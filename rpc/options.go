package rpc

import "github.com/basecomplextech/spec/tcp"

// Options is a type alias for SpecTCP options.
type Options = tcp.Options

// Default returns default options.
func Default() Options {
	return tcp.Default()
}
