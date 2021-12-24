package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFile(t *testing.T) string {
	b, err := os.ReadFile("testdata/test.spec")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestParse_Parse__should_parse_file(t *testing.T) {
	p := newParser()
	s := testFile(t)

	file, err := p.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	require.NotNil(t, file)
	assert.Equal(t, "test", file.Module)
	t.Fatal()
}
