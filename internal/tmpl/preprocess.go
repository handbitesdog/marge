// Package tmpl implements marge's component macro-preprocessor: rewriting
// the {{<name/>}} / {{<name>}}...{{</name>}} component syntax into standard
// html/template source before it reaches the Go template parser.
package tmpl

import (
	"fmt"
	"strings"
	"unicode"
)

// attr is a single key="value" (or key=.Pipeline) pair on a component tag.
// value is the raw pipeline-operand source, passed through verbatim into
// the generated dict(...) call.
type attr struct {
	key   string
	value string
}

// directive is a parsed {{<name ...>}}, {{<name .../>}}, or {{</name>}} tag.
type directive struct {
	closing     bool
	selfClosing bool
	name        string
	attrs       []attr
}

// openBlock is a frame on the preprocessor's stack for a block-form
// component awaiting its {{</name>}} close. body accumulates everything
// (literal text and nested emissions) between the open and close tags.
type openBlock struct {
	name  string
	attrs []attr
	body  strings.Builder
}

// Preprocess rewrites marge's {{<name/>}} / {{<name>}}...{{</name>}}
// component syntax found in src into plain html/template source. sourceName
// (typically a file path) is used both in error messages and, sanitized, as
// the prefix for generated block template names.
func Preprocess(src, sourceName string) (string, error) {
	prefix := sanitizeName(sourceName)
	var root strings.Builder
	var stack []*openBlock
	counter := 0

	out := func() *strings.Builder {
		if len(stack) == 0 {
			return &root
		}
		return &stack[len(stack)-1].body
	}

	i, n := 0, len(src)
	for i < n {
		idx := strings.Index(src[i:], "{{")
		if idx == -1 {
			out().WriteString(src[i:])
			break
		}
		start := i + idx
		out().WriteString(src[i:start])

		// A directive is marked by a leading '<' (skipping whitespace)
		// right after "{{" — no valid Go template action can start with
		// '<' (comparisons are the functions lt/gt), so this is
		// unambiguous without reserving a keyword.
		j := start + 2
		for j < n && isSpace(src[j]) {
			j++
		}
		if j >= n || src[j] != '<' {
			out().WriteString("{{")
			i = start + 2
			continue
		}

		end, err := findDirectiveEnd(src, start)
		if err != nil {
			return "", fmt.Errorf("%s: %w", sourceName, err)
		}
		d, err := parseDirective(src[start+2 : end])
		if err != nil {
			return "", fmt.Errorf("%s: %w", sourceName, err)
		}

		switch {
		case d.closing:
			if len(stack) == 0 {
				return "", fmt.Errorf("%s: closing tag %q has no matching open tag", sourceName, d.name)
			}
			top := stack[len(stack)-1]
			if top.name != d.name {
				return "", fmt.Errorf("%s: mismatched closing tag: expected %q, got %q", sourceName, top.name, d.name)
			}
			stack = stack[:len(stack)-1]
			counter++
			blockName := fmt.Sprintf("__block_%s_%d", prefix, counter)
			// {{define}} is only recognized by the Go template parser at
			// the top level of a parse — never nested inside another
			// block's captured body — so it always goes to root, however
			// deeply the source tag was nested; only the {{template}} call
			// (valid anywhere) is emitted at the point of use. Named
			// templates resolve globally, so this doesn't change behavior.
			root.WriteString(fmt.Sprintf(`{{define %q}}%s{{end}}`, blockName, top.body.String()))
			out().WriteString(fmt.Sprintf(`{{template %q (dict %s"children" (render %q .))}}`,
				top.name, dictArgs(top.attrs), blockName))

		case d.selfClosing:
			out().WriteString(fmt.Sprintf(`{{template %q (dict %s"children" "")}}`, d.name, dictArgs(d.attrs)))

		default:
			stack = append(stack, &openBlock{name: d.name, attrs: d.attrs})
		}

		i = end + 2
	}

	if len(stack) > 0 {
		names := make([]string, len(stack))
		for idx, f := range stack {
			names[idx] = f.name
		}
		return "", fmt.Errorf("%s: unclosed component tag(s): %s", sourceName, strings.Join(names, ", "))
	}

	return root.String(), nil
}

// dictArgs renders attrs as "key1" val1 "key2" val2, always followed by a
// trailing space (or empty) so callers can append "children" ... directly.
func dictArgs(attrs []attr) string {
	if len(attrs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(attrs)*2)
	for _, a := range attrs {
		parts = append(parts, fmt.Sprintf("%q", a.key), a.value)
	}
	return strings.Join(parts, " ") + " "
}

// findDirectiveEnd scans quote-aware from src[start:] (start is the index of
// "{{") for the matching "}}", so a quoted attribute value containing '}'
// isn't mistaken for the directive's end. It returns the index of the first
// '}' of that closing "}}".
func findDirectiveEnd(src string, start int) (int, error) {
	i, n := start+2, len(src)
	for i < n {
		switch c := src[i]; c {
		case '"', '`':
			i, _ = skipQuoted(src, i)
		case '}':
			if i+1 < n && src[i+1] == '}' {
				return i, nil
			}
			i++
		default:
			i++
		}
	}
	return -1, fmt.Errorf(`unclosed directive (missing "}}")`)
}

// skipQuoted returns the index just past the quoted string starting at i
// (src[i] is the opening quote), and whether a matching close quote was
// found. Double-quoted strings support backslash escapes; backtick raw
// strings do not, matching Go template pipeline literal syntax.
func skipQuoted(src string, i int) (end int, closed bool) {
	quote := src[i]
	n := len(src)
	i++
	for i < n {
		if quote == '"' && src[i] == '\\' && i+1 < n {
			i += 2
			continue
		}
		if src[i] == quote {
			return i + 1, true
		}
		i++
	}
	return n, false
}

// parseDirective tokenizes body, the raw text between "{{" and "}}" of a
// confirmed directive (body starts with '<', modulo leading whitespace).
func parseDirective(body string) (directive, error) {
	i, n := 0, len(body)
	skipSpace := func() {
		for i < n && isSpace(body[i]) {
			i++
		}
	}

	skipSpace()
	if i >= n || body[i] != '<' {
		return directive{}, fmt.Errorf("expected '<' to start component tag")
	}
	i++
	skipSpace()

	var d directive
	if i < n && body[i] == '/' {
		d.closing = true
		i++
		skipSpace()
		start := i
		for i < n && !isSpace(body[i]) && body[i] != '>' {
			i++
		}
		d.name = body[start:i]
		if d.name == "" {
			return directive{}, fmt.Errorf("closing tag missing component name")
		}
		skipSpace()
		if i >= n || body[i] != '>' {
			return directive{}, fmt.Errorf("closing tag %q: missing '>'", d.name)
		}
		i++
		skipSpace()
		if i != n {
			return directive{}, fmt.Errorf("closing tag %q: unexpected content after '>'", d.name)
		}
		return d, nil
	}

	start := i
	for i < n && !isSpace(body[i]) && body[i] != '/' && body[i] != '>' {
		i++
	}
	d.name = body[start:i]
	if d.name == "" {
		return directive{}, fmt.Errorf("component tag missing name")
	}

	for {
		skipSpace()
		if i >= n {
			return directive{}, fmt.Errorf("component %q: unterminated tag", d.name)
		}
		if body[i] == '/' {
			i++
			skipSpace()
			if i >= n || body[i] != '>' {
				return directive{}, fmt.Errorf("component %q: expected '>' after '/'", d.name)
			}
			d.selfClosing = true
			i++
			break
		}
		if body[i] == '>' {
			i++
			break
		}

		keyStart := i
		for i < n && body[i] != '=' && !isSpace(body[i]) && body[i] != '/' && body[i] != '>' {
			i++
		}
		key := body[keyStart:i]
		if key == "" {
			return directive{}, fmt.Errorf("component %q: expected attribute", d.name)
		}
		skipSpace()
		if i >= n || body[i] != '=' {
			return directive{}, fmt.Errorf("component %q: attribute %q missing '='", d.name, key)
		}
		i++
		skipSpace()
		if i >= n {
			return directive{}, fmt.Errorf("component %q: attribute %q missing value", d.name, key)
		}

		var value string
		if body[i] == '"' || body[i] == '`' {
			end, closed := skipQuoted(body, i)
			if !closed {
				return directive{}, fmt.Errorf("component %q: attribute %q: unterminated quoted value", d.name, key)
			}
			value = body[i:end]
			i = end
		} else {
			valStart := i
			for i < n && !isSpace(body[i]) && body[i] != '/' && body[i] != '>' {
				i++
			}
			value = body[valStart:i]
		}
		d.attrs = append(d.attrs, attr{key: key, value: value})
	}

	skipSpace()
	if i != n {
		return directive{}, fmt.Errorf("component %q: unexpected content after tag", d.name)
	}
	return d, nil
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// sanitizeName maps sourceName to a string safe for use inside a generated
// Go template identifier: letters, digits, and underscores pass through;
// everything else becomes '_'.
func sanitizeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}
