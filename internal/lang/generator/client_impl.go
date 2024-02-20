package generator

import (
	"fmt"
	"strings"

	"github.com/basecomplextech/spec/internal/lang/model"
)

type clientImplWriter struct {
	*writer
}

func newClientImplWriter(w *writer) *clientImplWriter {
	return &clientImplWriter{w}
}

func (w *writer) clientImpl(def *model.Definition) error {
	w1 := &clientImplWriter{w}
	return w1.clientImpl(def)
}

func (w *clientImplWriter) clientImpl(def *model.Definition) error {
	if err := w.def(def); err != nil {
		return err
	}
	if err := w.methods(def); err != nil {
		return err
	}
	if err := w.unwrap(def); err != nil {
		return err
	}
	if err := w.channels(def); err != nil {
		return err
	}
	return nil
}

func (w *clientImplWriter) def(def *model.Definition) error {
	name := clientImplName(def)

	if def.Service.Sub {
		w.linef(`// %v`, name)
		w.line()
		w.linef(`type %v struct {`, name)
		w.line(`client rpc.Client`)
		w.line(`req *rpc.Request`)
		w.line(`st status.Status`)
		w.line(`}`)
		w.line()
	} else {
		w.linef(`// %v`, name)
		w.line()
		w.linef(`type %v struct {`, name)
		w.line(`client rpc.Client`)
		w.line(`}`)
		w.line()
	}
	return nil
}

func (w *clientImplWriter) methods(def *model.Definition) error {
	for _, m := range def.Service.Methods {
		if err := w.method(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *clientImplWriter) method(def *model.Definition, m *model.Method) error {
	name := clientImplName(def)
	methodName := toUpperCamelCase(m.Name)
	w.writef(`func (c *%v) %v`, name, methodName)

	if err := w.method_input(def, m); err != nil {
		return err
	}
	if err := w.method_output(def, m); err != nil {
		return err
	}
	w.line(`{`)

	if err := w.method_call(def, m); err != nil {
		return err
	}

	switch {
	case m.Sub:
		if err := w.method_subservice(def, m); err != nil {
			return err
		}
	case m.Chan:
		if err := w.method_channel(def, m); err != nil {
			return err
		}
	default:
		if err := w.method_request(def, m); err != nil {
			return err
		}
		if err := w.method_response(def, m); err != nil {
			return err
		}
	}

	w.line(`}`)
	w.line()
	return nil
}

func (w *clientImplWriter) method_input(def *model.Definition, m *model.Method) error {
	cancel := "cancel <-chan struct{}, "
	if m.Sub {
		cancel = ""
	}

	switch {
	default:
		w.writef(`(%v) `, cancel)

	case m.Input != nil:
		typeName := typeName(m.Input)
		w.writef(`(%v req_ %v) `, cancel, typeName)

	case m.InputFields != nil:
		w.writef(`(%v`, cancel)

		fields := m.InputFields.List
		multi := len(fields) > 3
		if multi {
			w.line()
		}

		for _, field := range fields {
			argName := toLowerCameCase(field.Name)
			typeName := typeName(field.Type)

			if multi {
				w.linef(`%v_ %v, `, argName, typeName)
			} else {
				w.writef(`%v_ %v, `, argName, typeName)
			}
		}

		w.write(`)`)
	}
	return nil
}

func (w *clientImplWriter) method_output(def *model.Definition, m *model.Method) error {
	switch {
	default:
		w.write(`(_st status.Status)`)

	case m.Sub:
		typeName := typeName(m.Output)
		w.writef(`%vClient`, typeName)

	case m.Chan:
		name := clientChannel_name(m)
		w.writef(`(_ %v, _st status.Status)`, name)

	case m.Output != nil:
		typeName := typeName(m.Output)
		w.writef(`(_ *ref.R[%v], _st status.Status)`, typeName)

	case m.OutputFields != nil:
		fields := m.OutputFields.List
		multi := len(fields) > 3

		if multi {
			w.line(`(`)
		} else {
			w.write(`(`)
		}

		for _, f := range fields {
			name := toLowerCameCase(f.Name)
			typeName := typeName(f.Type)

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

func (w *clientImplWriter) method_error(def *model.Definition, m *model.Method) error {
	if !m.Sub {
		w.line(`return`)
		return nil
	}

	name := clientImplNewErr(m.Output)
	w.linef(`return %v(_st)`, name)
	return nil
}

func (w *clientImplWriter) method_call(def *model.Definition, m *model.Method) error {
	// Subservice methods do not return status
	if m.Sub {
		w.line(`var _st status.Status`)
		w.line(``)
	}

	// Begin request
	if def.Service.Sub {
		w.line(`// Continue request`)
		w.line(`if _st = c.st; !_st.OK() {`)
		w.method_error(def, m)
		w.line(`}`)
		w.line(`req := c.req`)
		w.line(`c.req = nil`)
	} else {
		w.line(`// Begin request`)
		w.line(`req := rpc.NewRequest()`)
	}

	// Free request
	if m.Sub {
		w.line(`ok := false`)
		w.line(`defer func() {`)
		w.line(`if !ok {`)
		w.line(`req.Free()`)
		w.line(`}`)
		w.line(`}()`)
		w.line()
	} else {
		w.line(`defer req.Free()`)
		w.line()
	}

	// Add call
	w.line(`// Add call`)
	switch {
	default:
		w.linef(`st := req.AddEmpty("%v")`, m.Name)
		w.line(`if !st.OK() {`)
		w.line(`_st = st`)
		w.method_error(def, m)
		w.line(`}`)

	case m.Input != nil:
		w.linef(`st := req.AddMessage("%v", req_.Unwrap())`, m.Name)
		w.line(`if !st.OK() {`)
		w.line(`_st = st`)
		w.method_error(def, m)
		w.line(`}`)

	case m.InputFields != nil:
		w.line(`{`)
		w.linef(`call := req.Add("%v")`, m.Name)
		w.line(`in := call.Input().Message()`)

		for _, f := range m.InputFields.List {
			typ := f.Type
			switch typ.Kind {
			case model.KindBool:
				w.linef(`in.Field(%d).Bool(%v_)`, f.Tag, f.Name)
			case model.KindByte:
				w.linef(`in.Field(%d).Byte(%v_)`, f.Tag, f.Name)

			case model.KindInt16:
				w.linef(`in.Field(%d).Int16(%v_)`, f.Tag, f.Name)
			case model.KindInt32:
				w.linef(`in.Field(%d).Int32(%v_)`, f.Tag, f.Name)
			case model.KindInt64:
				w.linef(`in.Field(%d).Int64(%v_)`, f.Tag, f.Name)

			case model.KindUint16:
				w.linef(`in.Field(%d).Uint16(%v_)`, f.Tag, f.Name)
			case model.KindUint32:
				w.linef(`in.Field(%d).Uint32(%v_)`, f.Tag, f.Name)
			case model.KindUint64:
				w.linef(`in.Field(%d).Uint64(%v_)`, f.Tag, f.Name)

			case model.KindBin64:
				w.linef(`in.Field(%d).Bin64(%v_)`, f.Tag, f.Name)
			case model.KindBin128:
				w.linef(`in.Field(%d).Bin128(%v_)`, f.Tag, f.Name)
			case model.KindBin256:
				w.linef(`in.Field(%d).Bin256(%v_)`, f.Tag, f.Name)

			case model.KindFloat32:
				w.linef(`in.Field(%d).Float32(%v_)`, f.Tag, f.Name)
			case model.KindFloat64:
				w.linef(`in.Field(%d).Float64(%v_)`, f.Tag, f.Name)

			case model.KindString:
				w.linef(`in.Field(%d).String(%v_)`, f.Tag, f.Name)
			case model.KindBytes:
				w.linef(`in.Field(%d).Bytes(%v_)`, f.Tag, f.Name)
			}
		}

		w.line()
		w.line(`if err := in.End(); err != nil {`)
		w.line(`_st = status.WrapError(err)`)
		w.method_error(def, m)
		w.line(`}`)
		w.line(`if err := call.End(); err != nil {`)
		w.line(`_st = status.WrapError(err)`)
		w.method_error(def, m)
		w.line(`}`)
		w.line(`}`)
	}

	// End request
	w.line()
	return nil
}

func (w *clientImplWriter) method_subservice(def *model.Definition, m *model.Method) error {
	// Return subservice
	newFunc := clientImplNew(m.Output)

	w.line(`// Return subservice`)
	w.linef(`sub := %v(c.client, req)`, newFunc)
	w.line(`ok = true`)
	w.linef(`return sub`)
	return nil
}

func (w *clientImplWriter) method_channel(def *model.Definition, m *model.Method) error {
	// Build request
	w.line(`// Build request`)
	w.line(`preq, st := req.Build()`)
	w.line(`if !st.OK() {`)
	w.line(`_st = st`)
	w.line(`return`)
	w.line(`}`)
	w.line()

	// Open channel
	name := clientChannelImpl_name(m)
	w.line(`// Open channel`)
	w.line(`ch, st := c.client.Channel(cancel, preq)`)
	w.line(`if !st.OK() {`)
	w.line(`_st = st`)
	w.line(`return`)
	w.line(`}`)
	w.linef(`return new%v(ch), status.OK`, strings.Title(name))
	return nil
}

func (w *clientImplWriter) method_request(def *model.Definition, m *model.Method) error {
	// Build request
	w.line(`// Build request`)
	w.line(`preq, st := req.Build()`)
	w.line(`if !st.OK() {`)
	w.line(`_st = st`)
	w.line(`return`)
	w.line(`}`)
	w.line()

	// Send request
	w.line(`// Send request`)
	w.line(`resp, st := c.client.Request(cancel, preq)`)
	w.line(`if !st.OK() {`)
	w.line(`_st = st`)
	w.line(`return`)
	w.line(`}`)
	w.line(`defer resp.Release()`)
	w.line(``)
	return nil
}

func (w *clientImplWriter) method_response(def *model.Definition, m *model.Method) error {
	switch {
	default:
		w.line(`return status.OK`)

	case m.Output != nil:
		parseFunc := typeParseFunc(m.Output)
		w.line(`// Parse result`)
		w.linef(`result, _, err := %v(resp.Unwrap())`, parseFunc)
		w.line(`if err != nil {`)
		w.line(`_st = status.WrapError(err)`)
		w.line(`return`)
		w.line(`}`)
		w.line(`return ref.NewParentRetain(result, resp), status.OK`)

	case m.OutputFields != nil:
		w.line(`// Parse results`)
		w.linef(`result, err := resp.Unwrap().MessageErr()`)
		w.line(`if err != nil {`)
		w.line(`_st = status.WrapError(err)`)
		w.line(`return`)
		w.line(`}`)

		fields := m.OutputFields.List
		for _, f := range fields {
			typ := f.Type
			name := toLowerCameCase(f.Name)

			switch typ.Kind {
			case model.KindBool:
				w.linef(`_%v, err = result.Field(%d).BoolErr()`, name, f.Tag)
			case model.KindByte:
				w.linef(`_%v, err = result.Field(%d).ByteErr()`, name, f.Tag)

			case model.KindInt16:
				w.linef(`_%v, err = result.Field(%d).Int16Err()`, name, f.Tag)
			case model.KindInt32:
				w.linef(`_%v, err = result.Field(%d).Int32Err()`, name, f.Tag)
			case model.KindInt64:
				w.linef(`_%v, err = result.Field(%d).Int64Err()`, name, f.Tag)

			case model.KindUint16:
				w.linef(`_%v, err = result.Field(%d).Uint16Err()`, name, f.Tag)
			case model.KindUint32:
				w.linef(`_%v, err = result.Field(%d).Uint32Err()`, name, f.Tag)
			case model.KindUint64:
				w.linef(`_%v, err = result.Field(%d).Uint64Err()`, name, f.Tag)

			case model.KindBin64:
				w.linef(`_%v, err = result.Field(%d).Bin64Err()`, name, f.Tag)
			case model.KindBin128:
				w.linef(`_%v, err = result.Field(%d).Bin128Err()`, name, f.Tag)
			case model.KindBin256:
				w.linef(`_%v, err = result.Field(%d).Bin256Err()`, name, f.Tag)

			case model.KindFloat32:
				w.linef(`_%v, err = result.Field(%d).Float32Err()`, name, f.Tag)
			case model.KindFloat64:
				w.linef(`_%v, err = result.Field(%d).Float64Err()`, name, f.Tag)
			}

			w.line(`if err != nil {`)
			w.line(`_st = status.WrapError(err)`)
			w.line(`return`)
			w.line(`}`)
		}

		w.write(`return `)
		for _, f := range fields {
			name := toLowerCameCase(f.Name)
			w.writef(`_%v, `, name)
		}
		w.write(`status.OK`)
	}

	return nil
}

// unwrap

func (w *clientImplWriter) unwrap(def *model.Definition) error {
	name := clientImplName(def)
	w.linef(`func (c *%v) Unwrap() rpc.Client {`, name)
	w.line(`return c.client `)
	w.line(`}`)
	w.line()
	return nil
}

// channel

func (w *clientImplWriter) channels(def *model.Definition) error {
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

func (w *clientImplWriter) channel(def *model.Definition, m *model.Method) error {
	if err := w.channel_def(def, m); err != nil {
		return err
	}
	if err := w.channel_read(def, m); err != nil {
		return err
	}
	if err := w.channel_write(def, m); err != nil {
		return err
	}
	if err := w.channel_response(def, m); err != nil {
		return err
	}
	if err := w.channel_free(def, m); err != nil {
		return err
	}
	return nil
}

func (w *clientImplWriter) channel_def(def *model.Definition, m *model.Method) error {
	name := clientChannelImpl_name(m)

	w.linef(`// %v`, name)
	w.line()
	w.linef(`type %v struct {`, name)
	w.line(`ch rpc.Channel`)
	w.line(`}`)
	w.line()
	w.linef(`func new%v(ch rpc.Channel) *%v {`, strings.Title(name), name)
	w.linef(`return &%v{ch: ch}`, name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *clientImplWriter) channel_read(def *model.Definition, m *model.Method) error {
	in := m.Channel.In
	if in == nil {
		return nil
	}

	name := clientChannelImpl_name(m)
	typeName := typeName(in)
	parseFunc := typeParseFunc(in)

	// Read
	w.linef(`func (c *%v) Read(cancel <-chan struct{}) (%v, bool, status.Status) {`, name, typeName)
	w.line(`b, ok, st := c.ch.Read(cancel)`)
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

	// ReadSync
	w.linef(`func (c *%v) ReadSync(cancel <-chan struct{}) (%v, status.Status) {`, name, typeName)
	w.line(`b, st := c.ch.ReadSync(cancel)`)
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

	// ReadWait
	w.linef(`func (c *%v) ReadWait() <-chan struct{} {`, name)
	w.line(`return c.ch.ReadWait()`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *clientImplWriter) channel_write(def *model.Definition, m *model.Method) error {
	out := m.Channel.Out
	if out == nil {
		return nil
	}

	name := clientChannelImpl_name(m)
	typeName := typeName(out)

	// Write
	w.linef(`func (c *%v) Write(cancel <-chan struct{}, msg %v) status.Status {`, name, typeName)
	switch out.Kind {
	case model.KindList, model.KindMessage:
		w.line(`return c.ch.Write(cancel, msg.Unwrap().Raw())`)
	default:
		w.line(`return c.ch.Write(cancel, msg)`)
	}
	w.line(`}`)
	w.line()

	// WriteEnd
	w.linef(`func (c *%v) WriteEnd(cancel <-chan struct{}) status.Status {`, name)
	w.line(`return c.ch.WriteEnd(cancel)`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *clientImplWriter) channel_response(def *model.Definition, m *model.Method) error {
	if err := w.channel_response_def(m); err != nil {
		return err
	}
	if err := w.channel_response_receive(m); err != nil {
		return err
	}
	if err := w.channel_response_parse(m); err != nil {
		return err
	}
	return nil
}

func (w *clientImplWriter) channel_response_def(m *model.Method) error {
	// Response method
	name := clientChannelImpl_name(m)
	w.writef(`func (c *%v) Response(cancel <-chan struct{}) `, name)

	switch {
	default:
		w.write(`(_st status.Status)`)

	case m.Output != nil:
		typeName := typeName(m.Output)
		w.writef(`(_ *ref.R[%v], _st status.Status)`, typeName)

	case m.OutputFields != nil:
		fields := m.OutputFields.List
		multi := len(fields) > 3
		w.line(`(`)

		for _, f := range fields {
			name := toLowerCameCase(f.Name)
			typeName := typeName(f.Type)
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

	w.line(`{`)
	return nil
}

func (w *clientImplWriter) channel_response_receive(m *model.Method) error {
	// Receive response
	w.line(`// Receive response`)
	w.line(`resp, st := c.ch.Response(cancel)`)
	w.line(`if !st.OK() {`)
	w.line(`_st = st`)
	w.line(`return`)
	w.line(`}`)
	w.line(`defer resp.Release()`)
	w.line(``)
	return nil
}

func (w *clientImplWriter) channel_response_parse(m *model.Method) error {
	// Parse results
	switch {
	default:
		w.line(`return status.OK`)

	case m.Output != nil:
		parseFunc := typeParseFunc(m.Output)
		w.line(`// Parse result`)
		w.linef(`result, _, err := %v(resp.Unwrap())`, parseFunc)
		w.line(`if err != nil {`)
		w.line(`_st = status.WrapError(err)`)
		w.line(`return`)
		w.line(`}`)
		w.line(`return ref.NewParentRetain(result, resp), status.OK`)

	case m.OutputFields != nil:
		w.line(`// Parse results`)
		w.linef(`result, err := resp.Unwrap().MessageErr()`)
		w.line(`if err != nil {`)
		w.line(`_st = status.WrapError(err)`)
		w.line(`return`)
		w.line(`}`)

		fields := m.OutputFields.List
		for _, f := range fields {
			typ := f.Type
			name := toLowerCameCase(f.Name)

			switch typ.Kind {
			case model.KindBool:
				w.linef(`_%v, err = result.Field(%d).BoolErr()`, name, f.Tag)
			case model.KindByte:
				w.linef(`_%v, err = result.Field(%d).ByteErr()`, name, f.Tag)

			case model.KindInt16:
				w.linef(`_%v, err = result.Field(%d).Int16Err()`, name, f.Tag)
			case model.KindInt32:
				w.linef(`_%v, err = result.Field(%d).Int32Err()`, name, f.Tag)
			case model.KindInt64:
				w.linef(`_%v, err = result.Field(%d).Int64Err()`, name, f.Tag)

			case model.KindUint16:
				w.linef(`_%v, err = result.Field(%d).Uint16Err()`, name, f.Tag)
			case model.KindUint32:
				w.linef(`_%v, err = result.Field(%d).Uint32Err()`, name, f.Tag)
			case model.KindUint64:
				w.linef(`_%v, err = result.Field(%d).Uint64Err()`, name, f.Tag)

			case model.KindBin64:
				w.linef(`_%v, err = result.Field(%d).Bin64Err()`, name, f.Tag)
			case model.KindBin128:
				w.linef(`_%v, err = result.Field(%d).Bin128Err()`, name, f.Tag)
			case model.KindBin256:
				w.linef(`_%v, err = result.Field(%d).Bin256Err()`, name, f.Tag)

			case model.KindFloat32:
				w.linef(`_%v, err = result.Field(%d).Float32Err()`, name, f.Tag)
			case model.KindFloat64:
				w.linef(`_%v, err = result.Field(%d).Float64Err()`, name, f.Tag)
			}

			w.line(`if err != nil {`)
			w.line(`_st = status.WrapError(err)`)
			w.line(`return`)
			w.line(`}`)
		}

		w.write(`return `)
		for _, f := range fields {
			name := toLowerCameCase(f.Name)
			w.writef(`_%v, `, name)
		}
		w.write(`status.OK`)
	}

	w.line(`}`)
	w.line()
	return nil
}

func (w *clientImplWriter) channel_free(def *model.Definition, m *model.Method) error {
	name := clientChannelImpl_name(m)
	w.linef(`func (c *%v) Free() {`, name)
	w.line(`c.ch.Free()`)
	w.line(`}`)
	w.line()
	return nil
}

// util

func clientImplName(def *model.Definition) string {
	return fmt.Sprintf("%vClient", toLowerCameCase(def.Name))
}

func clientImplNew(typ *model.Type) string {
	if typ.Import != nil {
		return fmt.Sprintf("%v.New%vClient", typ.ImportName, typ.Name)
	}
	return fmt.Sprintf("New%vClient", typ.Name)
}

func clientImplNewErr(typ *model.Type) string {
	if typ.Import != nil {
		return fmt.Sprintf("%v.New%vClientErr", typ.ImportName, typ.Name)
	}
	return fmt.Sprintf("New%vClientErr", typ.Name)
}

func clientChannelImpl_name(m *model.Method) string {
	return fmt.Sprintf("%v%vClientChannel", toLowerCameCase(m.Service.Def.Name), toUpperCamelCase(m.Name))
}
