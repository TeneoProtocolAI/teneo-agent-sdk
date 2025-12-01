package modules

import (
	"encoding/json"
	"os"
	"time"
)

// Detection represents a single detection / signal from the scanner.
type Detection struct {
	KOL        string    `json:"kol"`        // nama KOL / influencer
	Token      string    `json:"token"`      // ticker / token id / nama
	Signal     string    `json:"signal"`     // e.g. "early_call", "dump_warning"
	Confidence float64   `json:"confidence"` // 0..1
	Source     string    `json:"source"`     // source layanan: "x", "coingecko", etc.
	Text       string    `json:"text"`       // raw text/snippet yang memicu deteksi
	Link       string    `json:"link"`       // optional link (tweet, post, tx)
	Timestamp  time.Time `json:"timestamp"`  // waktu deteksi
}

// SaveDetection writes a Detection as pretty JSON to filename (overwrites/creates).
// This is simple and safe for mock/stub usage.
func SaveDetection(filename string, d Detection) error {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	// write file (overwrite). If you prefer append, change os.WriteFile to open/append style.
	return os.WriteFile(filename, b, 0644)
}
