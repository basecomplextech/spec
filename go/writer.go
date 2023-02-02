package spec

import (
	"errors"
	"fmt"

	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/baselibrary/types"
)

var writerClosed = errors.New("operation on a closed writer")

// Writer writes spec elements.
// It is not reusable, but small enough to have negligible effect on memory allocation.
type Writer struct {
	*writerState
	err error
}

// NewWriter returns a new writer with an empty buffer.
func NewWriter() *Writer {
	buf := buffer.New()
	return newWriter(buf)
}

// NewWriterBuffer returns a new writer with a buffer.
func NewWriterBuffer(buf buffer.Buffer) *Writer {
	return newWriter(buf)
}

func newWriter(buf buffer.Buffer) *Writer {
	s := getWriterState()
	s.init(buf)

	return &Writer{writerState: s}
}

// close closes the writer and releases its state.
func (e *Writer) close(err error) error {
	if e.err != nil {
		return e.err
	}

	if err != nil {
		e.err = err
	} else {
		e.err = writerClosed
	}

	s := e.writerState
	e.writerState = nil

	releaseWriterState(s)
	return err
}

func (e *Writer) closef(format string, args ...any) error {
	var err error
	if len(args) == 0 {
		err = errors.New(format)
	} else {
		err = fmt.Errorf(format, args...)
	}
	return e.close(err)
}

// Reset resets the writer state with the buffer.
func (e *Writer) Reset(buf buffer.Buffer) {
	e.close(nil)
	e.err = nil

	if buf == nil {
		buf = buffer.New()
	}

	s := getWriterState()
	s.init(buf)
	e.writerState = s
}

// End ends a nested object and a parent field/element if present.
func (e *Writer) End() (result []byte, err error) {
	if e.err != nil {
		return nil, e.err
	}

	// end top object
	entry, ok := e.stack.peek()
	if !ok {
		return nil, e.closef("end: encode stack is empty")
	}

	switch entry.type_ {
	case entryList:
		result, err = e.endList()
		if err != nil {
			return nil, err
		}

	case entryMessage:
		result, err = e.endMessage()
		if err != nil {
			return nil, err
		}

	default:
		return nil, e.closef("end: not list or message")
	}

	// end parent field/element
	entry, ok = e.stack.peek()
	if !ok {
		return result, e.close(nil)
	}

	switch entry.type_ {
	case entryElement:
		return e.endElement()
	case entryField:
		return e.endField()
	}
	return result, nil
}

// Primitive

func (e *Writer) Bool(v bool) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeBool(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Writer) Byte(v byte) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeByte(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Writer) Int32(v int32) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeInt32(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Writer) Int64(v int64) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeInt64(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Writer) Uint32(v uint32) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeUint32(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Writer) Uint64(v uint64) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeUint64(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

// Bin64/128/256

func (e *Writer) Bin64(v types.Bin64) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeBin64(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Writer) Bin128(v types.Bin128) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeBin128(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Writer) Bin256(v types.Bin256) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeBin256(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

// Float

func (e *Writer) Float32(v float32) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeFloat32(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Writer) Float64(v float64) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeFloat64(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

// Bytes/string

func (e *Writer) Bytes(v []byte) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	if _, err := EncodeBytes(e.buf, v); err != nil {
		return e.close(err)
	}
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Writer) String(v string) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	if _, err := EncodeString(e.buf, v); err != nil {
		return e.close(err)
	}
	end := e.buf.Len()

	return e.setData(start, end)
}

// List

func (e *Writer) BeginList() error {
	if e.err != nil {
		return e.err
	}

	// push list
	start := e.buf.Len()
	tableStart := e.elements.offset()

	e.stack.pushList(start, tableStart)
	return nil
}

func (e *Writer) BeginElement() error {
	if e.err != nil {
		return e.err
	}

	// check list
	list, ok := e.stack.peek()
	switch {
	case !ok:
		return e.closef("begin element: cannot begin element, parent not list")
	case list.type_ != entryList:
		return e.closef("begin element: cannot begin element, parent not list")
	}

	// push list element
	start := e.buf.Len()
	e.stack.pushElement(start)
	return nil
}

func (e *Writer) Element() error {
	if e.err != nil {
		return e.err
	}

	// check list
	list, ok := e.stack.peek()
	switch {
	case !ok:
		return e.closef("element: cannot encode element, parent not list")
	case list.type_ != entryList:
		return e.closef("element: cannot encode element, parent not list")
	}

	// pop data
	data := e.popData()

	// append element relative offset
	offset := uint32(data.end - list.start)
	element := listElement{offset: offset}
	e.elements.push(element)
	return nil
}

func (e *Writer) ListLen() int {
	if e.err != nil {
		return 0
	}

	// check list
	list, ok := e.stack.peek()
	switch {
	case !ok:
		return 0
	case list.type_ != entryList:
		return 0
	}

	start := list.start
	return e.elements.len(start)
}

func (e *Writer) endElement() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}

	// check element
	elem, ok := e.stack.pop()
	switch {
	case !ok:
		return nil, e.closef("end element: not element")
	case elem.type_ != entryElement:
		return nil, e.closef("end element: not element")
	}

	// check list
	list, ok := e.stack.peek()
	switch {
	case !ok:
		return nil, e.closef("end element: parent not list")
	case list.type_ != entryList:
		return nil, e.closef("end element: parent not list")
	}

	// pop data
	data := e.popData()

	// append element relative offset
	offset := uint32(data.end - list.start)
	element := listElement{offset: offset}
	e.elements.push(element)

	// return data
	b := e.buf.Bytes()
	b = b[elem.start:data.end]
	return b, nil
}

func (e *Writer) endList() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}

	// pop list
	list, ok := e.stack.pop()
	switch {
	case !ok:
		return nil, e.closef("end list: not list")
	case list.type_ != entryList:
		return nil, e.closef("end list: not list")
	}

	bodySize := e.buf.Len() - list.start
	table := e.elements.pop(list.tableStart)

	// encode list
	if _, err := encodeListMeta(e.buf, bodySize, table); err != nil {
		return nil, e.close(err)
	}

	// push data entry
	start := list.start
	end := e.buf.Len()
	if err := e.setData(start, end); err != nil {
		return nil, err
	}

	// return data
	b := e.buf.Bytes()
	b = b[start:end]
	return b, nil
}

// Message

func (e *Writer) BeginMessage() error {
	if e.err != nil {
		return e.err
	}

	// push message
	start := e.buf.Len()
	tableStart := e.fields.offset()

	e.stack.pushMessage(start, tableStart)
	return nil
}

func (e *Writer) BeginField(tag uint16) error {
	if e.err != nil {
		return e.err
	}

	// check message
	message, ok := e.stack.peek()
	switch {
	case !ok:
		return e.closef("begin field: cannot begin field, parent not message")
	case message.type_ != entryMessage:
		return e.closef("begin field: cannot begin field, parent not message")
	}

	// push field
	start := e.buf.Len()
	e.stack.pushField(start, tag)
	return nil
}

func (e *Writer) Field(tag uint16) error {
	if e.err != nil {
		return e.err
	}

	// check message
	message, ok := e.stack.peek()
	switch {
	case !ok:
		return e.closef("field: cannot encode field, parent not message")
	case message.type_ != entryMessage:
		return e.closef("field: cannot encode field, parent not message")
	}

	// pop data
	data := e.popData()

	// insert field tag and relative offset
	f := messageField{
		tag:    tag,
		offset: uint32(data.end - message.start),
	}
	e.fields.insert(message.tableStart, f)
	return nil
}

func (e *Writer) FieldBytes(tag uint16, data []byte) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	e.buf.Write(data)
	end := e.buf.Len()

	if err := e.setData(start, end); err != nil {
		return err
	}
	return e.Field(tag)
}

func (e *Writer) HasField(tag uint16) bool {
	if e.err != nil {
		return false
	}

	// peek message
	message, ok := e.stack.peek()
	switch {
	case !ok:
		return false
	case message.type_ != entryMessage:
		return false
	}

	// check field table
	offset := message.tableStart
	return e.fields.hasField(offset, tag)
}

func (e *Writer) endField() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}

	// check field
	field, ok := e.stack.pop()
	switch {
	case !ok:
		return nil, e.closef("end field: not field")
	case field.type_ != entryField:
		return nil, e.closef("end field: not field")
	}
	tag := field.tag()

	// check message
	message, ok := e.stack.peek()
	switch {
	case !ok:
		return nil, e.closef("field: cannot encode field, parent not message")
	case message.type_ != entryMessage:
		return nil, e.closef("field: cannot encode field, parent not message")
	}

	// pop data
	data := e.popData()

	// insert field with tag and relative offset
	f := messageField{
		tag:    tag,
		offset: uint32(data.end - message.start),
	}
	e.fields.insert(message.tableStart, f)

	// return data
	b := e.buf.Bytes()
	b = b[field.start:data.end]
	return b, nil
}

func (e *Writer) endMessage() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}

	// pop message
	message, ok := e.stack.pop()
	switch {
	case !ok:
		return nil, e.closef("end message: parent not message")
	case message.type_ != entryMessage:
		return nil, e.closef("end message: parent not message")
	}

	dataSize := e.buf.Len() - message.start
	table := e.fields.pop(message.tableStart)

	// encode message
	if _, err := encodeMessageMeta(e.buf, dataSize, table); err != nil {
		return nil, e.close(err)
	}

	// push data
	start := message.start
	end := e.buf.Len()
	if err := e.setData(start, end); err != nil {
		return nil, err
	}

	// return data
	b := e.buf.Bytes()
	b = b[start:end]
	return b, nil
}

// Value

func EncodeValue[T any](e *Writer, v T, encode EncodeFunc[T]) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	if _, err := encode(e.buf, v); err != nil {
		return e.close(err)
	}
	end := e.buf.Len()

	return e.setData(start, end)
}

// data

// encodeData holds the last written data start/end.
// there is no data stack because the data must be consumed immediatelly after it is written.
type encodeData struct {
	start int
	end   int
}

// TODO: Rename into pushData and move to stack.
func (e *Writer) setData(start, end int) error {
	if e.data.start != 0 || e.data.end != 0 {
		return e.closef("encode: cannot encode more data, element/field must be written first")
	}

	e.data = encodeData{
		start: start,
		end:   end,
	}
	return nil
}

func (e *Writer) popData() encodeData {
	d := e.data
	e.data = encodeData{}
	return d
}
