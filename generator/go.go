package generator

import (
	"fmt"
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
	"packageName": goPackageName,
	"importName":  goImportName,
	"typeName":    goTypeName,

	"enumValueName":   goEnumValueName,
	"enumValueString": goEnumValueString,

	"messageFieldName": goMessageFieldName,
	"messageFieldTag":  goMessageFieldTag,

	"structFieldName": goStructFieldName,
	"structFieldTag":  goStructFieldTag,
}

func goPackageName(pkg *compiler.Package) string {
	return pkg.Name
}

func goImportName(imp *compiler.Import) string {
	id := imp.ID
	return id
}

func goEnumValueName(val *compiler.EnumValue) string {
	name := toUpperCamelCase(val.Name)
	return val.Enum.Def.Name + name
}

func goEnumValueString(val *compiler.EnumValue) string {
	return toLowerCase(val.Name)
}

func goMessageFieldName(field *compiler.MessageField) string {
	return toUpperCamelCase(field.Name)
}

func goMessageFieldTag(field *compiler.MessageField) string {
	return fmt.Sprintf("`tag:\"%d\" json:\"%v\"`", field.Tag, field.Name)
}

func goStructFieldName(field *compiler.StructField) string {
	return toUpperCamelCase(field.Name)
}

func goStructFieldTag(field *compiler.StructField) string {
	return fmt.Sprintf("`json:\"%v\"`", field.Name)
}

func goTypeName(t *compiler.Type) string {
	switch t.Kind {
	case compiler.KindBool:
		return "bool"

	case compiler.KindInt8:
		return "int8"
	case compiler.KindInt16:
		return "int16"
	case compiler.KindInt32:
		return "int32"
	case compiler.KindInt64:
		return "int64"

	case compiler.KindUint8:
		return "uint8"
	case compiler.KindUint16:
		return "uint16"
	case compiler.KindUint32:
		return "uint32"
	case compiler.KindUint64:
		return "uint64"

	case compiler.KindFloat32:
		return "float32"
	case compiler.KindFloat64:
		return "float64"

	case compiler.KindBytes:
		return "[]byte"
	case compiler.KindString:
		return "string"

	// references

	case compiler.KindReference:
		return t.Name
	case compiler.KindImport:
		return t.ImportName + "." + t.Name
	case compiler.KindList:
		elem := goTypeName(t.Element)
		return "[]" + elem
	case compiler.KindNullable:
		elem := goTypeName(t.Element)
		return "*" + elem
	}

	return ""
}

const goTemplate = `package {{ .Package | packageName }}

{{ if .Imports -}}
import (
	{{- range .Imports }}
	"{{ . | importName }}"
	{{- end }}
)
{{- end }}

{{ range .Definitions -}}
	{{ template "def" . }}
{{ end }}

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
	{{- range .Message.Fields }}
	{{ . | messageFieldName }} {{ .Type | typeName }} {{ . | messageFieldTag }}
	{{- end}}
}
{{- end }}

{{- define "struct" }}
type {{ .Name }} struct {
	{{- range .Struct.Fields }}
	{{ . | structFieldName }} {{ .Type | typeName }} {{ . | structFieldTag }}
	{{- end}}
}
{{- end }}
`
