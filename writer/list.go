package writer

import (
	"github.com/complex1tech/baselibrary/types"
)

// ListWriter writes a list of elements.
type ListWriter struct {
	w *writer
}

// Build ends the list and returns its bytes.
func (l ListWriter) Build() ([]byte, error) {
	return l.w.end()
}

// End ends the list.
func (l ListWriter) End() error {
	_, err := l.w.end()
	return err
}

// Elements

func (l ListWriter) Bool(v bool) error {
	if err := l.w.Value().Bool(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Byte(v byte) error {
	if err := l.w.Value().Byte(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Int32(v int32) error {
	if err := l.w.Value().Int32(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Int64(v int64) error {
	if err := l.w.Value().Int64(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Uint32(v uint32) error {
	if err := l.w.Value().Uint32(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Uint64(v uint64) error {
	if err := l.w.Value().Uint64(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Float32(v float32) error {
	if err := l.w.Value().Float32(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Float64(v float64) error {
	if err := l.w.Value().Float64(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Bin64(v types.Bin64) error {
	if err := l.w.Value().Bin64(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Bin128(v types.Bin128) error {
	if err := l.w.Value().Bin128(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Bin256(v types.Bin256) error {
	if err := l.w.Value().Bin256(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Bytes(v []byte) error {
	if err := l.w.Value().Bytes(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) String(v string) error {
	if err := l.w.Value().String(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) List() ListWriter {
	l.w.beginElement()
	return l.w.List()
}

func (l ListWriter) Message() MessageWriter {
	l.w.beginElement()
	return l.w.Message()
}