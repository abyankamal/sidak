package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/abyankamal/sidak/backend/config"
	"github.com/abyankamal/sidak/backend/internal/handler"
	"github.com/abyankamal/sidak/backend/internal/repository"
	"github.com/abyankamal/sidak/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TestEnv struct {
	DB               *pgxpool.Pool
	Config           *config.Config
	Router           http.Handler
	Server           *httptest.Server
	AuthService      *service.AuthService
	StorageService   *service.StorageService
	SyncService      *service.SyncService
	TransaksiService *service.TransaksiService
	CMSService       *service.CMSService
	SchemaCache      *service.SchemaCache
}

func SetupTestEnv(t *testing.T) *TestEnv {
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Use temporary test uploads directory
	cfg.StorageBasePath = t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		t.Fatalf("Failed to parse DB config: %v", err)
	}
	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1

	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}

	if err := dbPool.Ping(ctx); err != nil {
		t.Fatalf("DB ping failed: %v", err)
	}

	userRepo := repository.NewUserRepository(dbPool)
	templateRepo := repository.NewTemplateRepository(dbPool)
	transaksiRepo := repository.NewTransaksiRepository(dbPool)
	dokumenRepo := repository.NewDokumenRepository(dbPool)
	profilRepo := repository.NewProfilRepository(dbPool)
	menuRepo := repository.NewMenuRepository(dbPool)
	kontenRepo := repository.NewKontenRepository(dbPool)

	schemaCache := service.NewSchemaCache(templateRepo)
	if err := schemaCache.LoadAll(ctx); err != nil {
		t.Fatalf("Failed to load schema cache: %v", err)
	}

	authService := service.NewAuthService(userRepo, cfg.AppSecret)
	storageService := service.NewStorageService(cfg)
	syncService := service.NewSyncService(transaksiRepo, schemaCache, cfg)
	transaksiService := service.NewTransaksiService(transaksiRepo, cfg)
	pdfService := service.NewPDFService(dokumenRepo, transaksiRepo, templateRepo, profilRepo, userRepo, cfg)
	cmsService := service.NewCMSService(profilRepo, menuRepo, kontenRepo, cfg)

	// Start worker with dedicated context cancelled on test cleanup
	workerCtx, workerCancel := context.WithCancel(context.Background())
	pdfService.StartWorker(workerCtx)

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

	ts := httptest.NewServer(router)

	// Clean test transactions created during test runs & reset singleton profil
	_, _ = dbPool.Exec(context.Background(), "DELETE FROM transaksi_pelayanan WHERE warga_nik LIKE '320599%' OR warga_nik LIKE '32050199%' OR warga_nik LIKE '32050177%' OR warga_nik LIKE '32050166%'")
	_, _ = dbPool.Exec(context.Background(), `
		UPDATE profil_wilayah
		SET nama_kelurahan = 'Sukanegla',
		    kecamatan = 'Garut Kota',
		    kabupaten_kota = 'Kabupaten Garut',
		    visi = 'Terwujudnya Pelayanan Kelurahan Sukanegla yang Bersih, Prima, Responsif, dan Terpercaya Berbasis Digital.',
		    alamat_kantor = 'Jl. Sukanegla Raya No. 45, RT 02 / RW 03, Kelurahan Sukanegla, Garut Kota, Jawa Barat 44118'
		WHERE id = 1
	`)

	t.Cleanup(func() {
		workerCancel()
		ts.Close()
		_, _ = dbPool.Exec(context.Background(), "DELETE FROM transaksi_pelayanan WHERE warga_nik LIKE '320599%' OR warga_nik LIKE '32050199%' OR warga_nik LIKE '32050177%' OR warga_nik LIKE '32050166%'")
		_, _ = dbPool.Exec(context.Background(), `
			UPDATE profil_wilayah
			SET nama_kelurahan = 'Sukanegla',
			    kecamatan = 'Garut Kota',
			    kabupaten_kota = 'Kabupaten Garut',
			    visi = 'Terwujudnya Pelayanan Kelurahan Sukanegla yang Bersih, Prima, Responsif, dan Terpercaya Berbasis Digital.',
			    alamat_kantor = 'Jl. Sukanegla Raya No. 45, RT 02 / RW 03, Kelurahan Sukanegla, Garut Kota, Jawa Barat 44118'
			WHERE id = 1
		`)
		_ = os.RemoveAll(cfg.StorageBasePath)
		dbPool.Close()
	})

	return &TestEnv{
		DB:               dbPool,
		Config:           cfg,
		Router:           router,
		Server:           ts,
		AuthService:      authService,
		StorageService:   storageService,
		SyncService:      syncService,
		TransaksiService: transaksiService,
		CMSService:       cmsService,
		SchemaCache:      schemaCache,
	}
}
