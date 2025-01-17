// Copyright 2021 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package types

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/decode"
	"github.com/basecomplextech/spec/internal/format"
)

// List is a raw list of elements.
type List struct {
	table format.ListTable
	bytes []byte
}

// OpenList opens and returns a list from bytes, or an empty list on error.
// The method decodes the list table, but not the elements.
func OpenList(b []byte) List {
	l, _, _ := decodeList(b)
	return l
}

// OpenListErr opens and returns a list from bytes, or an error.
// The method decodes the list table, but not the elements.
func OpenListErr(b []byte) (List, error) {
	l, _, err := decodeList(b)
	return l, err
}

// ParseList recursively parses and returns a list.
func ParseList(b []byte) (l List, size int, err error) {
	l, size, err = decodeList(b)
	if err != nil {
		return List{}, 0, err
	}

	ln := l.Len()
	for i := 0; i < ln; i++ {
		b1 := l.GetBytes(i)
		if len(b1) == 0 {
			continue
		}

		if _, _, err = ParseValue(b1); err != nil {
			return
		}
	}
	return l, size, nil
}

func decodeList(b []byte) (l List, size int, err error) {
	table, size, err := decode.DecodeListTable(b)
	if err != nil {
		return List{}, 0, err
	}
	bytes := b[len(b)-size:]

	l = List{
		table: table,
		bytes: bytes,
	}
	return l, size, nil
}

// Len returns the number of elements in the list.
func (l List) Len() int {
	return l.table.Len()
}

// Empty returns true if bytes are empty or list has no elements.
func (l List) Empty() bool {
	return len(l.bytes) == 0 || l.table.Len() == 0
}

// Raw returns the underlying list bytes.
func (l List) Raw() []byte {
	return l.bytes
}

// Elements

// Get returns an element at index i, panics on out of range.
func (l List) Get(i int) Value {
	start, end := l.table.Offset(i)
	if start < 0 {
		panic(fmt.Sprintf("index out of range: %d", i))
	}

	size := l.table.DataSize()
	if end > int(size) {
		return Value{}
	}
	return l.bytes[start:end]
}

// GetBytes returns element bytes at index i, panics on out of range.
func (l List) GetBytes(i int) []byte {
	start, end := l.table.Offset(i)
	if start < 0 {
		panic(fmt.Sprintf("index out of range: %d", i))
	}

	size := l.table.DataSize()
	if end > int(size) {
		return nil
	}

	return l.bytes[start:end]
}

// Clone

// List returns a list clone.
func (l List) Clone() List {
	b := make([]byte, len(l.bytes))
	copy(b, l.bytes)
	return OpenList(b)
}

// CloneTo clones a list into a slice.
func (l List) CloneTo(b []byte) List {
	ln := len(l.bytes)
	if cap(b) < ln {
		b = make([]byte, ln)
	}
	b = b[:ln]

	copy(b, l.bytes)
	return OpenList(b)
}
