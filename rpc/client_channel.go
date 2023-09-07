package rpc

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/basecomplextech/spec/tcp"
)

// Channel is a client RPC channel.
type Channel interface {
	// Request sends a request to the server.
	Request(cancel <-chan struct{}, req prpc.Request) status.Status

	// Response receives a response and returns its status and result if status is OK.
	Response(cancel <-chan struct{}) (*alloc.Buffer, status.Status)

	// Internal

	// Free frees the channel.
	Free()
}

// internal

var _ Channel = (*channel)(nil)

type channel struct {
	tconn tcp.Conn
	tchan tcp.Channel

	freed bool
	*channelState
}

type channelState struct {
	wlock  async.Lock
	wreq   bool // request sent
	wbuf   *alloc.Buffer
	writer spec.Writer

	rlock async.Lock
}

func newChannel(tconn tcp.Conn, tchan tcp.Channel) Channel {
	s := acquireState()

	return &channel{
		tconn: tconn,
		tchan: tchan,

		channelState: s,
	}
}

// Request sends a request to the server.
func (ch *channel) Request(cancel <-chan struct{}, req prpc.Request) status.Status {
	select {
	case <-ch.wlock:
	case <-cancel:
		return status.Cancelled
	}
	defer ch.wlock.Unlock()

	if ch.wreq {
		return Error("request already sent")
	}

	// Make request
	var bytes []byte
	{
		ch.wbuf.Reset()
		ch.writer.Reset(ch.wbuf)

		w := prpc.NewMessageWriterTo(ch.writer.Message())
		w.Type(prpc.MessageType_Request)
		w.CopyReq(req)

		msg, err := w.Build()
		if err != nil {
			return WrapError(err)
		}

		bytes = msg.Unwrap().Raw()
	}

	// Send request
	ch.wreq = true
	return ch.tchan.Write(cancel, bytes)
}

// Response receives a response and returns its status and result if status is OK.
func (ch *channel) Response(cancel <-chan struct{}) (*alloc.Buffer, status.Status) {
	select {
	case <-ch.rlock:
	case <-cancel:
		return nil, status.Cancelled
	}
	defer ch.rlock.Unlock()

	// Read message
	bytes, st := ch.tchan.Read(cancel)
	if !st.OK() {
		return nil, st
	}

	// Parse message
	msg, _, err := prpc.ParseMessage(bytes)
	if err != nil {
		return nil, WrapError(err)
	}

	// Check response
	typ := msg.Type()
	if typ != prpc.MessageType_Response {
		return nil, Errorf("unexpected message type %d, expected response type %d",
			typ, prpc.MessageType_Response)
	}
	resp := msg.Resp()

	// Parse status
	st = parseStatus(resp.Status())
	if !st.OK() {
		return nil, st
	}

	// Copy result
	res := resp.Result()
	buf := alloc.NewBufferSize(len(res))
	buf.Write(res)
	return buf, st
}

// Internal

// Free frees the channel.
func (ch *channel) Free() {
	defer ch.tconn.Free()
	defer ch.tchan.Free()

	s := ch.channelState
	ch.channelState = nil
	releaseState(s)
}

// state pool

var statePool = &sync.Pool{}

func acquireState() *channelState {
	v := statePool.Get()
	if v != nil {
		return v.(*channelState)
	}

	return newChannelState()
}

func releaseState(s *channelState) {
	s.reset()
	statePool.Put(s)
}

func newChannelState() *channelState {
	wbuf := alloc.NewBuffer()
	return &channelState{
		wlock:  async.NewLock(),
		wbuf:   wbuf,
		writer: spec.NewWriterBuffer(wbuf),
		rlock:  async.NewLock(),
	}
}

func (s *channelState) reset() {
	select {
	case s.wlock <- struct{}{}:
	default:
	}

	select {
	case s.rlock <- struct{}{}:
	default:
	}

	s.wreq = false
	s.wbuf.Reset()
	s.writer.Reset(s.wbuf)
}
