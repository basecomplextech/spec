// Copyright 2024 Ivan Korobkov. All rights reserved.

package rpc

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
)

type builder struct{}

func newBuilder() builder {
	return builder{}
}

func (b builder) buildMessage(buf alloc.Buffer, data []byte) (prpc.Message, error) {
	w := prpc.NewMessageWriterBuffer(buf)
	w.Type(prpc.MessageType_Message)
	w.Msg(data)

	return w.Build()
}

func (b builder) buildEnd(buf alloc.Buffer) (prpc.Message, error) {
	w := prpc.NewMessageWriterBuffer(buf)
	w.Type(prpc.MessageType_End)

	return w.Build()
}

func (b builder) buildRequest(buf alloc.Buffer, req prpc.Request) (prpc.Message, error) {
	w := prpc.NewMessageWriterBuffer(buf)
	w.Type(prpc.MessageType_Request)
	w.CopyReq(req)

	return w.Build()
}

func (b builder) buildResponse(buf alloc.Buffer, result []byte, st status.Status) (prpc.Message, error) {
	w := prpc.NewMessageWriterBuffer(buf)
	w.Type(prpc.MessageType_Response)

	w1 := w.Resp()
	w2 := w1.Status()
	w2.Code(string(st.Code))
	w2.Message(st.Message)
	if err := w2.End(); err != nil {
		return prpc.Message{}, nil
	}
	if result != nil {
		w1.Result().Any(result)
	}
	if err := w1.End(); err != nil {
		return prpc.Message{}, err
	}
	return w.Build()
}
