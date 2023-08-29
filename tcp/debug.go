package tcp

import "fmt"

const (
	debug       = false
	debugStream = false
)

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
