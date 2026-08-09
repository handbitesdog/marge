// Command marge is a static-site generator with two subcommands: build
// (one-shot) and serve (build, serve dist/ over HTTP, rebuild on change).
package main

import (
	"fmt"
	"os"

	"github.com/handbitesdog/marge/internal/build"
	"github.com/handbitesdog/marge/internal/serve"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "usage: marge build <src> <dist>")
			os.Exit(1)
		}
		if err := build.Run(build.Options{SrcDir: os.Args[2], DistDir: os.Args[3]}); err != nil {
			fmt.Fprintln(os.Stderr, "marge build:", err)
			os.Exit(1)
		}

	case "serve":
		if len(os.Args) != 5 {
			fmt.Fprintln(os.Stderr, "usage: marge serve <src> <dist> <addr>")
			os.Exit(1)
		}
		if err := serve.Run(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			fmt.Fprintln(os.Stderr, "marge serve:", err)
			os.Exit(1)
		}

	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: marge build <src> <dist>")
	fmt.Fprintln(os.Stderr, "       marge serve <src> <dist> <addr>")
}
