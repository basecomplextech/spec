package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/async"
	"github.com/basecomplextech/baselibrary/logging"
	"github.com/basecomplextech/baselibrary/status"
	"github.com/basecomplextech/spec/proto/prpc"
)

// Server

// Server is an RPC server backed by an HTTP2 server.
type Server interface {
	// Running indicates that the server is running.
	Running() <-chan struct{}

	// Stopped indicates that the server is stopped.
	Stopped() <-chan struct{}

	// Listening indicates that the server is listening to an address.
	Listening() <-chan struct{}

	// Run runs the server.
	Run() (async.Routine[struct{}], status.Status)
}

// NewServer returns a new server with the given config and a single root handler.
func NewServer(config *ServerConfig, logger logging.Logger, handler Handler) (Server, status.Status) {
	h := map[string]Handler{"": handler}
	return newServer(config, logger, h)
}

// NewServerHandlers returns a new server with the given config and handlers.
func NewServerHandlers(config *ServerConfig, logger logging.Logger, handlers map[string]Handler) (Server, status.Status) {
	return newServer(config, logger, handlers)
}

// internal

var _ Server = (*server)(nil)

type server struct {
	config   *ServerConfig
	logger   logging.Logger
	handlers map[string]Handler

	running   *async.Flag
	stopped   *async.Flag
	listening *async.Flag

	mu      sync.Mutex
	main    async.Routine[struct{}]
	address string
}

func newServer(config *ServerConfig, logger logging.Logger, handlers map[string]Handler) (*server, status.Status) {
	if len(handlers) == 0 {
		return nil, Error("No server handlers")
	}

	s := &server{
		config: config,
		logger: logger,

		running:   async.UnsetFlag(),
		stopped:   async.SetFlag(),
		listening: async.UnsetFlag(),

		handlers: make(map[string]Handler),
	}

	for path, h := range handlers {
		if path == "" {
			path = "/"
		}

		if _, ok := s.handlers[path]; ok {
			return nil, Errorf("Duplicate server handler for path %q", path)
		}

		s.handlers[path] = h
	}
	return s, status.OK
}

// Running indicates that the server is running.
func (s *server) Running() <-chan struct{} {
	return s.running.Wait()
}

// Stopped indicates that the server is stopped.
func (s *server) Stopped() <-chan struct{} {
	return s.stopped.Wait()
}

// Listening indicates that the server is listening to an address.
func (s *server) Listening() <-chan struct{} {
	return s.listening.Wait()
}

// Run runs the server.
func (s *server) Run() (async.Routine[struct{}], status.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.main != nil {
		return nil, status.Error("Server already running")
	}

	s.main = async.Go(s.run)
	return s.main, status.OK
}

// private

func (s *server) checkConfig() status.Status {
	switch {
	case s.config.CertPath == "":
		return status.Error("no server certificate")
	case s.config.KeyPath == "":
		return status.Error("no server key")
	}

	if _, err := os.Stat(s.config.CertPath); err != nil {
		return status.WrapErrorf(err, "server certificate error")
	}
	if _, err := os.Stat(s.config.KeyPath); err != nil {
		return status.WrapErrorf(err, "server key error")
	}
	return status.OK
}

func (s *server) run(cancel <-chan struct{}) status.Status {
	s.stopped.Unset()
	s.running.Set()
	defer s.stopped.Set()
	defer s.running.Unset()
	defer s.listening.Unset()

	// Make server
	srv := &http.Server{
		Addr:    s.config.Listen,
		Handler: http.HandlerFunc(s.handle),
	}
	defer s.logger.Debug("Server stopped")
	defer srv.Close()

	// Check config
	if st := s.checkConfig(); st != status.OK {
		s.logger.Error("Invalid server config", "status", st)
		return st
	}

	// Listen to address
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Error("Failed to listen to address", "address", addr, "status", err)
		return status.WrapError(err)
	}
	ln = &loggingListener{ln}
	defer ln.Close()

	s.address = ln.Addr().String()
	s.listening.Set()
	s.logger.Info("Listening", "address", addr)

	// Run server
	serving := async.Go(func(cancel <-chan struct{}) status.Status {
		err := srv.ServeTLS(ln, s.config.CertPath, s.config.KeyPath)
		if err != nil {
			return status.WrapError(err)
		}
		return status.OK
	})

	// Wait for cancel or error
	var st status.Status
	select {
	case <-cancel:
		ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()
		s.logger.Debug("Server is shutting down...")

		if err := srv.Shutdown(ctx); err != nil {
			s.logger.Error("Server failed to shutdown gracefully", "status", err)
		}

		st = status.Cancelled

	case <-serving.Wait():
		st = serving.Status()
		s.logger.Error("Server error", "status", st)
	}

	return st
}

func (s *server) handle(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			st, stack := status.RecoverStack(e)
			s.logger.Error("Server panic", "status", st, "stack", string(stack))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()

	// Check method
	method := req.Method
	if method != "POST" {
		msg := fmt.Sprintf("Method %v is not allowed", method)
		http.Error(w, msg, http.StatusMethodNotAllowed)
		return
	}

	// Check content type
	ctype := req.Header.Get("Content-Type")
	if ctype != ContentType {
		msg := fmt.Sprintf("Content type %q is not allowed", ctype)
		http.Error(w, msg, http.StatusUnsupportedMediaType)
		return
	}

	// Get handler
	path := req.URL.Path
	handler, ok := s.handlers[path]
	if !ok {
		msg := fmt.Sprintf("RPC service at %q is not found", path)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	// Parse request
	req1, st := newServerRequest(req)
	if !st.OK() {
		msg := fmt.Sprintf("Failed to parse request: %v", st)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	defer req1.free()

	// Make response
	resp1 := newServerResponse()
	defer resp1.free()

	// Handle request
	ctx := req.Context()
	cancel := ctx.Done()

	if st := handler.Handle(cancel, req1, resp1); !st.OK() {
		s.rpcError(w, st)
		return
	}

	// Write response
	p, st := resp1.build()
	if !st.OK() {
		s.rpcError(w, st)
		return
	}
	s.rpcResponse(w, p)
}

func (s *server) rpcError(w http.ResponseWriter, st status.Status) {
	buf := alloc.NewBuffer()
	defer buf.Free()

	resp, st := newErrorResponse(buf, st)
	if !st.OK() {
		s.logger.Error("Failed to build error response", "status", st)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	s.rpcResponse(w, resp)
}

func (s *server) rpcResponse(w http.ResponseWriter, resp prpc.Response) {
	data := resp.Unwrap().Raw()
	clen := len(data)

	w.Header().Set("Content-Type", "application/spec-rpc")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", clen))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func newErrorResponse(buf *alloc.Buffer, st status.Status) (prpc.Response, status.Status) {
	w := prpc.NewResponseWriterBuffer(buf)

	w1 := w.Status()
	w1.Code(string(st.Code))
	w1.Message(st.Message)
	if err := w1.End(); err != nil {
		return prpc.Response{}, status.WrapError(err)
	}

	p, err := w.Build()
	if err != nil {
		return prpc.Response{}, status.WrapError(err)
	}
	return p, status.OK
}

type loggingListener struct {
	net.Listener
}

// Accept waits for and returns the next connection to the listener.
func (l *loggingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	fmt.Println("Accepted connection", conn.RemoteAddr().String())
	return conn, nil
}
