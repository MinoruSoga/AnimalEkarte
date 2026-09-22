# LINMIG-230 — Staff-batch provision execution method, secret supply, and rollback (docs-only)

Campaign `linmig-ops-prep-20260919` revision 1. Claim ID `LINMIG-230` (Linear issue not found; keep as claim only). Maps to [todo-operations.md](../../../todo-operations.md) P5 / GitHub [#255](https://github.com/MinoruSoga/AnimalEkarte/issues/255).

Worktree: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-linmig-230` on `feat/linmig-230-ops-prep` at HEAD `aac697645df92fd24611c7c13bf0f7dda12a6e08`. Sheet date: 2026-09-19. Repo evidence only. This unit **does not** run `staff-provision`, read secret files, invent roster rows, edit Go, apply migrations, or change shared STG/PROD.

Truth source order used: `backend/cmd/staff-provision` and `backend/internal/staff` over [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) over [todo-operations.md](../../../todo-operations.md) P5. Where they agree, both are cited. Live roster, live environment, authorized actor, and apply receipts are **UNKNOWN** (not observed this session).

## 1. Why this sheet exists

[STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L103:

> approved remote execution と read-only secret-file mount の仕組みは未定義です。`STAFF_PROVISION_ALLOW_REMOTE` だけを根拠にローカル Compose から共有環境へ apply しません。

[todo-operations.md](../../../todo-operations.md) P5 (L177–183) repeats that gate and forbids treating H3-9 `stg-uat-staff-attach` as P5 completion. This sheet is the **design** for remote apply method, repo-external secret supply, rollback, and I-ROSTER gaps. It is not apply authorization and not an implementation change.

## 2. Current executable contract (do not re-implement here)

Binary: `backend/cmd/staff-provision`. Commands: `preflight` | `apply` only (`main.go` L159–165). There is **no** rollback subcommand.

| Fact | Evidence |
|---|---|
| preflight is write-0; receipts are not consulted | [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L109–115; `staff_provisioning_apply.go` L17–19, L21–47 |
| apply re-validates inputs, then one transaction + `pg_advisory_xact_lock` on `staff-provision:`+`batch_id` | [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L117–127; `staff_provisioning_apply.go` L50–118; `staff_provisioning_repository.go` lock key |
| same-digest complete receipts → `status=noop`; partial/mismatch → conflict | [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L122–124; `staff_provisioning_apply.go` L81–94 |
| stdout / logs are PII-free (`status` / `batch_id` / `digest` / `staff_count` / `clinic_scope`) | [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L105; `main.go` L2–4, L134–153; `TestWriteJSON_EmitsDigestOnlySurface` |
| apply against non-local `DB_HOST` refuses unless `STAFF_PROVISION_ALLOW_REMOTE=YES_I_UNDERSTAND` | `main.go` L102–116; `TestRun_ApplyRefusesNonLocalWithoutOverride` |
| local hosts | `db`, `localhost`, `127.0.0.1`, `::1`, `[::1]` in `backend/internal/dbconn/dbconn.go` L107–119 |
| H3-9 `stg-uat-staff-attach` attaches accounts onto **existing** staff rows; it does not create staff | `backend/cmd/stg-uat-staff-attach/main.go` L1–2; [todo-operations.md](../../../todo-operations.md) L177–178 |

`STAFF_PROVISION_ALLOW_REMOTE=YES_I_UNDERSTAND` is a **laptop non-local host guard**, not shared STG/PROD authorization (`main.go` L103–112; [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L22, L101–103; [todo-operations.md](../../../todo-operations.md) L177).

## 3. Designed remote apply method (not executed)

Stop rule: do not apply from a developer laptop to shared STG/PROD even if `STAFF_PROVISION_ALLOW_REMOTE` is set. The approved path is **inside the target environment**, where `DB_HOST` is that environment's local compose name (`db`).

### 3.1 Target selection (UNKNOWN until operator fills)

| ID | Required input | Why | Status this session |
|---|---|---|---|
| I-ENV | Apply target: local disposable **or** named shared STG **or** PROD, plus human approval | `DB_NAME` / compose project / host identity | **UNKNOWN** |
| I-ACTOR | Authorized `actor_account_id` already present in that DB | preflight authorization | **UNKNOWN** |
| I-RECEIPT | Prior PII-free `batch_id` / `digest` / `count` for this `clinic_scope` | detect noop vs conflict vs first apply | **UNKNOWN** |

PROD remains behind #253/#254 gates in the ops checklist ([STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L160). This sheet does not open that gate.

### 3.2 Execution entry (design)

1. Operator confirms target compose project, `DB_HOST` expected to be `db`, and `DB_NAME` **before** any apply. Current `staff-provision` has **no** `--confirm-target-host` flag (unlike `stg-uat-staff-attach`). Until such a flag exists, confirmation is an operator checklist, not a binary gate. **Do not implement the flag in this unit.**
2. Copy manifest/secrets onto the **target host** as repo-external 0600 regular files (see §4). Do not scp into the git worktree.
3. Run **preflight** first with write-0:

```bash
docker compose run --rm --no-deps --entrypoint '' -T \
  -v /secure/staff-batch:/secure/staff-batch:ro backend \
  go run ./cmd/staff-provision preflight \
    --manifest=/secure/staff-batch/manifest.json \
    --secrets=/secure/staff-batch/secrets.json
```

Citation: [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L75–85.

4. Human reviews PII-free stdout (`batch_id`, `digest`, `staff_count`, `clinic_scope`). Compare digest to any prior I-RECEIPT. Do not paste names/emails/passwords into issues or this sheet.
5. After **explicit USER approval for that named environment**, run **apply** with the same mount and same files. Keep `STAFF_PROVISION_ALLOW_REMOTE` **unset** on this path so a mistaken public `DB_HOST` still fail-closes (`main.go` L102–116).
6. Record only PII-free receipt fields as I-RECEIPT. Login / clinic / permission / audit checks are USER work after apply ([todo-operations.md](../../../todo-operations.md) L183).

### 3.3 What this method is not

- Not `STAFF_PROVISION_ALLOW_REMOTE` from local Compose toward a shared hostname.
- Not H3-9 attach.
- Not a new CI workflow or Cloudflare Job. Those are unspecified; inventing them would exceed this unit.
- Not authorization to run apply in this session.

## 4. Repo-external secret supply (design)

Contract: [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L17–20, L49–60; `ValidateSecureInputPath` / `OpenSecureInputFile` in `backend/internal/staff/staff_provisioning_validate.go` L16–78.

| Rule | Binding check |
|---|---|
| Manifest and secrets live **outside** every configured repo root | `ValidateSecureInputPath` rejects paths inside repo roots |
| Absolute path, `filepath.Clean` + realpath | relative paths rejected |
| Regular file, **not** a symlink; open with `O_RDONLY\|O_NOFOLLOW` | L30–37, L73 |
| Permission **mode 0600** only | L36–38 |
| `secret_ref` 1:1 with manifest; surplus/missing/duplicate fails preflight | ops doc L59–60 |
| Password policy same as normal staff create (8+ chars, letter+digit, ≤72 bytes) | ops doc L60 — **do not record actual passwords here** |
| Logs/receipts/stdout never include name, email, password, or file body | ops doc L20; `main.go` sanitizeError |

Operator mount: host directory → container `:ro` as in the documented `docker compose run -v /secure/staff-batch:/secure/staff-batch:ro` example. The **approved remote** version of that mount is: the 0600 files already exist on the target host (or a host-local tmpfs), then the **same** read-only bind is used inside that environment's compose. Files are not committed, not copied into this worktree, and not echoed.

This session did not read any secrets file. Presence of `/secure/staff-batch` on any host is **UNKNOWN**.

## 5. Rollback design (no CLI; do not run)

`staff-provision` has no undo command. Rollback is two different procedures depending on whether the transaction committed.

### 5.1 Before commit (automatic)

Any error inside `WithTx` rolls the transaction back: no accounts, no staff rows, no receipts, no audit (`staff_provisioning_apply.go` L73–116; repository atomicity tests). Operator action: keep the same 0600 files, fix the **non-secret** cause (scope, actor, FK, conflict), re-run **preflight**, then stop for approval. Do not delete receipts that do not exist.

### 5.2 After `status=applied` (compensating, USER-only)

Atomic apply either creates the whole batch or nothing. There is no supported “delete the batch” API in this command. Compensating control after a mistaken successful apply:

1. Stop. Do not re-apply a different digest for the same `batch_id` / `clinic_scope` (conflict by design; ops doc L122–124).
2. Do **not** delete `audit_logs` or `staff_provision` receipts. Receipts are the idempotency key; deleting them would allow a second create.
3. Deactivate or permission-reduce **individual** staff through the existing staff-master path (`POST /api/v1/masters/staffs/{id}/account` is the add-account path; inactivation is the normal staff update path — exact UI steps are out of scope). Prefer `is_active=false` over physical delete.
4. Capture a PII-free note: `batch_id`, `digest`, `staff_count`, clinic_scope ids, actor id, environment name, timestamp. No names/emails/passwords.
5. Password material on disk: shred/unlink the **host** secrets file after the window; never git-rm what was never in git.

### 5.3 After `status=noop`

No data change. No rollback. Treat as success of idempotent re-entry. If the operator expected a new batch, the digest already matches — do not invent a new `batch_id` to force writes.

### 5.4 Conflict

Fail-closed. Resolve by comparing PII-free receipts. Do not bypass the advisory lock. Do not apply a second digest into the same `clinic_scope`+`batch_id`.

## 6. Missing-input table (I-ROSTER gaps)

Values, names, emails, passwords, and invented staff rows are not recorded. Unsupplied stays **未記入 / UNKNOWN**. Source: [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L148–163. This session did not receive a current roster file.

| ID | Missing input | Why needed | Supplier | Status this session |
|---|---|---|---|---|
| I-ROSTER | Current staff list (name / clinic / role) as `staff[]` source. Distinguish historical intake from an **applicable current** list | Manifest rows | 先方 / PO | **UNKNOWN** — current version and applicability not confirmed. Do not copy names into this sheet. Historical vs current distinction: ops doc L154 |
| I-EMAIL | Email policy (personal required / shared forbidden per Q&A No.30) | Manifest `email` | PO | **未記入 / UNKNOWN** |
| I-CLINIC | Clinic name → `clinic_id` map | `clinic_scope` / `main_clinic_id` | 運用 | **未記入 / UNKNOWN** |
| I-ROLE | Role → **explicit** `permission_group_ids` (no inference) | Permission assignment | PO | **未記入 / UNKNOWN** |
| I-LEAVE | Whether leave / retired / contractor accounts are issued | `is_active` policy | PO | **未記入 / UNKNOWN** |
| I-ACTOR | Authorized `actor_account_id` in the target DB | preflight auth | USER | **未記入 / UNKNOWN** |
| I-ENV | Target (local / STG / PROD) and approval | apply is USER; PROD after #253/#254 | USER | **未記入 / UNKNOWN** |
| I-RECEIPT | PII-free `batch_id` / digest / count after an authorized apply | #255 AC | USER after apply | **未記入 / UNKNOWN** |

**Not done in this unit:** inventing roster rows, committing real roster, production/STG apply, writing PII to Issue/Linear, running `staff-provision`.

## 7. Stop / safety boundary

- If a roster file, password file, or remote apply is required to proceed, **stop**. This sheet is complete as a design with UNKNOWN cells.
- First system admin is out of band ([FIRST_SYSTEM_ADMIN.md](../../ops/deploy/FIRST_SYSTEM_ADMIN.md)); `staff-provision` assumes an existing authorized actor.
- Agents must not auto-apply this command.

## 8. Verification performed for this docs unit

Docs-only. No Docker app tests. Commands: `git diff --check`; `git diff --name-only`; `git diff --cached --name-only`; `git ls-files --others --exclude-standard`; `rg STAFF_PROVISION_ALLOW_REMOTE`, `staff-provision`, `preflight`; `git ls-files --error-unmatch` on this path after add.
