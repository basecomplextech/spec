// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package decode

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec/internal/encode"
	"github.com/basecomplextech/spec/internal/format"
)

func testEncodeListTable(t tests.T, dataSize int, elements []format.ListElement) []byte {
	buf := buffer.New()
	buf.Grow(dataSize)

	_, err := encode.EncodeListTable(buf, dataSize, elements)
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
