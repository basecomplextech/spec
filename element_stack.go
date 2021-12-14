package protocol

const elementSize = 4

type element struct {
	offset uint32
}

// elementStack acts as a buffer for nested list elements.
//
// Each list externally stores its start offset in the buffer, and provides the offset
// when inserting new elements.
//
// 	+-------------------+-------------------+-------------------+
//	|       list0       |      sublist1     |      sublist2     |
//	|-------------------|-------------------|-------------------|
//	| e0 | e1 | e2 | e3 | e0 | e1 | e2 | e3 | e0 | e1 | e2 | e3 |
//	+-------------------+-------------------+-------------------+
//
type elementStack []element

// offset returns the next list buffer offset.
func (s elementStack) offset() int {
	return len(s)
}

// push appends a new element to the last list.
func (sptr *elementStack) push(elem element) {
	sptr._grow(1)

	s := *sptr
	s = append(s, elem)
	*sptr = s
}

// popList pops a list starting at offset.
func (sptr *elementStack) popList(offset int) []element {
	s := *sptr
	list := s[offset:]

	s = s[:offset]
	*sptr = s
	return list
}

// grow grows element stack capacity to store at least n elements.
func (sptr *elementStack) _grow(n int) {
	s := *sptr

	// check remaining
	rem := cap(s) - len(s)
	if rem >= n {
		return
	}

	// grow
	size := (cap(s) * 2) + n
	next := make([]element, len(s), size)
	copy(next, s)

	*sptr = s
}
