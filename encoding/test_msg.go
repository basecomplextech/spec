// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"math"

	"github.com/basecomplextech/spec/internal/core"
)

func TestFields() []core.MessageField {
	return TestFieldsSizeN(false, 10)
}

func TestFieldsN(n int) []core.MessageField {
	return TestFieldsSizeN(false, n)
}

func TestFieldsSize(big bool) []core.MessageField {
	return TestFieldsSizeN(big, 10)
}

func TestFieldsSizeN(big bool, n int) []core.MessageField {
	tagStart := uint16(0)
	offStart := uint32(0)
	if big {
		tagStart = math.MaxUint8 + 1
		offStart = math.MaxUint16 + 1
	}

	result := make([]core.MessageField, 0, n)
	for i := 0; i < n; i++ {
		field := core.MessageField{
			Tag:    tagStart + uint16(i+1),
			Offset: offStart + uint32(i*10),
		}
		result = append(result, field)
	}
	return result
}
