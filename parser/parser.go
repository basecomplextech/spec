package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Parser interface {
	Parse(s string) (*File, error)
	ParseFile(path string) (*File, error)
	ParseDirectory(path string) ([]*File, error)
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

func (p *parser) ParseFile(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	name := f.Name()
	src := bufio.NewReader(f)
	return p.parse(name, src)
}

func (p *parser) ParseDirectory(path string) ([]*File, error) {
	// check path is directory
	info, err := os.Stat(path)
	switch {
	case err != nil:
		return nil, err
	case !info.IsDir():
		return nil, fmt.Errorf("package not directory, path=%v", path)
	}

	// scan files
	pattern := filepath.Join(path, "*.spec")
	filepaths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// parse files
	files := make([]*File, 0, len(filepaths))
	for _, filepath := range filepaths {
		file, err := p.ParseFile(filepath)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

// private

func (p *parser) parse(filename string, src io.Reader) (*File, error) {
	lexer := newLexer(filename, src)
	parser := yyNewParser()
	parser.Parse(lexer)

	if err := lexer.err; err != nil {
		return nil, err
	}

	file := lexer.file
	file.Path = filename
	return file, nil
}
