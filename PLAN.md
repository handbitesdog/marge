# PLAN.md

## Context

`marge` is a brand-new Go static-site generator (the repo currently contains only `CLAUDE.md`). It needs to support reusable, argument-taking components; layouts that are just components; file-based page routing; a generic "collection" mechanism for grouping pages (used for a blog); and Markdown content with front matter for blog posts. This plan is the result of a design session covering the templating approach, URL structure, static assets, and CLI scope, followed by a detailed implementation design pass.

Decisions locked in during planning:
- **Component syntax**: an HTML-tag-like syntax inside `{{}}` — self-closing `{{<name key="val"/>}}` or block `{{<name key="val">}}...{{</name>}}` (not raw Go `html/template` calls), implemented as a thin macro-preprocessor that rewrites these into standard `html/template` source — not a hand-rolled template engine. A bare `<` right after `{{` can never start valid Go template syntax (comparisons are the functions `lt`/`gt`, not infix operators), so it's an unambiguous directive marker without reserving a keyword, and closing tags echo the component name so a mismatched open/close is a caught error rather than a silent misrender.
- **URLs**: pretty URLs everywhere (`about.html` → `/about/`).
- **Static assets**: a `static/` passthrough directory, copied byte-for-byte into the output.
- **CLI**: `marge build` (one-shot) and `marge serve` (build + serve `dist/` over HTTP + rebuild on source change, full rebuild each time, no browser auto-refresh).
- **Module path**: `github.com/handbitesdog/marge`.

## Package layout

```
marge/
  go.mod                        # module github.com/handbitesdog/marge
  main.go                       # CLI entrypoint: build/serve subcommand dispatch
  internal/
    site/
      site.go                    # Page, Collection, CollectionItem, PageData, ItemData
    tmpl/
      preprocess.go               # {{<name/>}} / {{<name>}}...{{</name>}} -> html/template source
      preprocess_test.go
      funcs.go                     # dict + render FuncMap helpers
      components.go                 # load components/*.html, auto-wrap as {{define}}
    content/
      frontmatter.go                 # YAML front matter split + parse
      frontmatter_test.go
      content.go                      # walk content/, goldmark render, build Collections, sort by date desc
    build/
      pages.go                        # walk pages/, classify page vs _item.html, prettyURL()
      pages_test.go
      staticcopy.go                    # byte-for-byte static/ -> dist/ copy
      build.go                          # Options, Run() — orchestrates the full pipeline
      build_test.go                      # integration test against example/
    serve/
      serve.go                            # http.FileServer + fsnotify watch + debounced rebuild
  example/                                # minimal reference site; doubles as build_test.go fixture
    components/{layout.html, card.html}
    pages/{index.html, about.html, blog/{index.html, _item.html}}
    content/blog/{hello-world.md, second-post.md, third-post.md}
    static/css/site.css
```

`internal/` hides implementation packages since nothing here is meant to be imported by other modules. Root-level `main.go` (not `cmd/marge/`) since this is a single binary.

## Key types (`internal/site/site.go`)

```go
type Page struct {
    SourcePath string // relative to pages/, e.g. "blog/index.html"
    OutputPath string // relative to dist/, e.g. "blog/index.html"
    URL        string // e.g. "/blog/"
}

type CollectionItem struct {
    Title   string
    Date    time.Time
    Slug    string
    URL     string
    Content template.HTML  // goldmark-rendered body
    Extra   map[string]any // front matter fields beyond title/date
}

type Collection struct {
    Name  string
    Items []CollectionItem // sorted by Date descending
}

type PageData struct {
    Collections map[string]Collection // keyed by content/ subdirectory name
}

type ItemData struct {
    CollectionItem
    Collections map[string]Collection
}
```

`html/template` (not `text/template`) is used throughout so plain output stays auto-escaped, while `Content` and `render()`'s return type are `template.HTML` to deliberately allow pre-rendered, trusted HTML through.

## Component syntax and macro-preprocessor

Authors write, in any file under `components/` or `pages/`:
- Self-closing: `{{<name key1="val1" key2="val2"/>}}`
- Block (captures children as pre-rendered HTML): `{{<name key1="val1">}} ...markup... {{</name>}}`

A space between `{{` and `<`, and between `<` and the name, is optional — `{{ <name/> }}` and `{{< name/>}}` both mean the same as `{{<name/>}}`. Spaces around `=` in `key = "val"` are likewise optional.

`key=val` values are plain Go template pipeline operands (`"literal"` or `.Title`) passed through verbatim — no expression parser needed.

`Preprocess(src, sourceName string) (string, error)` in `internal/tmpl/preprocess.go`:
1. Scan for `{{`; skip whitespace and peek the next character — only a leading `<` marks a directive, everything else (`.Title`, `range`, etc.) passes through untouched. This is unambiguous because no valid Go template action can begin with `<` — comparisons are the functions `lt`/`gt`, not infix operators.
2. For a directive, scan forward **quote-aware** to the matching `}}` (so a quoted value containing `}` or `>` isn't mistaken for the tag's end).
3. Manually tokenize: consume `<`; if the next character is `/`, this is a closing tag — read the name up to `>`. Otherwise read the component name (bare identifier up to whitespace, `/`, or `>`), then `key=value` pairs (quoted string or bare token up to whitespace). The tag is self-closing if the last non-whitespace character before `}}` is `/`; otherwise it opens a block awaiting `{{</name>}}`.
4. Maintain a stack of open blocks for nested `{{<name>}}...{{</name>}}`. Each frame records the component name and accumulates its body in a `strings.Builder`; on `{{</name>}}` the name must match the top of the stack (error otherwise — catches mismatched open/close tags), then emit into the parent frame (or top-level output):
   ```
   {{define "BLOCKNAME"}}BODY{{end}}{{template "NAME" (dict K1 V1 ... "Children" (render "BLOCKNAME" .))}}
   ```
   Self-closing tags emit directly with `"Children" ""` (must be present explicitly — a missing map key renders as literal `<no value>` in Go templates).
5. `blockName` = `__block_<sanitized sourceName>_<counter>` — collision-free across files and nesting depth.
6. Error at EOF on unmatched/unclosed blocks, and on a closing tag whose name doesn't match the innermost open tag.

`internal/tmpl/funcs.go` provides `dict(pairs ...any) (map[string]any, error)` and a `Renderer` with `renderFunc(name string, data any) (template.HTML, error)` that calls `ExecuteTemplate` on the shared template set. Because `render()` needs the *fully parsed* template set but `Funcs()` must be registered *before* parsing, `Renderer.tset` is a pointer field set once after all parsing completes — `render`'s closure only dereferences it at execution time. `render(name, .)` passes the **ambient dot at the call site**, so a captured children block can still reference the outer page's fields.

`internal/tmpl/components.go`: `LoadComponents(root *template.Template, componentsDir string) error` reads each `components/<name>.html` (flat, non-recursive), preprocesses it, wraps as `{{define "<name>"}}...{{end}}`, and parses it into the shared `root` set. Component names never collide with page names since pages are registered under their full relative path including `.html`.

## Collections

**Convention**: `pages/<collection>/_item.html` is the per-item template for the collection named `<collection>` (i.e. `content/<collection>/`). Any `pages/` file whose basename starts with `_` is excluded from ordinary routing; `_item.html` is specifically recognized and tied to its parent directory's collection name. It's preprocessed and parsed like any other page, just executed once per item with `site.ItemData` as its data context. If `content/<name>/` has `.md` files but no matching `_item.html` exists, `build.Run` fails fast with a clear error.

**Data exposure**: every discovered collection is exposed to *every* page and item template as `.Collections` (a `map[string]site.Collection`), so a homepage can list latest posts the same way a blog index does, and a future second collection needs zero new code. `.Collections.blog.Items` works via Go's map-index-via-dot; a hyphenated collection name would need `{{index .Collections "blog-posts"}}` instead.

No auto-generated excerpts — an author who wants one adds `excerpt: "..."` to front matter and reads `{{.Extra.excerpt}}`; `Extra` already captures arbitrary front matter via YAML's inline-map tag.

`internal/content/frontmatter.go`: splits leading `---`/`---` YAML block from the Markdown body.
```go
type frontMatter struct {
    Title string         `yaml:"title"`
    Date  time.Time      `yaml:"date"`
    Extra map[string]any `yaml:",inline"`
}
```
`yaml.v3` resolves unquoted ISO-8601 dates into `time.Time` directly. Missing/empty `Title` or zero `Date` → build error naming the file.

`internal/content/content.go`: `LoadAll(contentDir string) (map[string]site.Collection, error)` — each subdirectory of `content/` (flat, one level, `.md` files only) becomes one collection; goldmark renders the body with `html.WithUnsafe()` (content is author-trusted, same trust model as Jekyll/Hugo — without it, raw HTML embedded in a post's Markdown gets silently stripped). Items sorted by `Date` descending.

## Build pipeline (`internal/build/build.go`)

```go
type Options struct{ SrcDir, DistDir string }
func Run(opts Options) error
```
1. `os.RemoveAll(DistDir)` then recreate — a **clean** build, so deleted posts/pages don't leave stale output behind.
2. Create `root := template.New("root").Funcs(renderer.FuncMap())`.
3. `tmpl.LoadComponents(root, SrcDir/components)`.
4. `content.LoadAll(SrcDir/content)` → collections.
5. `build.Discover(SrcDir/pages)` → ordinary pages vs. `_item.html` paths keyed by collection name; preprocess + parse all of them into the same shared `root` set.
6. Set `renderer.tset = root`.
7. Execute each ordinary page with `site.PageData{Collections}`, write to `dist/<prettyOutputPath>`.
8. For each collection, execute its `_item.html` once per item with `site.ItemData{...}`, write to `dist/<name>/<slug>/index.html`.
9. `build.CopyStatic(SrcDir/static, DistDir)` — no-op if `static/` doesn't exist.

**Pretty URL mapping** (`internal/build/pages.go`, `prettyURL(relPath) (outputPath, url string)`):
| relPath | outputPath | url |
|---|---|---|
| `index.html` | `index.html` | `/` |
| `about.html` | `about/index.html` | `/about/` |
| `blog/index.html` | `blog/index.html` | `/blog/` |

Collection items follow the same shape directly: `dist/<collection>/<slug>/index.html`.

`internal/build/staticcopy.go`: `CopyStatic(srcDir, destDir string) error` — recursive copy preserving file mode; no-op if source is absent.

## Serve (`internal/serve/serve.go`)

```go
func Run(src, dist, addr string) error
```
1. Run the full build once.
2. Start `http.FileServer(http.Dir(dist))` in a goroutine.
3. `fsnotify` watcher: recursively add a watch per subdirectory under `components/`, `pages/`, `content/`, `static/` (fsnotify only watches one level per `Add`); new directories created at runtime (e.g. a new collection) get watched automatically.
4. On any event, debounce (~150ms, editors fire multiple events per save) then rebuild; a rebuild error is logged, not fatal — a bad edit must not kill the server.
5. Clean shutdown on `SIGINT` via `signal.NotifyContext`.

Deliberately no atomic dist-directory swap (build-to-temp-then-rename) — a request could theoretically see a half-written file mid-rebuild, but this matches the earlier "simplicity over incremental/atomic rebuilds" decision. Easy to add later if it matters.

## Dependencies

```
github.com/yuin/goldmark v1.7.x
gopkg.in/yaml.v3 v3.0.x
github.com/fsnotify/fsnotify v1.7.x
```
Everything else (CLI dispatch, HTTP server, file walking, templating) is stdlib — no CLI framework, no router, no websocket/live-reload library.

## File-by-file build order

1. `go.mod`
2. `internal/site/site.go`
3. `internal/tmpl/preprocess.go` (+ test)
4. `internal/tmpl/funcs.go`
5. `internal/tmpl/components.go`
6. `internal/content/frontmatter.go` (+ test)
7. `internal/content/content.go`
8. `internal/build/pages.go` (+ test)
9. `internal/build/staticcopy.go`
10. `internal/build/build.go` (+ integration test)
11. `example/` fixture site
12. `internal/serve/serve.go`
13. `main.go`

## Verification

**Unit tests** (`go test ./...`):
- `Preprocess`: self-closing components, nested blocks, unmatched/unclosed-block errors, mismatched closing-tag name, quoted value containing `}` or `>`.
- `prettyURL`: table-driven over the cases above.
- `splitFrontMatter` / date parsing: valid front matter, missing title/date.
- `internal/build/build_test.go`: integration test running `build.Run` against `example/` into a `t.TempDir()`, asserting `index.html`, `about/index.html`, `blog/index.html`, `blog/hello-world/index.html` exist with expected content; posts appear newest-first on the listing page; `css/site.css` matches byte-for-byte; no `blog/_item/` directory was created.

**Manual end-to-end**:
```bash
go run . build example dist_out
go run . serve example dist_out :8080
```
Then check `/`, `/about/`, `/blog/`, `/blog/hello-world/`, `/css/site.css` return 200 with expected content; edit a file under `example/content/blog/` while `serve` runs and confirm rebuilt output appears within ~1s; Ctrl+C for clean shutdown.

## Noted trade-offs (accepted, not oversights)

- No `{{-`/`-}}` whitespace-trim support on the custom directives themselves — control whitespace in surrounding markup instead.
- No nested collections (`content/<name>/` must be flat).
- No atomic dist swap during `serve` rebuilds.
- No site-wide config object (base URL, site title) — page titles are passed as literal component args.