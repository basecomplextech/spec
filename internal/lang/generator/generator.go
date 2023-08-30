package generator

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/basecomplextech/spec/internal/lang/compiler"
)

type Generator interface {
	// Package generates a go package.
	Package(pkg *compiler.Package, out string) error
}

// New returns a new generator.
func New(skipRPC bool) Generator {
	return newGenerator(skipRPC)
}

type generator struct {
	skipRPC bool
}

func newGenerator(skipRPC bool) *generator {
	return &generator{skipRPC: skipRPC}
}

// Package generates a go package.
func (g *generator) Package(pkg *compiler.Package, out string) error {
	for _, file := range pkg.Files {
		if err := g.file(file, out); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) file(file *compiler.File, out string) error {
	// Generate file
	w := newWriter(g.skipRPC)
	if err := w.file(file); err != nil {
		return err
	}

	// Format file
	bytes := w.b.Bytes()
	bytes, err := format.Source(bytes)
	if err != nil {
		return err
	}

	// Create file
	return g.createFile(file, out, bytes)
}

// private

func (g *generator) createFile(file *compiler.File, out string, bytes []byte) error {
	filename := filenameWithoutExt(file.Name) + "_generated.go"
	path := filepath.Join(out, filename)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(bytes); err != nil {
		return err
	}
	return f.Sync()
}

// filenameWithoutExt returns a filename without an extension.
func filenameWithoutExt(name string) string {
	ext := filepath.Ext(name)
	name = name[:len(name)-len(ext)]
	return strings.TrimRight(name, ".")
}
