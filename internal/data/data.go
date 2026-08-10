// Package data loads static JSON data files from data/, exposing each as a
// named value for use in templates (e.g. a table driven by a flat list of
// records, rather than one Markdown file per item).
package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadAll reads every top-level .json file in dataDir and decodes it into a
// Go value (object, array, or scalar — whatever the file contains), keyed by
// its filename without the .json extension. Subdirectories and non-.json
// files are ignored.
//
// A missing dataDir is not an error — LoadAll simply returns no data.
func LoadAll(dataDir string) (map[string]any, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read data dir %q: %w", dataDir, err)
	}

	data := make(map[string]any)
	for _, f := range entries {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		path := filepath.Join(dataDir, f.Name())
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", path, err)
		}

		var v any
		if err := json.Unmarshal(src, &v); err != nil {
			return nil, fmt.Errorf("parse json %q: %w", path, err)
		}

		name := strings.TrimSuffix(f.Name(), ".json")
		data[name] = v
	}

	return data, nil
}
