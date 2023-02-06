package tests

import spec "github.com/complex1tech/spec"

func (m *TestSmall) Write(e *spec.Writer) ([]byte, error) {
	w := e.Message()
	w.Field(1).Int64(m.Field0)
	w.Field(2).Int64(m.Field1)
	w.Field(3).Int64(m.Field2)
	w.Field(4).Int64(m.Field3)
	return w.Build()
}

func (m *TestObject) Write(e *spec.Writer) ([]byte, error) {
	w := e.Message()

	w.Field(1).Bool(m.Bool)
	w.Field(2).Byte(m.Byte)

	w.Field(10).Int32(m.Int32)
	w.Field(11).Int64(m.Int64)

	w.Field(20).Uint32(m.Uint32)
	w.Field(21).Uint64(m.Uint64)

	w.Field(24).Bin128(m.Bin128)
	w.Field(25).Bin256(m.Bin256)

	w.Field(30).Float32(m.Float32)
	w.Field(31).Float64(m.Float64)

	w.Field(40).String(m.String)
	w.Field(41).Bytes(m.Bytes)

	if m.Submessage != nil {
		w1 := w.Field(50).Message()
		if _, err := m.Submessage.Write(w1); err != nil {
			return nil, err
		}
	}

	if len(m.List) > 0 {
		w1 := w.Field(51).List()

		for _, v := range m.List {
			w1.Int64(v)
		}

		if err := w1.End(); err != nil {
			return nil, err
		}
	}

	if len(m.Messages) > 0 {
		w1 := w.Field(52).List()

		for _, v := range m.Messages {
			w2 := w1.Message()
			if _, err := v.Write(w2); err != nil {
				return nil, err
			}
		}

		if err := w1.End(); err != nil {
			return nil, err
		}
	}

	if len(m.Strings) > 0 {
		w1 := w.Field(53).List()

		for _, v := range m.Strings {
			w1.String(v)
		}

		if err := w1.End(); err != nil {
			return nil, err
		}
	}

	return w.Build()
}

func (m *TestSubobject) Write(w spec.MessageWriter) ([]byte, error) {
	w.Field(1).Int32(m.Int32)
	w.Field(2).Int64(m.Int64)
	return w.Build()
}

func (m *TestObjectElement) Write(w spec.MessageWriter) ([]byte, error) {
	w.Field(1).Byte(m.Byte)
	w.Field(2).Int32(m.Int32)
	w.Field(3).Int64(m.Int64)
	return w.Build()
}
