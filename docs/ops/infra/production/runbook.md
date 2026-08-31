# PROD operations runbook — Cloudflare

> Timeless post-build contract. **NOT RUNNABLE until [setup.md](setup.md) sections 1–6 are implemented and verified.** Checked-in HEAD `70dc7405` cannot deploy production through `backend-deploy.yml`. External resource, billing, Environment, backup, notification, DNS, and database state is verification-required.

Task details and live status belong in Linear. Root [`todo-po.md`](../../../../todo-po.md) is only a pointer; it has no `#253` detail section and is not a semantic SSOT.

## 1. Release gate and order

A human release owner verifies the reviewed ref, exact GitHub Environment name/protection, production secret scope, backup evidence, frontend API target, and rollback candidate. Then the activated workflow contract must run:

**deploy → migrate → `/health` → optional smoke**

Every production invocation is a hard stop until setup acceptance criteria are present in the workflow and verified. `/health` returns process health and does not prove DB access. CRUD smoke uses a deliberately provisioned synthetic account with named owner, expiry, and cleanup; production demo seed credentials are not expected because migrate is master-only.

## 2. Incident triage

1. Declare an owner and timestamp. Record no secret or PHI.
2. Compare configured domain and provider-direct health path.
3. Inspect the bounded Actions/log evidence and provider status.
4. Determine whether the failure is Worker/Container, route/DNS, DB, or frontend target.
5. Stop writes when data integrity is uncertain.
6. Choose forward fix or the reviewed rollback below. AWS is not a rollback target.

## 3. Rollback with `GOOD_SHA`

Branch mutation, approval, and deployment are human-only. A commit merely existing is insufficient. The deployment ref must resolve to that commit, and the resulting workflow run must report the same `headSha`.

```bash
GOOD_SHA='<reviewed-last-known-good-commit>'
git rev-parse --verify "${GOOD_SHA}^{commit}"

# Human creates/selects a reviewed immutable ref that resolves exactly to GOOD_SHA.
ROLLBACK_REF='<reviewed-ref-for-GOOD_SHA>'
test "$(git rev-parse "${ROLLBACK_REF}^{commit}")" = "$(git rev-parse "${GOOD_SHA}^{commit}")" || exit 1

# Only after setup is implemented and Environment approval is active:
gh workflow run backend-deploy.yml --ref "$ROLLBACK_REF"
```

After dispatch, obtain the run metadata through the approved operator flow. **Stop unless the workflow run `headSha` equals `GOOD_SHA`.** Also stop if Environment approval is absent, config is not `wrangler.production.jsonc`, migration compatibility is not reviewed, or the production Worker/route differs. Never dispatch `--ref production` while claiming it pins an older SHA.

Schema rollback is not automatic. If `GOOD_SHA` is incompatible with applied migrations, use a forward-compatible fix or a separately approved restore plan.

## 4. Backup/restore rehearsal

Backup existence and restore success are external facts. Record dated evidence. Restore only to a disposable isolated target. Before any destructive restore option, enforce an explicit identity assertion:

```bash
: "${ISOLATED_DB_HOST:?}"
: "${ISOLATED_DB_NAME:?}"
test "$ISOLATED_DB_NAME" = "$EXPECTED_DISPOSABLE_DB_NAME" || exit 1
test "$ISOLATED_DB_HOST" = "$EXPECTED_DISPOSABLE_DB_HOST" || exit 1
# Operator also verifies provider project/id is the approved disposable target.
pg_restore --clean --if-exists -h "$ISOLATED_DB_HOST" -U "$ISOLATED_DB_USER" -d "$ISOLATED_DB_NAME" '<snapshot>'
```

The allowlisted host/database and disposable provider ID must come from the approved rehearsal record, not visual similarity. Stop on mismatch or nonzero exit. Verify schema, representative non-sensitive data, application compatibility, timing, and cleanup. No production restore proceeds solely from this example.

## 5. Secrets, seed, and monitoring boundaries

- Use `npx wrangler secret put <NAME> -c wrangler.production.jsonc` from `backend/`.
- The required secret-name inventory is [setup.md](setup.md). Values never enter docs/logs.
- `APP_ENV` is not currently injected and is not the seed gate. Current migrate is master-only.
- Notification policies, backup schedules, DNS, certificate, DB, and Environment reviewers are required contracts, not achieved facts. Verify each externally with a date and evidence owner.
- Record Actions run URL/id, `headSha`, approver, health result, and rollback decision without secret/PHI.

## 6. Go-live checklist

- [ ] setup acceptance criteria implemented and verified
- [ ] exact production Environment protection verified
- [ ] required secret **names** reconciled with Wrangler config; values verified through protected channel
- [ ] backup and isolated restore rehearsal evidence reviewed
- [ ] workflow `headSha`, Worker/config/route, migrate, and health verified
- [ ] frontend deployed API target verified
- [ ] optional synthetic smoke account owner/expiry/cleanup recorded
- [ ] live issue/date/billing status verified in Linear/provider systems by a human

Unchecked items mean stop. This document does not assert current external state.
