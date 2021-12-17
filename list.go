package spec

type List struct {
	buffer []byte
	table  listTable
}

func ReadList(buf []byte) List {
	type_, b := readType(buf)
	if type_ != TypeList {
		return List{}
	}

	tsize, b := readListTableSize(b)
	dsize, b := readListDataSize(b)
	table, _ := readListTable(b, tsize)
	buffer := readListBuffer(buf, tsize, dsize) // slice initial buf

	return List{
		buffer: buffer,
		table:  table,
	}
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
	b := l.buffer[:off]
	return ReadData(b)
}

// Len returns the number of elements in the list.
func (l List) Len() int {
	return l.table.count()
}
