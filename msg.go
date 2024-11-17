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

// OpenMessage opens and returns a message from bytes, or an empty message on error.
// The method decodes the message table, but not the fields, see [ParseMessage].
func OpenMessage(b []byte) Message {
	return types.OpenMessage(b)
}

// OpenMessageErr opens and returns a message from bytes, or an error.
// The method decodes the message table, but not the fields, see [ParseMessage].
func OpenMessageErr(b []byte) (Message, error) {
	return types.OpenMessageErr(b)
}

// ParseMessage recursively parses and returns a message.
func ParseMessage(b []byte) (_ Message, size int, err error) {
	return types.ParseMessage(b)
}
