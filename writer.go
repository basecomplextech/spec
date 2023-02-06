package spec

import (
	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/spec/writer"
)

type (
	Writer        = writer.Writer
	ListWriter    = writer.ListWriter
	MessageWriter = writer.MessageWriter
)

func NewWriter() *Writer {
	return writer.NewWriter()
}

func NewWriterBuffer(buffer buffer.Buffer) *Writer {
	return writer.NewWriterBuffer(buffer)
}
