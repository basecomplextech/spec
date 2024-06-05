package mpx

import (
	"testing"
	"time"
)

func TestServer_Run__should_start_server(t *testing.T) {
	server := testRequestServer(t)

	select {
	case <-server.Running().Wait():
	case <-time.After(time.Second):
		t.Fatal("server not running")
	}

	select {
	case <-server.Listening().Wait():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}
}

func TestServer_Cancel__should_stop_server(t *testing.T) {
	server := testRequestServer(t)

	select {
	case <-server.Listening().Wait():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}

	select {
	case <-server.Stop():
	case <-time.After(time.Second):
		t.Fatal("server not stopped")
	}

	select {
	case <-server.Stopped().Wait():
	case <-time.After(time.Second):
		t.Fatal("server not stopped")
	}

	select {
	case <-server.Running().Wait():
		t.Fatal("server still running")
	default:
	}

	select {
	case <-server.Listening().Wait():
		t.Fatal("server still listening")
	default:
	}
}
