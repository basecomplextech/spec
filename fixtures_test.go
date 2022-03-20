package spec

import (
	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
)

// TestMessage

type TestMessage struct {
	m Message
}

func GetTestMessage(b []byte) TestMessage {
	m := GetMessage(b)
	return TestMessage{m}
}

func DecodeTestMessage(b []byte) (TestMessage, int, error) {
	m, n, err := DecodeMessage(b)
	if err != nil {
		return TestMessage{}, n, err
	}
	return TestMessage{m}, n, nil
}

func (m TestMessage) Bytes() []byte    { return m.m.Bytes() }
func (m TestMessage) Bool() bool       { return m.m.Bool(1) }
func (m TestMessage) Byte() byte       { return m.m.Byte(2) }
func (m TestMessage) Int32() int32     { return m.m.Int32(10) }
func (m TestMessage) Int64() int64     { return m.m.Int64(11) }
func (m TestMessage) Uint32() uint32   { return m.m.Uint32(20) }
func (m TestMessage) Uint64() uint64   { return m.m.Uint64(21) }
func (m TestMessage) U128() u128.U128  { return m.m.U128(22) }
func (m TestMessage) U256() u256.U256  { return m.m.U256(23) }
func (m TestMessage) Float32() float32 { return m.m.Float32(30) }
func (m TestMessage) Float64() float64 { return m.m.Float64(31) }
func (m TestMessage) String() string   { return m.m.String(40) }
func (m TestMessage) Bytes_() []byte   { return m.m.ByteSlice(41) }

func (m TestMessage) Submessage() TestSubmessage {
	b := m.m.Field(50)
	return GetTestSubmessage(b)
}

func (m TestMessage) List() List[int64] {
	b := m.m.Field(51)
	return GetList(b, DecodeInt64)
}

func (m TestMessage) Messages() List[TestElement] {
	b := m.m.Field(52)
	return GetList(b, DecodeTestElement)
}

func (m TestMessage) Strings() List[string] {
	b := m.m.Field(53)
	return GetList(b, DecodeString)
}

func (m TestMessage) Struct() TestStruct {
	data := m.m.Field(60)
	v, _ := DecodeTestStruct(data)
	return v
}

// TestMessageEncoder

type TestMessageEncoder struct {
	e *Encoder
}

func BeginTestMessage(e *Encoder) TestMessageEncoder {
	e.BeginMessage()
	return TestMessageEncoder{e}
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

func (e TestMessageEncoder) Submessage() TestSubmessageEncoder {
	e.e.BeginField(50)
	return BeginSubmessage(e.e)
}

func (e TestMessageEncoder) List() ListValueEncoder[int64] {
	e.e.BeginField(51)
	return BeginValueList(e.e, e.e.Int64)
}

func (e TestMessageEncoder) Messages() ListEncoder[TestElementEncoder] {
	e.e.BeginField(52)

	return BeginList(e.e, BeginTestSubmessage)
}

func (e TestMessageEncoder) Strings() ListValueEncoder[string] {
	e.e.BeginField(53)
	return BeginValueList(e.e, e.e.String)
}

func (e TestMessageEncoder) Struct(v TestStruct) error {
	// return e.e.Field(60)
	return nil
}

// TestSubmessage

type TestSubmessage struct {
	m Message
}

func GetTestSubmessage(b []byte) TestSubmessage {
	m := GetMessage(b)
	return TestSubmessage{m}
}

func DecodeTestSubmessage(b []byte) (TestSubmessage, int, error) {
	m, n, err := DecodeMessage(b)
	if err != nil {
		return TestSubmessage{}, n, err
	}
	return TestSubmessage{m}, n, nil
}

func (m TestSubmessage) Bytes() []byte { return m.m.Bytes() }
func (m TestSubmessage) Int32() int32  { return m.m.Int32(1) }
func (m TestSubmessage) Int64() int64  { return m.m.Int64(2) }

// TestSubmessageEncoder

type TestSubmessageEncoder struct {
	e *Encoder
}

func BeginSubmessage(e *Encoder) TestSubmessageEncoder {
	e.BeginMessage()

	return TestSubmessageEncoder{e}
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
	m Message
}

func GetTestElement(b []byte) TestElement {
	m := GetMessage(b)
	return TestElement{m}
}

func DecodeTestElement(b []byte) (TestElement, int, error) {
	m, n, err := DecodeMessage(b)
	if err != nil {
		return TestElement{}, n, err
	}
	return TestElement{m}, n, nil
}

func (m TestElement) Byte() byte {
	return m.m.Byte(1)
}

func (m TestElement) Int32() int32 {
	return m.m.Int32(2)
}

func (m TestElement) Int64() int64 {
	return m.m.Int64(3)
}

// TestElementEncoder

type TestElementEncoder struct {
	e *Encoder
}

func BeginTestSubmessage(e *Encoder) TestElementEncoder {
	e.BeginMessage()
	return TestElementEncoder{e}
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

func DecodeTestStruct(b []byte) (TestStruct, error) {
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

	// s.Y, r, err = r.DecodeInt64()
	// if err != nil {
	// 	return err
	// }
	// s.X, r, err = r.DecodeInt64()
	// if err != nil {
	// 	return err
	// }
	// return nil
}

func (s TestStruct) Write(e *Encoder) error {
	if err := e.BeginStruct(); err != nil {
		return err
	}

	e.Int64(s.X)
	e.StructField()

	e.Int64(s.Y)
	e.StructField()

	return e.EndStruct()
}
