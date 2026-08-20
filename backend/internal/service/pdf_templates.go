package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/abyankamal/sidak/backend/internal/domain"
)

type SuratTemplateData struct {
	NomorSurat        string
	NamaLayanan       string
	TanggalSurat      string
	ProfilKelurahan   *domain.ProfilWilayah
	WargaNIK          string
	DataIsian         map[string]any
	PejabatNama       string
	PejabatNIP        string
	PejabatJabatan    string
	TransaksiID       string
}

const htmlTemplateKop = `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <title>{{.NamaLayanan}} - {{.NomorSurat}}</title>
    <style>
        @page {
            size: A4;
            margin: 20mm 20mm 20mm 20mm;
        }
        body {
            font-family: 'Times New Roman', Times, serif;
            font-size: 12pt;
            line-height: 1.5;
            color: #000;
            background: #fff;
            margin: 0;
            padding: 0;
        }
        .kop-container {
            text-align: center;
            border-bottom: 3px double #000;
            padding-bottom: 8px;
            margin-bottom: 20px;
        }
        .kop-instansi {
            font-size: 14pt;
            font-weight: bold;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin: 0;
        }
        .kop-kelurahan {
            font-size: 16pt;
            font-weight: bold;
            text-transform: uppercase;
            letter-spacing: 1.5px;
            margin: 2px 0;
        }
        .kop-alamat {
            font-size: 9.5pt;
            font-style: italic;
            margin: 0;
            line-height: 1.3;
        }
        .surat-title {
            text-align: center;
            margin-bottom: 24px;
        }
        .surat-title h2 {
            font-size: 13pt;
            font-weight: bold;
            text-decoration: underline;
            text-transform: uppercase;
            margin: 0 0 2px 0;
        }
        .surat-title p {
            font-size: 11pt;
            margin: 0;
        }
        .isi-surat {
            text-align: justify;
            margin-bottom: 16px;
        }
        .data-table {
            width: 100%;
            margin: 12px 0 16px 20px;
            border-collapse: collapse;
        }
        .data-table td {
            padding: 3px 0;
            vertical-align: top;
            font-size: 11.5pt;
        }
        .data-table td.label {
            width: 200px;
        }
        .data-table td.separator {
            width: 20px;
            text-align: center;
        }
        .ttd-container {
            width: 100%;
            margin-top: 30px;
        }
        .ttd-table {
            width: 100%;
            border-collapse: collapse;
        }
        .ttd-table td {
            vertical-align: top;
        }
        .ttd-kanan {
            width: 260px;
            text-align: center;
            float: right;
        }
        .ttd-space {
            height: 65px;
        }
        .ttd-nama {
            font-weight: bold;
            text-decoration: underline;
            margin: 0;
        }
        .ttd-nip {
            margin: 0;
            font-size: 11pt;
        }
        .digital-stamp {
            display: inline-block;
            border: 1px dashed #444;
            padding: 4px 8px;
            font-size: 8pt;
            color: #333;
            margin-top: 6px;
            border-radius: 3px;
        }
    </style>
</head>
<body>
    <!-- KOP SURAT RESMI -->
    <div class="kop-container">
        <div class="kop-instansi">Pemerintah Kabupaten {{.ProfilKelurahan.KabupatenKota}}</div>
        <div class="kop-instansi">Kecamatan {{.ProfilKelurahan.Kecamatan}}</div>
        <div class="kop-kelurahan">Kelurahan {{.ProfilKelurahan.NamaKelurahan}}</div>
        <div class="kop-alamat">{{.ProfilKelurahan.AlamatKantor}} {{if .ProfilKelurahan.KontakTelepon}}| Telp: {{.ProfilKelurahan.KontakTelepon}}{{end}} {{if .ProfilKelurahan.KontakEmail}}| Email: {{.ProfilKelurahan.KontakEmail}}{{end}}</div>
    </div>

    <!-- JUDUL & NOMOR SURAT -->
    <div class="surat-title">
        <h2>{{.NamaLayanan}}</h2>
        <p>Nomor: {{.NomorSurat}}</p>
    </div>

    <!-- ISI SURAT -->
    <div class="isi-surat">
        <p>Yang bertanda tangan di bawah ini, Kepala Kelurahan {{.ProfilKelurahan.NamaKelurahan}}, Kecamatan {{.ProfilKelurahan.Kecamatan}}, Kabupaten {{.ProfilKelurahan.KabupatenKota}}, dengan ini menerangkan bahwa:</p>

        <table class="data-table">
            <tr>
                <td class="label">Nomor Induk Kependudukan (NIK)</td>
                <td class="separator">:</td>
                <td><strong>{{.WargaNIK}}</strong></td>
            </tr>
            {{range $key, $value := .DataIsian}}
            <tr>
                <td class="label">{{formatKey $key}}</td>
                <td class="separator">:</td>
                <td>{{formatValue $value}}</td>
            </tr>
            {{end}}
        </table>

        <p>Berdasarkan data dan verifikasi yang telah dilakukan oleh petugas di lapangan, keterangan yang tercantum di atas adalah benar dan sesuai dengan kondisi yang bersangkutan.</p>
        <p>Demikian surat keterangan ini dibuat dengan sebenarnya untuk dapat dipergunakan sebagaimana mestinya.</p>
    </div>

    <!-- TANDA TANGAN -->
    <div class="ttd-container">
        <table class="ttd-table">
            <tr>
                <td style="width: 50%;"></td>
                <td style="width: 50%; text-align: center;">
                    <div>{{.ProfilKelurahan.NamaKelurahan}}, {{.TanggalSurat}}</div>
                    <div style="font-weight: bold; margin-bottom: 5px;">{{.PejabatJabatan}} {{.ProfilKelurahan.NamaKelurahan}}</div>
                    <div class="ttd-space">
                        <div class="digital-stamp">
                            ✓ DIVERIFIKASI DIGITAL SIDAK<br>
                            Dokumen ID: {{.TransaksiID}}
                        </div>
                    </div>
                    <div class="ttd-nama">{{.PejabatNama}}</div>
                    <div class="ttd-nip">NIP. {{.PejabatNIP}}</div>
                </td>
            </tr>
        </table>
    </div>
</body>
</html>
`

func renderSuratHTML(data SuratTemplateData) (string, error) {
	funcMap := template.FuncMap{
		"formatKey": func(k string) string {
			k = strings.ReplaceAll(k, "_", " ")
			words := strings.Fields(k)
			for i, w := range words {
				if len(w) > 0 {
					words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
				}
			}
			return strings.Join(words, " ")
		},
		"formatValue": func(v any) string {
			if v == nil {
				return "-"
			}
			switch val := v.(type) {
			case float64:
				if val == float64(int64(val)) {
					return fmt.Sprintf("%d", int64(val))
				}
				return fmt.Sprintf("%.2f", val)
			case int, int64:
				return fmt.Sprintf("%d", val)
			case bool:
				if val {
					return "Ya"
				}
				return "Tidak"
			case string:
				return val
			default:
				b, _ := json.Marshal(val)
				return string(b)
			}
		},
	}

	tmpl, err := template.New("surat").Funcs(funcMap).Parse(htmlTemplateKop)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func formatIndonesianDate(t time.Time) string {
	months := [...]string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()-1], t.Year())
}
