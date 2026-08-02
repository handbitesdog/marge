"""Command line interface for marge."""

import argparse
from pathlib import Path

from marge import __version__


def build_parser() -> argparse.ArgumentParser:
    """Construct the top-level argument parser."""
    parser = argparse.ArgumentParser(
        prog="marge",
        description="Build a static site from source content.",
    )
    parser.add_argument(
        "--version",
        action="version",
        version=f"%(prog)s {__version__}",
    )

    subs = parser.add_subparsers(dest="command", required=True)
    build = subs.add_parser("build", help="Build the site.")
    build.add_argument(
        "src",
        nargs="?",
        default=Path("content"),
        type=Path,
        help="Source directory (default: content).",
    )
    build.add_argument(
        "-o",
        "--out",
        default=Path("site"),
        type=Path,
        help="Output directory (default: site).",
    )
    return parser


def build_site(src: Path, out: Path) -> int:
    """Build the site at `src` into `out`. Returns an exit code."""
    print(f"marge: would build {src} -> {out}")
    return 0


def main(argv: list[str] | None = None) -> int:
    """Run the CLI. Returns a process exit code."""
    args = build_parser().parse_args(argv)
    if args.command == "build":
        return build_site(args.src, args.out)
    return 1
