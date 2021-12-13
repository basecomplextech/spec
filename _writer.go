package protocol

import (
	"encoding/binary"
	"math"
)

const WriteBufferSize = 4096

// Writer is a protocol writer.1
type Writer interface {
	End() []byte
	Buffer() []byte

	Bool(v bool)
	Byte(v byte)

	Int8(v int8)
	Int16(v int16)
	Int32(v int32)
	Int64(v int64)

	UInt8(v uint8)
	UInt16(v uint16)
	UInt32(v uint32)
	UInt64(v uint64)

	Float32(v float32)
	Float64(v float64)

	Bytes(v []byte)
	String(v string)

	Map() MapWriter
	List() ListWriter
	Struct() StructWriter
}

// ListWriter writes a list of elements.
type ListWriter interface {
	// Next writes the next element.
	Next() Writer

	// End writes the list end.
	End()
}

// MapWriter writes a map of entries.
type MapWriter interface {
	// Next writes the next entry.
	Next() EntryWriter

	// End writes the map end.
	End()
}

// EntryWriter writes a map entry.
type EntryWriter interface {
	// Key writes the entry key.
	Key() Writer

	// Value writes the entry value.
	Value() Writer
}

// StructWriter writes a struct of fields.
type StructWriter interface {
	// Field writes a struct field.
	Field(num uint16) Writer

	// End writes the struct end.
	End()
}

func NewWriter() Writer {
	return newWriter()
}

// implementation

var _ Writer = (*writer)(nil)

type writer struct {
	buffer []byte
	offset int
}

func newWriter() *writer {
	return &writer{
		buffer: make([]byte, 0, 4096),
	}
}

func (w *writer) End() []byte {
	return w.buffer
}

func (w *writer) Buffer() []byte {
	return w.buffer
}

func (w *writer) Bool(v bool) {
	b := w.write(1)
	if v {
		b[0] = 1
	} else {
		b[0] = 0
	}
}

func (w *writer) Byte(v byte) {
	b := w.write(1)
	b[0] = v
}

func (w *writer) Int8(v int8) {
	b := w.write(1)
	b[0] = byte(v)
}

func (w *writer) Int16(v int16) {
	b := w.write(2)
	binary.BigEndian.PutUint16(b, uint16(v))
}

func (w *writer) Int32(v int32) {
	b := w.write(4)
	binary.BigEndian.PutUint32(b, uint32(v))
}

func (w *writer) Int64(v int64) {
	b := w.write(8)
	binary.BigEndian.PutUint64(b, uint64(v))
}

func (w *writer) UInt8(v uint8) {
	b := w.write(1)
	b[0] = v
}

func (w *writer) UInt16(v uint16) {
	b := w.write(2)
	binary.BigEndian.PutUint16(b, v)
}

func (w *writer) UInt32(v uint32) {
	b := w.write(4)
	binary.BigEndian.PutUint32(b, v)
}

func (w *writer) UInt64(v uint64) {
	b := w.write(8)
	binary.BigEndian.PutUint64(b, v)
}

func (w *writer) Float32(v float32) {
	b := w.write(4)
	binary.BigEndian.PutUint32(b, math.Float32bits(v))
}

func (w *writer) Float64(v float64) {
	b := w.write(8)
	binary.BigEndian.PutUint64(b, math.Float64bits(v))
}

func (w *writer) Bytes(v []byte) {
	size := len(v)
	if size > math.MaxUint32 {
		panic("bytes too large")
	}

	// write size
	w.UInt32(uint32(size))

	// write bytes
	b := w.write(size)
	copy(b, v)
}

func (w *writer) String(s string) {
	v := []byte(s)
	size := len(v)
	if size > math.MaxUint32 {
		panic("string too large")
	}

	// write size
	w.UInt32(uint32(size))

	// write bytes
	b := w.write(size)
	copy(b, v)
}

func (w *writer) Map() MapWriter {
	return nil
}

func (w *writer) List() ListWriter {
	return nil
}

func (w *writer) Struct() StructWriter {
	return nil
}

// write grows the underlying buffer and returns a write buffer of size.
func (w *writer) write(size int) []byte {
	// maybe realloc
	rem := cap(w.buffer) - len(w.buffer)
	if rem < size {
		nextSize := cap(w.buffer) * 2
		if nextSize == 0 {
			nextSize = WriteBufferSize
		}

		next := make([]byte, len(w.buffer), nextSize)
		copy(next, w.buffer)
		w.buffer = next
	}

	// compute next length
	length := len(w.buffer) + size

	// get next writer buffer
	buf := w.buffer[len(w.buffer):length]

	// grow underlying buffer
	w.buffer = w.buffer[:length]
	return buf
}
