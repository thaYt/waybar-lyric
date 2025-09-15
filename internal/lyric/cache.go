package lyric

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/Nadim147c/waybar-lyric/internal/player"
	"github.com/Nadim147c/waybar-lyric/internal/shared"
)

// storeValue is Lyrics with LastAccess time
type storeValue struct {
	LastAccess time.Time
	Lyrics     shared.Lyrics
}

// store is used to cache lyrics in memory
type store struct {
	mu   sync.RWMutex // Using RWMutex for better read performance
	data map[string]*storeValue
}

// newStore creates a new initialized Store
func newStore() *store {
	return &store{
		data: map[string]*storeValue{},
	}
}

// Save saves lyrics to Store
func (s *store) Save(id string, lyrics shared.Lyrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = &storeValue{
		LastAccess: time.Now(),
		Lyrics:     lyrics,
	}
}

// Load loads lyrics from Store
func (s *store) Load(key string) (shared.Lyrics, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, exists := s.data[key]
	if !exists {
		return nil, false
	}
	v.LastAccess = time.Now() // Update last access time
	return v.Lyrics, true
}

// Cleanup runs a blocking loop that periodically removes unused entries
// until the context is canceled.
func (s *store) Cleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return // Exit when context is canceled
		case <-ticker.C:
			s.cleanupExpired(interval)
		}
	}
}

// cleanupExpired removes entries not accessed within the interval
func (s *store) cleanupExpired(threshold time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, value := range s.data {
		if time.Since(value.LastAccess) > threshold {
			delete(s.data, key)
		}
	}
}

// CacheDir is waybar-lyric lyrics cache dir
var CacheDir string

func init() {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		slog.Error("Failed to find cache directory", "error", err)
		return
	}

	CacheDir = filepath.Join(userCacheDir, "waybar-lyric")

	if err := os.MkdirAll(CacheDir, 0755); err != nil {
		slog.Error("Failed to create cache directory")
	}
}

// SaveCache saves the lyrics to cache
func SaveCache(info *player.Info, lines shared.Lyrics, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write player info as comments before lyrics
	fmt.Fprintf(file, "# PLAYER: %s\n", info.Player)
	fmt.Fprintf(file, "# ID: %s\n", info.ID)
	metaLines := make([]string, 0, len(info.Metadata))
	for k, v := range info.Metadata {
		// Convert metadata key to SNAKE_CASE
		keysplit := strings.SplitN(k, ":", 2)
		if len(keysplit) != 2 {
			continue
		}
		var key strings.Builder
		for _, r := range keysplit[1] {
			if unicode.IsUpper(r) {
				key.WriteByte('_')
			}
			key.WriteRune(unicode.ToUpper(r))
		}
		metaLines = append(metaLines, fmt.Sprintf("# %s: %s", key.String(), v))
	}

	slices.SortFunc(metaLines, func(a, b string) int {
		return strings.Compare(a, b)
	})

	for line := range slices.Values(metaLines) {
		fmt.Fprintln(file, line)
	}

	// Write lyrics
	for line := range slices.Values(lines) {
		_, err := fmt.Fprintf(file, "%d,%s\n", line.Timestamp, line.Text)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadCache loads the lyrics from cache
func LoadCache(filePath string) (shared.Lyrics, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lyrics shared.Lyrics
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			continue // Skip invalid lines
		}

		ts, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}

		timestamp := time.Duration(ts)
		text := strings.TrimSpace(parts[1])

		lyric := shared.LyricLine{Timestamp: timestamp, Text: text}
		lyrics = append(lyrics, lyric)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lyrics) == 0 {
		return nil, errors.New("Number of line found is zero")
	}

	return lyrics, nil
}
