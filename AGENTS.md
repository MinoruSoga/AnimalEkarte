# Animal Ekarte Agent Instructions

This repository uses Claude Code and Codex-style agent workflows. The source of truth for project-specific development rules is [.claude/CLAUDE.md](.claude/CLAUDE.md).

## Required First Step

Before making changes, read [.claude/CLAUDE.md](.claude/CLAUDE.md), then load only the referenced files relevant to the task. Do not bulk-read unrelated docs.

Also load the project-wide rules in [.claude/rules/](.claude/rules/) (including the Go/Gin backend guidelines — mirrored at `.agents/rules/`). Claude Code auto-loads these; other agents must read them explicitly.

## Core Rules

- For backend work, follow [.claude/rules/go-gin-backend-guidelines.md](.claude/rules/go-gin-backend-guidelines.md). Design packages by cohesion, consumers, and dependency direction; do not treat a fixed layer layout as a Go/Gin requirement.
- Keep TypeScript and Go type-safe; do not introduce `any` or untyped escape hatches.
- Use Docker-based commands for this project; do not run local npm/go commands directly.
- Keep changes minimal and aligned with the nearest directory-level `CLAUDE.md`.
- Validate inputs at boundaries and preserve clinic, owner, pet, and staff data separation.
- Never expose secrets, tokens, credentials, private data, or operationally sensitive details.
- After pulling a commit that adds or changes migrations, developers must run `make migrate` before using the updated app.
- Agents must not auto-apply migrations; when the post-pull rule applies, surface `make migrate` for the user to run manually.

## Git / Parallel Agent Safety (Mandatory)

Full policy: [.claude/rules/git-worktree-safety.md](.claude/rules/git-worktree-safety.md).

- **Never** run `git reset --hard`, `git clean -fd(x)`, `git checkout -- .`, `git restore .`, or force-push. These are permission-deny + PreToolUse hook blocked.
- To sync with remote: `git fetch` + `git merge` / `git pull --ff-only` (after checking `git status` for foreign WIP).
- **Parallel Grok/Claude tasks must use separate git worktrees** (or isolation worktree). Do not share one working tree across concurrent agents.
- Prefer WIP commits over discarding work. Do not “clean the tree” to unblock yourself.

### Packet claim protocol (Mandatory)

Mutual exclusion on a ledger task ID / packet ID. Convention over existing `git branch` only — no script, hook, or CI.

- **Check (required before the first edit for any ledger task ID):** `git branch --list 'claim/<TASK-ID>'`. A non-empty result is a hard stop: report BLOCKED naming the claim branch; do not edit.
- **Acquire:** `git branch claim/<TASK-ID>` in the shared repository. A non-zero exit means another session already owns the task — hard stop and report BLOCKED; do not proceed.
- **Release (USER-only):** Agents must never delete a claim branch (own or another's). After the work is integrated into `main` or explicitly abandoned, the agent reports the claim branch name(s) to the user; the user releases with `git branch -D claim/<TASK-ID>` in a human terminal. Release is coupled to integration: merging into `main` is a user action, so releasing the claim is one too. An agent that released at the end of its own run would drop the lock while its work is still unmerged, letting another session take the same task — the exact failure this protocol exists to prevent.
- Never delete another session's claim to unblock yourself. A claim that looks abandoned is reported to the user with its branch name and commit date (`git log -1 --format='%ci %s' claim/<TASK-ID>`); removal is a user action.
- `<TASK-ID>` is the ledger task ID or packet ID exactly as written (example: `BE-ACT-BOUNDED-MASTER-LISTS`). If an ID contains characters invalid in a git reference name, replace each invalid character with `-` and state the transformed name in the session report before acquire.

## Execution Autonomy

- Ask specification questions only before execution starts, such as during /grill-me or an equivalent clarification phase.
- Once scope is clear, proceed through the in-scope work without asking mid-task confirmation questions.
- Stop only for explicit safety boundaries: destructive operations (including any working-tree wipe), credential or secret changes, external posting/publishing/pushing/merging, paid actions, production-impacting actions, or irreversible third-party changes.

## Verification

- Prefer scoped verification tied to changed files.
- Do not auto-run the full-project prohibited commands listed in [.claude/CLAUDE.md](.claude/CLAUDE.md).
- For docs-only changes, state that runtime verification was not needed.

## Directory Guidance

Nested `CLAUDE.md` files provide local rules for backend, frontend, infra, docs, and specific layers. The closest applicable file wins when it is more specific and does not conflict with this root guidance.
