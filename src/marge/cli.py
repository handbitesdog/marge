"""Command line interface for marge."""

import argparse
import sys
from pathlib import Path

from marge import __version__
from marge.build import BuildError, build_site
from marge.config import load_config


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
    build.add_argument(
        "-c",
        "--config",
        default=Path("config.toml"),
        type=Path,
        help="Site config file (default: config.toml).",
    )
    return parser


def main(argv: list[str] | None = None) -> int:
    """Run the CLI. Returns a process exit code."""
    args = build_parser().parse_args(argv)
    if args.command != "build":
        return 1

    config = load_config(args.config)
    try:
        build_site(config, args.src, args.out)
    except BuildError as exc:
        print(f"marge: {exc}", file=sys.stderr)
        return 1
    return 0
