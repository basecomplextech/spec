// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package mpx

import (
	"testing"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/baselibrary/bin"
)

func BenchmarkMessageBuild(b *testing.B) {
	buf := alloc.NewBuffer()
	input := messageInput{
		id:     bin.Random128(),
		data:   make([]byte, 128),
		window: 16 * 1024 * 1024,
		open:   true,
	}

	for i := 0; i < b.N; i++ {
		buf.Reset()

		msg, err := newBuilder().buildMessage(buf, input)
		if err != nil {
			b.Fatal(err)
		}
		_ = msg
	}

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec

	b.ReportMetric(ops/1000_000, "mops")
}
