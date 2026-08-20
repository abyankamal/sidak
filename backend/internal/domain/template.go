package domain

import (
	"encoding/json"
	"time"
)

type TemplateForm struct {
	LayananID   string          `json:"layanan_id"`
	NamaLayanan string          `json:"nama_layanan"`
	Deskripsi   *string         `json:"deskripsi,omitempty"`
	SkemaJSON   json.RawMessage `json:"skema_json"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at,omitempty"`
	UpdatedAt   time.Time       `json:"updated_at,omitempty"`
}
