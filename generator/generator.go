package generator

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/complexl/spec/compiler"
	"github.com/complexl/spec/generator/golang"
)

type Generator interface {
	// Golang generates a go package.
	Golang(pkg *compiler.Package, out string) error
}

// New returns a new generator.
func New() Generator {
	return newGenerator()
}

type generator struct{}

func newGenerator() *generator {
	return &generator{}
}

// Golang generates a go package.
func (g *generator) Golang(pkg *compiler.Package, out string) error {
	for _, file := range pkg.Files {
		if err := g.golangFile(file, out); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) golangFile(file *compiler.File, out string) error {
	bytes, err := golang.WriteFile(file)
	if err != nil {
		return err
	}

	filename := filenameWithoutExt(file.Name) + "_generated.go"
	path := filepath.Join(out, filename)

	f, err := g.createFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(bytes); err != nil {
		return err
	}
	return f.Sync()
}

// private

func (g *generator) generate(path string, template *template.Template, data interface{}) error {
	f, err := g.createFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return template.Execute(f, data)
}

func (g *generator) createFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}
	return os.Create(path)
}

// filenameWithoutExt returns a filename without an extension.
func filenameWithoutExt(name string) string {
	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)]
	return strings.TrimRight(name, ".")
}
