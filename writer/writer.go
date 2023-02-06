package writer

import (
	"errors"
	"fmt"

	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/baselibrary/types"
	"github.com/complex1tech/spec/encoding"
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
func (w *Writer) close(err error) error {
	if w.err != nil {
		return w.err
	}

	if err != nil {
		w.err = err
	} else {
		w.err = writerClosed
	}

	s := w.writerState
	w.writerState = nil

	releaseWriterState(s)
	return err
}

func (w *Writer) closef(format string, args ...any) error {
	var err error
	if len(args) == 0 {
		err = errors.New(format)
	} else {
		err = fmt.Errorf(format, args...)
	}
	return w.close(err)
}

// Reset resets the writer state with the buffer.
func (w *Writer) Reset(buf buffer.Buffer) {
	w.close(nil)
	w.err = nil

	if buf == nil {
		buf = buffer.New()
	}

	s := getWriterState()
	s.init(buf)
	w.writerState = s
}

// Err returns an error or nil.
func (w *Writer) Err() error {
	return w.err
}

// End ends a nested object and a parent field/element if present.
func (w *Writer) End() (result []byte, err error) {
	if w.err != nil {
		return nil, w.err
	}

	// end top object
	entry, ok := w.stack.peek()
	if !ok {
		return nil, w.closef("end: encode stack is empty")
	}

	switch entry.type_ {
	case entryList:
		result, err = w.endList()
		if err != nil {
			return nil, err
		}

	case entryMessage:
		result, err = w.endMessage()
		if err != nil {
			return nil, err
		}

	default:
		return nil, w.closef("end: not list or message")
	}

	// maybe end parent field/element
	entry, ok = w.stack.peekSecondLast()
	if !ok {
		return result, w.close(nil)
	}

	switch entry.type_ {
	case entryElement:
		return w.endElement()
	case entryField:
		return w.endField()
	}
	return result, nil
}

// Primitive

func (w *Writer) Bool(v bool) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeBool(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

func (w *Writer) Byte(v byte) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeByte(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

func (w *Writer) Int32(v int32) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeInt32(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

func (w *Writer) Int64(v int64) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeInt64(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

func (w *Writer) Uint32(v uint32) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeUint32(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

func (w *Writer) Uint64(v uint64) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeUint64(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

// Bin64/128/256

func (w *Writer) Bin64(v types.Bin64) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeBin64(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

func (w *Writer) Bin128(v types.Bin128) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeBin128(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

func (w *Writer) Bin256(v types.Bin256) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeBin256(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

// Float

func (w *Writer) Float32(v float32) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeFloat32(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

func (w *Writer) Float64(v float64) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	encoding.EncodeFloat64(w.buf, v)
	end := w.buf.Len()

	return w.pushData(start, end)
}

// Bytes/string

func (w *Writer) Bytes(v []byte) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	if _, err := encoding.EncodeBytes(w.buf, v); err != nil {
		return w.close(err)
	}
	end := w.buf.Len()

	return w.pushData(start, end)
}

func (w *Writer) String(v string) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	if _, err := encoding.EncodeString(w.buf, v); err != nil {
		return w.close(err)
	}
	end := w.buf.Len()

	return w.pushData(start, end)
}

// List

func (w *Writer) List() ListWriter {
	w.BeginList()
	return ListWriter{w}
}

func (w *Writer) BeginList() error {
	if w.err != nil {
		return w.err
	}

	// push list
	start := w.buf.Len()
	tableStart := w.elements.offset()

	w.stack.pushList(start, tableStart)
	return nil
}

func (w *Writer) BeginElement() error {
	if w.err != nil {
		return w.err
	}

	// check list
	list, ok := w.stack.peek()
	switch {
	case !ok:
		return w.closef("begin element: cannot begin element, parent not list")
	case list.type_ != entryList:
		return w.closef("begin element: cannot begin element, parent not list")
	}

	// push list element
	start := w.buf.Len()
	w.stack.pushElement(start)
	return nil
}

func (w *Writer) Element() error {
	if w.err != nil {
		return w.err
	}

	// pop data
	_, end, err := w.popData()
	if err != nil {
		return w.close(err)
	}

	// check list
	list, ok := w.stack.peek()
	switch {
	case !ok:
		return w.closef("element: cannot encode element, parent not list")
	case list.type_ != entryList:
		return w.closef("element: cannot encode element, parent not list")
	}

	// append element relative offset
	offset := uint32(end - list.start)
	element := encoding.ListElement{Offset: offset}
	w.elements.push(element)
	return nil
}

func (w *Writer) ListLen() int {
	if w.err != nil {
		return 0
	}

	// check list
	list, ok := w.stack.peek()
	switch {
	case !ok:
		return 0
	case list.type_ != entryList:
		return 0
	}

	start := list.start
	return w.elements.len(start)
}

func (w *Writer) endElement() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}

	// pop data
	_, end, err := w.popData()
	if err != nil {
		return nil, w.close(err)
	}

	// pop element
	elem, ok := w.stack.pop()
	switch {
	case !ok:
		return nil, w.closef("end element: not element")
	case elem.type_ != entryElement:
		return nil, w.closef("end element: not element")
	}

	// check list
	list, ok := w.stack.peek()
	switch {
	case !ok:
		return nil, w.closef("end element: parent not list")
	case list.type_ != entryList:
		return nil, w.closef("end element: parent not list")
	}

	// append element relative offset
	offset := uint32(end - list.start)
	element := encoding.ListElement{Offset: offset}
	w.elements.push(element)

	// return data
	b := w.buf.Bytes()
	b = b[elem.start:end]
	return b, nil
}

func (w *Writer) endList() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}

	// pop list
	list, ok := w.stack.pop()
	switch {
	case !ok:
		return nil, w.closef("end list: not list")
	case list.type_ != entryList:
		return nil, w.closef("end list: not list")
	}

	bodySize := w.buf.Len() - list.start
	table := w.elements.pop(list.tableStart)

	// encode list
	if _, err := encoding.EncodeListMeta(w.buf, bodySize, table); err != nil {
		return nil, w.close(err)
	}

	// push data entry
	start := list.start
	end := w.buf.Len()
	if err := w.pushData(start, end); err != nil {
		return nil, err
	}

	// return data
	b := w.buf.Bytes()
	b = b[start:end]
	return b, nil
}

// Message

func (w *Writer) Message() MessageWriter {
	w.BeginMessage()
	return MessageWriter{w}
}

func (w *Writer) BeginMessage() error {
	if w.err != nil {
		return w.err
	}

	// push message
	start := w.buf.Len()
	tableStart := w.fields.offset()

	w.stack.pushMessage(start, tableStart)
	return nil
}

func (w *Writer) BeginField(tag uint16) error {
	if w.err != nil {
		return w.err
	}

	// check message
	message, ok := w.stack.peek()
	switch {
	case !ok:
		return w.closef("begin field: cannot begin field, parent not message")
	case message.type_ != entryMessage:
		return w.closef("begin field: cannot begin field, parent not message")
	}

	// push field
	start := w.buf.Len()
	w.stack.pushField(start, tag)
	return nil
}

func (w *Writer) Field(tag uint16) error {
	if w.err != nil {
		return w.err
	}

	// pop data
	_, end, err := w.popData()
	if err != nil {
		return w.close(err)
	}

	// check message
	message, ok := w.stack.peek()
	switch {
	case !ok:
		return w.closef("field: cannot encode field, parent not message")
	case message.type_ != entryMessage:
		return w.closef("field: cannot encode field, parent not message")
	}

	// insert field tag and relative offset
	f := encoding.MessageField{
		Tag:    tag,
		Offset: uint32(end - message.start),
	}
	w.fields.insert(message.tableStart, f)
	return nil
}

func (w *Writer) FieldBytes(tag uint16, data []byte) error {
	if w.err != nil {
		return w.err
	}

	start := w.buf.Len()
	w.buf.Write(data)
	end := w.buf.Len()

	if err := w.pushData(start, end); err != nil {
		return err
	}
	return w.Field(tag)
}

func (w *Writer) HasField(tag uint16) bool {
	if w.err != nil {
		return false
	}

	// peek message
	message, ok := w.stack.peek()
	switch {
	case !ok:
		return false
	case message.type_ != entryMessage:
		return false
	}

	// check field table
	offset := message.tableStart
	return w.fields.hasField(offset, tag)
}

func (w *Writer) endField() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}

	// pop data
	_, end, err := w.popData()
	if err != nil {
		return nil, w.close(err)
	}

	// pop field
	field, ok := w.stack.pop()
	switch {
	case !ok:
		return nil, w.closef("end field: not field")
	case field.type_ != entryField:
		return nil, w.closef("end field: not field")
	}
	tag := field.tag()

	// check message
	message, ok := w.stack.peek()
	switch {
	case !ok:
		return nil, w.closef("field: cannot encode field, parent not message")
	case message.type_ != entryMessage:
		return nil, w.closef("field: cannot encode field, parent not message")
	}

	// insert field with tag and relative offset
	f := encoding.MessageField{
		Tag:    tag,
		Offset: uint32(end - message.start),
	}
	w.fields.insert(message.tableStart, f)

	// return data
	b := w.buf.Bytes()
	b = b[field.start:end]
	return b, nil
}

func (w *Writer) endMessage() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}

	// pop message
	message, ok := w.stack.pop()
	switch {
	case !ok:
		return nil, w.closef("end message: parent not message")
	case message.type_ != entryMessage:
		return nil, w.closef("end message: parent not message")
	}

	dataSize := w.buf.Len() - message.start
	table := w.fields.pop(message.tableStart)

	// encode message
	if _, err := encoding.EncodeMessageMeta(w.buf, dataSize, table); err != nil {
		return nil, w.close(err)
	}

	// push data
	start := message.start
	end := w.buf.Len()
	if err := w.pushData(start, end); err != nil {
		return nil, err
	}

	// return data
	b := w.buf.Bytes()
	b = b[start:end]
	return b, nil
}

// data

func (w *Writer) pushData(start, end int) error {
	entry, ok := w.stack.peek()
	if ok {
		if entry.type_ == entryData {
			return w.closef("cannot push more data, element/field must be written first")
		}
	}

	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) popData() (start, end int, err error) {
	entry, ok := w.stack.pop()
	switch {
	case !ok:
		return 0, 0, w.closef("cannot pop data, no data")
	case entry.type_ != entryData:
		return 0, 0, w.closef("cannot pop data, not data, type=%v", entry.type_)
	}

	start, end = entry.start, entry.end()
	return
}
