package writer

import (
	"github.com/complex1tech/baselibrary/types"
)

// ListWriter writes a list of elements.
type ListWriter struct {
	w *Writer
}

func newListWriter(w *Writer) ListWriter {
	l := ListWriter{w: w}
	l.w.BeginList()
	return l
}

// Build ends the list and returns its bytes.
func (l ListWriter) Build() ([]byte, error) {
	return l.w.End()
}

// End ends the list.
func (l ListWriter) End() error {
	_, err := l.w.End()
	return err
}

// Elements

func (l ListWriter) Bool(v bool) error {
	if err := l.w.Bool(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Byte(v byte) error {
	if err := l.w.Byte(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Int32(v int32) error {
	if err := l.w.Int32(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Int64(v int64) error {
	if err := l.w.Int64(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Uint32(v uint32) error {
	if err := l.w.Uint32(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Uint64(v uint64) error {
	if err := l.w.Uint64(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Float32(v float32) error {
	if err := l.w.Float32(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Float64(v float64) error {
	if err := l.w.Float64(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Bin64(v types.Bin64) error {
	if err := l.w.Bin64(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Bin128(v types.Bin128) error {
	if err := l.w.Bin128(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Bin256(v types.Bin256) error {
	if err := l.w.Bin256(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) Bytes(v []byte) error {
	if err := l.w.Bytes(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) String(v string) error {
	if err := l.w.String(v); err != nil {
		return err
	}
	return l.w.Element()
}

func (l ListWriter) List() ListWriter {
	l.w.BeginElement()
	return l.w.List()
}

func (l ListWriter) Message() MessageWriter {
	l.w.BeginElement()
	return l.w.Message()
}
