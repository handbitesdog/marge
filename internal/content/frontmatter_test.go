package content

import (
	"strings"
	"testing"
	"time"
)

func TestSplitFrontMatter(t *testing.T) {
	src := "---\ntitle: Hello World\ndate: 2024-01-15\nexcerpt: a short teaser\n---\nBody **text**.\n"
	fm, body, err := splitFrontMatter([]byte(src), "hello.md")
	if err != nil {
		t.Fatalf("splitFrontMatter: %v", err)
	}
	if fm.Title != "Hello World" {
		t.Fatalf("Title = %q, want %q", fm.Title, "Hello World")
	}
	wantDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	if !fm.Date.Equal(wantDate) {
		t.Fatalf("Date = %v, want %v", fm.Date, wantDate)
	}
	if fm.Extra["excerpt"] != "a short teaser" {
		t.Fatalf("Extra[excerpt] = %v, want %q", fm.Extra["excerpt"], "a short teaser")
	}
	if _, ok := fm.Extra["title"]; ok {
		t.Fatal("Extra should not contain the title field")
	}
	wantBody := "Body **text**.\n"
	if string(body) != wantBody {
		t.Fatalf("body = %q, want %q", body, wantBody)
	}
}

func TestSplitFrontMatterMissingTitle(t *testing.T) {
	src := "---\ndate: 2024-01-15\n---\nBody.\n"
	_, _, err := splitFrontMatter([]byte(src), "notitle.md")
	if err == nil {
		t.Fatal("splitFrontMatter should error when title is missing")
	}
	if !strings.Contains(err.Error(), "notitle.md") {
		t.Fatalf("error %q does not name the source file", err)
	}
}

func TestSplitFrontMatterMissingDate(t *testing.T) {
	src := "---\ntitle: No Date\n---\nBody.\n"
	_, _, err := splitFrontMatter([]byte(src), "nodate.md")
	if err == nil {
		t.Fatal("splitFrontMatter should error when date is missing")
	}
}

func TestSplitFrontMatterNoDelimiter(t *testing.T) {
	src := "Just a plain body, no front matter.\n"
	_, _, err := splitFrontMatter([]byte(src), "plain.md")
	if err == nil {
		t.Fatal("splitFrontMatter should error when the opening delimiter is missing")
	}
}

func TestSplitFrontMatterUnterminated(t *testing.T) {
	src := "---\ntitle: Unterminated\ndate: 2024-01-15\nBody without a closing delimiter.\n"
	_, _, err := splitFrontMatter([]byte(src), "unterminated.md")
	if err == nil {
		t.Fatal("splitFrontMatter should error when the closing delimiter is missing")
	}
}
