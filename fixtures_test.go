package spec

import (
	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

// TestMessage

type TestMessage struct {
	msg Message
}

func GetTestMessage(b []byte) TestMessage {
	msg := GetMessage(b)
	return TestMessage{msg}
}

func DecodeTestMessage(b []byte) (result TestMessage, n int, err error) {
	msg, n, err := DecodeMessage(b)
	if err != nil {
		return
	}
	result = TestMessage{msg}
	return
}

func EncodeTestMessage(e *Encoder) (result TestMessageEncoder, err error) {
	if err = e.BeginMessage(); err != nil {
		return
	}
	result = TestMessageEncoder{e}
	return
}

func (m TestMessage) RawBytes() []byte { return m.msg.Raw() }
func (m TestMessage) Bool() bool       { return m.msg.Bool(1) }
func (m TestMessage) Byte() byte       { return m.msg.Byte(2) }
func (m TestMessage) Int32() int32     { return m.msg.Int32(10) }
func (m TestMessage) Int64() int64     { return m.msg.Int64(11) }
func (m TestMessage) Uint32() uint32   { return m.msg.Uint32(20) }
func (m TestMessage) Uint64() uint64   { return m.msg.Uint64(21) }
func (m TestMessage) U128() u128.U128  { return m.msg.U128(22) }
func (m TestMessage) U256() u256.U256  { return m.msg.U256(23) }
func (m TestMessage) Float32() float32 { return m.msg.Float32(30) }
func (m TestMessage) Float64() float64 { return m.msg.Float64(31) }
func (m TestMessage) String() string   { return m.msg.String(40) }
func (m TestMessage) Bytes() []byte    { return m.msg.Bytes(41) }

func (m TestMessage) Submessage() TestSubmessage {
	b := m.msg.Field(50)
	return GetTestSubmessage(b)
}

func (m TestMessage) List() List[int64] {
	b := m.msg.Field(51)
	return GetList(b, DecodeInt64)
}

func (m TestMessage) Messages() List[TestElement] {
	b := m.msg.Field(52)
	return GetList(b, DecodeTestElement)
}

func (m TestMessage) Strings() List[string] {
	b := m.msg.Field(53)
	return GetList(b, DecodeString)
}

func (m TestMessage) Struct() TestStruct {
	b := m.msg.Field(60)
	return GetTestStruct(b)
}

// TestMessageEncoder

type TestMessageEncoder struct {
	e *Encoder
}

func (e TestMessageEncoder) End() ([]byte, error) {
	return e.e.End()
}

func (e TestMessageEncoder) Bool(v bool) error {
	e.e.Bool(v)
	return e.e.Field(1)
}

func (e TestMessageEncoder) Byte(v byte) error {
	e.e.Byte(v)
	return e.e.Field(2)
}

func (e TestMessageEncoder) Int32(v int32) error {
	e.e.Int32(v)
	return e.e.Field(10)
}

func (e TestMessageEncoder) Int64(v int64) error {
	e.e.Int64(v)
	return e.e.Field(11)
}

func (e TestMessageEncoder) Uint32(v uint32) error {
	e.e.Uint32(v)
	return e.e.Field(20)
}

func (e TestMessageEncoder) Uint64(v uint64) error {
	e.e.Uint64(v)
	return e.e.Field(21)
}

func (e TestMessageEncoder) U128(v u128.U128) error {
	e.e.U128(v)
	return e.e.Field(22)
}

func (e TestMessageEncoder) U256(v u256.U256) error {
	e.e.U256(v)
	return e.e.Field(23)
}

func (e TestMessageEncoder) Float32(v float32) error {
	e.e.Float32(v)
	return e.e.Field(30)
}

func (e TestMessageEncoder) Float64(v float64) error {
	e.e.Float64(v)
	return e.e.Field(31)
}

func (e TestMessageEncoder) String(v string) error {
	e.e.String(v)
	return e.e.Field(40)
}

func (e TestMessageEncoder) Bytes(v []byte) error {
	e.e.Bytes(v)
	return e.e.Field(41)
}

func (e TestMessageEncoder) Submessage() (result TestSubmessageEncoder, err error) {
	e.e.BeginField(50)
	return EncodeTestSubmessage(e.e)
}

func (e TestMessageEncoder) List() (result ValuesEncoder[int64], err error) {
	e.e.BeginField(51)
	return EncodeValues(e.e, e.e.Int64)
}

func (e TestMessageEncoder) Messages() (result ListEncoder[TestElementEncoder], err error) {
	e.e.BeginField(52)
	return EncodeList(e.e, EncodeTestElement)
}

func (e TestMessageEncoder) Strings() (result ValuesEncoder[string], err error) {
	e.e.BeginField(53)
	return EncodeValues(e.e, e.e.String)
}

func (e TestMessageEncoder) Struct(v TestStruct) ([]byte, error) {
	e.e.BeginField(60)
	return EncodeTestStruct(e.e, v)
}

// TestSubmessage

type TestSubmessage struct {
	msg Message
}

func GetTestSubmessage(b []byte) TestSubmessage {
	msg := GetMessage(b)
	return TestSubmessage{msg}
}

func DecodeTestSubmessage(b []byte) (result TestSubmessage, n int, err error) {
	msg, n, err := DecodeMessage(b)
	if err != nil {
		return
	}
	result = TestSubmessage{msg}
	return
}

func EncodeTestSubmessage(e *Encoder) (result TestSubmessageEncoder, err error) {
	if err = e.BeginMessage(); err != nil {
		return
	}
	result = TestSubmessageEncoder{e}
	return
}

func (m TestSubmessage) RawBytes() []byte { return m.msg.Raw() }
func (m TestSubmessage) Int32() int32     { return m.msg.Int32(1) }
func (m TestSubmessage) Int64() int64     { return m.msg.Int64(2) }

// TestSubmessageEncoder

type TestSubmessageEncoder struct {
	e *Encoder
}

func (e TestSubmessageEncoder) End() ([]byte, error) {
	return e.e.End()
}

func (e TestSubmessageEncoder) Int32(v int32) error {
	e.e.Int32(v)
	return e.e.Field(1)
}

func (e TestSubmessageEncoder) Int64(v int64) error {
	e.e.Int64(v)
	return e.e.Field(2)
}

// TestElement

type TestElement struct {
	msg Message
}

func GetTestElement(b []byte) TestElement {
	msg := GetMessage(b)
	return TestElement{msg}
}

func DecodeTestElement(b []byte) (result TestElement, n int, err error) {
	msg, n, err := DecodeMessage(b)
	if err != nil {
		return
	}
	result = TestElement{msg}
	return
}

func EncodeTestElement(e *Encoder) (result TestElementEncoder, err error) {
	if err = e.BeginMessage(); err != nil {
		return
	}
	result = TestElementEncoder{e}
	return
}

func (m TestElement) Byte() byte {
	return m.msg.Byte(1)
}

func (m TestElement) Int32() int32 {
	return m.msg.Int32(2)
}

func (m TestElement) Int64() int64 {
	return m.msg.Int64(3)
}

// TestElementEncoder

type TestElementEncoder struct {
	e *Encoder
}

func (e TestElementEncoder) End() ([]byte, error) {
	return e.e.End()
}

func (e TestElementEncoder) Byte(v byte) error {
	e.e.Byte(v)
	return e.e.Field(1)
}

func (e TestElementEncoder) Int32(v int32) error {
	e.e.Int32(v)
	return e.e.Field(2)
}

func (e TestElementEncoder) Int64(v int64) error {
	e.e.Int64(v)
	return e.e.Field(3)
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

func EncodeTestStruct(e *Encoder, s TestStruct) ([]byte, error) {
	if err := e.BeginStruct(); err != nil {
		return nil, err
	}

	e.Int64(s.X)
	e.StructField()

	e.Int64(s.Y)
	e.StructField()

	return e.End()
}
