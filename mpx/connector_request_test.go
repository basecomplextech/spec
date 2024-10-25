// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/stretchr/testify/assert"
)

func testRequestConnector(t tests.T, s *server) *requestConnector {
	addr := s.Address()
	c := newRequestConnector(addr, s.logger, s.options)

	t.Cleanup(func() {
		c.close()
	})
	return c
}

// connected

func TestRequestConnector__should_set_connected_flag_on_conn_opened(t *testing.T) {
	server := testRequestServer(t)
	ctx := async.NoContext()
	c := testRequestConnector(t, server)

	_, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	assert.True(t, c.connected().Get())
	assert.False(t, c.disconnected().Get())
}

// disconnected

func TestRequestConnector__should_set_disconnected_flag_when_all_conns_closed(t *testing.T) {
	server := testRequestServer(t)
	ctx := async.NoContext()
	c := testRequestConnector(t, server)

	conn, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	conn.Close()

	select {
	case <-c.disconnected().Wait():
	case <-time.After(time.Second):
		t.Fatal("disconnect timeout")
	}

	assert.False(t, c.connected().Get())
	assert.True(t, c.disconnected().Get())
}

// conn

func TestRequestConnector_conn__should_open_connection(t *testing.T) {
	server := testRequestServer(t)
	ctx := async.NoContext()
	c := testRequestConnector(t, server)

	conn, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	conn.Close()
}

func TestRequestConnector_conn__should_return_existing_connection(t *testing.T) {
	server := testRequestServer(t)
	ctx := async.NoContext()
	c := testRequestConnector(t, server)

	conn, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	conn1, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Same(t, conn, conn1)
}

func TestRequestConnector_conn__should_reconnect_when_connection_closed(t *testing.T) {
	server := testRequestServer(t)
	ctx := async.NoContext()
	c := testRequestConnector(t, server)

	conn, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}

	conn.Close()
	select {
	case <-conn.Closed().Wait():
	case <-time.After(time.Second):
		t.Fatal("close timeout")
	}

	conn1, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.NotSame(t, conn, conn1)
}

func TestRequestConnector_conn__should_open_more_connections_when_channels_target_reached(t *testing.T) {
	server := testRequestServer(t)
	ctx := async.NoContext()

	c := testRequestConnector(t, server)
	c.opts.ClientConns = 2
	c.opts.ConnChannels = 1

	conn, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	if _, st := conn.Channel(ctx); !st.OK() {
		t.Fatal(st)
	}
	if _, st := conn.Channel(ctx); !st.OK() {
		t.Fatal(st)
	}

	conn1, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.Same(t, conn, conn1)

	time.Sleep(50 * time.Millisecond)

	conn2, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	assert.NotSame(t, conn1, conn2)
}

// close

func TestRequestConnector_close__should_close_connections(t *testing.T) {
	server := testRequestServer(t)
	ctx := async.NoContext()
	c := testRequestConnector(t, server)

	conn, st := c.conn(ctx)
	if !st.OK() {
		t.Fatal(st)
	}
	if st := c.close(); !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-conn.Closed().Wait():
	case <-time.After(time.Second):
		t.Fatal("close timeout")
	}
}

func TestRequestConnector_close__should_set_closed_flag(t *testing.T) {
	server := testRequestServer(t)

	c := testRequestConnector(t, server)
	if st := c.close(); !st.OK() {
		t.Fatal(st)
	}

	select {
	case <-c.closed().Wait():
	case <-time.After(time.Second):
		t.Fatal("close timeout")
	}
}
