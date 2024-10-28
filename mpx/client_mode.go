// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

// ClientMode specifies how the client connects to the server.
type ClientMode int

const (
	// ClientMode_OnDemand connects to the server on demand, does not reconnect on errors.
	ClientMode_OnDemand ClientMode = iota

	// ClientMode_AutoConnect automatically connects and reconnects to the server.
	// The client reconnects with exponential backoff on errors.
	ClientMode_AutoConnect
)
