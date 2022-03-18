package spec

import (
	"fmt"
	"math"

	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

// Objects

type TestObject struct {
	Bool bool `tag:"1"`
	Byte byte `tag:"2"`

	Int32 int32 `tag:"10"`
	Int64 int64 `tag:"11"`

	Uint32 uint32 `tag:"20"`
	Uint64 uint64 `tag:"21"`

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
		Byte: math.MaxInt8,

		Int32: math.MaxInt32,
		Int64: math.MaxInt64,

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

	m.Bool = r.Bool()
	m.Byte = r.Byte()

	m.Int32 = r.Int32()
	m.Int64 = r.Int64()

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
	w.Byte(m.Byte)

	w.Int32(m.Int32)
	w.Int64(m.Int64)

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
			list.Element(value)
		}
		w.EndList()
	}

	if len(m.Messages) > 0 {
		list := w.BeginMessages()
		for _, msg := range m.Messages {
			next := list.BeginElement()
			msg.Write(next)
			list.EndElement()
		}
		w.EndMessages()
	}

	if len(m.Strings) > 0 {
		list := w.BeginStrings()
		for _, v := range m.Strings {
			list.Element(v)
		}
		w.EndStrings()
	}

	// struct:60
	// TODO: Uncomment
	// m.Struct.Write(b)
	// w.Field(60)
	return w.End()
}

func (m *TestObject) Marshal() ([]byte, error) {
	w := NewWriter()
	mw := BeginTestMessage(w)
	if err := m.Write(mw); err != nil {
		return nil, err
	}
	return w.End()
}

// TestSubobject

type TestSubobject struct {
	Byte  byte  `tag:"1"`
	Int32 int32 `tag:"2"`
	Int64 int64 `tag:"3"`
}

func newTestSubobject(i int) *TestSubobject {
	return &TestSubobject{
		Byte:  byte(i + 1),
		Int32: int32(i + 100),
		Int64: int64(i + 1000),
	}
}

func (m *TestSubobject) Read(b []byte) error {
	r, _, err := ReadMessage(b)
	if err != nil {
		return err
	}

	m.Byte = r.Byte(1)
	m.Int32 = r.Int32(2)
	m.Int64 = r.Int64(3)
	return nil
}

func (m TestSubobject) Write(w TestSubmessageWriter) error {
	w.Byte(m.Byte)
	w.Int32(m.Int32)
	w.Int64(m.Int64)
	return w.End()
}
