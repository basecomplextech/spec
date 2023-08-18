package golang

import (
	"fmt"

	"github.com/basecomplextech/spec/lang/compiler"
)

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
	if err := w.methodCall(def, m); err != nil {
		return err
	}
	if err := w.methodRequest(def, m); err != nil {
		return err
	}

	w.line()
	return nil
}

func (w *writer) methodArgs(def *compiler.Definition, m *compiler.Method) error {
	w.write(`(cancel <-chan struct{}`)

	for _, arg := range m.Args {
		if len(m.Args) <= 3 {
			w.write(`, `)
		} else {
			w.line(`,`)
		}

		argName := toLowerCameCase(arg.Name)
		typeName := typeName(arg.Type)
		w.writef(`%v %v`, argName, typeName)
	}
	if len(m.Args) > 3 {
		w.line(`,`)
	}

	w.write(`) `)
	return nil
}

func (w *writer) methodResults(def *compiler.Definition, m *compiler.Method) error {
	// TODO: Check number of results in compiler, zero results
	switch {
	case m.Chained:
		res := m.Results[0]
		serviceName := res.Type.Name
		w.writef(`%v`, serviceName)

	case len(m.Results) == 1:
		result := m.Results[0]
		typeName := typeRefName(result.Type)
		w.writef(`(rpc.Future[%v], status.Status)`, typeName)

	default:
		w.linef(`(rpc.Future%d[`, len(m.Results))
		for _, res := range m.Results {
			typeName := typeRefName(res.Type)
			w.linef(`%v,`, typeName)
		}
		w.writef(`], status.Status)`)
	}
	return nil
}

func (w *writer) methodCall(def *compiler.Definition, m *compiler.Method) error {
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
		kind := arg.Type.Kind
		name := toLowerCameCase(arg.Name)

		w.line(`{`)
		w.line(`_arg := _args.Add()`)
		w.linef(`_arg.Name("%v")`, arg.Name)

		switch kind {
		case compiler.KindBool:
			w.linef(`_arg.Value().Bool(%v)`, name)
		case compiler.KindByte:
			w.linef(`_arg.Value().Byte(%v)`, name)

		case compiler.KindInt16:
			w.linef(`_arg.Value().Int16(%v)`, name)
		case compiler.KindInt32:
			w.linef(`_arg.Value().Int32(%v)`, name)
		case compiler.KindInt64:
			w.linef(`_arg.Value().Int64(%v)`, name)

		case compiler.KindUint16:
			w.linef(`_arg.Value().Uint16(%v)`, name)
		case compiler.KindUint32:
			w.linef(`_arg.Value().Uint32(%v)`, name)
		case compiler.KindUint64:
			w.linef(`_arg.Value().Uint64(%v)`, name)

		case compiler.KindBin64:
			w.linef(`_arg.Value().Bin64(%v)`, name)
		case compiler.KindBin128:
			w.linef(`_arg.Value().Bin128(%v)`, name)
		case compiler.KindBin256:
			w.linef(`_arg.Value().Bin256(%v)`, name)

		case compiler.KindFloat32:
			w.linef(`_arg.Value().Float32(%v)`, name)
		case compiler.KindFloat64:
			w.linef(`_arg.Value().Float64(%v)`, name)

		case compiler.KindBytes:
			w.linef(`_arg.Value().Bytes(%v)`, name)
		case compiler.KindString:
			w.linef(`_arg.Value().String(%v)`, name)

		case compiler.KindEnum:
			writeFunc := typeWriteFunc(arg.Type)
			w.linef(`spec.WriteField(_arg.Value(), %v, %v)`, name, writeFunc)
		case compiler.KindList:
			w.linef(`_arg.Value().Any(%v.Raw())`, name)
		case compiler.KindMessage:
			w.linef(`_arg.Value().Any(%v.Unwrap().Raw())`, name)
		case compiler.KindStruct:
			writeFunc := typeWriteFunc(arg.Type)
			w.linef(`spec.WriteField(_arg.Value(), %v, %v)`, name, writeFunc)

		case compiler.KindAny:
			w.linef(`_arg.Value().Any(%v)`, name)
		case compiler.KindAnyMessage:
			w.linef(`_arg.Value().Any(%v.Raw())`, name)

		default:
			return fmt.Errorf("unknown arg kind: %v", kind)
		}

		w.line(`if err := _arg.End(); err != nil {`)
		w.line(`return nil, status.WrapError(err)`)
		w.line(`}`)
		w.line(`}`)
	}

	// End args
	w.line(`if err := _args.End(); err != nil {`)
	w.line(`return nil, status.WrapError(err)`)
	w.line(`}`)

	// End call
	w.line(`if err := _call.End(); err != nil {`)
	w.line(`return nil, status.WrapError(err)`)
	w.line(`}`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) methodRequest(def *compiler.Definition, m *compiler.Method) error {
	// Send request
	w.line(`// Send request`)
	w.line(`_resp, st := _c.client.Request(cancel, _req)`)
	w.line(`if !st.OK() {`)
	w.line(`return nil, st`)
	w.line(`}`)
	w.line(``)

	// Make result types
	resultTypes := ""
	switch {
	case m.Chained:
		w.line(`return nil, status.OK`)
		w.line(`}`)
		return nil
	case len(m.Results) == 1:
		res := m.Results[0]
		typeName := typeRefName(res.Type)
		resultTypes += typeName
	default:
		resultTypes += "\n"
		for _, res := range m.Results {
			typeName := typeRefName(res.Type)
			resultTypes += typeName
			resultTypes += ",\n"
		}
	}

	// Make results
	// TODO: Check number of results in compiler, zero results
	w.line(`// Parse results`)
	switch {
	case m.Chained:
	case len(m.Results) == 1:
		w.linef(`_result := rpc.Result[%v]{}`, resultTypes)
	default:
		w.linef(`_result := rpc.Result%d[%v]{}`, len(m.Results), resultTypes)
	}

	// Parse results
	w.line(`_results := _resp.Results()`)
	w.line(`for i := 0; i < _results.Len(); i++ {`)
	w.line(`_res := _results.Get(i)`)
	w.line(`_name := _res.Name().Unwrap()`)
	w.line()

	w.line(`switch _name {`)
	for i, res := range m.Results {
		kind := res.Type.Kind
		field := [...]string{"A", "B", "C", "D", "E"}[i]
		w.linef(`case "%v":`, res.Name)

		switch kind {
		case compiler.KindBool:
			w.linef(`_result.%v = _res.Value().Bool()`, field)
		case compiler.KindByte:
			w.linef(`_result.%v = _res.Value().Byte()`, field)

		case compiler.KindInt16:
			w.linef(`_result.%v = _res.Value().Int16()`, field)
		case compiler.KindInt32:
			w.linef(`_result.%v = _res.Value().Int32()`, field)
		case compiler.KindInt64:
			w.linef(`_result.%v = _res.Value().Int64()`, field)

		case compiler.KindUint16:
			w.linef(`_result.%v = _res.Value().Uint16()`, field)
		case compiler.KindUint32:
			w.linef(`_result.%v = _res.Value().Uint32()`, field)
		case compiler.KindUint64:
			w.linef(`_result.%v = _res.Value().Uint64()`, field)

		case compiler.KindBin64:
			w.linef(`_result.%v = _res.Value().Bin64()`, field)
		case compiler.KindBin128:
			w.linef(`_result.%v = _res.Value().Bin128()`, field)
		case compiler.KindBin256:
			w.linef(`_result.%v = _res.Value().Bin256()`, field)

		case compiler.KindFloat32:
			w.linef(`_result.%v = _res.Value().Float32()`, field)
		case compiler.KindFloat64:
			w.linef(`_result.%v = _res.Value().Float64()`, field)

		case compiler.KindBytes:
			w.linef(`_result.%v = _res.Value().Bytes()`, field)
		case compiler.KindString:
			w.linef(`_result.%v = _res.Value().String()`, field)

		case compiler.KindList:
			decodeFunc := typeDecodeRefFunc(res.Type.Element)

			w.writef(`_result.%v = spec.NewTypedList(_res.Value(), %v)`, field, decodeFunc)
			w.line()

		case compiler.KindEnum,
			compiler.KindMessage,
			compiler.KindStruct:
			newFunc := typeNewFunc(res.Type)

			w.writef(`_result.%v = %v(_res.Value())`, field, newFunc)
			w.line()

		case compiler.KindAny:
			w.linef(`_result.%v = _res.Value()`, field)
		case compiler.KindAnyMessage:
			w.linef(`_result.%v = _res.Value().Message()`, field)

		default:
			return fmt.Errorf("unknown arg kind: %v", kind)
		}
	}
	w.line(`}`)
	w.line(`}`)
	w.line()

	// Return future
	w.line(`// Return future`)
	w.line(`_st := rpc.ParseStatus(_resp.Status())`)

	switch {
	case m.Chained:
	case len(m.Results) == 1:
		w.linef(`_future := rpc.Completed[%v](_result.A, _st)`, resultTypes)
	default:
		w.linef(`_future := rpc.Completed%d[%v](_result, _st)`, len(m.Results), resultTypes)
	}
	w.line(`return _future, status.OK`)
	w.line(`}`)
	return nil
}
