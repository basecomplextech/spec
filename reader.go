package kproto

import (
	"encoding/binary"
	"math"
)

// Reader is a protocol reader.
type Reader interface {
	// Buffer returns the underlying reader byte buffer.
	Buffer() []byte

	Bool() bool
	Byte() byte

	Int8() int8
	Int16() int16
	Int32() int32
	Int64() int64

	UInt8() uint8
	UInt16() uint16
	UInt32() uint32
	Uint64() uint64

	Float32() float32
	Float64() float64

	Bytes() []byte
	String() string

	List() ListReader
	Map() MapReader
	Struct() StructReader
}

// ListReader reads a list of elements.
type ListReader interface {
	// Bytes returns the underlying list byte buffer.
	Bytes() []byte

	// Element returns an element reader by its index in the list.
	Element(index int) Reader

	// Elements returns the number of elements in the list.
	Elements() int
}

// MapReader reads a map of entries.
type MapReader interface {
	// Bytes returns the underlying map byte buffer.
	Bytes() []byte

	// Find finds and returns an entry index in the map.
	Find(fn func(EntryReader) bool) int

	// Entry returns a map entry reader by its index in the map.
	Entry(index int) EntryReader

	// Entries returns the number of entries in the map.
	Entries() int
}

// EntryReader reads a single map entry.
type EntryReader interface {
	// Key returns a map key reader.
	Key() Reader

	// Value returns a map value reader.
	Value() Reader
}

// StructReader reads a struct of fields.
type StructReader interface {
	// Bytes returns the underlying map byte buffer.
	Bytes() []byte

	// Field returns a field reader by a field number or false when field is absent.
	Field(num uint16) (Reader, bool)

	// FieldAt returns a field reader and its number by a field index in the struct table.
	FieldAt(index int) (Reader, uint16)

	// Fields returns the number of fields in the struct table.
	Fields() int
}

func NewReader(buffer []byte) Reader {
	return newReader(buffer)
}

// implementation

var _ Reader = (*reader)(nil)

type reader struct {
	buffer []byte
	offset int
}

func newReader(buffer []byte) *reader {
	return &reader{buffer: buffer}
}

func (r *reader) Buffer() []byte {
	return r.buffer
}

func (r *reader) Bool() bool {
	b := r.read(1)
	if b[0] == 1 {
		return true
	}
	return false
}

func (r *reader) Byte() byte {
	b := r.read(1)
	return b[0]
}

func (r *reader) Int8() int8 {
	b := r.read(1)
	v := b[0]
	return int8(v)
}

func (r *reader) Int16() int16 {
	b := r.read(2)
	v := binary.BigEndian.Uint16(b)
	return int16(v)
}

func (r *reader) Int32() int32 {
	b := r.read(4)
	v := binary.BigEndian.Uint32(b)
	return int32(v)
}

func (r *reader) Int64() int64 {
	b := r.read(8)
	v := binary.BigEndian.Uint64(b)
	return int64(v)
}

func (r *reader) UInt8() uint8 {
	b := r.read(1)
	return b[0]
}

func (r *reader) UInt16() uint16 {
	b := r.read(2)
	v := binary.BigEndian.Uint16(b)
	return v
}

func (r *reader) UInt32() uint32 {
	b := r.read(4)
	v := binary.BigEndian.Uint32(b)
	return v
}

func (r *reader) Uint64() uint64 {
	b := r.read(8)
	v := binary.BigEndian.Uint64(b)
	return v
}

func (r *reader) Float32() float32 {
	b := r.read(4)
	v := binary.BigEndian.Uint32(b)
	return math.Float32frombits(v)
}

func (r *reader) Float64() float64 {
	b := r.read(8)
	v := binary.BigEndian.Uint64(b)
	return math.Float64frombits(v)
}

func (r *reader) Bytes() []byte {
	// read size
	size := r.UInt32()

	// read bytes
	b := r.read(int(size))
	return b
}

func (r *reader) String() string {
	// read size
	size := r.UInt32()

	// read bytes
	b := r.read(int(size))

	// TODO: Unsafe cast to string
	return string(b)
}

func (r *reader) List() ListReader {
	return nil
}

func (r *reader) Map() MapReader {
	return nil
}

func (r *reader) Struct() StructReader {
	return nil
}

// read increments the offset and returns the next bytes of size.
func (r *reader) read(size int) []byte {
	buf := r.buffer[r.offset : r.offset+size]
	r.offset += size
	return buf
}
