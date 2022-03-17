package spec

type List[T any] struct {
	list
	read func(b []byte) (T, int, error)
}

// NewList reads and returns list data, but does not validate its elements.
func NewList[T any](b []byte, read func([]byte) (T, int, error)) List[T] {
	list, _, err := readList(b)
	if err != nil {
		return List[T]{}
	}

	l := List[T]{
		list: list,
		read: read,
	}
	return l
}

// ReadList reads and returns list data, and recursively validates its elements.
func ReadList[T any](b []byte, read func([]byte) (T, int, error)) (List[T], int, error) {
	list, n, err := readList(b)
	if err != nil {
		return List[T]{}, n, err
	}

	l := List[T]{
		list: list,
		read: read,
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
	return l.table.count(l.big)
}

// ElementBytes returns an element or zero.
func (l List[T]) Element(i int) (result T) {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return
	case end > int(l.body):
		return
	}

	b := l.data[start:end]
	result, _, _ = l.read(b)
	return result
}

// ElementBytes returns element data or nil.
func (l List[T]) ElementBytes(i int) []byte {
	start, end := l.table.offset(l.big, i)
	switch {
	case start < 0:
		return nil
	case end > int(l.body):
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
