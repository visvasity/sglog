package sglog

import (
	"log/slog"
	"os"
	"time"
)

type Options struct {
	// LogDirs if non-empty write log files in this directory.
	LogDirs []string

	// LogLink if non-empty, adds symbolic links in this directory to the log
	// files.
	LogLink string

	// MaxSize is the maximum size of a log file in bytes. Header and footer
	// messages are not accounted in this so final log file size can be larger
	// than this limit by a few hundred bytes.
	MaxSize uint64

	// BufferSize sizes the buffer associated with each log file. It's large so
	// that log records can accumulate without the logging thread blocking on
	// disk I/O. The flushDaemon will block instead.
	BufferSize int

	// FlushTimeout is the maximum buffering time interval before writing
	// messsages to the file.
	FlushTimeout time.Duration

	// LogBufLevel is the log level (and below) to buffer log messages.
	LogBufLevel slog.Level

	// Log levels enabled for logging.
	Levels []slog.Level

	// LogFileMode is the log file mode/permissions.
	LogFileMode os.FileMode

	// LogFileHeader when true writes the file header at the start of each log
	// file.
	LogFileHeader bool

	// ReuseFileDuration maximum duration to reuse the last log file as long as
	// it doesn't cross the maximum log file size.
	ReuseFileDuration time.Duration

	// MaxLogMessageLen is the limit on length of a formatted log message, including
	// the standard line prefix and trailing newline.
	MaxLogMessageLen int
}

func (v *Options) setDefaults() {
	if len(v.LogDirs) == 0 {
		v.LogDirs = []string{os.TempDir()}
	} else {
		v.LogDirs = append(v.LogDirs, os.TempDir())
	}
	if v.MaxSize == 0 {
		v.MaxSize = 1024 * 1024 * 1800
	}
	if v.BufferSize == 0 {
		v.BufferSize = 256 * 1024
	}
	if len(v.Levels) == 0 {
		v.Levels = []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	}
	if v.FlushTimeout == 0 {
		v.FlushTimeout = 30 * time.Second
	}
	if v.LogFileMode == 0 {
		v.LogFileMode = 0644
	}
	if v.ReuseFileDuration == 0 {
		v.ReuseFileDuration = 16 * time.Hour
	}
	if v.MaxLogMessageLen == 0 {
		v.MaxLogMessageLen = 15000
	}
}
