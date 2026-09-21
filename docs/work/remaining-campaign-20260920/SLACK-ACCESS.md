# SLACK-ACCESS — staff-access input gaps versus P5 and H3-9

Campaign `remaining-ops-20260920` revision 1, unit `SLACK-ACCESS`, claim `claim/SLACK-ACCESS`.
Worktree `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-rem-slack-access` on `feat/rem-slack-access-20260920` at HEAD `873685b0b`. Sheet date: 2026-09-20.

Docs-only. This unit does **not** invent a roster, transcribe July Slack names, run `staff-provision`, run `stg-uat-staff-attach`, read secret files, or treat demo catalog logins as production / formal UAT accounts.

Binding: [todo-issue.md](../../../todo-issue.md) heading `### SLACK-ACCESS` (L264–268); [P5](../../../todo-operations.md#p5--staff-provision); [H3-9](../../../todo-operations.md#h3-9); [seedlogin env](../../../backend/internal/seedlogin/env.go); [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md); [LINMIG-230.md](../linmig-campaign-20260919/LINMIG-230.md). July roster / forward text is **not in this repo**. Personal names stay **UNKNOWN**.

This file is a campaign investigation sheet. Product modules do not import it. The controller ledger names it as `owned_paths` / acceptance target (`units.json` unit `SLACK-ACCESS`; `unit-specs.json` in_scope). Sibling ops docs (`STAFF_ACCOUNT_PROVISIONING.md`, `LINMIG-230.md`) are contracts or a remote-apply design, not this clinic/role/demo-versus-formal gap sheet.

## 1. Why this sheet exists

[todo-issue.md](../../../todo-issue.md) L266–268: all-clinic login, demo privileges, and per-staff accounts are operational readiness. A 9 September login-success report is a partial example, not a full-staff confirmation. First work is to sort **P5 / H3-9 input gaps by clinic, role, and demo versus formal account**. Formal email / permission policy is PO. Credential changes are a separate approval. If remote apply is undefined or the current roster is missing, **stop that execution**. Do not reuse demo privileges on production.

This sheet is that input-gap table. It is **not** apply authorization.

## 2. Three account classes (do not mix)

| Class | What it is | Write owner / entry | May exist on | Not |
|---|---|---|---|---|
| **Demo catalog** | Synthetic LoginForm `DEMO_ACCOUNTS` upserted at migrate phase `003_login` | [`seedlogin.Apply`](../../../backend/internal/seedlogin/apply.go); gated by [`ShouldApply`](../../../backend/internal/seedlogin/env.go#L19) | `development` / `local` / `dev` / `test` / `staging` only | Production accounts; current hospital roster; P5 completion |
| **Formal provision (P5 / #255)** | Batch **create** account + staff + assignments + permission groups | `backend/cmd/staff-provision` `preflight` then `apply` | Authorized local disposable DB; shared STG/PROD only after USER approval **and** a defined in-environment method ([LINMIG-230.md](../linmig-campaign-20260919/LINMIG-230.md) §3) | Demo catalog; attach onto existing staff |
| **H3-9 attach** | Link synthetic UAT accounts onto **existing** `staffs` rows; does not insert staff | [`Makefile`](../../../Makefile) `stg-uat-staff-attach-preflight` / `stg-uat-staff-attach`; [`backend/cmd/stg-uat-staff-attach`](../../../backend/cmd/stg-uat-staff-attach/main.go) L1–2 | Target host/database confirmed on the command line | Proof that P5 remote apply exists ([todo-operations.md](../../../todo-operations.md) L177–179) |

Operator / first system admin is a fourth, out-of-band path ([STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L25; [FIRST_SYSTEM_ADMIN.md](../../ops/deploy/FIRST_SYSTEM_ADMIN.md)). `SEEDLOGIN_OPERATOR_*` values are repo-external. Catalog emails never receive the shared-password shortcut as operator accounts ([env.go](../../../backend/internal/seedlogin/env.go) L45–48).

## 3. Demo catalog (code facts — not a roster)

Source: [env.go](../../../backend/internal/seedlogin/env.go), [catalog.go](../../../backend/internal/seedlogin/catalog.go), [env_test.go](../../../backend/internal/seedlogin/env_test.go). Personal names in `clinicLocalPeople` / `hayashiSpec` are **fixture identities**. They are **not** copied here and are **not** the July or current hospital roster.

| Fact | Evidence | Gap for access readiness |
|---|---|---|
| `ShouldApply` is true only after trim/lowercase `APP_ENV` ∈ {`development`,`local`,`dev`,`test`,`staging`} | env.go L19–28; env_test.go L18–28 | Live `APP_ENV` on each host is **UNKNOWN** this session |
| `production`, empty, unknown (`preview`, `prod`, …) fail closed | env.go L20–27; env_test.go L24–28 | Demo seed must not be planned for production login |
| Shared-password shortcut requires allowlisted env **and** catalog email **and** `SharedPassword` | env.go L45–59 | Do not paste the constant value into this sheet or chat. Production auth must not accept it (env.go L10–12) |
| Catalog emails are `stg-staff-{staffID}@example.test` | catalog.go `emailPattern`; catalog_test.go first/last emails | `@example.test` is synthetic ([STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L24). Formal emails are I-EMAIL **未記入** |
| Four catalog clinic bands: id `1` 八王子病院, `2` 城東センター病院, `3` ノア動物病院　敷島病院, `4` ノア動物病院　Hako bu neco | catalog.go `clinicBands` L73–78 | These are **demo seed labels**, not a verified live `clinic_id` map (I-CLINIC **未記入**) |
| Permission groups assigned to demo logins: `一般` (home clinic) and `執行` (one all-clinic executive) | catalog.go L17–21, L80–92 | Formal role → `permission_group_ids` is I-ROLE **未記入**; inference is forbidden (ops doc L21) |
| Catalog size: 1 all-clinic executive + 9 home-clinic general templates × 4 clinics = **37** demo rows | catalog.go L61–92 | Count is fixture cardinality, not staff headcount |
| Production / empty / unknown `APP_ENV` skips migrate login seed | [login_seed.go](../../../backend/cmd/migrate/login_seed.go) L17–24 | Presence of these 37 rows on any live STG DB is **UNKNOWN** (not queried) |

**Do not** use catalog emails, shared password, or 執行-all-clinic demo as production or as a substitute for a current roster.

## 4. Clinic × role × demo-versus-formal gap matrix

Clinic labels below are catalog seed labels only. Live clinic identity, occupancy, and whether those ids still match STG/PROD are **UNKNOWN**. Role columns are demo permission-group names, not PO job titles. Formal cells stay **UNKNOWN** until I-ROSTER / I-CLINIC / I-ROLE are supplied **outside** git.

| Clinic (catalog id / label) | Demo `執行` (all-clinic) | Demo `一般` (home clinic) | Formal P5 account (create) | H3-9 attach on existing staff | Formal login/permission receipt |
|---|---|---|---|---|---|
| 1 / 八王子病院 | Fixture exists in catalog (one shared executive login, home band 1) | 9 home-clinic general fixtures | **UNKNOWN** — no current roster row | **UNKNOWN** — attach state not observed | **UNKNOWN** |
| 2 / 城東センター病院 | Same single executive covers catalog clinics 1–4 | 9 home-clinic general fixtures | **UNKNOWN** | **UNKNOWN** | **UNKNOWN** |
| 3 / ノア動物病院　敷島病院 | Same | 9 home-clinic general fixtures | **UNKNOWN** | **UNKNOWN** | **UNKNOWN** |
| 4 / ノア動物病院　Hako bu neco | Same | 9 home-clinic general fixtures | **UNKNOWN** | **UNKNOWN** | **UNKNOWN** |
| Any clinic **not** in catalog ids 1–4 | No demo catalog login | No demo catalog login | **UNKNOWN** | **UNKNOWN** | **UNKNOWN** |

Staff **names** (demo fixture or formal): **UNKNOWN** in this sheet. Do not backfill from Slack, GitHub issue comments, or catalog.go person strings.

Occupation labels on demo templates (`獣医師` / `看護師` / `動物看護師` / `VT` / `スタッフ`) are not `staff_type` (`doctor` / `nurse` / `trimmer` / `resource` in the P5 manifest) and are not `permission_group_ids`. Crossing them is an I-ROLE gap, not a mapping this sheet may invent.

## 5. P5 / #255 missing inputs (formal create)

Source: [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L148–163; restated by [LINMIG-230.md](../linmig-campaign-20260919/LINMIG-230.md) §6 and [todo-operations.md](../../../todo-operations.md) L177–183. This session did not receive a current roster file. Values, names, emails, and passwords are not recorded.

| ID | Missing input | Why needed for access | Supplier | Status this session | Demo catalog does **not** fill this |
|---|---|---|---|---|---|
| I-ROSTER | Current staff list (name / clinic / role) as manifest `staff[]`. Distinguish historical intake from an **applicable current** list | Formal accounts cannot be specified | 先方 / PO | **UNKNOWN** — current version and applicability not confirmed (ops doc L154) | Catalog fixtures are not a roster |
| I-EMAIL | Email policy (personal required / shared forbidden; Q&A No.30) | Manifest `email` | PO | **未記入 / UNKNOWN** | `@example.test` is synthetic only |
| I-CLINIC | Clinic name → `clinic_id` | `clinic_scope` / `main_clinic_id` | 運用 | **未記入 / UNKNOWN** | catalog.go labels are seed, not live confirmation |
| I-ROLE | Role → **explicit** `permission_group_ids` (no inference) | Permission assignment | PO | **未記入 / UNKNOWN** | `一般` / `執行` names are demo group names only |
| I-LEAVE | Whether leave / retired / contractor accounts are issued | `is_active` | PO | **未記入 / UNKNOWN** | Catalog rows are active demo logins |
| I-ACTOR | Authorized `actor_account_id` in the **target** DB | preflight authorization | USER | **未記入 / UNKNOWN** | Operator bootstrap is a different path |
| I-ENV | Target (local disposable / named STG / PROD) and approval | apply is USER; PROD after #253/#254 | USER | **未記入 / UNKNOWN** | `ShouldApply` allowlist is not apply approval |
| I-RECEIPT | PII-free `batch_id` / digest / count after authorized apply | #255 AC | USER after apply | **未記入 / UNKNOWN** | Login-seed checksum is not a staff-provision receipt |

Remote apply: [STAFF_ACCOUNT_PROVISIONING.md](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) L103 and [LINMIG-230.md](../linmig-campaign-20260919/LINMIG-230.md) §3 — `STAFF_PROVISION_ALLOW_REMOTE=YES_I_UNDERSTAND` is a laptop non-local host guard, **not** shared STG/PROD authorization. Designed path is inside the target environment with `DB_HOST=db`. This unit does **not** run that path.

## 6. H3-9 missing inputs (attach, not create)

Source: [todo-operations.md](../../../todo-operations.md) L133–137, L25; Makefile L340–381; `stg-uat-staff-attach` package comment L1–2. Current attach + existing execution evidence: **UNKNOWN** (todo-operations.md L60).

| ID | Missing input | Why needed | Status this session |
|---|---|---|---|
| H-ROSTER | Repo-external attach roster (`schema_version` `stg-uat-staff-attach-v1`) with existing `staff_id`, `clinic_id` / `clinic_ids`, email, `permission_group_ids`, `secret_ref` | Command requires `STG_UAT_STAFF_ATTACH_ROSTER` | **UNKNOWN** — path not supplied; do not invent staff_id rows |
| H-SECRETS | Matching 0600 secrets file | `STG_UAT_STAFF_ATTACH_SECRETS` | **UNKNOWN** — not read |
| H-TARGET | `STG_UAT_STAFF_ATTACH_CONFIRM_HOST` and `TARGET_DB_NAME` | Makefile L340–345; attach has `--confirm-target-host` / `--confirm-target-database` (staff-provision still lacks an equivalent flag — LINMIG-230 §3.2) | **UNKNOWN** |
| H-REMOTE | Whether shared STG attach is approved (`STG_UAT_STAFF_ATTACH_ALLOW_REMOTE`) | Makefile L341–342 | **UNKNOWN**; this session does not set it |
| H-EXISTING | Whether target staff rows already exist (attach does not create staff) | todo-operations.md L135 | **UNKNOWN** — DB not queried |
| H-RECEIPT | PII-free digest / staff_count / staff_ids after attach | stdout contract in attach `attachResult` | **UNKNOWN** |
| H3-11 | Same-target login / clinic / permission / reload cases | Depends on H3-9 receipt (todo-operations.md L26, L139–143) | **BLOCKED** until H-RECEIPT |

Name match on screen is not identity (todo-operations.md L143). Do not auto-merge by display name.

## 7. Access-readiness implications (what is still blocked)

| Intended use | Allowed from repo evidence? | Blocker |
|---|---|---|
| Local/STG **demo** login using catalog emails on an env where `ShouldApply` is true | Code allows the shortcut; **live env/DB presence UNKNOWN** | Confirm `APP_ENV` and that `003_login` actually applied; still not formal staff access |
| Formal STG/PROD login for real staff | **No** | I-ROSTER…I-RECEIPT all open; remote apply not executed |
| H3-9 UAT attach then H3-11 screen check | **No this session** | H-* inputs UNKNOWN; do not apply |
| Treating 9 September partial login success as all-clinic confirmation | **No** | todo-issue.md L266 |
| Reusing demo 執行 / shared password in production | **Forbidden** | env.go L10–12, L19–28; todo-issue.md L268 |

## 8. Stop / safety boundary

- If a current roster file, password file, or account **creation/attach apply** is required to proceed, **stop**. This sheet is complete with UNKNOWN cells.
- Do not invent names, emails, `staff_id`s, or `permission_group_ids`.
- Do not copy catalog person strings into ops receipts.
- Agents must not run `staff-provision` or `stg-uat-staff-attach`.
- First system admin remains out of band.

## 9. Verification for this docs unit

Docs-only. No Docker app tests. Expected commands: `git diff --check`; `git diff --name-only`; `git diff --cached --name-only`; `git ls-files --others --exclude-standard`; `rg seedlogin`, `staff-provision` on cited files and this path; `git ls-files --error-unmatch` after add.
