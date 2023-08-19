package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	url        = "https://localhost:7777"
	caCertPath = "../../../certs/localhost.crt"
)

func BenchmarkClient(b *testing.B) {
	// Create a pool with the server certificate since it is not signed
	// by a known CA
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		log.Fatalf("Reading server certificate: %s", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS configuration with the certificate of the server
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsConfig,
			ForceAttemptHTTP2: true,
		},
	}

	data := bytes.Repeat([]byte("a"), 16*1024)

	// Create a pipe - an object that implements `io.Reader` and `io.Writer`.
	// Whatever is written to the writer part will be read by the reader part.
	pr, pw := io.Pipe()

	// Create an `http.Request` and set its body as the reader part of the
	// pipe - after sending the request, whatever will be written to the pipe,
	// will be sent as the request body.
	// This makes the request content dynamic, so we don't need to define it
	// before sending the request.
	req, err := http.NewRequest(http.MethodPost, url, io.NopCloser(pr))
	if err != nil {
		log.Fatal(err)
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Got: %d", resp.StatusCode)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	t0 := time.Now()

	// Discard server response
	go func() {
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			b.Fatal(err)
		}
	}()

	for i := 0; i < b.N; i++ {
		_, err := pw.Write(data)
		if err != nil {
			b.Fatal(err)
		}
	}

	t1 := time.Now()
	sec := t1.Sub(t0).Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops, "ops")
}
