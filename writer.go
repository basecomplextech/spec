package spec

import (
	"fmt"
	"sync"

	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

const WriteBufferSize = 4096

var writerPool = &sync.Pool{
	New: func() interface{} {
		return EmptyWriter()
	},
}

// Write writes a value.
type Writer struct {
	buf  []byte
	err  error     // writer failed
	data writeData // last written data, must be consumed before writing next data

	objects  objectStack
	elements listStack    // stack of list element tables
	fields   messageStack // stack of message field tables

	// preallocated
	_objects  [16]objectEntry
	_elements [128]listElement
	_fields   [128]messageField
}

// Writable can write itself using a writer.
type Writable interface {
	Write(w *Writer) error
}

// Write writes a writable.
func Write(w Writable) ([]byte, error) {
	return WriteTo(w, nil)
}

// WriteTo writes a writeable to a buffer or allocates a new one when the buffer is too small.
func WriteTo(w Writable, buf []byte) ([]byte, error) {
	wr := writerPool.Get().(*Writer)
	wr.Init(buf)

	defer writerPool.Put(wr)
	defer wr.Reset()

	if err := w.Write(wr); err != nil {
		return nil, err
	}

	return wr.End()
}

// NewWriter returns a new writer with a default buffer.
//
// Usually, it is better to use Write(obj) and WriteTo(obj, buf), than to construct
// a new writer directly. These methods internally use a writer pool.
func NewWriter() *Writer {
	buf := make([]byte, 0, WriteBufferSize)
	return NewWriterBuffer(buf)
}

// NewWriterBuffer returns a new writer with a buffer.
//
// Usually, it is better to use Write(obj) and WriteTo(obj, buf), than to construct
// a new writer directly. These methods internally use a writer pool.
func NewWriterBuffer(buf []byte) *Writer {
	w := &Writer{}

	w.buf = buf[:0]
	w.data = writeData{}

	w.objects.stack = w._objects[:0]
	w.elements.stack = w._elements[:0]
	w.fields.stack = w._fields[:0]
	return w
}

// EmptyWriter return a new writer with an empty buffer.
func EmptyWriter() *Writer {
	return NewWriterBuffer(nil)
}

// End ends writing, returns the result bytes, and resets the writer.
func (w *Writer) End() ([]byte, error) {
	switch {
	case w.err != nil:
		return nil, w.err
	case w.objects.len() > 0:
		return nil, fmt.Errorf("end: incomplete objects, object stack size=%d", w.objects.len())
	}

	// pop data
	data := w.popData()

	// return and reset
	b := w.buf[data.start:data.end]
	w.Reset()
	return b, nil
}

// Init resets the writer and sets its buffer.
func (w *Writer) Init(b []byte) {
	w.Reset()
	w.buf = b[:0]
}

// Reset clears the writer and nils buffer.
func (w *Writer) Reset() {
	w.buf = nil
	w.err = nil
	w.data = writeData{}

	w.objects.reset()
	w.elements.reset()
	w.fields.reset()
}

// Primitive

func (w *Writer) Nil() error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeNil(w.buf)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Bool(v bool) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeBool(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Byte(v byte) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeByte(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Int32(v int32) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeInt32(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Int64(v int64) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeInt64(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Uint32(v uint32) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeUint32(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Uint64(v uint64) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeUint64(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

// U128/U256

func (w *Writer) U128(v u128.U128) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeU128(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) U256(v u256.U256) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeU256(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

// Float

func (w *Writer) Float32(v float32) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeFloat32(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Float64(v float64) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = EncodeFloat64(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

// Bytes/string

func (w *Writer) Bytes(v []byte) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)

	var err error
	w.buf, err = EncodeBytes(w.buf, v)
	if err != nil {
		return w.fail(err)
	}

	end := len(w.buf)
	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) String(v string) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)

	var err error
	w.buf, err = EncodeString(w.buf, v)
	if err != nil {
		return w.fail(err)
	}

	end := len(w.buf)
	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

// List

func (w *Writer) BeginList() error {
	if w.err != nil {
		return w.err
	}

	// push list
	start := len(w.buf)
	tableStart := w.elements.offset()

	w.objects.pushList(start, tableStart)
	return nil
}

func (w *Writer) Element() error {
	if w.err != nil {
		return w.err
	}

	// check list
	list, ok := w.objects.peek()
	switch {
	case !ok:
		return w.fail(fmt.Errorf("element: cannot write element, not list writer"))
	case list.type_ != objectTypeList:
		return w.fail(fmt.Errorf("element: cannot write element, not list writer"))
	}

	// pop data
	data := w.popData()

	// append element relative offset
	offset := uint32(data.end - list.start)
	element := listElement{offset: offset}
	w.elements.push(element)
	return nil
}

func (w *Writer) EndList() error {
	if w.err != nil {
		return w.err
	}

	// pop list
	list, ok := w.objects.pop()
	switch {
	case !ok:
		return w.fail(fmt.Errorf("end list: not list writer"))
	case list.type_ != objectTypeList:
		return w.fail(fmt.Errorf("end list: not list writer"))
	}

	bsize := len(w.buf) - list.start
	table := w.elements.pop(list.tableStart)

	// write list
	var err error
	w.buf, err = encodeList(w.buf, bsize, table)
	if err != nil {
		return w.fail(err)
	}

	// push data entry
	start := list.start
	end := len(w.buf)
	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

// Message

func (w *Writer) BeginMessage() error {
	if w.err != nil {
		return w.err
	}

	// push message
	start := len(w.buf)
	tableStart := w.fields.offset()

	w.objects.pushMessage(start, tableStart)
	return nil
}

func (w *Writer) Field(tag uint16) error {
	if w.err != nil {
		return w.err
	}

	// check message
	message, ok := w.objects.peek()
	switch {
	case !ok:
		return w.fail(fmt.Errorf("field: cannot write field, not message writer"))
	case message.type_ != objectTypeMessage:
		return w.fail(fmt.Errorf("field: cannot write field, not message writer"))
	}

	// pop data
	data := w.popData()

	// insert field tag and relative offset
	f := messageField{
		tag:    tag,
		offset: uint32(data.end - message.start),
	}
	w.fields.insert(message.tableStart, f)
	return nil
}

func (w *Writer) EndMessage() error {
	if w.err != nil {
		return w.err
	}

	// pop message
	message, ok := w.objects.pop()
	switch {
	case !ok:
		return w.fail(fmt.Errorf("end message: not message writer"))
	case message.type_ != objectTypeMessage:
		return w.fail(fmt.Errorf("end message: not message writer"))
	}

	bsize := len(w.buf) - message.start
	table := w.fields.pop(message.tableStart)

	// write message
	var err error
	w.buf, err = encodeMessage(w.buf, bsize, table)
	if err != nil {
		return w.fail(err)
	}

	// push data
	start := message.start
	end := len(w.buf)
	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

// Struct

func (w *Writer) BeginStruct() error {
	if w.err != nil {
		return w.err
	}

	// push struct
	start := len(w.buf)
	w.objects.pushStruct(start)
	return nil
}

func (w *Writer) StructField() error {
	if w.err != nil {
		return w.err
	}

	// check struct
	obj, ok := w.objects.peek()
	switch {
	case !ok:
		return w.fail(fmt.Errorf("field: cannot write struct field, not struct writer"))
	case obj.type_ != objectTypeStruct:
		return w.fail(fmt.Errorf("field: cannot write struct field, not struct writer"))
	}

	// just consume data
	w.popData()
	return nil
}

func (w *Writer) EndStruct() error {
	if w.err != nil {
		return w.err
	}

	// pop struct
	obj, ok := w.objects.pop()
	switch {
	case !ok:
		return w.fail(fmt.Errorf("end struct: not struct writer"))
	case obj.type_ != objectTypeStruct:
		return w.fail(fmt.Errorf("end struct: not struct writer"))
	}

	bsize := len(w.buf) - obj.start

	// write struct
	var err error
	w.buf, err = encodeStruct(w.buf, bsize)
	if err != nil {
		return w.fail(err)
	}

	// push data
	start := obj.start
	end := len(w.buf)
	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

// private

func (w *Writer) fail(err error) error {
	if w.err != nil {
		return err
	}

	w.err = err
	return err
}

// data

// writeData holds the last written data start/end.
// there is no data stack because the data must be consumed immediatelly after it is written.
type writeData struct {
	start int
	end   int
}

func (w *Writer) setData(start, end int) error {
	if w.data.start != 0 || w.data.end != 0 {
		return fmt.Errorf("write: cannot write more data, element/field must be written first")
	}

	w.data = writeData{
		start: start,
		end:   end,
	}
	return nil
}

func (w *Writer) popData() writeData {
	d := w.data
	w.data = writeData{}
	return d
}
