package writer

import (
	"errors"
	"fmt"

	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/spec/encoding"
)

// Writer writes spec objects.
type Writer interface {
	// Err returns an error or nil.
	Err() error

	// Reset resets the writer and sets its output buffer.
	Reset(buf buffer.Buffer)

	// Objects

	// List begins a new list and returns a list writer.
	List() ListWriter

	// Value returns a value writer.
	Value() ValueWriter

	// Message begins a new message and returns a message writer.
	Message() MessageWriter

	// Internal

	// Free frees the writer and releases its internal resources.
	Free()
}

// New returns a new writer with a new empty buffer.
func New() Writer {
	buf := buffer.New()
	return newWriter(buf)
}

// NewBuffer returns a new writer with the given buffer.
func NewBuffer(buf buffer.Buffer) Writer {
	return newWriter(buf)
}

// internal

var writerClosed = errors.New("operation on a closed writer")

type writer struct {
	*writerState
	err error
}

func newWriter(buf buffer.Buffer) *writer {
	s := getWriterState()
	s.init(buf)

	return &writer{writerState: s}
}

// Err returns an error or nil.
func (w *writer) Err() error {
	return w.err
}

// Reset resets the writer and sets its output buffer.
func (w *writer) Reset(buf buffer.Buffer) {
	w.close(nil)
	w.err = nil

	if buf == nil {
		buf = buffer.New()
	}

	s := getWriterState()
	s.init(buf)
	w.writerState = s
}

// Objects

// List begins a new list and returns a list writer.
func (w *writer) List() ListWriter {
	w.beginList()
	return ListWriter{w}
}

// Value returns a value writer.
func (w *writer) Value() ValueWriter {
	return ValueWriter{w}
}

// Message begins a new message and returns a message writer.
func (w *writer) Message() MessageWriter {
	w.beginMessage()
	return MessageWriter{w}
}

// Internal

// Free frees the writer and releases its internal resources.
func (w *writer) Free() {
	w.close(nil)
}

// end ends a nested object and a parent field/element if present.
func (w *writer) end() (result []byte, err error) {
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

// list

func (w *writer) beginList() error {
	if w.err != nil {
		return w.err
	}

	// push list
	start := w.buf.Len()
	tableStart := w.elements.offset()

	w.stack.pushList(start, tableStart)
	return nil
}

func (w *writer) beginElement() error {
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

func (w *writer) element() error {
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

func (w *writer) listLen() int {
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

func (w *writer) endElement() ([]byte, error) {
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

func (w *writer) endList() ([]byte, error) {
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

// message

func (w *writer) beginMessage() error {
	if w.err != nil {
		return w.err
	}

	// push message
	start := w.buf.Len()
	tableStart := w.fields.offset()

	w.stack.pushMessage(start, tableStart)
	return nil
}

func (w *writer) beginField(tag uint16) error {
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

func (w *writer) field(tag uint16) error {
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

func (w *writer) fieldAny(tag uint16, data []byte) error {
	if w.err != nil {
		return w.err
	}

	_, _, err := encoding.DecodeType(data)
	if err != nil {
		return w.close(err)
	}

	start := w.buf.Len()
	w.buf.Write(data)
	end := w.buf.Len()

	if err := w.pushData(start, end); err != nil {
		return err
	}
	return w.field(tag)
}

func (w *writer) hasField(tag uint16) bool {
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

func (w *writer) endField() ([]byte, error) {
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

func (w *writer) endMessage() ([]byte, error) {
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

func (w *writer) pushData(start, end int) error {
	entry, ok := w.stack.peek()
	if ok {
		if entry.type_ == entryData {
			return w.closef("cannot push more data, element/field must be written first")
		}
	}

	w.stack.pushData(start, end)
	return nil
}

func (w *writer) popData() (start, end int, err error) {
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

// close

// close closes the writer and releases its state.
func (w *writer) close(err error) error {
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

func (w *writer) closef(format string, args ...any) error {
	var err error
	if len(args) == 0 {
		err = errors.New(format)
	} else {
		err = fmt.Errorf(format, args...)
	}
	return w.close(err)
}
