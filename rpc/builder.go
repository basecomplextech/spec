package rpc

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/prpc"
)

type builder struct {
	buffer *alloc.Buffer
	writer spec.Writer
}

func newBuilder() builder {
	buffer := alloc.NewBuffer()
	writer := spec.NewWriterBuffer(buffer)

	return builder{
		buffer: buffer,
		writer: writer,
	}
}

func (b *builder) reset() {
	b.buffer.Reset()
	b.writer.Reset(b.buffer)
}

func (b *builder) buildMessage(data []byte) (prpc.Message, error) {
	b.reset()

	w := prpc.NewMessageWriterTo(b.writer.Message())
	w.Type(prpc.MessageType_Message)
	w.Msg(data)

	return w.Build()
}

func (b *builder) buildEnd() (prpc.Message, error) {
	b.reset()

	w := prpc.NewMessageWriterTo(b.writer.Message())
	w.Type(prpc.MessageType_End)

	return w.Build()
}

func (b *builder) buildRequest(req prpc.Request) (prpc.Message, error) {
	b.reset()

	w := prpc.NewMessageWriterTo(b.writer.Message())
	w.Type(prpc.MessageType_Request)
	w.CopyReq(req)

	return w.Build()
}

func (b *builder) buildResponse(result []byte, st status.Status) (prpc.Message, error) {
	b.reset()

	w := prpc.NewMessageWriterTo(b.writer.Message())
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
