// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package encoding

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec/internal/core"
)

func testEncodeMessageTable(t tests.T, dataSize int, fields []core.MessageField) []byte {
	buf := buffer.New()
	buf.Grow(dataSize)

	_, err := EncodeMessageTable(buf, dataSize, fields)
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
