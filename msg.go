// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import "github.com/basecomplextech/spec/internal/types"

type (
	// Message is a raw message.
	Message = types.Message

	// MessageType is a type implemented by generated messages.
	MessageType = types.MessageType
)

// NewMessage returns a new message from bytes or an empty message when not a message.
func NewMessage(b []byte) Message {
	return types.NewMessage(b)
}

// NewMessageErr returns a new message from bytes or an error when not a message.
func NewMessageErr(b []byte) (Message, error) {
	return types.NewMessageErr(b)
}

// ParseMessage recursively parses and returns a message.
func ParseMessage(b []byte) (_ Message, size int, err error) {
	return types.ParseMessage(b)
}
