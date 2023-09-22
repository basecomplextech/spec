package tcp

import (
	"time"

	"github.com/basecomplextech/baselibrary/units"
)

type Options struct {
	// Compress enables compression.
	Compress bool `json:"compress"`

	// DialTimeout is a client dial timeout.
	DialTimeout time.Duration `json:"dial_timeout"`

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
	return Options{
		Compress:          true,
		DialTimeout:       2 * time.Second,
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
