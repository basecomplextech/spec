package golang

import "github.com/basecomplextech/spec/internal/lang/model"

const (
	OptionPackage = "go_package"
)

type Package struct {
	Context *Context

	ID   string // id is "my/example/test", overriden by go_package option
	Name string // name is "test" in "my/example/test"

	Files []*File
	Defs  map[string]Definition
}

func newPackage(mp *model.Package) (*Package, error) {
	p := &Package{
		ID:   mp.ID,
		Name: mp.Name,
	}

	// Optional package id
	opt, ok := mp.OptionNames[OptionPackage]
	if ok {
		p.ID = opt.Value
	}

	// Add files
	if err := p.addFiles(mp); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Package) GetType(mtyp *model.Type) (Type, error) {
	if mtyp.Kind == model.KindReference {
		if mtyp.ImportName != "" {
			return p.Context.GetType(mtyp)
		}
	}

	switch mtyp.Kind {
	case model.KindEnum,
		model.KindMessage,
		model.KindStruct,
		model.KindService,
		model.KindReference:
		return p.GetDefinition(mtyp)
	default:
		return p.Context.GetType(mtyp)
	}

	name := mtyp.Name
	typ, ok := p.Types[name]
	if ok {
		return typ, nil
	}

	return nil, nil
}

func (p *Package) GetDefinition(mdef *model.Definition) (Definition, error) {
	switch mdef.Type {
	case model.DefinitionEnum:
	case model.DefinitionMessage:
	case model.DefinitionStruct:
	case model.DefinitionService:
	}
	return nil, nil
}

// parse

func (p *Package) addFiles(mp *model.Package) error {
	for _, mf := range mp.Files {
		if err := p.addFile(mf); err != nil {
			return err
		}
	}
	return nil
}

func (p *Package) addFile(mf *model.File) error {
	file, err := newFile(p, mf)
	if err != nil {
		return err
	}

	p.Files = append(p.Files, file)
	return nil
}

// types
