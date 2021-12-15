package protocol

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestWrite_size__must_be_less_or_equal_2048(t *testing.T) {
	// 2048 is 1/2 of 4kb page or 1/4 of 8kb page.

	w := Writer{}
	size := unsafe.Sizeof(w)

	assert.LessOrEqual(t, int(size), 2048)
}
