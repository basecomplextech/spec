package main

import (
	"log"
	"strings"

	"github.com/basecomplextech/spec/internal/lang"
	"github.com/spf13/cobra"
)

func main() {
	var importPath []string
	var skipRPC bool

	generate := &cobra.Command{
		Use:   "generate [-i import-dirs] [--skip-rpc] [src-dir] [dst-dir]",
		Short: "Generates a Go package",
		Long:  `Generates a Go package`,
		Args:  cobra.MaximumNArgs(2),
		Run: func(_ *cobra.Command, args []string) {
			srcPath := ""
			dstPath := ""

			switch len(args) {
			case 0:
				srcPath = "."
			case 1:
				srcPath = strings.TrimSpace(args[0])
			case 2:
				srcPath = strings.TrimSpace(args[0])
				dstPath = strings.TrimSpace(args[1])
			default:
				log.Fatal("Invalid src/dir paths")
			}

			spec := lang.New(importPath, skipRPC)
			if err := spec.Generate(srcPath, dstPath); err != nil {
				log.Fatal(err)
			}
		},
	}

	generate.Flags().StringArrayVarP(&importPath, "import", "i", nil, "import dirs")
	generate.Flags().BoolVar(&skipRPC, "skip-rpc", false, "skip generating RPC code")
	generate.MarkFlagRequired("out")

	var cmd = &cobra.Command{Use: "spec"}
	cmd.AddCommand(generate)
	cmd.Execute()
}
