"""Parsing of `+++`-fenced TOML front matter in content files."""

import tomllib
from typing import Any

FENCE = "+++"

# Injected by the build at render time; front matter must not shadow them.
RESERVED_KEYS = {"content", "site", "posts", "page"}


class FrontMatterError(Exception):
    """Raised when a content file's front matter is missing or invalid."""


def parse(text: str, source: str = "<string>") -> tuple[dict[str, Any], str]:
    """Split `text` into (metadata, body) at its `+++` front-matter fence."""
    if not text.startswith(FENCE):
        raise FrontMatterError(f"{source}: missing '{FENCE}' front matter fence")

    parts = text.split(FENCE, 2)
    if len(parts) < 3:
        raise FrontMatterError(f"{source}: unterminated '{FENCE}' front matter fence")

    _, raw_metadata, body = parts
    try:
        metadata = tomllib.loads(raw_metadata)
    except tomllib.TOMLDecodeError as exc:
        raise FrontMatterError(f"{source}: invalid TOML front matter: {exc}") from exc

    reserved = RESERVED_KEYS & metadata.keys()
    if reserved:
        names = ", ".join(sorted(reserved))
        raise FrontMatterError(
            f"{source}: front matter may not set reserved key(s): {names}"
        )

    return metadata, body.removeprefix("\n")
