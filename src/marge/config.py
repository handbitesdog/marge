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


def load_config(path: Path) -> SiteConfig:
    """Parse `path` (a config.toml) into a SiteConfig."""
    data = tomllib.loads(path.read_text())
    return SiteConfig(
        title=data["title"],
        base_url=data["base_url"],
        content_dir=Path(data.get("content_dir", "content")),
        output_dir=Path(data.get("output_dir", "site")),
        template_dir=Path(data.get("template_dir", "templates")),
    )
