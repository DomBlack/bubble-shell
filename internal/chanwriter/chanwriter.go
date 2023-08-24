package chanwriter

import (
	"sync"
	"time"

	"github.com/cockroachdb/errors"
)

// ChanWriter is a writer that writes to a channel after a debounced period
type ChanWriter struct {
	mu      sync.Mutex
	buf     []byte
	channel chan []byte
	timer   *time.Timer
	closed  bool
}

func New() *ChanWriter {
	return &ChanWriter{
		channel: make(chan []byte),
	}
}

// Write implements io.Writer
func (w *ChanWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, errors.New("closed")
	}

	w.buf = append(w.buf, p...)
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(50*time.Millisecond, w.Flush)

	return len(p), nil
}

// Flush the writer
func (w *ChanWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.closed {
		w.flush()
	}
}

// Close the writer
func (w *ChanWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.flush()

	close(w.channel)
	w.closed = true

	return nil
}

// flush must be called under lock
func (w *ChanWriter) flush() {
	if len(w.buf) > 0 {
		w.channel <- w.buf
		w.buf = nil
	}
}
