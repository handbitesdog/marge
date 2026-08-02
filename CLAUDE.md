# Project overview

marge is a simple static site generator written in Python 3 that outputs
HTML/CSS/JS. The focus is on clean syntax and producing performant and readable
code. This will be primarily used for my blog and browser tools.

# Environment

- Dependencies are managed with [Poetry](https://python-poetry.org/) (2.x).
- Requires Python ≥3.11 (stdlib `tomllib`). Locally installed: 3.14.6.
- The virtualenv lives at `.venv/` in the project root (`virtualenvs.in-project`,
  set in `poetry.toml`).
- Never `pip install` into `.venv` directly — use `poetry add` so `poetry.lock`
  stays accurate. `poetry.lock` is committed.
- The python.org framework build has no CA bundle configured, so plain `python3`
  HTTPS calls fail. Poetry and pip bundle their own certs and are unaffected. If a
  script needs it, set `SSL_CERT_FILE=/etc/ssl/cert.pem`.

# Commands

- `poetry install` — sync the environment with `poetry.lock`.
- `poetry add <pkg>` / `poetry add --group dev <pkg>` — add a dependency.
- `poetry run marge --help` — run the CLI.
- `poetry run ruff check .` — lint. `poetry run ruff format .` — format.
- `poetry run mypy` — type check (`strict`, scoped to `src` via `pyproject.toml`).

# Project structure

```
src/marge/
├── __init__.py   package version
├── __main__.py   enables `python -m marge`
└── cli.py        argparse entry point; `main()` returns an exit code
```

- `src/` layout: the package is only importable once installed, so `poetry run`
  exercises the real installed package.
- New modules go in `src/marge/`. The console script `marge` is wired to
  `marge.cli:main` in `pyproject.toml`.

# Code style

- Use comments only to document a function/file/etc. according to proper
  conventions, or where code would be confusing without extra context.
- Keep lines below ≈90 characters (enforced by `[tool.ruff] line-length = 90`).
- Use short and descriptive variable/function names.
- Avoid excessively long functions/files.

# Git conventions

- Do all work on a branch off `main` and merge it back; don't commit changes
  directly to `main`. The one exception is the root commit that scaffolds a
  repository, which has no branch to come from.
- Name branches `<type>/<short-description>`, reusing the commit types below:
  `feat/template-loader`, `fix/empty-content-dir`, `chore/bump-ruff`.
- Write commit subjects as `<type>: <summary>` — imperative mood, ≤72
  characters, no trailing period. Types: `feat`, `fix`, `refactor`, `perf`,
  `docs`, `test`, `chore`. Add a body when the *why* isn't obvious from the diff.
- Keep each commit to one logical change. Formatting-only churn goes in its own
  `chore:` commit.
- Before committing, `poetry run ruff format .`, `poetry run ruff check .`, and
  `poetry run mypy` must all pass.
- When `pyproject.toml` dependencies change, commit the updated `poetry.lock` in
  the same commit.
- Rebase an unmerged branch onto `main` to keep its history linear, then merge
  with `--no-ff` so the branch stays visible. Never rewrite commits already on
  `main`.
- Delete branches once they're merged.
