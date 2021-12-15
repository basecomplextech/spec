package protocol

import "fmt"

const WriteBufferSize = 4096

// Writer writes objects to a buffer.
type Writer interface {
	// End ends writing, returns the result bytes, and resets the writer.
	End() ([]byte, error)

	// Reset clears the writer.
	Reset()

	// List

	BeginList() error
	EndList() error

	BeginElement() error
	EndElement() error

	// Message

	BeginMessage() error
	EndMessage() error

	BeginField(tag uint16) error
	EndField() error

	// Values

	WriteBool(v bool) error
	WriteByte(v byte) error

	WriteInt8(v int8) error
	WriteInt16(v int16) error
	WriteInt32(v int32) error
	WriteInt64(v int64) error

	WriteUInt8(v uint8) error
	WriteUInt16(v uint16) error
	WriteUInt32(v uint32) error
	WriteUInt64(v uint64) error

	WriteFloat32(v float32) error
	WriteFloat64(v float64) error

	WriteBytes(v []byte) error
	WriteString(v string) error
}

// NewWriter returns a new writer.
func NewWriter() Writer {
	return newWriter()
}

type writer struct {
	data     writeBuffer
	stack    writeStack
	fields   fieldStack
	elements elementStack
}

func newWriter() *writer {
	return &writer{}
}

// End ends writing, returns the result bytes, and resets the writer.
func (w *writer) End() ([]byte, error) {
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
func (w *writer) Reset() {
	w.data = w.data[:0]
	w.stack = w.stack[:0]
	w.fields = w.fields[:0]
	w.elements = w.elements[:0]
}

// List

func (w *writer) BeginList() error {
	// push list
	start := w.data.offset()
	tableStart := w.elements.offset()

	w.stack.pushList(start, tableStart)
	return nil
}

func (w *writer) EndList() error {
	// pop list
	list, err := w.stack.popType(entryTypeList)
	if err != nil {
		return err
	}
	dataSize := uint32(w.data.offset() - list.list.start)

	// write table
	table := w.elements.popList(list.list.tableStart)
	tableSize := w.data.listTable(table)

	// write sizes and type
	w.data.listDataSize(dataSize)
	w.data.listTableSize(tableSize)
	w.data.type_(TypeList)

	// push data entry
	start := list.list.start
	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) BeginElement() error {
	// check parent
	if _, err := w.stack.peekType(entryTypeList); err != nil {
		return err
	}

	// push element
	start := w.data.offset()
	w.stack.pushElement(start)
	return nil
}

func (w *writer) EndElement() error {
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
	offset := uint32(data.data.end - list.list.start)
	element := element{offset: offset}
	w.elements.push(element)

	// push data
	start := elem.element.start
	end := data.data.end
	w.stack.pushData(start, end)
	return nil
}

// Message

func (w *writer) BeginMessage() error {
	// push message
	start := w.data.offset()
	fieldStart := w.fields.offset()

	w.stack.pushMessage(start, fieldStart)
	return nil
}

func (w *writer) EndMessage() error {
	// pop message
	message, err := w.stack.popType(entryTypeMessage)
	if err != nil {
		return err
	}
	dataSize := uint32(w.data.offset() - message.message.start)

	// write table
	table := w.fields.popTable(message.message.tableStart)
	tableSize := w.data.messageTable(table)

	// write sizes and type
	w.data.messageDataSize(dataSize)
	w.data.messageTableSize(tableSize)
	w.data.type_(TypeMessage)

	// push data
	start := message.message.start
	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) BeginField(tag uint16) error {
	// check parent
	if _, err := w.stack.peekType(entryTypeMessage); err != nil {
		return err
	}

	// push field
	start := w.data.offset()
	w.stack.pushField(tag, start)
	return nil
}

func (w *writer) EndField() error {
	// pop data and field
	data, err := w.stack.popType(entryTypeData)
	if err != nil {
		return err
	}
	fentry, err := w.stack.popType(entryTypeField)
	if err != nil {
		return err
	}
	message, err := w.stack.peekType(entryTypeMessage)
	if err != nil {
		return err
	}

	// insert tag and relative offset
	f := messageField{
		tag:    fentry.field.tag,
		offset: uint32(data.data.end - message.message.start),
	}
	w.fields.insert(message.message.tableStart, f)
	return nil
}

// Values

func (w *writer) WriteBool(v bool) error {
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

func (w *writer) WriteByte(v byte) error {
	return w.WriteUInt8(v)
}

func (w *writer) WriteInt8(v int8) error {
	start := w.data.offset()

	w.data.int8(v)
	w.data.type_(TypeInt8)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteInt16(v int16) error {
	start := w.data.offset()

	w.data.int16(v)
	w.data.type_(TypeInt16)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteInt32(v int32) error {
	start := w.data.offset()

	w.data.int32(v)
	w.data.type_(TypeInt32)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteInt64(v int64) error {
	start := w.data.offset()

	w.data.int64(v)
	w.data.type_(TypeInt64)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteUInt8(v uint8) error {
	start := w.data.offset()

	w.data.uint8(v)
	w.data.type_(TypeUInt8)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteUInt16(v uint16) error {
	start := w.data.offset()

	w.data.uint16(v)
	w.data.type_(TypeUInt16)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteUInt32(v uint32) error {
	start := w.data.offset()

	w.data.uint32(v)
	w.data.type_(TypeUInt32)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteUInt64(v uint64) error {
	start := w.data.offset()

	w.data.uint64(v)
	w.data.type_(TypeUInt64)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteFloat32(v float32) error {
	start := w.data.offset()

	w.data.float32(v)
	w.data.type_(TypeFloat32)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteFloat64(v float64) error {
	start := w.data.offset()

	w.data.float64(v)
	w.data.type_(TypeFloat64)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteBytes(v []byte) error {
	start := w.data.offset()

	size := w.data.bytes(v)
	w.data.bytesSize(size)
	w.data.type_(TypeBytes)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteString(v string) error {
	start := w.data.offset()

	size := w.data.string(v)
	w.data.stringZero()
	w.data.stringSize(size + 1) // plus zero byte
	w.data.type_(TypeString)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}
