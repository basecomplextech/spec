package generator

import (
	"fmt"
	"strings"

	"github.com/basecomplextech/spec/internal/lang/model"
)

func (w *writer) client(def *model.Definition) error {
	if err := w.clientDef(def); err != nil {
		return err
	}
	if err := w.clientMethods(def); err != nil {
		return err
	}
	if err := w.clientUnwrap(def); err != nil {
		return err
	}
	if err := w.clientChannels(def); err != nil {
		return err
	}
	return nil
}

func (w *writer) clientDef(def *model.Definition) error {
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
		w.line(`}`)
		w.line()
		w.linef(`func New%vClient(client rpc.Client) *%vClient {`, def.Name, def.Name)
		w.linef(`return &%vClient{client: client}`, def.Name)
		w.linef(`}`)
		w.line()
	}
	return nil
}

func (w *writer) clientMethods(def *model.Definition) error {
	for _, m := range def.Service.Methods {
		if err := w.clientMethod(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) clientMethod(def *model.Definition, m *model.Method) error {
	methodName := toUpperCamelCase(m.Name)
	w.writef(`func (c *%vClient) %v`, def.Name, methodName)

	if err := w.clientMethod_input(def, m); err != nil {
		return err
	}
	if err := w.clientMethod_output(def, m); err != nil {
		return err
	}
	w.line(`{`)

	if err := w.clientMethod_call(def, m); err != nil {
		return err
	}

	switch {
	case m.Sub:
		if err := w.clientMethod_subservice(def, m); err != nil {
			return err
		}
	case m.Chan:
		if err := w.clientMethod_channel(def, m); err != nil {
			return err
		}
	default:
		if err := w.clientMethod_request(def, m); err != nil {
			return err
		}
		if err := w.clientMethod_response(def, m); err != nil {
			return err
		}
	}

	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) clientMethod_input(def *model.Definition, m *model.Method) error {
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

func (w *writer) clientMethod_output(def *model.Definition, m *model.Method) error {
	switch {
	default:
		w.write(`(status.Status)`)

	case m.Sub:
		typeName := typeName(m.Output)
		w.line(`(`)
		w.writef(`*%vClient, status.Status)`, typeName)

	case m.Chan:
		name := clientChannel_name(m)
		w.line(`(`)
		w.writef(`*%v, status.Status)`, name)

	case m.Output != nil:
		typeName := typeName(m.Output)
		w.line(`(`)
		w.writef(`*ref.R[%v], status.Status)`, typeName)

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
	return nil
}

func (w *writer) clientMethod_call(def *model.Definition, m *model.Method) error {
	// Begin request
	if def.Service.Sub {
		w.line(`// Continue request`)
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
		w.linef(`return %v st`, clientMethod_zeroReturn(m))
		w.line(`}`)

	case m.Input != nil:
		w.linef(`st := req.AddMessage("%v", req_.Unwrap())`, m.Name)
		w.line(`if !st.OK() {`)
		w.linef(`return %v st`, clientMethod_zeroReturn(m))
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
		w.linef(`return %v status.WrapError(err)`, clientMethod_zeroReturn(m))
		w.line(`}`)
		w.line(`if err := call.End(); err != nil {`)
		w.linef(`return %v status.WrapError(err)`, clientMethod_zeroReturn(m))
		w.line(`}`)
		w.line(`}`)
	}

	// End request
	w.line()
	return nil
}

func (w *writer) clientMethod_subservice(def *model.Definition, m *model.Method) error {
	// Return subservice
	typeName := typeRefName(m.Output)

	w.line(`// Return subservice`)
	w.linef(`sub := New%vClient(c.client, req)`, typeName)
	w.line(`ok = true`)
	w.linef(`return sub, status.OK`)
	return nil
}

func (w *writer) clientMethod_channel(def *model.Definition, m *model.Method) error {
	// Build request
	w.line(`// Build request`)
	w.line(`preq, st := req.Build()`)
	w.line(`if !st.OK() {`)
	w.linef(`return %v st`, clientMethod_zeroReturn(m))
	w.line(`}`)
	w.line()

	// Open channel
	name := clientChannel_name(m)
	w.line(`// Open channel`)
	w.line(`ch, st := c.client.Channel(cancel, preq)`)
	w.line(`if !st.OK() {`)
	w.linef(`return %v st`, clientMethod_zeroReturn(m))
	w.line(`}`)
	w.linef(`return New%v(ch), status.OK`, name)
	return nil
}

func (w *writer) clientMethod_request(def *model.Definition, m *model.Method) error {
	// Build request
	w.line(`// Build request`)
	w.line(`preq, st := req.Build()`)
	w.line(`if !st.OK() {`)
	w.linef(`return %v st`, clientMethod_zeroReturn(m))
	w.line(`}`)
	w.line()

	// Send request
	w.line(`// Send request`)
	w.line(`resp, st := c.client.Request(cancel, preq)`)
	w.line(`if !st.OK() {`)
	w.linef(`return %v st`, clientMethod_zeroReturn(m))
	w.line(`}`)
	w.line(`defer resp.Release()`)
	w.line(``)
	return nil
}

func (w *writer) clientMethod_response(def *model.Definition, m *model.Method) error {
	switch {
	default:
		w.line(`return status.OK`)

	case m.Output != nil:
		parseFunc := typeParseFunc(m.Output)
		w.line(`// Parse result`)
		w.linef(`result, _, err := %v(resp.Unwrap())`, parseFunc)
		w.line(`if err != nil {`)
		w.linef(`return %v status.WrapError(err)`, clientMethod_zeroReturn(m))
		w.line(`}`)
		w.line(`return ref.NewParentRetain(result, resp), status.OK`)

	case m.OutputFields != nil:
		w.line(`// Parse results`)
		w.linef(`result, err := resp.Unwrap().MessageErr()`)
		w.line(`if err != nil {`)
		w.linef(`return %v status.WrapError(err)`, clientMethod_zeroReturn(m))
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
			w.linef(`return %v status.WrapError(err)`, clientMethod_zeroReturn(m))
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

func (w *writer) clientUnwrap(def *model.Definition) error {
	w.linef(`func (c *%vClient) Unwrap() rpc.Client {`, def.Name)
	w.line(`return c.client `)
	w.line(`}`)
	w.line()
	return nil
}

// channel

func (w *writer) clientChannels(def *model.Definition) error {
	for _, m := range def.Service.Methods {
		if !m.Chan {
			continue
		}

		if err := w.clientChannel(def, m); err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) clientChannel(def *model.Definition, m *model.Method) error {
	if err := w.clientChannel_def(def, m); err != nil {
		return err
	}
	if err := w.clientChannel_free(def, m); err != nil {
		return err
	}
	if err := w.clientChannel_send(def, m); err != nil {
		return err
	}
	if err := w.clientChannel_receive(def, m); err != nil {
		return err
	}
	if err := w.clientChannel_response(def, m); err != nil {
		return err
	}
	return nil
}

func (w *writer) clientChannel_def(def *model.Definition, m *model.Method) error {
	name := clientChannel_name(m)

	w.linef(`// %v`, name)
	w.line()
	w.linef(`type %v struct {`, name)
	w.line(`ch rpc.Channel`)
	w.line(`}`)
	w.line()
	w.linef(`func New%v(ch rpc.Channel) *%v {`, name, name)
	w.linef(`return &%v{ch: ch}`, name)
	w.linef(`}`)
	w.line()
	return nil
}

func (w *writer) clientChannel_free(def *model.Definition, m *model.Method) error {
	name := clientChannel_name(m)
	w.linef(`func (c *%v) Free() {`, name)
	w.line(`c.ch.Free()`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) clientChannel_send(def *model.Definition, m *model.Method) error {
	out := m.Channel.Out
	if out == nil {
		return nil
	}

	name := clientChannel_name(m)
	typeName := typeName(out)

	w.linef(`func (c *%v) Send(cancel <-chan struct{}, msg %v) status.Status {`, name, typeName)
	w.line(`return c.ch.Send(cancel, msg.Unwrap().Raw())`)
	w.line(`}`)
	w.line()
	return nil
}

func (w *writer) clientChannel_receive(def *model.Definition, m *model.Method) error {
	in := m.Channel.In
	if in == nil {
		return nil
	}

	name := clientChannel_name(m)
	typeName := typeName(in)
	parseFunc := typeParseFunc(in)

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

func (w *writer) clientChannel_response(def *model.Definition, m *model.Method) error {
	if err := w.clientChannel_response_def(m); err != nil {
		return err
	}
	if err := w.clientChannel_response_receive(m); err != nil {
		return err
	}
	if err := w.clientChannel_response_parse(m); err != nil {
		return err
	}
	return nil
}

func (w *writer) clientChannel_response_def(m *model.Method) error {
	// Response method
	name := clientChannel_name(m)
	w.writef(`func (c *%v) Response(cancel <-chan struct{}) `, name)

	switch {
	default:
		w.write(`(status.Status)`)

	case m.Output != nil:
		typeName := typeName(m.Output)
		w.writef(`(*ref.R[%v], status.Status)`, typeName)

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

func (w *writer) clientChannel_response_receive(m *model.Method) error {
	// Receive response
	w.line(`// Receive response`)
	w.line(`resp, st := c.ch.Response(cancel)`)
	w.line(`if !st.OK() {`)
	w.linef(`return %v st`, clientChannel_zeroReturn(m))
	w.line(`}`)
	w.line(`defer resp.Release()`)
	w.line(``)
	return nil
}

func (w *writer) clientChannel_response_parse(m *model.Method) error {
	// Parse results
	switch {
	default:
		w.line(`return status.OK`)

	case m.Output != nil:
		parseFunc := typeParseFunc(m.Output)
		w.line(`// Parse result`)
		w.linef(`result, _, err := %v(resp.Unwrap())`, parseFunc)
		w.line(`if err != nil {`)
		w.linef(`return %v status.WrapError(err)`, clientChannel_zeroReturn(m))
		w.line(`}`)
		w.line(`return ref.NewParentRetain(result, resp), status.OK`)

	case m.OutputFields != nil:
		w.line(`// Parse results`)
		w.linef(`result, err := resp.Unwrap().MessageErr()`)
		w.line(`if err != nil {`)
		w.linef(`return %v status.WrapError(err)`, clientChannel_zeroReturn(m))
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
			w.linef(`return %v status.WrapError(err)`, clientChannel_zeroReturn(m))
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

// util

func clientMethod_zeroReturn(m *model.Method) string {
	switch {
	default:
		return ``

	case m.Chan:
		return `nil, `

	case m.Output != nil:
		return `nil, `

	case m.OutputFields != nil:
		b := strings.Builder{}
		fields := m.OutputFields.List

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

func clientChannel_name(m *model.Method) string {
	return fmt.Sprintf("%v%vChannel", m.Service.Def.Name, toUpperCamelCase(m.Name))
}

func clientChannel_zeroReturn(m *model.Method) string {
	switch {
	default:
		return ``

	case m.Output != nil:
		return `nil, `

	case m.OutputFields != nil:
		b := strings.Builder{}
		fields := m.OutputFields.List

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
