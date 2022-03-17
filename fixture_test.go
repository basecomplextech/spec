package spec

import (
	"fmt"
	"math"

	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

// Fixture messages

type TestMessage struct {
	Bool bool `tag:"1"`

	Int8  int8  `tag:"10"`
	Int16 int16 `tag:"11"`
	Int32 int32 `tag:"12"`
	Int64 int64 `tag:"13"`

	Uint8  uint8  `tag:"20"`
	Uint16 uint16 `tag:"21"`
	Uint32 uint32 `tag:"22"`
	Uint64 uint64 `tag:"23"`

	U128 u128.U128 `tag:"24"`
	U256 u256.U256 `tag:"25"`

	Float32 float32 `tag:"30"`
	Float64 float64 `tag:"31"`

	String string `tag:"40"`
	Bytes  []byte `tag:"41"`

	List     []int64           `tag:"50"`
	Messages []*TestSubMessage `tag:"51"`
	Strings  []string          `tag:"52"`

	Struct TestStruct `tag:"60"`
}

func newTestMessage() *TestMessage {
	list := make([]int64, 0, 10)
	for i := 0; i < cap(list); i++ {
		list = append(list, int64(i))
	}

	messages := make([]*TestSubMessage, 0, 10)
	for i := 0; i < cap(messages); i++ {
		sub := newTestSubMessage(i)
		messages = append(messages, sub)
	}

	strings := make([]string, 0, 10)
	for i := 0; i < cap(strings); i++ {
		s := fmt.Sprintf("hello, world %03d", i)
		strings = append(strings, s)
	}

	return &TestMessage{
		Bool: true,

		Int8:  math.MaxInt8,
		Int16: math.MaxInt16,
		Int32: math.MaxInt32,
		Int64: math.MaxInt64,

		Uint8:  math.MaxUint8,
		Uint16: math.MaxUint16,
		Uint32: math.MaxUint32,
		Uint64: math.MaxUint64,

		Float32: math.MaxFloat32,
		Float64: math.MaxFloat64,

		String: "hello, world",
		Bytes:  []byte("goodbye, world"),

		List:     list,
		Messages: messages,
		Strings:  strings,

		Struct: TestStruct{
			X: 100,
			Y: 200,
		},
	}
}

// TestSubMessage

type TestSubMessage struct {
	Int8  int8  `tag:"1"`
	Int16 int16 `tag:"2"`
	Int32 int32 `tag:"3"`
	Int64 int64 `tag:"4"`
}

func newTestSubMessage(i int) *TestSubMessage {
	return &TestSubMessage{
		Int8:  int8(i + 1),
		Int16: int16(i + 10),
		Int32: int32(i + 100),
		Int64: int64(i + 1000),
	}
}

// TestStruct

type TestStruct struct {
	X int64
	Y int64
}

// Marshal

func (m *TestMessage) Marshal() ([]byte, error) {
	return Write(m)
}

func (m *TestMessage) MarshalTo(buf []byte) ([]byte, error) {
	return WriteTo(m, buf)
}

// Unmarshal

func (m *TestMessage) Unmarshal(b []byte) error {
	r, err := NewMessageReader(b)
	if err != nil {
		return err
	}

	// bool:1
	m.Bool, err = r.ReadBool(1)
	if err != nil {
		return err
	}

	// int:10-13
	m.Int8, err = r.ReadInt8(10)
	if err != nil {
		return err
	}
	m.Int16, err = r.ReadInt16(11)
	if err != nil {
		return err
	}
	m.Int32, err = r.ReadInt32(12)
	if err != nil {
		return err
	}
	m.Int64, err = r.ReadInt64(13)
	if err != nil {
		return err
	}

	// uint:20-22
	m.Uint8, err = r.ReadUint8(20)
	if err != nil {
		return err
	}
	m.Uint16, err = r.ReadUint16(21)
	if err != nil {
		return err
	}
	m.Uint32, err = r.ReadUint32(22)
	if err != nil {
		return err
	}
	m.Uint64, err = r.ReadUint64(23)
	if err != nil {
		return err
	}

	// u128/u256:24-25
	m.U128, err = r.ReadU128(24)
	if err != nil {
		return err
	}
	m.U256, err = r.ReadU256(25)
	if err != nil {
		return err
	}

	// float:30-31
	m.Float32, err = r.ReadFloat32(30)
	if err != nil {
		return err
	}
	m.Float64, err = r.ReadFloat64(31)
	if err != nil {
		return err
	}

	// string/bytes:40-41
	m.String, err = r.ReadString(40)
	if err != nil {
		return err
	}
	m.Bytes, err = r.ReadBytes(41)
	if err != nil {
		return err
	}

	// list:50
	{
		list, err := r.ReadList(50)
		if err != nil {
			return err
		}

		m.List = make([]int64, 0, list.Len())
		for i := 0; i < list.Len(); i++ {
			val, err := list.ReadInt64(i)
			if err != nil {
				return err
			}
			m.List = append(m.List, val)
		}
	}

	// messages:51
	{
		list, err := r.ReadList(51)
		if err != nil {
			return err
		}

		m.Messages = make([]*TestSubMessage, 0, list.Len())
		for i := 0; i < list.Len(); i++ {
			data, err := list.Read(i)
			if err != nil {
				return err
			}
			if len(data) == 0 {
				continue
			}

			val := &TestSubMessage{}
			if err := val.Unmarshal(data); err != nil {
				return err
			}
			m.Messages = append(m.Messages, val)
		}
	}

	// strings:52
	{
		list, err := r.ReadList(52)
		if err != nil {
			return err
		}

		m.Strings = make([]string, 0, list.Len())
		for i := 0; i < list.Len(); i++ {
			s, err := list.ReadString(i)
			if err != nil {
				return err
			}
			m.Strings = append(m.Strings, s)
		}
	}

	// struct:60
	{
		data, err := r.Read(60)
		if err != nil {
			return err
		}
		if err := m.Struct.Unmarshal(data); err != nil {
			return err
		}
	}
	return nil
}

func (m *TestSubMessage) Unmarshal(b []byte) error {
	r, err := NewMessageReader(b)
	if err != nil {
		return err
	}

	// int:1-4
	m.Int8, err = r.ReadInt8(1)
	if err != nil {
		return err
	}
	m.Int16, err = r.ReadInt16(2)
	if err != nil {
		return err
	}
	m.Int32, err = r.ReadInt32(3)
	if err != nil {
		return err
	}
	m.Int64, err = r.ReadInt64(4)
	if err != nil {
		return err
	}
	return nil
}

func (s *TestStruct) Unmarshal(b []byte) error {
	r, err := ReadStruct(b)
	switch {
	case err != nil:
		return err
	case len(r) == 0:
		return nil
	}

	s.Y, r, err = r.ReadInt64()
	if err != nil {
		return err
	}
	s.X, r, err = r.ReadInt64()
	if err != nil {
		return err
	}
	return nil
}

// Write

func (m TestMessage) Write(w *Writer) error {
	if err := w.BeginMessage(); err != nil {
		return err
	}

	// bool:1
	w.Bool(m.Bool)
	w.Field(1)

	// int:10-13
	w.Int8(m.Int8)
	w.Field(10)
	w.Int16(m.Int16)
	w.Field(11)
	w.Int32(m.Int32)
	w.Field(12)
	w.Int64(m.Int64)
	w.Field(13)

	// uint:20-23
	w.Uint8(m.Uint8)
	w.Field(20)
	w.Uint16(m.Uint16)
	w.Field(21)
	w.Uint32(m.Uint32)
	w.Field(22)
	w.Uint64(m.Uint64)
	w.Field(23)

	// u128/u256:24-25
	w.U128(m.U128)
	w.Field(24)
	w.U256(m.U256)
	w.Field(25)

	// float:30-31
	w.Float32(m.Float32)
	w.Field(30)
	w.Float64(m.Float64)
	w.Field(31)

	// bytes:40-41
	w.String(m.String)
	w.Field(40)
	w.Bytes(m.Bytes)
	w.Field(41)

	// list:50
	if len(m.List) > 0 {
		w.BeginList()
		for _, val := range m.List {
			w.Int64(val)
			w.Element()
		}
		w.EndList()
		w.Field(50)
	}

	// messages:51
	if len(m.Messages) > 0 {
		w.BeginList()
		for _, val := range m.Messages {
			val.Write(w)
			w.Element()
		}
		w.EndList()
		w.Field(51)
	}

	// strings:52
	if len(m.Strings) > 0 {
		w.BeginList()
		for _, val := range m.Strings {
			w.String(val)
			w.Element()
		}
		w.EndList()
		w.Field(52)
	}

	// struct:60
	m.Struct.Write(w)
	w.Field(60)

	return w.EndMessage()
}

func (m TestSubMessage) Write(w *Writer) error {
	if err := w.BeginMessage(); err != nil {
		return err
	}

	// int:1-4
	w.Int8(m.Int8)
	w.Field(1)
	w.Int16(m.Int16)
	w.Field(2)
	w.Int32(m.Int32)
	w.Field(3)
	w.Int64(m.Int64)
	w.Field(4)

	return w.EndMessage()
}

func (s TestStruct) Write(w *Writer) error {
	if err := w.BeginStruct(); err != nil {
		return err
	}

	w.Int64(s.X)
	w.StructField()

	w.Int64(s.Y)
	w.StructField()

	return w.EndStruct()
}

// Data

type TestMessageData struct{ d MessageData }

func (d TestMessageData) Bool() bool         { return d.d.Bool(1) }
func (d TestMessageData) Int8() int8         { return d.d.Int8(10) }
func (d TestMessageData) Int16() int16       { return d.d.Int16(11) }
func (d TestMessageData) Int32() int32       { return d.d.Int32(12) }
func (d TestMessageData) Int64() int64       { return d.d.Int64(13) }
func (d TestMessageData) Uint8() uint8       { return d.d.Uint8(20) }
func (d TestMessageData) Uint16() uint16     { return d.d.Uint16(21) }
func (d TestMessageData) Uint32() uint32     { return d.d.Uint32(22) }
func (d TestMessageData) Uint64() uint64     { return d.d.Uint64(23) }
func (d TestMessageData) U128() u128.U128    { return d.d.U128(24) }
func (d TestMessageData) U256() u256.U256    { return d.d.U256(25) }
func (d TestMessageData) Float32() float32   { return d.d.Float32(30) }
func (d TestMessageData) Float64() float64   { return d.d.Float64(31) }
func (d TestMessageData) String() string     { return d.d.String(40) }
func (d TestMessageData) Bytes() []byte      { return d.d.Bytes(41) }
func (d TestMessageData) List() ListData     { return d.d.List(50) }
func (d TestMessageData) Messages() ListData { return d.d.List(51) }
func (d TestMessageData) Strings() ListData  { return d.d.List(52) }

func (d TestMessageData) Struct() TestStruct {
	data := d.d.Field(60)
	v, _ := readTestStruct(data)
	return v
}

func getTestMessageData(b []byte) (TestMessageData, error) {
	m, err := NewMessageData(b)
	if err != nil {
		return TestMessageData{}, err
	}
	return TestMessageData{m}, nil
}

func readTestMessageData(b []byte) (TestMessageData, error) {
	m, err := ReadMessageData(b)
	if err != nil {
		return TestMessageData{}, err
	}
	return TestMessageData{m}, nil
}

type TestSubMessageData struct{ d MessageData }

func (d TestSubMessageData) Int8() int8   { return d.d.Int8(1) }
func (d TestSubMessageData) Int16() int16 { return d.d.Int16(2) }
func (d TestSubMessageData) Int32() int32 { return d.d.Int32(3) }
func (d TestSubMessageData) Int64() int64 { return d.d.Int64(4) }

func getTestSubMessageData(b []byte) (TestSubMessageData, error) {
	m, err := NewMessageData(b)
	if err != nil {
		return TestSubMessageData{}, err
	}
	return TestSubMessageData{m}, nil
}

func readTestSubMessageData(b []byte) (TestSubMessageData, error) {
	m, err := ReadMessageData(b)
	if err != nil {
		return TestSubMessageData{}, err
	}
	return TestSubMessageData{m}, nil
}

func readTestStruct(b []byte) (TestStruct, error) {
	s := TestStruct{}
	err := s.Unmarshal(b)
	return s, err
}
