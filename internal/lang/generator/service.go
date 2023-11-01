package generator

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/model"
)

type serviceWriter struct {
	*writer
}

func newServiceWriter(w *writer) *serviceWriter {
	return &serviceWriter{w}
}

func (w *serviceWriter) service(def *model.Definition) error {
	if err := w.iface(def); err != nil {
		return err
	}
	if err := w.channels(def); err != nil {
		return err
	}
	return nil
}

func (w *serviceWriter) iface(def *model.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef(`type %v interface {`, def.Name)

	for _, m := range def.Service.Methods {
		if err := w.method(def, m); err != nil {
			return err
		}
	}

	w.linef(`}`)
	w.line()
	return nil
}

func (w *serviceWriter) method(def *model.Definition, m *model.Method) error {
	if err := w.method_input(def, m); err != nil {
		return err
	}
	if err := w.method_output(def, m); err != nil {
		return err
	}
	w.line()
	return nil
}

func (w *serviceWriter) method_input(def *model.Definition, m *model.Method) error {
	name := toUpperCamelCase(m.Name)
	w.writef(`%v`, name)

	switch {
	default:
		w.write(`(cancel <-chan struct{}) `)

	case m.Chan:
		channel := serviceChannel_name(m)
		w.writef(`(cancel <-chan struct{}, ch %v) `, channel)

	case m.Input != nil:
		typeName := typeName(m.Input)
		w.writef(`(cancel <-chan struct{}, req %v) `, typeName)

	case m.InputFields != nil:
		w.writef(`(cancel <-chan struct{}, `)

		fields := m.InputFields.List
		multi := len(fields) > 3
		if multi {
			w.line()
		}

		for _, field := range fields {
			argName := toLowerCameCase(field.Name)
			typeName := typeRefName(field.Type)

			if multi {
				w.linef(`%v_ %v, `, argName, typeName)
			} else {
				w.writef(`%v_ %v, `, argName, typeName)
			}
		}

		w.write(`) `)
	}
	return nil
}

func (w *serviceWriter) method_output(def *model.Definition, m *model.Method) error {
	out := m.Output

	switch {
	default:
		w.write(`(status.Status)`)

	case m.Sub:
		typeName := typeName(out)
		w.writef(`(%v, status.Status)`, typeName)

	case m.Output != nil:
		typeName := typeName(out)
		w.writef(`(*ref.R[%v], status.Status)`, typeName)

	case m.OutputFields != nil:
		fields := m.OutputFields.List
		multi := len(fields) > 1
		w.line(`(`)

		for _, field := range fields {
			name := toLowerCameCase(field.Name)
			typeName := typeName(field.Type)

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

		w.write(`)`)
	}
	return nil
}

// channels

func (w *serviceWriter) channels(def *model.Definition) error {
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

func (w *serviceWriter) channel(def *model.Definition, m *model.Method) error {
	name := serviceChannel_name(m)
	w.linef(`type %v interface {`, name)

	// Request method
	switch {
	case m.Input != nil:
		typeName := typeName(m.Input)
		w.linef(`Request() (%v, status.Status)`, typeName)

	case m.InputFields != nil:
		w.write(`Request() (`)

		fields := m.InputFields.List
		for _, f := range fields {
			w.writef(`%v, `, typeName(f.Type))
		}
		w.line(`status.Status)`)
	}

	// Send method
	if in := m.Channel.In; in != nil {
		typeName := typeName(in)
		w.linef(`Send(cancel <-chan struct{}, msg %v) status.Status`, typeName)
	}

	// Receive method
	if out := m.Channel.Out; out != nil {
		typeName := typeName(out)
		w.linef(`Receive(cancel <-chan struct{}) (%v, status.Status)`, typeName)
	}

	w.linef(`}`)
	w.line()
	return nil
}

func serviceChannel_name(m *model.Method) string {
	return fmt.Sprintf("%v%vChannel", m.Service.Def.Name, toUpperCamelCase(m.Name))
}
