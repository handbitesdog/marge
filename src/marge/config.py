"""Site-wide configuration loaded from config.toml."""

import tomllib
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class SiteConfig:
    title: str
    base_url: str
    content_dir: Path = Path("content")
    output_dir: Path = Path("site")
    template_dir: Path = Path("templates")
    prune_css: bool = False
    prune_css_safelist: frozenset[str] = frozenset()


def load_config(path: Path) -> SiteConfig:
    """Parse `path` (a config.toml) into a SiteConfig."""
    data = tomllib.loads(path.read_text())
    return SiteConfig(
        title=data["title"],
        base_url=data["base_url"],
        content_dir=Path(data.get("content_dir", "content")),
        output_dir=Path(data.get("output_dir", "site")),
        template_dir=Path(data.get("template_dir", "templates")),
        prune_css=data.get("prune_css", False),
        prune_css_safelist=frozenset(data.get("prune_css_safelist", [])),
    )
