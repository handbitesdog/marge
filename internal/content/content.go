package content

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/handbitesdog/marge/internal/site"
)

// LoadAll walks one level of subdirectories under contentDir. Each
// subdirectory becomes a Collection named after itself, and each .md file
// within it becomes a CollectionItem: front matter is parsed and the
// remaining body is rendered as Markdown (raw HTML is passed through
// unescaped, since content authors are trusted). Items within a collection
// are sorted by Date descending.
//
// A missing contentDir is not an error — LoadAll simply returns no
// collections.
func LoadAll(contentDir string) (map[string]site.Collection, error) {
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]site.Collection{}, nil
		}
		return nil, fmt.Errorf("read content dir %q: %w", contentDir, err)
	}

	md := goldmark.New(goldmark.WithRendererOptions(html.WithUnsafe()))

	collections := make(map[string]site.Collection)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		dir := filepath.Join(contentDir, name)
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("read collection dir %q: %w", dir, err)
		}

		var items []site.CollectionItem
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
				continue
			}

			path := filepath.Join(dir, f.Name())
			src, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read %q: %w", path, err)
			}

			fm, body, err := splitFrontMatter(src, path)
			if err != nil {
				return nil, err
			}

			var buf bytes.Buffer
			if err := md.Convert(body, &buf); err != nil {
				return nil, fmt.Errorf("render markdown %q: %w", path, err)
			}

			slug := strings.TrimSuffix(f.Name(), ".md")
			items = append(items, site.CollectionItem{
				Title:   fm.Title,
				Date:    fm.Date,
				Slug:    slug,
				URL:     fmt.Sprintf("/%s/%s/", name, slug),
				Content: template.HTML(buf.String()),
				Extra:   fm.Extra,
			})
		}

		sort.Slice(items, func(i, j int) bool {
			return items[i].Date.After(items[j].Date)
		})

		collections[name] = site.Collection{Name: name, Items: items}
	}

	return collections, nil
}
