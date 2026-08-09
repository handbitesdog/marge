// Command marge is a static-site generator with two subcommands: build
// (one-shot) and serve (build, serve dist/ over HTTP, rebuild on change). By
// convention a project's source lives in src/ and builds to dist/; both
// subcommands default to those paths but accept explicit ones instead.
package main

import (
	"fmt"
	"os"

	"github.com/handbitesdog/marge/internal/build"
	"github.com/handbitesdog/marge/internal/serve"
)

const (
	defaultSrcDir  = "src"
	defaultDistDir = "dist"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		src, dist, ok := buildArgs(os.Args[2:])
		if !ok {
			fmt.Fprintln(os.Stderr, "usage: marge build [<src> <dist>]")
			os.Exit(1)
		}
		if err := build.Run(build.Options{SrcDir: src, DistDir: dist}); err != nil {
			fmt.Fprintln(os.Stderr, "marge build:", err)
			os.Exit(1)
		}

	case "serve":
		src, dist, addr, ok := serveArgs(os.Args[2:])
		if !ok {
			fmt.Fprintln(os.Stderr, "usage: marge serve [<src> <dist>] <addr>")
			os.Exit(1)
		}
		if err := serve.Run(src, dist, addr); err != nil {
			fmt.Fprintln(os.Stderr, "marge serve:", err)
			os.Exit(1)
		}

	default:
		usage()
		os.Exit(1)
	}
}

// buildArgs resolves marge build's positional arguments: none (src and dist
// default to src/ and dist/) or both given explicitly.
func buildArgs(args []string) (src, dist string, ok bool) {
	switch len(args) {
	case 0:
		return defaultSrcDir, defaultDistDir, true
	case 2:
		return args[0], args[1], true
	default:
		return "", "", false
	}
}

// serveArgs resolves marge serve's positional arguments: just <addr> (src
// and dist default to src/ and dist/), or <src> <dist> <addr> explicitly.
func serveArgs(args []string) (src, dist, addr string, ok bool) {
	switch len(args) {
	case 1:
		return defaultSrcDir, defaultDistDir, args[0], true
	case 3:
		return args[0], args[1], args[2], true
	default:
		return "", "", "", false
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: marge build [<src> <dist>]")
	fmt.Fprintln(os.Stderr, "       marge serve [<src> <dist>] <addr>")
}
