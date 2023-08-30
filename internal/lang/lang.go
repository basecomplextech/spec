package lang

import (
	"github.com/basecomplextech/spec/internal/lang/compiler"
	"github.com/basecomplextech/spec/internal/lang/generator"
)

type Spec struct {
	importPath []string
	skipRPC    bool
}

func New(importPath []string, skipRPC bool) *Spec {
	return &Spec{
		importPath: importPath,
		skipRPC:    skipRPC,
	}
}

func (s *Spec) Generate(srcPath string, dstPath string) error {
	if dstPath == "" {
		dstPath = srcPath
	}

	compiler, err := compiler.New(compiler.Options{
		ImportPath: s.importPath,
	})
	if err != nil {
		return err
	}

	pkg, err := compiler.Compile(srcPath)
	if err != nil {
		return err
	}

	gen := generator.New(s.skipRPC)
	return gen.Package(pkg, dstPath)
}
