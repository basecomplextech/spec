package generator

import "github.com/basecomplextech/spec/internal/lang/model"

func (w *writer) service(def *model.Definition) error {
	if err := w.iface(def); err != nil {
		return err
	}
	if err := w.client(def); err != nil {
		return err
	}
	// if err := w.handler(def); err != nil {
	// 	return err
	// }
	return nil
}

func (w *writer) iface(def *model.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef(`type %v interface {`, def.Name)

	for _, m := range def.Service.Methods {
		if err := w.ifaceMethod(def, m); err != nil {
			return err
		}
	}

	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) ifaceMethod(def *model.Definition, m *model.Method) error {
	if err := w.ifaceMethod_input(def, m); err != nil {
		return err
	}
	if err := w.ifaceMethod_output(def, m); err != nil {
		return err
	}
	w.line()
	return nil
}

func (w *writer) ifaceMethod_input(def *model.Definition, m *model.Method) error {
	name := toUpperCamelCase(m.Name)
	w.writef(`%v`, name)

	switch {
	default:
		w.write(`(cancel <-chan struct{}) `)

	case m.Input != nil:
		typeName := typeName(m.Input)
		w.writef(`(cancel <-chan struct{}, req %v) `, typeName)

	case m.InputFields != nil:
		w.write(`(cancel <-chan struct{}, `)

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

func (w *writer) ifaceMethod_output(def *model.Definition, m *model.Method) error {
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

func ifaceChannel_name(m *model.Method) string {
	return ""
}
