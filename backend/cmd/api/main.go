package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abyankamal/sidak/backend/config"
	"github.com/abyankamal/sidak/backend/internal/handler"
	"github.com/abyankamal/sidak/backend/internal/repository"
	"github.com/abyankamal/sidak/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize PostgreSQL Connection Pool
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Invalid database connection URL: %v", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.MaxConnLifetime = 1 * time.Hour

	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to create database connection pool: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Printf("Successfully connected to PostgreSQL at %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	// 2. Initialize Repositories
	userRepo := repository.NewUserRepository(dbPool)
	templateRepo := repository.NewTemplateRepository(dbPool)
	transaksiRepo := repository.NewTransaksiRepository(dbPool)
	dokumenRepo := repository.NewDokumenRepository(dbPool)
	profilRepo := repository.NewProfilRepository(dbPool)
	menuRepo := repository.NewMenuRepository(dbPool)
	kontenRepo := repository.NewKontenRepository(dbPool)

	// 3. Initialize Services & In-Memory Schema Cache
	schemaCache := service.NewSchemaCache(templateRepo)
	if err := schemaCache.LoadAll(ctx); err != nil {
		log.Printf("Warning: Initial schema cache loading failed: %v", err)
	} else {
		log.Printf("In-memory JSON Schema cache successfully loaded")
	}

	authService := service.NewAuthService(userRepo, cfg.AppSecret)
	storageService := service.NewStorageService(cfg)
	syncService := service.NewSyncService(transaksiRepo, schemaCache, cfg)
	transaksiService := service.NewTransaksiService(transaksiRepo, cfg)
	pdfService := service.NewPDFService(dokumenRepo, transaksiRepo, templateRepo, profilRepo, userRepo, cfg)
	cmsService := service.NewCMSService(profilRepo, menuRepo, kontenRepo, cfg)

	// Start Gotenberg PDF Worker (Single concurrent worker pool)
	pdfService.StartWorker(ctx)

	// 4. Initialize Router
	router := handler.NewRouter(handler.RouterParams{
		Config:           cfg,
		AuthService:      authService,
		StorageService:   storageService,
		SyncService:      syncService,
		TransaksiService: transaksiService,
		PDFService:       pdfService,
		CMSService:       cmsService,
		TemplateHandler:  handler.NewTemplateHandler(templateRepo),
	})

	// 5. Start HTTP Server
	serverAddr := fmt.Sprintf(":%s", cfg.AppPort)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("SIDAK Backend API server listening on %s (ENV: %s)", serverAddr, cfg.AppEnv)
		serverErrors <- srv.ListenAndServe()
	}()

	// 6. Graceful Shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Caught signal %v, initiating graceful shutdown...", sig)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = srv.Close()
			log.Fatalf("Could not stop server gracefully: %v", err)
		}
		log.Printf("Server stopped cleanly.")
	}
}
