package generator

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/model"
)

func (w *writer) handler(def *model.Definition) error {
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

func (w *writer) handlerDef(def *model.Definition) error {
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

func (w *writer) handlerHandle(def *model.Definition) error {
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

	w.linef(`return rpc.Errorf("Unknown %v method %%q", method)`, def.Name)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) handlerMethods(def *model.Definition) error {
	for _, m := range def.Service.Methods {
		if err := w.handlerMethod(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) handlerMethod(def *model.Definition, m *model.Method) error {
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
				case model.KindBool:
					w.linef(`%v, err = arg.Value().BoolErr()`, name)
				case model.KindByte:
					w.linef(`%v, err = arg.Value().ByteErr()`, name)

				case model.KindInt16:
					w.linef(`%v, err = arg.Value().Int16Err()`, name)
				case model.KindInt32:
					w.linef(`%v, err = arg.Value().Int32Err()`, name)
				case model.KindInt64:
					w.linef(`%v, err = arg.Value().Int64Err()`, name)

				case model.KindUint16:
					w.linef(`%v, err = arg.Value().Uint16Err()`, name)
				case model.KindUint32:
					w.linef(`%v, err = arg.Value().Uint32Err()`, name)
				case model.KindUint64:
					w.linef(`%v, err = arg.Value().Uint64Err()`, name)

				case model.KindBin64:
					w.linef(`%v, err = arg.Value().Bin64Err()`, name)
				case model.KindBin128:
					w.linef(`%v, err = arg.Value().Bin128Err()`, name)
				case model.KindBin256:
					w.linef(`%v, err = arg.Value().Bin256Err()`, name)

				case model.KindFloat32:
					w.linef(`%v, err = arg.Value().Float32Err()`, name)
				case model.KindFloat64:
					w.linef(`%v, err = arg.Value().Float64Err()`, name)

				case model.KindBytes:
					w.linef(`%v, err = arg.Value().BytesErr()`, name)
				case model.KindString:
					w.linef(`%v, err = arg.Value().StringErr()`, name)

				case model.KindList:
					decodeFunc := typeDecodeRefFunc(arg.Type.Element)

					w.writef(`%v, _, err = spec.ParseTypedList(arg.Value(), %v)`, name, decodeFunc)
					w.line()

				case model.KindEnum,
					model.KindMessage,
					model.KindStruct:
					parseFunc := typeParseFunc(arg.Type)

					w.writef(`%v, _, err = %v(arg.Value())`, name, parseFunc)
					w.line()

				case model.KindAny:
					w.linef(`%v = arg.Value()`, name)
				case model.KindAnyMessage:
					w.linef(`%v, err = arg.Value().MessageErr()`, name)

				default:
					return fmt.Errorf("unknown arg kind: %v", kind)
				}
			}
			w.line(`}`)

			w.line(`if err != nil {`)
			w.line(`return rpc.WrapErrorf(err, "Invalid argument %q", name)`)
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
			w.line(`res := resp.Result()`)
			w.linef(`res.Name("%v")`, res.Name)

			switch kind {
			case model.KindBool:
				w.linef(`res.Value().Bool(%v_)`, name)
			case model.KindByte:
				w.linef(`res.Value().Byte(%v_)`, name)

			case model.KindInt16:
				w.linef(`res.Value().Int16(%v_)`, name)
			case model.KindInt32:
				w.linef(`res.Value().Int32(%v_)`, name)
			case model.KindInt64:
				w.linef(`res.Value().Int64(%v_)`, name)

			case model.KindUint16:
				w.linef(`res.Value().Uint16(%v_)`, name)
			case model.KindUint32:
				w.linef(`res.Value().Uint32(%v_)`, name)
			case model.KindUint64:
				w.linef(`res.Value().Uint64(%v_)`, name)

			case model.KindBin64:
				w.linef(`res.Value().Bin64(%v_)`, name)
			case model.KindBin128:
				w.linef(`res.Value().Bin128(%v_)`, name)
			case model.KindBin256:
				w.linef(`res.Value().Bin256(%v_)`, name)

			case model.KindFloat32:
				w.linef(`res.Value().Float32(%v_)`, name)
			case model.KindFloat64:
				w.linef(`res.Value().Float64(%v_)`, name)

			case model.KindBytes:
				w.linef(`res.Value().Bytes(%v_)`, name)
			case model.KindString:
				w.linef(`res.Value().String(%v_)`, name)

			case model.KindEnum:
				writeFunc := typeWriteFunc(res.Type)
				w.linef(`spec.WriteField(res.Value(), %v_, %v)`, name, writeFunc)
			case model.KindList:
				w.linef(`res.Value().Any(%v_.Raw())`, name)
			case model.KindMessage:
				w.linef(`res.Value().Any(%v_.Unwrap().Raw())`, name)
			case model.KindStruct:
				writeFunc := typeWriteFunc(res.Type)
				w.linef(`spec.WriteField(res.Value(), %v_, %v)`, name, writeFunc)

			case model.KindAny:
				w.linef(`res.Value().Any(%v_)`, name)
			case model.KindAnyMessage:
				w.linef(`res.Value().Any(%v_.Raw())`, name)

			default:
				return fmt.Errorf("unknown arg kind: %v", kind)
			}

			w.line(`if err := res.End(); err != nil {`)
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
