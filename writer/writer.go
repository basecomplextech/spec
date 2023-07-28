package writer

import (
	"errors"
	"fmt"

	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/encoding"
)

// Writer writes spec objects.
type Writer interface {
	// Err returns the current write error or nil.
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
// The writer must be freed manually.
func New(autoRelease bool) Writer {
	return newWriter(nil, autoRelease)
}

// NewBuffer returns a new writer with the given buffer.
// The writer must be freed manually.
func NewBuffer(buf buffer.Buffer, autoRelease bool) Writer {
	return newWriter(buf, autoRelease)
}

// internal

var errClosed = errors.New("operation on closed writer")

type writer struct {
	*writerState

	err         error
	autoRelease bool // whether to release the writer state on close
}

func newWriter(buf buffer.Buffer, autoRelease bool) *writer {
	w := &writer{autoRelease: autoRelease}
	w.Reset(buf)
	return w
}

// Err returns the current write error or nil.
func (w *writer) Err() error {
	return w.err
}

// Reset resets the writer and sets its output buffer.
func (w *writer) Reset(buf buffer.Buffer) {
	w.err = nil

	if buf == nil {
		buf = buffer.New()
	}

	s := w.writerState
	if s == nil {
		s = acquireWriterState()
	}

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
	w.err = errClosed
	w.free()
}

// end ends the top object and its parent field/element if present.
func (w *writer) end() (result []byte, err error) {
	if w.err != nil {
		return nil, w.err
	}

	// End top object
	entry, ok := w.stack.peek()
	if !ok {
		return nil, w.failf("end: stack is empty")
	}

	switch entry.type_ {
	case entryData:
		result, err = w.endValue()
		if err != nil {
			return nil, err
		}

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
		return nil, w.failf("end: cannot end object, invalid entry type: %v", entry.type_)
	}

	// Maybe end parent field/element
	entry, ok = w.stack.peekSecondLast()
	if !ok {
		return result, w.close()
	}

	switch entry.type_ {
	case entryElement:
		return w.endElement()
	case entryField:
		return w.endField()
	}
	return result, nil
}

// value

func (w *writer) endValue() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}

	if w.stack.len() > 1 {
		return nil, w.failf("end value: cannot end value, not root value")
	}

	// Pop data
	entry, ok := w.stack.pop()
	switch {
	case !ok:
		return nil, w.failf("end value: no data entry")
	case entry.type_ != entryData:
		return nil, w.failf("end value: not data entry, type=%v", entry.type_)
	}

	// Return data
	start := entry.start
	end := w.buf.Len()

	b := w.buf.Bytes()
	b = b[start:end]
	return b, nil
}

// list

func (w *writer) beginList() error {
	if w.err != nil {
		return w.err
	}

	// Push list
	start := w.buf.Len()
	tableStart := w.elements.offset()

	w.stack.pushList(start, tableStart)
	return nil
}

func (w *writer) beginElement() error {
	if w.err != nil {
		return w.err
	}

	// Check list
	list, ok := w.stack.peek()
	switch {
	case !ok:
		return w.failf("begin element: cannot begin element, parent not list")
	case list.type_ != entryList:
		return w.failf("begin element: cannot begin element, parent not list")
	}

	// Push list element
	start := w.buf.Len()
	w.stack.pushElement(start)
	return nil
}

func (w *writer) element() error {
	if w.err != nil {
		return w.err
	}

	// Pop data
	_, end, err := w.popData()
	if err != nil {
		return w.fail(err)
	}

	// Check list
	list, ok := w.stack.peek()
	switch {
	case !ok:
		return w.failf("element: cannot encode element, parent not list")
	case list.type_ != entryList:
		return w.failf("element: cannot encode element, parent not list")
	}

	// Append element relative offset
	offset := uint32(end - list.start)
	element := encoding.ListElement{Offset: offset}
	w.elements.push(element)
	return nil
}

func (w *writer) listLen() int {
	if w.err != nil {
		return 0
	}

	// Check list
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

	// Pop data
	_, end, err := w.popData()
	if err != nil {
		return nil, w.fail(err)
	}

	// Pop element
	elem, ok := w.stack.pop()
	switch {
	case !ok:
		return nil, w.failf("end element: not element")
	case elem.type_ != entryElement:
		return nil, w.failf("end element: not element")
	}

	// Check list
	list, ok := w.stack.peek()
	switch {
	case !ok:
		return nil, w.failf("end element: parent not list")
	case list.type_ != entryList:
		return nil, w.failf("end element: parent not list")
	}

	// Append element relative offset
	offset := uint32(end - list.start)
	element := encoding.ListElement{Offset: offset}
	w.elements.push(element)

	// Return data
	b := w.buf.Bytes()
	b = b[elem.start:end]
	return b, nil
}

func (w *writer) endList() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}

	// Pop list
	list, ok := w.stack.pop()
	switch {
	case !ok:
		return nil, w.failf("end list: not list")
	case list.type_ != entryList:
		return nil, w.failf("end list: not list")
	}

	bodySize := w.buf.Len() - list.start
	table := w.elements.pop(list.tableStart)

	// Encode list
	if _, err := encoding.EncodeListMeta(w.buf, bodySize, table); err != nil {
		return nil, w.fail(err)
	}

	// Push data entry
	start := list.start
	end := w.buf.Len()
	if err := w.pushData(start, end); err != nil {
		return nil, err
	}

	// Return data
	b := w.buf.Bytes()
	b = b[start:end]
	return b, nil
}

// message

func (w *writer) beginMessage() error {
	if w.err != nil {
		return w.err
	}

	// Push message
	start := w.buf.Len()
	tableStart := w.fields.offset()

	w.stack.pushMessage(start, tableStart)
	return nil
}

func (w *writer) beginField(tag uint16) error {
	if w.err != nil {
		return w.err
	}

	// Check message
	message, ok := w.stack.peek()
	switch {
	case !ok:
		return w.failf("begin field: cannot begin field, parent not message")
	case message.type_ != entryMessage:
		return w.failf("begin field: cannot begin field, parent not message")
	}

	// Push field
	start := w.buf.Len()
	w.stack.pushField(start, tag)
	return nil
}

func (w *writer) field(tag uint16) error {
	if w.err != nil {
		return w.err
	}

	// Pop data
	_, end, err := w.popData()
	if err != nil {
		return w.fail(err)
	}

	// Check message
	message, ok := w.stack.peek()
	switch {
	case !ok:
		return w.failf("field: cannot encode field, parent not message")
	case message.type_ != entryMessage:
		return w.failf("field: cannot encode field, parent not message")
	}

	// Insert field tag and relative offset
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
		return w.fail(err)
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

	// Peek message
	message, ok := w.stack.peek()
	switch {
	case !ok:
		return false
	case message.type_ != entryMessage:
		return false
	}

	// Check field table
	offset := message.tableStart
	return w.fields.hasField(offset, tag)
}

func (w *writer) endField() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}

	// Pop data
	_, end, err := w.popData()
	if err != nil {
		return nil, w.fail(err)
	}

	// Pop field
	field, ok := w.stack.pop()
	switch {
	case !ok:
		return nil, w.failf("end field: not field")
	case field.type_ != entryField:
		return nil, w.failf("end field: not field")
	}
	tag := field.tag()

	// Check message
	message, ok := w.stack.peek()
	switch {
	case !ok:
		return nil, w.failf("field: cannot encode field, parent not message")
	case message.type_ != entryMessage:
		return nil, w.failf("field: cannot encode field, parent not message")
	}

	// Insert field with tag and relative offset
	f := encoding.MessageField{
		Tag:    tag,
		Offset: uint32(end - message.start),
	}
	w.fields.insert(message.tableStart, f)

	// Return data
	b := w.buf.Bytes()
	b = b[field.start:end]
	return b, nil
}

func (w *writer) endMessage() ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}

	// Pop message
	message, ok := w.stack.pop()
	switch {
	case !ok:
		return nil, w.failf("end message: parent not message")
	case message.type_ != entryMessage:
		return nil, w.failf("end message: parent not message")
	}

	dataSize := w.buf.Len() - message.start
	table := w.fields.pop(message.tableStart)

	// Encode message
	if _, err := encoding.EncodeMessageMeta(w.buf, dataSize, table); err != nil {
		return nil, w.fail(err)
	}

	// Push data
	start := message.start
	end := w.buf.Len()
	if err := w.pushData(start, end); err != nil {
		return nil, err
	}

	// Return data
	b := w.buf.Bytes()
	b = b[start:end]
	return b, nil
}

// data

func (w *writer) pushData(start, end int) error {
	entry, ok := w.stack.peek()
	if ok {
		if entry.type_ == entryData {
			return w.failf("cannot push more data, element/field must be written first")
		}
	}

	w.stack.pushData(start, end)
	return nil
}

func (w *writer) popData() (start, end int, err error) {
	entry, ok := w.stack.pop()
	switch {
	case !ok:
		return 0, 0, w.failf("cannot pop data, no data")
	case entry.type_ != entryData:
		return 0, 0, w.failf("cannot pop data, not data, type=%v", entry.type_)
	}

	start, end = entry.start, entry.end()
	return
}

// close

// close sets the writer error and frees its state.
func (w *writer) close() error {
	if w.err != nil {
		return w.err
	}

	w.err = errClosed
	if w.autoRelease {
		w.free()
	}
	return nil
}

// fail sets the writer error and frees its state.
func (w *writer) fail(err error) error {
	if w.err != nil {
		return w.err
	}

	if err == nil {
		return w.close()
	}

	w.err = err
	w.free()
	return w.err
}

// failf sets the writer error and frees its state.
func (w *writer) failf(format string, args ...any) error {
	var err error
	if len(args) == 0 {
		err = errors.New(format)
	} else {
		err = fmt.Errorf(format, args...)
	}

	return w.fail(err)
}

// free

func (w *writer) free() {
	s := w.writerState
	w.writerState = nil

	releaseWriterState(s)
}
