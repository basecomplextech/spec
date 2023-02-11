package pkg1

import (
	"github.com/complex1tech/baselibrary/basic"
	"github.com/complex1tech/spec/tests/pkg2"
	"github.com/complex1tech/spec/tests/pkg3/pkg3a"
)

type Object struct {
	Bool bool
	Byte byte

	Int16 int16
	Int32 int32
	Int64 int64

	Uint16 uint16
	Uint32 uint32
	Uint64 uint64

	Float32 float32
	Float64 float64

	Bin64  basic.Bin64
	Bin128 basic.Bin128
	Bin256 basic.Bin256

	String   string
	Bytes1   []byte
	Message1 map[uint16]int32

	Enum1      Enum
	Struct1    Struct
	Subobject  *Subobject
	Subobject1 *Subobject1

	Ints        []int64
	Strings     []string
	Structs     []Struct
	Subobjects  []*Subobject
	Subobjects1 []*Subobject1
}

type Subobject struct {
	Value string
	Next  *Subobject
}

type Subobject1 struct {
	Key   string
	Value pkg3a.Value
}

// Write

func (o *Object) Write(w MessageWriter) (Message, error) {
	w.Bool(o.Bool)
	w.Byte(o.Byte)

	w.Int16(o.Int16)
	w.Int32(o.Int32)
	w.Int64(o.Int64)

	w.Uint16(o.Uint16)
	w.Uint32(o.Uint32)
	w.Uint64(o.Uint64)

	w.Float32(o.Float32)
	w.Float64(o.Float64)

	w.Bin64(o.Bin64)
	w.Bin128(o.Bin128)
	w.Bin256(o.Bin256)

	w.String(o.String)
	w.Bytes1(o.Bytes1)

	w.Enum1(o.Enum1)
	w.Struct1(o.Struct1)

	if o.Message1 != nil {
		w1 := w.Message1()
		for tag, value := range o.Message1 {
			w1.Field(tag).Int32(value)
		}
		if err := w1.End(); err != nil {
			return Message{}, err
		}
	}

	if o.Subobject != nil {
		if _, err := o.Subobject.Write(w.Submessage()); err != nil {
			return Message{}, err
		}
	}
	if o.Subobject1 != nil {
		if _, err := o.Subobject1.Write(w.Submessage1()); err != nil {
			return Message{}, err
		}
	}

	if len(o.Ints) > 0 {
		list := w.Ints()
		for _, i := range o.Ints {
			list.Add(int64(i))
		}
		if err := list.End(); err != nil {
			return Message{}, err
		}
	}

	if len(o.Strings) > 0 {
		list := w.Strings()
		for _, s := range o.Strings {
			list.Add(s)
		}
		if err := list.End(); err != nil {
			return Message{}, err
		}
	}

	if len(o.Structs) > 0 {
		list := w.Structs()
		for _, s := range o.Structs {
			list.Add(s)
		}
		if err := list.End(); err != nil {
			return Message{}, err
		}
	}

	if len(o.Subobjects) > 0 {
		list := w.Submessages()
		for _, sub := range o.Subobjects {
			if _, err := sub.Write(list.Add()); err != nil {
				return Message{}, err
			}
		}
		if err := list.End(); err != nil {
			return Message{}, err
		}
	}

	if len(o.Subobjects1) > 0 {
		list := w.Submessages1()
		for _, sub := range o.Subobjects1 {
			if _, err := sub.Write(list.Add()); err != nil {
				return Message{}, err
			}
		}
		if err := list.End(); err != nil {
			return Message{}, err
		}
	}

	return w.Build()
}

func (o *Subobject) Write(w SubmessageWriter) (Submessage, error) {
	w.Value(o.Value)

	if o.Next != nil {
		if _, err := o.Next.Write(w.Next()); err != nil {
			return Submessage{}, err
		}
	}

	return w.Build()
}

func (o *Subobject1) Write(w pkg2.SubmessageWriter) (pkg2.Submessage, error) {
	w.Key(o.Key)
	w.Value(o.Value)
	return w.Build()
}
