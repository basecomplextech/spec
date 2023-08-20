package rpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/stretchr/testify/require"
)

func BenchmarkRPC(b *testing.B) {
	logger := logging.TestLogger(b)

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
		b.Fatal(st)
	}

	// Make client
	clientConfig := DefaultClientConfig()
	clientConfig.TLSRootCert = "../internal/certs/localhost.crt"
	c, st := newClient(clientConfig)
	if !st.OK() {
		b.Fatal(st)
	}

	// Run server
	_, st = s.Run()
	if !st.OK() {
		b.Fatal(st)
	}

	// Await listening
	select {
	case <-s.Listening():
	case <-time.After(time.Second):
		b.Fatal("server not listening")
	}

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	for i := 0; i < b.N; i++ {
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
				b.Fatal(err)
			}
			if err := args.End(); err != nil {
				b.Fatal(err)
			}
			if err := call.End(); err != nil {
				b.Fatal(err)
			}
		}

		// Send request
		resp, st := c.Request(nil, req)
		if !st.OK() {
			b.Fatal(st)
		}

		// Parse response
		results := resp.Results()
		require.Equal(b, 2, results.Len())

		req.Free()
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}

func BenchmarkRPC_Parallel(b *testing.B) {
	logger := logging.TestLogger(b)

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
		b.Fatal(st)
	}

	// Make client
	clientConfig := DefaultClientConfig()
	clientConfig.TLSRootCert = "../internal/certs/localhost.crt"
	c, st := newClient(clientConfig)
	if !st.OK() {
		b.Fatal(st)
	}

	// Run server
	_, st = s.Run()
	if !st.OK() {
		b.Fatal(st)
	}

	// Await listening
	select {
	case <-s.Listening():
	case <-time.After(time.Second):
		b.Fatal("server not listening")
	}

	b.ReportAllocs()
	b.ResetTimer()
	t0 := time.Now()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
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
					b.Fatal(err)
				}
				if err := args.End(); err != nil {
					b.Fatal(err)
				}
				if err := call.End(); err != nil {
					b.Fatal(err)
				}
			}

			// Send request
			resp, st := c.Request(nil, req)
			if !st.OK() {
				b.Fatal(st)
			}

			// Parse response
			results := resp.Results()
			require.Equal(b, 2, results.Len())

			req.Free()
		}
	})

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}
