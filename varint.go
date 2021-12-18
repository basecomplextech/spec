package spec

// maxReverseVarintLenN is the maximum length of a reverse varint-encoded N-bit integer.
const (
	maxReverseVarintLen16 = 3
	maxReverseVarintLen32 = 5
	maxReverseVarintLen64 = 10
)

// reverseUvarint decodes a uint64 from the buf in reverse order starting from
// the buf end and returns that value and the number of bytes read (> 0).
// If an error occurred, the value is 0 and the number of bytes n is <= 0 meaning:
//
// 	n == 0: buf too small
// 	n  < 0: value larger than 64 bits (overflow)
// 	        and -n is the number of bytes read
//
// This is a version of https://developers.google.com/protocol-buffers/docs/encoding#varints
// which encodes uint64 in reverse byte order.
//
func reverseUvarint(buf []byte) (uint64, int) {
	var result uint64
	var shift uint

	// slice last bytes upto max varint64 len
	if len(buf) > maxReverseVarintLen64 {
		buf = buf[len(buf)-maxReverseVarintLen64:]
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
		return result | uint64(b)<<shift, n + 1
	}

	return 0, 0
}

// putReverseUvarint encodes a uint64 into buf in reverse order starting from
// the buf end and returns the number of bytes written.
// If the buffer is too small, PutUvarint will panic.
func putReverseUvarint(buf []byte, v uint64) int {
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
