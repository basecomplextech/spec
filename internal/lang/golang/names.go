package golang

import "strings"

// name_upperCamelCase returns "HelloWord" from "hello_word".
func name_upperCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		part = strings.ToLower(part)
		part = strings.Title(part)
		parts[i] = part
	}

	s1 := strings.Join(parts, "")
	if strings.HasPrefix(s, "_") {
		s1 = "_" + s1
	}
	if strings.HasSuffix(s, "_") {
		s1 += "_"
	}
	return s1
}

// name_lowerCameCase returns "helloWord" from "HelloWord".
func name_lowerCameCase(s string) string {
	if len(s) == 0 {
		return ""
	}

	s = name_upperCamelCase(s)
	return strings.ToLower(s[:1]) + s[1:]
}

// name_lowerCase returns "hello_word" from "HELLO_WORLD".
func name_lowerCase(s string) string {
	return strings.ToLower(s)
}
