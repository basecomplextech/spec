package rpc

import (
	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/spec/mpx"
)

type (
	// Context is an RPC context, which is an alias for mpx.Context.
	Context = mpx.Context

	// Options is RPC options, which are a type alias for mpx.Options.
	Options = mpx.Options
)

// Default returns default options.
func Default() Options {
	return mpx.Default()
}

// NewBuffer returns a new alloc.Buffer.
// The method is used in generated code.
func NewBuffer() *alloc.Buffer {
	return alloc.NewBuffer()
}
