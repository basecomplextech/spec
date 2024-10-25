// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"time"

	"github.com/basecomplextech/baselibrary/units"
)

type Options struct {
	Client ClientOptions `json:"client"`
	Server ServerOptions `json:"server"`
	Conn   ConnOptions   `json:"conn"`
}

type ClientOptions struct {
	AutoConnect bool `json:"auto_connect"`

	// MaxConns is a maximum number of client connections, zero means one connection.
	MaxConns int `json:"max_conns"`

	// ConnChannels is a target number of channels per connection, zero means no limit.
	ConnChannels int `json:"conn_channels"`

	// DialTimeout is a client dial timeout.
	DialTimeout time.Duration `json:"dial_timeout"`
}

type ServerOptions struct{}

type ConnOptions struct {
	// Compress enables compression.
	Compress bool `json:"compress"`

	// ChannelWindowSize is an initial channel window size.
	ChannelWindowSize units.Bytes `json:"channel_window_size"`

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
		Client: ClientOptions{
			MaxConns:     4,
			ConnChannels: 128,
			DialTimeout:  2 * time.Second,
		},

		Conn: ConnOptions{
			Compress:          true,
			ChannelWindowSize: 16 * units.MiB,
			ReadBufferSize:    32 * units.KiB,
			WriteBufferSize:   32 * units.KiB,
			WriteQueueSize:    16 * units.MiB,
		},
	}
}

// internal

// clean cleans options, sets zero values to default values.
func (o Options) clean() Options {
	o.Conn = o.Conn.clean()
	return o
}

// clean cleans options, sets zero values to default values.
func (o ConnOptions) clean() ConnOptions {
	o1 := Default().Conn
	o1.Compress = o.Compress

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
