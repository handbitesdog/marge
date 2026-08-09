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
marge build [<src> <dist>]          # one-shot build
marge serve [<src> <dist>] <addr>   # build, serve dist/, rebuild on change
```

`src` and `dist` default to `src/` and `dist/` (relative to the current
directory) when omitted; pass them explicitly to override.

For example, to build and serve the reference site in [example/](example/):

```bash
marge build example/src example/dist
marge serve example/src example/dist :8080
```

or, from inside `example/` itself, relying on the defaults:

```bash
marge build
marge serve :8080
```

A `marge` site is expected to look like:

```
mysite/
  src/
    components/   # reusable HTML components (optional)
    layouts/      # page-wrapping components (same syntax, own folder; optional)
    pages/        # routed pages, mapped to pretty URLs (about.html -> /about/) — required
    content/      # Markdown content with YAML front matter, grouped into collections (optional)
    static/       # assets copied byte-for-byte into the output (optional)
  dist/           # build output (created by `marge build`)
```

`pages/` is the only required directory — everything else can be omitted
entirely if a site doesn't need it.

## Components

Files under `components/`, `layouts/`, and `pages/` can use a self-closing or
block tag syntax inside `{{}}` to render a component. `components/` and
`layouts/` share one flat namespace, so the same name can't be defined in
both. A component or layout's name — and its file's basename — must start
with an uppercase letter (`Card`, not `card`), so a call site reads as a
component invocation at a glance.

```html
{{<Card title=.Title url=.URL date=.Date/>}}
```

```html
{{<Layout title="Home">}}
<h1>Welcome to marge</h1>
{{</Layout>}}
```

A block tag's body is rendered and passed to the component as the variable
`$children`, stable regardless of any `{{range}}`/`{{with}}` elsewhere in the
component that changes what `.` refers to. `key=value` pairs are plain Go
template pipeline operands (string literals or `.Field`) — the key is an
arbitrary lowercase name you choose, so `layouts/Layout.html` above can
reference `.title` and `$children` like any other template:

```html
<title>{{.title}}</title>
<main>{{$children}}</main>
```

`$`-prefixed names are reserved for marge-injected values like `$children` —
avoid declaring your own `$`-prefixed template variables inside a component
or layout to prevent them from shadowing marge's own.

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
{{<Card title=.Title url=.URL date=.Date/>}}
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
