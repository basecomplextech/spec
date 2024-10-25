// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

type connPool map[*conn]struct{}

func newConnPool() connPool {
	return make(map[*conn]struct{})
}

// conn returns a connection with the minimum number of channels, or false.
// the method also removes closed connections.
func (p connPool) conn() (*conn, bool) {
	var conn *conn

	for conn1 := range p {
		if conn1.closed.Get() {
			delete(p, conn1)
			continue
		}

		if conn == nil {
			conn = conn1
			continue
		}

		n0 := conn.channelNum()
		n1 := conn1.channelNum()
		if n1 < n0 {
			conn = conn1
		}
	}

	ok := conn != nil
	return conn, ok
}
