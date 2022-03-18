package rvarint

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// varint

func TestInt64__should_encode_decode_varint(t *testing.T) {
	fn := func(v int64) {
		buf := make([]byte, MaxLen64)
		n := PutInt64(buf, v)
		v1, n1 := Int64(buf)
		if v != v1 {
			t.Errorf("Int64(%d): got %d", v, v1)
		}
		if n != n1 {
			t.Errorf("Int64(%d): expected n=%d; n=%d", v, n, n1)
		}
	}

	tests := []int64{
		0,
		1,
		2,
		10,
		20,
		63,
		64,
		65,
		127,
		128,
		129,
		255,
		256,
		257,
		math.MaxInt64 - 1,
		math.MaxInt64,

		-1,
		-2,
		-255,
		-256,
		-257,
		math.MinInt64 + 1,
		math.MinInt64,
	}

	for _, v := range tests {
		fn(v)
	}
}

// uvarint

func TestUint64__should_read_write_uvarint(t *testing.T) {
	fn := func(v uint64) {
		buf := make([]byte, MaxLen64)
		n := PutUint64(buf, v)
		v1, n1 := Uint64(buf)

		if v != v1 {
			t.Errorf("Uint64(%d): got %d", v, v1)
		}
		if n != n1 {
			t.Errorf("Uint64(%d): expected offset=%d; actual=%d", v, n, n1)
		}
	}

	tests := []uint64{
		0,
		1,
		2,
		10,
		20,
		63,
		64,
		65,
		127,
		128,
		129,
		255,
		256,
		257,
		math.MaxUint64 - 1,
		math.MaxUint64,
	}

	for _, v := range tests {
		fn(v)
	}
}

// max len

func TestUint64_max_length(t *testing.T) {
	fn := func(w uint, max int) {
		buf := make([]byte, MaxLen64)
		n := PutUint64(buf, 1<<w-1)
		if n != max {
			t.Errorf("invalid length, expected=%d, actual=%v", max, n)
		}
	}

	fn(16, MaxLen16)
	fn(32, MaxLen32)
	fn(64, MaxLen64)
}

// errors

func TestReadUint64__should_return_n_zero_on_small_buffer(t *testing.T) {
	b := []byte{}
	v, n := Uint64(b)
	assert.Equal(t, uint64(0), v)
	assert.Equal(t, 0, n)
}

func TestReadUint64__should_return_minus_n_on_overflow(t *testing.T) {
	b := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	v, n := Uint64(b)
	assert.Equal(t, uint64(0), v)
	assert.Equal(t, -8, n)
}
