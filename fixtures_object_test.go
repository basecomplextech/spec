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

	Submessage *TestSubobject       `tag:"50"`
	List       []int64              `tag:"51"`
	Messages   []*TestObjectElement `tag:"52"`
	Strings    []string             `tag:"53"`

	// Struct TestStruct `tag:"60"`
}

func newTestObject() *TestObject {
	list := make([]int64, 0, 10)
	for i := 0; i < cap(list); i++ {
		list = append(list, int64(i))
	}

	messages := make([]*TestObjectElement, 0, 10)
	for i := 0; i < cap(messages); i++ {
		sub := newTestObjectElement(i)
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

func (m *TestObject) Decode(b []byte) error {
	msg, _, err := DecodeTestMessage(b)
	if err != nil {
		return err
	}

	m.Bool = msg.Bool()
	m.Byte = msg.Byte()

	m.Int32 = msg.Int32()
	m.Int64 = msg.Int64()

	m.Uint32 = msg.Uint32()
	m.Uint64 = msg.Uint64()

	// u128/u256:24-25
	m.U128 = msg.U128()
	m.U256 = msg.U256()

	// float:30-31
	m.Float32 = msg.Float32()
	m.Float64 = msg.Float64()

	// string/bytes:40-41
	m.String = msg.String()
	m.Bytes = msg.Bytes()

	// submessage
	if p := msg.Submessage().Bytes(); p != nil {
		sub := &TestSubobject{}
		if err := sub.Decode(p); err != nil {
			return err
		}

		m.Submessage = sub
	}

	// list:51
	{
		list := msg.List()
		m.List = make([]int64, 0, list.Count())

		for i := 0; i < list.Count(); i++ {
			el := list.Element(i)
			m.List = append(m.List, el)
		}
	}

	// messages:52
	{
		list := msg.Messages()
		m.Messages = make([]*TestObjectElement, 0, list.Count())

		for i := 0; i < list.Count(); i++ {
			data := list.ElementBytes(i)
			if len(data) == 0 {
				continue
			}

			el := &TestObjectElement{}
			if err := el.Decode(data); err != nil {
				return err
			}
			m.Messages = append(m.Messages, el)
		}
	}

	// strings:53
	{
		list := msg.Strings()
		m.Strings = make([]string, 0, list.Count())

		for i := 0; i < list.Count(); i++ {
			s := list.Element(i)
			m.Strings = append(m.Strings, s)
		}
	}

	// struct:60
	// TODO: Uncomment
	// {
	// 	data := msg.Field(60)
	// 	if err := m.Struct.Unmarshal(data); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func (m *TestObject) Encode(e TestMessageEncoder) error {
	e.Bool(m.Bool)
	e.Byte(m.Byte)

	e.Int32(m.Int32)
	e.Int64(m.Int64)

	e.Uint32(m.Uint32)
	e.Uint64(m.Uint64)

	e.U128(m.U128)
	e.U256(m.U256)

	e.Float32(m.Float32)
	e.Float64(m.Float64)

	e.String(m.String)
	e.Bytes(m.Bytes)

	if m.Submessage != nil {
		sub := e.Submessage()
		if err := m.Submessage.Encode(sub); err != nil {
			return err
		}
		if err := sub.End(); err != nil {
			return err
		}
	}

	if len(m.List) > 0 {
		list := e.List()
		for _, value := range m.List {
			if err := list.Next(value); err != nil {
				return err
			}
		}
		if err := list.End(); err != nil {
			return err
		}
	}

	if len(m.Messages) > 0 {
		list := e.Messages()
		for _, msg := range m.Messages {
			next := list.Next()
			if err := msg.Encode(next); err != nil {
				return err
			}
			if err := next.End(); err != nil {
				return err
			}
		}
		if err := list.End(); err != nil {
			return err
		}
	}

	if len(m.Strings) > 0 {
		list := e.Strings()
		for _, v := range m.Strings {
			if err := list.Next(v); err != nil {
				return err
			}
		}
		if err := list.End(); err != nil {
			return err
		}
	}

	// struct:60
	// TODO: Uncomment
	// m.Struct.Write(b)
	// e.Field(60)
	return nil
}

func (m *TestObject) Marshal() ([]byte, error) {
	e := NewEncoder()
	me := BeginTestMessage(e)
	if err := m.Encode(me); err != nil {
		return nil, err
	}
	if err := me.End(); err != nil {
		return nil, err
	}
	return e.End()
}

// TestSubobject

type TestSubobject struct {
	Int32 int32 `tag:"1"`
	Int64 int64 `tag:"2"`
}

func newTestSubobject() *TestSubobject {
	return &TestSubobject{
		Int32: 1,
		Int64: 2,
	}
}

func (m *TestSubobject) Decode(b []byte) error {
	msg, _, err := DecodeTestSubmessage(b)
	if err != nil {
		return err
	}

	m.Int32 = msg.Int32()
	m.Int64 = msg.Int64()
	return nil
}

func (m *TestSubobject) Encode(e TestSubmessageEncoder) error {
	e.Int32(m.Int32)
	e.Int64(m.Int64)
	return nil
}

// TestObjectElement

type TestObjectElement struct {
	Byte  byte  `tag:"1"`
	Int32 int32 `tag:"2"`
	Int64 int64 `tag:"3"`
}

func newTestObjectElement(i int) *TestObjectElement {
	return &TestObjectElement{
		Byte:  byte(i + 1),
		Int32: int32(i + 100),
		Int64: int64(i + 1000),
	}
}

func (m *TestObjectElement) Decode(b []byte) error {
	msg, _, err := DecodeMessage(b)
	if err != nil {
		return err
	}

	m.Byte = msg.Byte(1)
	m.Int32 = msg.Int32(2)
	m.Int64 = msg.Int64(3)
	return nil
}

func (m TestObjectElement) Encode(e TestElementEncoder) error {
	e.Byte(m.Byte)
	e.Int32(m.Int32)
	e.Int64(m.Int64)
	return nil
}