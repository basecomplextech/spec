//go:generate spec generate --skip-rpc .

package pmpx

import (
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec"
)

type SendMessageInput struct {
	ID     bin.Bin128
	Data   []byte
	Window int32

	Open  bool
	Close bool
}

func MakeSendMessage(dst spec.MessageWriter, input SendMessageInput) (Message, error) {
	w := NewMessageWriterTo(dst)

	id := input.ID
	data := input.Data
	window := input.Window

	open := input.Open
	close := input.Close

	switch {
	// Open message
	case open:
		w.Code(Code_ChannelOpen)

		w1 := w.Open()
		w1.Id(id)
		w1.Window(window)

		if data != nil {
			w1.Data(data)
		}
		if close {
			w1.Close(close)
		}

		if err := w1.End(); err != nil {
			return Message{}, err
		}
		return w.Build()

	// Close message
	case close:
		w.Code(Code_ChannelClose)

		w1 := w.Close()
		w1.Id(id)
		if data != nil {
			w1.Data(data)
		}

		if err := w1.End(); err != nil {
			return Message{}, err
		}
		return w.Build()

	// Data message
	default:
		w.Code(Code_ChannelMessage)

		w1 := w.Message()
		w1.Id(id)
		w1.Data(data)
		if err := w1.End(); err != nil {
			return Message{}, err
		}
		return w.Build()
	}
}

func MakeSendWindow(dst spec.MessageWriter, id bin.Bin128, delta int32) (Message, error) {
	w := NewMessageWriterTo(dst)
	w.Code(Code_ChannelWindow)

	w1 := w.Window()
	w1.Id(id)
	w1.Delta(delta)
	if err := w1.End(); err != nil {
		return Message{}, err
	}
	return w.Build()
}
