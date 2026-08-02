from pathlib import Path

import pytest

from marge.template import TemplateError, render


def test_interpolation() -> None:
    assert render("Hello, {{ name }}!", {"name": "World"}, Path(".")) == "Hello, World!"


def test_dotted_lookup() -> None:
    context = {"page": {"title": "Home"}}
    assert render("{{ page.title }}", context, Path(".")) == "Home"


def test_undefined_variable_raises() -> None:
    with pytest.raises(TemplateError, match="undefined variable"):
        render("{{ missing }}", {}, Path("."))


def test_for_loop() -> None:
    context = {"items": ["a", "b", "c"]}
    template = "{% for item in items %}[{{ item }}]{% endfor %}"
    assert render(template, context, Path(".")) == "[a][b][c]"


def test_for_without_endfor_raises() -> None:
    with pytest.raises(TemplateError, match="endfor"):
        render("{% for item in items %}", {"items": []}, Path("."))


def test_stray_endfor_raises() -> None:
    with pytest.raises(TemplateError, match="endfor"):
        render("{% endfor %}", {}, Path("."))


def test_include(tmp_path: Path) -> None:
    (tmp_path / "partial.html").write_text("Partial says {{ msg }}")
    result = render('{% include "partial.html" %}', {"msg": "hi"}, tmp_path)
    assert result == "Partial says hi"


def test_include_inside_for_loop(tmp_path: Path) -> None:
    (tmp_path / "item.html").write_text("<{{ item }}>")
    template = '{% for item in items %}{% include "item.html" %}{% endfor %}'
    result = render(template, {"items": ["a", "b"]}, tmp_path)
    assert result == "<a><b>"


def test_include_cycle_detected(tmp_path: Path) -> None:
    (tmp_path / "a.html").write_text('{% include "b.html" %}')
    (tmp_path / "b.html").write_text('{% include "a.html" %}')
    with pytest.raises(TemplateError, match="cycle"):
        render('{% include "a.html" %}', {}, tmp_path)
