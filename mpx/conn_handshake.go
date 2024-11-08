// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/pmpx"
)

func (c *conn) handshake() status.Status {
	if c.client {
		return c.handshakeAsClient()
	} else {
		return c.handshakeAsServer()
	}
}

// private

func (c *conn) handshakeAsClient() status.Status {
	// Write protocol line
	if st := c.writer.writeLine(ProtocolLine); !st.OK() {
		return st
	}

	// Write connect request
	req, err := pmpx.NewConnectInput().
		WithCompression(c.options.Compression).
		Build()
	if err != nil {
		return mpxError(err)
	}
	if st := c.writer.writeAndFlush(req); !st.OK() {
		return st
	}

	// Read/check protocol line
	line, st := c.reader.readLine()
	if !st.OK() {
		return st
	}
	if line != ProtocolLine {
		return mpxErrorf("invalid protocol, expected %q, got %q", ProtocolLine, line)
	}

	// Read connect response
	resp, st := c.reader.readResponse()
	if !st.OK() {
		return st
	}
	if ok := resp.Ok(); !ok {
		return mpxErrorf("server refused connection: %v", resp.Error())
	}

	// Check version
	v := resp.Version()
	if v != pmpx.Version_Version10 {
		return mpxErrorf("server returned unsupported version %d", v)
	}

	// Init compression
	comp := resp.Compression()
	switch comp {
	case pmpx.ConnectCompression_None:
	case pmpx.ConnectCompression_Lz4:
		if st := c.reader.initLZ4(); !st.OK() {
			return st
		}
		if st := c.writer.initLZ4(); !st.OK() {
			return st
		}
	default:
		return mpxErrorf("server returned unsupported compression %d", comp)
	}

	c.handshaked.Set()
	return status.OK
}

func (c *conn) handshakeAsServer() status.Status {
	// Write protocol line
	if st := c.writer.writeLine(ProtocolLine); !st.OK() {
		return st
	}

	// Read/check protocol line
	line, st := c.reader.readLine()
	if !st.OK() {
		return st
	}
	if line != ProtocolLine {
		return mpxErrorf("invalid protocol, expected %q, got %q", ProtocolLine, line)
	}

	// Read connect request
	req, st := c.reader.readRequest()
	if !st.OK() {
		return st
	}

	// Check version
	ok := false
	versions := req.Versions()
	for i := 0; i < versions.Len(); i++ {
		v := versions.Get(i)
		if v == pmpx.Version_Version10 {
			ok = true
			break
		}
	}
	if !ok {
		resp, err := pmpx.BuildConnectError("unsupported protocol versions")
		if err != nil {
			return mpxError(err)
		}
		return c.writer.writeAndFlush(resp)
	}

	// Select compression
	comp := pmpx.ConnectCompression_None
	comps := req.Compression()
	for i := 0; i < comps.Len(); i++ {
		c := comps.Get(i)
		if c == pmpx.ConnectCompression_Lz4 {
			comp = pmpx.ConnectCompression_Lz4
			break
		}
	}

	// Write response
	resp, err := pmpx.BuildConnectResponse(pmpx.Version_Version10, comp)
	if err != nil {
		return mpxError(err)
	}
	if st := c.writer.writeAndFlush(resp); !st.OK() {
		return st
	}

	// Init compression
	switch comp {
	case pmpx.ConnectCompression_None:
	case pmpx.ConnectCompression_Lz4:
		if st := c.reader.initLZ4(); !st.OK() {
			return st
		}
		if st := c.writer.initLZ4(); !st.OK() {
			return st
		}
	}

	c.handshaked.Set()
	return status.OK
}
