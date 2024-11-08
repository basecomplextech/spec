// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/stretchr/testify/assert"
)

func testServer(t tests.T, handle HandleFunc) *server {
	opts := Default()
	return testServerOpts(t, handle, opts)
}

func testServerOpts(t tests.T, handle HandleFunc, opts Options) *server {
	logger := logging.TestLogger(t)
	server := newServer("localhost:0", handle, logger, opts)

	_, st := server.Start()
	if !st.OK() {
		t.Fatal(st)
	}

	cleanup := func() {
		select {
		case <-server.Stop():
		case <-time.After(time.Second):
			t.Fatal("stop timeout")
		}
	}
	t.Cleanup(cleanup)

	select {
	case <-server.listening.Wait():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}

	server.address = server.Address()
	return server
}

// request server

func testRequestServer(t tests.T) *server {
	opts := Default()
	return testRequestServerOpts(t, opts)
}

func testRequestServerOpts(t tests.T, opts Options) *server {
	handle := func(ctx Context, ch Channel) status.Status {
		msg, st := ch.Receive(ctx)
		if !st.OK() {
			return st
		}
		return ch.SendAndClose(ctx, msg)
	}

	return testServerOpts(t, handle, opts)
}

// connect

func testConnect(t tests.T, s *server) *conn {
	ctx := async.NoContext()
	addr := s.Address()

	c, st := Connect(ctx, addr, s.logger, s.options)
	if !st.OK() {
		t.Fatal(st)
	}
	return c.(*conn)
}

func testChannel(t tests.T, c Conn) *channel {
	ctx := async.NoContext()
	ch, st := c.Channel(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	return ch.(*channel)
}

// Open/Close

func TestOpenClose(t *testing.T) {
	handle := func(ctx Context, ch Channel) status.Status {
		for {
			msg, st := ch.Receive(ctx)
			if !st.OK() {
				return st
			}
			if st := ch.Send(ctx, msg); !st.OK() {
				return st
			}
		}
	}

	server := testServer(t, handle)
	conn := testConnect(t, server)
	defer conn.Free()

	ctx := async.NoContext()
	msg0 := []byte("hello, world")

	ch, st := conn.Channel(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	defer ch.Free()

	st = ch.Send(ctx, msg0)
	if !st.OK() {
		t.Fatal(st)
	}

	msg1, st := ch.Receive(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	if st := ch.SendAndClose(ctx, nil); !st.OK() {
		t.Fatal(st)
	}

	assert.Equal(t, msg0, msg1)
}
