# marge

A simple static site generator written in Python 3 that outputs HTML/CSS/JS.

## Setup

```bash
poetry install
```

## Usage

```bash
poetry run marge --help
poetry run marge build
```

## Authoring

A site is a `config.toml`, a `templates/` directory, and a `content/`
directory of `.html` files.

### Site config (`config.toml`)

```toml
title = "Jackson's Blog"
base_url = "https://example.com"

# optional overrides — defaults shown
content_dir = "content"
output_dir = "site"
template_dir = "templates"
prune_css = false
prune_css_safelist = []
```

### Content pages

Each content file starts with a `+++`-fenced TOML front-matter block,
followed by plain HTML:

```html
+++
title = "Hello, World"
date = 2026-08-01
layout = "base"
+++
<p>This is my first post, written in plain HTML.</p>
```

`title` and `layout` are required on every page; `date` is required on
pages under `content/posts/` (used to sort the `posts` listing). `slug`
overrides the output filename, and `draft = true` excludes a page from
the build.

Non-post pages can be organized under `content/pages/` — the `pages/`
segment is stripped from the output path and URL, so
`content/pages/about.html` builds to `about.html` (`/about.html`), not
`pages/about.html`. This is purely organizational; pages can also sit
directly under `content/` with the same result.

### Templates

Templates support three tags:

- `{{ page.title }}` — interpolate a value from `site`, `page`, `posts`,
  or (inside a layout) the rendered `content`. Dotted paths do nested
  lookups.
- `{% include "components/nav.html" %}` — inline another template, path
  relative to `template_dir`.
- `{% for post in posts %} ... {% endfor %}` — repeat a block once per
  item.

A content page's layout template (named in its front matter) receives
the page's rendered body as `{{ content }}` — that's the only hook
needed to wrap a page in a shared header/footer.

### CSS pruning

Setting `prune_css = true` drops CSS rules that the site's rendered HTML
never uses, so a large general-purpose stylesheet can be dropped into
`content/` without shipping unused rules to visitors. It's off by default.

Scope is whole-site, not per-page: a rule is kept if its selector could
match an element on *any* page in the site, not just the page the
stylesheet happens to sit next to. `@font-face`, `@keyframes`, `@import`,
`@charset`, `@page`, and any other at-rule that isn't `@media`/`@supports`
are always kept verbatim, since they aren't tied to a matchable selector.
`@media`/`@supports` blocks are kept if at least one rule inside them
survives. Pseudo-classes/elements (`:hover`, `::before`, `:not(...)`) and
attribute selectors (`[data-foo]`) are stripped before matching, so
`.btn:hover` is treated the same as `.btn`.

Because pruning only looks at rendered HTML, a class added purely by
JavaScript (e.g. toggled via `classList.add`) looks unused and would
normally get pruned. List those in `prune_css_safelist` to always keep
them:

```toml
prune_css_safelist = ["is-open", "active"]
```
