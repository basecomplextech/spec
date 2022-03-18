package rvarint

// MaxLenN is the maximum length of a reverse varint-encoded N-bit integer.
const (
	MaxLen16 = 3
	MaxLen32 = 5
	MaxLen64 = 10
)

// decode

// Int64 decodes an int64 from the buf in reverse order starting from
// the buf end and returns that value and the number of read bytes.
// If an error occurred, the value is 0 and the number of bytes n is <= 0 meaning:
//
// 	n == 0: buf too small
// 	n  < 0: value larger than 64 bits (overflow)
// 	        and -n is the number of bytes read
//
func Int64(buf []byte) (int64, int) {
	ux, off := Uint64(buf) // ok to continue in presence of error
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, off
}

// Uint32 decodes a uint32 from the buf in reverse order starting from
// the buf end and returns that value and the number of read bytes.
// If an error occurred, the value is 0 and the number of bytes n is <= 0 meaning:
//
// 	n == 0: buf too small
// 	n  < 0: value larger than 64 bits (overflow)
// 	        and -n is the number of bytes read]
//
func Uint32(buf []byte) (uint32, int) {
	v, n := Uint64(buf)
	return uint32(v), n
}

// Uint64 decodes a uint64 from the buf in reverse order starting from
// the buf end and returns that value and the number of read bytes.
// If an error occurred, the value is 0 and the number of bytes n is <= 0 meaning:
//
// 	n == 0: buf too small
// 	n  < 0: value larger than 64 bits (overflow)
// 	        and -n is the number of bytes read
//
// This is a version of https://developers.google.com/protocol-buffers/docs/encoding#varints
// which encodes uint64 in reverse byte order.
//
func Uint64(buf []byte) (uint64, int) {
	var result uint64
	var shift uint

	// slice last bytes upto max varint64 len
	if len(buf) > MaxLen64 {
		buf = buf[len(buf)-MaxLen64:]
	}

	// iterate in reverse order
	var n int
	for i := len(buf) - 1; i >= 0; i-- {
		b := buf[i]

		// check if most significant bit (msb) is set
		// this indicates that there are further bytes to come
		if b >= 0b1000_0000 {
			if i == 0 {
				// overflow, this is the last byte
				// there can be no more bytes
				return 0, -(n + 1)
			}

			// disable msb and shift byte
			result |= uint64(b&0b0111_1111) << shift
			shift += 7
			n++
			continue
		}

		// no most significat bit (msb)
		// this is the last byte
		return result | uint64(b)<<shift, (n + 1)
	}

	return 0, 0
}

// encode

// PutInt64 encodes an int64 into buf in reverse order
// starting from the buf end and returns the number of written bytes.
// If the buffer is too small, PutInt64 will panic.
func PutInt64(buf []byte, x int64) int {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	return PutUint64(buf, ux)
}

// PutUint64 encodes a uint64 into buf in reverse order
// starting from the buf end and returns the number of written bytes.
// If the buffer is too small, PutUvarint will panic.
func PutUint64(buf []byte, v uint64) int {
	i := len(buf) - 1
	n := 0

	// while v is >= most significat bit
	for v >= 0b1000_0000 {
		// encode last 7 bit with msb set
		buf[i] = byte(v) | 0b1000_0000

		// shift by 7 bit
		v >>= 7

		i--
		n++
	}

	// last byte without msb
	buf[i] = byte(v)
	return n + 1
}
