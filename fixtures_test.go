package spec

import (
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

func (m TestMessage) Bool() bool {
	return m.m.Bool(1)
}

func (m TestMessage) Byte() byte {
	return m.m.Byte(2)
}

func (m TestMessage) Int32() int32 {
	return m.m.Int32(10)
}

func (m TestMessage) Int64() int64 {
	return m.m.Int64(11)
}

func (m TestMessage) Uint32() uint32 {
	return m.m.Uint32(20)
}

func (m TestMessage) Uint64() uint64 {
	return m.m.Uint64(21)
}

func (m TestMessage) U128() u128.U128 {
	return m.m.U128(22)
}

func (m TestMessage) U256() u256.U256 {
	return m.m.U256(23)
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
	return GetList(b, DecodeInt64)
}

func (m TestMessage) Messages() List[TestSubmessage] {
	b := m.m.Field(51)
	return GetList(b, ReadTestSubmessage)
}

func (m TestMessage) Strings() List[string] {
	b := m.m.Field(52)
	return GetList(b, DecodeString)
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

func BeginTestMessage(w *Writer) TestMessageWriter {
	w.BeginMessage()
	return TestMessageWriter{w}
}

func (w TestMessageWriter) End() error {
	return w.w.EndMessage()
}

func (w TestMessageWriter) Bool(v bool) error {
	w.w.Bool(v)
	return w.w.Field(1)
}

func (w TestMessageWriter) Byte(v byte) error {
	w.w.Byte(v)
	return w.w.Field(2)
}

func (w TestMessageWriter) Int32(v int32) error {
	w.w.Int32(v)
	return w.w.Field(10)
}

func (w TestMessageWriter) Int64(v int64) error {
	w.w.Int64(v)
	return w.w.Field(11)
}

func (w TestMessageWriter) Uint32(v uint32) error {
	w.w.Uint32(v)
	return w.w.Field(20)
}

func (w TestMessageWriter) Uint64(v uint64) error {
	w.w.Uint64(v)
	return w.w.Field(21)
}

func (w TestMessageWriter) U128(v u128.U128) error {
	w.w.U128(v)
	return w.w.Field(22)
}

func (w TestMessageWriter) U256(v u256.U256) error {
	w.w.U256(v)
	return w.w.Field(23)
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

func (w TestMessageWriter) BeginList() ListValueEncoder[int64] {
	return BeginValueList(w.w, w.w.Int64)
}

func (w TestMessageWriter) EndList() error {
	w.w.EndList()
	return w.w.Field(50)
}

func (w TestMessageWriter) BeginMessages() ListEncoder[TestSubmessageWriter] {
	return BeginList(w.w, BeginTestSubmessage)
}

func (w TestMessageWriter) EndMessages() error {
	w.w.EndList()
	return w.w.Field(51)
}

func (w TestMessageWriter) BeginStrings() ListValueEncoder[string] {
	return BeginValueList(w.w, w.w.String)
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

func (m TestSubmessage) Byte() byte {
	return m.m.Byte(1)
}

func (m TestSubmessage) Int32() int32 {
	return m.m.Int32(2)
}

func (m TestSubmessage) Int64() int64 {
	return m.m.Int64(3)
}

// TestSubmessageWriter

type TestSubmessageWriter struct {
	w *Writer
}

func BeginTestSubmessage(w *Writer) TestSubmessageWriter {
	w.BeginMessage()
	return TestSubmessageWriter{w}
}

func (w TestSubmessageWriter) End() error {
	return w.w.EndMessage()
}

func (w TestSubmessageWriter) Byte(v byte) error {
	w.w.Byte(v)
	return w.w.Field(1)
}

func (w TestSubmessageWriter) Int32(v int32) error {
	w.w.Int32(v)
	return w.w.Field(2)
}

func (w TestSubmessageWriter) Int64(v int64) error {
	w.w.Int64(v)
	return w.w.Field(3)
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
