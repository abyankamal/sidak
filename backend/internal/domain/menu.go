package domain

import "time"

type NavigasiMenu struct {
	ID        string    `json:"id"`
	ParentID  *string   `json:"parent_id,omitempty"`
	Label     string    `json:"label"`
	URL       string    `json:"url"`
	Urutan    int       `json:"urutan"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NavigasiMenuPublicChild struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	URL    string `json:"url"`
	Urutan int    `json:"urutan"`
}

type NavigasiMenuPublic struct {
	ID       string                    `json:"id"`
	Label    string                    `json:"label"`
	URL      string                    `json:"url"`
	Urutan   int                       `json:"urutan"`
	Children []NavigasiMenuPublicChild `json:"children"`
}

type NavigasiMenuAdmin struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parent_id,omitempty"`
	Label    string  `json:"label"`
	URL      string  `json:"url"`
	Urutan   int     `json:"urutan"`
	IsActive bool    `json:"is_active"`
}

type NavigasiMenuInput struct {
	ParentID *string `json:"parent_id,omitempty"`
	Label    string  `json:"label"`
	URL      string  `json:"url"`
	Urutan   int     `json:"urutan"`
	IsActive *bool   `json:"is_active,omitempty"`
}
