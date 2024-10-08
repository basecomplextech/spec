// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"encoding/binary"
	"math"
	"unsafe"
)

const (
	messageFieldSmallSize = 1 + 2 // tag(1) + offset(2)
	messageFieldBigSize   = 2 + 4 // tag(2) + offset(4)
)

type MessageMeta struct {
	table messageTable

	data uint32 // message data size
	big  bool   // big/small table format
}

// Len returns the number of fields in the message.
func (m MessageMeta) Len() int {
	return m.table.count(m.big)
}

// DataSize returns the message data size.
func (m MessageMeta) DataSize() uint32 {
	return m.data
}

// Offset returns field end offset by a tag or -1.
func (m MessageMeta) Offset(tag uint16) int {
	if m.big {
		return m.table.offset_big(tag)
	} else {
		return m.table.offset_small(tag)
	}
}

// OffsetByIndex returns field end offset by an index or -1.
func (m MessageMeta) OffsetByIndex(i int) int {
	if m.big {
		return m.table.offsetByIndex_big(i)
	} else {
		return m.table.offsetByIndex_small(i)
	}
}

// Field returns a field by an index or false,
func (m MessageMeta) Field(i int) (MessageField, bool) {
	if m.big {
		return m.table.field_big(i)
	} else {
		return m.table.field_small(i)
	}
}

// MessageField specifies a tag and a value offset in a message byte array.
//
//	+----------+-------------------+
//	| tag(1/2) |    offset(2/4)    |
//	+----------+-------------------+
type MessageField struct {
	Tag    uint16
	Offset uint32
}

// internal

// messageTable is a serialized array of message fields ordered by tags.
// the serialization format depends on whether the message is big or small, see isBigMessage().
//
//	         field0                field1                field2
//	+---------------------+---------------------+---------------------+
//	|  tag0 |   offset0   |  tag1 |   offset1   |  tag2 |   offset3   |
//	+---------------------+---------------------+---------------------+
type messageTable []byte

// isBigList returns true if any field tag > uint8 or offset > uint16.
func isBigMessage(table []MessageField) bool {
	ln := len(table)
	if ln == 0 {
		return false
	}

	for i := ln - 1; i >= 0; i-- {
		field := table[i]

		switch {
		case field.Tag > math.MaxUint8:
			return true
		case field.Offset > math.MaxUint16:
			return true
		}
	}

	return false
}

// count returns the number of fields in the table.
func (t messageTable) count(big bool) int {
	var size int
	if big {
		size = messageFieldBigSize
	} else {
		size = messageFieldSmallSize
	}
	return len(t) / size
}

// offset

// offset_big returns a field end offset by a tag or -1.
//
// The method is an optimized version of offset_big_safe. Eliminating array bounds checks
// reduces the function call from 13.5 ns/op to 7.5 ns/op for a 100-field table.
// See benchmarks.
func (t messageTable) offset_big(tag uint16) int {
	if len(t) < messageFieldBigSize {
		return -1
	}

	n := len(t) / messageFieldBigSize
	ptr := unsafe.Pointer(&t[0])

	// Binary search table
	left, right := 0, (n - 1)
	for left <= right {
		// Middle
		middle := int(uint(left+right) >> 1) // avoid overflow

		// Offset
		off := middle * messageFieldBigSize
		ptr1 := unsafe.Add(ptr, off)

		// Current tag (uint16)
		var cur uint16
		{
			b0 := *(*byte)(ptr1)
			b1 := *(*byte)(unsafe.Add(ptr1, 1))
			cur = uint16(b1) | uint16(b0)<<8
		}

		// Check current
		switch {
		case cur < tag:
			left = middle + 1

		case cur > tag:
			right = middle - 1

		case cur == tag:
			// Read offset (uint32) after tag (uint16)
			b0 := *(*byte)(unsafe.Add(ptr1, 2))
			b1 := *(*byte)(unsafe.Add(ptr1, 3))
			b2 := *(*byte)(unsafe.Add(ptr1, 4))
			b3 := *(*byte)(unsafe.Add(ptr1, 5))
			return int(uint32(b3) |
				uint32(b2)<<8 |
				uint32(b1)<<16 |
				uint32(b0)<<24)
		}
	}

	return -1
}

func (t messageTable) offset_big_safe(tag uint16) int {
	size := messageFieldBigSize
	n := len(t) / size

	// Binary search table
	left, right := 0, (n - 1)
	for left <= right {
		// Middle
		middle := int(uint(left+right) >> 1) // avoid overflow

		// Offset
		off := middle * size
		b := t[off : off+size]

		// Current tag
		cur := binary.BigEndian.Uint16(b)

		// Check current
		switch {
		case cur < tag:
			left = middle + 1

		case cur > tag:
			right = middle - 1

		case cur == tag:
			// Read offset after tag
			return int(binary.BigEndian.Uint32(b[2:]))
		}
	}

	return -1
}

// offset_small returns a field end offset by a tag or -1.
//
// The method is an optimized version of offset_small_safe. Eliminating array bounds checks
// reduces the function call from 10 ns/op to 6.5 ns/op for a 100-field table.
// See benchmarks.
func (t messageTable) offset_small(tag uint16) int {
	if len(t) < messageFieldSmallSize {
		return -1
	}

	n := len(t) / messageFieldSmallSize
	ptr := unsafe.Pointer(&t[0])

	// Binary search table
	left, right := 0, (n - 1)
	for left <= right {
		// Middle
		middle := int(uint(left+right) >> 1) // avoid overflow

		// Offset
		off := middle * messageFieldSmallSize
		ptr1 := unsafe.Add(ptr, off)

		// Current tag (uint8)
		cur := uint16(*(*byte)(ptr1))

		// Check current
		switch {
		case cur < tag:
			left = middle + 1

		case cur > tag:
			right = middle - 1

		case cur == tag:
			// Read offset (uint16) after tag (uint8)
			b0 := *(*byte)(unsafe.Add(ptr1, 1))
			b1 := *(*byte)(unsafe.Add(ptr1, 2))
			return int(uint16(b1) | uint16(b0)<<8)
		}
	}

	return -1
}

func (t messageTable) offset_small_safe(tag uint16) int {
	size := messageFieldSmallSize
	n := len(t) / size

	// Binary search table
	left, right := 0, (n - 1)
	for left <= right {
		// Middle
		middle := int(uint(left+right) >> 1) // avoid overflow

		// Offset
		off := middle * size
		b := t[off : off+size]

		// Current tag
		cur := uint16(b[0])

		// Check current
		switch {
		case cur < tag:
			left = middle + 1

		case cur > tag:
			right = middle - 1

		case cur == tag:
			// Read offset after tag
			return int(binary.BigEndian.Uint16(b[1:]))
		}
	}

	return -1
}

// offsetByIndex

func (t messageTable) offsetByIndex_big(i int) int {
	size := messageFieldBigSize
	n := len(t) / size

	// Check count
	switch {
	case i < 0:
		return -1
	case i >= n:
		return -1
	}

	// Field offset
	off := i * size

	// Read end after tag
	return int(binary.BigEndian.Uint32(t[off+2:]))
}

func (t messageTable) offsetByIndex_small(i int) int {
	size := messageFieldSmallSize
	n := len(t) / size

	// Check count
	switch {
	case i < 0:
		return -1
	case i >= n:
		return -1
	}

	// Field offset
	off := i * size

	// Read end after tag
	return int(binary.BigEndian.Uint16(t[off+1:]))
}

// field

func (t messageTable) field_big(i int) (f MessageField, ok bool) {
	size := messageFieldBigSize
	n := len(t) / size

	// Check count
	switch {
	case i < 0:
		return
	case i >= n:
		return
	}

	off := i * size
	b := t[off : off+size]

	f = MessageField{
		Tag:    binary.BigEndian.Uint16(b),
		Offset: binary.BigEndian.Uint32(b[2:]),
	}

	ok = true
	return
}

func (t messageTable) field_small(i int) (f MessageField, ok bool) {
	size := messageFieldSmallSize
	n := len(t) / size

	// Check count
	switch {
	case i < 0:
		return
	case i >= n:
		return
	}

	off := i * size
	b := t[off : off+size]

	f = MessageField{
		Tag:    uint16(b[0]),
		Offset: uint32(binary.BigEndian.Uint16(b[1:])),
	}

	ok = true
	return
}

// fields parses the table and returns a slice of fields.
func (t messageTable) fields(big bool) []MessageField {
	n := t.count(big)

	result := make([]MessageField, 0, n)
	for i := 0; i < n; i++ {
		var field MessageField
		var ok bool

		if big {
			field, ok = t.field_big(i)
		} else {
			field, ok = t.field_small(i)
		}
		if !ok {
			continue
		}

		result = append(result, field)
	}
	return result
}
