package domain

import (
	"encoding/json"
	"time"
)

type TransaksiPelayanan struct {
	ID            string          `json:"id"`
	WargaNIK      string          `json:"warga_nik"`
	LayananID     string          `json:"layanan_id"`
	DataIsian     json.RawMessage `json:"data_isian"`
	Lampiran      []string        `json:"lampiran"`
	Status        string          `json:"status"`
	CatatanReview *string         `json:"catatan_review,omitempty"`
	ReviewedBy    *string         `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time      `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type PaginationMeta struct {
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	TotalRecords int `json:"total_records"`
}

type TransaksiListItem struct {
	ID        string    `json:"id"`
	WargaNIK  string    `json:"warga_nik"`
	LayananID string    `json:"layanan_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type TransaksiListResponse struct {
	Data []TransaksiListItem `json:"data"`
	Meta PaginationMeta      `json:"meta"`
}

type TransaksiDetailResponse struct {
	ID                     string          `json:"id"`
	WargaNIK               string          `json:"warga_nik"`
	LayananID              string          `json:"layanan_id"`
	DataIsian              json.RawMessage `json:"data_isian"`
	LampiranPresignedURLs  []string        `json:"lampiran_presigned_urls"`
	Status                 string          `json:"status"`
	CatatanReview          *string         `json:"catatan_review,omitempty"`
	ReviewedBy             *string         `json:"reviewed_by,omitempty"`
	ReviewedAt             *time.Time      `json:"reviewed_at,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
}

type ReviewTransaksiRequest struct {
	Status        string  `json:"status"`
	CatatanReview *string `json:"catatan_review,omitempty"`
}

type DokumenStatusResponse struct {
	DokumenID   string  `json:"dokumen_id"`
	Status      string  `json:"status"`
	DownloadURL *string `json:"download_url,omitempty"`
}

type TransaksiFilter struct {
	Status    string
	LayananID string
	Page      int
	Limit     int
}
