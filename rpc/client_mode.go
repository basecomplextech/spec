// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package rpc

import "github.com/basecomplextech/spec/mpx"

// ClientMode specifies how the client connects to the server.
type ClientMode = mpx.ClientMode

const (
	// ClientMode_OnDemand connects to the server on demand, does not reconnect on errors.
	ClientMode_OnDemand = mpx.ClientMode_OnDemand

	// ClientMode_AutoConnect automatically connects and reconnects to the server.
	// The client reconnects with exponential backoff on errors.
	ClientMode_AutoConnect = mpx.ClientMode_AutoConnect
)
