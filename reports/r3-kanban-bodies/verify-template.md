# [検証] Parent impl verify

## Workspace
Orchestrator sets dir-share to parent worktree + exact branch_name. Do not create second worktree of same branch.

## Pass criteria
1. Scope matches single BUG only.
2. Evidence: test commands + exit 0 logs (Docker --entrypoint '' for go test).
3. No migrate applied; no merge/push by engineer.
4. Code review: minimal diff, no PII, CLAUDE.md respected.

## Outcomes
- PASS → complete this card with summary + commit SHA on branch.
- FAIL → block parent with concrete gaps (request-changes path).

## Forbidden
- Implementing fixes yourself beyond tiny verify-only typos if policy allows; prefer FAIL.
- Marking PASS without running tests.
