# [検証] parent implement task

## Role
dev-reviewer only. Approve PASS or FAIL block. Do not implement product fix (tiny test-only OK if needed for proof; prefer FAIL).

## Steps
1. Work on parent impl worktree path (dir-share). Read branch diff vs main.
2. Confirm single-BUG scope.
3. Re-run scoped tests from card DoD with AnimalEkarte Docker:
   - backend: docker compose run --rm --no-deps --entrypoint '' -v "$PWD/backend:/app" -e TZ=Asia/Tokyo backend go test ...
   - frontend: docker compose run --rm --no-deps --entrypoint '' -v "$PWD/frontend:/app" -w /app frontend sh -c 'npm test -- --run <paths>'
   NEVER omit --entrypoint '' (air hang).
4. PASS only if code+tests support the bug fix claim.
5. Output first line exactly: APPROVE (PASS) or REJECT (FAIL) with reasons.

## Forbidden
merge, push, migrate, other BUGs, VERIFIED_FIXED, Linear Done.
