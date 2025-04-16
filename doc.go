// Package sglog provides a log/slog logging handler that writes to multiple
// log files based on the log severity -- similar to the Google's glog package.
//
// Most of the code is copied from the Google's glog package for Go. However,
// there are many differences and some new features, like log file reuse, etc.
//
// # DIFFERENCES
//
//   - The standard log/slog package doesn't define Fatal level, so FATAL
//     messages and log files are not supported by this package.
//
//   - Most of the global flags from glog package are replaced with an Options structure.
//
//   - Thread-ID field in the log file names is always set to zero (to enable
//     log file reuse across restarts). Thread-ID is still included in the
//     individual log messages.
//
//   - Google's glog package adds a footer message when a log file is rotated,
//     which is not supported.
//
//   - When the log file reuse feature is enabled, log file names do not
//     exactly match the log file creation time. However, timestamps in the log
//     file names still follow the sorted order.
//
// # LOG FILE REUSE
//
// Google's glog package creates a new log file every time the process is
// restarted. This can exhaust filesystem inodes if the process is crashing
// repeatedly.
//
// This package changes the above behavior and enables log file reuse with a
// certain timeout (ex: maximum one log file per hour.) As part of this,
// thread-id field in the log file name is replaced with a zero.
//
// Note that log file is still rotated when the file size reaches up to the
// maximum limit.
//
// # VMODULE USAGE
//
// In addition to the log levels, logging can also be enabled/disabled
// selectively using vmodule attributes. This is somewhat similar to glog
// package's vmodule feature.
//
// Users are required to create a reusable attribute (typically at global
// scope) for each module and use `slog.With` function to log module specific
// log messages.
//
//	var network = sglog.VModule("network", slog.LevelDebug)
//	...
//	slog.With(network).Info(...)
package sglog
