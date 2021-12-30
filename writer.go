package spec

import "fmt"

const WriteBufferSize = 4096

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

// NewWriter returns a new writer with a default buffer.
func NewWriter() *Writer {
	buf := make([]byte, 0, WriteBufferSize)
	return NewWriterBuffer(buf)
}

// NewWriterBuffer returns a new writer with a buffer.
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
	w.buf = writeNil(w.buf)
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
	w.buf = writeBool(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Byte(v byte) error {
	return w.Uint8(v)
}

func (w *Writer) Int8(v int8) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = writeInt8(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Int16(v int16) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = writeInt16(w.buf, v)
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
	w.buf = writeInt32(w.buf, v)
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
	w.buf = writeInt64(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Uint8(v uint8) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = writeUint8(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Uint16(v uint16) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = writeUint16(w.buf, v)
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
	w.buf = writeUint32(w.buf, v)
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
	w.buf = writeUint64(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) Float32(v float32) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = writeFloat32(w.buf, v)
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
	w.buf = writeFloat64(w.buf, v)
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
	w.buf, err = writeBytes(w.buf, v)
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
	w.buf, err = writeString(w.buf, v)
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
	w.buf, err = writeList(w.buf, bsize, table)
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

	// write mesasge
	var err error
	w.buf, err = writeMessage(w.buf, bsize, table)
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
