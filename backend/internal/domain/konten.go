package domain

import "time"

type KontenPublik struct {
	ID                string     `json:"id"`
	Tipe              string     `json:"tipe"`
	Judul             string     `json:"judul"`
	Slug              string     `json:"slug"`
	Ringkasan         string     `json:"ringkasan"`
	IsiKonten         string     `json:"isi_konten"`
	ThumbnailFilePath *string    `json:"thumbnail_file_path,omitempty"`
	ThumbnailURL      *string    `json:"thumbnail_url,omitempty"`
	IsPublished       bool       `json:"is_published"`
	PublishedAt       *time.Time `json:"published_at,omitempty"`
	AuthorID          *string    `json:"author_id,omitempty"`
	AuthorNama        string     `json:"author_nama"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type KontenPublikListItem struct {
	ID           string     `json:"id"`
	Tipe         string     `json:"tipe"`
	Judul        string     `json:"judul"`
	Slug         string     `json:"slug"`
	Ringkasan    string     `json:"ringkasan"`
	ThumbnailURL *string    `json:"thumbnail_url,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
}

type KontenPublikListResponse struct {
	Data []KontenPublikListItem `json:"data"`
	Meta PaginationMeta         `json:"meta"`
}

type KontenPublikDetail struct {
	ID           string     `json:"id"`
	Tipe         string     `json:"tipe"`
	Judul        string     `json:"judul"`
	Slug         string     `json:"slug"`
	Ringkasan    string     `json:"ringkasan"`
	IsiKonten    string     `json:"isi_konten"`
	ThumbnailURL *string    `json:"thumbnail_url,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	AuthorNama   string     `json:"author_nama"`
}

type KontenPublikAdminItem struct {
	ID          string     `json:"id"`
	Tipe        string     `json:"tipe"`
	Judul       string     `json:"judul"`
	Slug        string     `json:"slug"`
	IsPublished bool       `json:"is_published"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	AuthorNama  string     `json:"author_nama"`
}

type KontenPublikAdminListResponse struct {
	Data []KontenPublikAdminItem `json:"data"`
	Meta PaginationMeta          `json:"meta"`
}

type KontenPublikInput struct {
	Tipe              string  `json:"tipe"`
	Judul             string  `json:"judul"`
	Ringkasan         string  `json:"ringkasan"`
	IsiKonten         string  `json:"isi_konten"`
	ThumbnailFilePath *string `json:"thumbnail_file_path,omitempty"`
	IsPublished       *bool   `json:"is_published,omitempty"`
}
