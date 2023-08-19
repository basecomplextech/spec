package main

import (
	"log"
	"strings"

	"github.com/basecomplextech/spec/internal/lang"
	"github.com/spf13/cobra"
)

func main() {
	var importPath []string

	generateGo := &cobra.Command{
		Use:   "generate-go [-i import-paths] [src-dir] [dst-dir]",
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

			spec := lang.New()
			if err := spec.GenerateGo(srcPath, dstPath, importPath); err != nil {
				log.Fatal(err)
			}
		},
	}
	generateGo.Flags().StringArrayVarP(&importPath, "import", "i", nil, "import paths")
	generateGo.MarkFlagRequired("out")

	var cmd = &cobra.Command{Use: "spec"}
	cmd.AddCommand(generateGo)
	cmd.Execute()
}
