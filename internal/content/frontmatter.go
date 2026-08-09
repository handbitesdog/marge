// Package content loads Markdown-sourced collection items (e.g. blog posts)
// from content/, parsing front matter and rendering bodies to HTML.
package content

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// frontMatter is the YAML block at the top of a content Markdown file,
// delimited by "---" lines.
type frontMatter struct {
	Title string         `yaml:"title"`
	Date  time.Time      `yaml:"date"`
	Extra map[string]any `yaml:",inline"`
}

// splitFrontMatter splits the leading "---"/"---" YAML front matter block
// from src and parses it, returning the remaining Markdown body. sourceName
// is used to produce descriptive errors. Title and Date are required.
func splitFrontMatter(src []byte, sourceName string) (frontMatter, []byte, error) {
	var fm frontMatter

	lines := strings.Split(string(src), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fm, nil, fmt.Errorf("%s: missing front matter delimiter", sourceName)
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return fm, nil, fmt.Errorf("%s: unterminated front matter block", sourceName)
	}

	yamlBlock := strings.Join(lines[1:end], "\n")
	body := strings.Join(lines[end+1:], "\n")

	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return fm, nil, fmt.Errorf("%s: parse front matter: %w", sourceName, err)
	}
	if strings.TrimSpace(fm.Title) == "" {
		return fm, nil, fmt.Errorf("%s: front matter missing title", sourceName)
	}
	if fm.Date.IsZero() {
		return fm, nil, fmt.Errorf("%s: front matter missing date", sourceName)
	}

	return fm, []byte(body), nil
}
