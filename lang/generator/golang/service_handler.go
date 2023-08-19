package golang

import (
	"fmt"

	"github.com/basecomplextech/spec/lang/compiler"
)

func (w *writer) handler(def *compiler.Definition) error {
	if err := w.handlerDef(def); err != nil {
		return err
	}
	if err := w.handlerHandle(def); err != nil {
		return err
	}
	if err := w.handlerMethods(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) handlerDef(def *compiler.Definition) error {
	w.linef(`// %vHandler`, def.Name)
	w.line()
	w.linef(`type %vHandler struct {`, def.Name)
	w.linef(`service %v`, def.Name)
	w.line(`}`)
	w.line()
	w.linef(`func New%vHandler(s %v) *%vHandler {`, def.Name, def.Name, def.Name)
	w.linef(`return &%vHandler{service: s}`, def.Name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) handlerHandle(def *compiler.Definition) error {
	if def.Service.Sub {
		w.linef(`func (h *%vHandler) Handle(cancel <-chan struct{},
			req *rpc.ServerRequest,
			resp rpc.ServerResponse,
			index int,
		) status.Status {`, def.Name)
	} else {
		w.linef(`func (h *%vHandler) Handle(cancel <-chan struct{},
			req *rpc.ServerRequest,
			resp rpc.ServerResponse,
		) status.Status {
		index := 0
		`, def.Name)
	}

	w.line(`call, st := req.Call(index)
	if !st.OK() {
		return st
	}

	method := call.Method()
	switch method {`)

	for _, m := range def.Service.Methods {
		w.linef(`case %q:`, m.Name)
		w.linef(`return h._%v(cancel, req, resp, index)`, toLowerCameCase(m.Name))
	}
	w.line(`}`)
	w.line()

	w.linef(`return status.Newf("rpc_error", "Unknown %v method %%q", method)`, def.Name)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) handlerMethods(def *compiler.Definition) error {
	for _, m := range def.Service.Methods {
		if err := w.handlerMethod(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) handlerMethod(def *compiler.Definition, m *compiler.Method) error {
	methodName := toLowerCameCase(m.Name)
	// Declare method
	w.linef(`func (h *%vHandler) _%v(cancel <-chan struct{},
			req *rpc.ServerRequest,
			resp rpc.ServerResponse,
			index int,
		) status.Status {`, def.Name, methodName)

	// Get call
	{
		w.line(`call, st := req.Call(index)
	if !st.OK() {
		return st
	}`)
		w.line()
	}

	if len(m.Args) == 0 {
		w.line(`// No args`)
		w.line(`var _ = call`)
		w.line(``)
	} else {
		// Declare args
		{
			w.line(`// Declare args`)
			w.line(`var (`)
			for _, arg := range m.Args {
				typeName := typeRefName(arg.Type)
				w.linef(`_%v %v`, toLowerCameCase(arg.Name), typeName)
			}
			w.line(`)`)
			w.line()
		}

		// Parse args
		{
			w.line(`// Parse args`)
			w.line(`args := call.Args()`)
			w.line(`for i := 0; i < args.Len(); i++ {`)
			w.line(`arg := args.Get(i)`)
			w.line(`name := arg.Name()`)
			w.line(`var err error`)
			w.line()

			w.line(`switch name {`)
			for _, arg := range m.Args {
				kind := arg.Type.Kind
				name := "_" + toLowerCameCase(arg.Name)
				w.linef(`case "%v":`, arg.Name)

				switch kind {
				case compiler.KindBool:
					w.linef(`%v, err = arg.Value().BoolErr()`, name)
				case compiler.KindByte:
					w.linef(`%v, err = arg.Value().ByteErr()`, name)

				case compiler.KindInt16:
					w.linef(`%v, err = arg.Value().Int16Err()`, name)
				case compiler.KindInt32:
					w.linef(`%v, err = arg.Value().Int32Err()`, name)
				case compiler.KindInt64:
					w.linef(`%v, err = arg.Value().Int64Err()`, name)

				case compiler.KindUint16:
					w.linef(`%v, err = arg.Value().Uint16Err()`, name)
				case compiler.KindUint32:
					w.linef(`%v, err = arg.Value().Uint32Err()`, name)
				case compiler.KindUint64:
					w.linef(`%v, err = arg.Value().Uint64Err()`, name)

				case compiler.KindBin64:
					w.linef(`%v, err = arg.Value().Bin64Err()`, name)
				case compiler.KindBin128:
					w.linef(`%v, err = arg.Value().Bin128Err()`, name)
				case compiler.KindBin256:
					w.linef(`%v, err = arg.Value().Bin256Err()`, name)

				case compiler.KindFloat32:
					w.linef(`%v, err = arg.Value().Float32Err()`, name)
				case compiler.KindFloat64:
					w.linef(`%v, err = arg.Value().Float64Err()`, name)

				case compiler.KindBytes:
					w.linef(`%v, err = arg.Value().BytesErr()`, name)
				case compiler.KindString:
					w.linef(`%v, err = arg.Value().StringErr()`, name)

				case compiler.KindList:
					decodeFunc := typeDecodeRefFunc(arg.Type.Element)

					w.writef(`%v, _, err = spec.ParseTypedList(arg.Value(), %v)`, name, decodeFunc)
					w.line()

				case compiler.KindEnum,
					compiler.KindMessage,
					compiler.KindStruct:
					parseFunc := typeParseFunc(arg.Type)

					w.writef(`%v, _, err = %v(arg.Value())`, name, parseFunc)
					w.line()

				case compiler.KindAny:
					w.linef(`%v = arg.Value()`, name)
				case compiler.KindAnyMessage:
					w.linef(`%v, err = arg.Value().MessageErr()`, name)

				default:
					return fmt.Errorf("unknown arg kind: %v", kind)
				}
			}
			w.line(`}`)

			w.line(`if err != nil {`)
			w.line(`return status.Newf("rpc_error", "Invalid argument %q: %v", name, err)`)
			w.line(`}`)
			w.line(`}`)
			w.line()
		}
	}

	// Call method
	{
		name := toUpperCamelCase(m.Name)
		w.line(`// Call method`)

		for _, result := range m.Results {
			resultName := toLowerCameCase(result.Name)
			w.writef(`%v_, `, resultName)
		}

		w.writef(`st_ := h.service.%v(cancel`, name)
		for _, arg := range m.Args {
			if len(m.Args) > 3 {
				w.line(`, `)
			} else {
				w.write(`, `)
			}
			argName := toLowerCameCase(arg.Name)
			w.writef(`_%v`, argName)
		}
		if len(m.Args) > 3 {
			w.line(`, `)
		}
		w.line(`)`)

		w.line(`if !st_.OK() {`)
		w.line(`return st_`)
		w.line(`}`)
	}

	// Call subservice
	if m.Sub {
		res := m.Results[0]
		resName := toLowerCameCase(res.Name)

		w.line()
		w.line(`// Handle next call`)
		w.linef(`h1 := NewSubserviceHandler(%v_)`, resName)
		w.line(`return h1.Handle(cancel, req, resp, index+1)`)
		w.line(`}`)
		return nil
	}

	// Serialize results
	if len(m.Results) > 0 {
		w.line()
		w.line(`// Serialize results`)
		for _, res := range m.Results {
			kind := res.Type.Kind
			name := toLowerCameCase(res.Name)

			w.line(`{`)
			w.linef(`val := resp.Result("%v")`, res.Name)

			switch kind {
			case compiler.KindBool:
				w.linef(`err := val.Bool(%v_)`, name)
			case compiler.KindByte:
				w.linef(`err := val.Byte(%v_)`, name)

			case compiler.KindInt16:
				w.linef(`err := val.Int16(%v_)`, name)
			case compiler.KindInt32:
				w.linef(`err := val.Int32(%v_)`, name)
			case compiler.KindInt64:
				w.linef(`err := val.Int64(%v_)`, name)

			case compiler.KindUint16:
				w.linef(`err := val.Uint16(%v_)`, name)
			case compiler.KindUint32:
				w.linef(`err := val.Uint32(%v_)`, name)
			case compiler.KindUint64:
				w.linef(`err := val.Uint64(%v_)`, name)

			case compiler.KindBin64:
				w.linef(`err := val.Bin64(%v_)`, name)
			case compiler.KindBin128:
				w.linef(`err := val.Bin128(%v_)`, name)
			case compiler.KindBin256:
				w.linef(`err := val.Bin256(%v_)`, name)

			case compiler.KindFloat32:
				w.linef(`err := val.Float32(%v_)`, name)
			case compiler.KindFloat64:
				w.linef(`err := val.Float64(%v_)`, name)

			case compiler.KindBytes:
				w.linef(`err := val.Bytes(%v_)`, name)
			case compiler.KindString:
				w.linef(`err := val.String(%v_)`, name)

			case compiler.KindEnum:
				writeFunc := typeWriteFunc(res.Type)
				w.linef(`err := spec.WriteField(val, %v_, %v)`, name, writeFunc)
			case compiler.KindList:
				w.linef(`err := val.Any(%v_.Raw())`, name)
			case compiler.KindMessage:
				w.linef(`err := val.Any(%v_.Unwrap().Raw())`, name)
			case compiler.KindStruct:
				writeFunc := typeWriteFunc(res.Type)
				w.linef(`err := spec.WriteField(val, %v_, %v)`, name, writeFunc)

			case compiler.KindAny:
				w.linef(`err := val.Any(%v_)`, name)
			case compiler.KindAnyMessage:
				w.linef(`err := val.Any(%v_.Raw())`, name)

			default:
				return fmt.Errorf("unknown arg kind: %v", kind)
			}

			w.line(`if err != nil {`)
			w.linef(`return status.WrapError(err)`)
			w.line(`}`)
			w.line(`}`)
		}
	}

	// Return
	w.line(`return status.OK`)
	w.line(`}`)
	w.line()
	return nil
}
