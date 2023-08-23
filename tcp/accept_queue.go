package tcp

import (
	"sync"

	"github.com/basecomplextech/baselibrary/status"
)

type acceptQueue struct {
	mu sync.Mutex
	st status.Status

	streams  []*stream
	waitChan chan struct{}
}

func newAcceptQueue() *acceptQueue {
	return &acceptQueue{
		st:       status.OK,
		waitChan: make(chan struct{}),
	}
}

// close closes the queue.
func (q *acceptQueue) close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.st.OK() {
		return
	}

	q.st = status.End
	close(q.waitChan)
}

// poll returns the first stream in the queue or false.
func (q *acceptQueue) poll() (*stream, bool, status.Status) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.streams) == 0 {
		return nil, false, q.st
	}

	s := q.streams[0]
	copy(q.streams, q.streams[1:])
	q.streams = q.streams[:len(q.streams)-1]
	return s, true, status.OK
}

// push adds a stream to the queue.
func (q *acceptQueue) push(s *stream) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.st.OK() {
		panic("push to closed queue")
	}

	q.streams = append(q.streams, s)

	select {
	case q.waitChan <- struct{}{}:
	default:
	}
}

// wait returns a channel that is notified when the queue is not empty or closed.
func (q *acceptQueue) wait() <-chan struct{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	switch {
	case !q.st.OK():
		return closedChan
	case len(q.streams) > 0:
		return closedChan
	}

	select {
	case <-q.waitChan:
	default:
	}
	return q.waitChan
}

// closed chan

var closedChan = func() chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}()
