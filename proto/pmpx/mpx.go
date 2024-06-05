//go:generate spec generate --skip-rpc .

package pmpx

import (
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec"
)

type SendMessageInput struct {
	ID    bin.Bin128
	Data  []byte
	Open  bool
	Close bool
	End   bool
}

func MakeSendMessage(dst spec.MessageWriter, input SendMessageInput) (Message, error) {
	w := NewMessageWriterTo(dst)
	id := input.ID
	data := input.Data
	open := input.Open
	close := input.Close
	end := input.End

	switch {
	// Open message
	case open:
		w.Code(Code_ChannelOpen)

		w1 := w.Open()
		w1.Id(id)
		if data != nil {
			w1.Data(data)
		}
		if end {
			w1.End_(end)
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

	// End message
	case end:
		w.Code(Code_ChannelEnd)

		w1 := w.End_()
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

func MakeSendWindow(dst spec.MessageWriter, id bin.Bin128, delta uint32) (Message, error) {
	return Message{}, nil
}
