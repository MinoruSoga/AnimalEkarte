# Animal Ekarte Agent Instructions

This repository uses Claude Code and Codex-style agent workflows. The source of truth for project-specific development rules is [.claude/CLAUDE.md](.claude/CLAUDE.md).

## Required First Step

Before making changes, read [.claude/CLAUDE.md](.claude/CLAUDE.md), then load only the referenced files relevant to the task. Do not bulk-read unrelated docs.

Load `.claude/rules/` by path and task type. Frontend or docs-only work does not need the Go/Gin backend guidelines. Backend Go work must read [.claude/rules/go-gin-backend-guidelines.md](.claude/rules/go-gin-backend-guidelines.md) (mirrored at `.agents/rules/`). Claude Code auto-loads project rules; other agents must read the relevant files explicitly.

If a specification unknown blocks the current unit, ask before executing that unit. Independent, already-allowed work can continue.

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

- **Never** run `git reset --hard`, `git clean -fd(x)`, `git checkout -- .`, `git restore .`, or force-push. Do not assume Claude Hook enforcement applies to Codex; verify the effective runtime controls.
- To sync with remote: `git fetch` + `git merge` / `git pull --ff-only` (after checking `git status` for foreign WIP).
- **Parallel Grok/Claude/Codex tasks must use separate git worktrees** (or isolation worktree). Do not share one working tree across concurrent agents.
- Prefer WIP commits over discarding work. Do not “clean the tree” to unblock yourself.

### Branch deletion by creator (Mandatory)

- **User-created branches must not be deleted by agents.** This includes user-created claim branches.
- **AI-created branches may be deleted by agents** after verifying that the work is integrated, explicitly abandoned, or handed off with all remaining changes preserved, and that no active session or worktree still uses the branch. Do not delete `main`, `staging`, or `production` as routine cleanup.
- Determine the creator from session records or an explicit user statement; a branch name or commit author alone is not proof. If the creator or active use is unknown, preserve the branch and report the uncertainty.
- Prefer `git branch -d <exact-branch>` for local cleanup. If it refuses, inspect unmerged work instead of automatically forcing deletion. Remote branch deletion remains an external action requiring explicit authorization. Report which branches were deleted.

### Packet claim protocol (Mandatory)

Mutual exclusion on a ledger task ID / packet ID. Convention over existing `git branch` only — no script, hook, or CI.

- **Check (required before the first edit for any ledger task ID):** `git branch --list 'claim/<TASK-ID>'`. A non-empty result is a hard stop: report BLOCKED naming the claim branch; do not edit.
- **Acquire:** `git branch claim/<TASK-ID>` in the shared repository. A non-zero exit means another session already owns the task — hard stop and report BLOCKED; do not proceed.
- **Release (creator-aware):** Apply the branch deletion rules above. Agents may release AI-created claims after integration, explicit abandonment, or a confirmed handoff that preserves remaining work and ends the previous session's ownership. User-created claims must not be deleted by agents. A claim points to the acquisition commit, so `git branch -d` succeeding does not by itself prove that the claimed work is integrated; inspect the actual task changes and worktree first.
- Never delete an active session's claim to unblock yourself. Age or an apparently abandoned branch is not proof of release eligibility. Preserve uncertain claims and report the branch name and commit date (`git log -1 --format='%ci %s' claim/<TASK-ID>`). A session ending with unmerged, unhanded-off work is not sufficient reason to release its claim.
- `<TASK-ID>` is the ledger task ID or packet ID exactly as written (example: `BE-ACT-BOUNDED-MASTER-LISTS`). If an ID contains characters invalid in a git reference name, replace each invalid character with `-` and state the transformed name in the session report before acquire.

## Execution Autonomy

- Resolve specification blockers before the dependent unit; continue independent authorized work.
- Once scope is clear, proceed through the in-scope work without asking mid-task confirmation questions.
- Stop only for explicit safety boundaries: destructive operations (including any working-tree wipe), credential or secret changes, external posting/publishing/pushing/merging, paid actions, production-impacting actions, or irreversible third-party changes.

## Configuration scope

Generic AI workflow, model preferences, sandbox, Hooks and external connectors belong in user settings. Keep this repository's clinical invariants, Docker verification, worktree/claim policy and domain Skills here. Prefer project Skills for these project adaptations; use user Skills for generic tasks. See [agent harness operations](docs/ops/agent-harness.md).

## Completion contract

- Complete the accepted scope, review the owned diff, fix findings, and run applicable scoped checks.
- Report changed behavior, checked worktree/diff, commands and results, unresolved items and external state separately.
- Missing tools, skipped checks and untested behavior are BLOCKED, SKIP or UNKNOWN, never PASS. Static/candidate checks do not establish runtime or release readiness.
- Keep Linear as execution SoT. When unavailable, prepare a local draft and report its current state UNKNOWN; external posting needs authorization.

## Verification

- Prefer scoped verification tied to changed files.
- Do not auto-run the full-project prohibited commands listed in [.claude/CLAUDE.md](.claude/CLAUDE.md).
- For docs-only changes, state that runtime verification was not needed.

## Directory Guidance

Nested `CLAUDE.md` files provide local rules for backend, frontend, infra, docs, and specific layers. The closest applicable file wins when it is more specific and does not conflict with this root guidance.
