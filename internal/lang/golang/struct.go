package golang

import "github.com/basecomplextech/spec/internal/lang/model"

type Struct struct {
	Package *Package

	Name   string
	Fields []*StructField
}

func newStruct(pkg *Package, mdef *model.Definition) (*Struct, error) {
	s := &Struct{
		Package: pkg,
		Name:    mdef.Name,
	}

	for _, mf := range mdef.Struct.Fields.Values() {
		f, err := newStructField(pkg, s, mf)
		if err != nil {
			return nil, err
		}
		s.Fields = append(s.Fields, f)
	}
	return s, nil
}
