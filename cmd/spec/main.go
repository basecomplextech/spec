package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/basecomplextech/spec/internal/lang"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "spec",
		Usage: "Spec code generator",
		Commands: []*cli.Command{
			{
				Name:        "generate",
				Description: "Generate a Go package from a Spec package",
				UsageText:   "spec generate [-i import-paths] [--skip-rpc] [src-dir] [dst-dir]",
				Args:        true,
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "import",
						Aliases: []string{"i"},
						Usage:   "import paths",
					},
					&cli.BoolFlag{
						Name:  "skip-rpc",
						Usage: "skip generating RPC code",
					},
				},
				Action: func(x *cli.Context) error {
					// Source/dest args
					src := ""
					dst := ""

					args := x.Args().Slice()
					switch len(args) {
					case 0:
						src = "."
					case 1:
						src = strings.TrimSpace(x.Args().Get(0))
					case 2:
						src = strings.TrimSpace(x.Args().Get(0))
						dst = strings.TrimSpace(x.Args().Get(1))
					default:
						return fmt.Errorf("invalid src/dst args: %v", args)
					}

					// Flags
					imports := x.StringSlice("import")
					skipRPC := x.Bool("skip-rpc")

					// Generate
					spec := lang.New(imports, skipRPC)
					return spec.Generate(src, dst)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
