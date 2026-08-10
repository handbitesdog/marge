package data

import (
	"os"
	"path/filepath"
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

// TestLoadAll checks that LoadAll decodes each top-level .json file, keyed
// by its filename without extension, and ignores non-.json files and
// subdirectories.
func TestLoadAll(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "footbags.json"), `[
		{"name": "Whirled", "weight_g": 34.5, "image": "/static/footbags/whirled.jpg"},
		{"name": "Standard", "weight_g": 32, "image": "/static/footbags/standard.jpg"}
	]`)
	writeFile(t, filepath.Join(dir, "notes.txt"), "not json")
	writeFile(t, filepath.Join(dir, "ignored-dir", "nested.json"), `{"a": 1}`)

	got, err := LoadAll(dir)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len(data) = %d, want 1", len(got))
	}

	footbags, ok := got["footbags"].([]any)
	if !ok {
		t.Fatalf(`data["footbags"] = %#v, want []any`, got["footbags"])
	}
	if len(footbags) != 2 {
		t.Fatalf("len(footbags) = %d, want 2", len(footbags))
	}

	first, ok := footbags[0].(map[string]any)
	if !ok {
		t.Fatalf("footbags[0] = %#v, want map[string]any", footbags[0])
	}
	if first["name"] != "Whirled" {
		t.Fatalf(`footbags[0]["name"] = %v, want "Whirled"`, first["name"])
	}
	if first["weight_g"] != 34.5 {
		t.Fatalf(`footbags[0]["weight_g"] = %v, want 34.5`, first["weight_g"])
	}
}

func TestLoadAllInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "broken.json"), `{not valid json`)

	_, err := LoadAll(dir)
	if err == nil {
		t.Fatal("LoadAll should error on invalid JSON")
	}
}

func TestLoadAllMissingDir(t *testing.T) {
	got, err := LoadAll(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("LoadAll should not error when dataDir is missing: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("data = %#v, want empty", got)
	}
}
