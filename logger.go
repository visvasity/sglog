package sglog

import (
	"log/slog"
	"sync"
	"time"
)

type Logger struct {
	*Handler

	opts *Options

	mu sync.Mutex
	wg sync.WaitGroup

	fileMap map[slog.Level]*levelFile

	flushChan chan slog.Level
}

func New(opts *Options) *Logger {
	opts.setDefaults()
	l := &Logger{
		opts:      opts,
		fileMap:   make(map[slog.Level]*levelFile),
		flushChan: make(chan slog.Level, 1),
	}
	l.Handler = newHandler(opts, l.emit)
	l.wg.Add(1)
	go l.flushDaemon()
	return l
}

func (v *Logger) Close() {
	close(v.flushChan)
	v.wg.Wait()
}

func (v *Logger) createMissingFiles(uptoLevel slog.Level) error {
	if _, ok := v.fileMap[uptoLevel]; ok {
		return nil
	}
	// Files are created in increasing severity order, so we can be assured that
	// if a high severity logfile exists, then so do all of lower severity.
	for _, l := range v.opts.Levels {
		if _, ok := v.fileMap[l]; ok {
			continue
		}
		if l <= uptoLevel {
			lf := newLevelFile(v.opts, l)
			v.fileMap[l] = lf
		}
	}
	return nil
}

func (v *Logger) emit(level slog.Level, msg []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if err := v.createMissingFiles(level); err != nil {
		return err
	}

	var firstErr error
	for l, f := range v.fileMap {
		if l <= level {
			if _, err := f.Write(msg); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	if level > v.opts.LogBufLevel {
		v.flushChan <- level
	}

	return firstErr
}

func (v *Logger) Flush() error {
	return v.flush(slog.LevelInfo)
}

func (v *Logger) flush(level slog.Level) error {
	var firstErr error
	updateErr := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// Remember where we flushed, so we can call sync without holding
	// the lock.
	var files []*levelFile

	func() {
		v.mu.Lock()
		defer v.mu.Unlock()

		// Flush from fatal down, in case there's trouble flushing.
		for l, lf := range v.fileMap {
			if l >= level {
				updateErr(lf.bio.Flush())
				files = append(files, lf)
			}
		}
	}()

	for _, file := range files {
		updateErr(file.Sync())
	}
	return firstErr
}

func (v *Logger) flushDaemon() {
	defer v.wg.Done()

	tick := time.NewTicker(v.opts.FlushTimeout)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			v.flush(slog.LevelInfo)

		case sev, ok := <-v.flushChan:
			if !ok {
				return
			}
			v.flush(sev)
		}
	}
}
