package spec

type List[T any] struct {
	meta   listMeta
	bytes  []byte
	decode func(b []byte) (T, int, error)
}

// GetList decodes and returns a list without recursive validation, or an empty list on error.
func GetList[T any](b []byte, decode func([]byte) (T, int, error)) List[T] {
	meta, n, err := decodeListMeta(b)
	if err != nil {
		return List[T]{}
	}
	bytes := b[len(b)-n:]

	l := List[T]{
		meta:   meta,
		bytes:  bytes,
		decode: decode,
	}
	return l
}

// DecodeList decodes, recursively validates and returns a list.
func DecodeList[T any](b []byte, decode func([]byte) (T, int, error)) (List[T], int, error) {
	meta, n, err := decodeListMeta(b)
	if err != nil {
		return List[T]{}, n, err
	}
	bytes := b[len(b)-n:]

	l := List[T]{
		meta:   meta,
		bytes:  bytes,
		decode: decode,
	}
	if err := l.Validate(); err != nil {
		return List[T]{}, n, err
	}
	return l, n, nil
}

// Bytes returns the exact list bytes.
func (l List[T]) Bytes() []byte {
	return l.bytes
}

// Count returns the number of elements in the list.
func (l List[T]) Count() int {
	return l.meta.count()
}

// ElementBytes returns an element or zero.
func (l List[T]) Element(i int) (result T) {
	start, end := l.meta.offset(i)
	switch {
	case start < 0:
		return
	case end > int(l.meta.data):
		return
	}

	b := l.bytes[start:end]
	result, _, _ = l.decode(b)
	return result
}

// ElementBytes returns element data or nil.
func (l List[T]) ElementBytes(i int) []byte {
	start, end := l.meta.offset(i)
	switch {
	case start < 0:
		return nil
	case end > int(l.meta.data):
		return nil
	}
	return l.bytes[start:end]
}

// Validate recursively validates the list.
func (l List[T]) Validate() error {
	n := l.Count()

	for i := 0; i < n; i++ {
		data := l.ElementBytes(i)
		if len(data) == 0 {
			continue
		}
		if _, _, err := DecodeValue(data); err != nil {
			return err
		}
	}
	return nil
}

// Builder

// ListBuilder builds a list of values.
type ListBuilder[T any] struct {
	e      *Encoder
	encode EncodeFunc[T]
}

// BuildList begins and returns a new value list builder.
func BuildList[T any](e *Encoder, encode EncodeFunc[T]) (_ ListBuilder[T], err error) {
	if err = e.BeginList(); err != nil {
		return
	}

	b := ListBuilder[T]{e: e, encode: encode}
	return b, nil
}

// Build ends and returns the list.
func (b ListBuilder[T]) Build() ([]byte, error) {
	return b.e.End()
}

// Next encodes the next element.
func (b ListBuilder[T]) Next(value T) error {
	if err := EncodeValue(b.e, value, b.encode); err != nil {
		return err
	}
	return b.e.Element()
}

// MessageListBuilder builds a list of messages.
type MessageListBuilder[T any] struct {
	e    *Encoder
	next func(e *Encoder) (T, error)
}

// BuildMessageList begins and returns a new message list builder.
func BuildMessageList[T any](e *Encoder, next func(e *Encoder) (T, error)) (
	_ MessageListBuilder[T], err error,
) {
	if err = e.BeginList(); err != nil {
		return
	}

	b := MessageListBuilder[T]{e: e, next: next}
	return b, nil
}

// Build ends and returns the list.
func (b MessageListBuilder[T]) Build() ([]byte, error) {
	return b.e.End()
}

// Next returns the next message builder.
func (b MessageListBuilder[T]) Next() (_ T, err error) {
	if err = b.e.BeginElement(); err != nil {
		return
	}
	return b.next(b.e)
}
