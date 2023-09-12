package model

import "strings"

func toUpperCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		part = strings.ToLower(part)
		part = strings.Title(part)
		parts[i] = part
	}
	return strings.Join(parts, "")
}
