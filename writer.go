package protocol

import "fmt"

const WriteBufferSize = 4096

type Writer struct {
	data writeBuffer

	stack        writeStack
	listStack    listStack    // stack of list tables
	messageStack messageStack // stack of message tables

	// preallocated
	_stack        [16]entry
	_listStack    [64]listElement
	_messageStack [128]messageField
}

// NewWriter returns a new writer with a default buffer.
func NewWriter() *Writer {
	buf := make([]byte, 0, WriteBufferSize)
	return NewWriterBuffer(buf)
}

// NewWriterBuffer returns a new writer with a buffer.
func NewWriterBuffer(buf []byte) *Writer {
	w := &Writer{data: buf[:0]}
	w.stack = w._stack[:]
	w.listStack = w._listStack[:]
	w.messageStack = w._messageStack[:]
	return w
}

// End ends writing, returns the result bytes, and resets the writer.
func (w *Writer) End() ([]byte, error) {
	switch {
	case len(w.stack) == 0:
		return []byte{}, nil
	case len(w.stack) > 1:
		return nil, fmt.Errorf("writer end: incomplete write, stack size=%d", len(w.stack))
	}

	// pop data
	if _, err := w.stack.popType(entryTypeData); err != nil {
		return nil, err
	}

	// return and reset
	b := w.data[:]
	w.Reset()
	return b, nil
}

// Reset clears the writer.
func (w *Writer) Reset() {
	w.data = nil
	w.stack = w.stack[:0]
	w.listStack = w.listStack[:0]
	w.messageStack = w.messageStack[:0]
}

// Primitive

func (w *Writer) Bool(v bool) error {
	start := w.data.offset()

	if v {
		w.data.type_(TypeTrue)
	} else {
		w.data.type_(TypeFalse)
	}

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) Byte(v byte) error {
	return w.UInt8(v)
}

func (w *Writer) Int8(v int8) error {
	start := w.data.offset()

	w.data.int8(v)
	w.data.type_(TypeInt8)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) Int16(v int16) error {
	start := w.data.offset()

	w.data.int16(v)
	w.data.type_(TypeInt16)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) Int32(v int32) error {
	start := w.data.offset()

	w.data.int32(v)
	w.data.type_(TypeInt32)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) Int64(v int64) error {
	start := w.data.offset()

	w.data.int64(v)
	w.data.type_(TypeInt64)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) UInt8(v uint8) error {
	start := w.data.offset()

	w.data.uint8(v)
	w.data.type_(TypeUInt8)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) UInt16(v uint16) error {
	start := w.data.offset()

	w.data.uint16(v)
	w.data.type_(TypeUInt16)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) UInt32(v uint32) error {
	start := w.data.offset()

	w.data.uint32(v)
	w.data.type_(TypeUInt32)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) UInt64(v uint64) error {
	start := w.data.offset()

	w.data.uint64(v)
	w.data.type_(TypeUInt64)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) Float32(v float32) error {
	start := w.data.offset()

	w.data.float32(v)
	w.data.type_(TypeFloat32)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) Float64(v float64) error {
	start := w.data.offset()

	w.data.float64(v)
	w.data.type_(TypeFloat64)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

// Bytes/string

func (w *Writer) Bytes(v []byte) error {
	start := w.data.offset()

	size := w.data.bytes(v)
	w.data.bytesSize(size)
	w.data.type_(TypeBytes)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) String(v string) error {
	start := w.data.offset()

	size := w.data.string(v)
	w.data.stringZero()
	w.data.stringSize(size + 1) // plus zero byte
	w.data.type_(TypeString)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

// List

func (w *Writer) BeginList() error {
	// push list
	start := w.data.offset()
	tableStart := w.listStack.offset()

	w.stack.pushList(start, tableStart)
	return nil
}

func (w *Writer) EndList() error {
	// pop list
	list, err := w.stack.popType(entryTypeList)
	if err != nil {
		return err
	}
	dataSize := uint32(w.data.offset() - list.start)

	// write table
	table := w.listStack.pop(list.tableStart)
	tableSize := w.data.listTable(table)

	// write sizes and type
	w.data.listDataSize(dataSize)
	w.data.listTableSize(tableSize)
	w.data.type_(TypeList)

	// push data entry
	start := list.start
	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) BeginElement() error {
	// check parent
	if _, err := w.stack.peekType(entryTypeList); err != nil {
		return err
	}

	// push element
	start := w.data.offset()
	w.stack.pushElement(start)
	return nil
}

func (w *Writer) EndElement() error {
	// pop data and element
	data, err := w.stack.popType(entryTypeData)
	if err != nil {
		return err
	}
	elem, err := w.stack.popType(entryTypeElement)
	if err != nil {
		return err
	}
	list, err := w.stack.peekType(entryTypeList)
	if err != nil {
		return err
	}

	// append relative offset
	offset := uint32(data.end - list.start)
	element := listElement{offset: offset}
	w.listStack.push(element)

	// push data
	start := elem.start
	end := data.end
	w.stack.pushData(start, end)
	return nil
}

// Message

func (w *Writer) BeginMessage() error {
	// push message
	start := w.data.offset()
	tableStart := w.listStack.offset()

	w.stack.pushMessage(start, tableStart)
	return nil
}

func (w *Writer) EndMessage() error {
	// pop message
	message, err := w.stack.popType(entryTypeMessage)
	if err != nil {
		return err
	}
	dataSize := uint32(w.data.offset() - message.start)

	// write table
	table := w.messageStack.pop(message.tableStart)
	tableSize := w.data.messageTable(table)

	// write sizes and type
	w.data.messageDataSize(dataSize)
	w.data.messageTableSize(tableSize)
	w.data.type_(TypeMessage)

	// push data
	start := message.start
	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *Writer) BeginField(tag uint16) error {
	// check parent
	if _, err := w.stack.peekType(entryTypeMessage); err != nil {
		return err
	}

	// push field
	start := w.data.offset()
	w.stack.pushField(start, tag)
	return nil
}

func (w *Writer) EndField() error {
	// pop data and field
	data, err := w.stack.popType(entryTypeData)
	if err != nil {
		return err
	}
	field, err := w.stack.popType(entryTypeField)
	if err != nil {
		return err
	}
	message, err := w.stack.peekType(entryTypeMessage)
	if err != nil {
		return err
	}

	// insert tag and relative offset
	f := messageField{
		tag:    field.tag,
		offset: uint32(data.end - message.start),
	}
	w.messageStack.insert(message.tableStart, f)
	return nil
}
