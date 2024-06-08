package writer

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/pools"
	"github.com/basecomplextech/spec/encoding"
)

// writerState is a big pooled struct which holds an encoding state.
type writerState struct {
	buf         buffer.Buffer
	autoRelease bool // whether to release the writer state on close

	stack    stack
	elements listStack    // buffer for list element tables
	fields   messageStack // buffer for message field tables

	// Preallocated
	_stack    [14]stackEntry
	_elements [48]encoding.ListElement
	_fields   [48]encoding.MessageField
}

func newWriterState() *writerState {
	s := &writerState{}
	s.stack.stack = s._stack[:0]
	s.elements.stack = s._elements[:0]
	s.fields.stack = s._fields[:0]
	return s
}

func (s *writerState) init(b buffer.Buffer) {
	s.reset()
	s.buf = b
}

func (s *writerState) reset() {
	s.buf = nil

	s.stack.reset()
	s.elements.reset()
	s.fields.reset()
}

// state pool

var writerStatePool = pools.MakePool(newWriterState)

func acquireWriterState() *writerState {
	return writerStatePool.New()
}

func releaseWriterState(s *writerState) {
	s.reset()
	writerStatePool.Put(s)
}
