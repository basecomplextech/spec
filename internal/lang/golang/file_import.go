package golang

import "github.com/basecomplextech/spec/internal/lang/model"

type Import struct {
	ID string
}

func newImport(m *model.Import) (*Import, error) {
	id := m.Package.ID

	opt, ok := m.Package.OptionNames[OptionPackage]
	if ok {
		id = opt.Value
	}

	imp := &Import{ID: id}
	return imp, nil
}
