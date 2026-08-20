package domain

import "encoding/json"

type PresignUploadRequest struct {
	TransaksiID string `json:"transaksi_id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type PresignUploadResponse struct {
	UploadURL  string `json:"upload_url"`
	FilePathR2 string `json:"file_path_r2"`
}

type SyncCommitRequest struct {
	TransaksiID string          `json:"transaksi_id"`
	WargaNIK    string          `json:"warga_nik"`
	LayananID   string          `json:"layanan_id"`
	DataIsian   json.RawMessage `json:"data_isian"`
	Lampiran    []string        `json:"lampiran,omitempty"`
}
