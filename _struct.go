package spec

// ReadStruct reads a struct and returns its body reader or nil,
func ReadStruct(b []byte) (Reader, error) {
	bodySize, n, err := readStruct(b)
	switch {
	case err != nil:
		return nil, err
	case n == 0:
		return nil, err
	}

	start := len(b) - n
	end := start + bodySize
	// TODO: Check body size
	r := b[start:end]
	return r, nil
}
