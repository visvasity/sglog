package sglog

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
)

type Backend struct {
	mu sync.Mutex
	wg sync.WaitGroup

	opts *Options

	handler *slogHandler

	fileMap map[slog.Level]*levelFile

	currentLevel slog.LevelVar
}

// NewBackend creates a slog backend.
func NewBackend(opts *Options) *Backend {
	opts.setDefaults()
	v := &Backend{
		opts:    opts,
		fileMap: make(map[slog.Level]*levelFile),
	}
	v.handler = v.newHandler(opts)

	for _, l := range v.opts.Levels {
		v.fileMap[l] = v.newLevelFile(v.opts, l)
	}
	return v
}

// Close flushes the logs and waits for the background goroutine to finish.
func (v *Backend) Close() {
	v.wg.Wait()
}

// Handler returns slog.Handler for the log backend.
func (v *Backend) Handler() slog.Handler {
	return v.handler
}

// EnableDebugLog enables logging for slog.LevelDebug messages.
func (v *Backend) EnableDebugLog() {
	v.currentLevel.Set(slog.LevelDebug)
}

// DisableDebugLog disables logging for slog.LevelDebug messages.
func (v *Backend) DisableDebugLog() {
	v.currentLevel.Set(slog.LevelInfo)
}

func normalize(v slog.Level) slog.Level {
	if v >= slog.LevelError {
		return slog.LevelError
	}
	if v >= slog.LevelWarn {
		return slog.LevelWarn
	}
	if v >= slog.LevelInfo {
		return slog.LevelInfo
	}
	if v >= slog.LevelDebug {
		return slog.LevelDebug
	}
	return slog.LevelDebug
}

func (v *Backend) emit(level slog.Level, msg []byte) error {
	level = normalize(level)

	v.mu.Lock()
	var firstErr error
	for l, f := range v.fileMap {
		if l < v.currentLevel.Level() {
			continue
		}
		if l > level {
			continue
		}
		if _, err := f.Write(msg); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	v.mu.Unlock()

	if firstErr != nil {
		fmt.Fprintf(os.Stderr, "could not emit log message for level %d: %v\n", level, firstErr)
	}
	return firstErr
}

// Flush force writes log messages to the log files.
func (v *Backend) Flush() error {
	return nil // v.flush(slog.LevelDebug)
}
