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
	if err := w.free(def); err != nil {
		return err
	}
	if err := w.channels(def); err != nil {
		return err
	}
	return nil
}

func (w *clientImplWriter) def(def *model.Definition) error {
	name := clientImplName(def)
	w.linef(`// %v`, name)
	w.line()

	if def.Service.Sub {
		w.linef(`var %vPool = pools.NewPoolFunc(`, name)
		w.linef(`func() *%v {`, name)
		w.linef(`return &%v{}`, name)
		w.line(`},`)
		w.line(`)`)
		w.line()
		w.linef(`type %v struct {`, name)
		w.line(`client rpc.Client`)
		w.line(`req *rpc.Request`)
		w.line(`st status.Status`)
		w.line(`}`)
		w.line()
		w.linef(`func new%vCall(client rpc.Client, req *rpc.Request) %vCall {`, def.Name, def.Name)
		w.linef(`c := %vPool.New()`, name)
		w.line(`c.client = client`)
		w.line(`c.req = req`)
		w.line(`c.st = status.OK`)
		w.line(`return c`)
		w.line(`}`)
		w.line()
	} else {
		w.linef(`type %v struct {`, name)
		w.line(`client rpc.Client`)
		w.line(`}`)
		w.line()
	}
	return nil
}

func (w *clientImplWriter) free(def *model.Definition) error {
	if !def.Service.Sub {
		return nil
	}

	name := clientImplName(def)
	w.linef(`func (c *%v) free() {`, name)
	w.linef(`*c = %v{}`, name)
	w.linef(`%vPool.Put(c)`, name)
	w.line(`}`)
	w.line()
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
	case m.Subservice != nil:
		if err := w.method_subservice(def, m); err != nil {
			return err
		}
	case m.Channel != nil:
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
	ctx := "ctx async.Context, "
	if m.Subservice != nil {
		ctx = ""
	}

	switch {
	default:
		w.writef(`(%v) `, ctx)

	case m.Request != nil:
		typeName := typeName(m.Request)
		w.writef(`(%v req_ %v) `, ctx, typeName)
	}
	return nil
}

func (w *clientImplWriter) method_output(def *model.Definition, m *model.Method) error {
	switch {
	default:
		w.write(`(_st status.Status)`)

	case m.Subservice != nil:
		typeName := typeName(m.Subservice)
		w.writef(`%vCall`, typeName)

	case m.Channel != nil:
		name := clientChannel_name(m)
		w.writef(`(_ %v, _st status.Status)`, name)

	case m.Response != nil:
		typeName := typeName(m.Response)
		w.writef(`(_ ref.R[%v], _st status.Status)`, typeName)
	}
	return nil
}

func (w *clientImplWriter) method_error(def *model.Definition, m *model.Method) error {
	if m.Type != model.MethodType_Subservice {
		w.line(`return`)
		return nil
	}

	name := clientImplNewErr(m.Subservice)
	w.linef(`return %v(_st)`, name)
	return nil
}

func (w *clientImplWriter) method_call(def *model.Definition, m *model.Method) error {
	if def.Service.Sub {
		w.line(`defer c.free()`)
		w.line()
	}

	// Subservice methods do not return status
	if m.Subservice != nil {
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
	if m.Subservice != nil {
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

	case m.Request != nil:
		w.linef(`st := req.AddMessage("%v", req_.Unwrap())`, m.Name)
		w.line(`if !st.OK() {`)
		w.line(`_st = st`)
		w.method_error(def, m)
		w.line(`}`)
	}

	// End request
	w.line()
	return nil
}

func (w *clientImplWriter) method_subservice(def *model.Definition, m *model.Method) error {
	// Return subservice
	newFunc := clientImplNew(m.Subservice)

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
	w.line(`ch, st := c.client.Channel(ctx, preq)`)
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
	if m.Oneway {
		w.line(`// Send request`)
		w.line(`return c.client.RequestOneway(ctx, preq)`)
	} else {
		w.line(`// Send request`)
		w.line(`resp, st := c.client.Request(ctx, preq)`)
		w.line(`if !st.OK() {`)
		w.line(`_st = st`)
		w.line(`return`)
		w.line(`}`)
		w.line(`defer resp.Release()`)
		w.line(``)
	}
	return nil
}

func (w *clientImplWriter) method_response(def *model.Definition, m *model.Method) error {
	switch {
	default:
		w.line(`return status.OK`)

	case m.Oneway:
		// pass

	case m.Response != nil:
		parseFunc := typeParseFunc(m.Response)
		w.line(`// Parse result`)
		w.linef(`result, _, err := %v(resp.Unwrap())`, parseFunc)
		w.line(`if err != nil {`)
		w.line(`_st = status.WrapError(err)`)
		w.line(`return`)
		w.line(`}`)
		w.line(`return ref.NextRetain(result, resp), status.OK`)
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
		if m.Channel == nil {
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
	if err := w.channel_receive(def, m); err != nil {
		return err
	}
	if err := w.channel_send(def, m); err != nil {
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

func (w *clientImplWriter) channel_send(def *model.Definition, m *model.Method) error {
	in := m.Channel.In
	if in == nil {
		return nil
	}

	name := clientChannelImpl_name(m)
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

func (w *clientImplWriter) channel_receive(def *model.Definition, m *model.Method) error {
	out := m.Channel.Out
	if out == nil {
		return nil
	}

	name := clientChannelImpl_name(m)
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
	w.writef(`func (c *%v) Response(ctx async.Context) `, name)

	switch {
	default:
		w.write(`(_st status.Status)`)

	case m.Response != nil:
		typeName := typeName(m.Response)
		w.writef(`(_ %v, _st status.Status)`, typeName)
	}

	w.line(`{`)
	return nil
}

func (w *clientImplWriter) channel_response_receive(m *model.Method) error {
	// Receive response
	w.line(`// Receive response`)
	w.line(`resp, st := c.ch.Response(ctx)`)
	w.line(`if !st.OK() {`)
	w.line(`_st = st`)
	w.line(`return`)
	w.line(`}`)
	w.line(``)
	return nil
}

func (w *clientImplWriter) channel_response_parse(m *model.Method) error {
	// Parse results
	switch {
	default:
		w.line(`_ = resp`)
		w.line(`return status.OK`)

	case m.Response != nil:
		parseFunc := typeParseFunc(m.Response)
		w.line(`// Parse result`)
		w.linef(`result, _, err := %v(resp)`, parseFunc)
		w.line(`if err != nil {`)
		w.line(`_st = status.WrapError(err)`)
		w.line(`return`)
		w.line(`}`)
		w.line(`return result, status.OK`)
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
	if def.Service.Sub {
		return fmt.Sprintf("%vCall", toLowerCameCase(def.Name))
	}
	return fmt.Sprintf("%vClient", toLowerCameCase(def.Name))
}

func clientImplNew(typ *model.Type) string {
	var name string
	if typ.Ref.Service.Sub {
		name = fmt.Sprintf("New%vCall", typ.Name)
	} else {
		name = fmt.Sprintf("New%vClient", typ.Name)
	}

	if typ.Import != nil {
		return fmt.Sprintf("%v.%v", typ.ImportName, name)
	}
	return name
}

func clientImplNewErr(typ *model.Type) string {
	var name string
	if typ.Ref.Service.Sub {
		name = fmt.Sprintf("New%vCallErr", typ.Name)
	} else {
		name = fmt.Sprintf("New%vClientErr", typ.Name)
	}

	if typ.Import != nil {
		return fmt.Sprintf("%v.%v", typ.ImportName, name)
	}
	return name
}

func clientChannelImpl_name(m *model.Method) string {
	return fmt.Sprintf("%v%vClientChannel", toLowerCameCase(m.Service.Def.Name), toUpperCamelCase(m.Name))
}
