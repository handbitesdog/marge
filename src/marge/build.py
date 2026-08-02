"""Orchestrates a full site build: discover, render, and write pages."""

from pathlib import Path
from typing import Any

from marge.config import SiteConfig
from marge.css import UsedSelectors, collect_used_selectors, prune_css
from marge.frontmatter import FrontMatterError
from marge.page import Page, PageError, discover_pages
from marge.template import TemplateError, render


class BuildError(Exception):
    """Raised when the site fails to build; message is user-facing."""


def build_site(config: SiteConfig, src: Path, out: Path) -> None:
    """Build the site described by `config`, reading `src` and writing `out`."""
    try:
        pages = [page for page in discover_pages(src) if not page.is_draft]
    except (FrontMatterError, PageError) as exc:
        raise BuildError(str(exc)) from exc

    posts = _build_posts_context(pages, src)
    site = {"title": config.title, "base_url": config.base_url}

    rendered = [
        _render_page(page, site, posts, config.template_dir, out) for page in pages
    ]

    used = None
    if config.prune_css:
        scanned = collect_used_selectors(rendered)
        used = UsedSelectors(
            tags=scanned.tags,
            classes=scanned.classes | config.prune_css_safelist,
            ids=scanned.ids,
        )

    _copy_assets(src, out, used)


def _build_posts_context(pages: list[Page], content_dir: Path) -> list[dict[str, Any]]:
    posts_dir = content_dir / "posts"
    posts = [page for page in pages if posts_dir in page.source.parents]

    for post in posts:
        if "date" not in post.metadata:
            raise BuildError(f"{post.source}: posts require a 'date' front matter field")

    posts.sort(key=lambda page: page.metadata["date"], reverse=True)
    return [{**post.metadata, "url": post.url} for post in posts]


def _render_page(
    page: Page,
    site: dict[str, Any],
    posts: list[dict[str, Any]],
    template_dir: Path,
    out: Path,
) -> str:
    page_context = {**page.metadata, "url": page.url}
    context: dict[str, Any] = {"site": site, "page": page_context, "posts": posts}

    try:
        content = render(page.body, context, template_dir)
        layout_path = template_dir / f"{page.layout}.html"
        layout_context = {**context, "content": content}
        html = render(layout_path.read_text(), layout_context, template_dir)
    except (TemplateError, FileNotFoundError) as exc:
        raise BuildError(f"{page.source}: {exc}") from exc

    dest = out / page.rel_path
    dest.parent.mkdir(parents=True, exist_ok=True)
    dest.write_text(html)
    return html


def _copy_assets(src: Path, out: Path, used: UsedSelectors | None = None) -> None:
    for path in src.rglob("*"):
        if path.is_file() and path.suffix != ".html":
            dest = out / path.relative_to(src)
            dest.parent.mkdir(parents=True, exist_ok=True)
            if used is not None and path.suffix == ".css":
                dest.write_text(prune_css(path.read_text(), used))
            else:
                dest.write_bytes(path.read_bytes())
