package spec

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func _BenchmarkSizeDistribution(b *testing.B) {
	msg := newTestObject()

	builder, err := BuildTestMessage()
	if err != nil {
		b.Fatal(err)
	}
	if err := msg.Encode(builder); err != nil {
		b.Fatal(err)
	}
	data, err := builder.End()
	if err != nil {
		b.Fatal(err)
	}

	msg1 := &TestObject{}
	if err := msg1.Decode(data); err != nil {
		b.Fatal(err)
	}
	require.Equal(b, msg, msg1)

	_, p, err := computeSizeDistribution(data)
	if err != nil {
		b.Fatal(err)
	}
	if p.size == 0 {
		b.Fatal()
	}

	b.Fatalf("%+v", p)
}

type sizeDistrib struct {
	// Total size
	size int

	// Total distribution
	meta   int
	tables int
	data   int

	// Meta distribution
	types int
	sizes int

	// Data distribution
	values  int
	bytes   int
	strings int
}

type sizeDistribPercent struct {
	// Total size
	size int

	// Total distribution
	meta   float32
	tables float32
	data   float32

	// Meta distribution
	types float32
	sizes float32

	// Data distribution
	values  float32
	bytes   float32
	strings float32
}

func computeSizeDistribution(b []byte) (*sizeDistrib, *sizeDistribPercent, error) {
	d := &sizeDistrib{}
	p := &sizeDistribPercent{}
	if len(b) == 0 {
		return d, p, nil
	}

	if err := _computeSizeDistribution(b, d); err != nil {
		return nil, nil, err
	}
	d.size = len(b)
	d.meta = d.types + d.sizes
	d.data = d.values + d.bytes + d.strings

	p.size = d.size
	p.meta = float32(d.meta) / float32(d.size)
	p.tables = float32(d.tables) / float32(d.size)
	p.data = float32(d.data) / float32(d.size)

	p.types = float32(d.types) / float32(d.size)
	p.sizes = float32(d.sizes) / float32(d.size)

	p.values = float32(d.values) / float32(d.size)
	p.bytes = float32(d.bytes) / float32(d.size)
	p.strings = float32(d.strings) / float32(d.size)
	return d, p, nil
}

func _computeSizeDistribution(b []byte, d *sizeDistrib) error {
	typ, n := decodeType(b)
	if n < 0 {
		return fmt.Errorf("invalid type")
	}
	d.types += n

	switch typ {
	case TypeTrue, TypeFalse:
		return nil

	case TypeByte:
		_, m, err := DecodeByte(b)
		if err != nil {
			return err
		}
		d.values += m - n

	case TypeInt32:
		_, m, err := DecodeInt32(b)
		if err != nil {
			return err
		}
		d.values += m - n

	case TypeInt64:
		_, m, err := DecodeInt64(b)
		if err != nil {
			return err
		}
		d.values += m - n

	case TypeUint32:
		_, m, err := DecodeUint32(b)
		if err != nil {
			return err
		}
		d.values += m - n

	case TypeUint64:
		_, m, err := DecodeUint64(b)
		if err != nil {
			return err
		}
		d.values += m - n

	case TypeFloat32, TypeFloat64:
		_, m := decodeFloat64(b)
		d.values += m - n

	case TypeBin128:
		_, m, err := DecodeBin128(b)
		if err != nil {
			return err
		}
		d.values += m - n

	case TypeBin256:
		_, m, err := DecodeBin256(b)
		if err != nil {
			return err
		}
		d.values += m - n

	case TypeBytes:
		off := len(b) - 1
		size, m := decodeSize(b[:off])
		if m < 0 {
			return fmt.Errorf("invalid bytes size")
		}

		d.sizes += m
		d.bytes += int(size)

	case TypeString:
		off := len(b) - 1
		size, m := decodeSize(b[:off])
		if n < 0 {
			return fmt.Errorf("invalid string size")
		}

		d.sizes += m
		d.strings += int(size)

	case TypeList, TypeBigList:
		off := len(b) - 1

		// Read table size
		tableSize, m := decodeSize(b[:off])
		if m < 0 {
			return fmt.Errorf("invalid list table size")
		}
		off -= m
		d.sizes += m
		d.tables += int(tableSize)

		// Read data size
		_, m = decodeSize(b[:off])
		if m < 0 {
			return fmt.Errorf("invalid list data size")
		}
		d.sizes += m

		// Read list
		list, _, err := DecodeList(b, ParseValue)
		if err != nil {
			return err
		}

		// Read elements
		for i := 0; i < list.Len(); i++ {
			elem := list.GetBytes(i)
			if err := _computeSizeDistribution(elem, d); err != nil {
				return err
			}
		}

	case TypeMessage, TypeBigMessage:
		off := len(b) - 1

		// Read table size
		tableSize, m := decodeSize(b[:off])
		if m < 0 {
			return fmt.Errorf("invalid message table size")
		}
		off -= m
		d.sizes += m
		d.tables += int(tableSize)

		// Read data size
		_, m = decodeSize(b[:off])
		if m < 0 {
			return fmt.Errorf("invalid message data size")
		}
		d.sizes += m

		// Read message meta
		msg, _, err := DecodeMessage(b)
		if err != nil {
			return err
		}

		// Read fields
		for i := 0; i < msg.Len(); i++ {
			field := msg.FieldByIndex(i)
			if err := _computeSizeDistribution(field, d); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unsupported type %d", typ)
	}

	return nil
}
