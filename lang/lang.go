package lang

import (
	"github.com/complexl/spec/lang/compiler"
	"github.com/complexl/spec/lang/generator"
)

type Spec struct{}

func New() *Spec {
	return &Spec{}
}

func (s *Spec) GenerateGo(srcPath string, dstPath string, importPath []string) error {
	if dstPath == "" {
		dstPath = srcPath
	}

	compiler, err := compiler.New(compiler.Options{
		ImportPath: importPath,
	})
	if err != nil {
		return err
	}

	pkg, err := compiler.Compile(srcPath)
	if err != nil {
		return err
	}

	gen := generator.New()
	return gen.Golang(pkg, dstPath)
}
