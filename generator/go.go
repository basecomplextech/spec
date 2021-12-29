package generator

import (
	"text/template"

	"github.com/baseone-run/spec/compiler"
)

// GenerateGo generates a go package.
func (g *generator) GenerateGo(pkg *compiler.Package) error {
	template, err := template.New("file").
		Funcs(goFuncs).
		Parse(goTemplate)
	if err != nil {
		return err
	}

	for _, file := range pkg.Files {
		path := filenameWithExt(file.Name, "go")

		if err := g.generate(path, template, file); err != nil {
			return err
		}
	}
	return nil
}

var goFuncs = map[string]interface{}{
	"packageName":     goPackageName,
	"importName":      goImportName,
	"enumValueName":   goEnumValueName,
	"enumValueString": goEnumValueString,
}

func goPackageName(pkg *compiler.Package) string {
	return pkg.Name
}

func goImportName(imp *compiler.Import) string {
	id := imp.ID
	return id
}

func goEnumValueName(val *compiler.EnumValue) string {
	name := toCamelCase(val.Name)
	return val.Enum.Def.Name + name
}

func goEnumValueString(val *compiler.EnumValue) string {
	return toLowerCase(val.Name)
}

const goTemplate = `package {{ .Package | packageName }}

{{ if .Imports -}}
import (
	{{- range .Imports }}
	"{{ . | importName }}"
	{{- end }}
)
{{- end }}

{{- range .Definitions -}}
	{{- template "def" . }}
{{- end }}

{{- define "def" }}
	{{- if eq .Type "enum" }}
		{{- template "enum" . }}
	{{- else if eq .Type "message" }}
		{{- template "message" . }}
	{{- else if eq .Type "struct" }}
		{{- template "struct" . }}
	{{- else }}
		Invalid type {{ .Type }}
	{{- end }}
{{- end }}

{{- define "enum" }}
{{- with $name := .Name }}
type {{ $.Name }} int

const (
	{{- range $.Enum.Values }}
		{{ . | enumValueName }} {{ $name }} = {{ .Number }}
	{{- end }}
)

func (e {{ $name }}) String() string {
	switch e {
	{{- range $.Enum.Values }}
	case {{ . | enumValueName }}:
		return "{{ . | enumValueString }}"
	{{- end}}
	}
	return ""
}
{{- end }}
{{- end }}

{{- define "message" }}
type {{ .Name }} struct {

}
{{- end }}

{{- define "struct" }}
type {{ .Name }} struct {

}
{{- end }}
`
