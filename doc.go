// Package sglog provides a log/slog logging handler that writes to multiple
// files based on the severity similar to the Google's glog package.
//
// Most of the code is copied or derived from the Google glog package for Go
// however, there are some differences when compared to the original glog
// package.
//
// # DIFFERENCES
//
//   - The standard log/slog package doesn't offer support for FATAL messages, so
//     they are not supported by this package as well.
//
//   - Globally defined flags are replaced with an Options structure.
//
//   - Thread-ID field in the log file names is always set to zero to enable
//     log file reuse. Thread-ID is still included in the individual log
//     messages, even though it is not very useful in Go.
//
//   - Google's glog package adds a footer message when a log file is rotated,
//     which is not supported in this package.
//
//   - When the log file reuse feature is enabled, log file names do not
//     exactly match the log file creation time. However, timestamps in the log
//     file names still follow the sorted order.
//
// # REUSING LOG FILE NAMES
//
// Google's glog creates a new log file every time the process restarts. This
// can exhaust filesystem inodes when the process is crashing repeatedly. This
// package enables log file reuse with a certain timeout (eg: maximum one log
// file per hour, etc.)
//
// While the reuse timeout option limits number of log files created by time
// duration, this package also rotates the log file when the log file reaches
// the maximum size limit. Timestamps for the log files are chosen to allow for
// identifying the last used log file quickly.
//
// As part of implementing the above feature, we need to remove thread-id from
// the log file name. Since thread-id is not meaningful with Go runtime anyway,
// we decided to replace thread-id in the log file name with a zero.
package sglog
