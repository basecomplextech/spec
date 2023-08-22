package tcp

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/ptcp"
)

type writeQueue struct {
	queue *queue

	mu     sync.Mutex
	buf    *alloc.Buffer
	writer spec.Writer
}

func newWriteQueue() *writeQueue {
	buf := alloc.NewBuffer()

	return &writeQueue{
		queue: newQueue(),

		buf:    buf,
		writer: spec.NewWriterBuffer(buf),
	}
}

func (q *writeQueue) free() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue.free()
	q.queue = nil

	q.buf.Free()
	q.buf = nil
}

func (q *writeQueue) writeOpenStream(id bin.Bin128, data []byte) status.Status {
	q.mu.Lock()
	defer q.mu.Unlock()

	var msg ptcp.Message
	{
		q.buf.Reset()
		q.writer.Reset(q.buf)

		w := ptcp.NewMessageWriterTo(q.writer.Message())
		w.Code(ptcp.Code_OpenStream)

		w1 := w.Open()
		w1.Id(id)
		w1.Data(data)
		if err := w1.End(); err != nil {
			return tcpError(err)
		}

		var err error
		msg, err = w.Build()
		if err != nil {
			return tcpError(err)
		}
	}

	b := msg.Unwrap().Raw()
	return q.queue.append(b)
}

func (q *writeQueue) writeCloseStream(id bin.Bin128) status.Status {
	q.mu.Lock()
	defer q.mu.Unlock()

	// TODO: Close stream on error
	var msg ptcp.Message
	{
		q.buf.Reset()
		q.writer.Reset(q.buf)

		w := ptcp.NewMessageWriterTo(q.writer.Message())
		w.Code(ptcp.Code_CloseStream)

		w1 := w.Open()
		w1.Id(id)
		if err := w1.End(); err != nil {
			return tcpError(err)
		}

		var err error
		msg, err = w.Build()
		if err != nil {
			return tcpError(err)
		}
	}

	b := msg.Unwrap().Raw()
	return q.queue.append(b)
}

func (q *writeQueue) writeStreamMessage(id bin.Bin128, data []byte) status.Status {
	q.mu.Lock()
	defer q.mu.Unlock()

	var msg ptcp.Message
	{
		q.buf.Reset()
		q.writer.Reset(q.buf)

		w := ptcp.NewMessageWriterTo(q.writer.Message())
		w.Code(ptcp.Code_StreamMessage)

		w1 := w.Message()
		w1.Id(id)
		w1.Data(data)
		if err := w1.End(); err != nil {
			return tcpError(err)
		}

		var err error
		msg, err = w.Build()
		if err != nil {
			return tcpError(err)
		}
	}

	b := msg.Unwrap().Raw()
	return q.queue.append(b)
}
