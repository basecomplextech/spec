// Copyright 2023 Ivan Korobkov. All rights reserved.

package mpx

import "fmt"

const debug = false

func debugPrint(client bool, a ...any) {
	if !debug {
		return
	}

	prefix := " "
	if client {
		prefix = "c"
	}

	args := make([]any, 0, 1+len(a))
	args = append(args, prefix)
	args = append(args, a...)

	fmt.Println(args...)
}

func debugString(b []byte) string {
	switch {
	case b == nil:
		return "nil"
	case len(b) <= 64:
		return string(b)
	}
	return string(b[:64]) + "..."
}
