// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

//go:generate spec generate --skip-rpc .

package pmpx

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
	"github.com/basecomplextech/spec"
)

// ConnectInput

type ConnectInput struct {
	Versions     []Version
	Compressions []ConnectCompression
}

func NewConnectInput() ConnectInput {
	return ConnectInput{Versions: []Version{Version_Version10}}
}

func (in ConnectInput) WithCompression(enabled bool) ConnectInput {
	if !enabled {
		return in
	}
	return in.WithCompressionLZ4()
}

func (in ConnectInput) WithCompressionLZ4() ConnectInput {
	in.Compressions = []ConnectCompression{ConnectCompression_Lz4}
	return in
}

func (in ConnectInput) Build() (Message, error) {
	return BuildConnectRequest(in)
}

// ConnectRequest

func BuildConnectRequest(input ConnectInput) (Message, error) {
	w := NewMessageWriter()
	w.Code(Code_ConnectRequest)
	w1 := w.ConnectRequest()

	// Versions
	{
		w2 := w1.Versions()
		for _, v := range input.Versions {
			w2.Add(v)
		}
		if err := w2.End(); err != nil {
			return Message{}, err
		}
	}

	// Compression
	{
		w2 := w1.Compression()
		for _, c := range input.Compressions {
			w2.Add(c)
		}
		if err := w2.End(); err != nil {
			return Message{}, err
		}
	}

	// Build
	if err := w1.End(); err != nil {
		return Message{}, err
	}
	return w.Build()
}

// ConnectResponse

func BuildConnectError(errorMessage string) (Message, error) {
	w := NewMessageWriter()
	w.Code(Code_ConnectResponse)

	w1 := w.ConnectResponse()
	w1.Ok(false)
	w1.Error(errorMessage)

	if err := w1.End(); err != nil {
		return Message{}, err
	}
	return w.Build()
}

func BuildConnectResponse(version Version, comp ConnectCompression) (Message, error) {
	w := NewMessageWriter()
	w.Code(Code_ConnectResponse)

	w1 := w.ConnectResponse()
	w1.Ok(true)
	w1.Version(version)
	w1.Compression(comp)

	if err := w1.End(); err != nil {
		return Message{}, err
	}
	return w.Build()
}

// Channel

type MessageInput struct {
	Id     bin.Bin128
	Data   []byte
	Window int32

	Open  bool
	Close bool
}

func BuildChannelMessage(buf alloc.Buffer, input MessageInput) (Message, error) {
	w := NewMessageWriterBuffer(buf)

	id := input.Id
	data := input.Data
	window := input.Window

	open := input.Open
	close := input.Close
	if open && close {
		panic("open and close cannot be true at the same time")
	}

	switch {
	case open:
		return BuildChannelOpen(w, id, data, window)
	case close:
		return BuildChannelClose(w, id, data)
	default:
		return BuildChannelData(w, id, data)
	}
}

func BuildChannelOpen(w MessageWriter, id bin.Bin128, data []byte, window int32) (Message, error) {
	w.Code(Code_ChannelOpen)

	w1 := w.ChannelOpen()
	w1.Id(id)
	w1.Window(window)

	if data != nil {
		w1.Data(data)
	}

	if err := w1.End(); err != nil {
		return Message{}, err
	}
	return w.Build()
}

func BuildChannelClose(w MessageWriter, id bin.Bin128, data []byte) (Message, error) {
	w.Code(Code_ChannelClose)

	w1 := w.ChannelClose()
	w1.Id(id)
	w1.Data(data)

	if err := w1.End(); err != nil {
		return Message{}, err
	}
	return w.Build()
}

func BuildChannelData(w MessageWriter, id bin.Bin128, data []byte) (Message, error) {
	w.Code(Code_ChannelData)

	w1 := w.ChannelData()
	w1.Id(id)
	w1.Data(data)

	if err := w1.End(); err != nil {
		return Message{}, err
	}
	return w.Build()
}

func BuildChannelWindow(w MessageWriter, id bin.Bin128, delta int32) (Message, error) {
	w.Code(Code_ChannelWindow)

	w1 := w.ChannelWindow()
	w1.Id(id)
	w1.Delta(delta)

	if err := w1.End(); err != nil {
		return Message{}, err
	}
	return w.Build()
}

// Batch

type BatchBuilder struct {
	w  MessageWriter
	w1 BatchWriter
	w2 spec.MessageListWriter[MessageWriter]
}

func NewBatchBuilder(buf alloc.Buffer) BatchBuilder {
	w := NewMessageWriterBuffer(buf)
	w.Code(Code_Batch)

	w1 := w.Batch()
	w2 := w1.List()

	return BatchBuilder{
		w:  w,
		w1: w1,
		w2: w2,
	}
}

func (b BatchBuilder) Open(id bin.Bin128, window int32) (BatchBuilder, error) {
	_, err := BuildChannelOpen(b.w2.Add(), id, nil, window)
	return b, err
}

func (b BatchBuilder) Close(id bin.Bin128) (BatchBuilder, error) {
	_, err := BuildChannelClose(b.w2.Add(), id, nil)
	return b, err
}

func (b BatchBuilder) Data(id bin.Bin128, data []byte) (BatchBuilder, error) {
	_, err := BuildChannelData(b.w2.Add(), id, data)
	return b, err
}

func (b BatchBuilder) Build() (Message, error) {
	if err := b.w2.End(); err != nil {
		return Message{}, err
	}
	if err := b.w1.End(); err != nil {
		return Message{}, err
	}
	return b.w.Build()
}
