package protocol

// object

type writeObjectType int

const (
	writeObjectList = iota
	writeObjectStruct
)

type writeObject struct {
	type_         writeObjectType
	dataOffset    int
	fieldOffset   int
	elementOffset int
}

// stack

type writeStack []writeObject

func (s writeStack) last() writeObject {
	ln := len(s)
	return s[ln-1]
}

func (s *writeStack) push(obj writeObject) {
	q := *s

	// realloc
	if cap(q) == len(q) {
		size := cap(q) * 2
		if size == 0 {
			size = 32
		}

		buf := make([]writeObject, cap(q), size)
		copy(buf, q)
		*s = buf
		q = *s
	}

	q = append(q, obj)
	*s = q
}

func (s *writeStack) pop() writeObject {
	q := *s
	ln := len(q)
	last := q[ln-1]

	*s = q[:ln-1]
	return last
}
