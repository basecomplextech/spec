package protocol

// fieldStack acts as a buffer for nested message fields.
//
// Each message externally stores its start offset in the buffer, and provides the offset
// when inserting new fields. Message fields are kept sorted by tags using the insertion sort.
//
// 	+-------------------+-------------------+-------------------+
//	|      message0     |    submessage1    |    submessage2    |
//	|-------------------|-------------------|-------------------|
//	| f0 | f1 | f2 | f3 | f0 | f1 | f2 | f3 | f0 | f1 | f2 | f3 |
//	+-------------------+-------------------+-------------------+
//
type fieldStack []messageField

// offset returns the next message table buffer offset.
func (s fieldStack) offset() int {
	return len(s)
}

// insert inserts a new field into the last table starting at offset, keeps the table sorted.
func (sptr *fieldStack) insert(offset int, f messageField) {
	sptr._grow(1)

	// append new field
	s := *sptr
	s = append(s, f)
	*sptr = s

	// get table
	table := s[offset:]

	// walk table in reverse order
	// move new field to its position
	// using insertion sort
	for i := len(table) - 1; i > 0; i-- {
		left := table[i-1]
		right := table[i]

		if left.tag < right.tag {
			// sorted
			break
		}

		// swap fields
		table[i-1] = right
		table[i] = left
	}
}

// popTable pops a table starting at offset.
func (sptr *fieldStack) popTable(offset int) []messageField {
	s := *sptr
	table := s[offset:]

	s = s[:offset]
	*sptr = s
	return table
}

// grow grows field stack capacity to store at least n fields.
func (sptr *fieldStack) _grow(n int) {
	s := *sptr

	// check remaining
	rem := cap(s) - len(s)
	if rem >= n {
		return
	}

	// grow
	size := (cap(s) * 2) + n
	next := make([]messageField, len(s), size)
	copy(next, s)

	*sptr = s
}
