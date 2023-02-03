package writer

import "github.com/complex1tech/spec/go/encoding"

func WriteValue[T any](e *Writer, v T, encode encoding.EncodeFunc[T]) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	if _, err := encode(e.buf, v); err != nil {
		return e.close(err)
	}
	end := e.buf.Len()

	return e.pushData(start, end)
}
