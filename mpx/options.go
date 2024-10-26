// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"time"

	"github.com/basecomplextech/baselibrary/units"
)

type Options struct {
	// ClientMaxConns is a maximum number of client connections, zero means one connection.
	ClientMaxConns int `json:"client_max_conns"`

	// ClientConnChannels is a target number of channels per connection, zero means no limit.
	ClientConnChannels int `json:"client_conn_channels"`

	// ClientDialTimeout is a client dial timeout.
	ClientDialTimeout time.Duration `json:"client_dial_timeout"`

	// Protocol

	// Compression enables compression.
	Compression bool `json:"compress"`

	// ChannelWindowSize is an initial channel window size.
	ChannelWindowSize units.Bytes `json:"channel_window_size"`

	// Buffers

	// ReadBufferSize is a connection read buffer size.
	ReadBufferSize units.Bytes `json:"read_buffer_size"`

	// WriteBufferSize is a connection write buffer size.
	WriteBufferSize units.Bytes `json:"write_buffer_size"`

	// WriteQueueSize is a max connection write queue size (soft limit).
	WriteQueueSize units.Bytes `json:"write_queue_size"`
}

// Default

// Default returns the default options.
func Default() Options {
	return Options{
		ClientMaxConns:     4,
		ClientConnChannels: 128,
		ClientDialTimeout:  2 * time.Second,

		Compression:       true,
		ChannelWindowSize: 16 * units.MiB,

		ReadBufferSize:  32 * units.KiB,
		WriteBufferSize: 32 * units.KiB,
		WriteQueueSize:  16 * units.MiB,
	}
}

// Merge merges non-zero values from another options and returns new options.
func (o Options) Merge(o1 Options) Options {
	o.ClientMaxConns = nonzero(o.ClientMaxConns, o1.ClientMaxConns)
	o.ClientConnChannels = nonzero(o.ClientConnChannels, o1.ClientConnChannels)
	o.ClientDialTimeout = nonzero(o.ClientDialTimeout, o1.ClientDialTimeout)

	o.Compression = o1.Compression
	o.ChannelWindowSize = nonzero(o.ChannelWindowSize, o1.ChannelWindowSize)

	o.ReadBufferSize = nonzero(o.ReadBufferSize, o1.ReadBufferSize)
	o.WriteBufferSize = nonzero(o.WriteBufferSize, o1.WriteBufferSize)
	o.WriteQueueSize = nonzero(o.WriteQueueSize, o1.WriteQueueSize)
	return o
}

// internal

// clean cleans options, sets zero values to default values.
func (o Options) clean() Options {
	return Default().Merge(o)
}

// util

func nonzero[T comparable](a, b T) T {
	var zero T
	if b != zero {
		return b
	}
	return a
}
