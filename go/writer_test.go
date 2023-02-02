package spec

import (
	"testing"

	"github.com/complex1tech/baselibrary/buffer"
	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
	o := newTestObject()

	var data0 []byte
	{
		e := newEncoder(buffer.New())
		b, err := BuildTestMessageEncoder(e)
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Encode(b); err != nil {
			t.Fatal(err)
		}

		data0, err = e.End()
		if err != nil {
			t.Fatal(err)
		}
	}

	var data1 []byte
	{
		buf := buffer.New()
		e := newEncoder(buf)

		var err error
		data1, err = o.Write(e)
		if err != nil {
			t.Fatal(err)
		}
	}

	// assert.Equal(t, data0, data1)

	o0 := &TestObject{}
	if err := o0.Decode(data0); err != nil {
		t.Fatal(err)
	}

	o1 := &TestObject{}
	if err := o1.Decode(data1); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, o0, o1)
}
