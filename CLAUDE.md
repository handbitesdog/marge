# CLAUDE.md

## Project Overview

- Name: marge
- Description: Simple static-site generator.
- Stack: Written in Go.

## Code style

- Only use comments for properly documenting files/functions and explaining code that may be confusing. Never narrate your actions using comments.

## Git conventions

- Don't commit work to `main`. Run `git branch --show-current` before the first commit of a session; if it's `main`, create a branch first. The one exception is bootstrapping a repository with no commits yet — `main` has to exist before anything can branch from it.
- Branch from up-to-date `main`: `git switch main && git pull && git switch -c <name>`.
- Branch names: `<type>/<kebab-description>` using the same types as commits — `feat/card-component`, `fix/button-focus-ring`, `docs/custom-properties`.
- Conventional Commits: `<type>(<scope>): <subject>`. Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `build`, `chore`. Scope is the component or section touched — `feat(card):`, `fix(reset):`. Omit scope for repo-wide changes.
- Subject line: imperative mood, lowercase, no trailing period, under ~50 chars. "add pill button variant", not "Added pill button variant."
- Body only when the change needs justification — a workaround, a breaking rename, a non-obvious tradeoff. Wrap at 72. Describe the change and why, never the conversation that produced it. No "as requested", no "per feedback".
- Breaking changes: `!` after the type/scope and a `BREAKING CHANGE:` footer.
- One logical change per commit. Stage explicit paths (`git add example.css example.html`), never `git add -A` or `git add .`.
- Show me `git status` and `git diff --staged` before committing. Don't commit and report afterward.
- Don't amend or rebase anything already pushed.
- After a branch is merged into `main`, delete it (`git branch -d <name>`).
- Don't add `Co-Authored-By` or generated-with trailers to commits.

## Safety

- Never commit secrets. If .env is touched, verify .gitignore before any commit.
- Never run rm -rf, git reset --hard, git push --force, DROP TABLE, or similar destructive operations without explicit confirmation.

## Mindset

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.