package generator

import "github.com/basecomplextech/spec/internal/lang/model"

func (w *writer) handler(def *model.Definition) error {
	if err := w.handlerDef(def); err != nil {
		return err
	}
	// if err := w.handlerHandle(def); err != nil {
	// 	return err
	// }
	// if err := w.handlerMethods(def); err != nil {
	// 	return err
	// }
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
