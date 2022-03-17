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

func NewTestMessage(b []byte) (TestMessage, error) {
	m, err := NewMessage(b)
	if err != nil {
		return TestMessage{}, err
	}
	return TestMessage{m}, nil
}

func ReadTestMessage(b []byte) (TestMessage, error) {
	m, err := ReadMessage(b)
	if err != nil {
		return TestMessage{}, err
	}
	return TestMessage{m}, nil
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
func (m TestMessage) List() List {
	return m.m.List(50)
}
func (m TestMessage) Messages() List {
	return m.m.List(51)
}
func (m TestMessage) Strings() List {
	return m.m.List(52)
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
func (w TestMessageWriter) List(v []int64) error {
	if len(v) == 0 {
		return nil
	}

	w.w.BeginList()
	for _, val := range v {
		w.w.Int64(val)
		w.w.Element()
	}
	w.w.EndList()
	return w.w.Field(50)
}
func (w TestMessageWriter) Messages(v []*TestSubobject) error {
	if len(v) == 0 {
		return nil
	}
	w.w.BeginList()
	for _, val := range v {
		val.Write(w.w)
		w.w.Element()
	}
	w.w.EndList()
	return w.w.Field(51)
}
func (w TestMessageWriter) Strings(v []string) error {
	if len(v) == 0 {
		return nil
	}

	w.w.BeginList()
	for _, val := range v {
		w.w.String(val)
		w.w.Element()
	}
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

func NewTestSubmessage(b []byte) (TestSubmessage, error) {
	m, err := NewMessage(b)
	if err != nil {
		return TestSubmessage{}, err
	}
	return TestSubmessage{m}, nil
}

func ReadTestSubmessage(b []byte) (TestSubmessage, error) {
	m, err := ReadMessage(b)
	if err != nil {
		return TestSubmessage{}, err
	}
	return TestSubmessage{m}, nil
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

func (m *TestObject) Marshal() ([]byte, error) {
	return Write(m)
}

func (m *TestObject) MarshalTo(buf []byte) ([]byte, error) {
	return WriteTo(m, buf)
}

func (m *TestObject) Unmarshal(b []byte) error {
	r, err := ReadMessage(b)
	if err != nil {
		return err
	}

	// bool:1
	m.Bool = r.Bool(1)

	// int:10-13
	m.Int8 = r.Int8(10)
	m.Int16 = r.Int16(11)
	m.Int32 = r.Int32(12)
	m.Int64 = r.Int64(13)

	// uint:20-22
	m.Uint8 = r.Uint8(20)
	m.Uint16 = r.Uint16(21)
	m.Uint32 = r.Uint32(22)
	m.Uint64 = r.Uint64(23)

	// u128/u256:24-25
	m.U128 = r.U128(24)
	m.U256 = r.U256(25)

	// float:30-31
	m.Float32 = r.Float32(30)
	m.Float64 = r.Float64(31)

	// string/bytes:40-41
	m.String = r.String(40)
	m.Bytes = r.Bytes(41)

	// list:50
	{
		list := r.List(50)

		m.List = make([]int64, 0, list.Count())
		for i := 0; i < list.Count(); i++ {
			val := list.Int64(i)
			m.List = append(m.List, val)
		}
	}

	// messages:51
	{
		list := r.List(51)

		m.Messages = make([]*TestSubobject, 0, list.Count())
		for i := 0; i < list.Count(); i++ {
			data := list.Element(i)
			if len(data) == 0 {
				continue
			}

			val := &TestSubobject{}
			if err := val.Unmarshal(data); err != nil {
				return err
			}
			m.Messages = append(m.Messages, val)
		}
	}

	// strings:52
	{
		list := r.List(52)

		m.Strings = make([]string, 0, list.Count())
		for i := 0; i < list.Count(); i++ {
			s := list.String(i)
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

func (m *TestObject) Write(writer *Writer) error {
	w := WriteTestMessage(writer)

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

	w.List(m.List)
	w.Messages(m.Messages)
	w.Strings(m.Strings)

	// struct:60
	// TODO: Uncomment
	// m.Struct.Write(w)
	// w.Field(60)
	return w.w.EndMessage()
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

func (m *TestSubobject) Marshal() ([]byte, error) {
	return Write(m)
}

func (m *TestSubobject) MarshalTo(b []byte) ([]byte, error) {
	return WriteTo(m, b)
}

func (m *TestSubobject) Unmarshal(b []byte) error {
	r, err := ReadMessage(b)
	if err != nil {
		return err
	}

	m.Int8 = r.Int8(1)
	m.Int16 = r.Int16(2)
	m.Int32 = r.Int32(3)
	m.Int64 = r.Int64(4)
	return nil
}

func (m TestSubobject) Write(writer *Writer) error {
	w := WriteTestSubmessage(writer)

	// int:1-4
	w.Int8(m.Int8)
	w.Int16(m.Int16)
	w.Int32(m.Int32)
	w.Int64(m.Int64)

	return w.w.EndMessage()
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
