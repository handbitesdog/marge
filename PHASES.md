# PHASES.md

Phased build plan derived from [PLAN.md](PLAN.md). Each phase is meant to be
handed to a model as its own task, in its own session.

## How to use this

For each phase:
1. Start a fresh session/conversation.
2. Give the model `PLAN.md` (full design) and this file, and say: "Implement
   Phase N only, per PHASES.md."
3. Run that phase's verification commands yourself before starting the next
   phase. Do not let a model start Phase N+1 until Phase N's verification
   passes.
4. Each phase assumes all prior phases are already committed in the repo —
   the model should build on existing code, not restate it.

Phase boundaries follow package boundaries in PLAN.md's "Package layout" and
"File-by-file build order", regrouped so each phase (a) has its own
compile/test gate and (b) doesn't need files from a later phase to verify
itself. Where a phase's package has no dedicated test file in PLAN.md (e.g.
`components.go`, `content.go`, `staticcopy.go`), the model should still write
one using `t.TempDir()` fixtures — don't wait for the `example/` site, which
doesn't land until Phase 6.

---

## Phase 1 — Module scaffolding & core types

**Files**: `go.mod`, `internal/site/site.go`

**Scope**: `go mod init github.com/handbitesdog/marge`. Define `Page`,
`CollectionItem`, `Collection`, `PageData`, `ItemData` exactly as specified
in PLAN.md's "Key types" section. No logic yet — this is pure data
structures.

**Reference**: PLAN.md § "Package layout", § "Key types (`internal/site/site.go`)".

**Verification**:
```bash
go build ./...
go vet ./...
```
No tests expected — just confirm it compiles and the struct fields/types
match PLAN.md exactly (e.g. `Content template.HTML`, `Extra map[string]any`).

---

## Phase 2 — Macro preprocessor

**Files**: `internal/tmpl/preprocess.go`, `internal/tmpl/preprocess_test.go`

**Scope**: `Preprocess(src, sourceName string) (string, error)` implementing
the full component-tag scan/tokenize/stack algorithm.

**Reference**: PLAN.md § "Component syntax and macro-preprocessor" — the
5-step algorithm is fully specified there, including the `blockName`
collision-free naming scheme and the `dict`/`render` call shape it must
emit.

**Required test cases** (from PLAN.md § "Verification"):
- self-closing component
- nested blocks
- unmatched/unclosed block → error
- mismatched closing-tag name → error
- quoted value containing `}` or `>` (must not be mistaken for tag end)

**Verification**:
```bash
go test ./internal/tmpl/... -run Preprocess -v
```

---

## Phase 3 — Template funcs & component loader

**Files**: `internal/tmpl/funcs.go`, `internal/tmpl/components.go` (+ tests)

**Scope**: `dict(pairs ...any) (map[string]any, error)`; `Renderer` with
`renderFunc` and the deferred-pointer `tset` field (set once after parsing
completes, dereferenced only at execution time); `LoadComponents(root
*template.Template, componentsDir string) error`.

**Reference**: PLAN.md § "Component syntax and macro-preprocessor",
paragraphs on `internal/tmpl/funcs.go` and `internal/tmpl/components.go`.

**Note**: `example/components/` doesn't exist until Phase 6 — write test
fixtures with `t.TempDir()` (a couple of small component `.html` files) to
exercise `LoadComponents`.

**Verification**:
```bash
go test ./internal/tmpl/...
```

---

## Phase 4 — Content pipeline (front matter + Markdown)

**Files**: `internal/content/frontmatter.go` (+ test),
`internal/content/content.go` (+ test)

**Scope**: `frontMatter` struct and YAML split/parse; `LoadAll(contentDir
string) (map[string]site.Collection, error)` walking one level of
subdirectories, rendering Markdown via goldmark with `html.WithUnsafe()`,
sorting items by `Date` descending.

**Reference**: PLAN.md § "Collections", including the `frontMatter` struct
and the missing-title/zero-date error requirement.

**New dependencies**:
```bash
go get gopkg.in/yaml.v3@v3.0
go get github.com/yuin/goldmark@v1.7
```

**Test cases**: valid front matter parses; missing title → error; missing/zero
date → error; multiple `.md` files in one subdirectory sort newest-first.

**Verification**:
```bash
go test ./internal/content/...
```

---

## Phase 5 — Page discovery & static asset copy

**Files**: `internal/build/pages.go` (+ test), `internal/build/staticcopy.go`
(+ test)

**Scope**: `prettyURL(relPath) (outputPath, url string)`; page-vs-`_item.html`
classification (any `pages/` file whose basename starts with `_` is excluded
from ordinary routing); `CopyStatic(srcDir, destDir string) error`
(recursive, preserves file mode, no-op if source absent).

**Reference**: PLAN.md § "Build pipeline", the "Pretty URL mapping" table,
and the `staticcopy.go` description.

**Test cases**: table-driven `prettyURL` over `index.html`, `about.html`,
`blog/index.html`; `CopyStatic` byte-for-byte copy with mode preserved;
no-op when `static/` is absent.

**Verification**:
```bash
go test ./internal/build/...
```

---

## Phase 6 — Build orchestration + example fixture site

**Files**: `internal/build/build.go` (+ test), `example/` fixture site
(components, pages, content, static — full tree per PLAN.md's "Package
layout")

**Scope**: `Options{SrcDir, DistDir}` and `Run(opts Options) error`
implementing the 9-step pipeline: clean dist, create root template set,
`LoadComponents`, `content.LoadAll`, `Discover` pages, parse everything into
the shared set, set `renderer.tset`, execute ordinary pages, execute
collection items, `CopyStatic`.

**Reference**: PLAN.md § "Build pipeline (`internal/build/build.go`)"
(steps 1–9) and § "Package layout" for the exact `example/` tree to build
(3 blog posts, `layout.html`/`card.html` components, `index.html`/
`about.html`/`blog/index.html`/`blog/_item.html` pages, `static/css/site.css`).

**Verification**:
```bash
go test ./...
```
The integration test in `build_test.go` should assert (per PLAN.md §
"Verification"): `index.html`, `about/index.html`, `blog/index.html`,
`blog/hello-world/index.html` exist with expected content; posts appear
newest-first on the listing page; `css/site.css` matches byte-for-byte; no
`blog/_item/` directory was created.

This is the first point where the whole pipeline runs end-to-end — treat a
green `go test ./...` here as the main manual-review checkpoint before
moving on. There's no CLI yet, so full manual browser verification waits
for Phase 8.

---

## Phase 7 — Dev server

**Files**: `internal/serve/serve.go`

**Scope**: `Run(src, dist, addr string) error` — run a build, serve `dist/`
via `http.FileServer`, `fsnotify` watch (recursive `Add` per subdirectory
under `components/`, `pages/`, `content/`, `static/`, with new directories
watched as they appear), debounce (~150ms) then rebuild on change (log, not
fatal, on rebuild error), clean shutdown on `SIGINT` via
`signal.NotifyContext`.

**Reference**: PLAN.md § "Serve (`internal/serve/serve.go`)".

**New dependency**:
```bash
go get github.com/fsnotify/fsnotify@v1.7
```

**Verification**:
```bash
go build ./...
```
No CLI exists yet to drive this manually, so behavioral verification
(rebuild-on-edit, clean shutdown) is deferred to Phase 8. This phase's gate
is just a clean compile plus a code-level check that it matches the 5 steps
in PLAN.md.

---

## Phase 8 — CLI entrypoint & full manual verification

**Files**: `main.go`

**Scope**: subcommand dispatch for `marge build <src> <dist>` and `marge
serve <src> <dist> <addr>`.

**Reference**: PLAN.md § "Decisions locked in during planning" (CLI bullet).

**Verification**:
```bash
go build ./...
go test ./...
```
Then the full manual end-to-end from PLAN.md § "Verification":
```bash
go run . build example dist_out
go run . serve example dist_out :8080
```
- `/`, `/about/`, `/blog/`, `/blog/hello-world/`, `/css/site.css` all return
  200 with expected content
- edit a file under `example/content/blog/` while `serve` is running;
  confirm rebuilt output appears within ~1s
- Ctrl+C shuts down cleanly

This is the final checkpoint — once it passes, the project matches PLAN.md
in full.
