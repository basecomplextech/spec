package spec

type List[T any] struct {
	meta listMeta
	data []byte

	decode func(b []byte) (T, int, error)
}

// GetList decodes and returns a list without recursive validation, or an empty list on error.
func GetList[T any](b []byte, decode func([]byte) (T, int, error)) List[T] {
	meta, n, err := decodeListMeta(b)
	if err != nil {
		return List[T]{}
	}
	data := b[len(b)-n:]

	l := List[T]{
		meta:   meta,
		data:   data,
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
	data := b[len(b)-n:]

	l := List[T]{
		meta:   meta,
		data:   data,
		decode: decode,
	}
	if err := l.Validate(); err != nil {
		return List[T]{}, n, err
	}
	return l, n, nil
}

// Data returns the exact list bytes.
func (l List[T]) Data() []byte {
	return l.data
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
	case end > int(l.meta.body):
		return
	}

	b := l.data[start:end]
	result, _, _ = l.decode(b)
	return result
}

// ElementBytes returns element data or nil.
func (l List[T]) ElementBytes(i int) []byte {
	start, end := l.meta.offset(i)
	switch {
	case start < 0:
		return nil
	case end > int(l.meta.body):
		return nil
	}
	return l.data[start:end]
}

// Validate recursively validates the list.
func (l List[T]) Validate() error {
	n := l.Count()

	for i := 0; i < n; i++ {
		data := l.ElementBytes(i)
		if len(data) == 0 {
			continue
		}
		if _, _, err := ReadValue(data); err != nil {
			return err
		}
	}
	return nil
}

// Encoder

// ListEncoder encodes a list.
type ListEncoder[W any] struct {
	e    *Encoder
	next func(*Encoder) W
}

// BeginList begins and returns a new list encoder.
func BeginList[W any](e *Encoder, next func(*Encoder) W) ListEncoder[W] {
	e.BeginList()

	return ListEncoder[W]{
		e:    e,
		next: next,
	}
}

// End ends the list.
func (e ListEncoder[W]) End() ([]byte, error) {
	return e.e.End()
}

// Next returns the next element encoder.
func (e ListEncoder[W]) Next() W {
	e.e.BeginElement()

	return e.next(e.e)
}

// Value encoder

// ListValueEncoder encodes a list of primitive values.
type ListValueEncoder[T any] struct {
	e      *Encoder
	encode func(el T) error
}

// BeginValueList begins and returns a new list encoder for primitive values.
func BeginValueList[T any](e *Encoder, encode func(T) error) ListValueEncoder[T] {
	e.BeginList()

	return ListValueEncoder[T]{
		e:      e,
		encode: encode,
	}
}

// End ends the list.
func (e ListValueEncoder[T]) End() ([]byte, error) {
	return e.e.End()
}

// Next encodes the next element.
func (e ListValueEncoder[T]) Next(el T) error {
	e.encode(el)
	return e.e.Element()
}
