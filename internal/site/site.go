// Package site defines the core data types shared across marge's build
// pipeline: pages, collections, and the data contexts passed to templates.
package site

import (
	"html/template"
	"time"
)

// Page is an ordinary routed page discovered under pages/.
type Page struct {
	SourcePath string // relative to pages/, e.g. "blog/index.html"
	OutputPath string // relative to dist/, e.g. "blog/index.html"
	URL        string // e.g. "/blog/"
}

// CollectionItem is a single Markdown-sourced entry within a Collection.
type CollectionItem struct {
	Title   string
	Date    time.Time
	Slug    string
	URL     string
	Content template.HTML  // goldmark-rendered body
	Extra   map[string]any // front matter fields beyond title/date
}

// Collection groups CollectionItems under a content/ subdirectory name.
type Collection struct {
	Name  string
	Items []CollectionItem // sorted by Date descending
}

// PageData is the template data context for ordinary pages.
type PageData struct {
	Collections map[string]Collection // keyed by content/ subdirectory name
	Data        map[string]any        // keyed by data/ filename, without .json
}

// ItemData is the template data context for a single collection item.
type ItemData struct {
	CollectionItem
	Collections map[string]Collection
	Data        map[string]any
}
