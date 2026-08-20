package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abyankamal/sidak/backend/config"
	"github.com/abyankamal/sidak/backend/internal/domain"
	"github.com/abyankamal/sidak/backend/internal/repository"
	"github.com/oklog/ulid/v2"
)

var (
	ErrTransaksiNotReviewed = errors.New("dokumen PDF hanya dapat digenerate jika permohonan telah berstatus 'sudah_di_review'")
	ErrDokumenNotFound      = errors.New("data dokumen tidak ditemukan")
)

type PDFJob struct {
	DokumenID   string
	TransaksiID string
}

type PDFService struct {
	dokumenRepo   *repository.DokumenRepository
	transaksiRepo *repository.TransaksiRepository
	templateRepo  *repository.TemplateRepository
	profilRepo    *repository.ProfilRepository
	userRepo      *repository.UserRepository
	cfg           *config.Config
	jobQueue      chan PDFJob
	httpClient    *http.Client
}

func NewPDFService(
	dokumenRepo *repository.DokumenRepository,
	transaksiRepo *repository.TransaksiRepository,
	templateRepo *repository.TemplateRepository,
	profilRepo *repository.ProfilRepository,
	userRepo *repository.UserRepository,
	cfg *config.Config,
) *PDFService {
	return &PDFService{
		dokumenRepo:   dokumenRepo,
		transaksiRepo: transaksiRepo,
		templateRepo:  templateRepo,
		profilRepo:    profilRepo,
		userRepo:      userRepo,
		cfg:           cfg,
		jobQueue:      make(chan PDFJob, 100), // Buffer 100 jobs
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

// StartWorker initializes the single concurrent worker pool
func (s *PDFService) StartWorker(ctx context.Context) {
	go func() {
		log.Printf("Starting Gotenberg PDF Single Worker Pool (Max concurrency: 1)...")
		for {
			select {
			case <-ctx.Done():
				log.Printf("Stopping Gotenberg PDF Worker...")
				return
			case job, ok := <-s.jobQueue:
				if !ok {
					return
				}
				s.processJob(ctx, job)
			}
		}
	}()
}

func (s *PDFService) EnqueueJob(ctx context.Context, transaksiID string) (*repository.DokumenOutput, error) {
	trx, err := s.transaksiRepo.GetByID(ctx, transaksiID)
	if err != nil {
		return nil, err
	}
	if trx == nil {
		return nil, ErrTransaksiNotFound
	}

	if trx.Status != "sudah_di_review" {
		return nil, ErrTransaksiNotReviewed
	}

	// Generate or reuse Dokumen ID
	existingDoc, err := s.dokumenRepo.GetByTransaksiID(ctx, transaksiID)
	if err != nil {
		return nil, err
	}

	docID := ulid.Make().String()
	if existingDoc != nil {
		docID = existingDoc.ID
	}

	doc, err := s.dokumenRepo.UpsertProcessing(ctx, docID, transaksiID)
	if err != nil {
		return nil, err
	}

	// Non-blocking dispatch or queue
	select {
	case s.jobQueue <- PDFJob{DokumenID: docID, TransaksiID: transaksiID}:
	default:
		log.Printf("Warning: PDF job queue is full, dropping or waiting for job %s", docID)
		s.jobQueue <- PDFJob{DokumenID: docID, TransaksiID: transaksiID}
	}

	return doc, nil
}

func (s *PDFService) GetStatus(ctx context.Context, id string) (*domain.DokumenStatusResponse, error) {
	doc, err := s.dokumenRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		// Fallback: check if id is a transaksi_id
		doc, err = s.dokumenRepo.GetByTransaksiID(ctx, id)
		if err != nil {
			return nil, err
		}
		if doc == nil {
			return nil, ErrDokumenNotFound
		}
	}

	var downloadURL *string
	if doc.Status == "READY" && doc.FilePath != nil && *doc.FilePath != "" {
		cleanPath := strings.TrimPrefix(*doc.FilePath, "uploads/")
		url := fmt.Sprintf("%s/%s", strings.TrimRight(s.cfg.StoragePublicURL, "/"), cleanPath)
		downloadURL = &url
	}

	return &domain.DokumenStatusResponse{
		DokumenID:   doc.ID,
		Status:      doc.Status,
		DownloadURL: downloadURL,
	}, nil
}

func (s *PDFService) processJob(ctx context.Context, job PDFJob) {
	log.Printf("[PDF Worker] Processing job %s for transaksi %s", job.DokumenID, job.TransaksiID)

	trx, err := s.transaksiRepo.GetByID(ctx, job.TransaksiID)
	if err != nil || trx == nil {
		log.Printf("[PDF Worker Error] Failed to get transaksi %s: %v", job.TransaksiID, err)
		_ = s.dokumenRepo.UpdateStatus(ctx, job.DokumenID, "FAILED", nil)
		return
	}

	profil, err := s.profilRepo.Get(ctx)
	if err != nil || profil == nil {
		log.Printf("[PDF Worker Error] Failed to get profil: %v", err)
		_ = s.dokumenRepo.UpdateStatus(ctx, job.DokumenID, "FAILED", nil)
		return
	}

	templateName := trx.LayananID
	tpl, err := s.templateRepo.GetByID(ctx, trx.LayananID)
	if err == nil && tpl != nil {
		templateName = tpl.NamaLayanan
	}

	// Pejabat details
	pejabatNama := "Kepala Kelurahan Sukanegla"
	pejabatNIP := "-"
	pejabatJabatan := "Kepala"

	if trx.ReviewedBy != nil && *trx.ReviewedBy != "" {
		reviewer, err := s.userRepo.GetByID(ctx, *trx.ReviewedBy)
		if err == nil && reviewer != nil {
			pejabatNama = reviewer.Nama
			if reviewer.NIP != nil {
				pejabatNIP = *reviewer.NIP
			}
			switch reviewer.Role {
			case "LURAH":
				pejabatJabatan = "Lurah"
			case "SEKLUR":
				pejabatJabatan = "Sekretaris"
			case "KASI":
				pejabatJabatan = "Kasi Pelayanan"
			}
		}
	}

	var dataIsian map[string]any
	_ = json.Unmarshal(trx.DataIsian, &dataIsian)

	// Surat data
	romanMonths := [...]string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
	now := time.Now()
	nomorSurat := fmt.Sprintf("470/%s/Kel-Skn/%s/%d", job.TransaksiID[len(job.TransaksiID)-4:], romanMonths[now.Month()-1], now.Year())

	templateData := SuratTemplateData{
		NomorSurat:      nomorSurat,
		NamaLayanan:     templateName,
		TanggalSurat:    formatIndonesianDate(now),
		ProfilKelurahan: profil,
		WargaNIK:        trx.WargaNIK,
		DataIsian:       dataIsian,
		PejabatNama:     pejabatNama,
		PejabatNIP:      pejabatNIP,
		PejabatJabatan:  pejabatJabatan,
		TransaksiID:     job.TransaksiID,
	}

	htmlContent, err := renderSuratHTML(templateData)
	if err != nil {
		log.Printf("[PDF Worker Error] Failed to render HTML: %v", err)
		_ = s.dokumenRepo.UpdateStatus(ctx, job.DokumenID, "FAILED", nil)
		return
	}

	// Call Gotenberg Chromium API
	pdfBytes, err := s.callGotenbergHTML(ctx, htmlContent)
	if err != nil {
		log.Printf("[PDF Worker Error] Gotenberg conversion failed: %v", err)
		_ = s.dokumenRepo.UpdateStatus(ctx, job.DokumenID, "FAILED", nil)
		return
	}

	// Save to local disk (uploads/dokumen/{transaksi_id}_{layanan_id}.pdf)
	relDir := "dokumen"
	targetDir := filepath.Join(s.cfg.StorageBasePath, relDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Printf("[PDF Worker Error] Failed to create dir %s: %v", targetDir, err)
		_ = s.dokumenRepo.UpdateStatus(ctx, job.DokumenID, "FAILED", nil)
		return
	}

	fileName := fmt.Sprintf("%s_%s.pdf", job.TransaksiID, trx.LayananID)
	targetAbsPath := filepath.Join(targetDir, fileName)
	if err := os.WriteFile(targetAbsPath, pdfBytes, 0644); err != nil {
		log.Printf("[PDF Worker Error] Failed to write PDF: %v", err)
		_ = s.dokumenRepo.UpdateStatus(ctx, job.DokumenID, "FAILED", nil)
		return
	}

	storedPath := filepath.ToSlash(filepath.Join("uploads", relDir, fileName))
	if err := s.dokumenRepo.UpdateStatus(ctx, job.DokumenID, "READY", &storedPath); err != nil {
		log.Printf("[PDF Worker Error] Failed to update doc status: %v", err)
		return
	}

	log.Printf("[PDF Worker Success] Generated PDF for job %s: %s (%d bytes)", job.DokumenID, storedPath, len(pdfBytes))
}

func (s *PDFService) callGotenbergHTML(ctx context.Context, htmlContent string) ([]byte, error) {
	gotenbergEndpoint := fmt.Sprintf("%s/forms/chromium/convert/html", strings.TrimRight(s.cfg.GotenbergURL, "/"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add index.html part
	part, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, err
	}
	if _, err := io.WriteString(part, htmlContent); err != nil {
		return nil, err
	}

	// Add Chromium conversion properties (A4, print background, margins in inches)
	_ = writer.WriteField("paperWidth", "8.27")   // A4 width in inches
	_ = writer.WriteField("paperHeight", "11.69") // A4 height in inches
	_ = writer.WriteField("marginTop", "0.78")    // ~20mm
	_ = writer.WriteField("marginBottom", "0.78") // ~20mm
	_ = writer.WriteField("marginLeft", "0.78")   // ~20mm
	_ = writer.WriteField("marginRight", "0.78")  // ~20mm
	_ = writer.WriteField("printBackground", "true")

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gotenbergEndpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Gotenberg di %s: %w", gotenbergEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}
