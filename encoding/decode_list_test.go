// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec/internal/core"
)

func testEncodeListTable(t tests.T, dataSize int, elements []core.ListElement) []byte {
	buf := buffer.New()
	buf.Grow(dataSize)

	_, err := EncodeListTable(buf, dataSize, elements)
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
