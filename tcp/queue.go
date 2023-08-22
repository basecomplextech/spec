package tcp

import (
	"sync"

	"github.com/basecomplextech/baselibrary/status"
)

type queue struct {
	mu    sync.Mutex
	st    status.Status
	items [][]byte

	waitFlag bool // small performance optimization
	waitChan chan struct{}
}

func newQueue() *queue {
	return &queue{
		st:       status.OK,
		waitChan: make(chan struct{}, 1),
	}
}

func (q *queue) free() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.st.OK() {
		q.st = status.End
	}
	q.items = nil

	if q.waitFlag {
		select {
		case q.waitChan <- struct{}{}:
		default:
		}
	}
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

	if !q.st.OK() {
		return
	}
	q.st = status.End

	if q.waitFlag {
		select {
		case q.waitChan <- struct{}{}:
		default:
		}
	}
}

func (q *queue) closeWithError(st status.Status) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.st.OK() {
		return
	}
	q.st = st

	if q.waitFlag {
		select {
		case q.waitChan <- struct{}{}:
		default:
		}
	}
}

// append/next/wait

// append append a message to the end.
func (q *queue) append(msg []byte) status.Status {
	q.mu.Lock()
	defer q.mu.Unlock()

	switch q.st.Code {
	case status.CodeOK:
		// Pass
	case status.CodeEnd:
		return tcpErrorf("append to closed queue")
	default:
		return q.st
	}

	msg1 := make([]byte, len(msg))
	copy(msg1, msg)
	q.items = append(q.items, msg1)

	if q.waitFlag {
		select {
		case q.waitChan <- struct{}{}:
		default:
		}
	}
	return status.OK
}

// next moves to the next message and returns it or false.
func (q *queue) next() ([]byte, bool, status.Status) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.waitFlag = false

	if len(q.items) == 0 {
		return nil, false, q.st
	}

	msg := q.items[0]
	copy(q.items, q.items[1:])

	q.items[len(q.items)-1] = nil
	q.items = q.items[:len(q.items)-1]
	return msg, true, status.OK
}

// wait returns a channel which is notified when a message is available.
func (q *queue) wait() <-chan struct{} {
	q.mu.Lock()

	if len(q.items) > 0 || !q.st.OK() {
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
