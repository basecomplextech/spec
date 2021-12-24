package parser

import (
	"io"
	"os"
	"strings"
)

type Parser interface {
	Parse(s string) (*File, error)
	ParseFile(f *os.File) (*File, error)
}

// New returns a new reusable parser.
func New() Parser {
	return newParser()
}

type parser struct{}

func newParser() *parser {
	return &parser{}
}

func (p *parser) Parse(s string) (*File, error) {
	src := strings.NewReader(s)
	return p.parse("", src)
}

func (p *parser) ParseFile(f *os.File) (*File, error) {
	name := f.Name()
	return p.parse(name, f)
}

// private

func (p *parser) parse(filename string, src io.Reader) (*File, error) {
	lexer := newLexer(filename, src)
	parser := yyNewParser()
	parser.Parse(lexer)
	return lexer.file, lexer.err
}
