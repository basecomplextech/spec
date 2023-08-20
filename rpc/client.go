package rpc

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Clien is an RPC client.
type Client interface {
	// Free releases the client and its underlying connector.
	Free()

	// Request sends a request and returns a response.
	Request(cancel <-chan struct{}, req *Request) (prpc.Response, status.Status)
}

// NewClient returns a new client with the given config.
func NewClient(config *ClientConfig) (Client, status.Status) {
	return newClient(config)
}

// internal

var _ Client = (*client)(nil)

type client struct {
	client *http.Client
}

func newClient(config *ClientConfig) (*client, status.Status) {
	client_, st := newHttpClient(config)
	if !st.OK() {
		return nil, st
	}

	c := &client{client: client_}
	return c, status.OK
}

func newHttpClient(config *ClientConfig) (*http.Client, status.Status) {
	if config == nil {
		config = DefaultClientConfig()
	}

	// Dialer
	dialer := &net.Dialer{
		Timeout:   config.DialTimeout,
		KeepAlive: config.DialKeepAlive,
	}

	// Proxy
	proxy := http.ProxyFromEnvironment
	if !config.ProxyFromEnv {
		proxy = nil
	}

	// TLS config
	tlsConfig := &tls.Config{}
	{
		mod := false
		if config.TLSRootCert != "" {
			cert, err := os.ReadFile(config.TLSRootCert)
			if err != nil {
				return nil, status.WrapErrorf(err, "Failed to read root certificate")
			}

			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(cert)
			tlsConfig.RootCAs = pool
			mod = true
		}

		if config.TLSInsecureSkip {
			tlsConfig.InsecureSkipVerify = true
			mod = true
		}

		if !mod {
			tlsConfig = nil
		}
	}

	// Transport
	trans := &http.Transport{
		Proxy:                 proxy,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          config.MaxIdleConns,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSClientConfig:       tlsConfig,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
	}

	// Client
	c := &http.Client{Transport: trans}
	return c, status.OK
}

// Free releases the client.
func (c *client) Free() {
	c.client.CloseIdleConnections()
}

// Request sends a request and returns a response.
func (c *client) Request(cancel <-chan struct{}, req *Request) (prpc.Response, status.Status) {
	// Build request
	var req1 *http.Request
	{
		preq, st := req.Build()
		if !st.OK() {
			return prpc.Response{}, st
		}

		buf := bytes.NewBuffer(preq.Unwrap().Raw())
		var err error

		req1, err = http.NewRequest(http.MethodPost, req.url, buf)
		if err != nil {
			return prpc.Response{}, status.WrapError(err)
		}

		req1.Header.Set("Content-Type", ContentType)
	}

	// Send request
	resp1, err := c.client.Do(req1)
	if err != nil {
		return prpc.Response{}, WrapError(err)
	}

	// Check response
	{
		code := resp1.StatusCode
		if (code / 100) != 2 {
			return prpc.Response{}, Errorf("Unexpected response status: %v", resp1.Status)
		}
		ctype := resp1.Header.Get("Content-Type")
		if ctype != ContentType {
			return prpc.Response{}, Errorf("Unexpected response content type: %v", ctype)
		}
	}

	// Read body
	clen := resp1.ContentLength
	if clen < 0 {
		clen = 0
	}
	buf1 := alloc.NewBufferSize(int(clen))

	ok := false
	defer func() {
		if !ok {
			buf1.Free()
		}
	}()

	// TODO: Handle closed and EOF
	b := buf1.Grow(int(clen))
	if _, err := io.ReadFull(resp1.Body, b); err != nil {
		return prpc.Response{}, WrapError(err)
	}

	// Parse response
	presp, _, err := prpc.ParseResponse(buf1.Bytes())
	if err != nil {
		return prpc.Response{}, WrapError(err)
	}

	ok = true
	return presp, status.OK
}
