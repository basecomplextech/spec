package golang

import "github.com/basecomplextech/spec/internal/lang/model"

type File struct {
	Package *Package
	Name    string

	Imports     []*Import
	Definitions []Definition
}

func newFile(p *Package, mf *model.File) (*File, error) {
	f := &File{
		Package: p,
		Name:    mf.Name,
	}

	if err := f.addImports(mf); err != nil {
		return nil, err
	}
	if err := f.addDefinitions(mf); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *File) addImports(mf *model.File) error {
	for _, mi := range mf.Imports {
		imp, err := newImport(mi)
		if err != nil {
			return err
		}

		f.Imports = append(f.Imports, imp)
	}
	return nil
}

func (f *File) addDefinitions(mf *model.File) error {
	for _, mdef := range mf.Definitions {
		def, err := newDefinition(f.Package, mdef)
		if err != nil {
			return err
		}

		f.Definitions = append(f.Definitions, def)
	}
	return nil
}
