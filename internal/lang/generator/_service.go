package generator

import "github.com/basecomplextech/spec/internal/lang/model"

func (w *writer) service(def *model.Definition) error {
	if err := w.iface(def); err != nil {
		return err
	}
	if err := w.client(def); err != nil {
		return err
	}
	if err := w.handler(def); err != nil {
		return err
	}
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
	if err := w.ifaceMethod_args(def, m); err != nil {
		return err
	}
	if err := w.ifaceMethod_results(def, m); err != nil {
		return err
	}
	w.line()
	return nil
}

func (w *writer) ifaceMethod_args(def *model.Definition, m *model.Method) error {
	methodName := toUpperCamelCase(m.Name)

	w.writef(`%v`, methodName)
	w.write(`(cancel <-chan struct{}`)

	for _, arg := range m.Args {
		if len(m.Args) <= 3 {
			w.write(`, `)
		} else {
			w.line(`,`)
		}

		argName := toLowerCameCase(arg.Name)
		typeName := typeRefName(arg.Type)
		w.writef(`%v %v`, argName, typeName)
	}
	if len(m.Args) > 3 {
		w.line(`,`)
	}

	w.write(`) `)
	return nil
}

func (w *writer) ifaceMethod_results(def *model.Definition, m *model.Method) error {
	if len(m.Results) > 1 {
		w.linef(`(`)
	} else {
		w.write(`(`)
	}

	for _, res := range m.Results {
		resName := toLowerCameCase(res.Name)
		typeName := typeName(res.Type)

		if len(m.Results) > 1 {
			w.linef(`%v_ %v,`, resName, typeName)
		} else {
			w.writef(`%v_ %v, `, resName, typeName)
		}
	}

	if len(m.Results) > 1 {
		w.line(`st_ status.Status,`)
	} else {
		w.write(`st_ status.Status`)
	}
	w.line(`)`)
	return nil
}
