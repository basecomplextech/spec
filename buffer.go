package spec

type Buffer interface {
	// Len returns the number of bytes in the buffer; b.Len() == len(b.Bytes()).
	Len() int

	// Bytes returns a slice of with the written bytes.
	// The slice is valid for use only until the next buffer mutation.
	Bytes() []byte

	// Mutation

	// Alloc grows an internal buffer `capacity` and returns an n-byte slice.
	// The slice is intended to be passed to an immediately succeeding Write call.
	// The slice is only valid until the next buffer mutation.
	//
	// Usage:
	//
	//	p := b.Alloc(10)
	//	n := varint.PutUint64(p, 1234)
	//	b.Write(p[:n])
	//
	// Alloc(n int) []byte

	// Grow grows and internal buffer `length` and returns an n-byte slice.
	// The slice be should be used directly is only valid until the next buffer mutation.
	//
	// Usage:
	//
	//	p := b.Grow(8)
	//	binary.BigEndian.PutUint64(p, 1234)
	//
	Grow(n int) []byte

	// Write appends bytes from p to the buffer.
	//
	// Equivalent to:
	//
	//	buf := b.Grow(n)
	//	copy(buf, p)
	//
	Write(p []byte) (n int, err error)

	// Reset resets the buffer to be empty, but retains its internal byte slice.
	Reset()
}

// NewBuffer returns a new buffer and initializes it with a byte slice.
// The new buffer takes the ownership of the slice.
func NewBuffer(buf []byte) Buffer {
	return newBuffer(buf)
}

// NewBuffer returns a new buffer and initializes it with a byte slice.
func NewBufferSize(size int) Buffer {
	buf := make([]byte, 0, size)
	return newBuffer(buf)
}

type buffer struct {
	buf []byte
}

func newBuffer(buf []byte) *buffer {
	return &buffer{buf: buf}
}

// Len returns the number of bytes in the buffer; b.Len() == len(b.Bytes()).
func (b *buffer) Len() int {
	return len(b.buf)
}

// Bytes returns a slice of with the written bytes.
func (b *buffer) Bytes() []byte {
	return b.buf
}

// Mutation

// Alloc grows an internal buffer capacity and returns an n-byte slice.
func (b *buffer) Alloc(n int) []byte {
	cp := cap(b.buf)
	ln := len(b.buf)

	// increase capacity
	free := cp - ln
	if free < n {
		size := (cp * 2) + n
		buf := make([]byte, ln, size)
		copy(buf, b.buf)
		b.buf = buf
	}

	// return slice
	size := ln + n
	return b.buf[ln:size]
}

// Grow grows and internal buffer length and returns an n-byte slice.
func (b *buffer) Grow(n int) []byte {
	p := b.Alloc(n)

	b.buf = b.buf[:len(b.buf)+n]
	return p
}

// Write appends bytes from p to the buffer.
func (b *buffer) Write(p []byte) (n int, err error) {
	buf := b.Grow(len(p))

	n = copy(buf, p)
	return
}

// Reset resets the buffer to be empty, but retains its internal byte slice.
func (b *buffer) Reset() {
	b.buf = b.buf[:0]
}
