package spec

type List struct {
	buffer []byte
	table  listTable
	data   []byte
}

// Data returns the exact list bytes.
func (l List) Data() []byte {
	return l.buffer
}

// Element returns an element data by an index or false.
func (l List) Element(i int) (d Data) {
	off := l.table.offset(i)
	if off < 0 {
		return
	}
	b := l.data[:off]
	return ReadData(b)
}

// Len returns the number of elements in the list.
func (l List) Len() int {
	return l.table.count()
}
