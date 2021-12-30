package spec

import (
	"fmt"

	"github.com/baseone-run/library/u128"
	"github.com/baseone-run/library/u256"
)

// MessageReader reads a message from a byte slice.
type MessageReader struct {
	m message
}

// NewMessageReader returns a new message reader or an error.
func NewMessageReader(b []byte) (MessageReader, error) {
	m, err := readMessage(b)
	if err != nil {
		return MessageReader{}, err
	}
	return MessageReader{m}, nil
}

// Reflect access

// Len returns the number of fields in the message.
func (r MessageReader) Len() int {
	return r.m.len()
}

// Read returns a field data by a tag or nil.
// The field is at data end, but data slice can be larger than field.
func (r MessageReader) Read(tag uint16) ([]byte, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return nil, nil
	case end > int(r.m.body):
		return nil, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return b, nil
}

// ReadByIndex returns a field data by an index or nil.
// The field is at data end, but data slice can be larger than field.
func (r MessageReader) ReadByIndex(i int) ([]byte, error) {
	end := r.m.table.offsetByIndex(r.m.big, i)
	switch {
	case end < 0:
		return nil, nil
	case end > int(r.m.body):
		return nil, r.indexError(i, end)
	}

	b := r.m.data[:end]
	return b, nil
}

// Direct access

func (r MessageReader) ReadBool(tag uint16) (bool, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return false, nil
	case end > int(r.m.body):
		return false, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readBool(b)
}

func (r MessageReader) ReadInt8(tag uint16) (int8, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return 0, nil
	case end > int(r.m.body):
		return 0, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readInt8(b)
}

func (r MessageReader) ReadInt16(tag uint16) (int16, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return 0, nil
	case end > int(r.m.body):
		return 0, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readInt16(b)
}

func (r MessageReader) ReadInt32(tag uint16) (int32, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return 0, nil
	case end > int(r.m.body):
		return 0, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readInt32(b)
}

func (r MessageReader) ReadInt64(tag uint16) (int64, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return 0, nil
	case end > int(r.m.body):
		return 0, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readInt64(b)
}

func (r MessageReader) ReadUint8(tag uint16) (uint8, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return 0, nil
	case end > int(r.m.body):
		return 0, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readUint8(b)
}

func (r MessageReader) ReadUint16(tag uint16) (uint16, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return 0, nil
	case end > int(r.m.body):
		return 0, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readUint16(b)
}

func (r MessageReader) ReadUint32(tag uint16) (uint32, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return 0, nil
	case end > int(r.m.body):
		return 0, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readUint32(b)
}

func (r MessageReader) ReadUint64(tag uint16) (uint64, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return 0, nil
	case end > int(r.m.body):
		return 0, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readUint64(b)
}

func (r MessageReader) ReadU128(tag uint16) (u128.U128, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return u128.U128{}, nil
	case end > int(r.m.body):
		return u128.U128{}, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readU128(b)
}

func (r MessageReader) ReadU256(tag uint16) (u256.U256, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return u256.U256{}, nil
	case end > int(r.m.body):
		return u256.U256{}, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readU256(b)
}

func (r MessageReader) ReadFloat32(tag uint16) (float32, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return 0, nil
	case end > int(r.m.body):
		return 0, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readFloat32(b)
}

func (r MessageReader) ReadFloat64(tag uint16) (float64, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return 0, nil
	case end > int(r.m.body):
		return 0, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readFloat64(b)
}

func (r MessageReader) ReadBytes(tag uint16) ([]byte, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return nil, nil
	case end > int(r.m.body):
		return nil, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readBytes(b)
}

func (r MessageReader) ReadString(tag uint16) (string, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return "", nil
	case end > int(r.m.body):
		return "", r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return readString(b)
}

func (r MessageReader) ReadList(tag uint16) (ListReader, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return ListReader{}, nil
	case end > int(r.m.body):
		return ListReader{}, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return NewListReader(b)
}

func (r MessageReader) ReadMessage(tag uint16) (MessageReader, error) {
	end := r.m.table.offset(r.m.big, tag)
	switch {
	case end < 0:
		return MessageReader{}, nil
	case end > int(r.m.body):
		return MessageReader{}, r.rangeError(tag, end)
	}

	b := r.m.data[:end]
	return NewMessageReader(b)
}

// private

func (r MessageReader) indexError(index int, end int) error {
	return fmt.Errorf("field offset out of range: field index=%d, offset=%d, body=%d",
		index, end, r.m.body)
}

func (r MessageReader) rangeError(tag uint16, end int) error {
	return fmt.Errorf("field offset out of range: field=%d, offset=%d, body=%d", tag, end, r.m.body)
}
