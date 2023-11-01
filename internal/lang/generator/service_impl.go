package generator

import (
	"fmt"
	"strings"

	"github.com/basecomplextech/spec/internal/lang/model"
)

type serviceImplWriter struct {
	*writer
}

func newServiceImplWriter(w *writer) *serviceImplWriter {
	return &serviceImplWriter{w}
}

func (w *serviceImplWriter) serviceImpl(def *model.Definition) error {
	if err := w.def(def); err != nil {
		return err
	}
	if err := w.handle(def); err != nil {
		return err
	}
	if err := w.methods(def); err != nil {
		return err
	}
	if err := w.channels(def); err != nil {
		return err
	}
	return nil
}

func (w *serviceImplWriter) def(def *model.Definition) error {
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

func (w *serviceImplWriter) handle(def *model.Definition) error {
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

func (w *serviceImplWriter) methods(def *model.Definition) error {
	for _, m := range def.Service.Methods {
		if err := w.method(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *serviceImplWriter) method(def *model.Definition, m *model.Method) error {
	// Declare method
	w.linef(`func (h *%vHandler) _%v(cancel <-chan struct{}, ch rpc.ServerChannel, call prpc.Call, index int) (`,
		def.Name, toLowerCameCase(m.Name))
	w.line(`*ref.R[[]byte], status.Status) {`)

	// Parse input
	switch {
	case m.Chan:
		w.line(`// Make channel`)
		w.linef(`ch1 := New%v(ch)`, handlerChannel_name(m))
		w.line(`ch1.in = call.Input()`)
		w.line()

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
	case m.Chan:
		w.linef(`h.service.%v(cancel, ch1)`, toUpperCamelCase(m.Name))

	case m.Input != nil:
		w.linef(`h.service.%v(cancel, in)`, toUpperCamelCase(m.Name))

	case m.InputFields != nil:
		w.writef(`h.service.%v(cancel, `, toUpperCamelCase(m.Name))
		fields := m.InputFields.List
		for _, f := range fields {
			w.writef(`%v_, `, toLowerCameCase(f.Name))
		}
		w.line(`)`)

	default:
		w.linef(`h.service.%v(cancel)`, toUpperCamelCase(m.Name))
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

func (w *serviceImplWriter) channels(def *model.Definition) error {
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

func (w *serviceImplWriter) channel(def *model.Definition, m *model.Method) error {
	if err := w.channel_def(def, m); err != nil {
		return err
	}
	if err := w.channel_request(def, m); err != nil {
		return err
	}
	if err := w.channel_send(def, m); err != nil {
		return err
	}
	if err := w.channel_receive(def, m); err != nil {
		return err
	}
	return nil
}

func (w *serviceImplWriter) channel_def(def *model.Definition, m *model.Method) error {
	name := handlerChannel_name(m)

	w.linef(`// %v`, name)
	w.line()
	w.linef(`type %v struct {`, name)
	w.line(`ch rpc.ServerChannel`)
	w.line(`in spec.Value`)
	w.line(`}`)
	w.line()
	w.linef(`func New%v(ch rpc.ServerChannel) *%v {`, name, name)
	w.linef(`return &%v{ch: ch}`, name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *serviceImplWriter) channel_request(def *model.Definition, m *model.Method) error {
	name := handlerChannel_name(m)

	switch {
	case m.Input != nil:
		typeName := typeName(m.Input)
		parseFunc := typeParseFunc(m.Input)

		w.linef(`func (c *%v) Request() (%v, status.Status) {`, name, typeName)
		w.linef(`in, _, err := %v(c.in)`, parseFunc)
		w.line(`if err != nil {`)
		w.linef(`return %v{}, rpc.WrapError(err)`, typeName)
		w.line(`}`)

		w.line(`c.in = nil`)
		w.line(`return in, status.OK`)
		w.line(`}`)
		w.line()

	case m.InputFields != nil:
		w.writef(`func (c *%v) Request() (`, name)

		fields := m.InputFields.List
		for _, f := range fields {
			w.writef(`%v, `, typeName(f.Type))
		}
		w.line(`status.Status) {`)

		w.linef(`in, err := c.in.MessageErr()`)
		w.line(`if err != nil {`)
		w.linef(`return %v status.WrapError(err)`, handlerChannel_zeroReturn1(m.Input, m.InputFields))
		w.line(`}`)
		w.line()

		for _, f := range fields {
			typ := f.Type
			name := toLowerCameCase(f.Name)

			switch typ.Kind {
			case model.KindBool:
				w.linef(`_%v, err := in.Field(%d).BoolErr()`, name, f.Tag)
			case model.KindByte:
				w.linef(`_%v, err := in.Field(%d).ByteErr()`, name, f.Tag)

			case model.KindInt16:
				w.linef(`_%v, err := in.Field(%d).Int16Err()`, name, f.Tag)
			case model.KindInt32:
				w.linef(`_%v, err := in.Field(%d).Int32Err()`, name, f.Tag)
			case model.KindInt64:
				w.linef(`_%v, err := in.Field(%d).Int64Err()`, name, f.Tag)

			case model.KindUint16:
				w.linef(`_%v, err := in.Field(%d).Uint16Err()`, name, f.Tag)
			case model.KindUint32:
				w.linef(`_%v, err := in.Field(%d).Uint32Err()`, name, f.Tag)
			case model.KindUint64:
				w.linef(`_%v, err := in.Field(%d).Uint64Err()`, name, f.Tag)

			case model.KindBin64:
				w.linef(`_%v, err := in.Field(%d).Bin64Err()`, name, f.Tag)
			case model.KindBin128:
				w.linef(`_%v, err := in.Field(%d).Bin128Err()`, name, f.Tag)
			case model.KindBin256:
				w.linef(`_%v, err := in.Field(%d).Bin256Err()`, name, f.Tag)

			case model.KindFloat32:
				w.linef(`_%v, err := in.Field(%d).Float32Err()`, name, f.Tag)
			case model.KindFloat64:
				w.linef(`_%v, err := in.Field(%d).Float64Err()`, name, f.Tag)
			}

			w.line(`if err != nil {`)
			w.linef(`return %v status.WrapError(err)`, handlerChannel_zeroReturn1(m.Input, m.InputFields))
			w.line(`}`)
		}

		w.line(`c.in = nil`)
		w.write(`return `)
		for _, f := range fields {
			w.writef(`_%v, `, toLowerCameCase(f.Name))
		}
		w.line(`status.OK`)
		w.line(`}`)
		w.line()
	}
	return nil
}

func (w *serviceImplWriter) channel_send(def *model.Definition, m *model.Method) error {
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

func (w *serviceImplWriter) channel_receive(def *model.Definition, m *model.Method) error {
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

func handlerChannel_zeroReturn(m *model.Method) string {
	return handlerChannel_zeroReturn1(m.Output, m.OutputFields)
}

func handlerChannel_zeroReturn1(output *model.Type, outputFields *model.Fields) string {
	switch {
	default:
		return ``

	case output != nil:
		return `nil, `

	case outputFields != nil:
		b := strings.Builder{}
		fields := outputFields.List

		for _, f := range fields {
			typ := f.Type
			switch typ.Kind {
			case model.KindBool:
				b.WriteString(`false, `)
			case model.KindByte:
				b.WriteString(`0, `)

			case model.KindInt16:
				b.WriteString(`0, `)
			case model.KindInt32:
				b.WriteString(`0, `)
			case model.KindInt64:
				b.WriteString(`0, `)

			case model.KindUint16:
				b.WriteString(`0, `)
			case model.KindUint32:
				b.WriteString(`0, `)
			case model.KindUint64:
				b.WriteString(`0, `)

			case model.KindBin64:
				b.WriteString(`bin.Bin64{}, `)
			case model.KindBin128:
				b.WriteString(`bin.Bin128{}, `)
			case model.KindBin256:
				b.WriteString(`bin.Bin256{}, `)

			case model.KindFloat32:
				b.WriteString(`0, `)
			case model.KindFloat64:
				b.WriteString(`0, `)
			}
		}

		return b.String()
	}
}