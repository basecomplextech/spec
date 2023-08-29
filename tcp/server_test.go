package tcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServer_Run__should_start_server(t *testing.T) {
	server := testRequestServer(t)
	assert.NotNil(t, server.main)

	select {
	case <-server.Running():
	case <-time.After(time.Second):
		t.Fatal("server not running")
	}

	select {
	case <-server.Listening():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}
}

func TestServer_Cancel__should_stop_server(t *testing.T) {
	server := testRequestServer(t)
	assert.NotNil(t, server.main)

	select {
	case <-server.Listening():
	case <-time.After(time.Second):
		t.Fatal("server not listening")
	}

	main := server.main
	main.Cancel()
	select {
	case <-main.Wait():
	case <-time.After(time.Second):
		t.Fatal("server not stopped")
	}

	select {
	case <-server.Stopped():
	case <-time.After(time.Second):
		t.Fatal("server not stopped")
	}

	select {
	case <-server.Running():
		t.Fatal("server still running")
	default:
	}

	select {
	case <-server.Listening():
		t.Fatal("server still listening")
	default:
	}
}
