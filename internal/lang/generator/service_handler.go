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
	if err := w.handlerChannels(def); err != nil {
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
		w.linef(`func (h *%vHandler) Handle(cancel <-chan struct{}, ch rpc.ServerChannel, index int) (*ref.R[[]byte], status.Status) {`,
			def.Name)
	} else {
		w.linef(`func (h *%vHandler) Handle(cancel <-chan struct{}, ch rpc.ServerChannel) (*ref.R[[]byte], status.Status) {`,
			def.Name)
		w.line(`index := 0`)
	}

	w.line(`req, st := ch.Request(cancel)`)
	w.line(`if !st.OK() {`)
	w.line(`return nil, st`)
	w.line(`}`)
	w.line()

	w.line(`call, err := req.Calls().GetErr(index)`)
	w.line(`if err != nil {`)
	w.line(`return nil, rpc.WrapError(err)`)
	w.line(`}`)
	w.line()

	w.line(`method := call.Method()`)
	w.line(`switch method {`)
	for _, m := range def.Service.Methods {
		w.linef(`case %q:`, m.Name)
		w.linef(`return h._%v(cancel, ch, call, index)`, toLowerCameCase(m.Name))
	}
	w.line(`}`)
	w.line()

	w.linef(`return nil, rpc.Errorf("unknown %v method %%q", method)`, def.Name)
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
	// Declare method
	w.linef(`func (h *%vHandler) _%v(cancel <-chan struct{}, ch rpc.ServerChannel, call prpc.Call, index int) (`,
		def.Name, toLowerCameCase(m.Name))
	w.line(`*ref.R[[]byte], status.Status) {`)

	// Parse input
	switch {
	case m.Input != nil:
		parseFunc := typeParseFunc(m.Input)
		w.line(`// Parse input`)
		w.linef(`in, _, err := %v(call.Input())`, parseFunc)
		w.line(`if err != nil {`)
		w.line(`return nil, rpc.WrapError(err)`)
		w.line(`}`)
		w.line()

	case m.InputFields != nil:
		w.line(`// Parse input`)
		w.linef(`in, err := call.Input().MessageErr()`)
		w.line(`if err != nil {`)
		w.line(`return nil, rpc.WrapError(err)`)
		w.line(`}`)

		fields := m.InputFields.List
		for _, f := range fields {
			typ := f.Type
			name := toLowerCameCase(f.Name)

			switch typ.Kind {
			case model.KindBool:
				w.linef(`%v_, err := in.Field(%d).BoolErr()`, name, f.Tag)
			case model.KindByte:
				w.linef(`%v_, err := in.Field(%d).ByteErr()`, name, f.Tag)

			case model.KindInt16:
				w.linef(`%v_, err := in.Field(%d).Int16Err()`, name, f.Tag)
			case model.KindInt32:
				w.linef(`%v_, err := in.Field(%d).Int32Err()`, name, f.Tag)
			case model.KindInt64:
				w.linef(`%v_, err := in.Field(%d).Int64Err()`, name, f.Tag)

			case model.KindUint16:
				w.linef(`%v_, err := in.Field(%d).Uint16Err()`, name, f.Tag)
			case model.KindUint32:
				w.linef(`%v_, err := in.Field(%d).Uint32Err()`, name, f.Tag)
			case model.KindUint64:
				w.linef(`%v_, err := in.Field(%d).Uint64Err()`, name, f.Tag)

			case model.KindBin64:
				w.linef(`%v_, err := in.Field(%d).Bin64Err()`, name, f.Tag)
			case model.KindBin128:
				w.linef(`%v_, err := in.Field(%d).Bin128Err()`, name, f.Tag)
			case model.KindBin256:
				w.linef(`%v_, err := in.Field(%d).Bin256Err()`, name, f.Tag)

			case model.KindFloat32:
				w.linef(`%v_, err := in.Field(%d).Float32Err()`, name, f.Tag)
			case model.KindFloat64:
				w.linef(`%v_, err := in.Field(%d).Float64Err()`, name, f.Tag)
			}

			w.line(`if err != nil {`)
			w.linef(`return nil, status.WrapError(err)`)
			w.line(`}`)
		}
		w.line()
	}

	// Make channels
	channel := ""
	if m.Chan {
		w.line(`// Make channel`)
		w.linef(`ch1 := New%v(ch)`, handlerChannel_name(m))
		w.line()
		channel = ", ch1"
	}

	// Declare result
	w.line(`// Call method`)
	switch {
	case m.Sub:
		w.write(`sub, st := `)
	case m.Output != nil:
		w.write(`result, st := `)
	case m.OutputFields != nil:
		fields := m.OutputFields.List
		for _, f := range fields {
			w.writef(`_%v, `, toLowerCameCase(f.Name))
		}
		w.write(`st := `)
	default:
		w.write(`st := `)
	}

	// Call method
	switch {
	case m.Input != nil:
		w.linef(`h.service.%v(cancel%v, in)`, toUpperCamelCase(m.Name), channel)
	case m.InputFields != nil:
		w.writef(`h.service.%v(cancel%v, `, toUpperCamelCase(m.Name), channel)
		fields := m.InputFields.List
		for _, f := range fields {
			w.writef(`%v_, `, toLowerCameCase(f.Name))
		}
		w.line(`)`)
	default:
		w.linef(`h.service.%v(cancel%v)`, toUpperCamelCase(m.Name), channel)
	}

	// Handle output
	switch {
	case m.Sub:
		w.line(`if !st.OK() {`)
		w.line(`return nil, st`)
		w.line(`}`)
		w.line()
		w.line(`// Call subservice`)
		w.linef(`h1 := New%vHandler(sub)`, typeName(m.Output))
		w.line(`return h1.Handle(cancel, ch, index+1)`)

	case m.Output != nil:
		w.line(`if result != nil { `)
		w.line(`defer result.Release() `)
		w.line(`}`)
		w.line(`if !st.OK() {`)
		w.line(`return nil, st`)
		w.line(`}`)
		w.line()
		w.line(`// Return bytes`)
		w.line(`bytes := result.Unwrap().Unwrap().Raw()`)
		w.line(`return ref.NewParentRetain(bytes, result), status.OK`)

	case m.OutputFields != nil:
		w.line(`if !st.OK() {`)
		w.line(`return nil, st`)
		w.line(`}`)
		w.line()

		w.line(`// Build output`)
		w.line(`buf := rpc.NewBuffer()`)
		w.line(`out := spec.NewMessageWriterBuffer(buf)`)

		fields := m.OutputFields.List
		for _, f := range fields {
			typ := f.Type
			name := toLowerCameCase(f.Name)

			switch typ.Kind {
			case model.KindBool:
				w.linef(`out.Field(%d).Bool(_%v)`, f.Tag, name)
			case model.KindByte:
				w.linef(`out.Field(%d).Byte(_%v)`, f.Tag, name)

			case model.KindInt16:
				w.linef(`out.Field(%d).Int16(_%v)`, f.Tag, name)
			case model.KindInt32:
				w.linef(`out.Field(%d).Int32(_%v)`, f.Tag, name)
			case model.KindInt64:
				w.linef(`out.Field(%d).Int64(_%v)`, f.Tag, name)

			case model.KindUint16:
				w.linef(`out.Field(%d).Uint16(_%v)`, f.Tag, name)
			case model.KindUint32:
				w.linef(`out.Field(%d).Uint32(_%v)`, f.Tag, name)
			case model.KindUint64:
				w.linef(`out.Field(%d).Uint64(_%v)`, f.Tag, name)

			case model.KindBin64:
				w.linef(`out.Field(%d).Bin64(_%v)`, f.Tag, name)
			case model.KindBin128:
				w.linef(`out.Field(%d).Bin128(_%v)`, f.Tag, name)
			case model.KindBin256:
				w.linef(`out.Field(%d).Bin256(_%v)`, f.Tag, name)

			case model.KindFloat32:
				w.linef(`out.Field(%d).Float32(_%v)`, f.Tag, name)
			case model.KindFloat64:
				w.linef(`out.Field(%d).Float64(_%v)`, f.Tag, name)
			}
		}
		w.line(`bytes, err := out.Build()`)
		w.line(`if err != nil {`)
		w.line(`buf.Free()`)
		w.line(`return nil, status.WrapError(err)`)
		w.line(`}`)
		w.line(`return ref.NewFreer(bytes, buf), status.OK`)

	default:
		w.line(`return nil, st`)
	}

	w.line(`}`)
	w.line()
	return nil
}

// channels

func (w *writer) handlerChannels(def *model.Definition) error {
	for _, m := range def.Service.Methods {
		if !m.Chan {
			continue
		}

		if err := w.handlerChannel(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) handlerChannel(def *model.Definition, m *model.Method) error {
	if err := w.handlerChannel_def(def, m); err != nil {
		return err
	}
	if err := w.handlerChannel_send(def, m); err != nil {
		return err
	}
	if err := w.handlerChannel_receive(def, m); err != nil {
		return err
	}
	return nil
}

func (w *writer) handlerChannel_def(def *model.Definition, m *model.Method) error {
	name := handlerChannel_name(m)

	w.linef(`// %v`, name)
	w.line()
	w.linef(`type %v struct {`, name)
	w.line(`ch rpc.ServerChannel`)
	w.line(`}`)
	w.line()
	w.linef(`func New%v(ch rpc.ServerChannel) *%v {`, name, name)
	w.linef(`return &%v{ch: ch}`, name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) handlerChannel_send(def *model.Definition, m *model.Method) error {
	in := m.Channel.In
	if in == nil {
		return nil
	}

	name := handlerChannel_name(m)
	typeName := typeName(in)

	w.linef(`func (c *%v) Send(cancel <-chan struct{}, msg %v) status.Status {`, name, typeName)
	w.line(`return c.ch.Send(cancel, msg.Unwrap().Raw())`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) handlerChannel_receive(def *model.Definition, m *model.Method) error {
	out := m.Channel.Out
	if out == nil {
		return nil
	}

	name := handlerChannel_name(m)
	typeName := typeName(out)
	parseFunc := typeParseFunc(out)

	w.linef(`func (c *%v) Receive(cancel <-chan struct{}) (%v, status.Status) {`, name, typeName)
	w.line(`b, st := c.ch.Receive(cancel)`)
	w.line(`if !st.OK() {`)
	w.linef(`return %v{}, st`, typeName)
	w.line(`}`)
	w.linef(`msg, _, err := %v(b)`, parseFunc)
	w.line(`if err != nil {`)
	w.linef(`return %v{}, status.WrapError(err)`, typeName)
	w.line(`}`)
	w.line(`return msg, status.OK`)
	w.line(`}`)
	w.line()
	return nil
}

func handlerChannel_name(m *model.Method) string {
	return fmt.Sprintf("%v%vServerChannel", m.Service.Def.Name, toUpperCamelCase(m.Name))
}
