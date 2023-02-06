package writer

// MessageWriter writes a message.
type MessageWriter struct {
	w *Writer
}

func newMessageWriter(w *Writer) MessageWriter {
	m := MessageWriter{w: w}
	m.w.BeginMessage()
	return m
}

// Field returns a field writer.
func (m MessageWriter) Field(field uint16) FieldWriter {
	return newField(m.w, field)
}

// Build ends the message and returns its bytes.
func (m MessageWriter) Build() ([]byte, error) {
	return m.w.End()
}

// End ends the message.
func (m MessageWriter) End() error {
	_, err := m.w.End()
	return err
}
