// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package syntax

type Type struct {
	Kind    Kind
	Name    string
	Import  string // package name in imported type, "pkg" in "pkg.Name"
	Element *Type  // element type in list and nullable types
}

func (t *Type) String() string {
	switch t.Kind {
	case KindReference:
		if t.Import != "" {
			return t.Import + "." + t.Name
		}
		return t.Name
	case KindList:
		return "[]" + t.Element.String()
	}
	return t.Kind.String()
}
