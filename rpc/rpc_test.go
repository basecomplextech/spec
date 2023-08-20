package rpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPC(t *testing.T) {
	logger := logging.TestLogger(t)

	serverConfig := DefaultServerConfig()
	serverConfig.Listen = "localhost:0"
	serverConfig.CertPath = "../internal/certs/localhost.crt"
	serverConfig.KeyPath = "../internal/certs/localhost.key"

	handler := HandlerFunc(func(cancel <-chan struct{}, req *ServerRequest, resp ServerResponse) status.Status {
		call, st := req.Call(0)
		if !st.OK() {
			return st
		}

		arg := call.Args().Get(0)
		name := arg.Name().Unwrap()
		value := arg.Value().String().Unwrap()
		if name != "msg" || value != "hello, world" {
			return status.Error("invalid argument")
		}

		{
			res := resp.Result()
			res.Name("a")
			res.Value().Int64(123)
			if err := res.End(); err != nil {
				return status.WrapError(err)
			}
		}
		{
			res := resp.Result()
			res.Name("b")
			res.Value().String("hello, world")
			if err := res.End(); err != nil {
				return status.WrapError(err)
			}
		}
		return status.OK
	})

	// Make server
	s, st := newServer(serverConfig, logger, map[string]Handler{"": handler})
	if !st.OK() {
		t.Fatal(st)
	}

	// Make client
	clientConfig := DefaultClientConfig()
	clientConfig.TLSRootCert = "../internal/certs/localhost.crt"
	c, st := newClient(clientConfig)
	if !st.OK() {
		t.Fatal(st)
	}

	// Run server
	running, st := s.Run()
	if !st.OK() {
		t.Fatal(st)
	}
	defer async.CancelWait(running)

	// Await listening
	select {
	case <-s.Listening():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}

	// Make request
	url := fmt.Sprintf("https://%v", s.address)
	req := NewRequest(url)
	{
		call := req.Call("test")
		args := call.Args()
		arg := args.Add()
		arg.Name("msg")
		arg.Value().String("hello, world")
		if err := arg.End(); err != nil {
			t.Fatal(err)
		}
		if err := args.End(); err != nil {
			t.Fatal(err)
		}
		if err := call.End(); err != nil {
			t.Fatal(err)
		}
	}

	// Send request
	resp, st := c.Request(nil, req)
	if !st.OK() {
		t.Fatal(st)
	}

	// Parse response
	results := resp.Results()
	require.Equal(t, 2, results.Len())

	a := results.Get(0).Value().Int64()
	b := results.Get(1).Value().String().Unwrap()
	assert.Equal(t, a, int64(123))
	assert.Equal(t, b, "hello, world")
}
