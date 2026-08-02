from marge.css import UsedSelectors, collect_used_selectors, prune_css


def test_collect_used_selectors_finds_tags_classes_ids() -> None:
    html = '<div class="card highlight" id="main"><p>Hi</p></div>'
    used = collect_used_selectors([html])
    assert used.tags == {"div", "p"}
    assert used.classes == {"card", "highlight"}
    assert used.ids == {"main"}


def test_collect_used_selectors_unions_across_documents() -> None:
    used = collect_used_selectors(['<div class="a"></div>', '<span class="b"></span>'])
    assert used.classes == {"a", "b"}
    assert used.tags == {"div", "span"}


def test_prune_css_drops_unused_class_rule() -> None:
    used = collect_used_selectors(['<div class="used"></div>'])
    css = ".used { color: red; } .unused { color: blue; }"
    result = prune_css(css, used)
    assert ".used" in result
    assert ".unused" not in result


def test_prune_css_drops_unused_tag_rule() -> None:
    used = collect_used_selectors(["<p>Hi</p>"])
    css = "p { color: red; } aside { color: blue; }"
    result = prune_css(css, used)
    assert "color: red" in result
    assert "color: blue" not in result


def test_prune_css_drops_unused_id_rule() -> None:
    used = collect_used_selectors(['<div id="main"></div>'])
    css = "#main { color: red; } #missing { color: blue; }"
    result = prune_css(css, used)
    assert "#main" in result
    assert "#missing" not in result


def test_prune_css_keeps_rule_if_any_comma_alternative_matches() -> None:
    used = collect_used_selectors(["<p>Hi</p>"])
    css = "p, .missing { color: red; }"
    result = prune_css(css, used)
    assert "color: red" in result


def test_prune_css_media_query_survives_if_one_inner_rule_matches() -> None:
    used = collect_used_selectors(['<div class="used"></div>'])
    css = "@media (min-width: 600px) { .used { color: red; } .unused { color: blue; } }"
    result = prune_css(css, used)
    assert "@media" in result
    assert ".used" in result
    assert ".unused" not in result


def test_prune_css_media_query_dropped_if_no_inner_rules_match() -> None:
    used = collect_used_selectors(["<p>Hi</p>"])
    css = "@media (min-width: 600px) { .unused { color: blue; } }"
    result = prune_css(css, used)
    assert result == ""


def test_prune_css_always_keeps_keyframes() -> None:
    used = UsedSelectors(tags=frozenset(), classes=frozenset(), ids=frozenset())
    css = "@keyframes spin { from { opacity: 0; } to { opacity: 1; } }"
    result = prune_css(css, used)
    assert "@keyframes spin" in result


def test_prune_css_always_keeps_font_face() -> None:
    used = UsedSelectors(tags=frozenset(), classes=frozenset(), ids=frozenset())
    css = '@font-face { font-family: "Example"; src: url(example.woff2); }'
    result = prune_css(css, used)
    assert "@font-face" in result


def test_prune_css_strips_pseudo_class_before_matching() -> None:
    used = collect_used_selectors(['<a class="foo">Link</a>'])
    css = ".foo:hover { color: red; } .bar:hover { color: blue; }"
    result = prune_css(css, used)
    assert ".foo:hover" in result
    assert ".bar:hover" not in result


def test_prune_css_strips_attribute_selector_before_matching() -> None:
    used = collect_used_selectors(['<input class="foo">'])
    css = ".foo[data-active] { color: red; } .bar[data-active] { color: blue; }"
    result = prune_css(css, used)
    assert "color: red" in result
    assert "color: blue" not in result


def test_prune_css_keeps_selector_with_no_extractable_tokens() -> None:
    used = UsedSelectors(tags=frozenset(), classes=frozenset(), ids=frozenset())
    css = ":root { --spacing: 8px; } * { box-sizing: border-box; }"
    result = prune_css(css, used)
    assert ":root" in result
    assert "box-sizing" in result


def test_prune_css_not_with_comma_does_not_leak_as_top_level_or() -> None:
    used = collect_used_selectors(['<div class="a"></div>'])
    css = ".a:not(.b, .c) { color: red; }"
    result = prune_css(css, used)
    assert "color: red" in result
