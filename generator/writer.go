package generator

import (
	"bytes"
	"fmt"
)

type writer struct {
	b *bytes.Buffer
}

func newWriter() *writer {
	return &writer{
		b: &bytes.Buffer{},
	}
}

func (w *writer) newline() {
	w.b.WriteString("\n")
}

func (w *writer) line(args ...string) {
	for _, s := range args {
		w.b.WriteString(s)
	}
	w.newline()
}

func (w *writer) linef(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	w.b.WriteString(s)
	w.newline()
}
