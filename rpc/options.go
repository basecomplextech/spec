package rpc

import "github.com/basecomplextech/spec/mpx"

// Options is a type alias for SpecTCP options.
type Options = mpx.Options

// Default returns default options.
func Default() Options {
	return mpx.Default()
}
