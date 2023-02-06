package pkg1

import (
	"fmt"
	"math"

	"github.com/complex1tech/baselibrary/tests"
	"github.com/complex1tech/baselibrary/types"
	"github.com/complex1tech/spec/tests/pkg3/pkg3a"
)

func TestObject(t tests.T) *Object {
	o := &Object{
		Bool: true,
		Byte: 255,

		Int32: math.MaxInt32,
		Int64: math.MaxInt64,

		Uint32: math.MaxUint32,
		Uint64: math.MaxUint64,

		Float32: math.MaxFloat32,
		Float64: math.MaxFloat64,

		Bin64:  types.Bin64FromInt64(1),
		Bin128: types.Bin128FromInt64(2),
		Bin256: types.Bin256FromInt64(3),

		String: "hello, world",
		Bytes1: []byte("goodbye, world"),

		Enum1:      EnumOne,
		Struct1:    TestStruct(),
		Subobject:  TestSubobject(0),
		Subobject1: TestSubobject1(0),
	}

	for i := 0; i < 10; i++ {
		o.Ints = append(o.Ints, int64(i))
	}

	for i := 0; i < 10; i++ {
		o.Strings = append(o.Strings, fmt.Sprintf("%03d", i))
	}

	for i := 0; i < 10; i++ {
		o.Structs = append(o.Structs, Struct{
			Key:   int32(i),
			Value: -int32(i),
		})
	}

	for i := 0; i < 10; i++ {
		o.Subobjects = append(o.Subobjects, TestSubobject(i))
	}

	for i := 0; i < 10; i++ {
		o.Subobjects1 = append(o.Subobjects1, TestSubobject1(i))
	}

	return o
}

func TestSubobject(i int) *Subobject {
	return &Subobject{
		Value: fmt.Sprintf("value %03d", i),
	}
}

func TestSubobject1(i int) *Subobject1 {
	return &Subobject1{
		Key: fmt.Sprintf("key %03d", i),
		Value: pkg3a.Value{
			X: int32(i),
			Y: int32(-i),
		},
	}
}

func TestStruct() Struct {
	return Struct{
		Key:   1,
		Value: -1,
	}
}
