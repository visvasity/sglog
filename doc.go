// Package sglog provides a logging handler for the log/slog package that writes
// log messages to multiple files based on their severity, similar to Google's
// glog package.
//
// This package is inspired by Google's glog package for Go but introduces
// several differences and new features, such as log file reuse. Most of the
// code is adapted from glog, with modifications to support the slog ecosystem
// and additional functionality.
//
// # Differences from Google's glog
//
//   - Log messages are not buffered inline with the application control flow.
//   - The standard log/slog package does not define a Fatal level, so FATAL
//     messages and log files are not supported.
//   - Global flags from glog are replaced with an Options struct for configuration.
//   - Unlike glog, this package does not add a footer message when rotating log files.
//   - When log file reuse is enabled, log file names may not precisely reflect
//     the log file creation time, though timestamps in file names remain in
//     chronological order.
//
// # Log File Reuse
//
// Google's glog creates a new log file each time a process restarts, which can
// exhaust filesystem inodes if the process crashes repeatedly. The sglog package
// mitigates this by enabling log file reuse with a configurable timeout (e.g.,
// one log file per hour).
//
// Log files are still rotated when they reach the configured maximum size limit.
//
// # VModule Usage
//
// In addition to log levels, logging can be selectively enabled or disabled
// using vmodule attributes, similar to glog's vmodule feature. Users must define
// a reusable attribute for each module (typically at global scope) and use it
// with the slog.With function to log module-specific messages.
//
// Example:
//
//	var network = sglog.VModule("network", slog.LevelDebug)
//	...
//	slog.With(network).Info("Network event", ...)
package sglog
