package waybar

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/Nadim147c/waybar-lyric/internal/config"
	"github.com/Nadim147c/waybar-lyric/internal/player"
	"github.com/Nadim147c/waybar-lyric/internal/shared"
	"github.com/Nadim147c/waybar-lyric/internal/str"
)

// ForPlayer returns printable Waybar
func ForPlayer(p *player.Info) *Waybar {
	alt := Status(Playing)
	if p.Status == "Paused" {
		alt = Paused
	}

	var text string
	if !config.LyricOnly {
		text = fmt.Sprintf("%s - %s", p.Artist, p.Title)
	}

	waybar := &Waybar{
		Class:      Class{alt},
		Text:       text,
		Alt:        alt,
		Percentage: p.Percentage(),
	}

	if config.Detailed {
		waybar.Info = p
	}

	return waybar
}

// ForLyrics returns Waybar for lyrics
func ForLyrics(lyrics shared.Lyrics, idx int) *Waybar {
	currentLine := lyrics[idx]
	start := max(idx-2, 0)
	end := min(idx+config.TooltipLines-2, len(lyrics))

	lyricsContext := slices.Clone(lyrics[start:end])

	var tooltip strings.Builder

	tooltip.WriteString(fmt.Sprintf("<span foreground=\"%s\">", config.TooltipColor))

	for i, ttl := range lyricsContext {
		line := str.BreakLine(ttl.Text, config.BreakTooltip)
		if ttl.Text == "" {
			line = "󰝚 "
		}

		if start+i == idx {
			lyricsContext[i].Active = true
			newLine := fmt.Sprintf("</span><b><big>%s</big></b>\n<span foreground=\"%s\">", line, config.TooltipColor)
			tooltip.WriteString(newLine)
			continue
		}

		tooltip.WriteString(line + "\n")
	}

	line := str.Truncate(currentLine.Text, config.MaxTextLength)
	tt := strings.TrimSpace(tooltip.String()) + "</span>"

	class := Class{Lyric, Playing}
	waybar := &Waybar{Alt: Lyric, Class: class, Text: line, Tooltip: tt}

	if config.Detailed {
		waybar.Context = &lyricsContext
	}

	return waybar
}

// Zero is a empty Waybar
var Zero = &Waybar{}

// Status is the alt/class for waybar
type Status string

const (
	//revive:disable
	Music   Status = "music"
	Lyric   Status = "lyric"
	Playing Status = "playing"
	Paused  Status = "paused"
	NoLyric Status = "no_lyric"
	Getting Status = "getting"
	//revive:enable
)

// Class is waybar class which can be either a string slice or string
type Class []Status

// Waybar is structure data which can be printed to for waybar output
type Waybar struct {
	Text       string         `json:"text"`
	Class      Class          `json:"class"`
	Alt        Status         `json:"alt"`
	Tooltip    string         `json:"tooltip"`
	Percentage int            `json:"percentage"`
	Info       *player.Info   `json:"info,omitempty"`
	Context    *shared.Lyrics `json:"context,omitempty"`
}

// JSON is the json encoder for waybar
var JSON = json.NewEncoder(os.Stdout)

func init() {
	JSON.SetEscapeHTML(false)
}

var lastLine string

// Encode prints the Waybar as json to Stdout
func (w *Waybar) Encode() {
	if config.LyricOnly && (w.Alt == Paused || w.Alt == Music) {
		fmt.Println("")
		return
	}

	if config.Compact {
		if lastLine != w.Text {
			fmt.Println(w.Text)
			lastLine = w.Text
		}
		return
	}

	if w == Zero {
		fmt.Println("{}")
		return
	}

	JSON.Encode(w)
}

// Is indecates if current Waybar is equal to another Waybar
func (w *Waybar) Is(other *Waybar) bool {
	if w == other {
		return true
	}
	if other == nil {
		return false
	}
	if w.Text != other.Text ||
		w.Alt != other.Alt ||
		w.Tooltip != other.Tooltip ||
		w.Percentage != other.Percentage {
		return false
	}

	if len(w.Class) != len(other.Class) {
		return false
	}
	for i := range w.Class {
		if w.Class[i] != other.Class[i] {
			return false
		}
	}

	if config.Detailed {
		if w.Info == nil && other.Info == nil {
			return true
		}
		if w.Info == nil || other.Info == nil {
			return false
		}
		if w.Info.Shuffle != other.Info.Shuffle {
			return false
		}
		if w.Info.Status != other.Info.Status {
			return false
		}
		if w.Info.Volume != other.Info.Volume {
			return false
		}
	}

	return true
}

// Paused set Text to artist and title on default mode
func (w *Waybar) Paused(info *player.Info) {
	if !config.LyricOnly {
		w.Text = fmt.Sprintf("%s - %s", info.Artist, info.Title)
	}
	w.Alt = Paused
	w.Class = Class{Paused}
}
