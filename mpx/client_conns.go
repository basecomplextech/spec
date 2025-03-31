// Copyright 2025 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"math/rand/v2"
	"slices"

	"github.com/basecomplextech/baselibrary/collect/slices2"
)

type clientConns struct {
	conns []internalConn
}

func newClientConns() *clientConns {
	return &clientConns{}
}

// len returns the number of connections.
func (c *clientConns) len() int {
	return len(c.conns)
}

// add returns new connections with the given connection added.
func (c *clientConns) add(conn internalConn) *clientConns {
	c1 := c.clone()
	c1.conns = append(c1.conns, conn)
	return c1
}

// remove returns new connections without the given connection.
func (c *clientConns) remove(conn internalConn) *clientConns {
	c1 := c.clone()
	c1.conns = slices2.Remove(c1.conns, conn)
	return c1
}

// roundRoubin returns a random connection from the list of connections.
func (c *clientConns) roundRobin() (internalConn, bool) {
	if len(c.conns) == 0 {
		return nil, false
	}

	// Random start
	i := rand.IntN(len(c.conns))

	// Iterate from i
	for j := i; j < len(c.conns); j++ {
		conn := c.conns[(i+j)%len(c.conns)]
		closed := conn.Closed().IsSet()
		if !closed {
			return conn, true
		}
	}

	// Wrap around, iterate to i
	for j := range i {
		conn := c.conns[(i+j)%len(c.conns)]
		closed := conn.Closed().IsSet()
		if !closed {
			return conn, true
		}
	}

	// No available connections
	return nil, false
}

// private

func (c *clientConns) clone() *clientConns {
	return &clientConns{
		conns: slices.Clone(c.conns),
	}
}
