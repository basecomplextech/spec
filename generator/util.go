package generator

import (
	"path/filepath"
	"strings"
)

// filenameWithExt returns a filename with another extension.
func filenameWithExt(fileName string, ext string) string {
	name := filenameWithoutExt(fileName)
	return name + "." + ext
}

// filenameWithoutExt returns a filename without an extension.
func filenameWithoutExt(name string) string {
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)]
}

func toLowerCase(s string) string {
	return strings.ToLower(s)
}

func toUpperCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		part = strings.ToLower(part)
		part = strings.Title(part)
		parts[i] = part
	}
	return strings.Join(parts, "")
}
