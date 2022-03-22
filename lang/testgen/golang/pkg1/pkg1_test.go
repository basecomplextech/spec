package pkg1

import (
	"testing"

	"github.com/complexl/library/tests"
	"github.com/complexl/library/u128"
	"github.com/complexl/library/u256"
	"github.com/complexl/spec"
	"github.com/complexl/spec/lang/testgen/golang/sub/pkg3"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Decode(t *testing.T) {
	m := testMessage(t)
	b := m.RawBytes()

	m1, n, err := DecodeMessage(b)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(b), n)
	assert.Equal(t, m, m1)
}

// fixtures

func testMessage(t tests.T) Message {
	e := spec.NewEncoder()
	return testEncode(t, e)
}

func testEncode(t tests.T, e *spec.Encoder) Message {
	msg, err := EncodeMessage(e)
	if err != nil {
		t.Fatal(err)
	}

	msg.FieldBool(true)
	msg.FieldEnum(EnumOne)

	msg.FieldInt32(1)
	msg.FieldInt64(2)
	msg.FieldUint32(3)
	msg.FieldUint64(4)

	msg.FieldU128(u128.FromInt64(1))
	msg.FieldU256(u256.FromInt64(2))

	msg.FieldFloat32(10.0)
	msg.FieldFloat64(20.0)

	msg.FieldString("hello, world")
	msg.FieldBytes([]byte("abc"))

	msg.FieldStruct(Struct{
		Key:   123,
		Value: 456,
	})

	{
		node, err := msg.Node()
		if err != nil {
			t.Fatal(err)
		}
		node.Value("a")

		{
			next, err := node.Next()
			if err != nil {
				t.Fatal(err)
			}
			next.Value("b")
			next.End()
		}

		node.End()
	}

	msg.Value(Struct{
		Key:   123,
		Value: 456,
	})

	{
		submsg, err := msg.Imported()
		if err != nil {
			t.Fatal(err)
		}
		submsg.Key("key")
		submsg.Value(pkg3.Value{})
		submsg.End()
	}

	{
		list, err := msg.ListInts()
		if err != nil {
			t.Fatal(err)
		}
		for _, x := range []int64{1, 2, 3} {
			list.Next(x)
		}
		if _, err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	{
		list, err := msg.ListStrings()
		if err != nil {
			t.Fatal(err)
		}
		for _, x := range []string{"a", "b", "c"} {
			list.Next(x)
		}
		if _, err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	{
		list, err := msg.ListMessages()
		if err != nil {
			t.Fatal(err)
		}
		for _, x := range []string{"1", "2"} {
			elem, err := list.Next()
			if err != nil {
				t.Fatal(err)
			}
			elem.Value(x)
			elem.End()
		}
		if _, err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	{
		list, err := msg.ListImported()
		if err != nil {
			t.Fatal(err)
		}
		for _, x := range []string{"a", "b"} {
			elem, err := list.Next()
			if err != nil {
				t.Fatal(err)
			}
			elem.Key(x)
			elem.End()
		}
		if _, err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	b, err := msg.End()
	if err != nil {
		t.Fatal(err)
	}

	return GetMessage(b)
}
