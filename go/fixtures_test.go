package spec

import (
	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/baselibrary/types"
)

// TestMessage

type TestMessage struct {
	msg Message
}

func GetTestMessage(b []byte) TestMessage {
	msg := GetMessage(b)
	return TestMessage{msg}
}

func DecodeTestMessage(b []byte) (_ TestMessage, size int, err error) {
	msg, size, err := DecodeMessage(b)
	if err != nil {
		return
	}
	return TestMessage{msg}, size, nil
}

func BuildTestMessage() (_ TestMessageBuilder, err error) {
	e := NewEncoder()
	if err = e.BeginMessage(); err != nil {
		return
	}
	return TestMessageBuilder{e}, nil
}

func BuildTestMessageBuffer(b buffer.Buffer) (_ TestMessageBuilder, err error) {
	e := NewEncoderBuffer(b)
	if err = e.BeginMessage(); err != nil {
		return
	}
	return TestMessageBuilder{e}, nil
}

func BuildTestMessageEncoder(e *Encoder) (result TestMessageBuilder, err error) {
	if err = e.BeginMessage(); err != nil {
		return
	}
	result = TestMessageBuilder{e}
	return
}

func (m TestMessage) Bool() bool           { return m.msg.GetBool(1) }
func (m TestMessage) Byte() byte           { return m.msg.GetByte(2) }
func (m TestMessage) Int32() int32         { return m.msg.GetInt32(10) }
func (m TestMessage) Int64() int64         { return m.msg.GetInt64(11) }
func (m TestMessage) Uint32() uint32       { return m.msg.GetUint32(20) }
func (m TestMessage) Uint64() uint64       { return m.msg.GetUint64(21) }
func (m TestMessage) Bin128() types.Bin128 { return m.msg.GetBin128(22) }
func (m TestMessage) Bin256() types.Bin256 { return m.msg.GetBin256(23) }
func (m TestMessage) Float32() float32     { return m.msg.GetFloat32(30) }
func (m TestMessage) Float64() float64     { return m.msg.GetFloat64(31) }
func (m TestMessage) String() string       { return m.msg.GetString(40) }
func (m TestMessage) Bytes() []byte        { return m.msg.GetBytes(41) }
func (m TestMessage) Unwrap() Message      { return m.msg }

func (m TestMessage) Submessage() TestSubmessage {
	b := m.msg.Field(50)
	return GetTestSubmessage(b)
}

func (m TestMessage) List() List[int64] {
	b := m.msg.Field(51)
	return NewList(b, DecodeInt64)
}

func (m TestMessage) Messages() List[TestElement] {
	b := m.msg.Field(52)
	return NewList(b, DecodeTestElement)
}

func (m TestMessage) Strings() List[string] {
	b := m.msg.Field(53)
	return NewList(b, DecodeString)
}

func (m TestMessage) Struct() TestStruct {
	b := m.msg.Field(60)
	return GetTestStruct(b)
}

// TestMessageBuilder

type TestMessageBuilder struct {
	e *Encoder
}

func (b TestMessageBuilder) End() ([]byte, error) {
	return b.e.End()
}

func (b TestMessageBuilder) Bool(v bool) error {
	b.e.Bool(v)
	return b.e.Field(1)
}

func (b TestMessageBuilder) Byte(v byte) error {
	b.e.Byte(v)
	return b.e.Field(2)
}

func (b TestMessageBuilder) Int32(v int32) error {
	b.e.Int32(v)
	return b.e.Field(10)
}

func (b TestMessageBuilder) Int64(v int64) error {
	b.e.Int64(v)
	return b.e.Field(11)
}

func (b TestMessageBuilder) Uint32(v uint32) error {
	b.e.Uint32(v)
	return b.e.Field(20)
}

func (b TestMessageBuilder) Uint64(v uint64) error {
	b.e.Uint64(v)
	return b.e.Field(21)
}

func (b TestMessageBuilder) Bin128(v types.Bin128) error {
	b.e.Bin128(v)
	return b.e.Field(22)
}

func (b TestMessageBuilder) Bin256(v types.Bin256) error {
	b.e.Bin256(v)
	return b.e.Field(23)
}

func (b TestMessageBuilder) Float32(v float32) error {
	b.e.Float32(v)
	return b.e.Field(30)
}

func (b TestMessageBuilder) Float64(v float64) error {
	b.e.Float64(v)
	return b.e.Field(31)
}

func (b TestMessageBuilder) String(v string) error {
	b.e.String(v)
	return b.e.Field(40)
}

func (b TestMessageBuilder) Bytes(v []byte) error {
	b.e.Bytes(v)
	return b.e.Field(41)
}

func (b TestMessageBuilder) Submessage() (TestSubmessageBuilder, error) {
	b.e.BeginField(50)
	return BuildTestSubmessageEncoder(b.e)
}

func (b TestMessageBuilder) List() ValueListBuilder[int64] {
	b.e.BeginField(51)
	return NewValueListBuilder(b.e, EncodeInt64)
}

func (b TestMessageBuilder) Messages() ListBuilder[TestElementBuilder] {
	b.e.BeginField(52)
	return NewListBuilder(b.e, BuildTestElementEncoder)
}

func (b TestMessageBuilder) Strings() ValueListBuilder[string] {
	b.e.BeginField(53)
	return NewValueListBuilder(b.e, EncodeString)
}

func (b TestMessageBuilder) Struct(v TestStruct) error {
	EncodeValue(b.e, v, EncodeTestStruct)
	return b.e.Field(60)
}

// TestSubmessage

type TestSubmessage struct {
	msg Message
}

func GetTestSubmessage(b []byte) TestSubmessage {
	msg := GetMessage(b)
	return TestSubmessage{msg}
}

func DecodeTestSubmessage(b []byte) (_ TestSubmessage, size int, err error) {
	msg, size, err := DecodeMessage(b)
	if err != nil {
		return
	}
	return TestSubmessage{msg}, size, nil
}

func BuildTestSubmessage() (_ TestSubmessageBuilder, err error) {
	e := NewEncoder()
	if err = e.BeginMessage(); err != nil {
		return
	}
	return TestSubmessageBuilder{e}, nil
}

func BuildTestSubmessageBuffer(b buffer.Buffer) (_ TestSubmessageBuilder, err error) {
	e := NewEncoderBuffer(b)
	if err = e.BeginMessage(); err != nil {
		return
	}
	return TestSubmessageBuilder{e}, nil
}

func BuildTestSubmessageEncoder(e *Encoder) (_ TestSubmessageBuilder, err error) {
	if err = e.BeginMessage(); err != nil {
		return
	}
	return TestSubmessageBuilder{e}, nil
}

func (m TestSubmessage) Int32() int32    { return m.msg.GetInt32(1) }
func (m TestSubmessage) Int64() int64    { return m.msg.GetInt64(2) }
func (m TestSubmessage) Unwrap() Message { return m.msg }

// TestSubmessageBuilder

type TestSubmessageBuilder struct {
	e *Encoder
}

func (b TestSubmessageBuilder) End() ([]byte, error) {
	return b.e.End()
}

func (b TestSubmessageBuilder) Int32(v int32) error {
	b.e.Int32(v)
	return b.e.Field(1)
}

func (b TestSubmessageBuilder) Int64(v int64) error {
	b.e.Int64(v)
	return b.e.Field(2)
}

// TestElement

type TestElement struct {
	msg Message
}

func GetTestElement(b []byte) TestElement {
	msg := GetMessage(b)
	return TestElement{msg}
}

func DecodeTestElement(b []byte) (_ TestElement, size int, err error) {
	msg, size, err := DecodeMessage(b)
	if err != nil {
		return
	}
	return TestElement{msg}, size, nil
}

func BuildTestElement() (_ TestElementBuilder, err error) {
	e := NewEncoder()
	if err = e.BeginMessage(); err != nil {
		return
	}
	return TestElementBuilder{e}, nil
}

func BuildTestElementBuffer(b buffer.Buffer) (_ TestElementBuilder, err error) {
	e := NewEncoderBuffer(b)
	if err = e.BeginMessage(); err != nil {
		return
	}
	return TestElementBuilder{e}, nil
}

func BuildTestElementEncoder(e *Encoder) (_ TestElementBuilder) {
	if err := e.BeginMessage(); err != nil {
		return
	}
	return TestElementBuilder{e}
}

func (m TestElement) Byte() byte {
	return m.msg.GetByte(1)
}

func (m TestElement) Int32() int32 {
	return m.msg.GetInt32(2)
}

func (m TestElement) Int64() int64 {
	return m.msg.GetInt64(3)
}

// TestElementBuilder

type TestElementBuilder struct {
	e *Encoder
}

func (b TestElementBuilder) End() ([]byte, error) {
	return b.e.End()
}

func (b TestElementBuilder) Byte(v byte) error {
	b.e.Byte(v)
	return b.e.Field(1)
}

func (b TestElementBuilder) Int32(v int32) error {
	b.e.Int32(v)
	return b.e.Field(2)
}

func (b TestElementBuilder) Int64(v int64) error {
	b.e.Int64(v)
	return b.e.Field(3)
}

// TestStruct

type TestStruct struct {
	X int64
	Y int64
}

func GetTestStruct(b []byte) (result TestStruct) {
	result, _, _ = DecodeTestStruct(b)
	return result
}

func DecodeTestStruct(b []byte) (result TestStruct, total int, err error) {
	dataSize, size, err := DecodeStruct(b)
	if err != nil {
		return
	}

	b = b[len(b)-size:]
	n := size - dataSize
	off := len(b)

	// decode in reverse order

	off -= n
	result.Y, n, err = DecodeInt64(b[:off])
	if err != nil {
		return
	}

	off -= n
	result.X, n, err = DecodeInt64(b[:off])
	if err != nil {
		return
	}

	return
}

func EncodeTestStruct(b buffer.Buffer, s TestStruct) (int, error) {
	var dataSize, n int
	var err error

	n, err = EncodeInt64(b, s.X)
	if err != nil {
		return 0, err
	}
	dataSize += n

	n, err = EncodeInt64(b, s.Y)
	if err != nil {
		return 0, err
	}
	dataSize += n

	n, err = EncodeStruct(b, dataSize)
	if err != nil {
		return 0, err
	}
	return dataSize + n, nil
}
