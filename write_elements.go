package protocol

type writeElement uint32

type writeElements []uint32

func (e *writeElements) push(element uint32) {

}

func (e *writeElements) pop(offset int) []writeElement {
	return nil
}

func (e writeElements) offset() int {
	return len(e)
}
