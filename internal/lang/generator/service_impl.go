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
	name := handler_name(def)

	w.linef(`// %v`, name)
	w.line()
	w.linef(`type %v struct {`, name)
	w.linef(`service %v`, def.Name)
	w.line(`}`)
	w.line()
	return nil
}

func (w *serviceImplWriter) handle(def *model.Definition) error {
	name := handler_name(def)

	if def.Service.Sub {
		w.linef(`func (h *%v) Handle(ctx async.Context, ch rpc.ServerChannel, index int) (ref.R[[]byte], status.Status) {`,
			name)
	} else {
		w.linef(`func (h *%v) Handle(ctx async.Context, ch rpc.ServerChannel) (ref.R[[]byte], status.Status) {`,
			name)
		w.line(`index := 0`)
	}

	w.line(`req, st := ch.Request(ctx)`)
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
		w.linef(`return h._%v(ctx, ch, call, index)`, toLowerCameCase(m.Name))
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
	name := handler_name(def)
	w.linef(`func (h *%v) _%v(ctx async.Context, ch rpc.ServerChannel, call prpc.Call, index int) (`,
		name, toLowerCameCase(m.Name))
	w.line(`ref.R[[]byte], status.Status) {`)

	// Parse input
	switch {
	case m.Chan:
		channelName := handlerChannel_name(m)
		w.line(`// Make channel`)
		w.linef(`ch1 := new%v(ch, call.Input())`, strings.Title(channelName))
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
		w.linef(`h.service.%v(ctx, ch1)`, toUpperCamelCase(m.Name))

	case m.Input != nil:
		w.linef(`h.service.%v(ctx, in)`, toUpperCamelCase(m.Name))

	case m.InputFields != nil:
		w.writef(`h.service.%v(ctx, `, toUpperCamelCase(m.Name))
		fields := m.InputFields.List
		for _, f := range fields {
			w.writef(`%v_, `, toLowerCameCase(f.Name))
		}
		w.line(`)`)

	default:
		w.linef(`h.service.%v(ctx)`, toUpperCamelCase(m.Name))
	}

	// Handle output
	switch {
	case m.Sub:
		newFunc := handler_new(m.Output)
		w.line(`if !st.OK() {`)
		w.line(`return nil, st`)
		w.line(`}`)
		w.line()
		w.line(`// Call subservice`)
		w.linef(`h1 := %v(sub)`, newFunc)
		w.line(`return h1.Handle(ctx, ch, index+1)`)

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
		w.line(`return ref.NextRetain(bytes, result), status.OK`)

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
	if err := w.channel_receive(def, m); err != nil {
		return err
	}
	if err := w.channel_send(def, m); err != nil {
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
	w.line(`req spec.Value`)
	w.line(`}`)
	w.line()
	w.linef(`func new%v(ch rpc.ServerChannel, req spec.Value) *%v {`, strings.Title(name), name)
	w.linef(`return &%v{ch: ch, req: req}`, name)
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
		w.linef(`req, _, err := %v(c.req)`, parseFunc)
		w.line(`if err != nil {`)
		w.linef(`return %v{}, rpc.WrapError(err)`, typeName)
		w.line(`}`)

		w.line(`c.req = nil`)
		w.line(`return req, status.OK`)
		w.line(`}`)
		w.line()

	case m.InputFields != nil:
		w.writef(`func (c *%v) Request() (`, name)

		fields := m.InputFields.List
		for _, f := range fields {
			name := toLowerCameCase(f.Name)
			w.writef(`_%v %v, `, name, typeName(f.Type))
		}
		w.line(`_st status.Status) {`)

		w.linef(`req, err := c.req.MessageErr()`)
		w.line(`if err != nil {`)
		w.line(`_st = status.WrapError(err)`)
		w.line(`return`)
		w.line(`}`)
		w.line()

		for _, f := range fields {
			typ := f.Type
			name := toLowerCameCase(f.Name)

			switch typ.Kind {
			case model.KindBool:
				w.linef(`_%v, err = req.Field(%d).BoolErr()`, name, f.Tag)
			case model.KindByte:
				w.linef(`_%v, err = req.Field(%d).ByteErr()`, name, f.Tag)

			case model.KindInt16:
				w.linef(`_%v, err = req.Field(%d).Int16Err()`, name, f.Tag)
			case model.KindInt32:
				w.linef(`_%v, err = req.Field(%d).Int32Err()`, name, f.Tag)
			case model.KindInt64:
				w.linef(`_%v, err = req.Field(%d).Int64Err()`, name, f.Tag)

			case model.KindUint16:
				w.linef(`_%v, err = req.Field(%d).Uint16Err()`, name, f.Tag)
			case model.KindUint32:
				w.linef(`_%v, err = req.Field(%d).Uint32Err()`, name, f.Tag)
			case model.KindUint64:
				w.linef(`_%v, err = req.Field(%d).Uint64Err()`, name, f.Tag)

			case model.KindBin64:
				w.linef(`_%v, err = req.Field(%d).Bin64Err()`, name, f.Tag)
			case model.KindBin128:
				w.linef(`_%v, err = req.Field(%d).Bin128Err()`, name, f.Tag)
			case model.KindBin256:
				w.linef(`_%v, err = req.Field(%d).Bin256Err()`, name, f.Tag)

			case model.KindFloat32:
				w.linef(`_%v, err = req.Field(%d).Float32Err()`, name, f.Tag)
			case model.KindFloat64:
				w.linef(`_%v, err = req.Field(%d).Float64Err()`, name, f.Tag)
			}

			w.line(`if err != nil {`)
			w.line(`_st = status.WrapError(err)`)
			w.line(`return`)
			w.line(`}`)
		}

		w.line(`c.req = nil`)
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

func (w *serviceImplWriter) channel_receive(def *model.Definition, m *model.Method) error {
	out := m.Channel.Out
	if out == nil {
		return nil
	}

	name := handlerChannel_name(m)
	typeName := typeName(out)
	parseFunc := typeParseFunc(out)

	// Receive
	w.linef(`func (c *%v) Receive(ctx async.Context) (%v, status.Status) {`, name, typeName)
	w.line(`b, st := c.ch.Receive(ctx)`)
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

	// ReceiveAsync
	w.linef(`func (c *%v) ReceiveAsync(ctx async.Context) (%v, bool, status.Status) {`, name, typeName)
	w.line(`b, ok, st := c.ch.ReceiveAsync(ctx)`)
	w.line(`switch {`)
	w.line(`case !st.OK():`)
	w.linef(`return %v{}, false, st`, typeName)
	w.line(`case !ok:`)
	w.linef(`return %v{}, false, status.OK`, typeName)
	w.line(`}`)
	w.linef(`msg, _, err := %v(b)`, parseFunc)
	w.line(`if err != nil {`)
	w.linef(`return %v{}, false, status.WrapError(err)`, typeName)
	w.line(`}`)
	w.line(`return msg, true, status.OK`)
	w.line(`}`)
	w.line()

	// ReceiveWait
	w.linef(`func (c *%v) ReceiveWait() <-chan struct{} {`, name)
	w.line(`return c.ch.ReceiveWait()`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *serviceImplWriter) channel_send(def *model.Definition, m *model.Method) error {
	in := m.Channel.In
	if in == nil {
		return nil
	}

	name := handlerChannel_name(m)
	typeName := typeName(in)

	// Send
	w.linef(`func (c *%v) Send(ctx async.Context, msg %v) status.Status {`, name, typeName)
	switch in.Kind {
	case model.KindList, model.KindMessage:
		w.line(`return c.ch.Send(ctx, msg.Unwrap().Raw())`)

	case model.KindStruct:
		writeFunc := typeWriteFunc(in)
		w.line(`buf := alloc.AcquireBuffer()`)
		w.line(`defer buf.Free()`)
		w.linef(`if _, err := %v(buf, msg); err != nil {`, writeFunc)
		w.line(`return status.WrapError(err)`)
		w.line(`}`)
		w.line(`return c.ch.Send(ctx, buf.Bytes())`)

	default:
		w.line(`return c.ch.Send(ctx, msg)`)
	}
	w.line(`}`)
	w.line()

	// SendEnd
	w.linef(`func (c *%v) SendEnd(ctx async.Context) status.Status {`, name)
	w.line(`return c.ch.SendEnd(ctx)`)
	w.line(`}`)
	w.line()
	return nil
}

func handler_name(def *model.Definition) string {
	return fmt.Sprintf(`%vHandler`, toLowerCameCase(def.Name))
}

func handler_new(typ *model.Type) string {
	if typ.Import != nil {
		return fmt.Sprintf(`%v.New%vHandler`, typ.ImportName, typ.Name)
	}
	return fmt.Sprintf(`New%vHandler`, typ.Name)
}

func handlerChannel_name(m *model.Method) string {
	return fmt.Sprintf("%v%vChannel", toLowerCameCase(m.Service.Def.Name), toUpperCamelCase(m.Name))
}
