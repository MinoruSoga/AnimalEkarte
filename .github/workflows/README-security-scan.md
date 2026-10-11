# Security Scan Workflow

This repository uses `security-scan.yml` with AgentShield to run security checks.

The maintained action is vendored under `.github/actions/agentshield`, based on
`affaan-m/agentshield` v1.6.0 commit `b0891303bdcd6037376a94263d45cfd2ff3dfb98`.
It corrects Prime schema applicability and quoted-prohibition classification;
all other checks, input defaults and failure gates are retained.

Execution uses a separate checkout of the reviewed full repository commit
`50922e8e65db6248d435ec8b76d4ec752c8b6c96` and validates the bundle SHA-256.
The PR checkout is scanned as data; its scanner implementation is not executed.
The new checkout does not persist credentials, the job keeps `contents: read`,
and no secret, write permission, dependency install or privileged event is added.
See the action's README, MIT LICENSE, provenance.json and upstream.patch for
reproduction and maintenance. Updating the scanner requires review, regression
checks, rebuilding, refreshing provenance/hash, then an intentional commit-pin
change. Never select execution code from a PR head or mutable ref.

At PR606 head `1bb9c08`, local medium-threshold reproduction changes32→29
(High11/Medium18); fail-on-findings still fails. Capabilities26, backend settings2
and device1 remain unresolved. Runtime controls or capability reductions are
not included, and quoted prohibitions are not proof of runtime enforcement.

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
    1. PR targets `main` or `staging`, **and** agent-config paths changed
       (`.claude/**`, `.codex/**`, `.agents/**`, root `CLAUDE.md`,
       `**/AGENTS.md`, `**/.mcp.json`). Nested `docs/**/CLAUDE.md` is
       directory guidance, not the fail-gated agent-config surface, **or**
    2. `workflow_dispatch` with `force_fail_on_findings=true`
  - Push events and PRs that do **not** touch agent-config paths are **report-only**
    (same baseline debt must not block docs/backend/frontend PRs).
  - Edits to `security-scan.yml` itself are **not** in the fail path filter
    (chicken-and-egg: the gate-fix PR must change the workflow).

- **Why**
  - Previously every PR to `main` failed on medium+ findings that already lived on
    `main`, while `push` to `main` did not fail — so docs-only PRs were blocked without
    improving agent security. Path-gated fail restores a real signal.

- **Manual run**
  - `min_severity`: `doc` | `low` | `medium` | `high` | `critical` (default: `medium`)
  - `force_fail_on_findings`: force fail regardless of path filter (default: false)
