package spec

import "sync"

var writerPool = &sync.Pool{
	New: func() interface{} {
		return EmptyWriter()
	},
}

// Writable can write itself using a writer.
type Writable interface {
	Write(w *Writer) error
}

// Write writes a writable.
func Write(w Writable) ([]byte, error) {
	return WriteTo(w, nil)
}

// WriteTo writes a writeable to a buffer or allocates a new one when the buffer is too small.
func WriteTo(w Writable, buf []byte) ([]byte, error) {
	wr := writerPool.Get().(*Writer)
	wr.Init(buf)

	defer writerPool.Put(wr)
	defer wr.Reset()

	if err := w.Write(wr); err != nil {
		return nil, err
	}

	return wr.End()
}
