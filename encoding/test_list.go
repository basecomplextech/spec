// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import "math"

func TestElements() []ListElement {
	return TestElementsN(10)
}

func TestElementsN(n int) []ListElement {
	return TestElementsSizeN(false, n)
}

func TestElementsSize(big bool) []ListElement {
	return TestElementsSizeN(big, 10)
}

func TestElementsSizeN(big bool, n int) []ListElement {
	start := uint32(0)
	if big {
		start = math.MaxUint16 + 1
	}

	result := make([]ListElement, 0, n)
	for i := 0; i < n; i++ {
		elem := ListElement{
			Offset: start + uint32(i*10),
		}
		result = append(result, elem)
	}
	return result
}
