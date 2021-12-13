package protocol

import "fmt"

const WriteBufferSize = 4096

// Writer writes objects to a buffer.
type Writer interface {
	End() ([]byte, error)
	Reset()

	// List

	BeginList() error
	BeginElement() error
	EndList() error

	// Struct

	BeginStruct() error
	BeginField(tag uint16) error
	EndStruct() error

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
	state writerState
	err   error

	data     writeBuffer   // data buffer
	stack    writeStack    // nested objects (structs, lists)
	fields   writeFields   // field buffer
	elements writeElements // element buffer
}

func newWriter() *writer {
	return &writer{
		state: writerValue,
	}
}

func (w *writer) End() ([]byte, error) {
	switch {
	case w.state != writerValue:
		return nil, fmt.Errorf("end: cannot end incomplete writer, state=%v", w.state)
	case len(w.stack) > 0:
		return nil, fmt.Errorf("end: cannot end incomplete writer, stack size=%v", len(w.stack))
	}

	data := w.data
	w.Reset()
	return data, nil
}

func (w *writer) Reset() {
	w.data = w.data[:0]
	w.stack = w.stack[:0]
	w.fields = w.fields[:0]
	w.elements = w.elements[:0]
	w.state = writerValue
}

// List

func (w *writer) BeginList() error {
	if err := w.beginData("begin list: not data writer, state=%v"); err != nil {
		return err
	}

	// get offsets
	dataOffset := w.data.offset()
	elemOffset := w.elements.offset()

	// make list
	obj := writeObject{
		type_:         writeObjectList,
		dataOffset:    dataOffset,
		elementOffset: elemOffset,
	}

	// push list
	w.state = writerList
	w.stack.push(obj)
	return nil
}

func (w *writer) BeginElement() error {
	if w.state != writerList {
		return fmt.Errorf("write element: not list writer, state=%v", w.state)
	}

	// get list
	list := w.stack.last()

	// compute element data offset
	// relative to list start
	element := uint32(w.data.offset() - list.dataOffset)

	// push element
	w.elements.push(element)

	// await element data
	w.state = writerElement
	return nil
}

func (w *writer) EndList() error {
	if w.state != writerList {
		return fmt.Errorf("end list: not list writer, state=%v", w.state)
	}

	// pop list
	list := w.stack.pop()
	elements := w.elements.pop(list.elementOffset)

	// write elements and count
	w.data.writeElements(elements)
	w.data.writeElementCount(uint32(len(elements)))

	// compute data size
	offset := w.data.offset()
	size := uint32(offset - list.dataOffset)

	// write size and type
	w.data.writeSize(size)
	w.data.writeType(TypeList)

	// done
	w.state = writerValue
	return nil
}

// Struct

func (w *writer) BeginStruct() error {
	if err := w.beginData("begin struct: not data writer, state=%v"); err != nil {
		return err
	}

	// get offsets
	dataOffset := w.data.offset()
	fieldOffset := w.fields.offset()

	// make struct
	obj := writeObject{
		type_:       writeObjectStruct,
		dataOffset:  dataOffset,
		fieldOffset: fieldOffset,
	}

	// push struct
	w.state = writerStruct
	w.stack.push(obj)
	return nil
}

func (w *writer) BeginField(tag uint16) error {
	if w.state != writerStruct {
		return fmt.Errorf("write field: not struct writer, state=%v", w.state)
	}

	// get struct
	struct_ := w.stack.last()

	// compute field data offset
	// relative to struct start
	offset := uint32(w.data.offset() - struct_.dataOffset)

	// insert field sorted
	field := writeField{
		tag:    tag,
		offset: offset,
	}
	w.fields.insert(struct_.fieldOffset, field)

	// await field data
	w.state = writerField
	return nil
}

func (w *writer) EndStruct() error {
	if w.state != writerStruct {
		return fmt.Errorf("end struct: not struct writer, state=%v", w.state)
	}

	// pop struct
	struct_ := w.stack.pop()
	fields := w.fields.pop(struct_.fieldOffset)

	// write fields and count
	w.data.writeFields(fields)
	w.data.writeFieldCount(uint32(len(fields)))

	// compute data size
	offset := w.data.offset()
	size := uint32(offset - struct_.dataOffset)

	// write size and type
	w.data.writeSize(size)
	w.data.writeType(TypeStruct)

	// done
	w.state = writerValue
	return nil
}

// Data

func (w *writer) WriteBool(v bool) error {
	if err := w.beginData("write bool: not data writer, state=%v"); err != nil {
		return err
	}

	if v {
		w.data.writeType(TypeTrue)
	} else {
		w.data.writeType(TypeFalse)
	}
	return nil
}

func (w *writer) WriteByte(v byte) error {
	if err := w.beginData("write byte: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeByte(v)
	w.data.writeType(TypeByte)
	return nil
}

func (w *writer) WriteInt8(v int8) error {
	if err := w.beginData("write int8: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeInt8(v)
	w.data.writeType(TypeInt8)
	return nil
}

func (w *writer) WriteInt16(v int16) error {
	if err := w.beginData("write int16: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeInt16(v)
	w.data.writeType(TypeInt16)
	return nil
}

func (w *writer) WriteInt32(v int32) error {
	if err := w.beginData("write int32: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeInt32(v)
	w.data.writeType(TypeInt32)
	return nil
}

func (w *writer) WriteInt64(v int64) error {
	if err := w.beginData("write int64: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeInt64(v)
	w.data.writeType(TypeInt64)
	return nil
}

func (w *writer) WriteUInt8(v uint8) error {
	if err := w.beginData("write uint8: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeUInt8(v)
	w.data.writeType(TypeUInt8)
	return nil
}

func (w *writer) WriteUInt16(v uint16) error {
	if err := w.beginData("write uint16: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeUInt16(v)
	w.data.writeType(TypeUInt16)
	return nil
}

func (w *writer) WriteUInt32(v uint32) error {
	if err := w.beginData("write uint32: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeUInt32(v)
	w.data.writeType(TypeUInt32)
	return nil
}

func (w *writer) WriteUInt64(v uint64) error {
	if err := w.beginData("write uint64: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeUInt64(v)
	w.data.writeType(TypeUInt64)
	return nil
}

func (w *writer) WriteFloat32(v float32) error {
	if err := w.beginData("write float32: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeFloat32(v)
	w.data.writeType(TypeFloat32)
	return nil
}

func (w *writer) WriteFloat64(v float64) error {
	if err := w.beginData("write float64: not data writer, state=%v"); err != nil {
		return err
	}

	w.data.writeFloat64(v)
	w.data.writeType(TypeFloat64)
	return nil
}

func (w *writer) WriteBytes(v []byte) error {
	if err := w.beginData("write bytes: not data writer, state=%v"); err != nil {
		return err
	}

	size := uint32(len(v))
	w.data.writeBytes(v)
	w.data.writeSize(size)
	w.data.writeType(TypeBytes)
	return nil
}

func (w *writer) WriteString(v string) error {
	if err := w.beginData("write string: not data writer, state=%v"); err != nil {
		return err
	}

	size := uint32(len(v) + 1) // plus zero byte
	w.data.writeString(v)
	w.data.writeStringZero()
	w.data.writeSize(size)
	w.data.writeType(TypeString)
	return nil
}

// private

func (w *writer) beginData(msg string) error {
	switch w.state {
	case writerValue, writerField, writerElement:
		// ok, data expected
		return nil
	}

	return fmt.Errorf(msg, w.state)
}

// state

type writerState int

const (
	writerValue   writerState = iota // write any value
	writerStruct                     // write struct
	writerField                      // write struct field data
	writerList                       // write list
	writerElement                    // write list element data
)

func (s writerState) String() string {
	switch s {
	case writerValue:
		return "value"
	case writerStruct:
		return "struct"
	case writerField:
		return "field"
	case writerList:
		return "list"
	case writerElement:
		return "element"
	}
	return ""
}
