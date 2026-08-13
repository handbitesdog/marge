// Package data loads static JSON data files from data/, exposing each as a
// named value for use in templates (e.g. a table driven by a flat list of
// records, rather than one Markdown file per item). A data file may be
// paired with a sibling JSON Schema file to have its shape checked at build
// time.
package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

const schemaSuffix = ".schema.json"

// LoadAll reads every top-level .json file in dataDir and decodes it into a
// Go value (object, array, or scalar — whatever the file contains), keyed by
// its filename without the .json extension. Subdirectories and non-.json
// files are ignored.
//
// A data file named "<name>.json" may be paired with a sibling
// "<name>.schema.json" JSON Schema file in the same directory; if present,
// the data is validated against it and a mismatch fails the build. Schema
// files are not themselves exposed as data. A schema file with no matching
// data file is ignored.
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

	schemaPaths := make(map[string]string)
	for _, f := range entries {
		if !f.IsDir() && strings.HasSuffix(f.Name(), schemaSuffix) {
			name := strings.TrimSuffix(f.Name(), schemaSuffix)
			schemaPaths[name] = filepath.Join(dataDir, f.Name())
		}
	}

	data := make(map[string]any)
	for _, f := range entries {
		if f.IsDir() || strings.HasSuffix(f.Name(), schemaSuffix) || !strings.HasSuffix(f.Name(), ".json") {
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
		if schemaPath, ok := schemaPaths[name]; ok {
			if err := validate(schemaPath, v); err != nil {
				return nil, fmt.Errorf("data %q does not match schema %q: %w", path, schemaPath, err)
			}
		}

		data[name] = v
	}

	return data, nil
}

// validate compiles the JSON Schema at schemaPath and checks v against it.
func validate(schemaPath string, v any) error {
	schema, err := jsonschema.Compile(schemaPath)
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}
	return schema.Validate(v)
}
