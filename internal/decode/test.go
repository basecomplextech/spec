// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package decode

import "github.com/basecomplextech/baselibrary/encoding/compactint"

// appendSize appends size as compactint, for tests.
func appendSize(b []byte, big bool, size uint32) []byte {
	p := [compactint.MaxLen32]byte{}
	n := compactint.PutReverseUint32(p[:], size)
	off := compactint.MaxLen32 - n

	return append(b, p[off:]...)
}
