package domain

import "encoding/json"

type FileUploadResponse struct {
	FilePath string `json:"file_path"`
	FileURL  string `json:"file_url"`
}

type SyncCommitRequest struct {
	TransaksiID string          `json:"transaksi_id"`
	WargaNIK    string          `json:"warga_nik"`
	LayananID   string          `json:"layanan_id"`
	DataIsian   json.RawMessage `json:"data_isian"`
	Lampiran    []string        `json:"lampiran,omitempty"`
}
