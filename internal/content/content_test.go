package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestLoadAll checks that LoadAll turns each content/ subdirectory into a
// Collection, sorts its items newest-first, renders Markdown (passing raw
// HTML through unescaped), and ignores non-.md files and top-level files.
func TestLoadAll(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "blog", "oldest.md"),
		"---\ntitle: Oldest\ndate: 2024-01-01\n---\nFirst post.\n")
	writeFile(t, filepath.Join(dir, "blog", "newest.md"),
		"---\ntitle: Newest\ndate: 2024-03-01\n---\n<strong>Raw</strong> html and *markdown*.\n")
	writeFile(t, filepath.Join(dir, "blog", "notes.txt"), "not markdown")
	writeFile(t, filepath.Join(dir, "ignored-top-level.md"), "should be ignored")

	collections, err := LoadAll(dir)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}

	blog, ok := collections["blog"]
	if !ok {
		t.Fatal(`LoadAll did not produce a "blog" collection`)
	}
	if blog.Name != "blog" {
		t.Fatalf("Name = %q, want %q", blog.Name, "blog")
	}
	if len(blog.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(blog.Items))
	}

	if blog.Items[0].Title != "Newest" || blog.Items[1].Title != "Oldest" {
		t.Fatalf("items not sorted newest-first: got %q, %q", blog.Items[0].Title, blog.Items[1].Title)
	}

	newest := blog.Items[0]
	if newest.Slug != "newest" {
		t.Fatalf("Slug = %q, want %q", newest.Slug, "newest")
	}
	if newest.URL != "/blog/newest/" {
		t.Fatalf("URL = %q, want %q", newest.URL, "/blog/newest/")
	}
	wantContent := "<p><strong>Raw</strong> html and <em>markdown</em>.</p>\n"
	if string(newest.Content) != wantContent {
		t.Fatalf("Content = %q, want %q", newest.Content, wantContent)
	}
}

func TestLoadAllPropagatesFrontMatterError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "blog", "broken.md"), "---\ndate: 2024-01-01\n---\nNo title.\n")

	_, err := LoadAll(dir)
	if err == nil {
		t.Fatal("LoadAll should propagate front matter errors")
	}
	if !strings.Contains(err.Error(), "broken.md") {
		t.Fatalf("error %q does not name the source file", err)
	}
}

func TestLoadAllMissingDir(t *testing.T) {
	_, err := LoadAll(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("LoadAll should error when contentDir is missing")
	}
}
