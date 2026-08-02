from pathlib import Path

import pytest

from marge.build import BuildError, build_site
from marge.config import SiteConfig


@pytest.fixture
def project(tmp_path: Path) -> tuple[Path, Path, Path, SiteConfig]:
    content = tmp_path / "content"
    templates = tmp_path / "templates"
    out = tmp_path / "site"
    (content / "posts").mkdir(parents=True)
    templates.mkdir()

    (templates / "base.html").write_text(
        "<title>{{ page.title }}</title><body>{{ content }}</body>"
    )
    (content / "posts" / "hello.html").write_text(
        '+++\ntitle = "Hello"\ndate = 2026-01-01\nlayout = "base"\n+++\n<p>Hi</p>\n'
    )
    (content / "index.html").write_text(
        '+++\ntitle = "Home"\nlayout = "base"\n+++\n'
        "{% for post in posts %}{{ post.title }}{% endfor %}\n"
    )
    (content / "style.css").write_text("body { color: red; }")

    config = SiteConfig(
        title="Test Site",
        base_url="https://example.com",
        template_dir=templates,
    )
    return content, templates, out, config


def test_build_renders_pages_and_copies_assets(
    project: tuple[Path, Path, Path, SiteConfig],
) -> None:
    content, _templates, out, config = project
    build_site(config, content, out)

    assert "Hello" in (out / "index.html").read_text()
    assert "<p>Hi</p>" in (out / "posts" / "hello.html").read_text()
    assert (out / "style.css").read_text() == "body { color: red; }"


def test_build_skips_drafts(project: tuple[Path, Path, Path, SiteConfig]) -> None:
    content, _templates, out, config = project
    (content / "posts" / "draft.html").write_text(
        '+++\ntitle = "Draft"\ndate = 2026-01-02\nlayout = "base"\ndraft = true\n+++\n'
        "<p>Shh</p>\n"
    )
    build_site(config, content, out)
    assert not (out / "posts" / "draft.html").exists()


def test_build_missing_date_on_post_raises(
    project: tuple[Path, Path, Path, SiteConfig],
) -> None:
    content, _templates, out, config = project
    (content / "posts" / "no-date.html").write_text(
        '+++\ntitle = "No Date"\nlayout = "base"\n+++\n<p>Oops</p>\n'
    )
    with pytest.raises(BuildError, match="date"):
        build_site(config, content, out)
