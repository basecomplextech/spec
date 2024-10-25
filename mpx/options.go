// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"runtime"
	"time"

	"github.com/basecomplextech/baselibrary/units"
)

type Options struct {
	// Client

	// ClientConns is a maximum number of client connections, zero means one connection.
	ClientConns int

	// ConnChannels is a target number of channels per connection, zero means no limit.
	ConnChannels int

	// Connection

	// Compress enables compression.
	Compress bool `json:"compress"`

	// DialTimeout is a client dial timeout.
	DialTimeout time.Duration `json:"dial_timeout"`

	// Buffers

	// ChannelWindowSize is an initial channel window size.
	ChannelWindowSize units.Bytes `json:"channel_window_size"`

	// ReadBufferSize is a connection read buffer size.
	ReadBufferSize units.Bytes `json:"read_buffer_size"`

	// WriteBufferSize is a connection write buffer size.
	WriteBufferSize units.Bytes `json:"write_buffer_size"`

	// WriteQueueSize is a max connection write queue size (soft limit).
	WriteQueueSize units.Bytes `json:"write_queue_size"`
}

// Default returns default options.
func Default() Options {
	cpus := runtime.NumCPU()
	maxConns := max(4, cpus/4)

	return Options{
		ClientConns:  maxConns,
		ConnChannels: 128,

		Compress:    true,
		DialTimeout: 2 * time.Second,

		ChannelWindowSize: 16 * units.MiB,
		ReadBufferSize:    32 * units.KiB,
		WriteBufferSize:   32 * units.KiB,
		WriteQueueSize:    16 * units.MiB,
	}
}

// internal

// clean cleans options, sets zero values to default values.
func (o Options) clean() Options {
	o1 := Default()
	if o.ClientConns != 0 {
		o1.ClientConns = o.ClientConns
	}
	if o.ConnChannels != 0 {
		o1.ConnChannels = o.ConnChannels
	}

	o1.Compress = o.Compress
	o1.DialTimeout = o.DialTimeout

	if o.ChannelWindowSize != 0 {
		o1.ChannelWindowSize = o.ChannelWindowSize
	}
	if o.ReadBufferSize != 0 {
		o1.ReadBufferSize = o.ReadBufferSize
	}
	if o.WriteBufferSize != 0 {
		o1.WriteBufferSize = o.WriteBufferSize
	}
	if o.WriteQueueSize != 0 {
		o1.WriteQueueSize = o.WriteQueueSize
	}
	return o1
}
