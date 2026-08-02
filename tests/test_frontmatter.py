import pytest

from marge.frontmatter import FrontMatterError, parse


def test_parse_valid_front_matter() -> None:
    text = '+++\ntitle = "Hi"\nlayout = "base"\n+++\n<p>Body</p>\n'
    metadata, body = parse(text)
    assert metadata == {"title": "Hi", "layout": "base"}
    assert body == "<p>Body</p>\n"


def test_parse_missing_fence() -> None:
    with pytest.raises(FrontMatterError, match="missing"):
        parse("<p>No front matter</p>")


def test_parse_unterminated_fence() -> None:
    with pytest.raises(FrontMatterError, match="unterminated"):
        parse('+++\ntitle = "Hi"\n')


def test_parse_invalid_toml() -> None:
    with pytest.raises(FrontMatterError, match="invalid TOML"):
        parse("+++\nnot valid toml ===\n+++\nbody")


def test_parse_reserved_key() -> None:
    with pytest.raises(FrontMatterError, match="reserved"):
        parse('+++\ncontent = "oops"\n+++\nbody')
