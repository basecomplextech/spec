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

// Encoder

// ListEncoder encodes a list of values.
type ListEncoder[T any] struct {
	e      *Encoder
	encode EncodeFunc[T]
}

// EncodeList begins and returns a new value list encoder.
func EncodeList[T any](e *Encoder, encode EncodeFunc[T]) (
	result ListEncoder[T], err error,
) {
	if err = e.BeginList(); err != nil {
		return
	}

	result = ListEncoder[T]{e: e, encode: encode}
	return
}

// End ends the list.
func (e ListEncoder[T]) End() ([]byte, error) {
	return e.e.End()
}

// Next encodes the next element.
func (e ListEncoder[T]) Next(value T) error {
	if err := EncodeValue(e.e, value, e.encode); err != nil {
		return err
	}
	return e.e.Element()
}

// Message encoder

// MessageListEncoder encodes a list of messages.
type MessageListEncoder[T any] struct {
	e    *Encoder
	next MessageEncoderFunc[T]
}

// EncodeMessageList begins and returns a new message list encoder.
func EncodeMessageList[T any](e *Encoder, next MessageEncoderFunc[T]) (
	result MessageListEncoder[T], err error,
) {
	if err = e.BeginList(); err != nil {
		return
	}

	result = MessageListEncoder[T]{e: e, next: next}
	return
}

// End ends the list.
func (e MessageListEncoder[T]) End() ([]byte, error) {
	return e.e.End()
}

// Next returns the next element encoder.
func (e MessageListEncoder[T]) Next() (result T, err error) {
	if err = e.e.BeginElement(); err != nil {
		return
	}
	return e.next(e.e)
}
