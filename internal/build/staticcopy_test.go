package build

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyStatic(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	writeFile(t, filepath.Join(src, "css", "site.css"), "body { margin: 0; }")
	scriptPath := filepath.Join(src, "js", "app.js")
	writeFile(t, scriptPath, "console.log('hi');")
	if err := os.Chmod(scriptPath, 0o755); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	if err := CopyStatic(src, dest); err != nil {
		t.Fatalf("CopyStatic: %v", err)
	}

	gotCSS, err := os.ReadFile(filepath.Join(dest, "css", "site.css"))
	if err != nil {
		t.Fatalf("read copied css: %v", err)
	}
	if string(gotCSS) != "body { margin: 0; }" {
		t.Fatalf("css content = %q, want %q", gotCSS, "body { margin: 0; }")
	}

	info, err := os.Stat(filepath.Join(dest, "js", "app.js"))
	if err != nil {
		t.Fatalf("stat copied js: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("mode = %v, want %v", info.Mode().Perm(), os.FileMode(0o755))
	}
}

func TestCopyStaticMissingSrcIsNoop(t *testing.T) {
	dest := t.TempDir()
	if err := CopyStatic(filepath.Join(t.TempDir(), "does-not-exist"), dest); err != nil {
		t.Fatalf("CopyStatic should be a no-op when src is absent: %v", err)
	}

	entries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("dest should remain empty, got %d entries", len(entries))
	}
}
