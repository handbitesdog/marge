# Project overview

marge is a simple static site generator written in Python 3 that outputs HTML/CSS/JS. The focus is on clean syntax and producing performant and readable code.

# Environment

- Dependencies are managed with [Poetry](https://python-poetry.org/) (2.x).
- Requires Python ≥3.10. Locally installed: 3.14.6.
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

- Use comments only when documenting a function/file/etc. according to proper conventions or if code may be confusing without extra context.
- Keep lines below ≈90 characters (enforced by `[tool.ruff] line-length = 90`).
- Use short and descriptive variable/function names.
- Avoid excessively long functions/files.

# Git conventions

- Always create a new branch from `main` for making changes. Never commit directly to `main`.