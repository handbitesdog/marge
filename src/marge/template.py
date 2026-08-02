"""A minimal template engine: `{{ var }}`, `{% include %}`, `{% for %}`.

Rendering is a single recursive-descent pass over the template text: the
next tag is located, handled, and its expansion is itself scanned for more
tags. This is what lets `{% include %}` calls work correctly inside a
`{% for %}` body, and vice versa — a naive sequence of separate global
substitution passes would only give includes and loops a single pass each.

Nested `{% for %}` loops are not supported: a `{% for %}` always closes at
the next `{% endfor %}` in the text, whichever loop it belongs to.
"""

import re
from pathlib import Path
from typing import Any

TAG_RE = re.compile(r"\{\{(.*?)\}\}|\{%\s*(\w+)(.*?)%\}", re.DOTALL)
INCLUDE_ARG_RE = re.compile(r'\s*"([^"]*)"\s*')
FOR_HEADER_RE = re.compile(r"\s*(\w+)\s+in\s+([\w.]+)\s*")
ENDFOR_RE = re.compile(r"\{%\s*endfor\s*%\}")


class TemplateError(Exception):
    """Raised for invalid template syntax, an unknown variable, or a cycle."""


def render(
    template: str,
    context: dict[str, Any],
    template_dir: Path,
    _including: frozenset[Path] = frozenset(),
) -> str:
    """Render `template` against `context`, resolving includes under `template_dir`."""
    out: list[str] = []
    pos = 0
    while pos < len(template):
        match = TAG_RE.search(template, pos)
        if match is None:
            out.append(template[pos:])
            break
        out.append(template[pos : match.start()])

        var_expr, tag_name, tag_arg = match.groups()
        if var_expr is not None:
            out.append(str(_lookup(var_expr.strip(), context)))
            pos = match.end()
        elif tag_name == "include":
            path = _parse_include_arg(tag_arg)
            out.append(_render_include(path, context, template_dir, _including))
            pos = match.end()
        elif tag_name == "for":
            body, end_pos = _extract_for_body(template, match.end())
            out.append(_render_for(tag_arg, body, context, template_dir, _including))
            pos = end_pos
        elif tag_name == "endfor":
            raise TemplateError("'{% endfor %}' without matching '{% for %}'")
        else:
            raise TemplateError(f"unknown tag: {match.group(0)!r}")

    return "".join(out)


def _lookup(path: str, context: dict[str, Any]) -> Any:
    value: Any = context
    for part in path.split("."):
        if not isinstance(value, dict) or part not in value:
            raise TemplateError(f"undefined variable: {path!r}")
        value = value[part]
    return value


def _parse_include_arg(raw: str) -> str:
    match = INCLUDE_ARG_RE.fullmatch(raw)
    if match is None:
        raise TemplateError(f"'{{% include %}}' expects a quoted path, got: {raw!r}")
    return match.group(1)


def _render_include(
    rel_path: str,
    context: dict[str, Any],
    template_dir: Path,
    including: frozenset[Path],
) -> str:
    path = (template_dir / rel_path).resolve()
    if path in including:
        raise TemplateError(f"include cycle detected at {rel_path!r}")
    text = path.read_text()
    return render(text, context, template_dir, including | {path})


def _extract_for_body(template: str, body_start: int) -> tuple[str, int]:
    end_match = ENDFOR_RE.search(template, body_start)
    if end_match is None:
        raise TemplateError("'{% for %}' without matching '{% endfor %}'")
    return template[body_start : end_match.start()], end_match.end()


def _render_for(
    header: str,
    body: str,
    context: dict[str, Any],
    template_dir: Path,
    including: frozenset[Path],
) -> str:
    match = FOR_HEADER_RE.fullmatch(header)
    if match is None:
        raise TemplateError(f"invalid '{{% for %}}' header: {header!r}")
    loop_var, list_expr = match.group(1), match.group(2)
    items = _lookup(list_expr, context)

    rendered = []
    for item in items:
        loop_context = {**context, loop_var: item}
        rendered.append(render(body, loop_context, template_dir, including))
    return "".join(rendered)
