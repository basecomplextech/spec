package tcp

import "sync"

type messageBuffer struct {
	mu sync.RWMutex
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

// close closes the buffer for writing.
func (b *messageBuffer) close() {
}

// next moves to the next message and returns it or false.
func (b *messageBuffer) next() ([]byte, bool) {
	return nil, false
}

func (b *messageBuffer) wait() <-chan struct{} {
	return nil
}
