package tmpl

import (
	"fmt"
	"html/template"
	"strings"
)

// dict builds a map[string]any from alternating key/value arguments, for use
// as the single-argument data context in {{template "name" (dict ...)}}
// calls emitted by Preprocess. pairs must have even length with string keys.
func dict(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("dict: odd number of arguments (%d)", len(pairs))
	}
	m := make(map[string]any, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: argument %d is a %T, not a string key", i, pairs[i])
		}
		m[key] = pairs[i+1]
	}
	return m, nil
}

// Renderer supplies the "render" template func used by Preprocess's output
// to render a captured block by name against an ambient dot. render needs
// the fully parsed template set to execute against, but Funcs() must be
// registered before parsing happens, so tset is a pointer field left nil
// until SetTemplateSet is called once parsing completes; renderFunc's
// closure only dereferences it later, at execution time.
type Renderer struct {
	tset *template.Template
}

// SetTemplateSet records the fully parsed template set that renderFunc
// executes against. Must be called after all templates (components, pages,
// collection items) have been parsed into t, and before any template is
// executed.
func (r *Renderer) SetTemplateSet(t *template.Template) {
	r.tset = t
}

// FuncMap returns the "dict" and "render" funcs, bound to r, for
// registration on a template set via Funcs() before parsing.
func (r *Renderer) FuncMap() template.FuncMap {
	return template.FuncMap{
		"dict":   dict,
		"render": r.renderFunc,
	}
}

// renderFunc executes the named template against data and returns its
// output as trusted, pre-escaped HTML.
func (r *Renderer) renderFunc(name string, data any) (template.HTML, error) {
	if r.tset == nil {
		return "", fmt.Errorf("render %q: template set not initialized", name)
	}
	var buf strings.Builder
	if err := r.tset.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("render %q: %w", name, err)
	}
	return template.HTML(buf.String()), nil
}
