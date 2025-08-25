package lyric

import (
	"encoding/json"
	"time"
)

// LrcLibResponse is the response sent from LrcLib api
type LrcLibResponse struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

// Line is a line of synchronized lyrics
type Line struct {
	Timestamp time.Duration `json:"time"`
	Text      string        `json:"line"`

	// Active is used for detailed context
	Active bool `json:"active"`
}

// MarshalJSON implemetions json.Marshaller interface
func (l Line) MarshalJSON() ([]byte, error) {
	type Alias Line
	return json.Marshal(&struct {
		Alias
		Timestamp float64 `json:"time"`
	}{
		Alias:     (Alias)(l),
		Timestamp: l.Timestamp.Seconds(),
	})
}

// Lyrics is a slice of LyricLine
type Lyrics []Line
