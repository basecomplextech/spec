package writer

import "github.com/complex1tech/baselibrary/buffer"

// WriteFunc specifies a generic function to write a value directly into a buffer.
type WriteFunc[T any] func(b buffer.Buffer, value T) (int, error)

// WriteValue writes a generic value using the given write function.
func WriteValue[T any](w Writer, v T, write WriteFunc[T]) error {
	w1 := w.(*writer)
	if w1.err != nil {
		return w1.err
	}

	start := w1.buf.Len()
	if _, err := write(w1.buf, v); err != nil {
		return w1.fail(err)
	}
	end := w1.buf.Len()

	return w1.pushData(start, end)
}
