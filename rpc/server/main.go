package main

import (
	"io"
	"log"
	"net/http"
)

const (
	certPath    = "../../../certs/localhost.crt"
	certKeyPath = "../../../certs/localhost.key"
)

func main() {
	// Create a server on port 7777
	// Exactly how you would run an HTTP/1.1 server
	srv := &http.Server{Addr: ":7777", Handler: http.HandlerFunc(handler)}

	// Start the server with TLS, since we are running HTTP/2 it must be
	// run with TLS.
	// Exactly how you would run an HTTP/1.1 server with TLS connection.
	log.Printf("Serving on https://0.0.0.0:7777")
	log.Fatal(srv.ListenAndServeTLS(certPath, certKeyPath))
}

type flushWriter struct {
	w io.Writer
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	// Flush - send the buffered written data to the client
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Log the request protocol
	log.Printf("Got connection: %s", r.Proto)

	// First flash response headers
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	// Copy from the request body to the response writer and flush
	// (send to client)
	io.Copy(flushWriter{w: w}, r.Body)
}
