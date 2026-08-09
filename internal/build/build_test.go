package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const exampleDir = "../../example/src"

// TestRunBuildsExampleSite is the integration test for the full pipeline: it
// runs Run against the example/ fixture site and checks the output pages,
// item pages, collection ordering, static asset copy, and that _item.html
// itself was not routed as an ordinary page.
func TestRunBuildsExampleSite(t *testing.T) {
	dist := t.TempDir()

	if err := Run(Options{SrcDir: exampleDir, DistDir: dist}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	readFile := func(rel string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dist, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		return string(b)
	}

	for _, rel := range []string{
		"index.html",
		filepath.Join("about", "index.html"),
		filepath.Join("blog", "index.html"),
		filepath.Join("blog", "hello-world", "index.html"),
	} {
		if _, err := os.Stat(filepath.Join(dist, rel)); err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
		}
	}

	home := readFile("index.html")
	if !strings.Contains(home, "<title>Home</title>") {
		t.Errorf("index.html missing expected title, got:\n%s", home)
	}

	about := readFile(filepath.Join("about", "index.html"))
	if !strings.Contains(about, "<title>About</title>") {
		t.Errorf("about/index.html missing expected title, got:\n%s", about)
	}

	item := readFile(filepath.Join("blog", "hello-world", "index.html"))
	if !strings.Contains(item, "Hello, World!") {
		t.Errorf("blog/hello-world/index.html missing expected title, got:\n%s", item)
	}
	if !strings.Contains(item, "<strong>first post</strong>") {
		t.Errorf("blog/hello-world/index.html missing rendered markdown, got:\n%s", item)
	}

	blogIndex := readFile(filepath.Join("blog", "index.html"))
	posSecond := strings.Index(blogIndex, "Second Post")
	posThird := strings.Index(blogIndex, "Third Post")
	posHello := strings.Index(blogIndex, "Hello, World!")
	if posSecond == -1 || posThird == -1 || posHello == -1 {
		t.Fatalf("blog/index.html missing expected posts, got:\n%s", blogIndex)
	}
	if !(posSecond < posThird && posThird < posHello) {
		t.Errorf("blog/index.html posts not newest-first: Second=%d, Third=%d, Hello=%d", posSecond, posThird, posHello)
	}

	wantCSS, err := os.ReadFile(filepath.Join(exampleDir, "static", "css", "site.css"))
	if err != nil {
		t.Fatalf("read source css: %v", err)
	}
	if gotCSS := readFile(filepath.Join("css", "site.css")); gotCSS != string(wantCSS) {
		t.Errorf("css/site.css does not match source byte-for-byte")
	}

	if _, err := os.Stat(filepath.Join(dist, "blog", "_item")); !os.IsNotExist(err) {
		t.Errorf("blog/_item/ directory should not exist, stat err = %v", err)
	}
}

// TestRunCleansExistingDist checks that Run performs a clean build, removing
// files left over in DistDir from a previous run (e.g. a deleted post).
func TestRunCleansExistingDist(t *testing.T) {
	dist := t.TempDir()
	stale := filepath.Join(dist, "stale.html")
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	if err := Run(Options{SrcDir: exampleDir, DistDir: dist}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("Run should remove pre-existing dist contents, stat err = %v", err)
	}
}

// TestRunMissingItemTemplate checks that Run fails fast, naming the
// collection, when content exists for a collection but pages/<name>/_item.html
// does not.
func TestRunMissingItemTemplate(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "components", "Layout.html"), `{{$children}}`)
	writeFile(t, filepath.Join(src, "pages", "index.html"), `home`)
	writeFile(t, filepath.Join(src, "content", "blog", "post.md"),
		"---\ntitle: Post\ndate: 2024-01-01\n---\nBody.\n")

	err := Run(Options{SrcDir: src, DistDir: t.TempDir()})
	if err == nil {
		t.Fatal("Run should error when a collection has content but no _item.html")
	}
	if !strings.Contains(err.Error(), "blog") {
		t.Fatalf("error %q does not name the collection", err)
	}
}

// TestRunOnlyPagesRequired checks that Run succeeds with nothing but a
// pages/ directory present — components/, layouts/, content/, and static/
// are all optional.
func TestRunOnlyPagesRequired(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "pages", "index.html"), `home`)

	dist := t.TempDir()
	if err := Run(Options{SrcDir: src, DistDir: dist}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dist, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	if string(got) != "home" {
		t.Fatalf("index.html = %q, want %q", got, "home")
	}
}

// TestRunMissingPagesDir checks that pages/ remains the one required
// subdirectory: Run fails when it's absent, even though every other
// subdirectory is optional.
func TestRunMissingPagesDir(t *testing.T) {
	src := t.TempDir()
	if err := Run(Options{SrcDir: src, DistDir: t.TempDir()}); err == nil {
		t.Fatal("Run should error when pages/ is missing")
	}
}
