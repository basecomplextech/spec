// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package writer

import "github.com/basecomplextech/baselibrary/bin"

// ListWriter writes a list of elements.
type ListWriter struct {
	w *writer
}

// Err returns the current write error.
func (l ListWriter) Err() error {
	return l.w.err
}

// Len returns the number of written elements.
// The method is only valid when there is no pending element.
func (l ListWriter) Len() int {
	return l.w.listLen()
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

// WriteElement writes a generic element using the given write function.
func WriteElement[T any](w ListWriter, value T, write WriteFunc[T]) error {
	if err := WriteValue(w.w, value, write); err != nil {
		return err
	}
	return w.w.element()
}

// Elements

func (l ListWriter) Any(v []byte) error {
	if err := l.w.Value().Any(v); err != nil {
		return err
	}
	return l.w.element()
}

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

// Int

func (l ListWriter) Int16(v int16) error {
	if err := l.w.Value().Int16(v); err != nil {
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

// Uint

func (l ListWriter) Uint16(v uint16) error {
	if err := l.w.Value().Uint16(v); err != nil {
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

// Float

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

// Bin

func (l ListWriter) Bin64(v bin.Bin64) error {
	if err := l.w.Value().Bin64(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Bin128(v bin.Bin128) error {
	if err := l.w.Value().Bin128(v); err != nil {
		return err
	}
	return l.w.element()
}

func (l ListWriter) Bin256(v bin.Bin256) error {
	if err := l.w.Value().Bin256(v); err != nil {
		return err
	}
	return l.w.element()
}

// Bytes/string

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

// List/message

func (l ListWriter) List() ListWriter {
	l.w.beginElement()
	return l.w.List()
}

func (l ListWriter) Message() MessageWriter {
	l.w.beginElement()
	return l.w.Message()
}
