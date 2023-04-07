package pkg1

import (
	"fmt"
	"math"

	"github.com/complex1tech/baselibrary/bin"
	"github.com/complex1tech/baselibrary/tests"
	"github.com/complex1tech/spec/tests/pkg3/pkg3a"
)

func TestObject(t tests.T) *Object {
	message := map[uint16]int32{
		1: 1,
		2: 2,
		3: 3,
	}

	ints := make([]int64, 0, 10)
	for i := 0; i < 10; i++ {
		ints = append(ints, int64(i))
	}

	strings := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		s := fmt.Sprintf("hello, world %03d", i)
		strings = append(strings, s)
	}

	structs := make([]Struct, 0, 10)
	for i := 0; i < 10; i++ {
		s := Struct{
			Key:   int32(i),
			Value: -int32(i),
		}
		structs = append(structs, s)
	}

	subObjects := make([]*Subobject, 0, 10)
	for i := 0; i < 10; i++ {
		subObjects = append(subObjects, TestSubobject(i))
	}

	subObjects1 := make([]*Subobject1, 0, 10)
	for i := 0; i < 10; i++ {
		subObjects1 = append(subObjects1, TestSubobject1(i))
	}

	return &Object{
		Bool: true,
		Byte: 255,

		Int16: math.MaxInt16,
		Int32: math.MaxInt32,
		Int64: math.MaxInt64,

		Uint16: math.MaxUint16,
		Uint32: math.MaxUint32,
		Uint64: math.MaxUint64,

		Float32: math.MaxFloat32,
		Float64: math.MaxFloat64,

		Bin64:  bin.Bin64FromInt(1),
		Bin128: bin.Bin128FromInt(2),
		Bin256: bin.Bin256FromInt(3),

		String:   "hello, world",
		Bytes1:   []byte("goodbye, world"),
		Message1: message,

		Enum1:      EnumOne,
		Struct1:    TestStruct(),
		Subobject:  TestSubobject(0),
		Subobject1: TestSubobject1(0),

		Ints:        ints,
		Strings:     strings,
		Structs:     structs,
		Subobjects:  subObjects,
		Subobjects1: subObjects1,
	}
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
