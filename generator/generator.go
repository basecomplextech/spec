package generator

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/baseone-run/spec/compiler"
)

type Generator interface {
	// GenerateGo generates a go package.
	GenerateGo(pkg *compiler.Package) error
}

type Options struct {
	OutPath string
}

// New returns a new generator.
func New(opts Options) Generator {
	return newGenerator(opts)
}

type generator struct {
	opts Options
}

func newGenerator(opts Options) *generator {
	return &generator{opts: opts}
}

func (g *generator) generate(path string, template *template.Template, data interface{}) error {
	f, err := g.createFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return template.Execute(f, data)
}

func (g *generator) createFile(filename string) (*os.File, error) {
	path := filepath.Join(g.opts.OutPath, filename)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}

	return os.Create(path)
}
