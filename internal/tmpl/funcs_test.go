package tmpl

import (
	"html/template"
	"reflect"
	"strings"
	"testing"
)

func TestDict(t *testing.T) {
	got, err := dict("a", 1, "b", "two")
	if err != nil {
		t.Fatalf("dict returned error: %v", err)
	}
	want := map[string]any{"a": 1, "b": "two"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dict = %#v, want %#v", got, want)
	}
}

func TestDictEmpty(t *testing.T) {
	got, err := dict()
	if err != nil {
		t.Fatalf("dict returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("dict() = %#v, want empty map", got)
	}
}

func TestDictOddArgs(t *testing.T) {
	_, err := dict("a", 1, "b")
	if err == nil {
		t.Fatal("dict with odd argument count should error")
	}
}

func TestDictNonStringKey(t *testing.T) {
	_, err := dict(1, "a")
	if err == nil {
		t.Fatal("dict with non-string key should error")
	}
}

func TestRendererBeforeSetTemplateSet(t *testing.T) {
	r := &Renderer{}
	_, err := r.renderFunc("whatever", nil)
	if err == nil {
		t.Fatal("renderFunc before SetTemplateSet should error, not panic or silently succeed")
	}
}

// TestRendererWithPreprocessedComponent exercises the full round-trip: a
// component-syntax source is run through Preprocess, parsed alongside a
// component define using the Renderer's FuncMap, then executed. It checks
// that render(name, .) passes the ambient dot at the call site so a
// captured children block can still see the outer page's fields.
func TestRendererWithPreprocessedComponent(t *testing.T) {
	src := `{{<layout title=.Title>}}Hello {{.Name}}{{</layout>}}`
	out, err := Preprocess(src, "pages/index.html")
	if err != nil {
		t.Fatalf("Preprocess: %v", err)
	}

	r := &Renderer{}
	root := template.New("root").Funcs(r.FuncMap())
	root, err = root.Parse(`{{define "layout"}}<h1>{{.title}}</h1><div>{{.children}}</div>{{end}}`)
	if err != nil {
		t.Fatalf("parse layout define: %v", err)
	}
	if _, err := root.New("page").Parse(out); err != nil {
		t.Fatalf("parse preprocessed page: %v", err)
	}
	r.SetTemplateSet(root)

	var buf strings.Builder
	data := struct{ Title, Name string }{Title: "Home", Name: "World"}
	if err := root.ExecuteTemplate(&buf, "page", data); err != nil {
		t.Fatalf("execute page: %v", err)
	}
	want := `<h1>Home</h1><div>Hello World</div>`
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}
