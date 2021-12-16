package protocol

import "fmt"

const WriteBufferSize = 4096

type Writer struct {
	buffer writeBuffer

	data     dataStack
	objects  objectStack
	elements listStack    // stack of list element tables
	fields   messageStack // stack of message field tables

	// preallocated
	_data     [16]dataEntry
	_objects  [16]objectEntry
	_elements [64]listElement
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
	w.buffer.buffer = buf[:0]

	w.data.stack = w._data[:0]
	w.objects.stack = w._objects[:0]
	w.elements.stack = w._elements[:0]
	w.fields.stack = w._fields[:0]
	return w
}

// End ends writing, returns the result bytes, and resets the writer.
func (w *Writer) End() ([]byte, error) {
	switch {
	case w.objects.len() > 0:
		return nil, fmt.Errorf("end: incomplete objects, object stack size=%d", w.objects.len())
	case w.data.len() > 1:
		return nil, fmt.Errorf("end: incomplete write, data stack size=%d", w.data.len())
	case w.data.len() == 0:
		return nil, fmt.Errorf("end: empty writer")
	}

	// pop data
	data, err := w.data.pop()
	if err != nil {
		return nil, err
	}

	// return and reset
	b := w.buffer.buffer[data.start:data.end]
	w.Reset()
	return b, nil
}

// Reset clears the writer.
func (w *Writer) Reset() {
	w.buffer.reset()
	w.data.reset()
	w.objects.reset()
	w.elements.reset()
	w.fields.reset()
}

// Primitive

func (w *Writer) Bool(v bool) error {
	start := w.buffer.offset()

	if v {
		w.buffer.type_(TypeTrue)
	} else {
		w.buffer.type_(TypeFalse)
	}

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) Byte(v byte) error {
	return w.UInt8(v)
}

func (w *Writer) Int8(v int8) error {
	start := w.buffer.offset()

	w.buffer.int8(v)
	w.buffer.type_(TypeInt8)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) Int16(v int16) error {
	start := w.buffer.offset()

	w.buffer.int16(v)
	w.buffer.type_(TypeInt16)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) Int32(v int32) error {
	start := w.buffer.offset()

	w.buffer.int32(v)
	w.buffer.type_(TypeInt32)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) Int64(v int64) error {
	start := w.buffer.offset()

	w.buffer.int64(v)
	w.buffer.type_(TypeInt64)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) UInt8(v uint8) error {
	start := w.buffer.offset()

	w.buffer.uint8(v)
	w.buffer.type_(TypeUInt8)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) UInt16(v uint16) error {
	start := w.buffer.offset()

	w.buffer.uint16(v)
	w.buffer.type_(TypeUInt16)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) UInt32(v uint32) error {
	start := w.buffer.offset()

	w.buffer.uint32(v)
	w.buffer.type_(TypeUInt32)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) UInt64(v uint64) error {
	start := w.buffer.offset()

	w.buffer.uint64(v)
	w.buffer.type_(TypeUInt64)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) Float32(v float32) error {
	start := w.buffer.offset()

	w.buffer.float32(v)
	w.buffer.type_(TypeFloat32)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) Float64(v float64) error {
	start := w.buffer.offset()

	w.buffer.float64(v)
	w.buffer.type_(TypeFloat64)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

// Bytes/string

func (w *Writer) Bytes(v []byte) error {
	start := w.buffer.offset()

	size := w.buffer.bytes(v)
	w.buffer.bytesSize(size)
	w.buffer.type_(TypeBytes)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

func (w *Writer) String(v string) error {
	start := w.buffer.offset()

	size := w.buffer.string(v)
	w.buffer.stringZero()
	w.buffer.stringSize(size + 1) // plus zero byte
	w.buffer.type_(TypeString)

	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

// List

func (w *Writer) BeginList() error {
	// push list
	start := w.buffer.offset()
	tableStart := w.elements.offset()

	w.objects.pushList(start, tableStart)
	return nil
}

func (w *Writer) Element() error {
	// pop data
	data, err := w.data.pop()
	if err != nil {
		return err
	}
	list, err := w.objects.peekList()
	if err != nil {
		return err
	}

	// append element relative offset
	offset := uint32(data.end - list.start)
	element := listElement{offset: offset}
	w.elements.push(element)
	return nil
}

func (w *Writer) EndList() error {
	// pop list
	list, err := w.objects.popList()
	if err != nil {
		return err
	}
	dataSize := uint32(w.buffer.offset() - list.start)

	// write table
	table := w.elements.pop(list.tableStart)
	tableSize := w.buffer.listTable(table)

	// write sizes and type
	w.buffer.listDataSize(dataSize)
	w.buffer.listTableSize(tableSize)
	w.buffer.type_(TypeList)

	// push data entry
	start := list.start
	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}

// Message

func (w *Writer) BeginMessage() error {
	// push message
	start := w.buffer.offset()
	tableStart := w.fields.offset()

	w.objects.pushMessage(start, tableStart)
	return nil
}

func (w *Writer) Field(tag uint16) error {
	// pop data
	data, err := w.data.pop()
	if err != nil {
		return err
	}
	message, err := w.objects.peekMessage()
	if err != nil {
		return err
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
	// pop message
	message, err := w.objects.popMessage()
	if err != nil {
		return err
	}
	dataSize := uint32(w.buffer.offset() - message.start)

	// write table
	table := w.fields.pop(message.tableStart)
	tableSize := w.buffer.messageTable(table)

	// write sizes and type
	w.buffer.messageDataSize(dataSize)
	w.buffer.messageTableSize(tableSize)
	w.buffer.type_(TypeMessage)

	// push data
	start := message.start
	end := w.buffer.offset()
	w.data.push(start, end)
	return nil
}
