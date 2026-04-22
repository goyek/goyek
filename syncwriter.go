package goyek

import (
	"io"
	"sync"
)

type syncWriter struct {
	io.Writer
	mtx sync.Mutex
}

func (w *syncWriter) Write(p []byte) (int, error) {
	defer func() { w.mtx.Unlock() }()
	w.mtx.Lock()
	return w.Writer.Write(p)
}
