package generator

import (
	"github.com/basecomplextech/spec/internal/lang/model"
)

func (w *writer) client(def *model.Definition) error {
	if err := w.clientDef(def); err != nil {
		return err
	}
	if err := w.clientMethods(def); err != nil {
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
		// TODO: Remove me
		if m.Sub {
			continue
		}

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

	if err := w.clientMethod_request(def, m); err != nil {
		return err
	}

	if m.Sub {
		if err := w.clientMethod_sub(def, m); err != nil {
			return err
		}
	} else {
		if err := w.clientMethod_send(def, m); err != nil {
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
	if m.Input == nil {
		w.write(`(cancel <-chan struct{}) `)
	} else {
		typeName := typeName(m.Input)
		w.writef(`(cancel <-chan struct{}, req_ %v) `, typeName)
	}
	return nil
}

func (w *writer) clientMethod_output(def *model.Definition, m *model.Method) error {
	if m.Output == nil {
		w.write(`(status.Status)`)
	} else {
		typeName := typeName(m.Output)
		if m.Sub {
			w.line(`(`)
			w.writef(`%v, status.Status)`, typeName)
		} else {
			w.line(`(`)
			w.writef(`*ref.R[%v], status.Status)`, typeName)
		}
	}
	return nil
}

func (w *writer) clientMethod_request(def *model.Definition, m *model.Method) error {
	// Build request
	w.line(`// Build request`)
	w.line(`req := rpc.NewRequest()`)
	w.line(`defer req.Free()`)
	w.line()

	// Add call
	if m.Input == nil {
		w.linef(`st := req.Add("%v")`, m.Name)
	} else {
		w.linef(`st := req.AddInput("%v", req_.Unwrap())`, m.Name)
	}
	w.line(`if !st.OK() {`)
	w.linef(`return %v st`, clientMethod_zeroReturn(m))
	w.line(`}`)

	// End request
	w.line()
	return nil
}

func (w *writer) clientMethod_sub(def *model.Definition, m *model.Method) error {
	// Return subservice
	typeName := typeRefName(m.Output)

	w.line(`// Return subservice`)
	w.linef(`sub := New%vClient(c.client, req)`, typeName)
	w.line(`ok = true`)
	w.linef(`return sub, status.OK`)
	return nil
}

func (w *writer) clientMethod_send(def *model.Definition, m *model.Method) error {
	// Build request
	w.line(`preq, st := req.Build()`)
	w.line(`if !st.OK() {`)
	w.linef(`return %v st`, clientMethod_zeroReturn(m))
	w.line(`}`)
	w.line()

	// Send request
	w.line(`// Send request`)
	w.line(`value, st := c.client.Request(cancel, preq)`)
	w.line(`if !st.OK() {`)
	w.linef(`return %v st`, clientMethod_zeroReturn(m))
	w.line(`}`)
	w.line(`defer value.Release()`)
	w.line(``)
	return nil
}

func (w *writer) clientMethod_response(def *model.Definition, m *model.Method) error {
	if m.Output == nil {
		w.line(`return status.OK`)
		return nil
	}

	parseFunc := typeParseFunc(m.Output)
	w.line(`// Parse result`)
	w.linef(`result, _, err := %v(value.Unwrap())`, parseFunc)
	w.line(`if err != nil {`)
	w.linef(`return %v status.WrapError(err)`, clientMethod_zeroReturn(m))
	w.line(`}`)
	w.line(`return ref.NewParentRetain(result, value), status.OK`)
	return nil
}

// util

func clientMethod_zeroReturn(m *model.Method) string {
	if m.Output == nil {
		return ``
	} else {
		return `nil, `
	}
}
