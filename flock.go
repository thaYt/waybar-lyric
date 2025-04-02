package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"
)

// Flock ensures only one instance of the CLI is running.
//
// Note: returned *os.File needs to be closed once the program is exited. Run defer lock.Close()
func Flock() (*os.File, error) {
	lockFile := filepath.Join(os.TempDir(), "waybar-lyric.lock")

	file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		slog.Error("Failed to open or create lock file", "error", err)
		return file, err
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK {
			slog.Warn("Another instance of the CLI is already running. Exiting.")
			return file, err
		}
		slog.Error("Failed to acquire lock", "error", err)
		return file, fmt.Errorf("Failed to acquire lock: %+v", err)
	}

	return file, err
}
