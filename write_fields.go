package protocol

type writeField struct {
	tag    uint16
	offset uint32
}

type writeFields []writeField

func (q writeFields) offset() int {
	return len(q)
}

func (q *writeFields) insert(offset int, field writeField) {
	// insertion sort
}

func (q *writeFields) pop(offset int) []writeField {
	return nil
}
