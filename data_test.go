package spec

import (
	"fmt"
	"math"

	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

// TestMessage

type TestMessage struct {
	m Message
}

func NewTestMessage(b []byte) TestMessage {
	m := NewMessage(b)
	return TestMessage{m}
}

func ReadTestMessage(b []byte) (TestMessage, int, error) {
	m, n, err := ReadMessage(b)
	if err != nil {
		return TestMessage{}, n, err
	}
	return TestMessage{m}, n, nil
}

func WriteTestMessage(w *Writer) TestMessageWriter {
	w.BeginMessage()
	return TestMessageWriter{w}
}

func (m TestMessage) Bool() bool {
	return m.m.Bool(1)
}
func (m TestMessage) Int8() int8 {
	return m.m.Int8(10)
}
func (m TestMessage) Int16() int16 {
	return m.m.Int16(11)
}
func (m TestMessage) Int32() int32 {
	return m.m.Int32(12)
}
func (m TestMessage) Int64() int64 {
	return m.m.Int64(13)
}
func (m TestMessage) Uint8() uint8 {
	return m.m.Uint8(20)
}
func (m TestMessage) Uint16() uint16 {
	return m.m.Uint16(21)
}
func (m TestMessage) Uint32() uint32 {
	return m.m.Uint32(22)
}
func (m TestMessage) Uint64() uint64 {
	return m.m.Uint64(23)
}
func (m TestMessage) U128() u128.U128 {
	return m.m.U128(24)
}
func (m TestMessage) U256() u256.U256 {
	return m.m.U256(25)
}
func (m TestMessage) Float32() float32 {
	return m.m.Float32(30)
}
func (m TestMessage) Float64() float64 {
	return m.m.Float64(31)
}
func (m TestMessage) String() string {
	return m.m.String(40)
}
func (m TestMessage) Bytes() []byte {
	return m.m.Bytes(41)
}
func (m TestMessage) List() List[int64] {
	b := m.m.Field(50)
	return NewList[int64](b, ReadInt64)
}
func (m TestMessage) Messages() List[TestSubmessage] {
	b := m.m.Field(51)
	return NewList[TestSubmessage](b, ReadTestSubmessage)
}
func (m TestMessage) Strings() List[string] {
	b := m.m.Field(52)
	return NewList[string](b, ReadString)
}
func (m TestMessage) Struct() TestStruct {
	data := m.m.Field(60)
	v, _ := ReadTestStruct(data)
	return v
}

// TestMessageWriter

type TestMessageWriter struct {
	w *Writer
}

func (w TestMessageWriter) End() error {
	return w.w.EndMessage()
}
func (w TestMessageWriter) Bool(v bool) error {
	w.w.Bool(v)
	return w.w.Field(1)
}
func (w TestMessageWriter) Int8(v int8) error {
	w.w.Int8(v)
	return w.w.Field(10)
}
func (w TestMessageWriter) Int16(v int16) error {
	w.w.Int16(v)
	return w.w.Field(11)
}
func (w TestMessageWriter) Int32(v int32) error {
	w.w.Int32(v)
	return w.w.Field(12)
}
func (w TestMessageWriter) Int64(v int64) error {
	w.w.Int64(v)
	return w.w.Field(13)
}
func (w TestMessageWriter) Uint8(v uint8) error {
	w.w.Uint8(v)
	return w.w.Field(20)
}
func (w TestMessageWriter) Uint16(v uint16) error {
	w.w.Uint16(v)
	return w.w.Field(21)
}
func (w TestMessageWriter) Uint32(v uint32) error {
	w.w.Uint32(v)
	return w.w.Field(22)
}
func (w TestMessageWriter) Uint64(v uint64) error {
	w.w.Uint64(v)
	return w.w.Field(23)
}
func (w TestMessageWriter) U128(v u128.U128) error {
	w.w.U128(v)
	return w.w.Field(24)
}
func (w TestMessageWriter) U256(v u256.U256) error {
	w.w.U256(v)
	return w.w.Field(25)
}
func (w TestMessageWriter) Float32(v float32) error {
	w.w.Float32(v)
	return w.w.Field(30)
}
func (w TestMessageWriter) Float64(v float64) error {
	w.w.Float64(v)
	return w.w.Field(31)
}
func (w TestMessageWriter) String(v string) error {
	w.w.String(v)
	return w.w.Field(40)
}
func (w TestMessageWriter) Bytes(v []byte) error {
	w.w.Bytes(v)
	return w.w.Field(41)
}
func (w TestMessageWriter) BeginList() ListValueWriter[int64] {
	return WriteValueList(w.w, w.w.Int64)
}
func (w TestMessageWriter) EndList() error {
	w.w.EndList()
	return w.w.Field(50)
}
func (w TestMessageWriter) BeginMessages() ListWriter[TestSubmessageWriter] {
	return WriteList(w.w, WriteTestSubmessage)
}
func (w TestMessageWriter) EndMessages() error {
	w.w.EndList()
	return w.w.Field(51)
}
func (w TestMessageWriter) BeginStrings() ListValueWriter[string] {
	return WriteValueList(w.w, w.w.String)
}
func (w TestMessageWriter) EndStrings() error {
	w.w.EndList()
	return w.w.Field(52)
}
func (w TestMessageWriter) Struct(v TestStruct) error {
	// return w.w.Field(60)
	return nil
}

// TestSubmessage

type TestSubmessage struct {
	m Message
}

func NewTestSubmessage(b []byte) TestSubmessage {
	m := NewMessage(b)
	return TestSubmessage{m}
}

func ReadTestSubmessage(b []byte) (TestSubmessage, int, error) {
	m, n, err := ReadMessage(b)
	if err != nil {
		return TestSubmessage{}, n, err
	}
	return TestSubmessage{m}, n, nil
}

func WriteTestSubmessage(w *Writer) TestSubmessageWriter {
	w.BeginMessage()
	return TestSubmessageWriter{w}
}

func (m TestSubmessage) Int8() int8 {
	return m.m.Int8(1)
}
func (m TestSubmessage) Int16() int16 {
	return m.m.Int16(2)
}
func (m TestSubmessage) Int32() int32 {
	return m.m.Int32(3)
}
func (m TestSubmessage) Int64() int64 {
	return m.m.Int64(4)
}

// TestSubmessageWriter

type TestSubmessageWriter struct {
	w *Writer
}

func (w TestSubmessageWriter) End() error {
	return w.w.EndMessage()
}

func (w TestSubmessageWriter) Int8(v int8) error {
	w.w.Int8(v)
	return w.w.Field(1)
}
func (w TestSubmessageWriter) Int16(v int16) error {
	w.w.Int16(v)
	return w.w.Field(2)
}
func (w TestSubmessageWriter) Int32(v int32) error {
	w.w.Int32(v)
	return w.w.Field(3)
}
func (w TestSubmessageWriter) Int64(v int64) error {
	w.w.Int64(v)
	return w.w.Field(4)
}

// Objects

type TestObject struct {
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

	List     []int64          `tag:"50"`
	Messages []*TestSubobject `tag:"51"`
	Strings  []string         `tag:"52"`

	// Struct TestStruct `tag:"60"`
}

func newTestObject() *TestObject {
	list := make([]int64, 0, 10)
	for i := 0; i < cap(list); i++ {
		list = append(list, int64(i))
	}

	messages := make([]*TestSubobject, 0, 10)
	for i := 0; i < cap(messages); i++ {
		sub := newTestSubobject(i)
		messages = append(messages, sub)
	}

	strings := make([]string, 0, 10)
	for i := 0; i < cap(strings); i++ {
		s := fmt.Sprintf("hello, world %03d", i)
		strings = append(strings, s)
	}

	return &TestObject{
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

		// Struct: TestStruct{
		// 	X: 100,
		// 	Y: 200,
		// },
	}
}

func (m *TestObject) Read(b []byte) error {
	r, _, err := ReadTestMessage(b)
	if err != nil {
		return err
	}

	// bool:1
	m.Bool = r.Bool()

	// int:10-13
	m.Int8 = r.Int8()
	m.Int16 = r.Int16()
	m.Int32 = r.Int32()
	m.Int64 = r.Int64()

	// uint:20-22
	m.Uint8 = r.Uint8()
	m.Uint16 = r.Uint16()
	m.Uint32 = r.Uint32()
	m.Uint64 = r.Uint64()

	// u128/u256:24-25
	m.U128 = r.U128()
	m.U256 = r.U256()

	// float:30-31
	m.Float32 = r.Float32()
	m.Float64 = r.Float64()

	// string/bytes:40-41
	m.String = r.String()
	m.Bytes = r.Bytes()

	// list:50
	{
		list := r.List()
		m.List = make([]int64, 0, list.Count())

		for i := 0; i < list.Count(); i++ {
			el := list.Element(i)
			m.List = append(m.List, el)
		}
	}

	// messages:51
	{
		list := r.Messages()
		m.Messages = make([]*TestSubobject, 0, list.Count())

		for i := 0; i < list.Count(); i++ {
			data := list.ElementBytes(i)
			if len(data) == 0 {
				continue
			}

			el := &TestSubobject{}
			if err := el.Read(data); err != nil {
				return err
			}
			m.Messages = append(m.Messages, el)
		}
	}

	// strings:52
	{
		list := r.Strings()
		m.Strings = make([]string, 0, list.Count())

		for i := 0; i < list.Count(); i++ {
			s := list.Element(i)
			m.Strings = append(m.Strings, s)
		}
	}

	// struct:60
	// TODO: Uncomment
	// {
	// 	data := r.Field(60)
	// 	if err := m.Struct.Unmarshal(data); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func (m *TestObject) Write(w TestMessageWriter) error {
	w.Bool(m.Bool)

	w.Int8(m.Int8)
	w.Int16(m.Int16)
	w.Int32(m.Int32)
	w.Int64(m.Int64)

	w.Uint8(m.Uint8)
	w.Uint16(m.Uint16)
	w.Uint32(m.Uint32)
	w.Uint64(m.Uint64)

	w.U128(m.U128)
	w.U256(m.U256)

	w.Float32(m.Float32)
	w.Float64(m.Float64)

	w.String(m.String)
	w.Bytes(m.Bytes)

	if len(m.List) > 0 {
		list := w.BeginList()
		for _, value := range m.List {
			list.Next(value)
		}
		w.EndList()
	}

	if len(m.Messages) > 0 {
		list := w.BeginMessages()
		for _, msg := range m.Messages {
			next := list.BeginNext()
			msg.Write(next)
			list.EndNext()
		}
		w.EndMessages()
	}

	if len(m.Strings) > 0 {
		list := w.BeginStrings()
		for _, v := range m.Strings {
			list.Next(v)
		}
		w.EndStrings()
	}

	// struct:60
	// TODO: Uncomment
	// m.Struct.Write(w)
	// w.Field(60)
	return w.End()
}

func (m *TestObject) Marshal() ([]byte, error) {
	w := NewWriter()
	mw := WriteTestMessage(w)
	if err := m.Write(mw); err != nil {
		return nil, err
	}
	return w.End()
}

// TestSubobject

type TestSubobject struct {
	Int8  int8  `tag:"1"`
	Int16 int16 `tag:"2"`
	Int32 int32 `tag:"3"`
	Int64 int64 `tag:"4"`
}

func newTestSubobject(i int) *TestSubobject {
	return &TestSubobject{
		Int8:  int8(i + 1),
		Int16: int16(i + 10),
		Int32: int32(i + 100),
		Int64: int64(i + 1000),
	}
}

func (m *TestSubobject) Read(b []byte) error {
	r, _, err := ReadMessage(b)
	if err != nil {
		return err
	}

	m.Int8 = r.Int8(1)
	m.Int16 = r.Int16(2)
	m.Int32 = r.Int32(3)
	m.Int64 = r.Int64(4)
	return nil
}

func (m TestSubobject) Write(w TestSubmessageWriter) error {
	w.Int8(m.Int8)
	w.Int16(m.Int16)
	w.Int32(m.Int32)
	w.Int64(m.Int64)
	return w.End()
}

// TestStruct

type TestStruct struct {
	X int64
	Y int64
}

func ReadTestStruct(b []byte) (TestStruct, error) {
	s := TestStruct{}
	err := s.Unmarshal(b)
	return s, err
}

func (s *TestStruct) Unmarshal(b []byte) error {
	return nil

	// TODO: Uncomment
	// r, err := ReadStruct(b)
	// switch {
	// case err != nil:
	// 	return err
	// case len(r) == 0:
	// 	return nil
	// }

	// s.Y, r, err = r.ReadInt64()
	// if err != nil {
	// 	return err
	// }
	// s.X, r, err = r.ReadInt64()
	// if err != nil {
	// 	return err
	// }
	// return nil
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
