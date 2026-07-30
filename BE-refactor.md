# BE-refactor — backend audit (closed 2026-07-30)

## Purpose and authority

This file is the backend audit and actionability source, not the implementation-task execution ledger. Current source, tests, schema, and accepted ADR semantics outrank historical summaries. Executable task authority is [`3-session-agent.html#ledger`](3-session-agent.html#ledger). Release, reset, secret, and environment operations belong to [`q&a.html#ops`](q&a.html#ops) and its current runbooks.

## Status: closed

This audit is complete. There is no open work here and no packet is executable from this file.

- 141 findings: 118 `CLOSED_VERIFIED`, 23 `WITHDRAWN`, 0 `ACTIVE`.
- All 8 remaining `BE-ACT-*` packets were integrated on 2026-07-30: `b385cfdbb` (animal-species FE system-admin gate), `c3179776b` (campaign target share-lock serialization), `ce79e0c23` (LSTEP delivery-batch bulk reads), `a45046985` (LSTEP failure-contract docs), `700beb087` (merchandise atomic delete), `a3174b84e` (animal-species spec authorization), `4f51998e4` (bounded master lists), `d9b0568de` (session-ledger convergence).
- The full per-finding resolution index — each finding mapped to the current path, symbol, and behavioral reason it is defeated — was removed when the audit closed. It is preserved in Git at `3d3410f93` and earlier; read it with `git show 3d3410f93:BE-refactor.md`.

Code comments citing `BE-refactor.md` with identifiers from earlier generations of this file (`R1-2`, `R3-5`, `E-8`, `§5-1-3`, …) refer to superseded revisions. Resolve them through Git history, not through this file.

## Starting a new backend audit

Write the new frontier here rather than reviving the closed one. Two constraints carry forward because both were established by measurement, not convention:

- **Claim before editing.** Acquire `git branch claim/<TASK-ID>` in the shared repository before the first edit, per the packet claim protocol in [`AGENTS.md`](AGENTS.md). Worktree isolation alone does not prevent two sessions implementing the same task; on 2026-07-30 that failure occurred four times in one day.
- **DOCKER-MOUNT-PROOF.** Container verification counts only after comparing the SHA-256 of at least one owned source file on the host task worktree with the same path inside the container, or using an isolated Compose project mounted from that worktree. The `backend` service bind-mounts the shared main tree, so a container started there never observes a worktree's bytes — a false green that must stop the unit.
