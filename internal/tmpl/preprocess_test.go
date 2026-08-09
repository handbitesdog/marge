package tmpl

import (
	"html/template"
	"strings"
	"testing"
)

// stubFuncs registers no-op dict/render implementations with matching
// signatures so Preprocess's output can be checked for valid Go template
// syntax via Parse, without needing the real Phase 3 implementations.
func stubFuncs() template.FuncMap {
	return template.FuncMap{
		"dict": func(pairs ...any) (map[string]any, error) {
			return nil, nil
		},
		"render": func(name string, data any) (template.HTML, error) {
			return "", nil
		},
	}
}

func mustParse(t *testing.T, out string) {
	t.Helper()
	if _, err := template.New("t").Funcs(stubFuncs()).Parse(out); err != nil {
		t.Fatalf("preprocessed output is not valid template syntax: %v\noutput:\n%s", err, out)
	}
}

func TestPreprocessSelfClosing(t *testing.T) {
	src := `Hello {{<card title="Post" featured=.Featured/>}} world`
	want := `Hello {{template "card" (dict "title" "Post" "featured" .Featured "children" "")}} world`

	got, err := Preprocess(src, "pages/index.html")
	if err != nil {
		t.Fatalf("Preprocess returned error: %v", err)
	}
	if got != want {
		t.Fatalf("Preprocess output mismatch:\n got: %s\nwant: %s", got, want)
	}
	mustParse(t, got)
}

func TestPreprocessSelfClosingNoAttrs(t *testing.T) {
	src := `{{<spacer/>}}`
	want := `{{template "spacer" (dict "children" "")}}`

	got, err := Preprocess(src, "components/spacer.html")
	if err != nil {
		t.Fatalf("Preprocess returned error: %v", err)
	}
	if got != want {
		t.Fatalf("Preprocess output mismatch:\n got: %s\nwant: %s", got, want)
	}
	mustParse(t, got)
}

func TestPreprocessNestedBlocks(t *testing.T) {
	src := `{{<layout title="Home">}}A{{<card/>}}B{{</layout>}}`
	want := `{{define "__block_pages_index_html_1"}}A{{template "card" (dict "children" "")}}B{{end}}` +
		`{{template "layout" (dict "title" "Home" "children" (render "__block_pages_index_html_1" .))}}`

	got, err := Preprocess(src, "pages/index.html")
	if err != nil {
		t.Fatalf("Preprocess returned error: %v", err)
	}
	if got != want {
		t.Fatalf("Preprocess output mismatch:\n got: %s\nwant: %s", got, want)
	}
	mustParse(t, got)
}

func TestPreprocessDeeplyNestedBlocks(t *testing.T) {
	// Go's template parser only recognizes {{define}} at the top level of a
	// parse, never nested inside another {{define}}'s body — so both block
	// defines must land as top-level siblings, however deeply the source
	// tags were nested. Only the {{template}} calls stay inline.
	src := `{{<outer>}}A{{<inner>}}B{{</inner>}}C{{</outer>}}`
	want := `{{define "__block_pages_index_html_1"}}B{{end}}` +
		`{{define "__block_pages_index_html_2"}}A{{template "inner" (dict "children" (render "__block_pages_index_html_1" .))}}C{{end}}` +
		`{{template "outer" (dict "children" (render "__block_pages_index_html_2" .))}}`

	got, err := Preprocess(src, "pages/index.html")
	if err != nil {
		t.Fatalf("Preprocess returned error: %v", err)
	}
	if got != want {
		t.Fatalf("Preprocess output mismatch:\n got: %s\nwant: %s", got, want)
	}
	mustParse(t, got)
}

func TestPreprocessNestedBlockSurroundedByText(t *testing.T) {
	// Definitions hoisted to root must not corrupt literal text that
	// appears before/after the block in the original source.
	src := `Header {{<outer>}}A{{<inner>}}B{{</inner>}}C{{</outer>}} Footer`
	want := `Header ` +
		`{{define "__block_pages_index_html_1"}}B{{end}}` +
		`{{define "__block_pages_index_html_2"}}A{{template "inner" (dict "children" (render "__block_pages_index_html_1" .))}}C{{end}}` +
		`{{template "outer" (dict "children" (render "__block_pages_index_html_2" .))}}` +
		` Footer`

	got, err := Preprocess(src, "pages/index.html")
	if err != nil {
		t.Fatalf("Preprocess returned error: %v", err)
	}
	if got != want {
		t.Fatalf("Preprocess output mismatch:\n got: %s\nwant: %s", got, want)
	}
	mustParse(t, got)
}

func TestPreprocessUnclosedBlock(t *testing.T) {
	src := `{{<card title="X">}}no closing tag`

	_, err := Preprocess(src, "pages/index.html")
	if err == nil {
		t.Fatal("Preprocess returned no error for unclosed block")
	}
	if !strings.Contains(err.Error(), "unclosed") || !strings.Contains(err.Error(), "card") {
		t.Fatalf("error %q does not report the unclosed component", err)
	}
}

func TestPreprocessMismatchedClosingTag(t *testing.T) {
	src := `{{<card>}}body{{</widget>}}`

	_, err := Preprocess(src, "pages/index.html")
	if err == nil {
		t.Fatal("Preprocess returned no error for mismatched closing tag")
	}
	if !strings.Contains(err.Error(), "card") || !strings.Contains(err.Error(), "widget") {
		t.Fatalf("error %q does not name both the expected and actual tag", err)
	}
}

func TestPreprocessQuotedValueWithBraceAndAngle(t *testing.T) {
	src := `{{<card title="a}b" note="x>y"/>}}`
	want := `{{template "card" (dict "title" "a}b" "note" "x>y" "children" "")}}`

	got, err := Preprocess(src, "pages/index.html")
	if err != nil {
		t.Fatalf("Preprocess returned error: %v", err)
	}
	if got != want {
		t.Fatalf("Preprocess output mismatch:\n got: %s\nwant: %s", got, want)
	}
	mustParse(t, got)
}

func TestPreprocessPassesThroughOrdinaryActions(t *testing.T) {
	src := `{{.Title}} and {{ if .Featured }}yes{{ end }}`

	got, err := Preprocess(src, "pages/index.html")
	if err != nil {
		t.Fatalf("Preprocess returned error: %v", err)
	}
	if got != src {
		t.Fatalf("Preprocess should pass non-directive actions through unchanged:\n got: %s\nwant: %s", got, src)
	}
}
