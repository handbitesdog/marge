package build

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestPrettyURL(t *testing.T) {
	tests := []struct {
		relPath        string
		wantOutputPath string
		wantURL        string
	}{
		{"index.html", "index.html", "/"},
		{"about.html", filepath.Join("about", "index.html"), "/about/"},
		{"blog/index.html", filepath.Join("blog", "index.html"), "/blog/"},
	}

	for _, tt := range tests {
		t.Run(tt.relPath, func(t *testing.T) {
			gotOutputPath, gotURL := prettyURL(filepath.FromSlash(tt.relPath))
			if gotOutputPath != tt.wantOutputPath {
				t.Errorf("outputPath = %q, want %q", gotOutputPath, tt.wantOutputPath)
			}
			if gotURL != tt.wantURL {
				t.Errorf("url = %q, want %q", gotURL, tt.wantURL)
			}
		})
	}
}

// TestDiscover checks that ordinary pages are routed via prettyURL,
// _item.html is captured as the item template for its parent directory's
// collection, and other underscore-prefixed files are excluded entirely.
func TestDiscover(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "index.html"), "home")
	writeFile(t, filepath.Join(dir, "about.html"), "about")
	writeFile(t, filepath.Join(dir, "blog", "index.html"), "blog index")
	writeFile(t, filepath.Join(dir, "blog", "_item.html"), "item template")
	writeFile(t, filepath.Join(dir, "_partial.html"), "partial, not routed")

	pages, itemTemplates, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if len(pages) != 3 {
		t.Fatalf("len(pages) = %d, want 3: %+v", len(pages), pages)
	}

	gotURLs := make(map[string]bool)
	for _, p := range pages {
		gotURLs[p.URL] = true
	}
	for _, url := range []string{"/", "/about/", "/blog/"} {
		if !gotURLs[url] {
			t.Errorf("missing page for URL %q", url)
		}
	}

	wantItemTemplates := map[string]string{"blog": filepath.Join("blog", "_item.html")}
	if !reflect.DeepEqual(itemTemplates, wantItemTemplates) {
		t.Errorf("itemTemplates = %+v, want %+v", itemTemplates, wantItemTemplates)
	}
}

func TestDiscoverMissingDir(t *testing.T) {
	_, _, err := Discover(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("Discover should error when pagesDir is missing")
	}
}
