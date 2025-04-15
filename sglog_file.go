// Go support for leveled logs, analogous to https://github.com/google/glog.
//
// Copyright 2023 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// File I/O for logs.

package sglog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	pid      = os.Getpid()
	program  = filepath.Base(os.Args[0])
	host     = "unknownhost"
	userName = "unknownuser"
)

func init() {
	h, err := os.Hostname()
	if err == nil {
		host = shortHostname(h)
	}

	if u := lookupUser(); u != "" {
		userName = u
	}
	// Sanitize userName since it is used to construct file paths.
	userName = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		default:
			return '_'
		}
		return r
	}, userName)
}

// shortHostname returns its argument, truncating at the first period.
// For instance, given "www.google.com" it returns "www".
func shortHostname(hostname string) string {
	if i := strings.Index(hostname, "."); i >= 0 {
		return hostname[:i]
	}
	return hostname
}

type levelFile struct {
	backend *Backend

	level slog.Level

	file   *os.File
	nbytes uint64

	names []string
}

func (v *Backend) newLevelFile(opts *Options, level slog.Level) *levelFile {
	return &levelFile{
		backend: v,
		level:   level,
	}
}

func (f *levelFile) Write(p []byte) (int, error) {
	if f.file == nil || f.nbytes >= f.backend.opts.LogFileMaxSize {
		if err := f.rotateFile(time.Now()); err != nil {
			return 0, fmt.Errorf("could not create/rotate log file: %w", err)
		}
	}

	for nwrote := 0; nwrote < len(p); {
		n, err := f.file.Write(p)
		nwrote += n
		f.nbytes += uint64(n)

		if err != nil {
			if errors.Is(err, io.ErrShortWrite) {
				continue
			}
			return nwrote, err
		}
	}

	return len(p), nil
}

func (f *levelFile) Sync() error {
	return f.file.Sync()
}

func (f *levelFile) levelName() string {
	return f.level.String()
}

func (f *levelFile) fileName(t time.Time) string {
	return fmt.Sprintf("%s.%s.%s.log.%s.%04d%02d%02d-%02d%02d%02d.%d",
		program,
		host,
		userName,
		f.levelName(),
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		0 /* zero pid in the filename */)
}

func (f *levelFile) linkName(t time.Time) string {
	return program + "." + f.levelName()
}

func (f *levelFile) filePath(dir string, t time.Time) (string, uint64, error) {
	var fpaths []string
	for d := time.Second; d < f.backend.opts.ReuseFileDuration; d = d * 2 {
		fpath := filepath.Join(dir, f.fileName(t.Truncate(d)))
		fpaths = append(fpaths, fpath)
	}
	final := filepath.Join(dir, f.fileName(t.Truncate(f.backend.opts.ReuseFileDuration)))
	fpaths = append(fpaths, final)

	for i := 1; i < len(fpaths); i++ {
		fstat, err := os.Stat(fpaths[i])
		if err == nil {
			if fsize := uint64(fstat.Size()); fsize < f.backend.opts.LogFileMaxSize {
				return fpaths[i], fsize, nil
			}
			return fpaths[i-1], 0, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", 0, err
		}
	}
	return final, 0, nil
}

func (f *levelFile) createFile(t time.Time) (fp *os.File, filename string, err error) {
	link := f.linkName(t)

	var lastErr error
	for _, dir := range f.backend.opts.LogDirs {
		fpath, fsize, err := f.filePath(dir, t)
		if err != nil {
			lastErr = err
			continue
		}
		flags := os.O_WRONLY | os.O_CREATE
		fp, err := os.OpenFile(fpath, flags, f.backend.opts.LogFileMode)
		if err != nil {
			lastErr = err
			continue
		}
		if _, err := fp.Seek(0, os.SEEK_END); err != nil {
			lastErr = err
			if err := fp.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "could not close file (ignored): %v\n", err)
			}
			continue
		}
		f.nbytes = uint64(fsize)

		{
			fname := filepath.Base(fpath)
			symlink := filepath.Join(dir, link)
			if err := os.Remove(symlink); err != nil && !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "could not remove symlink %q (ignored): %v\n", symlink, err)
			}
			if err := os.Symlink(fname, symlink); err != nil {
				fmt.Fprintf(os.Stderr, "could not create symlink %q->%q (ignored): %v\n", symlink, fname, err)
			}

			if f.backend.opts.LogLinkDir != "" {
				lsymlink := filepath.Join(f.backend.opts.LogLinkDir, link)
				if err := os.Remove(lsymlink); err != nil && !errors.Is(err, os.ErrNotExist) {
					fmt.Fprintf(os.Stderr, "could not remove symlink %q (ignored): %v\n", lsymlink, err)
				}
				if err := os.Symlink(fname, lsymlink); err != nil {
					fmt.Fprintf(os.Stderr, "could not create symlink %q->%q (ignroed): %v\n", lsymlink, fname, err)
				}
			}
		}
		return fp, fpath, nil
	}
	return nil, "", fmt.Errorf("log: cannot create log: %w", lastErr)
}

func (f *levelFile) rotateFile(now time.Time) error {
	var err error
	pn := "<none>"
	file, name, err := f.createFile(now)
	if err != nil {
		return err
	}

	if f.file != nil {
		// The current log file becomes the previous log at the end of this block,
		// so save its name for use in the header of the next file.
		pn = f.file.Name()
		if err := f.file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "could not close file (ignored): %v", err)
		}
	}

	f.file = file
	f.names = append(f.names, name)

	if f.backend.opts.LogFileHeader {
		// Write header.
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "Log file created at: %s\n", now.Format("2006/01/02 15:04:05"))
		fmt.Fprintf(&buf, "Running on machine: %s\n", host)
		fmt.Fprintf(&buf, "Binary: Built with %s %s for %s/%s\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		fmt.Fprintf(&buf, "Previous log: %s\n", pn)
		fmt.Fprintf(&buf, "Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg\n")
		if _, err := f.Write(buf.Bytes()); err != nil {
			return fmt.Errorf("could not write log file header: %w", err)
		}
	}
	return nil
}
