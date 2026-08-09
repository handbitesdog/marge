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
// syntax to reference another), ignores non-.html files, doesn't recurse
// into subdirectories, and makes a block invocation's captured body
// available as $children.
func TestLoadComponents(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "card.html"), `<div class="card">{{.title}}{{$children}}</div>`)
	writeFile(t, filepath.Join(dir, "layout.html"), `<body>{{<card title="Sidebar"/>}}{{$children}}</body>`)
	writeFile(t, filepath.Join(dir, "notes.txt"), `not a component`)
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	writeFile(t, filepath.Join(dir, "nested", "ignored.html"), `<p>ignored</p>`)

	r := &Renderer{}
	root := template.New("root").Funcs(r.FuncMap())
	if err := LoadComponents(root, dir, true, map[string]string{}); err != nil {
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
	data := map[string]any{"children": template.HTML("body-content")}
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
	err := LoadComponents(root, dir, true, map[string]string{})
	if err == nil {
		t.Fatal("LoadComponents should error on malformed component source")
	}
	if !strings.Contains(err.Error(), "broken.html") {
		t.Fatalf("error %q does not name the source file", err)
	}
}

func TestLoadComponentsMissingDirRequired(t *testing.T) {
	r := &Renderer{}
	root := template.New("root").Funcs(r.FuncMap())
	err := LoadComponents(root, filepath.Join(t.TempDir(), "does-not-exist"), true, map[string]string{})
	if err == nil {
		t.Fatal("LoadComponents should error when a required dir is missing")
	}
}

func TestLoadComponentsMissingDirOptional(t *testing.T) {
	r := &Renderer{}
	root := template.New("root").Funcs(r.FuncMap())
	err := LoadComponents(root, filepath.Join(t.TempDir(), "does-not-exist"), false, map[string]string{})
	if err != nil {
		t.Fatalf("LoadComponents should not error when an optional dir is missing: %v", err)
	}
}

// TestLoadComponentsCollision checks that reusing the same seen map across
// two LoadComponents calls (e.g. components/ then layouts/) reports a
// same-name collision as a build error rather than silently letting the
// second definition win.
func TestLoadComponentsCollision(t *testing.T) {
	componentsDir := t.TempDir()
	layoutsDir := t.TempDir()
	writeFile(t, filepath.Join(componentsDir, "card.html"), `<div class="card"></div>`)
	writeFile(t, filepath.Join(layoutsDir, "card.html"), `<div class="layout-card"></div>`)

	r := &Renderer{}
	root := template.New("root").Funcs(r.FuncMap())
	seen := map[string]string{}
	if err := LoadComponents(root, componentsDir, true, seen); err != nil {
		t.Fatalf("LoadComponents(componentsDir): %v", err)
	}
	err := LoadComponents(root, layoutsDir, false, seen)
	if err == nil {
		t.Fatal("LoadComponents should error when a name is defined in two directories")
	}
	if !strings.Contains(err.Error(), "card") {
		t.Fatalf("error %q does not name the colliding component", err)
	}
}
