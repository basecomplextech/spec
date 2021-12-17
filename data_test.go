package spec

import (
	"math"
)

// Test messages

type TestMessage struct {
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

	List     []int64           `tag:"40"`
	Messages []*TestSubMessage `tag:"41"`
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

	return &TestMessage{
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

		List:     list,
		Messages: messages,
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
	m := ReadMessage(b)

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

	// list:40
	{
		list := m.List(40)
		msg.List = make([]int64, 0, list.Len())

		for i := 0; i < list.Len(); i++ {
			val := list.Element(i).Int64()
			msg.List = append(msg.List, val)
		}
	}

	// list:41
	{
		list := m.List(41)
		msg.Messages = make([]*TestSubMessage, 0, list.Len())

		for i := 0; i < list.Len(); i++ {
			el := list.Element(i)
			val := &TestSubMessage{}
			if err := val.Read(el); err != nil {
				return err
			}
			msg.Messages = append(msg.Messages, val)
		}
	}
	return nil
}

func (msg *TestSubMessage) Read(b []byte) error {
	m := ReadMessage(b)

	// int:1-4
	msg.Int8 = m.Int8(1)
	msg.Int16 = m.Int16(2)
	msg.Int32 = m.Int32(3)
	msg.Int64 = m.Int64(4)
	return nil
}

// Write

func (msg TestMessage) Write(w *Writer) error {
	w.BeginMessage()

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

	// list:40
	w.BeginList()
	for _, val := range msg.List {
		w.Int64(val)
		w.Element()
	}
	w.EndList()
	w.Field(40)

	// list:41
	w.BeginList()
	for _, val := range msg.Messages {
		val.Write(w)
		w.Element()
	}
	w.EndList()
	w.Field(41)

	return w.EndMessage()
}

func (msg TestSubMessage) Write(w *Writer) error {
	w.BeginMessage()

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

// Data

type TestMessageData struct{ m Message }

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
func (d TestMessageData) List() List       { return d.m.List(40) }
func (d TestMessageData) Messages() List   { return d.m.List(41) }

func readTestMessageData(b []byte) TestMessageData {
	d := ReadMessage(b)
	return TestMessageData{d}
}

type TestSubMessageData struct{ m Message }

func (d TestSubMessageData) Int8() int8   { return d.m.Int8(1) }
func (d TestSubMessageData) Int16() int16 { return d.m.Int16(2) }
func (d TestSubMessageData) Int32() int32 { return d.m.Int32(3) }
func (d TestSubMessageData) Int64() int64 { return d.m.Int64(4) }

func readTestSubMessageData(b []byte) TestSubMessageData {
	d := ReadMessage(b)
	return TestSubMessageData{d}
}
