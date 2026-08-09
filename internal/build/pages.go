// Package build discovers pages under pages/, computes their pretty output
// paths and URLs, and copies static assets into dist/.
package build

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/handbitesdog/marge/internal/site"
)

// Discover walks pagesDir recursively and separates ordinary routed pages
// from per-collection item templates. Any file whose basename starts with
// "_" is excluded from ordinary routing; a file named "_item.html" is
// specifically recognized as the item template for the collection named
// after its parent directory.
func Discover(pagesDir string) (pages []site.Page, itemTemplates map[string]string, err error) {
	itemTemplates = make(map[string]string)

	err = filepath.WalkDir(pagesDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".html") {
			return nil
		}

		rel, err := filepath.Rel(pagesDir, p)
		if err != nil {
			return err
		}

		if strings.HasPrefix(d.Name(), "_") {
			if d.Name() == "_item.html" {
				itemTemplates[filepath.Base(filepath.Dir(rel))] = rel
			}
			return nil
		}

		outputPath, url := prettyURL(rel)
		pages = append(pages, site.Page{
			SourcePath: rel,
			OutputPath: outputPath,
			URL:        url,
		})
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("discover pages in %q: %w", pagesDir, err)
	}

	return pages, itemTemplates, nil
}

// prettyURL computes a page's output path within dist/ and its route URL
// from its path relative to pages/. An index.html file maps to its own
// directory; any other name gets a directory of its own so links stay
// extension-free (about.html -> /about/).
func prettyURL(relPath string) (outputPath, url string) {
	relPath = filepath.ToSlash(relPath)
	dir := path.Dir(relPath)
	if dir == "." {
		dir = ""
	}
	base := strings.TrimSuffix(path.Base(relPath), ".html")

	outDir := dir
	if base != "index" {
		if dir == "" {
			outDir = base
		} else {
			outDir = dir + "/" + base
		}
	}

	outputPath = filepath.FromSlash(path.Join(outDir, "index.html"))
	if outDir == "" {
		url = "/"
	} else {
		url = "/" + outDir + "/"
	}
	return outputPath, url
}
