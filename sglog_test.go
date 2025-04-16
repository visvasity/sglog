package sglog

import (
	"log"
	"testing"

	"log/slog"
)

func TestBasic(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	backend := NewBackend(&Options{
		LogFileMaxSize: 5 * 1024 * 1024,
	})
	defer backend.Flush()

	slog.SetDefault(slog.New(backend.Handler()))

	slog.Info("info message", "key", "value", "one", 1)
	slog.Warn("warning message", "key", "value", "one", 1)
	slog.Error("error message", "key", "value", "one", 1)

	slog.Debug("debug message before EnableDebugLog")
	backend.SetLevel(slog.LevelDebug)
	slog.Debug("debug message after EnableDebugLog")
	slog.Info("info message after EnableDebugLog")
	backend.SetLevel(0)
	slog.Debug("debug message after DisableDebugLog")
	slog.Info("info message after DisableDebugLog")

	slog.Info("info message with group", slog.Group("g", slog.Int("a", 1), slog.Int("b", 2)))
	slog.Info("info message with group", slog.Group("g1", slog.Group("g2", slog.Int("a", 1), slog.Int("b", 2))))

	log.Printf("hello world %d", 123)
}

func TestLogFileRotation(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	backend := NewBackend(&Options{
		LogFileMaxSize: 1024 * 1024,
		LogDirs:        []string{"."},
	})
	defer backend.Flush()

	slog.SetDefault(slog.New(backend.Handler()))

	for i := 0; i < 1024; i++ {
		slog.Info("info message", "key", "value", "iteration", i)
		slog.Warn("warning message", "key", "value", "iteration", i)
		slog.Error("error message", "key", "value", "iteration", i)

		slog.Debug("debug message before EnableDebugLog", "iteration", i)
		backend.SetLevel(slog.LevelDebug)
		slog.Debug("debug message after EnableDebugLog", "iteration", i)
		slog.Info("info message after EnableDebugLog", "iteration", i)
		backend.SetLevel(slog.LevelInfo)
		slog.Debug("debug message after DisableDebugLog", "iteration", i)
		slog.Info("info message after DisableDebugLog", "iteration", i)

		slog.Info("info message with group", "iteration", i, slog.Group("g", slog.Int("a", 1), slog.Int("b", 2)))
		slog.Info("info message with group", "iteration", i, slog.Group("g1", slog.Group("g2", slog.Int("a", 1), slog.Int("b", 2))))

		log.Printf("hello world [iteration=%d]", i)
	}
}
