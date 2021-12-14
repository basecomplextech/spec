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
	elements writeElements
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
	elementStart := w.elements.offset()

	w.stack.pushList(start, elementStart)
	return nil
}

func (w *writer) EndList() error {
	// pop list and elements
	list, err := w.stack.popType(entryTypeList)
	if err != nil {
		return err
	}
	elements := w.elements.pop(list.list.elementStart)

	// write elements
	w.data.writeElements(elements)
	w.data.writeElementCount(uint32(len(elements)))

	// write size and type
	size := uint32(w.data.offset() - list.list.start)
	w.data.writeSize(size)
	w.data.writeType(TypeList)

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
	w.elements.push(offset)

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
	// pop message and fields
	message, err := w.stack.popType(entryTypeMessage)
	if err != nil {
		return err
	}
	fields := w.fields.popTable(message.message.fieldStart)

	// write fields and count
	w.data.writeFields(fields)
	w.data.writeFieldCount(uint32(len(fields)))

	// write size and type
	size := uint32(w.data.offset() - message.message.start)
	w.data.writeSize(size)
	w.data.writeType(TypeMessage)

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
	f := field{
		tag:    fentry.field.tag,
		offset: uint32(data.data.end - message.message.start),
	}
	w.fields.insert(message.message.fieldStart, f)
	return nil
}

// Values

func (w *writer) WriteBool(v bool) error {
	start := w.data.offset()

	if v {
		w.data.writeType(TypeTrue)
	} else {
		w.data.writeType(TypeFalse)
	}

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteByte(v byte) error {
	start := w.data.offset()

	w.data.writeByte(v)
	w.data.writeType(TypeByte)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteInt8(v int8) error {
	start := w.data.offset()

	w.data.writeInt8(v)
	w.data.writeType(TypeInt8)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteInt16(v int16) error {
	start := w.data.offset()

	w.data.writeInt16(v)
	w.data.writeType(TypeInt16)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteInt32(v int32) error {
	start := w.data.offset()

	w.data.writeInt32(v)
	w.data.writeType(TypeInt32)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteInt64(v int64) error {
	start := w.data.offset()

	w.data.writeInt64(v)
	w.data.writeType(TypeInt64)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteUInt8(v uint8) error {
	start := w.data.offset()

	w.data.writeUInt8(v)
	w.data.writeType(TypeUInt8)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteUInt16(v uint16) error {
	start := w.data.offset()

	w.data.writeUInt16(v)
	w.data.writeType(TypeUInt16)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteUInt32(v uint32) error {
	start := w.data.offset()

	w.data.writeUInt32(v)
	w.data.writeType(TypeUInt32)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteUInt64(v uint64) error {
	start := w.data.offset()

	w.data.writeUInt64(v)
	w.data.writeType(TypeUInt64)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteFloat32(v float32) error {
	start := w.data.offset()

	w.data.writeFloat32(v)
	w.data.writeType(TypeFloat32)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteFloat64(v float64) error {
	start := w.data.offset()

	w.data.writeFloat64(v)
	w.data.writeType(TypeFloat64)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteBytes(v []byte) error {
	start := w.data.offset()

	size := uint32(len(v))
	w.data.writeBytes(v)
	w.data.writeSize(size)
	w.data.writeType(TypeBytes)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}

func (w *writer) WriteString(v string) error {
	start := w.data.offset()

	size := uint32(len(v) + 1) // plus zero byte
	w.data.writeString(v)
	w.data.writeStringZero()
	w.data.writeSize(size)
	w.data.writeType(TypeString)

	end := w.data.offset()
	w.stack.pushData(start, end)
	return nil
}
