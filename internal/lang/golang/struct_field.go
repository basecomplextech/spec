package golang

import "github.com/basecomplextech/spec/internal/lang/model"

type StructField struct {
	Package *Package
	Struct  *Struct

	Name string
	Type Type
}

func newStructField(pkg *Package, s *Struct, mf *model.StructField) (*StructField, error) {
	name := name_upperCamelCase(mf.Name)

	typ, err := pkg.GetType(mf.Type)
	if err != nil {
		return nil, err
	}

	f := &StructField{
		Package: pkg,
		Struct:  s,

		Name: name,
		Type: typ,
	}
	return f, nil
}
