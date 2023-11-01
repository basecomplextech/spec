package generator

import (
	"github.com/basecomplextech/spec/internal/lang/model"
)

type fileWriter struct {
	*writer
}

func newFileWriter(w *writer) *fileWriter {
	return &fileWriter{w}
}

func (w *fileWriter) file(file *model.File) error {
	// Package
	w.line("package ", file.Package.Name)
	w.line()

	// Imports
	w.line("import (")
	w.line(`"github.com/basecomplextech/baselibrary/bin"`)
	w.line(`"github.com/basecomplextech/baselibrary/buffer"`)
	w.line(`"github.com/basecomplextech/baselibrary/ref"`)
	w.line(`"github.com/basecomplextech/baselibrary/status"`)
	w.line(`"github.com/basecomplextech/spec"`)
	w.line(`"github.com/basecomplextech/spec/encoding"`)

	if !w.skipRPC {
		w.line(`"github.com/basecomplextech/spec/rpc"`)
		w.line(`"github.com/basecomplextech/spec/proto/prpc"`)
	}

	for _, imp := range file.Imports {
		pkg := importPackage(imp)
		w.linef(`"%v"`, pkg)
	}
	w.line(")")
	w.line()

	// Empty values for imports
	w.line(`var (`)
	w.line(`_ bin.Bin128`)
	w.line(`_ buffer.Buffer`)
	w.line(`_ encoding.Type`)
	w.line(`_ ref.Ref`)

	if !w.skipRPC {
		w.line(`_ rpc.Client`)
		w.line(`_ prpc.Request`)
	}

	w.line(`_ spec.Type`)
	w.line(`_ status.Status`)
	w.line(`)`)

	// Definitions
	return w.definitions(file)
}

func (w *fileWriter) definitions(file *model.File) error {
	// Services
	if !w.skipRPC {
		for _, def := range file.Definitions {
			if def.Type != model.DefinitionService {
				continue
			}

			if err := w.service(def); err != nil {
				return err
			}

		}

		for _, def := range file.Definitions {
			if def.Type != model.DefinitionService {
				continue
			}

			if err := w.client(def); err != nil {
				return err
			}
		}
	}

	// Messages and types
	for _, def := range file.Definitions {
		switch def.Type {
		case model.DefinitionEnum:
			if err := w.enum(def); err != nil {
				return err
			}
		case model.DefinitionMessage:
			if err := w.message(def); err != nil {
				return err
			}
		case model.DefinitionStruct:
			if err := w.struct_(def); err != nil {
				return err
			}
		}
	}

	// Message writers
	for _, def := range file.Definitions {
		if def.Type != model.DefinitionMessage {
			continue
		}
		if err := w.messageWriter(def); err != nil {
			return err
		}
	}

	// Service impls
	if !w.skipRPC {
		for _, def := range file.Definitions {
			if def.Type != model.DefinitionService {
				continue
			}
			if err := w.serviceImpl(def); err != nil {
				return err
			}
		}

		for _, def := range file.Definitions {
			if def.Type != model.DefinitionService {
				continue
			}
			if err := w.clientImpl(def); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *fileWriter) enum(def *model.Definition) error {
	return newEnumWriter(w.writer).enum(def)
}

func (w *fileWriter) message(def *model.Definition) error {
	return newMessageWriter(w.writer).message(def)
}

func (w *fileWriter) messageWriter(def *model.Definition) error {
	return newMessageWriter(w.writer).messageWriter(def)
}

func (w *fileWriter) struct_(def *model.Definition) error {
	return newStructWriter(w.writer).struct_(def)
}

func (w *fileWriter) client(def *model.Definition) error {
	return newClientWriter(w.writer).client(def)
}

func (w *fileWriter) clientImpl(def *model.Definition) error {
	return newClientImplWriter(w.writer).clientImpl(def)
}

func (w *fileWriter) service(def *model.Definition) error {
	return newServiceWriter(w.writer).service(def)
}

func (w *fileWriter) serviceImpl(def *model.Definition) error {
	return newServiceImplWriter(w.writer).serviceImpl(def)
}
