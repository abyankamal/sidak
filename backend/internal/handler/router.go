package handler

import (
	"net/http"

	"github.com/abyankamal/sidak/backend/config"
	"github.com/abyankamal/sidak/backend/internal/middleware"
	"github.com/abyankamal/sidak/backend/internal/service"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type RouterParams struct {
	Config           *config.Config
	AuthService      *service.AuthService
	StorageService   *service.StorageService
	SyncService      *service.SyncService
	TransaksiService *service.TransaksiService
	PDFService       *service.PDFService
	CMSService       *service.CMSService
	TemplateHandler  *TemplateHandler
}

func NewRouter(p RouterParams) http.Handler {
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	// CORS Configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", p.Config.WebPublicURL, p.Config.WebAdminURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Static File Server for Local Uploads
	uploadsDir := http.Dir(p.Config.StorageBasePath)
	r.Handle("/uploads/*", http.StripPrefix("/uploads", http.FileServer(uploadsDir)))

	// Handlers
	authHandler := NewAuthHandler(p.AuthService)
	storageHandler := NewStorageHandler(p.StorageService)
	syncHandler := NewSyncHandler(p.SyncService)
	transaksiHandler := NewTransaksiHandler(p.TransaksiService)
	pdfHandler := NewPDFHandler(p.PDFService)
	cmsPublicHandler := NewCMSPublicHandler(p.CMSService)
	cmsAdminHandler := NewCMSAdminHandler(p.CMSService)

	// API v1 Routing
	r.Route("/api/v1", func(v1 chi.Router) {
		// ---------------------------------------------------------------------
		// 1. AUTENTIKASI
		// ---------------------------------------------------------------------
		v1.Post("/auth/login", authHandler.Login)

		v1.Group(func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware(p.AuthService))
			auth.Post("/auth/logout", authHandler.Logout)
			auth.Get("/auth/me", authHandler.Me)
		})

		// ---------------------------------------------------------------------
		// 2. PENYIMPANAN BERKAS LOKAL
		// ---------------------------------------------------------------------
		v1.Group(func(storage chi.Router) {
			storage.Use(middleware.AuthMiddleware(p.AuthService))
			storage.Post("/storage/upload", storageHandler.Upload)
		})

		// ---------------------------------------------------------------------
		// 3. SINKRONISASI MOBILE (OFFLINE-FIRST)
		// ---------------------------------------------------------------------
		v1.Group(func(sync chi.Router) {
			sync.Use(middleware.AuthMiddleware(p.AuthService))
			sync.Post("/sync/commit", syncHandler.Commit)
		})

		// ---------------------------------------------------------------------
		// 4. PELAYANAN & VERIFIKASI
		// ---------------------------------------------------------------------
		v1.Group(func(pelayanan chi.Router) {
			pelayanan.Use(middleware.AuthMiddleware(p.AuthService))
			pelayanan.Get("/template-form", p.TemplateHandler.ListTemplates)
			pelayanan.Get("/transaksi", transaksiHandler.List)
			pelayanan.Get("/transaksi/{id}", transaksiHandler.GetDetail)
			pelayanan.Patch("/transaksi/{id}/review", transaksiHandler.Review)
		})

		// ---------------------------------------------------------------------
		// 5. DOKUMEN & PDF ENGINE
		// ---------------------------------------------------------------------
		v1.Group(func(doc chi.Router) {
			doc.Use(middleware.AuthMiddleware(p.AuthService))
			doc.Post("/layanan/{id}/generate-pdf", pdfHandler.GeneratePDF)
			doc.Get("/dokumen/{id}/status", pdfHandler.GetStatus)
		})

		// ---------------------------------------------------------------------
		// 6. CMS PUBLIK (UNAUTHENTICATED)
		// ---------------------------------------------------------------------
		v1.Get("/public/profil", cmsPublicHandler.GetProfil)
		v1.Get("/public/menu", cmsPublicHandler.GetMenu)
		v1.Get("/public/konten", cmsPublicHandler.GetKontenList)
		v1.Get("/public/konten/{slug}", cmsPublicHandler.GetKontenDetail)

		// ---------------------------------------------------------------------
		// 7. CMS ADMIN (AUTHENTICATED LURAH/SEKLUR/KASI)
		// ---------------------------------------------------------------------
		v1.Group(func(admin chi.Router) {
			admin.Use(middleware.AuthMiddleware(p.AuthService))
			admin.Use(middleware.RequireRoles("LURAH", "SEKLUR", "KASI"))

			admin.Put("/cms/profil", cmsAdminHandler.UpdateProfil)
			admin.Get("/cms/menu", cmsAdminHandler.ListMenu)
			admin.Post("/cms/menu", cmsAdminHandler.CreateMenu)
			admin.Put("/cms/menu/{id}", cmsAdminHandler.UpdateMenu)
			admin.Delete("/cms/menu/{id}", cmsAdminHandler.DeleteMenu)

			admin.Get("/cms/konten", cmsAdminHandler.ListKonten)
			admin.Post("/cms/konten", cmsAdminHandler.CreateKonten)
			admin.Put("/cms/konten/{id}", cmsAdminHandler.UpdateKonten)
			admin.Delete("/cms/konten/{id}", cmsAdminHandler.DeleteKonten)
		})
	})

	return r
}
