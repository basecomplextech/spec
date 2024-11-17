// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"math"

	"github.com/basecomplextech/spec/internal/format"
)

func TestElements() []format.ListElement {
	return TestElementsN(10)
}

func TestElementsN(n int) []format.ListElement {
	return TestElementsSizeN(false, n)
}

func TestElementsSize(big bool) []format.ListElement {
	return TestElementsSizeN(big, 10)
}

func TestElementsSizeN(big bool, n int) []format.ListElement {
	start := uint32(0)
	if big {
		start = math.MaxUint16 + 1
	}

	result := make([]format.ListElement, 0, n)
	for i := 0; i < n; i++ {
		elem := format.ListElement{
			Offset: start + uint32(i*10),
		}
		result = append(result, elem)
	}
	return result
}
