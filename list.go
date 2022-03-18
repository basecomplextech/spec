package spec

type List[T any] struct {
	meta listMeta
	data []byte

	decode func(b []byte) (T, int, error)
}

// GetList decodes and returns a list without validation, or an empty list on error.
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
	w    *Writer
	next func(*Writer) W
}

// BeginList begins and returns a new list encoder.
func BeginList[W any](w *Writer, next func(*Writer) W) ListEncoder[W] {
	w.BeginList()

	return ListEncoder[W]{
		w:    w,
		next: next,
	}
}

// BeginElement an encoder for the next element.
func (e ListEncoder[W]) BeginElement() W {
	return e.next(e.w)
}

// EndElement ends an element.
func (e ListEncoder[W]) EndElement() error {
	return e.w.Element()
}

// Value

// ListValueEncoder encodes a list of primitive values.
type ListValueEncoder[T any] struct {
	w      *Writer
	encode func(el T) error
}

// BeginValueList begins and returns a new list encoder for primitive values.
func BeginValueList[T any](w *Writer, encode func(T) error) ListValueEncoder[T] {
	w.BeginList()

	return ListValueEncoder[T]{
		w:      w,
		encode: encode,
	}
}

// Element encodes the next element.
func (e ListValueEncoder[T]) Element(el T) error {
	e.encode(el)
	return e.w.Element()
}
