package encode

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/encoding/compactint"
	"github.com/basecomplextech/spec/internal/format"
)

func encodeSize(b buffer.Buffer, size uint32) int {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], size)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n)
	copy(buf, p[off:])

	return n
}

func encodeSizeType(b buffer.Buffer, size uint32, type_ format.Type) int {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], size)
	off := compactint.MaxLen32 - n

	buf := b.Grow(n + 1)
	copy(buf[:n], p[off:])
	buf[n] = byte(type_)

	return n + 1
}
