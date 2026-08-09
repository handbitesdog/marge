# marge

A simple static-site generator written in Go, with reusable HTML components,
file-based page routing, and a Markdown content pipeline for things like
blog posts.

Full design is in [PLAN.md](PLAN.md). The build is broken into phases,
tracked in [PHASES.md](PHASES.md) — see that file for current progress.

## Status

All phases in [PHASES.md](PHASES.md) are complete. `marge` builds and serves
sites via the CLI described below.

## Requirements

- Go 1.25+

## Building

```bash
go build ./...
go vet ./...
```

## Testing

```bash
go test ./...
```

## Usage

```bash
marge build <src> <dist>          # one-shot build
marge serve <src> <dist> <addr>   # build, serve dist/, rebuild on change
```

A `marge` site is expected to look like:

```
mysite/
  components/   # reusable HTML components (layouts are just components)
  pages/        # routed pages, mapped to pretty URLs (about.html -> /about/)
  content/      # Markdown content with YAML front matter, grouped into collections
  static/       # assets copied byte-for-byte into the output
```

See [PLAN.md](PLAN.md) for the full design, including component syntax,
collections, and URL mapping.
