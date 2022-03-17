package spec

import (
	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

type List struct {
	data  []byte
	table listTable
	body  uint32 // body size
	big   bool   // small/big table format
}

// NewList reads and returns list data, but does not validate its elements.
func NewList(b []byte) (List, error) {
	l, _, err := readList(b)
	if err != nil {
		return List{}, err
	}
	return l, nil
}

// ReadList reads and returns list data, and recursively validates its elements.
func ReadList(b []byte) (List, error) {
	l, _, err := readList(b)
	if err != nil {
		return List{}, err
	}
	if err := l.Validate(); err != nil {
		return List{}, err
	}
	return l, nil
}

// Data returns the exact list bytes.
func (l List) Data() []byte {
	return l.data
}

// Count returns the number of elements in the list.
func (l List) Count() int {
	return l.table.count(l.big)
}

// Element returns element data or nil.
func (l List) Element(i int) []byte {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return nil
	case end > int(l.body):
		return nil
	}
	return l.data[start:end]
}

// Validate recursively validates the list.
func (l List) Validate() error {
	n := l.Count()

	for i := 0; i < n; i++ {
		data := l.Element(i)
		if len(data) == 0 {
			continue
		}
		if _, err := ReadData(data); err != nil {
			return err
		}
	}
	return nil
}

// Direct access

func (l List) Bool(i int) bool {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return false
	case end > int(l.body):
		return false
	}

	b := l.data[start:end]
	v, _, _ := readBool(b)
	return v
}

func (l List) Int8(i int) int8 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _, _ := readInt8(b)
	return v
}

func (l List) Int16(i int) int16 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _, _ := readInt16(b)
	return v
}

func (l List) Int32(i int) int32 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _, _ := readInt32(b)
	return v
}

func (l List) Int64(i int) int64 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _, _ := readInt64(b)
	return v
}

func (l List) Uint8(i int) uint8 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _, _ := readUint8(b)
	return v
}

func (l List) Uint16(i int) uint16 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _, _ := readUint16(b)
	return v
}

func (l List) Uint32(i int) uint32 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _, _ := readUint32(b)
	return v
}

func (l List) Uint64(i int) uint64 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _, _ := readUint64(b)
	return v
}

func (l List) U128(i int) u128.U128 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return u128.U128{}
	case end > int(l.body):
		return u128.U128{}
	}

	b := l.data[start:end]
	v, _, _ := readU128(b)
	return v
}

func (l List) U256(i int) u256.U256 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return u256.U256{}
	case end > int(l.body):
		return u256.U256{}
	}

	b := l.data[start:end]
	v, _, _ := readU256(b)
	return v
}

func (l List) Float32(i int) float32 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _, _ := readFloat32(b)
	return v
}

func (l List) Float64(i int) float64 {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return 0
	case end > int(l.body):
		return 0
	}

	b := l.data[start:end]
	v, _, _ := readFloat64(b)
	return v
}

func (l List) Bytes(i int) []byte {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return nil
	case end > int(l.body):
		return nil
	}

	b := l.data[start:end]
	v, _, _ := readBytes(b)
	return v
}

func (l List) String(i int) string {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return ""
	case end > int(l.body):
		return ""
	}

	b := l.data[start:end]
	v, _, _ := readString(b)
	return v
}

func (l List) List(i int) List {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return List{}
	case end > int(l.body):
		return List{}
	}

	b := l.data[start:end]
	v, _ := NewList(b)
	return v
}

func (l List) Message(i int) Message {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return Message{}
	case end > int(l.body):
		return Message{}
	}

	b := l.data[start:end]
	v, _ := NewMessage(b)
	return v
}
