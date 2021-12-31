package main

import (
	"log"

	"github.com/baseone-run/spec/compiler"
	"github.com/baseone-run/spec/generator"
	"github.com/spf13/cobra"
)

func main() {
	var importPath []string
	generateGo := &cobra.Command{
		Use:   "go [-i import_paths] -o input_dir output_dir",
		Short: "Generates a Go package",
		Long:  `Generates a Go package`,
		Args:  cobra.ExactArgs(2),
		Run: func(_ *cobra.Command, args []string) {
			inputPath := args[0]
			outputPath := args[1]

			compiler, err := compiler.New(compiler.Options{
				ImportPath: importPath,
			})
			if err != nil {
				log.Fatal(err)
			}

			pkg, err := compiler.Compile(inputPath)
			if err != nil {
				log.Fatal(err)
			}

			gen := generator.New()
			if err := gen.Golang(pkg, outputPath); err != nil {
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
