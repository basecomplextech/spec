package mpx

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec"
	"github.com/basecomplextech/spec/proto/pmpx"
)

type builder struct {
	buffer *alloc.Buffer
	writer spec.Writer
}

type messageInput struct {
	id     bin.Bin128
	data   []byte
	window int32

	open  bool
	close bool
}

func newBuilder() *builder {
	buffer := alloc.NewBuffer()
	writer := spec.NewWriterBuffer(buffer)

	return &builder{
		buffer: buffer,
		writer: writer,
	}
}

func (b *builder) reset() {
	b.buffer.Reset()
	b.writer.Reset(b.buffer)
}

func (b *builder) buildMessage(input messageInput) (pmpx.Message, error) {
	b.reset()
	w := pmpx.NewMessageWriterTo(b.writer.Message())

	id := input.id
	data := input.data
	window := input.window

	open := input.open
	close := input.close
	if open && close {
		panic("open and close cannot be true at the same time")
	}

	switch {
	// Open message
	case open:
		w.Code(pmpx.Code_ChannelOpen)

		w1 := w.Open()
		w1.Id(id)
		w1.Window(window)

		if data != nil {
			w1.Data(data)
		}

		if err := w1.End(); err != nil {
			return pmpx.Message{}, err
		}
		return w.Build()

	// Close message
	case close:
		w.Code(pmpx.Code_ChannelClose)

		w1 := w.Close()
		w1.Id(id)
		if data != nil {
			w1.Data(data)
		}

		if err := w1.End(); err != nil {
			return pmpx.Message{}, err
		}
		return w.Build()

	// Data message
	default:
		w.Code(pmpx.Code_ChannelMessage)

		w1 := w.Message()
		w1.Id(id)
		w1.Data(data)
		if err := w1.End(); err != nil {
			return pmpx.Message{}, err
		}
		return w.Build()
	}
}

func (b *builder) buildWindow(id bin.Bin128, delta int32) (pmpx.Message, error) {
	b.reset()

	w := pmpx.NewMessageWriterTo(b.writer.Message())
	w.Code(pmpx.Code_ChannelWindow)

	w1 := w.Window()
	w1.Id(id)
	w1.Delta(delta)
	if err := w1.End(); err != nil {
		return pmpx.Message{}, err
	}
	return w.Build()

}
