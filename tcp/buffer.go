package tcp

type messageBuffer struct {
}

func newMessageBuffer() *messageBuffer {
	return nil
}

func (b *messageBuffer) free() {

}

func (b *messageBuffer) len() int {
	return 0
}

// append append a message to the end.
func (b *messageBuffer) append(msg []byte) {
}

// next moves to the next message and returns it or false.
func (b *messageBuffer) next() ([]byte, bool) {
	return nil, false
}
