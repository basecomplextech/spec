package generator

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/model"
)

type clientWriter struct {
	*writer
}

func newClientWriter(w *writer) *clientWriter {
	return &clientWriter{w}
}

func (w *clientWriter) client(def *model.Definition) error {
	if err := w.iface(def); err != nil {
		return err
	}
	if err := w.methods(def); err != nil {
		return err
	}
	if err := w.ifaceEnd(def); err != nil {
		return err
	}
	if err := w.new_client(def); err != nil {
		return err
	}
	if err := w.channels(def); err != nil {
		return err
	}
	return nil
}

// iface

func (w *clientWriter) iface(def *model.Definition) error {
	if def.Service.Sub {
		w.linef(`// %vCall`, def.Name)
		w.line()
		w.linef(`type %vCall interface {`, def.Name)
		w.line()
	} else {
		w.linef(`// %vClient`, def.Name)
		w.line()
		w.linef(`type %vClient interface {`, def.Name)
		w.line()
	}
	return nil
}

// new_client

func (w *clientWriter) new_client(def *model.Definition) error {
	name := clientImplName(def)

	if def.Service.Sub {
		w.linef(`func New%vCall(client rpc.Client, req *rpc.Request) %vCall {`, def.Name, def.Name)
		w.linef(`return new%vCall(client, req)`, def.Name)
		w.line(`}`)
		w.line()
		w.linef(`func New%vCallErr(st status.Status) %vCall {`, def.Name, def.Name)
		w.linef(`return &%v{st: st}`, name)
		w.linef(`}`)
		w.line()
	} else {
		w.linef(`func New%vClient(client rpc.Client) %vClient {`, def.Name, def.Name)
		w.linef(`return &%v{client: client}`, name)
		w.linef(`}`)
		w.line()
	}

	return nil
}

// methods

func (w *clientWriter) methods(def *model.Definition) error {
	for _, m := range def.Service.Methods {
		if err := w.method(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *clientWriter) method(def *model.Definition, m *model.Method) error {
	methodName := toUpperCamelCase(m.Name)
	w.write(methodName)

	if err := w.method_input(def, m); err != nil {
		return err
	}
	if err := w.method_output(def, m); err != nil {
		return err
	}
	return nil
}

func (w *clientWriter) method_input(def *model.Definition, m *model.Method) error {
	ctx := "ctx async.Context, "
	if m.Sub {
		ctx = ""
	}

	switch {
	default:
		w.writef(`(%v) `, ctx)

	case m.Input != nil:
		typeName := typeName(m.Input)
		w.writef(`(%v req_ %v) `, ctx, typeName)
	}
	return nil
}

func (w *clientWriter) method_output(def *model.Definition, m *model.Method) error {
	switch {
	default:
		w.line(`(status.Status)`)

	case m.Sub:
		typeName := typeName(m.Output)
		w.linef(`%vCall`, typeName)

	case m.Chan:
		name := clientChannel_name(m)
		w.linef(`(%v, status.Status)`, name)

	case m.Output != nil:
		typeName := typeName(m.Output)
		w.linef(`(ref.R[%v], status.Status)`, typeName)
	}
	return nil
}

// ifaceEnd

func (w *clientWriter) ifaceEnd(def *model.Definition) error {
	w.linef(`Unwrap() rpc.Client`)
	w.line(`}`)
	w.line()
	return nil
}

// channel

func (w *clientWriter) channels(def *model.Definition) error {
	for _, m := range def.Service.Methods {
		if !m.Chan {
			continue
		}

		if err := w.channel(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *clientWriter) channel(def *model.Definition, m *model.Method) error {
	name := clientChannel_name(m)
	w.linef(`type %v interface {`, name)

	// Receive methods
	if in := m.Channel.In; in != nil {
		typeName := typeName(in)
		w.linef(`Receive(ctx async.Context) (%v, status.Status)`, typeName)
		w.linef(`ReceiveAsync(ctx async.Context) (%v, bool, status.Status)`, typeName)
		w.line(`ReceiveWait() <-chan struct{}`)
	}

	// Send methods
	if out := m.Channel.Out; out != nil {
		typeName := typeName(out)
		w.linef(`Send(ctx async.Context, msg %v) status.Status `, typeName)
		w.line(`SendEnd(ctx async.Context) status.Status `)
	}

	// Response method
	{
		w.write(`Response(ctx async.Context) `)

		switch {
		default:
			w.line(`(status.Status)`)

		case m.Output != nil:
			typeName := typeName(m.Output)
			w.linef(`(%v, status.Status)`, typeName)
		}
	}

	// Free method
	w.line(`Free()`)
	w.line(`}`)
	w.line()
	return nil
}

func clientChannel_name(m *model.Method) string {
	return fmt.Sprintf("%v%vClientChannel", m.Service.Def.Name, toUpperCamelCase(m.Name))
}
