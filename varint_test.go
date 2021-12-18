package spec

import (
	"testing"
)

// reverseUvarint

func TestReverseUvarint(t *testing.T) {
	fn := func(v uint64) {
		buf := make([]byte, maxReverseVarintLen64)
		n := putReverseUvarint(buf, v)
		v1, m := reverseUvarint(buf[len(buf)-n:])

		if v != v1 {
			t.Errorf("reverseUvarint(%d): got %d", v, v1)
		}
		if n != m {
			t.Errorf("reverseUvarint(%d): got n = %d; want %d", v, m, n)
		}
	}

	tests := []int64{
		-1 << 63,
		-1<<63 + 1,
		-1,
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
		1<<63 - 1,
	}

	for _, v := range tests {
		fn(uint64(v))
	}
	for v := uint64(0x7); v != 0; v <<= 1 {
		fn(v)
	}
}

// reverseUvarint max

func TestReverseUvarint_max(t *testing.T) {
	fn := func(w uint, max int) {
		buf := make([]byte, maxReverseVarintLen64)
		n := putReverseUvarint(buf, 1<<w-1)
		if n != max {
			t.Errorf("invalid length, expected=%d, actual=%v", max, n)
		}
	}

	fn(16, maxReverseVarintLen16)
	fn(32, maxReverseVarintLen32)
	fn(64, maxReverseVarintLen64)
}
