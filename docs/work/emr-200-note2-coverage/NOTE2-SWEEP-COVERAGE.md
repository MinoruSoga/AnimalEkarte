# NOTE2-SWEEP-COVERAGE — 未確認 route×operation 受入カバレッジ（Plane `EMR-200`）

> Current task status is tracked in Plane `EMR-200`. This local sheet preserves the route×operation
> coverage matrix and the evidence pointers. See the [migration receipt](../plane-md-migration-20260923-receipt.md).
> 索引の正本は [todo-verification.md](../../../todo-verification.md)、履歴は [bug-2.md](../../../bug-2.md#plan-note2-coverage)。

- campaign: `emr-200-note2-coverage` revision 1
- unit: `NOTE2-SWEEP-COVERAGE`（Plane `EMR-200`）
- 照合 revision: `923bb99ba`（`docs: migrate markdown tasks to Plane`）。worktree ブランチ `MinoruSoga/emr-200`
- 状態: **カバレッジ表 完了。API/契約レベルは scoped 実DBテストで PASS、ブラウザ UAT は未実施（UNKNOWN/BLOCKED）**

## 目的とスコープ

[bug-2.md#plan-note2-coverage](../../../bug-2.md#plan-note2-coverage) の `NOTE2-SWEEP-COVERAGE` を
現行 route inventory と突合し、**未確認だった detail・入院・検査・カルテ/健診の route×operation** を
1 行ずつ「fixture 必須項目 / 操作 / 期待結果 / 証拠 / 判定」で固定する。

**除外（再実装しない）**: 既に修正済み・再検証済みの項目はやり直さない。

- `BUG2-MR-ENTERED-BY-CLINIC` → 修正済み `b58a77fd9`。本票は「actor 修正後に新規カルテ→再読込→健診経路を再確認する」受入のみを担う。
- `BUG2-PAYMETHOD-CREATE-FORBIDDEN` → SPEC-OK 済み。
- `BUG2-RES-DIALOG-A11Y` → FIXED/reverified 済み。

**混同禁止**: 82 ページ到達（route reachability）を CRUD 完了と扱わない。到達性と書込み/読取の成立は別判定にする。

## 判定の読み方

| 判定 | 意味 |
|:--|:--|
| **PASS（API/契約）** | scoped 実DBテスト（`ekarte_db_test`）が現行 revision で成功。handler/service/repository 契約レベルの成立証拠 |
| **PASS（到達）** | Playwright spec 定義（allowlist）が route 到達・フォーム表示を対象化。実行結果ではない |
| **UNKNOWN** | 現行 revision で実行証拠を照合していない |
| **BLOCKED** | 前提（fixture・環境・承認）が不足し実行できない |

API/契約レベルの証拠は「ブラウザ UAT」「本番/STG 準備」ではない。ブラウザ操作の成立は別レーン（下記「残件」）。

## 証拠の実行方法（再現手順）

対象 revision の backend コンテナ内で scoped に実行する（テスト DB は `ekarte_db_test`。共有 `ekarte_db` は使わない）。

```sh
docker exec <backend-container> go test ./internal/medicalrecord/ -count=1 -v -run '<regex>'
docker exec <backend-container> go test ./internal/billing/ ./internal/inventory/ -count=1 -v -run '<regex>'
```

`backend/internal/testdb/testdb.go` は `DB_NAME` + `_test`（= `ekarte_db_test`）へ接続する。
`001_init.sql` は適用しない GORM AutoMigrate double のため、実 DDL 固有の制約は
`medical_record_entered_by_clinic_fk_test.go` のように明示 DDL を当てて再現する。

## カバレッジマトリクス

列は [todo-verification.md](../../../todo-verification.md) の共通列（`ID / case / revision / 環境・fixture参照 / 操作 / 期待値 / 実際値 / 証拠参照 / 後処理 / 判定`）に準拠。

### A. カルテ新規 → 再読込 → 健診（actor 修正後の再確認）

| ID | case | fixture 必須項目 | 操作 | 期待結果 | 実際値 | 証拠参照 | 判定 |
|:--|:--|:--|:--|:--|:--|:--|:--|
| EMR200-MR-CREATE | `/medical-records/new` → `POST /api/v1/medical-records` | `owner_id`(string)・`pet_id`(string) 必須。任意: `visit_date`(YYYY-MM-DD)・`visit_type`・`doctor_id`・`status`(draft/finalized)・`recommendation_reason`。actor は認証由来の staff ID | 選択医院 B に有効所属＋B の create 権限を持つ本人で、B の飼主・ペットに対し POST | 201。`entered_by` は本人 staff ID。再取得・一覧・詳細・件数が整合 | 実DBテスト PASS（actor 契約） | `medical_record_entered_by_clinic_fk_test.go`: `TestEnteredBy_AssignedCreator_SucceedsAfterSingleColumnFKAndActorValidation` / `TestEnteredBy_SameClinicCreate_StillWorks` | **PASS（API/契約）** |
| EMR200-MR-CREATE-OLD | 旧複合 FK の再現 | 主所属 A・選択医院 B | 旧 `(entered_by, clinic_id)` FK 下で B 作成を再現 | 旧 FK が拒否（バグ再現）→ 単列 FK + actor 検証で成功 | 実DBテスト PASS | 同ファイル: `TestEnteredBy_LegacyCompositeFK_RejectsHomeClinicAActorOnClinicB` | **PASS（API/契約）** |
| EMR200-MR-ACTOR-GUARD | actor 認可境界 | 非所属／B の create 権限なし／system admin（assignment なし） | 上記 actor で POST | 非所属・無権限は既定認可で拒否しカルテを残さない。assignment なし system admin は既存仕様で許可 | 実DBテスト PASS | 同ファイル: `TestEnteredBy_UnassignedActor_RejectedFailClosed` / `TestEnteredBy_SystemAdminWithoutAssignment_Allowed` / `TestEnteredBy_AutoCreateNilActor_SkipsAssignmentGate` | **PASS（API/契約）** |
| EMR200-MR-RELOAD | `/medical-records/:id` → `GET /api/v1/medical-records/:id` | 有効な medical record ID | 作成後、一覧を経由せず詳細を再取得 | 作成内容が再読込で整合。他院 ID は 404 で本文非漏洩 | 実DBテスト PASS | `realdb_records_isolation_test.go`: `TestRealDB_RecordsSelectedClinicBGrantAIsolation`（`cross_get_*` subtest） | **PASS（API/契約）** |
| EMR200-CHECKUP-CREATE | 健診タブ → `POST /api/v1/medical-records/:id/checkups` | `checkup_type_id`(必須)・`date`(必須)。親 medical record ID | 作成したカルテ配下に健診を POST | 201。`GET /:id/checkups` と `checkup_field_results` に反映 | 実DBテスト PASS | `checkup_repository_test.go`: `TestCheckupRepository_Create` / `TestCheckupRepository_FindByMedicalRecordID`; `realdb_records_isolation_test.go`（`checkups_*` / `checkup_field_results_*` subtest） | **PASS（API/契約）** |
| EMR200-CHECKUP-STANDALONE | `POST /api/v1/checkups` | — | 単独健診作成を試行 | **ルートなし**。健診は `POST /medical-records/:id/checkups` 経由のみ（`checkups` group は GET のみ） | ルート定義で確認 | `backend/internal/medicalrecord/routes.go` `registerVaccinationAndCheckupRoutes` | **PASS（設計どおり）** |
| EMR200-MR-CHECKUP-UI | `/checkups/new`（ペット選択→フォーム） | `petId` query | フォーム表示・保存導線 | フォーム表示（実書込みは別レーン） | spec 定義のみ | `frontend/e2e/checkups-flow.spec.ts` / `medical-records-create.spec.ts`（後者は synthetic interceptor で POST をローカル fulfill。実 DB 書込みなし） | **PASS（到達）/ 書込み UNKNOWN** |

### B. 入院（cage 依存）

| ID | case | fixture 必須項目 | 操作 | 期待結果 | 実際値 | 証拠参照 | 判定 |
|:--|:--|:--|:--|:--|:--|:--|:--|
| EMR200-HOSP-CREATE | `/hospitalization/new` → `POST /api/v1/hospitalizations` | `owner_id`・`pet_id`・`hospitalization_type`(hospitalization/hotel)・`start_date`・`end_date` 必須。**`cage_id` 必須**（BUG-037: `nil`/`0` は invalid input）。任意: `doctor_id`・`status`・`treatment_plans` | 有効 cage を持つ選択医院で入院を POST | 201。`end_date >= start_date` をアプリ境界で検証。cage 未指定は 400 | 実DBテスト PASS（cage 必須・原子性・境界） | `hospitalization_create_tx_atomicity_test.go`: `TestHospitalizationService_Create_NestedPlansAtomicity`; `hospitalization_service_test.go`: `TestHospitalizationService_Create` / `_RejectsDeceasedPet` / `_InsuranceFields`; 契約は `hospitalization_request.go` L137–140 | **PASS（API/契約）** |
| EMR200-HOSP-CAGE-PREREQ | `POST /api/v1/masters/cages` | `name`・`cage_type`(icu/dog/cat/general)・`cage_size`(small/medium/large) | 選択医院に cage を作成 | 201。`GET /masters/cages/:id` で取得可。他院不可 | 実DBテスト PASS | `cage_repository_test.go`: `TestCageRepository_Create_And_FindByID` / `_FindAll_TypeFilterAndClinicIsolation` | **PASS（API/契約）** |
| EMR200-HOSP-DETAIL | `/hospitalization/:id` → `GET /api/v1/hospitalizations/:id` | 有効な入院 ID | 一覧を経由せず詳細を取得 | 200。他院 ID は非漏洩 | 実DBテスト PASS | `realdb_hosp_isolation_test.go`: `TestRealDB_HospSelectedClinicBGrantAIsolation` | **PASS（API/契約）** |
| EMR200-HOSP-EDIT | `/hospitalization/:id/edit` → `PATCH /api/v1/hospitalizations/:id` | 有効な入院 ID ＋ edit 権限 | 編集画面で PATCH | 更新が反映。日付逆転はアプリ境界で拒否 | 契約は `hospitalization_request.go` `updateHospitalizationRequest`。**実行証拠は未照合** | 契約: `updateHospitalizationRequest.toServiceInput`（MRB-06） | **UNKNOWN** |
| EMR200-HOSP-UI | `/hospitalization`（一覧/ボード）・`/hospitalization/new` | 有効 cage（clinicale2e fixture は cage を生成しない） | 一覧・ペット選択・ステータスタブ | 一覧/ボード表示、`/hospitalization/select-pet` 遷移 | spec 定義のみ | `frontend/e2e/hospitalization-flow.spec.ts`; `docs/ops/testing/scenarios/S05-hospitalization-cycle.md`（cage・active ケアマスタ前提） | **PASS（到達）/ 実書込み BLOCKED**（cage を UI 前に合成作成できない限り、S05 依存ケースのみ BLOCKED） |

### C. 検査

| ID | case | fixture 必須項目 | 操作 | 期待結果 | 実際値 | 証拠参照 | 判定 |
|:--|:--|:--|:--|:--|:--|:--|:--|
| EMR200-EXAM-CREATE | `/examinations/new` → `POST /api/v1/examinations` | `exam_type_id`(必須)・`date`(必須)。任意: `medical_record_id`・`pet_id`・`doctor_id`・`result_summary`・`machine`・`status`(pending/in_progress/result_entered/completed/confirmed)・`items[]`（各 `name` 必須） | 有効な exam type を選び検査を POST | 201。項目はマスタ基準値から算出。他院 exam type/field は拒否 | 実DB・handler テスト PASS | `examination_handler_test.go`: `TestCreateExamination`; `examination_atomic_write_test.go`; `examination_cross_tenant_master_fk_write_test.go`: `TestExaminationService_Create_RejectsCrossClinicExamType`; `exam_reference_range_resolution_test.go`: `TestExaminationService_CreateUsesMasterRanges`; `examination_service_test.go`: `TestExaminationService_Create` / `_RejectsInvalidClinicalRelations` | **PASS（API/契約）** |
| EMR200-EXAM-400 | 旧 400 の切り分け | exam type 取得後の POST | 必須フィールド欠落で POST | `exam_type_id`/`date` 欠落は 400（契約バリデーションとして正常） | handler テスト PASS | `examination_handler_test.go`: `TestCreateExamination/returns_400_when_required_field_missing` | **PASS（契約）** |
| EMR200-EXAM-DETAIL | `/examinations/:id` → `GET /api/v1/examinations/:id` | 有効な検査 ID | 詳細を取得 | 200。他院 ID は非漏洩 | 実DBテスト PASS | `realdb_examination_isolation_test.go`: `TestRealDB_ExaminationSelectedClinicBGrantAIsolation` | **PASS（API/契約）** |

### D. 詳細画面（会計・在庫）

`/accounting/:id`・`/inventory/:id` の到達と権限境界は [S31-detail-route-direct-access.md](../../ops/testing/scenarios/S31-detail-route-direct-access.md) を正本とする。本票は API/契約レベルの証拠を追加する。

| ID | case | fixture 必須項目 | 操作 | 期待結果 | 実際値 | 証拠参照 | 判定 |
|:--|:--|:--|:--|:--|:--|:--|:--|
| EMR200-ACCT-DETAIL | `/accounting/:id` → `GET /api/v1/accountings/:id` | 有効な会計 ID（選択医院） | 詳細を取得 | 200。view 権限必須 | handler テスト PASS | `accounting_handler_test.go`: `TestGetAccounting` | **PASS（API/契約）** |
| EMR200-INV-DETAIL | `/inventory/:id` → `GET /api/v1/inventory/:id` | 有効な在庫 ID（選択医院） | 詳細を取得 | 200。view 権限必須。他院/seed ID は 403/404 で本文非漏洩 | handler・実DBテスト PASS | `inventory_handler_test.go`: `TestGetInventory`; `inventory_service_test.go`: `TestInventoryService_GetByID`; `repository_test.go`: `TestInventoryRepository_FindByID`; `realdb_selected_clinic_b_grant_a_isolation_test.go`: `TestRealDB_Inventory_Get_MembershipABGrantASelectedB_Returns403` / `TestRealDB_Inventory_Get_SelectedB_ASeedID_Returns404NoABody` | **PASS（API/契約）** |

## bug-2.md#plan-note2-coverage との突合（reconciliation）

| 旧記載（2026-09-13） | 現行 route inventory | 本票の扱い |
|:--|:--|:--|
| `/accounting/:id`・`/hospitalization/:id`・`/hospitalization/:id/edit`・`/inventory/:id` は前提 ID 未確保で skip | FE route 定義は現行も存在（`accounting-routes.tsx` L65、`clinical-care-routes.tsx` L131/139、`clinical-business-routes.tsx` L44） | 詳細到達は S31 が正本。API 側は D 節で PASS。`/hospitalization/:id/edit` の PATCH 実行のみ UNKNOWN |
| 入院は `cage_id` 必須で追加確認未完了 | `createHospitalizationRequest` は現行も `cage_id` 必須（BUG-037） | B 節。cage 前提は PASS、UI 書込みは BLOCKED（cage 合成が UI 前段） |
| 検査 create は examination-types 取得後も 400 | 必須は `exam_type_id`・`date` | C 節。400 は契約バリデーションで正常 |
| カルテ新規は BUG2 でブロック | `b58a77fd9` で actor 契約修正済み | A 節。actor 契約は PASS、UI 書込みは別レーン |
| 「82 ページ到達」 | 到達 ≠ CRUD | 到達（PASS（到達））と API/契約（PASS）を分離 |

## 残件（本票では閉じない）

| 残件 | 種別 | 理由 |
|:--|:--|:--|
| カルテ新規 → 再読込 → 健診のブラウザ実操作 | UNKNOWN | 実ブラウザ＋実 backend への POST が必要。`medical-records-create.spec.ts` は synthetic interceptor で POST をローカル fulfill するため実 DB 書込みを検証しない |
| 入院新規のブラウザ実操作 | BLOCKED | `clinicale2e` fixture は cage を生成しない。有効 cage を先に合成作成する前提が UI 実行前に必要（依存ケースのみ BLOCKED） |
| 検査新規のブラウザ実操作 | UNKNOWN | 同上（実書込みレーン未実施） |
| `/hospitalization/:id/edit` PATCH の実行証拠 | UNKNOWN | 契約は確認、実行未照合 |
| 上記を実行する環境 | BLOCKED | 実行中 stack は main checkout の共有 dev stack（`APP_ENV=development`・共有 `ekarte_db`）。`clinicale2e` lane は `APP_ENV=test` を要求し、共有 DB への書込みは規約で不可（[todo-verification.md](../../../todo-verification.md) L9/L229）。実行には承認済みの専用 `APP_ENV=test` stack が必要 |

## 環境ゲート（ブラウザ実書込みレーン）

[todo-verification.md](../../../todo-verification.md) L9「他タスクの起動中コンテナや共有DBを検証先に流用しない」・L229「共有 `ekarte_db` … をテスト DB に使わず」に従う。

| ゲート | 合格条件 | 現状 |
|:--|:--|:--|
| stack | 専用 `APP_ENV=test` の local stack（FE `:3003`・BE `:8080`） | 実行中 stack は `development`。**BLOCKED** |
| DB | disposable な検証先（`ekarte_db_test` または専用 DB） | 実書込みは共有 `ekarte_db` になるため不可 |
| fixture | 承認済み合成 clinic（`clinicale2e`）＋ cage 等の追加前提 | cage が未生成 |
| 後処理 | deterministic teardown（`clinicale2e.Delete`） | 実行時に付与 |
| 証拠保存先 | gitignore 対象 `reports/uat-YYYY-MM-DD/` | 本票では run なし |

## 変更サマリ

- 新規: 本ファイル（route×operation カバレッジ表）。
- 製品コード変更: なし。migration 追加・適用: なし。共有 DB への書込み: なし。
- 実行した検証: backend コンテナ内 scoped `go test`（`./internal/medicalrecord/`・`./internal/billing/`・`./internal/inventory/`、`ekarte_db_test`）。
