# Animal Ekarte Agent Instructions

This repository uses Claude Code and Codex-style agent workflows. The source of truth for project-specific development rules is [.claude/CLAUDE.md](.claude/CLAUDE.md).

## Required First Step

Before making changes, read [.claude/CLAUDE.md](.claude/CLAUDE.md), then load only the referenced files relevant to the task. Do not bulk-read unrelated docs.

Also load the project-wide rules in [.claude/rules/](.claude/rules/) (code conventions such as Go package layout — mirrored at `.agents/rules/`). Claude Code auto-loads these; other agents must read them explicitly.

## Core Rules

- Preserve the Handler → Service → Repository architecture.
- Keep TypeScript and Go type-safe; do not introduce `any` or untyped escape hatches.
- Use Docker-based commands for this project; do not run local npm/go commands directly.
- Keep changes minimal and aligned with the nearest directory-level `CLAUDE.md`.
- Validate inputs at boundaries and preserve clinic, owner, pet, and staff data separation.
- Never expose secrets, tokens, credentials, private data, or operationally sensitive details.

## Execution Autonomy

- Ask specification questions only before execution starts, such as during /grill-me or an equivalent clarification phase.
- Once scope is clear, proceed through the in-scope work without asking mid-task confirmation questions.
- Stop only for explicit safety boundaries: destructive operations, credential or secret changes, external posting/publishing/pushing/merging, paid actions, production-impacting actions, or irreversible third-party changes.

## Verification

- Prefer scoped verification tied to changed files.
- Do not auto-run the full-project prohibited commands listed in [.claude/CLAUDE.md](.claude/CLAUDE.md).
- For docs-only changes, state that runtime verification was not needed.

## Directory Guidance

Nested `CLAUDE.md` files provide local rules for backend, frontend, infra, docs, and specific layers. The closest applicable file wins when it is more specific and does not conflict with this root guidance.
