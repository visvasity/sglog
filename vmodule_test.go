package sglog

import (
	"log"
	"testing"

	"log/slog"
)

func TestVModuleLogging(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	backend := NewBackend(&Options{
		LogFileMaxSize: 5 * 1024 * 1024,
		LogDirs:        []string{"/tmp", t.TempDir()},
	})
	defer backend.Close()

	slog.SetDefault(slog.New(backend.Handler()))

	network := VModule("network", slog.LevelInfo)
	storage := VModule("storage", slog.LevelInfo)

	slog.Debug("this is a debug message")
	slog.Info("this is an info message")
	slog.Warn("this is a warning message")
	slog.Error("this is an error message")

	slog.With(storage).Debug("this is a storage module's first debug message ")
	slog.With(storage).Info("this is a storage module's first info message ")
	slog.With(storage).Warn("this is a storage module's first warn message ")
	slog.With(storage).Error("this is a storage module's first error message ")

	slog.With(network).Debug("this is a network module's first debug message ")
	slog.With(network).Info("this is a network module's first info message ")
	slog.With(network).Warn("this is a network module's first warn message ")
	slog.With(network).Error("this is a network module's first error message ")

	SetVModuleLevel(network, slog.LevelDebug)
	SetVModuleLevel(storage, slog.LevelError)

	slog.Debug("this is a debug message")
	slog.Info("this is an info message")
	slog.Warn("this is a warning message")
	slog.Error("this is an error message")

	slog.With(storage).Debug("this is a storage module's second debug message ")
	slog.With(storage).Info("this is a storage module's second info message ")
	slog.With(storage).Warn("this is a storage module's second warn message ")
	slog.With(storage).Error("this is a storage module's second error message ")

	slog.With(network).Debug("this is a network module's second debug message ")
	slog.With(network).Info("this is a network module's second info message ")
	slog.With(network).Warn("this is a network module's second warn message ")
	slog.With(network).Error("this is a network module's second error message ")
}
