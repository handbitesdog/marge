package build

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/handbitesdog/marge/internal/content"
	"github.com/handbitesdog/marge/internal/site"
	"github.com/handbitesdog/marge/internal/tmpl"
)

// Options configures a Run.
type Options struct {
	SrcDir  string // site source root: components/, layouts/, pages/, content/, static/ — only pages/ is required
	DistDir string // build output directory; cleaned at the start of Run
}

// Run executes marge's full build pipeline: it cleans DistDir, loads
// components, layouts, content, and pages into one shared template set,
// executes every ordinary page and collection item into DistDir, then
// copies SrcDir/static alongside them.
func Run(opts Options) error {
	if err := os.RemoveAll(opts.DistDir); err != nil {
		return fmt.Errorf("clean dist dir %q: %w", opts.DistDir, err)
	}
	if err := os.MkdirAll(opts.DistDir, 0o755); err != nil {
		return fmt.Errorf("create dist dir %q: %w", opts.DistDir, err)
	}

	renderer := &tmpl.Renderer{}
	root := template.New("root").Funcs(renderer.FuncMap())

	seen := map[string]string{}
	if err := tmpl.LoadComponents(root, filepath.Join(opts.SrcDir, "components"), seen); err != nil {
		return err
	}
	if err := tmpl.LoadComponents(root, filepath.Join(opts.SrcDir, "layouts"), seen); err != nil {
		return err
	}

	collections, err := content.LoadAll(filepath.Join(opts.SrcDir, "content"))
	if err != nil {
		return err
	}

	pagesDir := filepath.Join(opts.SrcDir, "pages")
	pages, itemTemplates, err := Discover(pagesDir)
	if err != nil {
		return err
	}
	for name, coll := range collections {
		if len(coll.Items) > 0 {
			if _, ok := itemTemplates[name]; !ok {
				return fmt.Errorf("collection %q has content but no pages/%s/_item.html", name, name)
			}
		}
	}

	for _, p := range pages {
		if err := parseIntoSet(root, pagesDir, p.SourcePath); err != nil {
			return err
		}
	}
	for _, rel := range itemTemplates {
		if err := parseIntoSet(root, pagesDir, rel); err != nil {
			return err
		}
	}

	renderer.SetTemplateSet(root)

	for _, p := range pages {
		data := site.PageData{Collections: collections}
		out := filepath.Join(opts.DistDir, p.OutputPath)
		if err := executeToFile(root, p.SourcePath, data, out); err != nil {
			return err
		}
	}

	for name, rel := range itemTemplates {
		for _, item := range collections[name].Items {
			data := site.ItemData{CollectionItem: item, Collections: collections}
			out := filepath.Join(opts.DistDir, name, item.Slug, "index.html")
			if err := executeToFile(root, rel, data, out); err != nil {
				return err
			}
		}
	}

	return CopyStatic(filepath.Join(opts.SrcDir, "static"), opts.DistDir)
}

// parseIntoSet reads, preprocesses, and parses the page at pagesDir/rel as
// the template named rel (slash-separated), so every page and item template
// can be looked up by source path once all of them share root. It parses
// via root.New(name).Parse rather than wrapping the body in an explicit
// {{define}}: a page using a block-form component has a top-level
// {{define "__block_..."}} in its own preprocessed output (see Preprocess),
// and Go's template parser rejects a {{define}} nested inside another
// {{define}}'s body. New+Parse assigns the leftover body to the named
// template while still registering any such nested defines as siblings in
// the same set.
func parseIntoSet(root *template.Template, pagesDir, rel string) error {
	path := filepath.Join(pagesDir, rel)
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read page %q: %w", path, err)
	}

	body, err := tmpl.Preprocess(string(src), path)
	if err != nil {
		return err
	}

	name := filepath.ToSlash(rel)
	if _, err := root.New(name).Parse(body); err != nil {
		return fmt.Errorf("parse page %q: %w", path, err)
	}
	return nil
}

// executeToFile executes the template registered as name (a page or item
// template's slash-separated source path) against data, writing the result
// to outputPath, creating its parent directory as needed.
func executeToFile(root *template.Template, name string, data any, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create dir for %q: %w", outputPath, err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create %q: %w", outputPath, err)
	}
	defer f.Close()

	if err := root.ExecuteTemplate(f, filepath.ToSlash(name), data); err != nil {
		return fmt.Errorf("execute template %q: %w", name, err)
	}
	return nil
}
