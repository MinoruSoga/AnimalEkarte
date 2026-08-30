# Security Scan Workflow

This repository uses `security-scan.yml` with AgentShield to run security checks.

`affaan-m/agentshield` is pinned to a full commit SHA, not a mutable version tag.
When upgrading AgentShield, resolve the target tag to its commit SHA with
`git ls-remote https://github.com/affaan-m/agentshield.git refs/tags/<tag>^{}` and
update the workflow intentionally.

- **Triggers**
  - `push` on `main`, `staging`
  - `pull_request` targeting `main`, `staging`, `production`
  - `workflow_dispatch` (manual run)

- **What it scans**
  - Claude/Codex agent configuration surfaces (`.claude/**`, agents, hooks, etc.)
  - Baseline findings already exist on `main` (historically score ~40 / grade D).
    Cleaning that debt is a separate hardening track, not a gate on every product/docs PR.

- **Failure behavior (2026-08-30 root fix)**
  - AgentShield **always runs** (report visible on the check).
  - **Fails the job only when**:
    1. PR targets `main`, **and** agent-config paths changed
       (`.claude/**`, `.codex/**`, `.agents/**`, `**/CLAUDE.md`, `**/.mcp.json`,
       or this workflow file), **or**
    2. `workflow_dispatch` with `force_fail_on_findings=true`
  - Push events and PRs that do **not** touch agent-config paths are **report-only**
    (same baseline debt must not block docs/backend/frontend PRs).

- **Why**
  - Previously every PR to `main` failed on medium+ findings that already lived on
    `main`, while `push` to `main` did not fail — so docs-only PRs were blocked without
    improving agent security. Path-gated fail restores a real signal.

- **Manual run**
  - `min_severity`: `doc` | `low` | `medium` | `high` | `critical` (default: `medium`)
  - `force_fail_on_findings`: force fail regardless of path filter (default: false)
