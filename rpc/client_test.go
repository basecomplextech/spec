package rpc

import (
	"testing"

	"github.com/basecomplextech/spec/proto/prpc"
	"github.com/stretchr/testify/assert"
)

func TestClient_Send__should_send_request_receive_response(t *testing.T) {
	ts := NewTestTransport()
	c := newClient(ts)

	// Push test response
	{
		w := prpc.NewResponseWriter()
		stat := w.Status()
		stat.Code("ok")
		stat.Message("Success")
		if err := stat.End(); err != nil {
			t.Fatal(err)
		}

		results := w.Results()
		arg := results.Add()
		arg.Name("result")
		arg.Value().String("hello, world")
		if err := arg.End(); err != nil {
			t.Fatal(err)
		}
		if err := results.End(); err != nil {
			t.Fatal(err)
		}

		p, err := w.Build()
		if err != nil {
			t.Fatal(err)
		}

		b := p.Unwrap().Raw()
		ts.Push(b)
	}

	// Make test request
	req := NewRequest()
	{
		call := req.Call("echo")
		args := call.Args()
		arg := args.Add()
		arg.Name("arg")
		arg.Value().String("hello, world")
		if err := args.End(); err != nil {
			t.Fatal(err)
		}
		if err := args.End(); err != nil {
			t.Fatal(err)
		}
		if err := call.End(); err != nil {
			t.Fatal(err)
		}
	}

	// Send request, receive response
	resp, st := c.Request(nil, req)
	if !st.OK() {
		t.Fatal(st)
	}

	stat := resp.Status()
	result := resp.Results().Get(0)
	assert.Equal(t, "ok", stat.Code().Unwrap())
	assert.Equal(t, "Success", stat.Message().Unwrap())
	assert.Equal(t, "result", result.Name().Unwrap())
	assert.Equal(t, "hello, world", result.Value().String().Unwrap())
}
