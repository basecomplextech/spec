package spec

import "fmt"

const WriteBufferSize = 4096

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

// Reset clears the writer.
func (w *Writer) Reset() {
	w.buf = nil
	w.err = nil
	w.data = writeData{}

	w.objects.reset()
	w.elements.reset()
	w.fields.reset()
}

// Primitive

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
	return w.UInt8(v)
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

func (w *Writer) UInt8(v uint8) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = writeUInt8(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) UInt16(v uint16) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = writeUInt16(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) UInt32(v uint32) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = writeUInt32(w.buf, v)
	end := len(w.buf)

	if err := w.setData(start, end); err != nil {
		return w.fail(err)
	}
	return nil
}

func (w *Writer) UInt64(v uint64) error {
	if w.err != nil {
		return w.err
	}

	start := len(w.buf)
	w.buf = writeUInt64(w.buf, v)
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

	// pop data
	data := w.popData()
	list, err := w.objects.lastList()
	if err != nil {
		return w.fail(err)
	}

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
	list, err := w.objects.popList()
	if err != nil {
		return w.fail(err)
	}

	dsize := len(w.buf) - list.start
	table := w.elements.pop(list.tableStart)

	// write list
	w.buf, err = writeList(w.buf, dsize, table)
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

	// pop data
	data := w.popData()
	message, err := w.objects.lastMessage()
	if err != nil {
		return w.fail(err)
	}

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
	message, err := w.objects.popMessage()
	if err != nil {
		return w.fail(err)
	}

	dsize := len(w.buf) - message.start
	table := w.fields.pop(message.tableStart)

	// write mesasge
	w.buf, err = writeMessage(w.buf, dsize, table)
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

// object

type objectType byte

const (
	objectTypeUndefined objectType = iota
	objectTypeList
	objectTypeMessage
)

type objectEntry struct {
	start      int // start offset in data buffer
	tableStart int // table offset in list/message stack
	type_      objectType
}

// stack

type objectStack struct {
	stack []objectEntry
}

func (s *objectStack) reset() {
	s.stack = s.stack[:0]
}

func (s *objectStack) len() int {
	return len(s.stack)
}

// last

// last returns the last object and checks its type.
func (s *objectStack) last(type_ objectType) (objectEntry, error) {
	ln := len(s.stack)
	if ln == 0 {
		return objectEntry{}, fmt.Errorf("last: object stack is empty")
	}

	e := s.stack[ln-1]
	if e.type_ != type_ {
		return e, fmt.Errorf("last: unexpected stack object, expected=%v, actual=%v, ",
			objectTypeList, e.type_)
	}
	return e, nil
}

func (s *objectStack) lastList() (objectEntry, error) {
	return s.last(objectTypeList)
}

func (s *objectStack) lastMessage() (objectEntry, error) {
	return s.last(objectTypeMessage)
}

// pop

// pop removes the top object from the stack and checks its type.
func (s *objectStack) pop(type_ objectType) (objectEntry, error) {
	ln := len(s.stack)
	if ln == 0 {
		return objectEntry{}, fmt.Errorf("pop: stack is empty")
	}

	e := s.stack[ln-1]
	if e.type_ != type_ {
		return e, fmt.Errorf("peek: unexpected object, expected=%v, actual=%v, ", type_, e.type_)
	}

	s.stack = s.stack[:ln-1]
	return e, nil
}

func (s *objectStack) popList() (objectEntry, error) {
	return s.pop(objectTypeList)
}

func (s *objectStack) popMessage() (objectEntry, error) {
	return s.pop(objectTypeMessage)
}

// push

func (s *objectStack) pushList(start int, tableStart int) {
	e := objectEntry{
		type_:      objectTypeList,
		start:      start,
		tableStart: tableStart,
	}
	s.stack = append(s.stack, e)
}

func (s *objectStack) pushMessage(start int, tableStart int) {
	e := objectEntry{
		type_:      objectTypeMessage,
		start:      start,
		tableStart: tableStart,
	}
	s.stack = append(s.stack, e)
}

// util

func (t objectType) String() string {
	switch t {
	case objectTypeUndefined:
		return "undefined"
	case objectTypeList:
		return "list"
	case objectTypeMessage:
		return "message"
	}
	return ""
}
