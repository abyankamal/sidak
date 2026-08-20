package domain

type ProfilWilayah struct {
	NamaKelurahan         string   `json:"nama_kelurahan"`
	Kecamatan             string   `json:"kecamatan"`
	KabupatenKota         string   `json:"kabupaten_kota"`
	Visi                  string   `json:"visi"`
	Misi                  []string `json:"misi"`
	Sejarah               *string  `json:"sejarah,omitempty"`
	AlamatKantor          string   `json:"alamat_kantor"`
	KontakTelepon         *string  `json:"kontak_telepon,omitempty"`
	KontakEmail           *string  `json:"kontak_email,omitempty"`
	StrukturOrganisasiURL *string  `json:"struktur_organisasi_url,omitempty"`
}

type ProfilWilayahInput struct {
	NamaKelurahan              string   `json:"nama_kelurahan"`
	Kecamatan                  string   `json:"kecamatan"`
	KabupatenKota              string   `json:"kabupaten_kota"`
	Visi                       string   `json:"visi"`
	Misi                       []string `json:"misi"`
	Sejarah                    *string  `json:"sejarah,omitempty"`
	AlamatKantor               string   `json:"alamat_kantor"`
	KontakTelepon              *string  `json:"kontak_telepon,omitempty"`
	KontakEmail                *string  `json:"kontak_email,omitempty"`
	StrukturOrganisasiFilePath *string  `json:"struktur_organisasi_file_path,omitempty"`
}
