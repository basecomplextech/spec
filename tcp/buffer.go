package tcp

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
)

var bufferPool = &sync.Pool{}

func acquireBuffer() *alloc.Buffer {
	b := bufferPool.Get()
	if b == nil {
		return alloc.NewBuffer()
	}
	return b.(*alloc.Buffer)
}

func releaseBuffer(b *alloc.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}
