package pkg1

import (
	"testing"

	"github.com/epochtimeout/baselibrary/tests"
	"github.com/epochtimeout/baselibrary/types"
	spec "github.com/epochtimeout/spec/go"
	"github.com/epochtimeout/spec/lang/testgen/golang/sub/pkg3"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Decode(t *testing.T) {
	m := testMessage(t)
	b := m.Unwrap().Bytes()

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
	msg := BuildMessageEncoder(e)

	msg.FieldBool(true)
	msg.FieldEnum(EnumOne)

	msg.FieldInt32(1)
	msg.FieldInt64(2)
	msg.FieldUint32(3)
	msg.FieldUint64(4)

	msg.FieldBin128(types.Bin128FromInt64(1))
	msg.FieldBin256(types.Bin256FromInt64(2))

	msg.FieldFloat32(10.0)
	msg.FieldFloat64(20.0)

	msg.FieldString("hello, world")
	msg.FieldBytes([]byte("abc"))

	msg.FieldStruct(Struct{
		Key:   123,
		Value: 456,
	})

	{
		node := msg.Node()
		node.Value("a")
		{
			next := node.Next()
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
		submsg := msg.Imported()
		submsg.Key("key")
		submsg.Value(pkg3.Value{})
		submsg.End()
	}

	{
		list := msg.ListInts()
		for _, x := range []int64{1, 2, 3} {
			list.Next(x)
		}
		if err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	{
		list := msg.ListStrings()
		for _, x := range []string{"a", "b", "c"} {
			list.Next(x)
		}
		if err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	{
		list := msg.ListMessages()
		for _, x := range []string{"1", "2"} {
			elem := list.Next()
			elem.Value(x)
			elem.End()
		}
		if err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	{
		list := msg.ListImported()
		for _, x := range []string{"a", "b"} {
			elem := list.Next()
			elem.Key(x)
			elem.End()
		}
		if err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	m, err := msg.Build()
	if err != nil {
		t.Fatal(err)
	}
	return m
}
