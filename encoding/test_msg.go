// Copyright 2023 Ivan Korobkov. All rights reserved.

package encoding

import "math"

func TestFields() []MessageField {
	return TestFieldsSizeN(false, 10)
}

func TestFieldsN(n int) []MessageField {
	return TestFieldsSizeN(false, n)
}

func TestFieldsSize(big bool) []MessageField {
	return TestFieldsSizeN(big, 10)
}

func TestFieldsSizeN(big bool, n int) []MessageField {
	tagStart := uint16(0)
	offStart := uint32(0)
	if big {
		tagStart = math.MaxUint8 + 1
		offStart = math.MaxUint16 + 1
	}

	result := make([]MessageField, 0, n)
	for i := 0; i < n; i++ {
		field := MessageField{
			Tag:    tagStart + uint16(i+1),
			Offset: offStart + uint32(i*10),
		}
		result = append(result, field)
	}
	return result
}
