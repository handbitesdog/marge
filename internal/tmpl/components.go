package tmpl

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// LoadComponents reads each components/<name>.html file directly under
// componentsDir (flat, non-recursive — subdirectories are ignored),
// preprocesses it, wraps it as {{define "<name>"}}...{{end}}, and parses it
// into root. Component names never collide with page names since pages are
// registered under their full relative path including ".html".
func LoadComponents(root *template.Template, componentsDir string) error {
	entries, err := os.ReadDir(componentsDir)
	if err != nil {
		return fmt.Errorf("read components dir %q: %w", componentsDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		path := filepath.Join(componentsDir, entry.Name())
		src, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read component %q: %w", path, err)
		}

		name := strings.TrimSuffix(entry.Name(), ".html")
		body, err := Preprocess(string(src), path)
		if err != nil {
			return err
		}

		defined := fmt.Sprintf(`{{define %q}}%s{{end}}`, name, body)
		if _, err := root.Parse(defined); err != nil {
			return fmt.Errorf("parse component %q: %w", path, err)
		}
	}

	return nil
}
