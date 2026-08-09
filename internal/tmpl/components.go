package tmpl

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// LoadComponents reads each <name>.html file directly under dir (flat,
// non-recursive — subdirectories are ignored), preprocesses it, wraps it as
// {{define "<name>"}}{{$children := .children}}...{{end}}, and parses it
// into root. The leading {{$children := .children}} makes a block-form
// invocation's captured body available as the stable variable $children,
// regardless of any {{range}}/{{with}} later in the component that changes
// what "." refers to. Component names never collide with page names since
// pages are registered under their full relative path including ".html".
//
// A missing dir is not an error — LoadComponents simply defines nothing.
//
// seen tracks each defined name back to the directory it came from, so a
// name reused across multiple LoadComponents calls against the same root
// (e.g. once for components/, once for layouts/) is a build-time error
// instead of a silent last-definition-wins override. Pass a fresh map on the
// first call and reuse it across subsequent calls against the same root.
func LoadComponents(root *template.Template, dir string, seen map[string]string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read components dir %q: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		src, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read component %q: %w", path, err)
		}

		name := strings.TrimSuffix(entry.Name(), ".html")
		if !startsUpper(name) {
			return fmt.Errorf("component %q: component/layout names must start with an uppercase letter", path)
		}
		if prevDir, ok := seen[name]; ok {
			return fmt.Errorf("component %q is defined in both %q and %q", name, prevDir, dir)
		}
		seen[name] = dir

		body, err := Preprocess(string(src), path)
		if err != nil {
			return err
		}

		defined := fmt.Sprintf(`{{define %q}}{{$children := .children}}%s{{end}}`, name, body)
		if _, err := root.Parse(defined); err != nil {
			return fmt.Errorf("parse component %q: %w", path, err)
		}
	}

	return nil
}
