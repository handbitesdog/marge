"""CSS pruning: drop rules whose selectors don't match any rendered HTML."""

import re
from collections.abc import Iterable
from dataclasses import dataclass
from html.parser import HTMLParser

import tinycss2
from tinycss2 import ast

PSEUDO_RE = re.compile(r"::?[a-zA-Z-]+(\([^)]*\))?")
ATTR_RE = re.compile(r"\[[^\]]*\]")
TOKEN_RE = re.compile(r"[.#]?[a-zA-Z][\w-]*")
CONDITIONAL_AT_RULES = frozenset({"media", "supports"})


@dataclass(frozen=True)
class UsedSelectors:
    tags: frozenset[str]
    classes: frozenset[str]
    ids: frozenset[str]


class _SelectorCollector(HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.tags: set[str] = set()
        self.classes: set[str] = set()
        self.ids: set[str] = set()

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        self.tags.add(tag.lower())
        for name, value in attrs:
            if value is None:
                continue
            if name == "class":
                self.classes.update(value.split())
            elif name == "id":
                self.ids.add(value)


def collect_used_selectors(html_documents: Iterable[str]) -> UsedSelectors:
    """Scan already-rendered page HTML for tags/classes/ids in use site-wide."""
    collector = _SelectorCollector()
    for html in html_documents:
        collector.feed(html)
    return UsedSelectors(
        tags=frozenset(collector.tags),
        classes=frozenset(collector.classes),
        ids=frozenset(collector.ids),
    )


def prune_css(css_text: str, used: UsedSelectors) -> str:
    """Return `css_text` with rules whose selectors don't match `used` removed."""
    rules = tinycss2.parse_stylesheet(css_text, skip_comments=True, skip_whitespace=True)
    kept = [node for rule in rules if (node := _prune_rule(rule, used)) is not None]
    return "".join(node.serialize() for node in kept)


def _prune_rule(rule: ast.Node, used: UsedSelectors) -> ast.Node | None:
    if isinstance(rule, ast.QualifiedRule):
        selector = tinycss2.serialize(rule.prelude)
        return rule if _selector_matches(selector, used) else None
    if isinstance(rule, ast.AtRule) and rule.lower_at_keyword in CONDITIONAL_AT_RULES:
        return _prune_conditional_at_rule(rule, used)
    # @font-face, @keyframes, @import, @charset, @page, unknown at-rules,
    # and parse errors: can't safely verify against HTML, always keep.
    return rule


def _prune_conditional_at_rule(
    rule: ast.AtRule, used: UsedSelectors
) -> ast.AtRule | None:
    if rule.content is None:
        return rule
    inner = tinycss2.parse_rule_list(
        rule.content, skip_comments=True, skip_whitespace=True
    )
    kept = [node for r in inner if (node := _prune_rule(r, used)) is not None]
    if not kept:
        return None
    return ast.AtRule(
        rule.source_line,
        rule.source_column,
        rule.at_keyword,
        rule.lower_at_keyword,
        rule.prelude,
        kept,
    )


def _selector_matches(selector: str, used: UsedSelectors) -> bool:
    stripped = ATTR_RE.sub("", PSEUDO_RE.sub("", selector))
    return any(_alternative_matches(alt, used) for alt in stripped.split(","))


def _alternative_matches(alternative: str, used: UsedSelectors) -> bool:
    tokens = TOKEN_RE.findall(alternative)
    return not tokens or all(_token_matches(token, used) for token in tokens)


def _token_matches(token: str, used: UsedSelectors) -> bool:
    if token.startswith("."):
        return token[1:] in used.classes
    if token.startswith("#"):
        return token[1:] in used.ids
    return token.lower() in used.tags
