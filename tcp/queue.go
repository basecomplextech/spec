package tcp

import (
	"sync"

	"github.com/basecomplextech/baselibrary/status"
)

type queue struct {
	mu     sync.Mutex
	items  [][]byte
	closed bool

	waitFlag bool // small performance optimization
	waitChan chan struct{}
}

func newQueue() *queue {
	return &queue{
		waitChan: make(chan struct{}, 1),
	}
}

func (q *queue) free() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = nil
}

func (q *queue) len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}

// close

func (q *queue) close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return
	}

	q.closed = true

	if q.waitFlag {
		select {
		case q.waitChan <- struct{}{}:
		default:
		}
	}
}

func (q *queue) closeWithError(st status.Status) {

}

// append/next/wait

// append append a message to the end.
func (q *queue) append(msg []byte) {
	q.mu.Lock()
	defer q.mu.Unlock()

	msg1 := make([]byte, len(msg))
	copy(msg1, msg)
	q.items = append(q.items, msg1)

	if q.waitFlag {
		select {
		case q.waitChan <- struct{}{}:
		default:
		}
	}
}

// next moves to the next message and returns it or false.
func (q *queue) next() ([]byte, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.waitFlag = false

	if len(q.items) == 0 {
		return nil, false
	}

	msg := q.items[0]
	copy(q.items, q.items[1:])

	q.items[len(q.items)-1] = nil
	q.items = q.items[:len(q.items)-1]
	return msg, true
}

func (q *queue) wait() <-chan struct{} {
	q.mu.Lock()

	if len(q.items) > 0 || q.closed {
		q.mu.Unlock()
		return closedChan
	}

	q.waitFlag = true
	q.mu.Unlock()
	return q.waitChan
}

// util

var closedChan = func() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}()
