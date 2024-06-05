package mpx

import (
	"testing"

	"github.com/basecomplextech/baselibrary/async"
)

func testChannelSend(t *testing.T, ctx async.Context, ch Channel, msg string) {
	st := ch.Send(ctx, []byte(msg))
	if !st.OK() {
		t.Fatal(st)
	}
}
