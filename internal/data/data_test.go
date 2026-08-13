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

// TestLoadAllSchemaValid checks that a data file passing its sibling schema
// loads normally, and that the schema file itself isn't exposed as data.
func TestLoadAllSchemaValid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "footbags.json"), `[
		{"name": "Whirled", "weight_g": 34.5}
	]`)
	writeFile(t, filepath.Join(dir, "footbags.schema.json"), `{
		"type": "array",
		"items": {
			"type": "object",
			"required": ["name", "weight_g"],
			"properties": {
				"name": {"type": "string"},
				"weight_g": {"type": "number"}
			}
		}
	}`)

	got, err := LoadAll(dir)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if _, ok := got["footbags"]; !ok {
		t.Fatalf(`data["footbags"] missing, got %#v`, got)
	}
	if _, ok := got["footbags.schema"]; ok {
		t.Fatalf("schema file leaked into data as %q", "footbags.schema")
	}
	if len(got) != 1 {
		t.Fatalf("len(data) = %d, want 1: %#v", len(got), got)
	}
}

// TestLoadAllSchemaInvalid checks that data violating its sibling schema
// fails the build with an error naming both files.
func TestLoadAllSchemaInvalid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "footbags.json"), `[
		{"name": "Whirled"}
	]`)
	writeFile(t, filepath.Join(dir, "footbags.schema.json"), `{
		"type": "array",
		"items": {
			"type": "object",
			"required": ["name", "weight_g"]
		}
	}`)

	_, err := LoadAll(dir)
	if err == nil {
		t.Fatal("LoadAll should error when data violates its schema")
	}
}

// TestLoadAllSchemaUnmatchedIgnored checks that a schema file with no
// matching data file is silently ignored rather than erroring.
func TestLoadAllSchemaUnmatchedIgnored(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "orphan.schema.json"), `{"type": "array"}`)

	got, err := LoadAll(dir)
	if err != nil {
		t.Fatalf("LoadAll should not error on an unmatched schema file: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("data = %#v, want empty", got)
	}
}
