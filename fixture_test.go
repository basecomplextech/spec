package spec

import (
	"fmt"
	"math"
)

// Fixture messages

type TestMessage struct {
	Bool bool `tag:"1"`

	Int8  int8  `tag:"10"`
	Int16 int16 `tag:"11"`
	Int32 int32 `tag:"12"`
	Int64 int64 `tag:"13"`

	UInt8  uint8  `tag:"20"`
	UInt16 uint16 `tag:"21"`
	UInt32 uint32 `tag:"22"`
	UInt64 uint64 `tag:"23"`

	Float32 float32 `tag:"30"`
	Float64 float64 `tag:"31"`

	String string `tag:"40"`
	Bytes  []byte `tag:"41"`

	List     []int64           `tag:"50"`
	Messages []*TestSubMessage `tag:"51"`
	Strings  []string          `tag:"52"`
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

		UInt8:  math.MaxUint8,
		UInt16: math.MaxUint16,
		UInt32: math.MaxUint32,
		UInt64: math.MaxUint64,

		Float32: math.MaxFloat32,
		Float64: math.MaxFloat64,

		String: "hello, world",
		Bytes:  []byte("goodbye, world"),

		List:     list,
		Messages: messages,
		Strings:  strings,
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

// Read

func (msg *TestMessage) Read(b []byte) error {
	m, err := ReadMessage(b)
	if err != nil {
		return err
	}

	// bool:1
	msg.Bool = m.Bool(1)

	// int:10-13
	msg.Int8 = m.Int8(10)
	msg.Int16 = m.Int16(11)
	msg.Int32 = m.Int32(12)
	msg.Int64 = m.Int64(13)

	// uint:20-23
	msg.UInt8 = m.UInt8(20)
	msg.UInt16 = m.UInt16(21)
	msg.UInt32 = m.UInt32(22)
	msg.UInt64 = m.UInt64(23)

	// float:30-31
	msg.Float32 = m.Float32(30)
	msg.Float64 = m.Float64(31)

	// string/bytes:40-41
	msg.String = m.String(40)
	msg.Bytes = m.Bytes(41)

	// list:50
	{
		list := m.List(50)
		msg.List = make([]int64, 0, list.Len())

		for i := 0; i < list.Len(); i++ {
			val := list.Int64(i)
			msg.List = append(msg.List, val)
		}
	}

	// messages:51
	{
		list := m.List(51)
		msg.Messages = make([]*TestSubMessage, 0, list.Len())

		for i := 0; i < list.Len(); i++ {
			data := list.Element(i)
			if len(data) == 0 {
				continue
			}

			val := &TestSubMessage{}
			if err := val.Read(data); err != nil {
				return err
			}
			msg.Messages = append(msg.Messages, val)
		}
	}

	// strings:52
	{
		list := m.List(52)
		msg.Strings = make([]string, 0, list.Len())

		for i := 0; i < list.Len(); i++ {
			s := list.String(i)
			msg.Strings = append(msg.Strings, s)
		}
	}
	return nil
}

func (msg *TestSubMessage) Read(b []byte) error {
	m, err := ReadMessage(b)
	if err != nil {
		return err
	}

	// int:1-4
	msg.Int8 = m.Int8(1)
	msg.Int16 = m.Int16(2)
	msg.Int32 = m.Int32(3)
	msg.Int64 = m.Int64(4)
	return nil
}

// Write

func (msg TestMessage) Write(w *Writer) error {
	if err := w.BeginMessage(); err != nil {
		return err
	}

	// bool:1
	w.Bool(msg.Bool)
	w.Field(1)

	// int:10-13
	w.Int8(msg.Int8)
	w.Field(10)
	w.Int16(msg.Int16)
	w.Field(11)
	w.Int32(msg.Int32)
	w.Field(12)
	w.Int64(msg.Int64)
	w.Field(13)

	// uint:20-23
	w.UInt8(msg.UInt8)
	w.Field(20)
	w.UInt16(msg.UInt16)
	w.Field(21)
	w.UInt32(msg.UInt32)
	w.Field(22)
	w.UInt64(msg.UInt64)
	w.Field(23)

	// float:30-31
	w.Float32(msg.Float32)
	w.Field(30)
	w.Float64(msg.Float64)
	w.Field(31)

	// bytes:40-41
	w.String(msg.String)
	w.Field(40)
	w.Bytes(msg.Bytes)
	w.Field(41)

	// list:50
	if len(msg.List) > 0 {
		w.BeginList()
		for _, val := range msg.List {
			w.Int64(val)
			w.Element()
		}
		w.EndList()
		w.Field(50)
	}

	// messages:51
	if len(msg.Messages) > 0 {
		w.BeginList()
		for _, val := range msg.Messages {
			val.Write(w)
			w.Element()
		}
		w.EndList()
		w.Field(51)
	}

	// strings:52
	if len(msg.Strings) > 0 {
		w.BeginList()
		for _, val := range msg.Strings {
			w.String(val)
			w.Element()
		}
		w.EndList()
		w.Field(52)
	}

	return w.EndMessage()
}

func (msg TestSubMessage) Write(w *Writer) error {
	if err := w.BeginMessage(); err != nil {
		return err
	}

	// int:1-4
	w.Int8(msg.Int8)
	w.Field(1)
	w.Int16(msg.Int16)
	w.Field(2)
	w.Int32(msg.Int32)
	w.Field(3)
	w.Int64(msg.Int64)
	w.Field(4)

	return w.EndMessage()
}

// Value

type TestMessageData struct{ m Message }

func (d TestMessageData) Bool() bool       { return d.m.Bool(1) }
func (d TestMessageData) Int8() int8       { return d.m.Int8(10) }
func (d TestMessageData) Int16() int16     { return d.m.Int16(11) }
func (d TestMessageData) Int32() int32     { return d.m.Int32(12) }
func (d TestMessageData) Int64() int64     { return d.m.Int64(13) }
func (d TestMessageData) UInt8() uint8     { return d.m.UInt8(20) }
func (d TestMessageData) UInt16() uint16   { return d.m.UInt16(21) }
func (d TestMessageData) UInt32() uint32   { return d.m.UInt32(22) }
func (d TestMessageData) UInt64() uint64   { return d.m.UInt64(23) }
func (d TestMessageData) Float32() float32 { return d.m.Float32(30) }
func (d TestMessageData) Float64() float64 { return d.m.Float64(31) }
func (d TestMessageData) String() string   { return d.m.String(40) }
func (d TestMessageData) Bytes() []byte    { return d.m.Bytes(41) }
func (d TestMessageData) List() List       { return d.m.List(50) }
func (d TestMessageData) Messages() List   { return d.m.List(51) }
func (d TestMessageData) Strings() List    { return d.m.List(52) }

func getTestMessageData(b []byte) (TestMessageData, error) {
	msg, err := GetMessage(b)
	if err != nil {
		return TestMessageData{}, err
	}
	return TestMessageData{msg}, nil
}

func readTestMessageData(b []byte) (TestMessageData, error) {
	msg, err := ReadMessage(b)
	if err != nil {
		return TestMessageData{}, err
	}
	return TestMessageData{msg}, nil
}

type TestSubMessageData struct{ m Message }

func (d TestSubMessageData) Int8() int8   { return d.m.Int8(1) }
func (d TestSubMessageData) Int16() int16 { return d.m.Int16(2) }
func (d TestSubMessageData) Int32() int32 { return d.m.Int32(3) }
func (d TestSubMessageData) Int64() int64 { return d.m.Int64(4) }

func getTestSubMessageData(b []byte) (TestSubMessageData, error) {
	msg, err := GetMessage(b)
	if err != nil {
		return TestSubMessageData{}, err
	}
	return TestSubMessageData{msg}, nil
}

func readTestSubMessageData(b []byte) (TestSubMessageData, error) {
	msg, err := ReadMessage(b)
	if err != nil {
		return TestSubMessageData{}, err
	}
	return TestSubMessageData{msg}, nil
}