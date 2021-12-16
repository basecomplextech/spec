package protocol

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

type TestSubMessage struct {
	Int8  int8  `tag:"1"`
	Int16 int16 `tag:"2"`
	Int32 int32 `tag:"3"`
	Int64 int64 `tag:"4"`
}

func (msg *TestMessage) Read(v Value) error {
	var m = v.Message()
	var f Value
	var ok bool

	// int:10-13
	f, ok = m.Field(10)
	if ok {
		msg.Int8 = f.Int8()
	}
	f, ok = m.Field(11)
	if ok {
		msg.Int16 = f.Int16()
	}
	f, ok = m.Field(12)
	if ok {
		msg.Int32 = f.Int32()
	}
	f, ok = m.Field(13)
	if ok {
		msg.Int64 = f.Int64()
	}

	// uint:20-23
	f, ok = m.Field(20)
	if ok {
		msg.UInt8 = f.UInt8()
	}
	f, ok = m.Field(21)
	if ok {
		msg.UInt16 = f.UInt16()
	}
	f, ok = m.Field(22)
	if ok {
		msg.UInt32 = f.UInt32()
	}
	f, ok = m.Field(23)
	if ok {
		msg.UInt64 = f.UInt64()
	}

	// float:30-31
	f, ok = m.Field(30)
	if ok {
		msg.Float32 = f.Float32()
	}
	f, ok = m.Field(31)
	if ok {
		msg.Float64 = f.Float64()
	}

	// list:40
	f, ok = m.Field(40)
	if ok {
		list := f.List()
		msg.List = make([]int64, 0, list.Len())

		for i := 0; i < list.Len(); i++ {
			el, ok := list.Element(i)
			if !ok {
				continue
			}

			val := el.Int64()
			msg.List = append(msg.List, val)
		}
	}

	// list:41
	f, ok = m.Field(41)
	if ok {
		list := f.List()
		msg.Messages = make([]*TestSubMessage, 0, list.Len())

		for i := 0; i < list.Len(); i++ {
			el, ok := list.Element(i)
			if !ok {
				continue
			}

			msg1 := &TestSubMessage{}
			if err := msg1.Read(el); err != nil {
				return err
			}
			msg.Messages = append(msg.Messages, msg1)
		}
	}

	return nil
}

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
	if err := w.BeginList(); err != nil {
		return err
	}
	for _, val := range msg.List {
		if err := w.Int64(val); err != nil {
			return err
		}
		if err := w.Element(); err != nil {
			return err
		}
	}
	if err := w.EndList(); err != nil {
		return err
	}
	if err := w.Field(40); err != nil {
		return err
	}

	// list:41
	if err := w.BeginList(); err != nil {
		return err
	}
	for _, msg1 := range msg.Messages {
		if err := msg1.Write(w); err != nil {
			return err
		}
		if err := w.Element(); err != nil {
			return err
		}
	}
	if err := w.EndList(); err != nil {
		return err
	}
	if err := w.Field(41); err != nil {
		return err
	}

	return w.EndMessage()
}

func (msg *TestSubMessage) Read(v Value) error {
	var m = v.Message()
	var f Value
	var ok bool

	// int:1-4
	f, ok = m.Field(1)
	if ok {
		msg.Int8 = f.Int8()
	}
	f, ok = m.Field(2)
	if ok {
		msg.Int16 = f.Int16()
	}
	f, ok = m.Field(3)
	if ok {
		msg.Int32 = f.Int32()
	}
	f, ok = m.Field(4)
	if ok {
		msg.Int64 = f.Int64()
	}
	return nil
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

// Tests

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

func newTestSubMessage(i int) *TestSubMessage {
	return &TestSubMessage{
		Int8:  int8(i + 0),
		Int16: int16(i + 10),
		Int32: int32(i + 100),
		Int64: int64(i + 1000),
	}
}

func TestMessage_Write(t *testing.T) {
	msg := newTestMessage()

	w := NewWriter()
	if err := msg.Write(w); err != nil {
		t.Fatal(err)
	}
	b, err := w.End()
	if err != nil {
		t.Fatal(err)
	}

	msg1 := &TestMessage{}
	if err := msg1.Read(ReadValue(b)); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg, msg1)
}
