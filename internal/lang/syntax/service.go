// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package syntax

type Service struct {
	Sub     bool // Subservice
	Methods []*Method
}

type Method struct {
	Name string

	Input   MethodInput
	Output  MethodOutput
	Channel *MethodChannel
	Oneway  bool
}

// MethodInput is a union type for method inputs.
//
//	MethodInput:
//	| Reference
//	| Fields
//	| nil
type MethodInput interface{}

// MethodOutput is a union type for method outputs.
//
//	MethodOutput:
//	| Reference
//	| Fields
//	| nil
type MethodOutput interface{}

// MethodChannel defines method in/out messages, at lesyntax one field must be set.
type MethodChannel struct {
	In  *Type
	Out *Type
}
