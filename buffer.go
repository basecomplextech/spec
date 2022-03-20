package spec

type Buffer interface {
	// Alloc grows an internal buffer capacity and returns an n-byte slice.
	// This buffer is intended to be appended to and passed to an immediately succeeding
	// Write call. The buffer is only valid until the next write operation on b.
	//
	// Usage:
	//
	//	p := b.Alloc(8)
	//	binary.BigEndian.PutUint64(p, 1234)
	//	b.Write(p)
	//
	Alloc(n int) []byte

	// Bytes returns a slice of written bytes.
	// The slice is valid for use only until the next buffer modification.
	Bytes() []byte

	// Len returns the number of bytes in the buffer; b.Len() == len(b.Bytes()).
	Len() int

	// Write appends bytes from p to the buffer.
	Write(p []byte) (n int, err error)

	// WriteBytes appends a single byte to the buffer.
	WriteByte(p byte) error

	// WriteString appends bytes from s to the buffer.
	WriteString(s string) (n int, err error)
}

// NewBuffer returns a new buffer and initializes it with a byte slice.
// The new buffer takes the ownership of the slice.
func NewBuffer(buf []byte) Buffer {
	return newBuffer(buf)
}

type buffer struct {
	buf []byte
}

func newBuffer(buf []byte) *buffer {
	return &buffer{buf: buf}
}

// Alloc grows an internal buffer capacity and returns an n-byte slice.
func (b *buffer) Alloc(n int) []byte {
	return b.alloc(n)
}

// Bytes returns a slice of written bytes.
// The slice is valid for use only until the next buffer modification.
func (b *buffer) Bytes() []byte {
	return b.buf
}

// Len returns the number of bytes in the buffer; b.Len() == len(b.Bytes()).
func (b *buffer) Len() int {
	return len(b.buf)
}

// Reset resets the buffer to be empty, but retains its internal byte slice.
func (b *buffer) Reset() {
	b.buf = b.buf[:0]
}

// Write appends bytes from p to the buffer.
func (b *buffer) Write(p []byte) (n int, err error) {
	buf := b.grow(len(p))
	n = copy(buf, p)
	return
}

// WriteBytes appends a single byte to the buffer.
func (b *buffer) WriteByte(p byte) error {
	buf := b.grow(1)
	buf[0] = p
	return nil
}

// WriteString appends bytes from s to the buffer.
func (b *buffer) WriteString(s string) (n int, err error) {
	buf := b.grow(len(s))
	n = copy(buf, s)
	return
}

// private

func (b *buffer) alloc(n int) []byte {
	cp := cap(b.buf)
	ln := len(b.buf)

	// alloc
	free := cp - ln
	if free < n {
		size := (cp * 2) + n
		buf := make([]byte, ln, size)
		copy(buf, b.buf)
		b.buf = buf
	}

	// return
	size := ln + n
	return b.buf[ln:size]
}

func (b *buffer) grow(n int) []byte {
	out := b.alloc(n)
	b.buf = b.buf[:len(b.buf)+n]
	return out
}
