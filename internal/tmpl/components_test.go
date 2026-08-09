package tmpl

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestLoadComponents checks that LoadComponents preprocesses and defines
// each components/*.html file (including one that itself uses component
// syntax to reference another), ignores non-.html files, and doesn't
// recurse into subdirectories.
func TestLoadComponents(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "card.html"), `<div class="card">{{.title}}{{.Children}}</div>`)
	writeFile(t, filepath.Join(dir, "layout.html"), `<body>{{<card title="Sidebar"/>}}{{.Children}}</body>`)
	writeFile(t, filepath.Join(dir, "notes.txt"), `not a component`)
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	writeFile(t, filepath.Join(dir, "nested", "ignored.html"), `<p>ignored</p>`)

	r := &Renderer{}
	root := template.New("root").Funcs(r.FuncMap())
	if err := LoadComponents(root, dir); err != nil {
		t.Fatalf("LoadComponents: %v", err)
	}
	r.SetTemplateSet(root)

	if root.Lookup("ignored") != nil {
		t.Fatal("LoadComponents should not recurse into subdirectories")
	}
	if root.Lookup("notes") != nil {
		t.Fatal("LoadComponents should ignore non-.html files")
	}

	var buf strings.Builder
	data := map[string]any{"Children": template.HTML("body-content")}
	if err := root.ExecuteTemplate(&buf, "layout", data); err != nil {
		t.Fatalf("execute layout: %v", err)
	}
	want := `<body><div class="card">Sidebar</div>body-content</body>`
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestLoadComponentsPreprocessError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "broken.html"), `{{<card>}}no closing tag`)

	r := &Renderer{}
	root := template.New("root").Funcs(r.FuncMap())
	err := LoadComponents(root, dir)
	if err == nil {
		t.Fatal("LoadComponents should error on malformed component source")
	}
	if !strings.Contains(err.Error(), "broken.html") {
		t.Fatalf("error %q does not name the source file", err)
	}
}

func TestLoadComponentsMissingDir(t *testing.T) {
	r := &Renderer{}
	root := template.New("root").Funcs(r.FuncMap())
	err := LoadComponents(root, filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("LoadComponents should error when componentsDir is missing")
	}
}
