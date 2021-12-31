package spec

import (
	"fmt"

	"github.com/baseone-run/library/u128"
	"github.com/baseone-run/library/u256"
)

// ListReader reads a list from a byte slice.
type ListReader struct {
	l list
}

// NewListReader parses and returns list data, but does not validate it.
func NewListReader(b []byte) (ListReader, error) {
	l, _, err := readList(b)
	if err != nil {
		return ListReader{}, err
	}
	return ListReader{l}, nil
}

// Reflect access

// Len returns the number of elements in the list.
func (r ListReader) Len() int {
	return r.l.len()
}

// Read returns an element data by an index.
func (r ListReader) Read(i int) ([]byte, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return nil, nil
	case end > int(r.l.body):
		return nil, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	return b, nil
}

// Direct access

func (r ListReader) ReadBool(i int) (bool, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return false, nil
	case end > int(r.l.body):
		return false, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readBool(b)
	return v, err
}

func (r ListReader) ReadInt8(i int) (int8, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return 0, nil
	case end > int(r.l.body):
		return 0, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readInt8(b)
	return v, err
}

func (r ListReader) ReadInt16(i int) (int16, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return 0, nil
	case end > int(r.l.body):
		return 0, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readInt16(b)
	return v, err
}

func (r ListReader) ReadInt32(i int) (int32, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return 0, nil
	case end > int(r.l.body):
		return 0, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readInt32(b)
	return v, err
}

func (r ListReader) ReadInt64(i int) (int64, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return 0, nil
	case end > int(r.l.body):
		return 0, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readInt64(b)
	return v, err
}

func (r ListReader) ReadUint8(i int) (uint8, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return 0, nil
	case end > int(r.l.body):
		return 0, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readUint8(b)
	return v, err
}

func (r ListReader) ReadUint16(i int) (uint16, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return 0, nil
	case end > int(r.l.body):
		return 0, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readUint16(b)
	return v, err
}

func (r ListReader) ReadUint32(i int) (uint32, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return 0, nil
	case end > int(r.l.body):
		return 0, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readUint32(b)
	return v, err
}

func (r ListReader) ReadUint64(i int) (uint64, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return 0, nil
	case end > int(r.l.body):
		return 0, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readUint64(b)
	return v, err
}

func (r ListReader) ReadU128(i int) (u128.U128, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return u128.U128{}, nil
	case end > int(r.l.body):
		return u128.U128{}, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readU128(b)
	return v, err
}

func (r ListReader) ReadU256(i int) (u256.U256, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return u256.U256{}, nil
	case end > int(r.l.body):
		return u256.U256{}, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readU256(b)
	return v, err
}

func (r ListReader) ReadFloat32(i int) (float32, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return 0, nil
	case end > int(r.l.body):
		return 0, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readFloat32(b)
	return v, err
}

func (r ListReader) ReadFloat64(i int) (float64, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return 0, nil
	case end > int(r.l.body):
		return 0, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readFloat64(b)
	return v, err
}

func (r ListReader) ReadBytes(i int) ([]byte, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return nil, nil
	case end > int(r.l.body):
		return nil, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readBytes(b)
	return v, err
}

func (r ListReader) ReadString(i int) (string, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return "", nil
	case end > int(r.l.body):
		return "", r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	v, _, err := readString(b)
	return v, err
}

func (r ListReader) ReadList(i int) (ListReader, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return ListReader{}, nil
	case end > int(r.l.body):
		return ListReader{}, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	return NewListReader(b)
}

func (r ListReader) ReadMessage(i int) (MessageReader, error) {
	start, end := r.l.table.offset(r.l.big, i)
	switch {
	case start < 0:
		return MessageReader{}, nil
	case end > int(r.l.body):
		return MessageReader{}, r.rangeError(i, end)
	}

	b := r.l.data[start:end]
	return NewMessageReader(b)
}

// private

func (r ListReader) rangeError(index int, end int) error {
	return fmt.Errorf("element offset out of range: element=%d, offset=%d, body=%d",
		index, end, r.l.body)
}
