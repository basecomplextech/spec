package spec

import "testing"

func TestMetrics(t *testing.T) {
	msg := newTestMessage()

	w := NewWriter()
	if err := msg.Write(w); err != nil {
		t.Fatal(err)
	}
	b, err := w.End()
	if err != nil {
		t.Fatal(err)
	}

	_, p, err := computeSizeDistribution(b)
	if err != nil {
		t.Fatal(err)
	}
	if p.size == 0 {
		t.Fatal()
	}

	t.Fatalf("%+v", p)
}
