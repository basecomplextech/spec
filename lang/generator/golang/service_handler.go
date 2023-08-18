package golang

import (
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

	// Declare args
	{
		w.line(`// Declare args`)
		w.line(`var (`)
		for _, arg := range m.Args {
			typeName := typeName(arg.Type)
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
			argName := toLowerCameCase(arg.Name)
			w.linef(`case "%v":`, arg.Name)
			w.linef(`_%v, err = arg.Value().Bin128Err()`, argName)
		}
		w.line(`}`)

		w.line(`if err != nil {`)
		w.line(`return status.Newf("rpc_error", "Invalid argument %%q: %%v", name, err)`)
		w.line(`}`)
		w.line(`}`)
		w.line()
	}

	// Call method
	{
		name := toUpperCamelCase(m.Name)
		w.line(`// Call service`)

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
		w.line()
	}

	// Call subservice
	if m.Sub {
		res := m.Results[0]
		resName := toLowerCameCase(res.Name)

		w.line(`// Handle next call`)
		w.linef(`h1 := NewSubserviceHandler(%v_)`, resName)
		w.line(`return h1.Handle(cancel, req, resp, index+1)`)
		w.line(`}`)
		w.line()
		return nil
	}

	// Return
	w.line(`return status.OK`)
	w.line(`}`)
	w.line()
	return nil
}
