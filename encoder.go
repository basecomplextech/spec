package spec

import (
	"errors"
	"fmt"
	"sync"

	"github.com/complexl/library/buffer"
	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

const BufferSize = 4096

var encoderPool = &sync.Pool{
	New: func() interface{} {
		return newEncoder(nil)
	},
}

// Encoder encodes values.
type Encoder struct {
	buf  buffer.Buffer
	err  error      // encoding failed
	data encodeData // last written data, must be consumed before writing next data

	stack    stack
	elements listBuffer    // buffer for list element tables
	fields   messageBuffer // buffer for message field tables

	// preallocated
	_stack    [16]stackEntry
	_elements [128]listElement
	_fields   [128]messageField
}

// NewEncoder returns a new encoder with an empty buffer.
//
// Usually, it is better to use Encode(obj) and EncodeTo(obj, buf), than to construct
// a new encoder directly. These methods internally use an encoder pool.
func NewEncoder() *Encoder {
	buf := buffer.New(nil)
	return newEncoder(buf)
}

// NewEncoderBuffer returns a new encoder with a buffer.
//
// Usually, it is better to use Encode(obj) and EncodeTo(obj, buf), than to construct
// a new encoder directly. These methods internally use an encoder pool.
func NewEncoderBuffer(b []byte) *Encoder {
	buf := buffer.New(b)
	return newEncoder(buf)
}

func newEncoder(buf buffer.Buffer) *Encoder {
	e := &Encoder{
		buf:  buf,
		data: encodeData{},
	}

	e.stack.stack = e._stack[:0]
	e.elements.stack = e._elements[:0]
	e.fields.stack = e._fields[:0]
	return e
}

// Init resets the encoder and sets its buffer.
func (e *Encoder) Init(buf buffer.Buffer) {
	e.Reset()
	e.buf = buf
}

// Reset clears the encoder and nils buffer.
func (e *Encoder) Reset() {
	e.buf = nil
	e.err = nil
	e.data = encodeData{}

	e.stack.reset()
	e.elements.reset()
	e.fields.reset()
}

// End ends a nested object and a parent field/element if present.
func (e *Encoder) End() (result []byte, err error) {
	if e.err != nil {
		return nil, e.err
	}

	// end top object
	entry, ok := e.stack.peek()
	if !ok {
		return nil, e.fail(fmt.Errorf("end: encode stack is empty"))
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
		return nil, e.fail(errors.New("end: not nested encoder"))
	}

	// end parent field/element
	entry, ok = e.stack.peek()
	if !ok {
		return result, nil
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

func (e *Encoder) Nil() error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeNil(e.buf)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Encoder) Bool(v bool) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeBool(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Encoder) Byte(v byte) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeByte(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Encoder) Int32(v int32) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeInt32(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Encoder) Int64(v int64) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeInt64(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Encoder) Uint32(v uint32) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeUint32(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Encoder) Uint64(v uint64) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeUint64(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

// U128/U256

func (e *Encoder) U128(v u128.U128) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeU128(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Encoder) U256(v u256.U256) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeU256(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

// Float

func (e *Encoder) Float32(v float32) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeFloat32(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Encoder) Float64(v float64) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	EncodeFloat64(e.buf, v)
	end := e.buf.Len()

	return e.setData(start, end)
}

// Bytes/string

func (e *Encoder) Bytes(v []byte) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	if _, err := EncodeBytes(e.buf, v); err != nil {
		return e.fail(err)
	}
	end := e.buf.Len()

	return e.setData(start, end)
}

func (e *Encoder) String(v string) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	if _, err := EncodeString(e.buf, v); err != nil {
		return e.fail(err)
	}
	end := e.buf.Len()

	return e.setData(start, end)
}

// List

func (e *Encoder) BeginList() error {
	if e.err != nil {
		return e.err
	}

	// push list
	start := e.buf.Len()
	tableStart := e.elements.offset()

	e.stack.pushList(start, tableStart)
	return nil
}

func (e *Encoder) BeginElement() error {
	if e.err != nil {
		return e.err
	}

	// check list
	list, ok := e.stack.peek()
	switch {
	case !ok:
		return e.fail(errors.New("begin element: cannot begin element, not list encoder"))
	case list.type_ != entryList:
		return e.fail(errors.New("begin element: cannot begin element, not list encoder"))
	}

	// push list element
	start := e.buf.Len()
	e.stack.pushElement(start)
	return nil
}

func (e *Encoder) Element() error {
	if e.err != nil {
		return e.err
	}

	// check list
	list, ok := e.stack.peek()
	switch {
	case !ok:
		return e.fail(errors.New("element: cannot encode element, not list encoder"))
	case list.type_ != entryList:
		return e.fail(errors.New("element: cannot encode element, not list encoder"))
	}

	// pop data
	data := e.popData()

	// append element relative offset
	offset := uint32(data.end - list.start)
	element := listElement{offset: offset}
	e.elements.push(element)
	return nil
}

func (e *Encoder) endElement() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}

	// check element
	elem, ok := e.stack.pop()
	switch {
	case !ok:
		return nil, e.fail(errors.New("end element: not an element"))
	case elem.type_ != entryElement:
		return nil, e.fail(errors.New("end element: not an element"))
	}

	// check list
	list, ok := e.stack.peek()
	switch {
	case !ok:
		return nil, e.fail(errors.New("end element: parent not a list"))
	case list.type_ != entryList:
		return nil, e.fail(errors.New("end element: parent not list encoder"))
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

func (e *Encoder) endList() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}

	// pop list
	list, ok := e.stack.pop()
	switch {
	case !ok:
		return nil, e.fail(errors.New("end list: not list encoder"))
	case list.type_ != entryList:
		return nil, e.fail(errors.New("end list: not list encoder"))
	}

	bodySize := e.buf.Len() - list.start
	table := e.elements.pop(list.tableStart)

	// encode list
	if _, err := encodeListMeta(e.buf, bodySize, table); err != nil {
		return nil, e.fail(err)
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

func (e *Encoder) BeginMessage() error {
	if e.err != nil {
		return e.err
	}

	// push message
	start := e.buf.Len()
	tableStart := e.fields.offset()

	e.stack.pushMessage(start, tableStart)
	return nil
}

func (e *Encoder) BeginField(tag uint16) error {
	if e.err != nil {
		return e.err
	}

	// check message
	message, ok := e.stack.peek()
	switch {
	case !ok:
		return e.fail(errors.New("begin field: cannot begin field, not message encoder"))
	case message.type_ != entryMessage:
		return e.fail(errors.New("begin field: cannot begin field, not message encoder"))
	}

	// push field
	start := e.buf.Len()
	e.stack.pushField(start, tag)
	return nil
}

func (e *Encoder) Field(tag uint16) error {
	if e.err != nil {
		return e.err
	}

	// check message
	message, ok := e.stack.peek()
	switch {
	case !ok:
		return e.fail(errors.New("field: cannot encode field, not message encoder"))
	case message.type_ != entryMessage:
		return e.fail(errors.New("field: cannot encode field, not message encoder"))
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

func (e *Encoder) endField() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}

	// check field
	field, ok := e.stack.pop()
	switch {
	case !ok:
		return nil, e.fail(errors.New("end field: not a field"))
	case field.type_ != entryField:
		return nil, e.fail(errors.New("end field: not a field"))
	}
	tag := field.tag()

	// check message
	message, ok := e.stack.peek()
	switch {
	case !ok:
		return nil, e.fail(errors.New("field: cannot encode field, not message encoder"))
	case message.type_ != entryMessage:
		return nil, e.fail(errors.New("field: cannot encode field, not message encoder"))
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

func (e *Encoder) endMessage() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}

	// pop message
	message, ok := e.stack.pop()
	switch {
	case !ok:
		return nil, e.fail(errors.New("end message: not message encoder"))
	case message.type_ != entryMessage:
		return nil, e.fail(errors.New("end message: not message encoder"))
	}

	bsize := e.buf.Len() - message.start
	table := e.fields.pop(message.tableStart)

	// encode message
	if _, err := encodeMessageMeta(e.buf, bsize, table); err != nil {
		return nil, e.fail(err)
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

// Struct

func (e *Encoder) BeginStruct() error {
	if e.err != nil {
		return e.err
	}

	// push struct
	start := e.buf.Len()
	e.stack.pushStruct(start)
	return nil
}

func (e *Encoder) StructField() error {
	if e.err != nil {
		return e.err
	}

	// check struct
	entry, ok := e.stack.peek()
	switch {
	case !ok:
		return e.fail(errors.New("field: cannot encode struct field, not struct encoder"))
	case entry.type_ != entryStruct:
		return e.fail(errors.New("field: cannot encode struct field, not struct encoder"))
	}

	// just consume data
	e.popData()
	return nil
}

func (e *Encoder) EndStruct() error {
	if e.err != nil {
		return e.err
	}

	// pop struct
	entry, ok := e.stack.pop()
	switch {
	case !ok:
		return e.fail(errors.New("end struct: not struct encoder"))
	case entry.type_ != entryStruct:
		return e.fail(errors.New("end struct: not struct encoder"))
	}

	bsize := e.buf.Len() - entry.start

	// encode struct
	if _, err := encodeStruct(e.buf, bsize); err != nil {
		return e.fail(err)
	}

	// push data
	start := entry.start
	end := e.buf.Len()
	return e.setData(start, end)
}

// private

func (e *Encoder) fail(err error) error {
	if e.err != nil {
		return err
	}

	e.err = err
	return err
}

// data

// encodeData holds the last written data start/end.
// there is no data stack because the data must be consumed immediatelly after it is written.
type encodeData struct {
	start int
	end   int
}

// TODO: Rename into pushData and move to stack.
func (e *Encoder) setData(start, end int) error {
	if e.data.start != 0 || e.data.end != 0 {
		err := errors.New("encode: cannot encode more data, element/field must be written first")
		return e.fail(err)
	}

	e.data = encodeData{
		start: start,
		end:   end,
	}
	return nil
}

func (e *Encoder) popData() encodeData {
	d := e.data
	e.data = encodeData{}
	return d
}
