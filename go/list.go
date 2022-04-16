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
func DecodeList[T any](b []byte, decode func([]byte) (T, int, error)) (_ List[T], size int, err error) {
	meta, size, err := decodeListMeta(b)
	if err != nil {
		return
	}
	bytes := b[len(b)-size:]

	l := List[T]{
		meta:   meta,
		bytes:  bytes,
		decode: decode,
	}

	ln := l.Count()
	for i := 0; i < ln; i++ {
		elem := l.ElementBytes(i)
		if len(elem) == 0 {
			continue
		}
		if _, _, err = decode(elem); err != nil {
			return
		}
	}
	return l, size, nil
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

// Values converts a list into a slice.
func (l List[T]) Values() []T {
	result := make([]T, 0, l.meta.count())
	for i := 0; i < l.meta.count(); i++ {
		elem := l.Element(i)
		result = append(result, elem)
	}
	return result
}

// Builder

// ListBuilder builds a list of values.
type ListBuilder[T any] struct {
	e      *Encoder
	encode EncodeFunc[T]
}

// BuildList begins and returns a new value list builder.
func BuildList[T any](e *Encoder, encode EncodeFunc[T]) (_ ListBuilder[T]) {
	e.BeginList()
	return ListBuilder[T]{e: e, encode: encode}
}

// Len returns the number of elements in the builder.
func (b ListBuilder[T]) Len() int {
	return b.e.ListLen()
}

// Build ends and returns the list.
func (b ListBuilder[T]) Build() error {
	_, err := b.e.End()
	return err
}

// Next encodes the next element.
func (b ListBuilder[T]) Next(value T) error {
	if err := EncodeValue(b.e, value, b.encode); err != nil {
		return err
	}
	return b.e.Element()
}

// NestedListBuilder builds a list using nested element builder.
type NestedListBuilder[T any] struct {
	e    *Encoder
	next func(e *Encoder) T
}

// BuildNestedList begins and returns a new list.
func BuildNestedList[T any](e *Encoder, next func(e *Encoder) T) (_ NestedListBuilder[T]) {
	e.BeginList()
	return NestedListBuilder[T]{e: e, next: next}
}

// Err returns the current build error.
func (b NestedListBuilder[T]) Err() error {
	return b.e.err
}

// Len returns the number of elements in the builder.
func (b NestedListBuilder[T]) Len() int {
	return b.e.ListLen()
}

// Build ends and returns the list.
func (b NestedListBuilder[T]) Build() error {
	_, err := b.e.End()
	return err
}

// Next returns the next element builder.
func (b NestedListBuilder[T]) Next() (_ T) {
	b.e.BeginElement()
	return b.next(b.e)
}
