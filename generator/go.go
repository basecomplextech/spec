package generator

import (
	"fmt"
	"text/template"

	"github.com/baseone-run/spec/compiler"
)

// GenerateGo generates a go package.
func (g *generator) GenerateGo1(pkg *compiler.Package) error {
	t := template.New("file").Funcs(goFuncs)

	t, err := t.Parse(goFile)
	if err != nil {
		return err
	}
	t, err = t.Parse(goEnum)
	if err != nil {
		return err
	}
	t, err = t.Parse(goMessage)
	if err != nil {
		return err
	}
	t, err = t.Parse(goStruct)
	if err != nil {
		return err
	}

	for _, file := range pkg.Files {
		path := filenameWithExt(file.Name, "go")

		if err := g.generate(path, t, file); err != nil {
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

const goFile = `package {{ .Package | packageName }}

{{ if .Imports -}}
import (
	{{- range .Imports }}
	"{{ . | importName }}"
	{{- end }}
	"github.com/baseone-run/spec"
)
{{- end }}

{{ range .Definitions }}
	{{ if eq .Type "enum" }}
		{{ block "enum" . }}{{ end }}
	{{ else if eq .Type "message" }}
		{{ block "message" . }}{{ end }}
	{{ else if eq .Type "struct" }}
		{{ block "struct" . }}{{ end }}
	{{ else }}
		Invalid type {{ .Type }}
	{{ end }}
{{ end }}
`

const goEnum = `{{ define "enum" }}
{{- with $name := .Name }}
type {{ $.Name }} int32

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
`

const goMessage = `{{ define "message" }}
type {{ .Name }} struct {
{{- range .Message.Fields }}
	{{ . | messageFieldName }} {{ .Type | typeName }} {{ . | messageFieldTag }}
{{- end}}
}

func (m {{ .Name }}) Write(w spec.Writer) error {
	if err := w.BeginMessage(); err != nil {
		return err
	}

{{ range .Message.Fields }}
	{{- $name := . | messageFieldName }}
	// {{ $name }} {{ .Tag }}

	{{- if .Type.Bool }}
	{{- template "messageWriteField" . }}
	w.Field({{ .Tag }})

	{{- else if .Type.Number }}
	if m.{{ $name }} != 0 {
		{{- template "messageWriteField" . }}
		w.Field({{ .Tag }})
	}

	{{- else if .Type.Bytes }}
	if len(m.{{ $name }}) > 0 {
		{{- template "messageWriteField" . }}
		w.Field({{ .Tag }})
	}

	{{- else if .Type.String }}
	if m.{{ $name }} != "" {
		{{- template "messageWriteField" . }}
		w.Field({{ .Tag }})
	}

	{{- else if .Type.Nullable }}
	if m.{{ $name }} != nil {
		{{- template "messageWriteField" . }}
		w.Field({{ .Tag }})
	}

	{{- else if .Type.List }}
	if len(m.{{ $name }}) > 0 {
		{{- template "messageWriteField" . }}
		w.Field({{ .Tag }})
	}

	{{- else if .Type.Enum }}
	if m.{{ $name }} != 0 {
		{{- template "messageWriteField" . }}
		w.Field({{ .Tag }})
	}

	{{- else if .Type.Message }}
		{{- template "messageWriteField" . }}
		w.Field({{ .Tag }})

	{{- else if .Type.Struct }}
		{{- template "messageWriteField" . }}
		w.Field({{ .Tag }})
	{{ end }}
{{ end }}

	return w.EndMessage()
}

{{ end }}

{{- define "messageWriteField" }}
{{- $name := . | messageFieldName }}
{{- $access := . | messageFieldName }}

{{- if .Type.Bool }}
	w.Bool(m.{{ $name }})
{{- else if .Type.Int8 }}
	w.Int8(m.{{ $name }})
{{- else if .Type.Int16 }}
	w.Int16(m.{{ $name }})
{{- else if .Type.Int32 }}
	w.Int32(m.{{ $name }})
{{- else if .Type.Int64 }}
	w.Int64(m.{{ $name }})

{{- else if .Type.Uint8 }}
	w.Uint8(m.{{ $name }})
{{- else if .Type.Uint16 }}
	w.Uint16(m.{{ $name }})
{{- else if .Type.Uint32 }}
	w.Uint32(m.{{ $name }})
{{- else if .Type.Uint64 }}
	w.Uint64(m.{{ $name }})

{{- else if .Type.Float32 }}
	w.Float32(m.{{ $name }})
{{- else if .Type.Float64 }}
	w.Float64(m.{{ $name }})

{{- else if .Type.Bytes }}
	w.Bytes(m.{{ $name }})
{{- else if .Type.String }}
	w.String(m.{{ $name }})

{{- else if .Type.List }}
	if err := w.BeginList(); err != nil {
		return err
	}
	for v := range m.{{ $name }} {
	}
	if err := w.EndList(); err != nil {
		return err
	}

{{- else if .Type.Enum }}
	w.Int32(int32(m.{{ $name }}))

{{- else if .Type.Message }}
	if err := m.{{ $name }}.Write(w); err != nil {
		return err
	}
{{- else if .Type.Struct }}
	if err := m.{{ $name }}.Write(w); err != nil {
		return err
	}
{{- end }}

{{- end }}
`

const goStruct = `{{ define "struct" }}
type {{ .Name }} struct {
	{{- range .Struct.Fields }}
	{{ . | structFieldName }} {{ .Type | typeName }} {{ . | structFieldTag }}
	{{- end}}
}
{{- end }}
`
