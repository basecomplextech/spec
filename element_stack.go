package protocol

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
type elementStack []listElement

// offset returns the next list buffer offset.
func (s elementStack) offset() int {
	return len(s)
}

// push appends a new element to the last list.
func (sptr *elementStack) push(elem listElement) {
	sptr._grow(1)

	s := *sptr
	s = append(s, elem)
	*sptr = s
}

// popList pops a list starting at offset.
func (sptr *elementStack) popList(offset int) []listElement {
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
	next := make([]listElement, len(s), size)
	copy(next, s)

	*sptr = s
}
