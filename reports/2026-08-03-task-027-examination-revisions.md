# TASK-027 examination revisions 実行報告（2026-08-03）

## Completion Report

- Run status: COMPLETE

### Checklist Results

| Checklist item | Expected behavior | Actual behavior | Status | Verification method | Evidence |
|----------------|-------------------|-----------------|--------|--------------------|----------|
| Saved prompt validation | resume prompt が `dynamic-workflow/v1` を満たす | validator exit 0 | PASS | `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fast-task-027-resume-lint-gate.md` | `Prompt Craft Harness Validation: PASS` / `Execution contract: dynamic-workflow/v1` / `EXIT_CODE=0` |
| Slice A append-only revision | confirm が official revision/items → audit → pointer CAS を同一 transaction で行い、失敗時に全 rollback する | 実装・DB-backed regression と failure injection が green | PASS | revision/confirm/audit/rollback/cross-clinic gate | `ok  \tgithub.com/animal-ekarte/backend/internal/medicalrecord\t5.834s` / `EXIT_CODE=0` |
| Migration FK/index | 004 の全 3 FK を child の非 partial index が左端から覆う | 3/3 catalog test と独立 SQL review が PASS | PASS | migration catalog tests、全 SQL review | `backend/migrations/004_examination_revisions.sql:34,84-107`; 対応表は後掲 |
| Migration append-only/RLS | revision parent/items の UPDATE/DELETE と selected/gap version insert を拒否し、clinic RLS を適用する | trigger 3 個、RLS 2 表、nonowner/NOBYPASSRLS probe が green | PASS | migration DB tests | `backend/migrations/004_examination_revisions.sql:110-191` |
| Slice B unconfirm | reason 必須、`confirmed -> completed` のみ、working revision append、audit/CAS atomic | 空白/500超/status/legacy pointer/audit failure を fail-closed、reconfirm で official append | PASS | unconfirm/permission/item/pet gate | `ok  \tgithub.com/animal-ekarte/backend/internal/medicalrecord\t4.117s` / `EXIT_CODE=0` |
| Dedicated permission | unconfirm は既存 examination edit から分離した default-deny resource | route・default rule・frontend の専用権限を実装 | PASS | apicontract/clinic gate、security review | `backend/internal/medicalrecord/routes.go:369`; `ok .../apicontract 0.029s`; `ok .../clinic 0.369s` |
| Confirmed/ever-confirmed mutation safety | confirmed mutation、history 後の patient relation change/delete、死亡患者への新規紐付けを拒否する | confirmed は 409、revision pointer 後の reassignment/delete は拒否。直接 pet・カルテ由来 create/rebind は transaction 内の clinic-scoped status check で write 前に拒否 | PASS | medicalrecord tests + race/DB/transaction gates + final clinical/isolation review | `TestExaminationService_DeceasedPetWriteBoundary` race PASS; `TestDB_ExaminationServiceRejectsPollutedClinicalRelations` PASS; final C/H/M `0/0/0` |
| Legacy confirmed fail-closed | revision pointer のない legacy confirmed を official/unconfirm できない | not-found/conflict として拒否 | PASS | official/unconfirm regression tests | final medicalrecord gates exit 0 |
| OpenAPI/codegen | response pointer、unconfirm request/route、generated types が同期する | optional read-only revision pointer と reason contract を同期 | PASS | `make codegen-check` | `git diff --exit-code frontend/src/types/generated/` / `EXIT_CODE=0` |
| Slice C result rows | 手動行 add/delete/name/value が immutable update、空名 value loss を拒否、focus を保持 | 44px control、一意 label、add/delete focus handoff、blank-name fail-closed。items query pending/error/cached-error は保存不可、成功した空応答だけ authoritative | PASS | exact examinations gate | `13 passed`; `160 passed`; `EXIT_CODE=0` |
| Slice C patient change | 初回 confirm 前だけ、生存・同一 candidate・正整数 ID を mutation 直前に再検証する | revision/confirmed/permission race、死亡/不明、ID drift は保存全体を fail-closed。route id 変更でフォーム全体を keyed remount | PASS | hook/dialog/table tests + security review | `use-examination-form.test.ts` 57 tests PASS; route permissions 11 tests PASS; `PatientSelectionTable.test.tsx` 22 tests PASS |
| Slice C unconfirm UI | dedicated permission、理由 label/required/max500、成功後 cache refresh | fieldset 外の Radix dialog、latest permission recheck、3 cache key invalidationをawait | PASS | examinations gate + independent React/security review | final React/security C/H/M `0/0/0`; APPROVE |
| Shared update request | `pet_id` を PATCH body に保持する | request type/runtime body test が green | PASS | exact shared hook gate | `1 passed`; `2 passed`; `EXIT_CODE=0` |
| Clinic lintscan | Preload scope、master FK write、DBOrTx inventory が green | scoped gate green | PASS | exact lintscan command | `ok  \tgithub.com/animal-ekarte/backend/internal/lintscan\t1.549s` / `EXIT_CODE=0` |
| Go vet | medicalrecord/model scoped vet が diagnostics なし | compose env warning のみ、Go diagnosticsなし | PASS | exact vet command | `EXIT_CODE=0` |
| Go lint truth | Makefile と同じ pinned version、cap 無効化で全件を出す | v2.11.4 が 62 baseline issues を報告。Slice B 時点と同数・同分類で新規 finding なし | PASS | corrected official-image command | `62 issues:` / categories `1,5,8,11,17,2,1,6,1,3,4,3` / `EXIT_CODE=1`; **baseline-red / delta-green** |
| Migration non-apply | agent は migration/seed apply を行わない | 004 は committed、未適用。001 は RLS baseline だけをread-only確認、001/002/003は未変更、seedsは未アクセス | PASS | command log、commit path reconciliation | `make migrate` 未実行。ユーザー手動 command は後掲 |
| Slice commits | A/B/C と final repair を green で path-scoped commit | 4 code commit 完了、foreign WIP 非包含 | PASS | `git show --name-only`、commit直後 status | A `046615f4...`; B `1dd1cf04...`; C `c161baff...`; repair `fb0cf9c9...` |
| Foreign WIP preservation | `bug.md` と着手時の ReservationFormModal 2 files を編集/stage/commitしない | 本 unit は全3 pathを非所有として維持。ReservationFormModal 2 path は別 owner の `617f6f9bf`、`bug.md` は別 owner の `fc1cc5a8e` により解消 | PASS | start/final porcelain、commit path lists | start 3 paths; final foreign WIPなし; TASK commitsに非包含 |
| Independent reviews | DB、Go、TS、React/a11y、security、clinic isolation、clinical safety を独立 review | 実装後 review の全 C/H/M を修正し、最終 verdict は APPROVE | PASS | joined reviewer evidence | Go/OpenAPI、TS、React、security は C/H/M `0/0/0`; final clinic/clinical verdict は後掲 |
| Workflow orchestration | real fan-out、sole-writer、falsification/reconciliation、全 launch join/cancel | native Workflow unavailableのため multi-agent parity mode。全 launch を join/cancel | PASS | orchestration ledger | 後掲の各 role 行 |
| Final reconciliation | 全 acceptance item を current HEAD で再実行する | exact backend/frontend/codegen/vet gates green。lint は baseline-red として限定 | PASS | Verification Strategy 1-10 | 本表と verbatim output 節 |

### Run Summary

- Changed files: Slice A 14 paths、Slice B 30 paths、Slice C 23 paths、final repair 7 paths。正確な所有境界は各 commit の `git show --pretty= --name-only` で固定（一覧は後掲）。さらに `todo.md` の TASK-027 outcome と本 report を更新。
- Failure Signature log: FS-1〜FS-11（後掲）。全 source failure は修復済み。package lint 62 件は baseline-red として限定。
- Staged plan ledger: `todo.md:559` の TASK-027 実施結果を COMPLETE / IMPLEMENTED_UNMERGED へ更新。本 report は `reports/2026-08-03-task-027-examination-revisions.md`。
- Risk Tier: Local write | Safety boundary events: migration/seed apply、push/merge/Issue close、claim delete、credential操作はなし。formatter で一度 `npx` が container cache に `prettier@3.9.6` を一時取得した逸脱を記録。manifest/lockfile差分なし、再実行せず、後続 `pnpm exec prettier` は command absent で exit 254。

## Root-cause summary and minimal patch

既存 confirm は mutable `exams` status と audit だけを更新し、確定時点の official snapshot、working revision、確定解除理由、専用権限、ever-confirmed patient identity lock を持っていなかった。最小 patch は次の3 sliceと最終安全修復である。

1. Slice A: `examination_revisions` / `examination_revision_items`、pointer、append-only trigger、RLS、revision repository/service、confirm の revision → audit → CAS。
2. Slice B: reason 必須 unconfirm、working append、reconfirm official append、confirmed/ever-confirmed mutation guard、dedicated default-deny permission、OpenAPI/codegen。
3. Slice C: revision pointer response、手動 result rows、初回 confirm 前 patient change、dedicated unconfirm dialog、latest-state mutation guards。
4. Final repair: 死亡患者への direct/record-derived create・relation rebind を transaction 内で fail-closed にし、items query readiness と record ID ごとの override/items/test-type state、route remount を固定。

## Harness Selection and Execution Loop Selection

- Contract: `dynamic-workflow/v1`。
- Chosen harness: TDD。`tdd-workflow`、Go test、migration/clinic isolation、security、scoped verification instructionsを読み、RED → GREEN → review → repair → reconciliationを実行。
- Execution loop: sequential。A が B の schema/service foundation、B が C の API/permission foundationであるため slice 間は直列。slice 内の read-only probes/reviewsだけを並列化。
- Stop condition: checklist 全項目が PASS、または同一 failure signature 3 回で genuine BLOCKED。3 回反復した signature はなし。
- Native Workflow: available tools に存在しなかったため未使用。real parallel subagents + single active writer + independent proof reviewで parity orchestration。
- Loop monitoring: native recurring loop monitor は不要な bounded run。各 iteration は Failure Signature log と gate outputで記録。

## Failure Signature log

| ID | Expected | Actual / error | Verification | Attempt / fix | Result |
|----|----------|----------------|--------------|---------------|--------|
| FS-1 | confirm event `revision,audit,cas` | `legacy-status-update,audit`; official reader/migration absent | initial Slice A RED | revision model/repository/service/migrationをtest-first実装 | GREEN |
| FS-2 | selected/gap item version sealed、schema/catalog/RLS一致 | SQL review HIGH 2: late insert可能、full catalog/RLS proof不足 | migration tests + independent full-file review | insert-version trigger、complete 004 fixture、catalog/RLS/relation tests | GREEN / final SQL APPROVE |
| FS-3 | saved-prompt lint gate executes | backend image exact entrypointは `golangci-lint` absent | original prompt exact command | resume promptでMakefile正本のofficial pinned imageに修正 | harness起動成功。62 baseline issuesを完全計数 |
| FS-4 | B compile/test green | response helper型が別 test fileに依存し package単体生成で壊れた | Slice B gate | `LinkedTreatmentHistoryItem` を production response DTOへ移動 | GREEN |
| FS-5 | C RED contracts absent | response pointer compile error、row callbacks/dialog/API missing、`include_deceased` absent | focused Vitest/Go RED | response/transform、rows、dialogs、hook/API実装 | GREEN |
| FS-6 | scoped ESLint 0 | restricted generated-domain import 2 errors | explicit changed-file ESLint | feature-local `BackendExamination` DTOへ置換 | exit 0 |
| FS-7 | fail-closed clinical UI | review tests: locked patient save、unknown patient、blank manual name、OpenAPI pointer contractが失敗 | independent TS/React/security/Go review | whole-save abort、positive alive/ID check、blank-name alert、OpenAPI assertion、awaited invalidations | GREEN; final reviewers APPROVE |
| FS-8 | keyboard focus remains after row mutation | add/delete focus tests 2 failures | `ExamItemsTable.test.tsx` | stable-key refs + layout focus handoff（new → next → previous → add） | 38/38 GREEN |
| FS-9 | final clinical write boundaries are closed | direct/record-derived deceased pet create/rebind と items pending/error が write 可能 | final clinical review + focused RED | shared deceased-pet adapter、tx-scoped status lookup、query-success readiness gate | GREEN |
| FS-10 | A→B navigation cannot reuse A state | Aのlocal overrides/test type/itemsがBへ漏れ、B server rowsをtemplateで上書き | adversarial React/security review + RED | ID-scoped override/items/type refs + keyed route remount | GREEN |
| FS-11 | A→B pending→A cached restores A rows | initialized refがAのままでAをready扱いし、空items PATCHが可能 | React re-review + RED | ID変更時にinitializedを無効化し、owner-scoped visible rowsを再初期化 | GREEN |

## De-Sloppify Pass result

- `FormDataEntryValue` を unsafe cast せず `typeof === "string"` で narrow。
- unconfirm cache invalidation 3 promises を `await Promise.all`。
- manual row value + blank name の silent drop を visible error + no-mutationへ変更。
- patient candidate は latest ref、exact ID、`status === "生存"`、positive safe integer、revision/permission lockを mutation直前に再検証。
- row add/delete後のfocus、44px controls、一意な accessible names、dirty trackingを追加。
- response DTO と OpenAPI の `current_revision_version` driftを契約testで固定。
- 死亡判定は既存 shared-kernel validatorをadapter経由で再利用し、既存死亡患者の非relation履歴訂正は過剰拒否しない。
- route keyで患者選択・action stateを含む全record-local stateをatomic remountし、hook側にもID-scoped defenseを保持。
- changed-file ESLint exit 0、`git diff --check` exit 0、package/lockfile差分なし。
- 全体 `tsc` / full frontend build / full repo test は禁止されたため未実行。scoped greenをrepo-wide greenへ拡張しない。

## Independent Review Gate result

- Migration full-file reviewer: round 1 HIGH 2件を修正後 APPROVE。最終 CRITICAL/HIGH/MEDIUM `0/0/0`。
- Slice B Go、clinic isolation、clinical safety、security、correctness reviewers: 最終 APPROVE、CRITICAL/HIGH `0/0`。
- Slice C Go/OpenAPI: APPROVE、C/H/M `0/0/0`。response mapper、OpenAPI optional/readOnly int64/min1、全 response pathを確認。
- Slice C TypeScript: APPROVE、C/H/M `0/0/0`。ESLint 0、154 + 2 + 22 tests。
- Slice C React/a11y: APPROVE、C/H/M `0/0/0`。独立 128/128 tests。
- Slice C security: APPROVE、C/H/M `0/0/0`。独立 94 tests、XSS/credential/storage/logging追加なし。
- Final repair Go、clinic isolation、React/a11y、security、clinical safety: 全て APPROVE、C/H/M `0/0/0`。race、DB pollution、ambient transaction/share-lock、68 frontend testsを独立確認。

## Saved Prompt Validation Gate result

```text
$ node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fast-task-027-resume-lint-gate.md
Prompt Craft Harness Validation: PASS
Profile: standard (declared-risk-tier)
Target: agent (source-path)
Quality mode: standard
Execution contract: dynamic-workflow/v1
EXIT_CODE=0
```

## Verification command outputs and regression checks completed

```text
$ docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*Examination.*(Revision|Confirm|Audit|Rollback|CrossClinic)' -count=1
ok  	github.com/animal-ekarte/backend/internal/medicalrecord	5.834s
EXIT_CODE=0

$ docker compose exec -T backend go test -p 1 ./internal/lintscan -run 'Test.*(PreloadClinicScope|MasterFKWrite|DBOrTx)' -count=1
ok  	github.com/animal-ekarte/backend/internal/lintscan	1.549s
EXIT_CODE=0

$ docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*Examination.*(Unconfirm|Permission|Item|Pet)' -count=1
ok  	github.com/animal-ekarte/backend/internal/medicalrecord	4.117s
EXIT_CODE=0

$ docker compose exec -T backend go test -p 1 ./internal/apicontract ./internal/clinic -run 'Test.*(Examination|Permission|DefaultPermissionRuleTable)' -count=1
ok  	github.com/animal-ekarte/backend/internal/apicontract	0.029s
ok  	github.com/animal-ekarte/backend/internal/clinic	0.369s
EXIT_CODE=0
```

```text
$ make codegen-check
mkdir -p frontend/src/types/generated
docker compose --env-file .env.local run --rm codegen
git diff --exit-code frontend/src/types/generated/
EXIT_CODE=0
```

```text
$ docker compose exec -T frontend npx vitest run src/features/examinations src/features/master/components/permission-rule-table-model.test.ts
Test Files  13 passed (13)
Tests       160 passed (160)
EXIT_CODE=0

$ docker compose exec -T frontend npx vitest run src/hooks/use-update-examination.test.tsx
Test Files  1 passed (1)
Tests       2 passed (2)
EXIT_CODE=0

$ docker compose exec -T frontend npx vitest run src/components/shared/ReservationFormModal/PatientSelectionTable.test.tsx
Test Files  1 passed (1)
Tests       22 passed (22)
EXIT_CODE=0
```

Final repair gates:

```text
$ docker compose exec -T backend go test -race -p 1 ./internal/medicalrecord -run '^TestExaminationService_DeceasedPetWriteBoundary$' -count=1
ok  	github.com/animal-ekarte/backend/internal/medicalrecord	1.025s
EXIT_CODE=0

$ docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run '^TestDB_ExaminationServiceRejectsPollutedClinicalRelations$' -count=1
ok  	github.com/animal-ekarte/backend/internal/medicalrecord	0.323s
EXIT_CODE=0

$ docker compose exec -T backend go test -p 1 ./internal/reservation -run '^TestReservationRepository_FindPetByIDInClinic_(SeesUncommittedDeceasedUpdate|ShareLockBlocksConcurrentWriter)$' -count=1
ok  	github.com/animal-ekarte/backend/internal/reservation	0.540s
EXIT_CODE=0

$ docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*Examination' -count=1
ok  	github.com/animal-ekarte/backend/internal/medicalrecord	9.479s
EXIT_CODE=0

$ docker compose exec -T frontend npx eslint src/features/examinations/hooks/use-examination-form.ts src/features/examinations/hooks/use-examination-form.test.ts src/features/examinations/routes/ExaminationForm.tsx src/features/examinations/routes/ExaminationForm.permissions.test.tsx
[no ESLint findings]
EXIT_CODE=0
```

```text
$ docker compose exec -T backend go vet ./internal/medicalrecord/... ./internal/model/...
EXIT_CODE=0
```

テスト stderr の `useActionState` transition warning、Radix Select `act(...)` warning、意図的 API error log は既存 test-harness warningであり、test failureではない。

## lint 実行結果

```text
$ rg -n '^GOLANGCI_LINT_VERSION[[:space:]]*[:?+]?=' Makefile
268:GOLANGCI_LINT_VERSION := v2.11.4

$ docker run --rm -v "$PWD/backend:/app" -v ekarte-go-mod-cache:/go/pkg/mod -v ekarte-golangci-cache:/root/.cache -w /app golangci/golangci-lint:v2.11.4 golangci-lint run --max-same-issues 0 --max-issues-per-linter 0 ./internal/medicalrecord/... ./internal/model/... ; echo "EXIT_CODE=$?"
62 issues:
* contextcheck: 1
* errorlint: 5
* gocritic: 8
* gofmt: 11
* goimports: 17
* gosec: 2
* prealloc: 1
* staticcheck: 6
* unconvert: 1
* unparam: 3
* unused: 4
* wrapcheck: 3
EXIT_CODE=1
```

`--max-same-issues 0 --max-issues-per-linter 0` を保持し、`tail`/pipelineを使わず真の終了コードを取得した。0件ではないため偽0 probeは不要。Slice B gateと最終 gateは同じ62件・同分類で、TASK-027の新規 findingは0。`examination_response.go` の goimports 1件は同 file の parent commitにも同じ import blockで存在する。したがって判定は **package baseline-red / TASK delta-green** であり、repo-wide lint greenを主張しない。

## Slice 到達点 / commit boundary

- Slice A: GREEN / committed — `046615f4bc923869f189c4e104e27d0539d8c88d` (`feat: add examination revision foundation`)、14 paths。
- Slice B: GREEN / committed — `1dd1cf04e77fa7adef38b0230a1b824e4f9abff6` (`feat: add examination unconfirm workflow`)、30 paths。
- Slice C: GREEN / committed — `c161baffb2372b8da3195a3b3474f4824f23ada6` (`feat: add examination revision frontend workflow`)、23 paths。
- Final repair: GREEN / committed — `fb0cf9c910aef842fdde1a0206bb5546163096c3` (`fix: fail closed on unsafe examination edits`)、7 paths。
- 未完 slice: なし。status は **IMPLEMENTED_UNMERGED**（mainのlocal commitであり、push/remote mergeは未実施）。

Owned path evidence:

```text
$ git show --pretty= --name-only 046615f4bc923869f189c4e104e27d0539d8c88d
backend/internal/lintscan/dbortx_inventory_lint_test.go
backend/internal/medicalrecord/examination_audit.go
backend/internal/medicalrecord/examination_parent_audit_test.go
backend/internal/medicalrecord/examination_parent_audit_tx_test.go
backend/internal/medicalrecord/examination_repository_test.go
backend/internal/medicalrecord/examination_revision_repository.go
backend/internal/medicalrecord/examination_revision_repository_tx_atomicity_test.go
backend/internal/medicalrecord/examination_revision_service.go
backend/internal/medicalrecord/examination_revision_service_test.go
backend/internal/medicalrecord/examination_service.go
backend/internal/medicalrecord/examination_service_test.go
backend/internal/model/examination_record.go
backend/internal/model/examination_revision.go
backend/migrations/004_examination_revisions.sql

$ git show --pretty= --name-only 1dd1cf04e77fa7adef38b0230a1b824e4f9abff6
backend/docs/api.yaml
backend/internal/apicontract/openapi_examination_mutation_contract_test.go
backend/internal/clinic/clinic_service.go
backend/internal/clinic/clinic_service_test.go
backend/internal/identitylink/response.go
backend/internal/identitylink/types.go
backend/internal/lintscan/dbortx_inventory_lint_test.go
backend/internal/medicalrecord/exam_reference_range_resolution_test.go
backend/internal/medicalrecord/examination_audit.go
backend/internal/medicalrecord/examination_handler.go
backend/internal/medicalrecord/examination_handler_test.go
backend/internal/medicalrecord/examination_items_handler_test.go
backend/internal/medicalrecord/examination_request.go
backend/internal/medicalrecord/examination_revision_repository.go
backend/internal/medicalrecord/examination_revision_service.go
backend/internal/medicalrecord/examination_revision_service_test.go
backend/internal/medicalrecord/examination_revision_workflow_repository.go
backend/internal/medicalrecord/examination_revision_workflow_safety_test.go
backend/internal/medicalrecord/examination_service.go
backend/internal/medicalrecord/examination_service_test.go
backend/internal/medicalrecord/routes.go
backend/internal/medicalrecord/routes_snapshot_test.go
backend/internal/model/audit_log.go
backend/internal/model/permission.go
frontend/src/features/master/components/permission-rule-table-model.test.ts
frontend/src/features/master/components/permission-rule-table-model.ts
frontend/src/types/generated/identitylink-responses.ts
frontend/src/types/generated/medicalrecord-responses.ts
frontend/src/types/generated/models.ts
frontend/src/types/generated/owner-responses.ts

$ git show --pretty= --name-only c161baffb2372b8da3195a3b3474f4824f23ada6
backend/docs/api.yaml
backend/internal/apicontract/openapi_examination_mutation_contract_test.go
backend/internal/medicalrecord/examination_response.go
backend/internal/medicalrecord/examination_response_test.go
frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.test.tsx
frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx
frontend/src/features/examinations/api/transforms.test.ts
frontend/src/features/examinations/api/types.ts
frontend/src/features/examinations/api/unconfirm-examination.test.tsx
frontend/src/features/examinations/api/unconfirm-examination.ts
frontend/src/features/examinations/components/ExamItemsTable.test.tsx
frontend/src/features/examinations/components/ExamItemsTable.tsx
frontend/src/features/examinations/components/ExaminationPatientChangeDialog.test.tsx
frontend/src/features/examinations/components/ExaminationPatientChangeDialog.tsx
frontend/src/features/examinations/components/ExaminationUnconfirmDialog.test.tsx
frontend/src/features/examinations/components/ExaminationUnconfirmDialog.tsx
frontend/src/features/examinations/hooks/use-examination-form.test.ts
frontend/src/features/examinations/hooks/use-examination-form.ts
frontend/src/features/examinations/routes/ExaminationForm.permissions.test.tsx
frontend/src/features/examinations/routes/ExaminationForm.tsx
frontend/src/hooks/use-update-examination.test.tsx
frontend/src/hooks/use-update-examination.ts
frontend/src/lib/transforms/examination.ts

$ git show --pretty= --name-only fb0cf9c910aef842fdde1a0206bb5546163096c3
backend/internal/medicalrecord/examination_pet_safety.go
backend/internal/medicalrecord/examination_service.go
backend/internal/medicalrecord/examination_service_test.go
frontend/src/features/examinations/hooks/use-examination-form.test.ts
frontend/src/features/examinations/hooks/use-examination-form.ts
frontend/src/features/examinations/routes/ExaminationForm.permissions.test.tsx
frontend/src/features/examinations/routes/ExaminationForm.tsx
```

各 commit直前に `git rev-parse HEAD` と `git diff --cached --name-only` を再取得し、path-scoped `git commit -m "..." -- <allowlist>` を使用した。HEAD movement時はreset/rebaseせず再anchor・再verifyした。

## migration SQL レビュー結果

Reviewer: `slice_a_sql_review`（実装者とは別のread-only role）。最初からゼロ指摘ではなく、round 1の HIGH 2件を修正後に final APPROVEとなった。

1. HIGH: revision item は UPDATE/DELETE拒否だけで、選択済み version へ後挿入できた。対応: `trg_examination_revision_items_insert_version` を追加し、pointer NULL はv1、pointer NはN+1だけ許可。selected/gap/late insertは`23514`。
2. HIGH: 004全文、FK action/deferrability、partial index、column/default/CHECK、RLS policy、cross-clinic relationのproofが不足。対応: complete migration rollback fixture、catalog assertions、nonowner role probe、8 relation DB tests、model/schema alignment。

Final: CRITICAL 0 / HIGH 0 / MEDIUM 0 / APPROVE。

## FK と index の対応表

| FK | Child columns | Parent | ON DELETE / deferrability | Covering non-partial index |
|----|---------------|--------|---------------------------|----------------------------|
| `fk_examination_revisions_exam` | `examination_revisions (clinic_id, examination_id)` | `exams (clinic_id, id)` | `RESTRICT`, `NOT DEFERRABLE` | `uq_examination_revisions_clinic_exam_version (clinic_id, examination_id, version)` left prefix |
| `fk_examination_revision_items_revision` | `examination_revision_items (clinic_id, examination_id, version)` | `examination_revisions (clinic_id, examination_id, version)` | `RESTRICT`, `NOT DEFERRABLE` | `idx_examination_revision_items_revision_sort (clinic_id, examination_id, version, sort_order, id)` |
| `fk_exams_current_revision` | `exams (clinic_id, id, current_revision_version)` | `examination_revisions (clinic_id, examination_id, version)` | `RESTRICT`, `NOT DEFERRABLE` | `idx_exams_current_revision_pointer (clinic_id, id, current_revision_version)` |

Catalog testは3 indexの`indpred IS NULL`も確認した。

## RLS 判断

RLSを両新表へ適用した。`001_init.sql:2964-2999` の `has_clinic_access` / `apply_rls_policy` helper定義と、`001_init.sql:3017-3048` の直接clinic列・親join policyの既存2パターンを確認し、004では直接`clinic_id`を持つ両表に同 helperを使用した（`004_examination_revisions.sql:179-191`）。

- `USING` / `WITH CHECK`: `app_private.has_clinic_access(clinic_id)`。
- catalog: `relrowsecurity=true`、policy expression一致、`relforcerowsecurity=false`。
- role probe: `NOLOGIN NOSUPERUSER NOBYPASSRLS`。same-clinic SELECT/INSERT許可、cross-clinic SELECT非表示、INSERT `42501`。
- `FORCE ROW LEVEL SECURITY` はbaselineに合わせて追加していない。table owner/superuser/BYPASSRLSのPostgreSQL標準迂回を解消したとは主張しない。runtime nonowner roleとrepository clinic predicatesが境界。

## Orchestration evidence

Mode: real multi-agent read-only fan-out + single active writer。native Workflow unavailable。

| ID / label | Role / responsibility | Writer-owned paths | Completion | Evidence / integration decision |
|------------|-----------------------|--------------------|------------|---------------------------------|
| `phase1_backend_probe` | backend current behavior | none | completed / integrated | RED surfaceとconfirm contractへ統合 |
| `phase1_schema_probe` | schema/FK/index/RLS probe | none | completed / integrated | 004 contractへ統合 |
| `phase1_contracts_probe` | permission/API probe | none | completed / integrated | Slice B contractへ統合 |
| `phase1_frontend_probe` | frontend caller/UI probe | none | completed / integrated | Slice C planへ統合 |
| `phase1_test_probe` | test/CI matrix | none | completed / integrated | scoped gateへ統合 |
| `phase1_isolation_probe` | clinic isolation probe | none | completed / integrated | tenant testsへ統合 |
| `phase2_docs` | project/docs contract | none | completed / integrated | migration/non-action boundariesへ統合 |
| `phase2_db_review_final` | preimplementation DB review | none | completed / integrated | design CHANGES_REQUIRED → PASS |
| `tdd_writer` | Slice A sole writer | Slice A 14 paths | completed / integrated | RED→GREEN、SQL repair |
| `slice_a_sql_review` | full 004 reviewer | none | completed / integrated | HIGH 2修正後 APPROVE |
| `phase2_planner` | planner | none | cancelled | stall後明示cancel、editなし |
| `phase2_planner/tenant_audit` | nested tenant audit | none | cancelled | recursive stall後cancel |
| `phase2_planner/tenant_audit/independent_boundary_audit` | nested audit | none | cancelled | editなし |
| `phase2_planner/tenant_audit/independent_boundary_audit/specialist_check` | nested specialist | none | cancelled | editなし |
| `phase2_planner_fast` | replacement planner | none | cancelled | stall後cancel |
| `phase2_db_design_review` | DB design review | none | cancelled | stall後cancel |
| `phase2_db_review_retry` | DB review retry | none | cancelled | stall後cancel |
| `slice_b_go_review` | Go correctness | none | completed / integrated | final APPROVE |
| `slice_b_isolation_review` | tenant boundaries | none | completed / integrated | final APPROVE |
| `slice_b_clinical_review` | clinical integrity | none | completed / integrated | final APPROVE |
| `slice_b_security_review` | auth/input/error security | none | completed / integrated | final APPROVE |
| `slice_b_correctness_review` | acceptance falsification | none | completed / integrated | final APPROVE |
| `slice_b_tenant_child` | independent tenant child review | none | completed / integrated | final APPROVE |
| `slice_c_frontend_probe` | existing UI/route evidence | none | completed / integrated | inline selectorを選択、create-only route guardを保持 |
| `slice_c_react_probe` | preimplementation hooks/a11y | none | completed / integrated | revision pointer/active pet/fieldset findingsを修正 |
| `slice_c_test_probe` | C RED matrix | none | completed / integrated | tests-first casesへ統合 |
| `slice_c_ts_review` | final TypeScript review | none | completed / integrated | restricted DTO/FormData/async findings修正後 APPROVE |
| `slice_c_react_final` | final React/a11y | none | completed / integrated | focus/44px/fail-closed修正後 APPROVE |
| `slice_c_security_final` | final security | none | completed / integrated | 94 tests、APPROVE |
| `slice_c_go_response_review` | response/OpenAPI contract | none | completed / integrated | contract drift修正後 APPROVE |
| `task027_final_isolation` | full A/B/C clinic isolation | none | joined result below | DB path final proof |
| `task027_final_isolation/independent_isolation` | nested independent isolation | none | joined result below | parent reviewerのfalsification |
| `task027_final_clinical` | full clinical safety | none | joined result below | clinical invariant final proof |
| `task027_final_clinical/task027_clinic_isolation` | nested isolation child | none | joined result below | clinical reviewerのtenant proof |
| `task027_repair_tdd` | repair RED strategy / zero-write assertions | none | completed / integrated | focused RED matrixとtx marker設計 |
| `task027_repair_go_review` | final Go correctness | none | completed / integrated | gofmt/vet/race/full service tests、APPROVE |
| `task027_repair_isolation` | final clinic boundary | none | completed / integrated | 5 clinic/DB/lock gates exit 0、APPROVE |
| `task027_repair_react_review` | final hooks/a11y adversarial review | none | completed / integrated | navigation bounce findings修正後 APPROVE |
| `task027_repair_security` | final OWASP/data-integrity review | none | completed / integrated | C/H/M 0/0/0、APPROVE |
| `task027_repair_clinical` | final EMR clinical safety | none | completed / integrated | death-state/zero-write/history-edit review、APPROVE |

Cancelled rolesは全てread-onlyでintegrated editなし。completed rolesのeditもなし。各 phaseのwriteは一人だけが担当し、foreign WIPへ触れなかった。

## Final full-commit reviewer results

- Initial full-commit reviewers were all joined. The first clinical/React/security falsification pass returned the actionable findings recorded as FS-9〜FS-11; no finding was waived。
- `task027_final_clinical/task027_clinic_isolation`: revision path C/H/M `0/0/0` / APPROVE。
- Fresh repair reviewers (`task027_repair_go_review`, `task027_repair_isolation`, `task027_repair_react_review`, `task027_repair_security`, `task027_repair_clinical`) all returned C/H/M `0/0/0` / APPROVE after the repair commit candidate stabilized。
- Every launched role is completed/joined or explicitly cancelled; no role remains pending。

## Git worktree reconciliation

Resume re-anchor evidence:

```text
eb7db0dc9 fix: validate pet death dates before recording
```

After Slice A, another owner advanced HEAD with `fc3c12b28 fix: fail closed on unknown pet status`. This run did not reset/rebase it; Slice B was re-anchored and reverified on top. After Slice C, foreign commits `dee0d39fb` and `617f6f9bf` advanced HEAD; the run again re-anchored without reset/rebase and reverified before repair. After repair, docs-only foreign commit `fc1cc5a8e` advanced HEAD and absorbed `bug.md`; repair ancestry remained true. Final pre-admin ancestry is `eb7db0dc9 -> 046615f4b -> fc3c12b28 -> 1dd1cf04e -> c161baffb -> dee0d39fb -> 617f6f9bf -> fb0cf9c91 -> fc1cc5a8e`。

Recorded start-of-resume foreign projection:

```text
 M bug.md
 M frontend/src/components/shared/ReservationFormModal/ReservationFormModal.test.tsx
 M frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx
```

Final code-commit status:

```text
 M todo.md
?? reports/2026-08-03-task-027-examination-revisions.md
```

Slice C直前 staged listは23 owned pathsだけで、foreign files、`todo.md`、reportは不包含。Repair直前 staged listも7 owned pathsだけで、`bug.md`、`todo.md`、reportは不包含。着手時のReservationFormModal 2 pathsは本 unitが触れないまま別ownerの`617f6f9bf`へ入り、`bug.md`もrepair後に別ownerの`fc1cc5a8e`へ入った。最終foreign WIPはなく、`git diff --cached --name-only` は空。report probe:

```text
$ git check-ignore -v reports/2026-08-03-task-027-examination-revisions.md; echo "ignored exit code is $?"
ignored exit code is 1
```

`claim/TASK-027` は live。agentは削除しない。main統合後にUSERが解放する。

Formatter deviation後のdependency-file probe:

```text
$ git diff --name-only -- package.json pnpm-lock.yaml frontend/package.json frontend/pnpm-lock.yaml
[no output]
```

## Remaining risks and follow-ups

- 004 migrationはruntimeへ未適用。統合・pull後にユーザーが手動で適用するまでschema/runtime機能は利用不可。
- package lint baseline 62件はTASK-027外も含む既存負債。TASK-027でrepo-wide lint greenとはしていない。
- browser E2E、full repo test/build/typecheck、external provider run、staging/production smokeは未実行。保存promptのscoped gateだけを完了。
- dependency vulnerability scanはfinal security repair reviewerでは未実行。依存ファイルは本 unitで変更しておらず、脆弱性ゼロは主張しない。
- existing `useActionState` / Radix test warningsは残るがfailureではない。
- formatter harness deviation: `npx prettier` がcontainer cacheへ一時取得した。package/lockfile変更なし。今後はrepoにformatterが存在することをpreflightしてから実行する。

## Harness Improvement Feedback

1. Resume promptが official pinned golangci imageへ修正されたことで、旧 `executable not found` blockerを正しく解消できた。
2. formatter commandもlint同様にbinary preflightを明示すべき。`npx` は未導入packageを自動取得するため、このrepoでは許可 commandにしない。
3. baseline-red lintを扱うpromptは、parent baselineとの差分比較方法と「touched fileの既存 finding」の判定語彙を最初から規定すると、repo-wide greenへの誤拡張を防げる。
4. frontend scoped gateへ shared `PatientSelectionTable.test.tsx` を明示追加すると、feature外共有componentの変更もbinding evidenceに含められる。
5. final clinical falsificationで死亡患者のrecord-derived writeとquery readiness/cross-record state漏れが見つかった。今後のgeneratorは「FE表示guard」だけでなく、derived relationのserver-side再検証とA→B→A query-cache遷移を明示ケースにするべき。

## ユーザーが手で実行するコマンド

統合・pull後、開発者が手動で実行する。

```bash
make migrate
```

本 unitが実行しなかった適用/外部 action: migration apply、seed apply/read/write、push、PR作成、merge、Issue close、claim branch delete、staging/production deploy。
