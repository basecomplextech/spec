package golang

import "github.com/basecomplextech/spec/lang/compiler"

func (w *writer) service(def *compiler.Definition) error {
	if err := w.serviceDef(def); err != nil {
		return err
	}
	if err := w.serviceMethods(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) serviceDef(def *compiler.Definition) error {
	w.linef(`// %v`, def.Name)
	w.line()
	w.linef(`type %v struct {`, def.Name)
	w.line(`client rpc.Client`)
	w.line(`}`)
	w.line()
	w.linef(`func New%v(client rpc.Client) *%v {`, def.Name, def.Name)
	w.linef(`return &%v{client}`, def.Name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) serviceMethods(def *compiler.Definition) error {
	for _, m := range def.Service.Methods {
		if err := w.method(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) method(def *compiler.Definition, m *compiler.Method) error {
	methodName := toUpperCamelCase(m.Name)
	w.writef(`func (_c %v) %v`, def.Name, methodName)

	if err := w.methodArgs(def, m); err != nil {
		return err
	}
	if err := w.methodResults(def, m); err != nil {
		return err
	}
	if err := w.methodBody(def, m); err != nil {
		return err
	}

	w.line()
	return nil
}

func (w *writer) methodArgs(def *compiler.Definition, m *compiler.Method) error {
	w.write(`(`)
	w.write(`cancel <-chan struct{}`)

	for _, arg := range m.Args {
		argName := toLowerCameCase(arg.Name)
		typeName := typeName(arg.Type)
		w.write(`, `)
		w.writef(`%v %v`, argName, typeName)
	}

	w.write(`) `)
	return nil
}

func (w *writer) methodResults(def *compiler.Definition, m *compiler.Method) error {
	if m.Chained {
		res := m.Results[0]
		resName := toLowerCameCase(res.Name)
		serviceName := res.Type.Name
		w.write(`(`)
		w.writef(`%v %v`, resName, serviceName)
		w.write(`)`)
		return nil
	}

	w.write(`(`)
	for _, res := range m.Results {
		resName := toLowerCameCase(res.Name)
		typeName := typeRefName(res.Type)
		w.writef(`%v %v, `, resName, typeName)
	}
	w.writef(`st status.Status`)
	w.write(`) `)
	return nil
}

func (w *writer) methodBody(def *compiler.Definition, m *compiler.Method) error {
	w.line(`{`)

	// Make request
	w.line(`// Make request`)
	w.line(`_req := rpc.NewRequest()`)
	w.line(`defer _req.Free()`)
	w.line()

	// Make call
	w.line(`// Make call`)
	w.linef(`_call := _req.Call("%v")`, m.Name)
	w.line(`{`)
	w.line(`_args := _call.Args()`)

	// Make args
	for _, arg := range m.Args {
		argName := toLowerCameCase(arg.Name)

		w.line(`{`)
		w.line(`_arg := _args.Add()`)
		w.linef(`_arg.Name(%v)`, arg.Name)
		w.linef(`_arg.Value(%v)`, argName)

		w.line(`if err := _arg.End(); err != nil {`)
		w.line(`return status.WrapError(err)`)
		w.line(`}`)
		w.line(`}`)
	}

	// End args
	w.line(`if err := _args.End(); err != nil {`)
	w.line(`return status.WrapError(err)`)
	w.line(`}`)

	// End call
	w.line(`if err := _call.End(); err != nil {`)
	w.line(`return status.WrapError(err)`)
	w.line(`}`)
	w.line(`}`)
	w.line()

	// Send request
	w.line(`// Send request`)
	w.line(`_resp, st := _c.client.Request(cancel, _req)`)
	w.line(`if !st.OK() {`)
	w.line(`return st`)
	w.line(`}`)
	w.line(``)

	// Parse results
	w.line(`// Parse results`)
	w.line(`_results := _resp.Results()`)
	w.line(`for i := 0; i < _results.Len(); i++ {`)
	w.line(`_result := _results.Get(i)`)
	w.line(`_name := _result.Name().Unwrap()`)
	w.line()
	w.line(`switch _name {`)
	for _, res := range m.Results {
		resName := toLowerCameCase(res.Name)
		w.linef(`case "%v":`, res.Name)
		w.linef(`%v = _result.Value().String()`, resName)
	}
	w.line(`}`)
	w.line(`}`)
	w.line()

	// Done
	w.line(`// Done`)
	w.line(`st = rpc.ParseStatus(_resp.Status())`)
	w.line(`return msg1, st`)
	w.line(`}`)
	return nil
}
