package spec

import "testing"

func BenchmarkBuffer_AllocWrite(b *testing.B) {
	buf := NewBuffer(make([]byte, 0, 1024))
	hello := []byte("hello, world")

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(hello)))

	for i := 0; i < b.N; i++ {
		p := buf.Alloc(len(hello))
		copy(p, hello)

		_, err := buf.Write(p)
		if err != nil {
			b.Fatal(err)
		}

		(buf.(*buffer)).Reset()
	}
}
