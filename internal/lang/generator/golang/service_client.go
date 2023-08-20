package golang

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/compiler"
)

func (w *writer) client(def *compiler.Definition) error {
	if err := w.clientDef(def); err != nil {
		return err
	}
	if err := w.clientMethods(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) clientDef(def *compiler.Definition) error {
	if def.Service.Sub {
		w.linef(`// %vClient`, def.Name)
		w.line()
		w.linef(`type %vClient struct {`, def.Name)
		w.line(`client rpc.Client`)
		w.line(`req *rpc.Request`)
		w.line(`}`)
		w.line()
		w.linef(`func New%vClient(client rpc.Client, req *rpc.Request) *%vClient {`, def.Name, def.Name)
		w.linef(`return &%vClient{`, def.Name)
		w.linef(`client: client,`)
		w.linef(`req: req,`)
		w.linef(`}`)
		w.linef(`}`)
		w.line()
	} else {
		w.linef(`// %vClient`, def.Name)
		w.line()
		w.linef(`type %vClient struct {`, def.Name)
		w.line(`client rpc.Client`)
		w.line(`url string`)
		w.line(`}`)
		w.line()
		w.linef(`func New%vClient(client rpc.Client, url string) *%vClient {`, def.Name, def.Name)
		w.linef(`return &%vClient{`, def.Name)
		w.linef(`client: client,`)
		w.linef(`url: url,`)
		w.linef(`}`)
		w.linef(`}`)
		w.line()
	}
	return nil
}

func (w *writer) clientMethods(def *compiler.Definition) error {
	for _, m := range def.Service.Methods {
		if err := w.clientMethod(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) clientMethod(def *compiler.Definition, m *compiler.Method) error {
	methodName := toUpperCamelCase(m.Name)
	w.writef(`func (_c *%vClient) %v`, def.Name, methodName)

	if err := w.clientMethod_args(def, m); err != nil {
		return err
	}
	if err := w.clientMethod_results(def, m); err != nil {
		return err
	}
	w.line(`{`)

	if err := w.clientMethod_call(def, m); err != nil {
		return err
	}

	if m.Sub {
		if err := w.clientMethod_subservice(def, m); err != nil {
			return err
		}
	} else {
		if err := w.clientMethod_request(def, m); err != nil {
			return err
		}
	}

	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) clientMethod_args(def *compiler.Definition, m *compiler.Method) error {
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

func (w *writer) clientMethod_results(def *compiler.Definition, m *compiler.Method) error {
	if len(m.Results) > 1 {
		w.linef(`(`)
	} else {
		w.write(`(`)
	}

	for _, res := range m.Results {
		resName := toLowerCameCase(res.Name)
		typeName := typeRefName(res.Type)

		if res.Type.Kind == compiler.KindService {
			typeName = fmt.Sprintf("*%vClient", typeName)
		}

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
	w.write(`)`)
	return nil
}

func (w *writer) clientMethod_call(def *compiler.Definition, m *compiler.Method) error {
	return_ := methodReturn(m)

	// Make request
	if def.Service.Sub {
		w.line(`// Continue request`)
		w.line(`_req := _c.req`)
	} else {
		w.line(`// Begin request`)
		w.line(`_req := rpc.NewRequest(_c.url)`)
	}

	// Free request
	if m.Sub {
		w.line(`_ok := false`)
		w.line(`defer func() {`)
		w.line(`if !_ok {`)
		w.line(`_req.Free()`)
		w.line(`}`)
		w.line(`}()`)
		w.line()
	} else {
		w.line(`defer _req.Free()`)
		w.line()
	}

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
		w.linef(`return %v status.WrapError(err)`, return_)
		w.line(`}`)
		w.line(`}`)
	}

	// End args
	w.line(`if err := _args.End(); err != nil {`)
	w.linef(`return %v status.WrapError(err)`, return_)
	w.line(`}`)

	// End call
	w.line(`if err := _call.End(); err != nil {`)
	w.linef(`return %v status.WrapError(err)`, return_)
	w.line(`}`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) clientMethod_request(def *compiler.Definition, m *compiler.Method) error {
	return_ := methodReturn(m)

	// Send request
	w.line(`// Send request`)
	w.line(`_resp, st := _c.client.Request(cancel, _req)`)
	w.line(`if !st.OK() {`)
	w.linef(`return %v st`, return_)
	w.line(`}`)
	w.line(``)

	// Make result types
	resultTypes := ""
	if len(m.Results) == 1 {
		res := m.Results[0]
		typeName := typeRefName(res.Type)
		resultTypes += typeName
	} else {
		resultTypes += "\n"
		for _, res := range m.Results {
			typeName := typeRefName(res.Type)
			resultTypes += typeName
			resultTypes += ",\n"
		}
	}

	// Parse results
	w.line(`// Parse results`)
	w.line(`_results := _resp.Results()`)
	w.line(`for i := 0; i < _results.Len(); i++ {`)
	w.line(`_res, _err := _results.GetErr(i)`)
	w.line(`if _err != nil {`)
	w.linef(`return %v status.WrapError(_err)`, return_)
	w.line(`}`)
	w.line()

	w.line(`_name := _res.Name().Unwrap()`)
	w.line(`switch _name {`)
	for _, res := range m.Results {
		kind := res.Type.Kind
		name := toLowerCameCase(res.Name) + "_"
		w.linef(`case "%v":`, res.Name)

		switch kind {
		case compiler.KindBool:
			w.linef(`%v, _err = _res.Value().BoolErr()`, name)
		case compiler.KindByte:
			w.linef(`%v, _err = _res.Value().ByteErr()`, name)

		case compiler.KindInt16:
			w.linef(`%v, _err = _res.Value().Int16Err()`, name)
		case compiler.KindInt32:
			w.linef(`%v, _err = _res.Value().Int32Err()`, name)
		case compiler.KindInt64:
			w.linef(`%v, _err = _res.Value().Int64Err()`, name)

		case compiler.KindUint16:
			w.linef(`%v, _err = _res.Value().Uint16Err()`, name)
		case compiler.KindUint32:
			w.linef(`%v, _err = _res.Value().Uint32Err()`, name)
		case compiler.KindUint64:
			w.linef(`%v, _err = _res.Value().Uint64Err()`, name)

		case compiler.KindBin64:
			w.linef(`%v, _err = _res.Value().Bin64Err()`, name)
		case compiler.KindBin128:
			w.linef(`%v, _err = _res.Value().Bin128Err()`, name)
		case compiler.KindBin256:
			w.linef(`%v, _err = _res.Value().Bin256Err()`, name)

		case compiler.KindFloat32:
			w.linef(`%v, _err = _res.Value().Float32Err()`, name)
		case compiler.KindFloat64:
			w.linef(`%v, _err = _res.Value().Float64Err()`, name)

		case compiler.KindBytes:
			w.linef(`%v, _err = _res.Value().BytesErr()`, name)
		case compiler.KindString:
			w.linef(`%v, _err = _res.Value().StringErr()`, name)

		case compiler.KindList:
			decodeFunc := typeDecodeRefFunc(res.Type.Element)

			w.writef(`%v, _, _err = spec.ParseTypedList(_res.Value(), %v)`, name, decodeFunc)
			w.line()

		case compiler.KindEnum,
			compiler.KindMessage,
			compiler.KindStruct:
			parseFunc := typeParseFunc(res.Type)

			w.writef(`%v, _, _err = %v(_res.Value())`, name, parseFunc)
			w.line()

		case compiler.KindAny:
			w.linef(`%v = _res.Value()`, name)
		case compiler.KindAnyMessage:
			w.linef(`%v, _err = _res.Value().MessageErr()`, name)

		default:
			return fmt.Errorf("unknown arg kind: %v", kind)
		}
	}
	w.line(`}`)
	w.line(`if _err != nil {`)
	w.linef(`return %v rpc.WrapErrorf(_err, "Invalid result")`, return_)
	w.line(`}`)

	w.line(`}`)
	w.line()

	w.line(`// Return result`)
	w.line(`st_ = rpc.ParseStatus(_resp.Status())`)
	w.linef(`return %v st_`, return_)
	return nil
}

func (w *writer) clientMethod_subservice(def *compiler.Definition, m *compiler.Method) error {
	// Call subservice
	res := m.Results[0]
	resName := toLowerCameCase(res.Name) + "_"
	typeName := typeRefName(res.Type)

	w.line(`// Return subservice`)
	w.linef(`%v = New%vClient(_c.client, _req)`, resName, typeName)
	w.line(`_ok = true`)
	w.linef(`return %v, status.OK`, resName)
	return nil
}

func methodReturn(m *compiler.Method) string {
	s := ""
	for _, res := range m.Results {
		s += toLowerCameCase(res.Name) + "_, "
	}
	return s
}
