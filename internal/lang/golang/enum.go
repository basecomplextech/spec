package golang

import "github.com/basecomplextech/spec/internal/lang/model"

type Enum struct {
	Package *Package
	Name    string
	Values  []*EnumValue
}

func newEnum(pkg *Package, mdef *model.Definition) (*Enum, error) {
	e := &Enum{
		Package: pkg,
		Name:    mdef.Name,
	}

	for _, mval := range mdef.Enum.Values {
		val, err := newEnumValue(mval)
		if err != nil {
			return nil, err
		}
		e.Values = append(e.Values, val)
	}
	return e, nil
}
