package sglog

import (
	"os"
	"time"
)

type Options struct {
	// Name holds the program name to use with the log files.
	Name string

	// LogDirs if non-empty create/write log files in one of these directories.
	LogDirs []string

	// LogLinkDir if non-empty, adds symbolic links in this directory to the log
	// files.
	LogLinkDir string

	// LogFileMaxSize is the maximum size of a log file in bytes. Header and
	// footer messages are not accounted in this so final log file size can be
	// larger than this limit by a few hundred bytes.
	LogFileMaxSize uint64

	// LogFileMode is the log file mode/permissions.
	LogFileMode os.FileMode

	// LogFileHeader when true writes the file header at the start of each log
	// file.
	LogFileHeader bool

	// LogFileReuseDuration is the maximum duration to reuse/reopen an existing
	// log file as long as it doesn't cross the maximum log file size.
	LogFileReuseDuration time.Duration

	// LogMessageMaxLen is the limit on length of a formatted log message,
	// including the standard line prefix and trailing newline. Messages longer
	// than this value are truncated.
	LogMessageMaxLen int
}

func (v *Options) setDefaults() {
	if v.Name == "" {
		v.Name = program
	}
	if len(v.LogDirs) == 0 {
		v.LogDirs = []string{os.TempDir()}
	} else {
		v.LogDirs = append(v.LogDirs, os.TempDir())
	}
	if v.LogFileMaxSize == 0 {
		v.LogFileMaxSize = 1024 * 1024 * 1800
	}
	if v.LogFileMode == 0 {
		v.LogFileMode = 0644
	}
	if v.LogFileReuseDuration == 0 {
		v.LogFileReuseDuration = 16 * time.Hour
	}
	if v.LogMessageMaxLen == 0 {
		v.LogMessageMaxLen = 15000
	}
}
