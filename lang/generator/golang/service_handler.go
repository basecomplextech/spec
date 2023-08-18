package golang

import "github.com/basecomplextech/spec/lang/compiler"

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
	w.writef(`func (h *%vHandler) Handle(cancel <-chan struct{},`, def.Name)
	w.line(`
	req *rpc.ServerRequest,
	resp rpc.ServerResponse,
) status.Status {
	index := 0

	call, st := req.Call(index)
	if !st.OK() {
		return st
	}

	method := call.Method()
	switch call.Method() {`)

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
	return nil
}
