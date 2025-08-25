package player

import (
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"slices"
	"strconv"

	"github.com/Nadim147c/go-mpris"
	"github.com/godbus/dbus/v5"
	"github.com/spf13/cast"
)

var (
	// ErrNoPlayerVolume when failed to get player volume
	ErrNoPlayerVolume = errors.New("failed to get player volume")
	// ErrNoArtists when failed to get artists
	ErrNoArtists = errors.New("failed to get artists")
	// ErrNoTitle when failed to get title
	ErrNoTitle = errors.New("failed to get title")
	// ErrNoID when failed to get id
	ErrNoID = errors.New("failed to get track id")
)

// Parser parses player information from mpris metadata
type Parser func(*mpris.Player) (*Info, error)

// IDFunc extracts a stable ID for a player.
type IDFunc func(p *mpris.Player) (uint64, error)

// trackIDFunc: uses mpris:trackid as ID source
func trackIDFunc(p *mpris.Player) (uint64, error) {
	meta, err := p.GetMetadata()
	if err != nil {
		return 0, err
	}
	val, ok := meta["mpris:trackid"]
	if !ok {
		return 0, ErrNoID
	}
	trackid := cast.ToString(val.Value())

	h := fnv.New64a()
	_, _ = fmt.Fprint(h, trackid)
	return h.Sum64(), nil
}

// artistTitleFunc: uses artist+title combo as ID source
func artistTitleFunc(p *mpris.Player) (uint64, error) {
	artists, err := p.GetArtist()
	if err != nil || len(artists) == 0 {
		return 0, ErrNoArtists
	}
	artist := artists[0]

	title, err := p.GetTitle()
	if err != nil || title == "" {
		return 0, ErrNoTitle
	}

	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%s:%s", artist, title)
	return h.Sum64(), nil
}

var supportedPlayers = map[string]IDFunc{
	"spotify":          trackIDFunc,
	"YoutubeMusic":     trackIDFunc,
	"amarok":           artistTitleFunc,
	"io.bassi.Amberol": artistTitleFunc,
}

// Select selects correct parses for player
func Select(conn *dbus.Conn) (*mpris.Player, Parser, error) {
	players, err := mpris.List(conn)
	if err != nil {
		return nil, nil, err
	}
	slog.Debug("Player names", "players", players)

	if len(players) == 0 {
		return nil, nil, errors.New("No player exists")
	}

	for name, idFunc := range supportedPlayers {
		for player := range slices.Values(players) {
			if mpris.BaseInterface+"."+name == player {
				return mpris.New(conn, player), parserWithIDFunc(DefaultParser, idFunc), nil
			}
		}
	}

	return nil, nil, errors.New("No player exists")
}

func parserWithIDFunc(f Parser, i IDFunc) Parser {
	return func(p *mpris.Player) (*Info, error) {
		info, err := f(p)
		if err != nil {
			return info, err
		}
		id, err := i(p)
		if err != nil {
			return info, err
		}

		info.ID = strconv.FormatUint(id, 32)
		return info, nil
	}
}

// DefaultParser takes *mpris.Player of spotify and return *PlayerInfo
func DefaultParser(player *mpris.Player) (*Info, error) {
	meta, err := player.GetMetadata()
	if err != nil {
		return nil, err
	}
	for k, v := range meta {
		slog.Debug("MPRIS", k, v)
	}

	shuffle, err := player.GetShuffle()
	if err != nil {
		return nil, err
	}

	status, err := player.GetPlaybackStatus()
	if err != nil {
		return nil, err
	}

	// Cover is optional
	cover, _ := player.GetCoverURL()

	volume, err := player.GetVolume()
	if err != nil {
		return nil, ErrNoPlayerVolume
	}

	length, err := player.GetLength()
	if err != nil {
		return nil, err
	}

	artistList, err := player.GetArtist()
	if err != nil {
		return nil, err
	}

	if len(artistList) == 0 {
		return nil, ErrNoArtists
	}

	artist := artistList[0]

	title, err := player.GetTitle()
	if err != nil {
		return nil, err
	}

	if title == "" {
		return nil, ErrNoArtists
	}

	idValue, _ := meta["mpris:trackid"]
	trackid := cast.ToString(idValue.Value())

	album, _ := player.GetAlbum()

	info := &Info{
		Player:  player.GetName(),
		ID:      trackid,
		Artist:  artist,
		Title:   title,
		Album:   album,
		Status:  status,
		Volume:  volume,
		Length:  length,
		Shuffle: shuffle,
		Cover:   cover,
	}

	err = info.UpdatePosition(player)
	return info, err
}
