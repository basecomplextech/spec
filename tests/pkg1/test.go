package pkg1

import (
	"fmt"
	"math"

	"github.com/complex1tech/baselibrary/tests"
	"github.com/complex1tech/baselibrary/types"
	"github.com/complex1tech/spec/tests/pkg2"
)

func TestMessage(t tests.T) Message {
	w := NewMessageWriter()
	w.Bool(true)
	w.Byte(255)

	w.Int32(math.MaxInt32)
	w.Int64(math.MaxInt64)

	w.Uint32(math.MaxUint32)
	w.Uint64(math.MaxUint64)

	w.Float32(math.MaxFloat32)
	w.Float64(math.MaxFloat64)

	w.Bin64(types.Bin64FromInt64(1))
	w.Bin128(types.Bin128FromInt64(2))
	w.Bin256(types.Bin256FromInt64(3))

	w.String("hello, world")
	w.Bytes1([]byte("goodbye, world"))

	w.Enum1(EnumOne)
	w.Struct1(TestStruct())
	TestSubmessage(t, w.Submessage(), 0)
	pkg2.TestSubmessage(t, w.Submessage1())

	{
		list := w.Ints()
		for i := 0; i < 10; i++ {
			list.Add(int64(i))
		}
		if err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	{
		list := w.Strings()
		for i := 0; i < 10; i++ {
			list.Add(fmt.Sprintf("%03d", i))
		}
		if err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	{
		list := w.Structs()
		for i := 0; i < 10; i++ {
			list.Add(Struct{
				Key:   int32(i),
				Value: -int32(i),
			})
		}
		if err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	{
		list := w.Submessages()
		for i := 0; i < 10; i++ {
			TestSubmessage(t, list.Add(), i)
		}
		if err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	{
		list := w.Submessages1()
		for i := 0; i < 10; i++ {
			pkg2.TestSubmessage(t, list.Add())
		}
		if err := list.End(); err != nil {
			t.Fatal(err)
		}
	}

	m, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestSubmessage(t tests.T, w SubmessageWriter, i int) Submessage {
	w.Value(fmt.Sprintf("value %03d", i))

	n, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestStruct() Struct {
	return Struct{
		Key:   1,
		Value: -1,
	}
}
