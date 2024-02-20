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
	w.linef(`// %vClient`, def.Name)
	w.line()
	w.linef(`type %vClient interface {`, def.Name)
	w.line()
	return nil
}

// new_client

func (w *clientWriter) new_client(def *model.Definition) error {
	name := clientImplName(def)

	if def.Service.Sub {
		w.linef(`func New%vClient(client rpc.Client, req *rpc.Request) %vClient {`, def.Name, def.Name)
		w.linef(`return &%v{`, name)
		w.linef(`client: client,`)
		w.linef(`req: req,`)
		w.linef(`st: status.OK,`)
		w.linef(`}`)
		w.linef(`}`)
		w.line()
		w.linef(`func New%vClientErr(st status.Status) %vClient {`, def.Name, def.Name)
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
	cancel := "cancel <-chan struct{}, "
	if m.Sub {
		cancel = ""
	}

	switch {
	default:
		w.writef(`(%v) `, cancel)

	case m.Input != nil:
		typeName := typeName(m.Input)
		w.writef(`(%v req_ %v) `, cancel, typeName)

	case m.InputFields != nil:
		w.writef(`(%v`, cancel)

		fields := m.InputFields.List
		multi := len(fields) > 3
		if multi {
			w.line()
		}

		for _, field := range fields {
			argName := toLowerCameCase(field.Name)
			typeName := typeName(field.Type)

			if multi {
				w.linef(`%v_ %v, `, argName, typeName)
			} else {
				w.writef(`%v_ %v, `, argName, typeName)
			}
		}

		w.write(`)`)
	}
	return nil
}

func (w *clientWriter) method_output(def *model.Definition, m *model.Method) error {
	switch {
	default:
		w.line(`(status.Status)`)

	case m.Sub:
		typeName := typeName(m.Output)
		w.linef(`%vClient`, typeName)

	case m.Chan:
		name := clientChannel_name(m)
		w.linef(`(%v, status.Status)`, name)

	case m.Output != nil:
		typeName := typeName(m.Output)
		w.linef(`(*ref.R[%v], status.Status)`, typeName)

	case m.OutputFields != nil:
		fields := m.OutputFields.List
		multi := len(fields) > 3

		if multi {
			w.line(`(`)
		} else {
			w.write(`(`)
		}

		for _, f := range fields {
			name := toLowerCameCase(f.Name)
			typeName := typeName(f.Type)

			if multi {
				w.linef(`_%v %v, `, name, typeName)
			} else {
				w.writef(`_%v %v, `, name, typeName)
			}
		}

		if multi {
			w.line(`_st status.Status,`)
		} else {
			w.write(`_st status.Status`)
		}

		w.line(`)`)
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

	// Read methods
	if in := m.Channel.In; in != nil {
		typeName := typeName(in)
		w.linef(`Read(cancel <-chan struct{}) (%v, bool, status.Status)`, typeName)
		w.linef(`ReadSync(cancel <-chan struct{}) (%v, status.Status)`, typeName)
		w.line(`ReadWait() <-chan struct{}`)
	}

	// Write methods
	if out := m.Channel.Out; out != nil {
		typeName := typeName(out)
		w.linef(`Write(cancel <-chan struct{}, msg %v) status.Status `, typeName)
		w.line(`WriteEnd(cancel <-chan struct{}) status.Status `)
	}

	// Response method
	{
		w.write(`Response(cancel <-chan struct{}) `)

		switch {
		default:
			w.line(`(status.Status)`)

		case m.Output != nil:
			typeName := typeName(m.Output)
			w.linef(`(*ref.R[%v], status.Status)`, typeName)

		case m.OutputFields != nil:
			fields := m.OutputFields.List
			multi := len(fields) > 3

			if multi {
				w.line(`(`)
			} else {
				w.write(`(`)
			}

			for _, f := range fields {
				name := toLowerCameCase(f.Name)
				typeName := typeName(f.Type)
				if multi {
					w.linef(`_%v %v, `, name, typeName)
				} else {
					w.writef(`_%v %v, `, name, typeName)
				}
			}

			if multi {
				w.line(`_st status.Status,`)
			} else {
				w.write(`_st status.Status`)
			}

			w.line(`)`)
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
