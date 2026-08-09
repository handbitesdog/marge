# marge

A simple static-site generator written in Go, with reusable HTML components,
file-based page routing, and a Markdown content pipeline for things like
blog posts.

## Features

- **Components** — an HTML-tag-like syntax for reusable, argument-taking
  template fragments. Layouts are just components kept in their own
  `layouts/` folder.
- **File-based routing with pretty URLs** — `pages/about.html` becomes
  `/about/`.
- **Collections** — group Markdown content (e.g. blog posts) under
  `content/<name>/` and list them from any page via `.Collections`.
- **Static passthrough** — anything in `static/` is copied byte-for-byte
  into the output.
- **Dev server** — `marge serve` builds, serves the output over HTTP, and
  rebuilds automatically as source files change.

## Requirements

- Go 1.25+

## Installation

Install the `marge` command with `go install`:

```bash
go install .
```

This builds the binary and places it in `$(go env GOPATH)/bin`. Make sure
that directory is on your `PATH` — add this to your shell profile
(`~/.zshrc`, `~/.bashrc`, etc.) if it isn't already:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

After that, `marge` is available as a command from any project. Rerun
`go install .` from this repo whenever the source changes, to refresh the
installed binary.

## Usage

```bash
marge build <src> <dist>          # one-shot build
marge serve <src> <dist> <addr>   # build, serve dist/, rebuild on change
```

For example, to build and serve the reference site in [example/](example/):

```bash
marge build example dist
marge serve example dist :8080
```

A `marge` site is expected to look like:

```
mysite/
  components/   # reusable HTML components
  layouts/      # page-wrapping components (same syntax, own folder; optional)
  pages/        # routed pages, mapped to pretty URLs (about.html -> /about/)
  content/      # Markdown content with YAML front matter, grouped into collections
  static/       # assets copied byte-for-byte into the output
```

## Components

Files under `components/`, `layouts/`, and `pages/` can use a self-closing or
block tag syntax inside `{{}}` to render a component. `components/` and
`layouts/` share one flat namespace, so the same name can't be defined in
both.

```html
{{<card title=.Title url=.URL date=.Date/>}}
```

```html
{{<layout title="Home">}}
<h1>Welcome to marge</h1>
{{</layout>}}
```

A block tag's body is rendered and passed to the component as the variable
`$children`, stable regardless of any `{{range}}`/`{{with}}` elsewhere in the
component that changes what `.` refers to. `key=value` pairs are plain Go
template pipeline operands (string literals or `.Field`) — the key is an
arbitrary lowercase name you choose, so `layouts/layout.html` above can
reference `.title` and `$children` like any other template:

```html
<title>{{.title}}</title>
<main>{{$children}}</main>
```

## Content & collections

Each subdirectory of `content/` is a collection. A Markdown file with YAML
front matter:

```markdown
---
title: "Hello, World!"
date: 2024-01-05
---

This is the **first post** on the example site.
```

is exposed to every page and item template as `.Collections.<name>`, so a
homepage can list posts the same way a blog index does:

```html
{{range .Collections.blog.Items}}
{{<card title=.Title url=.URL date=.Date/>}}
{{end}}
```

`pages/<collection>/_item.html`, if present, is the per-item template for
that collection — it's executed once per item and written to
`dist/<collection>/<slug>/index.html`.

## Development

```bash
go build ./...
go vet ./...
go test ./...
```

## Status

All phases in [PHASES.md](PHASES.md) are complete. `marge` builds and serves
sites via the CLI described above.

## Further reading

Full design is in [PLAN.md](PLAN.md), including the macro-preprocessor
algorithm, key types, and the build pipeline. The build was broken into
phases, tracked in [PHASES.md](PHASES.md).
