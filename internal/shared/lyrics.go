package shared

import (
	"encoding/json"
	"time"
)

// LyricLine is a line of synchronized lyrics
type LyricLine struct {
	Timestamp time.Duration `json:"time"`
	Text      string        `json:"line"`

	// Active is used for detailed context
	Active bool `json:"active"`
}

// MarshalJSON implemetions json.Marshaller interface
func (l LyricLine) MarshalJSON() ([]byte, error) {
	type Alias LyricLine
	return json.Marshal(&struct {
		Alias
		Timestamp float64 `json:"time"`
	}{
		Alias:     (Alias)(l),
		Timestamp: l.Timestamp.Seconds(),
	})
}

// Lyrics is a slice of LyricLine
type Lyrics []LyricLine
