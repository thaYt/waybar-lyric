package player

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Nadim147c/go-mpris"
)

// Info holds all information of currently playing track metadata
type Info struct {
	Player string `json:"player"`
	ID     string `json:"id"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
	Album  string `json:"album"`
	Cover  string `json:"cover"`

	Volume   float64       `json:"volume"`
	Position time.Duration `json:"position"`
	Length   time.Duration `json:"length"`
	Shuffle  bool          `json:"shuffle"`

	Status mpris.PlaybackStatus `json:"status"`
}

// MarshalJSON encodes PlayerInfo with durations in seconds (float)
func (p Info) MarshalJSON() ([]byte, error) {
	p.Player = p.Player[23:]
	type Alias Info // create alias to avoid recursion
	return json.Marshal(&struct {
		Alias
		// Position is seconds as float64
		Position float64 `json:"position"`
		// Length is seconds as float64
		Length float64 `json:"length"`
	}{
		Alias:    (Alias)(p),
		Position: p.Position.Seconds(),
		Length:   p.Length.Seconds(),
	})
}

// Percentage is player position in percentage rounded to int
func (p *Info) Percentage() int {
	return int(((p.Position * 100) / p.Length))
}

// UpdatePosition updates the position of player
func (p *Info) UpdatePosition(player *mpris.Player) error {
	pos, err := player.GetPosition()
	if err != nil {
		return err
	}
	p.Position = pos

	// HACK: YoutubeMusic dbus position ≈ 1.1 slow
	if player.GetName() == mpris.BaseInterface+".YoutubeMusic" {
		slog.Debug("Adding 1.1 second to adjust mpris delay")
		p.Position += 1100 * time.Millisecond
	}

	return nil
}
