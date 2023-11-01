package generator

import "github.com/basecomplextech/spec/internal/lang/model"

type clientWriter struct {
	*writer
}

func newClientWriter(w *writer) *clientWriter {
	return &clientWriter{w}
}

func (w *clientWriter) client(def *model.Definition) error {
	return nil
}
