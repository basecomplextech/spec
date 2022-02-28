package pkg1

import (
	"testing"

	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
	"github.com/complexl/spec/testgen/golang/pkg2"
	"github.com/complexl/spec/testgen/golang/sub/pkg3"
	"github.com/stretchr/testify/assert"
)

func testMessage() *Message {
	return &Message{
		FieldBool: true,
		FieldEnum: EnumOne,

		FieldInt8:  1,
		FieldInt16: 2,
		FieldInt32: 3,
		FieldInt64: 4,

		FieldUint8:  1,
		FieldUint16: 2,
		FieldUint32: 3,
		FieldUint64: 4,

		FieldU128: u128.Random(),
		FieldU256: u256.Random(),

		FieldFloat32: 10.0,
		FieldFloat64: 20.0,

		FieldString: "hello, world",
		FieldBytes:  []byte("abc"),

		Msg: &Node{
			Value: "a",
			Next: &Node{
				Value: "b",
			},
		},
		Value: Struct{
			Key:   123,
			Value: 456,
		},
		Imported: &pkg2.SubMessage{
			Key:   "key",
			Value: pkg3.Value{},
		},

		ListInts:     []int64{1, 2, 3},
		ListStrings:  []string{"a", "b", "c"},
		ListValues:   []Struct{},
		ListMessages: []*Node{{Value: "1"}, {Value: "2"}},
		ListImported: []*pkg2.SubMessage{{Key: "A"}, {Key: "B"}},
	}
}

func testMessageData(t *testing.T) []byte {
	m := testMessage()

	data, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestMessage_Marshal_Unmarshal(t *testing.T) {
	m := testMessage()

	data, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	m1 := &Message{}
	if err := m1.Unmarshal(data); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, m, m1)
}

func TestReadMessageData(t *testing.T) {
	d := testMessageData(t)

	data, err := ReadMessageData(d)
	if err != nil {
		t.Fatal(err)
	}

	ok := data.FieldBool()
	assert.True(t, ok)
}
