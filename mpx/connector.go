// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/status"
)

type connector interface {
	// closed returns a flag which indicates the connector is closed.
	closed() async.Flag

	// connected returns a flag which indicates there is at least one connected connection.
	connected() async.Flag

	// disconnected returns a flag which indicates there are no connected connections.
	disconnected() async.Flag

	// methods

	// connect returns a connection or a future.
	conn(ctx async.Context) (*conn, async.Future[*conn], status.Status)

	// close stops and closes the connector.
	close() status.Status
}
