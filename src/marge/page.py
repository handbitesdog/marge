"""Discovery of content pages and their output locations."""

from dataclasses import dataclass
from pathlib import Path
from typing import Any

from marge.frontmatter import parse as parse_front_matter

REQUIRED_FIELDS = ("title", "layout")


class PageError(Exception):
    """Raised when a content page is missing required front matter."""


@dataclass
class Page:
    source: Path
    metadata: dict[str, Any]
    body: str
    rel_path: Path
    url: str

    @property
    def title(self) -> str:
        return str(self.metadata["title"])

    @property
    def layout(self) -> str:
        return str(self.metadata["layout"])

    @property
    def is_draft(self) -> bool:
        return bool(self.metadata.get("draft", False))


def discover_pages(content_dir: Path) -> list[Page]:
    """Find and parse every `.html` content file under `content_dir`."""
    return [_load_page(path, content_dir) for path in sorted(content_dir.rglob("*.html"))]


def _load_page(path: Path, content_dir: Path) -> Page:
    metadata, body = parse_front_matter(path.read_text(), source=str(path))

    missing = [field for field in REQUIRED_FIELDS if field not in metadata]
    if missing:
        names = ", ".join(missing)
        raise PageError(f"{path}: missing required front matter field(s): {names}")

    rel_path = path.relative_to(content_dir)
    slug = metadata.get("slug")
    if slug is not None:
        rel_path = rel_path.parent / f"{slug}.html"

    return Page(
        source=path,
        metadata=metadata,
        body=body,
        rel_path=rel_path,
        url="/" + rel_path.as_posix(),
    )
