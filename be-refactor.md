# backend コード規約再監査（2026-09-03 full file audit）

`backend/` の production `.go` を規約正本に 1 ファイルずつ照合した。コードは修正していない。

- HEAD: `aeaa3b08472a6ee4d86fcda56f977caada063d7e`
- 母集団: `git ls-files 'backend/**/*.go' | grep -v '_test.go$' | grep -v '/_archive/' | sort` → **981**（欠番 0）
- 規約正本: `backend/CLAUDE.md`、`backend/CODING_RULES.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/go-gin-backend-review.md`、`.claude/refs/backend-application-invariants.md`、`.claude/refs/error-handling.md`、`.claude/refs/naming-conventions.md`、`.claude/refs/go-language.md`、ADR-006、`docs/architecture/be9-2a-boundary-map.md`
- 方法: 隔離 worktree で全 production パスをレーン分割精読（auth/clinic/staff、owner/pet/identitylink、reservation/trimming、billing/inventory、medicalrecord×2、lstep、model+cmd+cross-cutting）。残差 ID は現行コードで再実測。旧 DONE は未修正として再掲しない
- 各所見の `path` は `backend/` 起点。行番号は `aeaa3b084` 時点
- Handler → Service → Repository / Clean Architecture は Go/Gin 公式要件ではない。再導入しない
- カバレッジ行: OK 702 / FINDING 277 / SKIP 2

---

## 1. フォルダ／ファイル構成

`backend/internal/` は ADR-006 の 14 target domain（owner / pet / staff / auth / reservation / trimming / medicalrecord / billing / inventory / lstep / clinic / manualarticle / httpapi / identitylink）と命名済み cross-cutting（config / dbconn / middleware / infra / model / timeutil / seedbundle / logger / csvimport / labdeviceagent / authjwt / apperrors / apicontract / lintscan / audit / persistence / scheduler / sharedkernel / textsearch / testdb）で構成される。

`backend/cmd/` は `api`（composition root、22 file）、`migrate`、`csv-import*`、`lab-device-agent`、`lstep-migrate`、`seed-export`、`staff-provision`、`stg-uat-*`、`coverage-ratchet`。`cmd/_archive` は母集団外。

**合格**

| 項目 | 実測 |
|---|---|
| 旧 `internal/handler` `internal/service` `internal/repository` ディレクトリ | 不存在（`test ! -d` 3 つとも成功） |
| domain 内 `handler/` `service/` `repository/` サブパッケージ | 0 |
| production `package util\|common\|misc` | 0。`timeutil` / `sharedkernel` は固有名 |
| 極端な package サイズ | `medicalrecord` 239、`lstep` 133、`model` 94、`billing` 93、`reservation` 84。800 行超 production ファイル 0（最大 672: `medicalrecord/lab_device_receive_service.go`） |
| `cmd/api` | domain を直接 import する composition root。業務ルールの複製ではない |

フォルダ構成の逸脱（新設すべき util/common、layer 再分割）は提案しない。

---

## 2. HIGH（開いた所見のみ）

#### BE-RC-026 [HIGH] admin 予約 Preload が `Doctor` / `CreatedByStaff` を clinic 所属で scope していない
- 対象: `internal/reservation/appointment_admin_repository.go:53,74-75,120,136`
- 現状: 本体は `Scopes(persistence.ClinicScope(clinicID))`。`ReservationType` / `Owner` / `Pet` / `LineCustomer` には `clinic_id` がある。`Preload("Doctor"|"CreatedByStaff", "deleted_at IS NULL")` だけ所属条件が無い
- 規約: nested Preload は clinic-owned の中間 association も独立 clinic predicate。破損 FK 経由で他院 staff 行を復元し得る
- 対比: `internal/reservation/reservation_repository.go:15-19,248-250` の `staffAssignedToClinicsCond`
- 改善案: admin の該当 Preload に `staffAssignedToClinicsCond`（と必要なら parent clinic fail-close）を付ける。`staffs.clinic_id` 単独では多院所属を落とす。一括リネームや layer 分割はしない

---

## 3. MEDIUM（開いた所見のみ）

#### BE-RC-004 [MEDIUM][既知 residual] 他マスター Delete が条件付き原子 DELETE 未達
- 規約: 正しさは `clinic_id + id` と usage 不在を同一 SQL に束ねた条件付き原子 DELETE。Find→Count→Delete を防御本体にしない。**一括 retrofit はしない**（CODING_RULES）。触る Delete で直す
- inventory `DeleteIfUnused` と payment_method / insurance の原子 DELETE は完了済みのためここでは再掲しない。service 側 Count は UX のみ
- **tx なし Count→Delete**: `medicalrecord/consultation_service.go:220-228`、`vaccine_service.go:212-220`、`checkup_type_service.go:200-208`、`diagnosis_service.go:346-358`、`chief_complaint_service.go:140-148`、`hospitalization_plan_service.go:188-196`、`inquiry_template_service.go:147-155`、`reservation/reservation_type_group_service.go:146-154`、`reservation_type_service_core.go:159-167`
- **tx 内だが同一 SQL ではない**: `procedure_service.go:239-263`、`exam_type_service.go:207-230`、`cage_service.go:199-215`、`medicine_service.go:451-458` + `medicine_service_delete.go:17-37`、`staff/occupation_service.go:164-178`、`auth/permission_group_service_mutate.go:297-305`、`pet/animal_species_service.go:202-221`、`clinic/clinic_service.go:371-405`、`staff/staff_service_core.go:395-419`、`reservation_type_liff_repository.go:116-154`
- **tx 内 Delete→Count（残があれば rollback）**: `inventory/merchandise_item_service.go:225-238`、`trimming/trimming_course_service.go:206-217`、`trimming_option_service.go:159-170`、`trimming_course_type_service.go:132-143`
- 改善案: 当該 Delete を変更する PR で `DeleteIfUnused` 型へ。汎用 helper package は作らない。owner は各 master の既存 repository

#### BE-RC-005 [MEDIUM] service の `slog.ErrorContext` と handler `RespondError` が 5xx を二重ログする
- 規約: 同じ error を複数層で重複ログしない。未知 pg だけ request 境界の `c.Error` 重複を例外許容
- 現状: production 191 ファイルに `slog.ErrorContext`。`*service.go` は 88/108。5xx は `httpapi/response.go:15-19` が `c.Error` → `middleware/logging.go:51-52` が再記録
- 代表: `billing/insurance_service.go:148-167`、`inventory/inventory_service.go:160-181`、`lstep/lstep_lifecycle_service.go`、`medicalrecord/diagnosis_service.go`、`reservation/reservation_staff_service.go`
- 改善案: 既知 4xx は service でログしない。5xx は middleware 一本化。新規から増やさない。一括削除は別タスク

#### BE-RC-009 [MEDIUM] 実装側に定義された広い `XxxRepository` が残る
- 規約: interface は利用側の最小メソッド。実装は concrete を返す
- 現行: `identitylink.Repository` ~28（`identitylink/repository.go:19-61`）、`medicalrecord.MedicalRecordRepository` ~22（`medical_record_repository.go:47-107`）、`billing.AccountingRepository` ~22（`accounting_repository.go:131-169`）、`staff.StaffRepository` 18（`staff_repository.go:20-50`）、`clinic.ClinicRepository` 12（`clinic/ports.go:37-50`、`UpdateClinic` 化済み）
- 正例: `lstep/composition.go` の consumer-side 最小 port
- 改善案: 新規は呼び出し側で切る。既存の一括分割はしない

#### BE-RC-014 [MEDIUM][residual] pgx encode 判定が `err.Error()` 文字列 Contains
- 対象: `internal/apperrors/errors.go:344-381` — `isPgxEncodeRangeMessage(err.Error())`
- 規約: `errors.Is` / `As`。BUG-138 の既知例外（pgx が typed error を出さない）
- 改善案: 新規に同パターンを増やさない。pgx が typed error を出したら `errors.As` へ

#### BE-RC-024 [MEDIUM] 先行 `max` 対応の外に HTTP 境界の長さ上限欠落が残る
- LINE 本文・返金理由・カルテ追記・飼主/ペット氏名 leftover 等は完了済み。未修正として再掲しない
- 残件（現行 path）:
  - `clinic/clinic_request.go:47-55,74-91` — `Name` / `Address` / `DirectorName` / `Website` / `FooterNote` に `max` なし
  - `clinic/company_request.go:4-14`、`closing_settings_request.go` の `Note`、`clinic_holiday_handler.go` の `Reason`
  - `billing/insurance_request.go` `Name`、`payment_method_master_request.go` `Name`、`accounting_request.go` `OtherReason` / `InsuranceName`、`billing_item_request.go` `OtherReason`、`cash_register_request.go:59` `Memo`、`estimate_request.go` PATCH `Title`
  - `inventory/inventory_request.go` `Name`/`Unit`/`Location`、`merchandise_item_request.go` `Name`
  - `pet/animal_species_request.go` `Name`、`chronic_condition_request.go`、`pet_request.go` list `Search` / `PetNumber`、`owner/http_request.go` list `Search` / `HomePostalCode`
  - `identitylink/handler.go:57,73` 検索 `q`
- 改善案: 同系統の正例（name 255、reason 500、memo 1000）。一括 sweep は別タスク。`util` 検証 package は作らない

#### BE-RC-025 [MEDIUM] LIFF スタッフ一覧が `is_active` を見ない
- 規約: LIFF で明示指定された staff は clinic 所属・対応可能種別に加え `is_active=true` かつ `reservation_visible=true`
- 書込は充足: `reservation_staff_capability_validator.go:57`
- 一覧/空き枠は `ReservationVisible` のみ: `liff_service_availability_staff.go:19,46-48,67`、`liff_service_catalog.go:112-122`、`liff_service_availability.go:43`
- 改善案: filter に `IsActive` を足す。inactive を出さない。書込側の二重防御は残す

#### BE-RC-028 [MEDIUM] LSTEP バッチに silent skip が残る
- 規約: 成立条件欠落は fail-closed。intentional best-effort でも Failed 計上とログ。`return 0, nil` で成功に見せない
- 対象:
  - `lstep/lstep_batch_delivery.go:26-27` — `lstepDeliveryTrigger == nil` で `return 0, nil`
  - `lstep/lstep_tag_sync_service.go:279` — `shouldSkipSync` が `settingsSvc == nil` を `(true, nil)`（dormant は同条件 fail-closed）
  - `lstep/lstep_lifecycle_service.go:177` — `clearAllTagsIfLinked` が `FindByID` 失敗をログなし return
- 改善案: 必須依存欠落は fail-closed か Partial/Failed 計上。副次 cleanup は少なくとも slog。汎用 batch helper package は作らない

#### BE-RC-032 [MEDIUM] payment_method 以外の update 後 reload が成功を失敗へ反転し得る
- payment_method の同一 tx reload は完了済み。未修正として再掲しない
- 残件: `billing/insurance_repository.go:61-65` — `UpdateScopedByID` の後 `FindByID` を自前 Transaction で括らない。`billing/accounting_service_correction.go:67-72` — `WithTx` commit 後の `FindByID`
- 改善案: 同一 tx で update+reload、または reload 失敗時は更新済みエンティティを返す。owner は billing

---

## 4. LOW（開いた所見のみ）

#### BE-RC-015 [LOW] package.Type stutter が系統的
- 例: `clinic.ClinicService`、`auth.AuthService`、`reservation.ReservationHandler`、`trimming.TrimmingService`、`lstep.LstepSettingsHandler`、`pet.PetResponse`、`staff.StaffRepository`
- 改善案: 新規/触る面から stutter を避ける。一括 rename はしない

#### BE-RC-017 [LOW] 同一 package の exported `Update(..., map[string]any)` が常態
- 機械検出 49 production ファイル。inventory / billing / medicalrecord / pet / auth permission_group / trimming / staff / reservation type 等
- 001/008 の境界（reservation 向け staff、consumer `ClinicRepository`）は typed 済み。回帰なし
- 改善案: 触る repository から unexported `update` にし、外には typed command だけ出す。一括 unexport はしない

#### BE-RC-019 [LOW] `medicalrecord` 本番 239 file の凝集圧
- layer サブパッケージ化は禁止方針どおり避けている。800 行超ファイルなし（最大 672）
- 改善案: 分割するなら業務能力（lab / hospitalization）単位。急がない

#### BE-RC-021 [LOW] exported GoDoc 欠如は系統的
- revive `exported` / `package-comments` は `.golangci.yml` で disable 継続
- 例: `config.Config` / `Load`、`clinic.ClinicService`
- 改善案: 新規 export には GoDoc。一括はしない

#### BE-RC-023 [LOW] `init()` で gin validator をグローバル登録
- 対象: `internal/clinic/clinic_request.go:18-35`（`jp_email` / `jp_phone` / `jp_postal`）
- 改善案: テスト順副作用が出たら constructor 登録へ。現状は実用上問題なし

#### BE-RC-027 [LOW] `LineLinkToken.Token` が `json:"token"`
- 対象: `internal/model/line_link_token.go:10`。永続値は digest。HTTP は別 DTO（`lstep/line_link_handler.go` の `linkTokenResponse`）で raw token を返す
- 対比: `PasswordResetToken.TokenHash` は `json:"-"`
- 改善案: モデル側を `json:"-"` にして誤 marshal に備える。owner は `model`（タグ）と `lstep`（発行）

#### BE-RC-030 [LOW] BE9 移行コメントが削除済み `internal/service` を現行配置として残す
- 022（`replace_audit_tail.go`）は完了。**022 回帰ではない**
- 対象: `medicalrecord/pagination.go:6`、`service_deps.go:13-16` が `internal/service|handler|repository` を pending 配置として記述
- 改善案: コメントを現行 domain package に直す

#### BE-RC-031 [LOW] 検査機器 receive が transactor nil でも persist を続ける
- 対象: `medicalrecord/lab_device_receive_service.go:47-51`、`lab_device_exam_persist.go:58`
- 規約: lock / 親カルテ整合に tx が必要なら ambient 不在は fail-closed
- 改善案: composition が必ず transactor を渡し、nil は拒否。ローカル agent 専用ならコメントで contract を固定する

---

## 5. 再検証した合格（未修正として再掲しない）

| 項目 | 根拠 |
|---|---|
| BE-RC-001 typed staff update | `staff/reservation_staff_update.go`、`UpdateForReservation(..., ReservationStaffUpdate)`。回帰なし |
| BE-RC-002 閉集合の `max` | LINE text、refund reason、addendum、OwnerName/Pet Name leftover 等は `max` 維持。残件は 024 |
| BE-RC-003 原子 DELETE 3 面 | inventory `DeleteIfUnused`、payment_method / insurance 原子 DELETE。service Count は UX |
| BE-RC-006 列挙 handler `err.Error()` | leftover 閉集合に連結なし。残件 CSV 経路は固定日本語 |
| BE-RC-007 payment_method reload | 同一 `Transaction` で update+reload。残件は 032 |
| BE-RC-008 `UpdateClinic` | consumer `ClinicRepository` から map `Update` 除去 |
| BE-RC-010 `LstepRepository` | exported 0。`LifecycleOwnerRepository` |
| BE-RC-012 wrapcheck `internal/*` wildcard | 廃止済み。残るのは明示 package 列挙（§6） |
| BE-RC-013 vaccination `t.Setenv` | `os.Setenv` 0 |
| BE-RC-016 stale `// Package handler\|service\|repository` | production 0 |
| BE-RC-018 staffs/shift_entries AST gate | `staff/staff_table_write_owner_lint_test.go` 存在 |
| BE-RC-022 `replace_audit_tail` | `internal/service` を正本扱いしない。他ファイルの古いコメントは 030 |
| BE-RC-011 見積通常 CRUD 監査 | 意図的 post-commit best-effort。CreateSuccessor のみ fail-closed。再提案しない |
| BE-RC-020 `nested_summary_response.go` | billing / medicalrecord / reservation の意図的コピー。統合しない |
| 旧 3 layer directory | 不存在 |
| ADR-006 DAG / appointments write owner | AST gate 維持。trimming は typed intent |
| `ShouldBind*` | 本番 `BindJSON`/`MustBind*` 0 |
| Context struct 保持 / 生 `*gin.Context` goroutine | 未検出 |
| CORS wildcard + credentials | なし |
| AutoMigrate | testdb のみ |
| 本番 800 行超 | 0 |

---

## 6. lint / 設定上の既知緩和

- wrapcheck: `internal/*` wildcard は廃止（012 DONE）。未触 package は `.golangci.yml:110-137` で明示 ignore。`trimming` も ignore のまま。**一括 ignore 解除はしない**（新 ID にしない。設定変更は Out of scope）
- `gocritic` の `hugeParam` / `unnamedResult` / `rangeValCopy` disable（composition/DTO）
- `revive` の `exported` / `package-comments` / `context-as-argument` / `unexported-return` disable
- `contextcheck` は `internal/middleware` と `cmd/api` で除外
- `cmd/csv-import*` / `internal/csvimport` の gocritic/gosec 緩和

---

## 7. 却下済み・再提案しない（再開条件付き）

- カルテ同日重複に DB unique を採らない（2026-07-27）。再開条件 = 手動作成経路で実害が出た場合のみ
- auto-create に clock seam を導入しない（2026-07-27）。予約日基準が正
- Count→Delete の**一括** retrofit を本監査の実装スコープにしない（CODING_RULES。触る Delete では直す）
- medicalrecord を `handler/service/repository` サブパッケージへ層分割しない
- `map[string]any` の監査 metadata / テスト fixture を禁止しない
- wrapcheck を host の full `golangci-lint run ./...` でエージェントが回さない
- stutter 一括 rename、GoDoc 一括、同一 package map Update 一括 unexport
- `nested_summary` 3 package コピーの統合（import cycle / 意図的非統合）
- `util` / `common` / `misc` 新設

---

## 8. カバレッジ表（production `.go` 全 981）

各行は `OK` / `FINDING(IDs)` / `SKIP(理由)`。FINDING の ID は開いた所見のみ（DONE 済み ID を未修正として付けない）。

### `backend/cmd/api` (22)

| path | status |
|---|---|
| `backend/cmd/api/base_routes.go` | OK |
| `backend/cmd/api/batch_scheduler.go` | OK |
| `backend/cmd/api/composition_auth.go` | OK |
| `backend/cmd/api/composition_auth_routes.go` | FINDING(BE-RC-005) |
| `backend/cmd/api/composition_billing.go` | OK |
| `backend/cmd/api/composition_billing_repositories.go` | OK |
| `backend/cmd/api/composition_billing_services.go` | OK |
| `backend/cmd/api/composition_clinic.go` | OK |
| `backend/cmd/api/composition_credential_audit.go` | OK |
| `backend/cmd/api/composition_medicalrecord.go` | OK |
| `backend/cmd/api/composition_medicalrecord_repositories.go` | OK |
| `backend/cmd/api/composition_medicalrecord_services.go` | OK |
| `backend/cmd/api/composition_owner_pet.go` | OK |
| `backend/cmd/api/composition_reservation.go` | OK |
| `backend/cmd/api/composition_runtime.go` | OK |
| `backend/cmd/api/composition_staff.go` | OK |
| `backend/cmd/api/composition_staff_account.go` | OK |
| `backend/cmd/api/lstep_adapters.go` | OK |
| `backend/cmd/api/main.go` | OK |
| `backend/cmd/api/runtime_execution.go` | OK |
| `backend/cmd/api/server_runner.go` | OK |
| `backend/cmd/api/smtp_adapters.go` | OK |

### `backend/cmd/coverage-ratchet` (1)

| path | status |
|---|---|
| `backend/cmd/coverage-ratchet/main.go` | OK |

### `backend/cmd/csv-import-failure-rehearsal` (1)

| path | status |
|---|---|
| `backend/cmd/csv-import-failure-rehearsal/main.go` | OK |

### `backend/cmd/csv-import-stg-uat` (1)

| path | status |
|---|---|
| `backend/cmd/csv-import-stg-uat/main.go` | OK |

### `backend/cmd/csv-import` (1)

| path | status |
|---|---|
| `backend/cmd/csv-import/main.go` | OK |

### `backend/cmd/lab-device-agent` (1)

| path | status |
|---|---|
| `backend/cmd/lab-device-agent/main.go` | OK |

### `backend/cmd/lstep-migrate` (4)

| path | status |
|---|---|
| `backend/cmd/lstep-migrate/cli.go` | OK |
| `backend/cmd/lstep-migrate/main.go` | OK |
| `backend/cmd/lstep-migrate/migrator.go` | OK |
| `backend/cmd/lstep-migrate/reporter.go` | OK |

### `backend/cmd/migrate` (3)

| path | status |
|---|---|
| `backend/cmd/migrate/csvbundle.go` | OK |
| `backend/cmd/migrate/main.go` | OK |
| `backend/cmd/migrate/reconcile_keys.go` | OK |

### `backend/cmd/seed-export` (3)

| path | status |
|---|---|
| `backend/cmd/seed-export/dump.go` | OK |
| `backend/cmd/seed-export/main.go` | OK |
| `backend/cmd/seed-export/tables.go` | OK |

### `backend/cmd/staff-provision` (1)

| path | status |
|---|---|
| `backend/cmd/staff-provision/main.go` | OK |

### `backend/cmd/stg-uat-skeleton` (3)

| path | status |
|---|---|
| `backend/cmd/stg-uat-skeleton/ensure.go` | OK |
| `backend/cmd/stg-uat-skeleton/main.go` | OK |
| `backend/cmd/stg-uat-skeleton/verify.go` | OK |

### `backend/cmd/stg-uat-staff-attach` (4)

| path | status |
|---|---|
| `backend/cmd/stg-uat-staff-attach/attach.go` | OK |
| `backend/cmd/stg-uat-staff-attach/main.go` | OK |
| `backend/cmd/stg-uat-staff-attach/repository.go` | OK |
| `backend/cmd/stg-uat-staff-attach/validate.go` | OK |

### `backend/internal/apicontract` (1)

| path | status |
|---|---|
| `backend/internal/apicontract/doc.go` | OK |

### `backend/internal/apperrors` (1)

| path | status |
|---|---|
| `backend/internal/apperrors/errors.go` | FINDING(BE-RC-014) |

### `backend/internal/audit` (2)

| path | status |
|---|---|
| `backend/internal/audit/repository.go` | OK |
| `backend/internal/audit/service.go` | FINDING(BE-RC-005) |

### `backend/internal/auth` (34)

| path | status |
|---|---|
| `backend/internal/auth/account_repository.go` | OK |
| `backend/internal/auth/account_service.go` | FINDING(BE-RC-005) |
| `backend/internal/auth/auth_service.go` | FINDING(BE-RC-005,BE-RC-015) |
| `backend/internal/auth/credential_audit.go` | OK |
| `backend/internal/auth/current_access_service.go` | OK |
| `backend/internal/auth/current_access_staff_reader.go` | OK |
| `backend/internal/auth/doc.go` | OK |
| `backend/internal/auth/http_binding.go` | OK |
| `backend/internal/auth/http_password.go` | FINDING(BE-RC-005) |
| `backend/internal/auth/http_permission.go` | OK |
| `backend/internal/auth/http_response.go` | FINDING(BE-RC-005) |
| `backend/internal/auth/http_routes.go` | OK |
| `backend/internal/auth/http_session.go` | OK |
| `backend/internal/auth/http_session_cookies.go` | OK |
| `backend/internal/auth/http_session_login.go` | FINDING(BE-RC-005) |
| `backend/internal/auth/http_session_me.go` | FINDING(BE-RC-005) |
| `backend/internal/auth/http_session_refresh.go` | FINDING(BE-RC-005) |
| `backend/internal/auth/http_types.go` | OK |
| `backend/internal/auth/login_failure_audit.go` | OK |
| `backend/internal/auth/login_response_floor.go` | OK |
| `backend/internal/auth/password_reset_persist.go` | OK |
| `backend/internal/auth/password_reset_response_floor.go` | OK |
| `backend/internal/auth/password_reset_service.go` | FINDING(BE-RC-005) |
| `backend/internal/auth/password_reset_token_repository.go` | OK |
| `backend/internal/auth/password_timing.go` | OK |
| `backend/internal/auth/permission_group_create_request.go` | OK |
| `backend/internal/auth/permission_group_repository.go` | FINDING(BE-RC-009,BE-RC-017) |
| `backend/internal/auth/permission_group_service.go` | FINDING(BE-RC-005) |
| `backend/internal/auth/permission_group_service_mutate.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/auth/permission_group_service_rules.go` | FINDING(BE-RC-005) |
| `backend/internal/auth/refresh_family.go` | OK |
| `backend/internal/auth/token_blacklist_repository.go` | OK |
| `backend/internal/auth/token_blacklist_service.go` | FINDING(BE-RC-005) |
| `backend/internal/auth/token_service.go` | FINDING(BE-RC-005) |

### `backend/internal/authjwt` (1)

| path | status |
|---|---|
| `backend/internal/authjwt/claims.go` | OK |

### `backend/internal/billing` (93)

| path | status |
|---|---|
| `backend/internal/billing/accounting_complete.go` | OK |
| `backend/internal/billing/accounting_complete_digest.go` | OK |
| `backend/internal/billing/accounting_complete_tx.go` | OK |
| `backend/internal/billing/accounting_handler.go` | OK |
| `backend/internal/billing/accounting_report_handler.go` | OK |
| `backend/internal/billing/accounting_report_request.go` | OK |
| `backend/internal/billing/accounting_report_response.go` | OK |
| `backend/internal/billing/accounting_report_service.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/accounting_report_service_monthly_build.go` | OK |
| `backend/internal/billing/accounting_reports_dto.go` | OK |
| `backend/internal/billing/accounting_repository.go` | FINDING(BE-RC-009,BE-RC-017) |
| `backend/internal/billing/accounting_repository_actor.go` | OK |
| `backend/internal/billing/accounting_repository_ltv.go` | OK |
| `backend/internal/billing/accounting_repository_reports.go` | OK |
| `backend/internal/billing/accounting_repository_reports_allocation.go` | OK |
| `backend/internal/billing/accounting_repository_reports_close.go` | OK |
| `backend/internal/billing/accounting_repository_reports_close_queries.go` | OK |
| `backend/internal/billing/accounting_repository_reports_daily.go` | OK |
| `backend/internal/billing/accounting_repository_reports_monthly.go` | OK |
| `backend/internal/billing/accounting_repository_reports_monthly_scan.go` | OK |
| `backend/internal/billing/accounting_repository_reports_shared.go` | OK |
| `backend/internal/billing/accounting_repository_unpaid.go` | OK |
| `backend/internal/billing/accounting_request.go` | FINDING(BE-RC-024) |
| `backend/internal/billing/accounting_response.go` | OK |
| `backend/internal/billing/accounting_service.go` | FINDING(BE-RC-009) |
| `backend/internal/billing/accounting_service_builders.go` | OK |
| `backend/internal/billing/accounting_service_core.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/accounting_service_correction.go` | FINDING(BE-RC-005,BE-RC-032) |
| `backend/internal/billing/accounting_service_reports.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/accounting_service_update.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/allocation.go` | OK |
| `backend/internal/billing/billing_confirmation_handler.go` | OK |
| `backend/internal/billing/billing_confirmation_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/billing/billing_confirmation_request.go` | OK |
| `backend/internal/billing/billing_confirmation_response.go` | OK |
| `backend/internal/billing/billing_confirmation_service.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/billing_item_exam.go` | OK |
| `backend/internal/billing/billing_item_handler.go` | OK |
| `backend/internal/billing/billing_item_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/billing/billing_item_repository_trimming_refs.go` | OK |
| `backend/internal/billing/billing_item_repository_unbilled.go` | OK |
| `backend/internal/billing/billing_item_repository_vaccination.go` | OK |
| `backend/internal/billing/billing_item_repository_vaccination_lock.go` | OK |
| `backend/internal/billing/billing_item_request.go` | FINDING(BE-RC-024) |
| `backend/internal/billing/billing_item_service.go` | OK |
| `backend/internal/billing/billing_item_service_create.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/billing_item_service_update.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/billing_item_unbilled.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/billing_service.go` | OK |
| `backend/internal/billing/campaign_discount.go` | OK |
| `backend/internal/billing/campaign_handler.go` | OK |
| `backend/internal/billing/campaign_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/billing/campaign_request.go` | OK |
| `backend/internal/billing/campaign_response.go` | OK |
| `backend/internal/billing/campaign_service.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/cash_register_close_repository.go` | OK |
| `backend/internal/billing/cash_register_handler.go` | OK |
| `backend/internal/billing/cash_register_request.go` | FINDING(BE-RC-024) |
| `backend/internal/billing/cash_register_response.go` | OK |
| `backend/internal/billing/cash_register_service.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/doc.go` | OK |
| `backend/internal/billing/estimate_audit_diff.go` | OK |
| `backend/internal/billing/estimate_handler.go` | OK |
| `backend/internal/billing/estimate_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/billing/estimate_request.go` | FINDING(BE-RC-024) |
| `backend/internal/billing/estimate_response.go` | OK |
| `backend/internal/billing/estimate_service.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/estimate_service_successor.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/estimate_service_successor_tx.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/estimate_service_tx.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/insurance_handler.go` | OK |
| `backend/internal/billing/insurance_repository.go` | FINDING(BE-RC-017,BE-RC-032) |
| `backend/internal/billing/insurance_request.go` | FINDING(BE-RC-024) |
| `backend/internal/billing/insurance_response.go` | OK |
| `backend/internal/billing/insurance_service.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/list_query_helpers_test_free.go` | OK |
| `backend/internal/billing/nested_summary_response.go` | OK |
| `backend/internal/billing/payment_method_master_handler.go` | OK |
| `backend/internal/billing/payment_method_master_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/billing/payment_method_master_request.go` | FINDING(BE-RC-024) |
| `backend/internal/billing/payment_method_master_response.go` | OK |
| `backend/internal/billing/payment_method_master_service.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/payment_method_scope.go` | OK |
| `backend/internal/billing/refund_handler.go` | OK |
| `backend/internal/billing/refund_repository.go` | OK |
| `backend/internal/billing/refund_request.go` | OK |
| `backend/internal/billing/refund_service.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/refund_service_tx.go` | FINDING(BE-RC-005) |
| `backend/internal/billing/routes.go` | OK |
| `backend/internal/billing/service_deps.go` | OK |
| `backend/internal/billing/unpaid_amount.go` | OK |
| `backend/internal/billing/validators_billing_item.go` | OK |
| `backend/internal/billing/validators_insurance.go` | OK |

### `backend/internal/clinic` (25)

| path | status |
|---|---|
| `backend/internal/clinic/clinic_handler.go` | OK |
| `backend/internal/clinic/clinic_holiday_handler.go` | FINDING(BE-RC-024) |
| `backend/internal/clinic/clinic_holiday_repository.go` | OK |
| `backend/internal/clinic/clinic_holiday_request.go` | OK |
| `backend/internal/clinic/clinic_holiday_response.go` | OK |
| `backend/internal/clinic/clinic_holiday_service.go` | FINDING(BE-RC-005) |
| `backend/internal/clinic/clinic_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/clinic/clinic_request.go` | FINDING(BE-RC-023,BE-RC-024) |
| `backend/internal/clinic/clinic_response.go` | OK |
| `backend/internal/clinic/clinic_service.go` | FINDING(BE-RC-004,BE-RC-005,BE-RC-015,BE-RC-021) |
| `backend/internal/clinic/clinic_service_update.go` | OK |
| `backend/internal/clinic/clinic_settings_repository.go` | OK |
| `backend/internal/clinic/closing_settings_handler.go` | OK |
| `backend/internal/clinic/closing_settings_request.go` | FINDING(BE-RC-024) |
| `backend/internal/clinic/closing_settings_response.go` | OK |
| `backend/internal/clinic/closing_settings_service.go` | FINDING(BE-RC-005) |
| `backend/internal/clinic/closing_special_period_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/clinic/company_handler.go` | OK |
| `backend/internal/clinic/company_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/clinic/company_request.go` | FINDING(BE-RC-024) |
| `backend/internal/clinic/company_response.go` | OK |
| `backend/internal/clinic/company_service.go` | FINDING(BE-RC-005) |
| `backend/internal/clinic/handler.go` | OK |
| `backend/internal/clinic/ports.go` | FINDING(BE-RC-009) |
| `backend/internal/clinic/repositories.go` | OK |

### `backend/internal/config` (2)

| path | status |
|---|---|
| `backend/internal/config/config.go` | FINDING(BE-RC-021) |
| `backend/internal/config/timezone.go` | OK |

### `backend/internal/csvimport` (10)

| path | status |
|---|---|
| `backend/internal/csvimport/cutover_contract.go` | OK |
| `backend/internal/csvimport/cutover_contract_validate.go` | OK |
| `backend/internal/csvimport/cutover_import.go` | OK |
| `backend/internal/csvimport/cutover_import_copy.go` | OK |
| `backend/internal/csvimport/cutover_import_validate.go` | OK |
| `backend/internal/csvimport/cutover_payment_contract.go` | OK |
| `backend/internal/csvimport/cutover_payment_graph.go` | OK |
| `backend/internal/csvimport/cutover_payment_target.go` | OK |
| `backend/internal/csvimport/failure_rehearsal.go` | OK |
| `backend/internal/csvimport/import.go` | OK |

### `backend/internal/dbconn` (2)

| path | status |
|---|---|
| `backend/internal/dbconn/dbconn.go` | OK |
| `backend/internal/dbconn/gorm.go` | OK |

### `backend/internal/httpapi` (13)

| path | status |
|---|---|
| `backend/internal/httpapi/bind_errors.go` | OK |
| `backend/internal/httpapi/clinic_permission.go` | OK |
| `backend/internal/httpapi/context.go` | OK |
| `backend/internal/httpapi/date.go` | OK |
| `backend/internal/httpapi/discount_permission.go` | OK |
| `backend/internal/httpapi/enum.go` | OK |
| `backend/internal/httpapi/pagination_response.go` | OK |
| `backend/internal/httpapi/query_helpers.go` | OK |
| `backend/internal/httpapi/response.go` | OK |
| `backend/internal/httpapi/response_pg.go` | OK |
| `backend/internal/httpapi/slice.go` | OK |
| `backend/internal/httpapi/string_query.go` | OK |
| `backend/internal/httpapi/time.go` | OK |

### `backend/internal/identitylink` (12)

| path | status |
|---|---|
| `backend/internal/identitylink/handler.go` | FINDING(BE-RC-024) |
| `backend/internal/identitylink/repository.go` | FINDING(BE-RC-009) |
| `backend/internal/identitylink/request.go` | OK |
| `backend/internal/identitylink/response.go` | OK |
| `backend/internal/identitylink/routes.go` | OK |
| `backend/internal/identitylink/service.go` | OK |
| `backend/internal/identitylink/service_helpers.go` | OK |
| `backend/internal/identitylink/service_history.go` | OK |
| `backend/internal/identitylink/service_owner.go` | OK |
| `backend/internal/identitylink/service_pet.go` | OK |
| `backend/internal/identitylink/service_pet_create.go` | OK |
| `backend/internal/identitylink/types.go` | OK |

### `backend/internal/infra` (19)

| path | status |
|---|---|
| `backend/internal/infra/crypto/aes_gcm.go` | OK |
| `backend/internal/infra/file_storage.go` | OK |
| `backend/internal/infra/httpx/retry.go` | OK |
| `backend/internal/infra/line/client.go` | OK |
| `backend/internal/infra/line/errors.go` | OK |
| `backend/internal/infra/line/push.go` | OK |
| `backend/internal/infra/local_file_storage.go` | OK |
| `backend/internal/infra/local_uploader.go` | OK |
| `backend/internal/infra/lstep/breed_codes.go` | OK |
| `backend/internal/infra/lstep/client.go` | OK |
| `backend/internal/infra/lstep/dial.go` | OK |
| `backend/internal/infra/lstep/errors.go` | OK |
| `backend/internal/infra/lstep/tag.go` | OK |
| `backend/internal/infra/lstep/user.go` | OK |
| `backend/internal/infra/s3_endpoint.go` | OK |
| `backend/internal/infra/s3_file_storage.go` | OK |
| `backend/internal/infra/s3_uploader.go` | OK |
| `backend/internal/infra/smtp/sender.go` | OK |
| `backend/internal/infra/uploader.go` | OK |

### `backend/internal/inventory` (12)

| path | status |
|---|---|
| `backend/internal/inventory/handler.go` | OK |
| `backend/internal/inventory/inventory_handler.go` | OK |
| `backend/internal/inventory/inventory_request.go` | FINDING(BE-RC-024) |
| `backend/internal/inventory/inventory_response.go` | OK |
| `backend/internal/inventory/inventory_service.go` | FINDING(BE-RC-005) |
| `backend/internal/inventory/merchandise_item_handler.go` | OK |
| `backend/internal/inventory/merchandise_item_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/inventory/merchandise_item_request.go` | FINDING(BE-RC-024) |
| `backend/internal/inventory/merchandise_item_response.go` | OK |
| `backend/internal/inventory/merchandise_item_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/inventory/repository.go` | FINDING(BE-RC-017) |
| `backend/internal/inventory/routes.go` | OK |

### `backend/internal/labdeviceagent` (5)

| path | status |
|---|---|
| `backend/internal/labdeviceagent/agent.go` | OK |
| `backend/internal/labdeviceagent/http.go` | OK |
| `backend/internal/labdeviceagent/queue.go` | OK |
| `backend/internal/labdeviceagent/serial_darwin.go` | OK |
| `backend/internal/labdeviceagent/serial_other.go` | OK |

### `backend/internal/lintscan` (1)

| path | status |
|---|---|
| `backend/internal/lintscan/lintscan.go` | OK |

### `backend/internal/logger` (1)

| path | status |
|---|---|
| `backend/internal/logger/logger.go` | OK |

### `backend/internal/lstep` (133)

| path | status |
|---|---|
| `backend/internal/lstep/aggregation_handler.go` | OK |
| `backend/internal/lstep/aggregation_request.go` | OK |
| `backend/internal/lstep/aggregation_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/checkup_sync_handler.go` | OK |
| `backend/internal/lstep/checkup_sync_repository.go` | OK |
| `backend/internal/lstep/checkup_sync_repository_preview_sql.go` | OK |
| `backend/internal/lstep/checkup_sync_request.go` | OK |
| `backend/internal/lstep/checkup_sync_response.go` | OK |
| `backend/internal/lstep/checkup_sync_service.go` | OK |
| `backend/internal/lstep/checkup_sync_service_create.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/checkup_sync_service_metadata.go` | OK |
| `backend/internal/lstep/checkup_sync_service_preview.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/composition.go` | OK |
| `backend/internal/lstep/composition_lifecycle_ports.go` | OK |
| `backend/internal/lstep/composition_repositories.go` | OK |
| `backend/internal/lstep/composition_services.go` | OK |
| `backend/internal/lstep/doc.go` | OK |
| `backend/internal/lstep/line_credentials.go` | OK |
| `backend/internal/lstep/line_customer_handler.go` | OK |
| `backend/internal/lstep/line_customer_repository.go` | OK |
| `backend/internal/lstep/line_customer_request.go` | OK |
| `backend/internal/lstep/line_customer_response.go` | OK |
| `backend/internal/lstep/line_customer_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/line_link_handler.go` | OK |
| `backend/internal/lstep/line_link_request.go` | OK |
| `backend/internal/lstep/line_link_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/line_link_token_repository.go` | OK |
| `backend/internal/lstep/line_messaging_service.go` | OK |
| `backend/internal/lstep/line_send_handler.go` | OK |
| `backend/internal/lstep/line_send_log_repository.go` | OK |
| `backend/internal/lstep/line_send_request.go` | OK |
| `backend/internal/lstep/line_send_response.go` | OK |
| `backend/internal/lstep/line_send_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_analytics_handler.go` | OK |
| `backend/internal/lstep/lstep_analytics_request.go` | OK |
| `backend/internal/lstep/lstep_analytics_response.go` | OK |
| `backend/internal/lstep/lstep_analytics_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_batch_delivery.go` | FINDING(BE-RC-005,BE-RC-028) |
| `backend/internal/lstep/lstep_batch_dormant.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_batch_noshow.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_batch_segmentation.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_batch_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_csv_helpers.go` | OK |
| `backend/internal/lstep/lstep_csv_import_concurrency.go` | OK |
| `backend/internal/lstep/lstep_csv_import_handler.go` | OK |
| `backend/internal/lstep/lstep_csv_import_owner_lookup.go` | OK |
| `backend/internal/lstep/lstep_csv_import_prepare.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_csv_import_processing.go` | OK |
| `backend/internal/lstep/lstep_csv_import_repository.go` | OK |
| `backend/internal/lstep/lstep_csv_import_request.go` | OK |
| `backend/internal/lstep/lstep_csv_import_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_csv_stream.go` | OK |
| `backend/internal/lstep/lstep_delivery_deps.go` | OK |
| `backend/internal/lstep/lstep_delivery_monitor_handler.go` | OK |
| `backend/internal/lstep/lstep_delivery_monitor_request.go` | OK |
| `backend/internal/lstep/lstep_delivery_monitor_response.go` | OK |
| `backend/internal/lstep/lstep_delivery_monitor_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_delivery_trigger_batch.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_delivery_trigger_client.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_delivery_trigger_log_repository.go` | OK |
| `backend/internal/lstep/lstep_delivery_trigger_methods.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_delivery_trigger_service.go` | OK |
| `backend/internal/lstep/lstep_delivery_trigger_state.go` | OK |
| `backend/internal/lstep/lstep_delivery_trigger_suppression.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_friend_attribute_snapshot_repository.go` | OK |
| `backend/internal/lstep/lstep_health_codes.go` | OK |
| `backend/internal/lstep/lstep_health_tag_sync.go` | OK |
| `backend/internal/lstep/lstep_health_tag_sync_batch.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_health_tag_sync_checkup.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_health_tag_sync_food.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_health_tag_sync_prevention.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_health_tag_sync_vaccine.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_lifecycle_deps.go` | OK |
| `backend/internal/lstep/lstep_lifecycle_handler.go` | OK |
| `backend/internal/lstep/lstep_lifecycle_request.go` | OK |
| `backend/internal/lstep/lstep_lifecycle_service.go` | FINDING(BE-RC-005,BE-RC-028) |
| `backend/internal/lstep/lstep_settings_connection.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_settings_credentials.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_settings_handler.go` | FINDING(BE-RC-015) |
| `backend/internal/lstep/lstep_settings_repository.go` | OK |
| `backend/internal/lstep/lstep_settings_request.go` | OK |
| `backend/internal/lstep/lstep_settings_response.go` | OK |
| `backend/internal/lstep/lstep_settings_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_settings_thresholds.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_settings_update.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_sync_error_counter_repository.go` | OK |
| `backend/internal/lstep/lstep_sync_settings_repository.go` | OK |
| `backend/internal/lstep/lstep_tag_cache_repository.go` | OK |
| `backend/internal/lstep/lstep_tag_code_mapping_handler.go` | OK |
| `backend/internal/lstep/lstep_tag_code_mapping_repository.go` | OK |
| `backend/internal/lstep/lstep_tag_code_mapping_request.go` | OK |
| `backend/internal/lstep/lstep_tag_code_mapping_response.go` | OK |
| `backend/internal/lstep/lstep_tag_code_mapping_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_config_handler.go` | OK |
| `backend/internal/lstep/lstep_tag_config_repository.go` | OK |
| `backend/internal/lstep/lstep_tag_config_request.go` | OK |
| `backend/internal/lstep/lstep_tag_config_response.go` | OK |
| `backend/internal/lstep/lstep_tag_config_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_handler.go` | OK |
| `backend/internal/lstep/lstep_tag_request.go` | OK |
| `backend/internal/lstep/lstep_tag_response.go` | OK |
| `backend/internal/lstep/lstep_tag_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_summary_handler.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_summary_request.go` | OK |
| `backend/internal/lstep/lstep_tag_summary_response.go` | OK |
| `backend/internal/lstep/lstep_tag_summary_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_api.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_care.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_care_checkup.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_care_chronic.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_care_resync.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_pet.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_pet_basic.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_pet_exclusion.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_service.go` | FINDING(BE-RC-028) |
| `backend/internal/lstep/lstep_tag_sync_shared.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_vaccine.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_visit.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_visit_cpm.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_visit_dormant.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_visit_ltv.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_tag_sync_visit_next.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/lstep_trigger_priority_handler.go` | OK |
| `backend/internal/lstep/lstep_trigger_priority_repository.go` | OK |
| `backend/internal/lstep/lstep_trigger_priority_request.go` | OK |
| `backend/internal/lstep/lstep_trigger_priority_service.go` | FINDING(BE-RC-005) |
| `backend/internal/lstep/routes.go` | OK |
| `backend/internal/lstep/service_deps.go` | OK |
| `backend/internal/lstep/shared_file_handler.go` | OK |
| `backend/internal/lstep/shared_file_repository.go` | OK |
| `backend/internal/lstep/shared_file_request.go` | OK |
| `backend/internal/lstep/shared_file_response.go` | OK |
| `backend/internal/lstep/shared_file_service.go` | FINDING(BE-RC-005) |

### `backend/internal/manualarticle` (6)

| path | status |
|---|---|
| `backend/internal/manualarticle/audit.go` | OK |
| `backend/internal/manualarticle/handler.go` | FINDING(BE-RC-005) |
| `backend/internal/manualarticle/repository.go` | OK |
| `backend/internal/manualarticle/request.go` | OK |
| `backend/internal/manualarticle/response.go` | OK |
| `backend/internal/manualarticle/service.go` | FINDING(BE-RC-005) |

### `backend/internal/medicalrecord` (239)

| path | status |
|---|---|
| `backend/internal/medicalrecord/audit_diff.go` | OK |
| `backend/internal/medicalrecord/cage_handler.go` | OK |
| `backend/internal/medicalrecord/cage_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/cage_request.go` | OK |
| `backend/internal/medicalrecord/cage_response.go` | OK |
| `backend/internal/medicalrecord/cage_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/care_plan_item_handler.go` | OK |
| `backend/internal/medicalrecord/care_plan_item_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/care_plan_item_request.go` | OK |
| `backend/internal/medicalrecord/care_plan_item_response.go` | OK |
| `backend/internal/medicalrecord/care_plan_item_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/checkup_field_handler.go` | OK |
| `backend/internal/medicalrecord/checkup_field_repository.go` | OK |
| `backend/internal/medicalrecord/checkup_field_request.go` | OK |
| `backend/internal/medicalrecord/checkup_field_response.go` | OK |
| `backend/internal/medicalrecord/checkup_field_result_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/checkup_handler.go` | OK |
| `backend/internal/medicalrecord/checkup_package_import_apply.go` | OK |
| `backend/internal/medicalrecord/checkup_package_import_apply_tx.go` | OK |
| `backend/internal/medicalrecord/checkup_package_import_handler.go` | OK |
| `backend/internal/medicalrecord/checkup_package_import_service.go` | OK |
| `backend/internal/medicalrecord/checkup_package_manifest.go` | OK |
| `backend/internal/medicalrecord/checkup_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/checkup_request.go` | OK |
| `backend/internal/medicalrecord/checkup_response.go` | OK |
| `backend/internal/medicalrecord/checkup_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/checkup_type_handler.go` | OK |
| `backend/internal/medicalrecord/checkup_type_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/checkup_type_request.go` | OK |
| `backend/internal/medicalrecord/checkup_type_response.go` | OK |
| `backend/internal/medicalrecord/checkup_type_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/chief_complaint_handler.go` | OK |
| `backend/internal/medicalrecord/chief_complaint_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/chief_complaint_request.go` | OK |
| `backend/internal/medicalrecord/chief_complaint_response.go` | OK |
| `backend/internal/medicalrecord/chief_complaint_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/clinical_plan_handler.go` | OK |
| `backend/internal/medicalrecord/clinical_plan_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/clinical_plan_request.go` | OK |
| `backend/internal/medicalrecord/clinical_plan_response.go` | OK |
| `backend/internal/medicalrecord/clinical_plan_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/clinical_relation_validation.go` | OK |
| `backend/internal/medicalrecord/consultation_handler.go` | OK |
| `backend/internal/medicalrecord/consultation_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/consultation_request.go` | OK |
| `backend/internal/medicalrecord/consultation_response.go` | OK |
| `backend/internal/medicalrecord/consultation_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/daily_record_handler.go` | OK |
| `backend/internal/medicalrecord/daily_record_repository.go` | OK |
| `backend/internal/medicalrecord/daily_record_request.go` | OK |
| `backend/internal/medicalrecord/daily_record_response.go` | OK |
| `backend/internal/medicalrecord/daily_record_service.go` | FINDING(BE-RC-005,BE-RC-019) |
| `backend/internal/medicalrecord/diagnosis_handler.go` | OK |
| `backend/internal/medicalrecord/diagnosis_name_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/diagnosis_request.go` | OK |
| `backend/internal/medicalrecord/diagnosis_response.go` | OK |
| `backend/internal/medicalrecord/diagnosis_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/diagnosis_type_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/discount_permission.go` | OK |
| `backend/internal/medicalrecord/dose_calc.go` | OK |
| `backend/internal/medicalrecord/dose_revalidation.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/dose_validators.go` | OK |
| `backend/internal/medicalrecord/exam_reference_range_repository.go` | OK |
| `backend/internal/medicalrecord/exam_result_assessment.go` | OK |
| `backend/internal/medicalrecord/exam_type_field.go` | OK |
| `backend/internal/medicalrecord/exam_type_handler.go` | OK |
| `backend/internal/medicalrecord/exam_type_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/exam_type_request.go` | OK |
| `backend/internal/medicalrecord/exam_type_response.go` | OK |
| `backend/internal/medicalrecord/exam_type_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/examination_audit.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/examination_handler.go` | OK |
| `backend/internal/medicalrecord/examination_items.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/examination_lock.go` | OK |
| `backend/internal/medicalrecord/examination_pet_safety.go` | OK |
| `backend/internal/medicalrecord/examination_print_snapshot.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/examination_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/examination_request.go` | OK |
| `backend/internal/medicalrecord/examination_response.go` | OK |
| `backend/internal/medicalrecord/examination_revision_repository.go` | OK |
| `backend/internal/medicalrecord/examination_revision_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/examination_revision_workflow_repository.go` | OK |
| `backend/internal/medicalrecord/examination_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/examination_service_create.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/examination_service_update.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/handler_deps.go` | OK |
| `backend/internal/medicalrecord/hospitalization_discharge.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/hospitalization_discharge_tx.go` | OK |
| `backend/internal/medicalrecord/hospitalization_handler.go` | OK |
| `backend/internal/medicalrecord/hospitalization_plan_handler.go` | OK |
| `backend/internal/medicalrecord/hospitalization_plan_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/hospitalization_plan_request.go` | OK |
| `backend/internal/medicalrecord/hospitalization_plan_response.go` | OK |
| `backend/internal/medicalrecord/hospitalization_plan_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/hospitalization_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/hospitalization_request.go` | OK |
| `backend/internal/medicalrecord/hospitalization_response.go` | OK |
| `backend/internal/medicalrecord/hospitalization_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/hospitalization_service_create.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/inquiry_handler.go` | OK |
| `backend/internal/medicalrecord/inquiry_repository.go` | OK |
| `backend/internal/medicalrecord/inquiry_request.go` | OK |
| `backend/internal/medicalrecord/inquiry_response.go` | OK |
| `backend/internal/medicalrecord/inquiry_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/inquiry_template_handler.go` | OK |
| `backend/internal/medicalrecord/inquiry_template_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/inquiry_template_request.go` | OK |
| `backend/internal/medicalrecord/inquiry_template_response.go` | OK |
| `backend/internal/medicalrecord/inquiry_template_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/lab_audit_logger.go` | OK |
| `backend/internal/medicalrecord/lab_device_agent_consumer_handler.go` | OK |
| `backend/internal/medicalrecord/lab_device_decode.go` | OK |
| `backend/internal/medicalrecord/lab_device_exam_persist.go` | FINDING(BE-RC-031) |
| `backend/internal/medicalrecord/lab_device_frame.go` | OK |
| `backend/internal/medicalrecord/lab_device_fuji.go` | OK |
| `backend/internal/medicalrecord/lab_device_idexx.go` | OK |
| `backend/internal/medicalrecord/lab_device_idexx_pims.go` | OK |
| `backend/internal/medicalrecord/lab_device_item_catalog.go` | OK |
| `backend/internal/medicalrecord/lab_device_item_master_handler.go` | OK |
| `backend/internal/medicalrecord/lab_device_item_master_repository.go` | OK |
| `backend/internal/medicalrecord/lab_device_item_master_request.go` | OK |
| `backend/internal/medicalrecord/lab_device_item_master_service.go` | OK |
| `backend/internal/medicalrecord/lab_device_receive.go` | OK |
| `backend/internal/medicalrecord/lab_device_receive_repository.go` | OK |
| `backend/internal/medicalrecord/lab_device_receive_service.go` | FINDING(BE-RC-019,BE-RC-031) |
| `backend/internal/medicalrecord/lab_device_today_visit.go` | OK |
| `backend/internal/medicalrecord/lab_device_urine.go` | OK |
| `backend/internal/medicalrecord/lab_import_examination_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/lab_import_examination_write.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/lab_import_handler.go` | OK |
| `backend/internal/medicalrecord/lab_import_repository.go` | OK |
| `backend/internal/medicalrecord/lab_import_request.go` | OK |
| `backend/internal/medicalrecord/lab_import_response.go` | OK |
| `backend/internal/medicalrecord/lab_import_revert_service.go` | OK |
| `backend/internal/medicalrecord/lab_import_revert_tx.go` | OK |
| `backend/internal/medicalrecord/lab_import_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/lab_import_usage_tracker.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/lab_report_handler.go` | OK |
| `backend/internal/medicalrecord/lab_report_query_service.go` | OK |
| `backend/internal/medicalrecord/lab_report_response.go` | OK |
| `backend/internal/medicalrecord/lab_result_import_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/master_validators.go` | OK |
| `backend/internal/medicalrecord/medical_record_addendum_handler.go` | OK |
| `backend/internal/medicalrecord/medical_record_addendum_repository.go` | OK |
| `backend/internal/medicalrecord/medical_record_addendum_request.go` | OK |
| `backend/internal/medicalrecord/medical_record_addendum_response.go` | OK |
| `backend/internal/medicalrecord/medical_record_addendum_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/medical_record_appointment_context.go` | OK |
| `backend/internal/medicalrecord/medical_record_auto_create.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/medical_record_builders.go` | OK |
| `backend/internal/medicalrecord/medical_record_crud.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/medical_record_crud_update.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/medical_record_delete_conflict.go` | OK |
| `backend/internal/medicalrecord/medical_record_handler.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_handler.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_repository.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_request.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_response.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/medical_record_image_upload_quota.go` | OK |
| `backend/internal/medicalrecord/medical_record_lock.go` | OK |
| `backend/internal/medicalrecord/medical_record_lstep_sync.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/medical_record_owner_visit_repository.go` | OK |
| `backend/internal/medicalrecord/medical_record_repository.go` | FINDING(BE-RC-009,BE-RC-017,BE-RC-019) |
| `backend/internal/medicalrecord/medical_record_repository_list.go` | OK |
| `backend/internal/medicalrecord/medical_record_repository_list_search.go` | OK |
| `backend/internal/medicalrecord/medical_record_request.go` | OK |
| `backend/internal/medicalrecord/medical_record_response.go` | OK |
| `backend/internal/medicalrecord/medical_record_service.go` | OK |
| `backend/internal/medicalrecord/medical_record_subrecords.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/medicine_dose_param_handler.go` | OK |
| `backend/internal/medicalrecord/medicine_dose_param_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/medicine_dose_param_request.go` | OK |
| `backend/internal/medicalrecord/medicine_dose_param_response.go` | OK |
| `backend/internal/medicalrecord/medicine_dose_param_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/medicine_handler.go` | OK |
| `backend/internal/medicalrecord/medicine_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/medicine_request.go` | OK |
| `backend/internal/medicalrecord/medicine_response.go` | OK |
| `backend/internal/medicalrecord/medicine_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/medicine_service_create.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/medicine_service_delete.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/nested_summary_response.go` | OK |
| `backend/internal/medicalrecord/pagination.go` | FINDING(BE-RC-030) |
| `backend/internal/medicalrecord/prescription_handler.go` | OK |
| `backend/internal/medicalrecord/prescription_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/prescription_request.go` | OK |
| `backend/internal/medicalrecord/prescription_response.go` | OK |
| `backend/internal/medicalrecord/prescription_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/procedure_handler.go` | OK |
| `backend/internal/medicalrecord/procedure_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/procedure_request.go` | OK |
| `backend/internal/medicalrecord/procedure_response.go` | OK |
| `backend/internal/medicalrecord/procedure_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/query_filter_helpers.go` | OK |
| `backend/internal/medicalrecord/replace_audit_tail.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/routes.go` | OK |
| `backend/internal/medicalrecord/routes_hospitalization.go` | OK |
| `backend/internal/medicalrecord/routes_lab.go` | OK |
| `backend/internal/medicalrecord/routes_masters.go` | OK |
| `backend/internal/medicalrecord/routes_records.go` | OK |
| `backend/internal/medicalrecord/service_deps.go` | FINDING(BE-RC-030) |
| `backend/internal/medicalrecord/to_service_input_error.go` | OK |
| `backend/internal/medicalrecord/treatment_dose_save.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/treatment_fields.go` | OK |
| `backend/internal/medicalrecord/treatment_handler.go` | OK |
| `backend/internal/medicalrecord/treatment_master_fk.go` | OK |
| `backend/internal/medicalrecord/treatment_plan_handler.go` | OK |
| `backend/internal/medicalrecord/treatment_plan_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/treatment_plan_request.go` | OK |
| `backend/internal/medicalrecord/treatment_plan_response.go` | OK |
| `backend/internal/medicalrecord/treatment_plan_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/treatment_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/treatment_request.go` | OK |
| `backend/internal/medicalrecord/treatment_response.go` | OK |
| `backend/internal/medicalrecord/treatment_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/treatment_service_tx.go` | OK |
| `backend/internal/medicalrecord/vaccination_handler.go` | OK |
| `backend/internal/medicalrecord/vaccination_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/vaccination_request.go` | OK |
| `backend/internal/medicalrecord/vaccination_response.go` | OK |
| `backend/internal/medicalrecord/vaccination_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/vaccine_handler.go` | OK |
| `backend/internal/medicalrecord/vaccine_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/vaccine_request.go` | OK |
| `backend/internal/medicalrecord/vaccine_response.go` | OK |
| `backend/internal/medicalrecord/vaccine_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/medicalrecord/validators.go` | OK |
| `backend/internal/medicalrecord/validators_accounting.go` | OK |
| `backend/internal/medicalrecord/validators_master.go` | OK |
| `backend/internal/medicalrecord/vital_audit.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/vital_handler.go` | OK |
| `backend/internal/medicalrecord/vital_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/medicalrecord/vital_request.go` | OK |
| `backend/internal/medicalrecord/vital_response.go` | OK |
| `backend/internal/medicalrecord/vital_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/vital_service_create.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/vital_service_update.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/vital_validation.go` | OK |

### `backend/internal/middleware` (9)

| path | status |
|---|---|
| `backend/internal/middleware/auth.go` | FINDING(BE-RC-005) |
| `backend/internal/middleware/cors.go` | OK |
| `backend/internal/middleware/csrf.go` | OK |
| `backend/internal/middleware/liff_auth.go` | FINDING(BE-RC-005) |
| `backend/internal/middleware/logging.go` | OK |
| `backend/internal/middleware/rate_limit.go` | OK |
| `backend/internal/middleware/response.go` | OK |
| `backend/internal/middleware/sanitize_null_bytes.go` | OK |
| `backend/internal/middleware/security_headers.go` | OK |

### `backend/internal/model` (94)

| path | status |
|---|---|
| `backend/internal/model/account.go` | OK |
| `backend/internal/model/accounting.go` | OK |
| `backend/internal/model/animal_species.go` | OK |
| `backend/internal/model/audit_log.go` | OK |
| `backend/internal/model/billing_confirmation.go` | OK |
| `backend/internal/model/billing_refund.go` | OK |
| `backend/internal/model/cage.go` | OK |
| `backend/internal/model/campaign.go` | OK |
| `backend/internal/model/cash_register_close.go` | OK |
| `backend/internal/model/cash_register_close_adjustment.go` | OK |
| `backend/internal/model/checkup_field.go` | OK |
| `backend/internal/model/checkup_package_import.go` | OK |
| `backend/internal/model/checkup_record.go` | OK |
| `backend/internal/model/checkup_type.go` | OK |
| `backend/internal/model/chief_complaint_type.go` | OK |
| `backend/internal/model/clinic.go` | OK |
| `backend/internal/model/clinic_holiday.go` | OK |
| `backend/internal/model/clinic_integration.go` | OK |
| `backend/internal/model/clinic_settings.go` | OK |
| `backend/internal/model/clinical_plan.go` | OK |
| `backend/internal/model/closing_special_period.go` | OK |
| `backend/internal/model/company.go` | OK |
| `backend/internal/model/consultation.go` | OK |
| `backend/internal/model/cpm_v1_thresholds.go` | OK |
| `backend/internal/model/cpm_v2_thresholds.go` | OK |
| `backend/internal/model/diagnosis.go` | OK |
| `backend/internal/model/dormant_thresholds.go` | OK |
| `backend/internal/model/estimate.go` | OK |
| `backend/internal/model/exam_reference_range.go` | OK |
| `backend/internal/model/examination_record.go` | OK |
| `backend/internal/model/examination_revision.go` | OK |
| `backend/internal/model/examination_type.go` | OK |
| `backend/internal/model/health_prevention_thresholds.go` | OK |
| `backend/internal/model/hospitalization.go` | OK |
| `backend/internal/model/hospitalization_plan.go` | OK |
| `backend/internal/model/identity_link.go` | OK |
| `backend/internal/model/inquiry.go` | OK |
| `backend/internal/model/inquiry_template.go` | OK |
| `backend/internal/model/insurance.go` | OK |
| `backend/internal/model/inventory.go` | OK |
| `backend/internal/model/lab_device.go` | OK |
| `backend/internal/model/lab_device_item_master.go` | OK |
| `backend/internal/model/lab_device_receive.go` | OK |
| `backend/internal/model/lab_import.go` | OK |
| `backend/internal/model/lab_report.go` | OK |
| `backend/internal/model/line_customer.go` | OK |
| `backend/internal/model/line_link_token.go` | FINDING(BE-RC-027) |
| `backend/internal/model/line_reservation_setting.go` | OK |
| `backend/internal/model/line_send_log.go` | OK |
| `backend/internal/model/lstep_auto_managed_prefix.go` | OK |
| `backend/internal/model/lstep_condition_tag_mapping.go` | OK |
| `backend/internal/model/lstep_csv_import.go` | OK |
| `backend/internal/model/lstep_delivery_trigger_log.go` | OK |
| `backend/internal/model/lstep_friend_attribute_snapshot.go` | OK |
| `backend/internal/model/lstep_send_purpose_tag_prefix.go` | OK |
| `backend/internal/model/lstep_settings.go` | OK |
| `backend/internal/model/lstep_sync_error_counter.go` | OK |
| `backend/internal/model/lstep_tag_cache.go` | OK |
| `backend/internal/model/lstep_tag_code_mapping.go` | OK |
| `backend/internal/model/lstep_trigger_priority.go` | OK |
| `backend/internal/model/manual_article.go` | OK |
| `backend/internal/model/medical_record.go` | OK |
| `backend/internal/model/medical_record_addendum.go` | OK |
| `backend/internal/model/medical_record_image.go` | OK |
| `backend/internal/model/medicine.go` | OK |
| `backend/internal/model/medicine_dose_param.go` | OK |
| `backend/internal/model/merchandise_item.go` | OK |
| `backend/internal/model/occupation.go` | OK |
| `backend/internal/model/owner.go` | OK |
| `backend/internal/model/password_reset_token.go` | OK |
| `backend/internal/model/payment_method_master.go` | OK |
| `backend/internal/model/permission.go` | OK |
| `backend/internal/model/permission_group.go` | OK |
| `backend/internal/model/pet.go` | OK |
| `backend/internal/model/pet_chronic_condition.go` | OK |
| `backend/internal/model/pet_owner.go` | OK |
| `backend/internal/model/prescription.go` | OK |
| `backend/internal/model/procedure.go` | OK |
| `backend/internal/model/reservation.go` | OK |
| `backend/internal/model/reservation_type.go` | OK |
| `backend/internal/model/reservation_type_group.go` | OK |
| `backend/internal/model/shared_file.go` | OK |
| `backend/internal/model/shift_entry_break.go` | OK |
| `backend/internal/model/staff.go` | OK |
| `backend/internal/model/staff_reservation_capability.go` | OK |
| `backend/internal/model/staff_reservation_exclusion.go` | OK |
| `backend/internal/model/token_blacklist.go` | OK |
| `backend/internal/model/treatment.go` | OK |
| `backend/internal/model/trimming.go` | OK |
| `backend/internal/model/trimming_course_type.go` | OK |
| `backend/internal/model/trimming_master.go` | OK |
| `backend/internal/model/vaccination_record.go` | OK |
| `backend/internal/model/vaccine.go` | OK |
| `backend/internal/model/vital.go` | OK |

### `backend/internal/owner` (18)

| path | status |
|---|---|
| `backend/internal/owner/http_date.go` | OK |
| `backend/internal/owner/http_handler.go` | OK |
| `backend/internal/owner/http_owner.go` | OK |
| `backend/internal/owner/http_request.go` | FINDING(BE-RC-024) |
| `backend/internal/owner/http_response.go` | OK |
| `backend/internal/owner/http_routes.go` | OK |
| `backend/internal/owner/ltv_repository.go` | OK |
| `backend/internal/owner/ltv_repository_query.go` | OK |
| `backend/internal/owner/mapper.go` | OK |
| `backend/internal/owner/pet_registration.go` | OK |
| `backend/internal/owner/repository.go` | FINDING(BE-RC-009) |
| `backend/internal/owner/service.go` | FINDING(BE-RC-015) |
| `backend/internal/owner/service_builders.go` | OK |
| `backend/internal/owner/service_core.go` | FINDING(BE-RC-005) |
| `backend/internal/owner/service_delivery.go` | FINDING(BE-RC-005) |
| `backend/internal/owner/service_line.go` | FINDING(BE-RC-005) |
| `backend/internal/owner/validators.go` | OK |
| `backend/internal/owner/validators_contact.go` | OK |

### `backend/internal/persistence` (5)

| path | status |
|---|---|
| `backend/internal/persistence/constraints.go` | OK |
| `backend/internal/persistence/junction.go` | OK |
| `backend/internal/persistence/scope.go` | OK |
| `backend/internal/persistence/transactor.go` | OK |
| `backend/internal/persistence/tx.go` | OK |

### `backend/internal/pet` (27)

| path | status |
|---|---|
| `backend/internal/pet/animal_species_handler.go` | OK |
| `backend/internal/pet/animal_species_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/pet/animal_species_request.go` | FINDING(BE-RC-024) |
| `backend/internal/pet/animal_species_response.go` | OK |
| `backend/internal/pet/animal_species_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/pet/chronic_condition_handler.go` | OK |
| `backend/internal/pet/chronic_condition_repository.go` | OK |
| `backend/internal/pet/chronic_condition_request.go` | FINDING(BE-RC-024) |
| `backend/internal/pet/chronic_condition_service.go` | FINDING(BE-RC-005) |
| `backend/internal/pet/date.go` | OK |
| `backend/internal/pet/handler.go` | OK |
| `backend/internal/pet/mapper.go` | OK |
| `backend/internal/pet/owner_registration.go` | OK |
| `backend/internal/pet/owner_registration_adapter.go` | OK |
| `backend/internal/pet/pet_handler.go` | OK |
| `backend/internal/pet/pet_owner_handler.go` | OK |
| `backend/internal/pet/pet_owner_repository.go` | OK |
| `backend/internal/pet/pet_owner_request.go` | FINDING(BE-RC-024) |
| `backend/internal/pet/pet_owner_response.go` | OK |
| `backend/internal/pet/pet_owner_service.go` | OK |
| `backend/internal/pet/pet_request.go` | FINDING(BE-RC-024) |
| `backend/internal/pet/pet_response.go` | FINDING(BE-RC-015) |
| `backend/internal/pet/ports.go` | OK |
| `backend/internal/pet/repository.go` | FINDING(BE-RC-009,BE-RC-017) |
| `backend/internal/pet/routes.go` | OK |
| `backend/internal/pet/service.go` | FINDING(BE-RC-005) |
| `backend/internal/pet/validators.go` | OK |

### `backend/internal/reservation` (84)

| path | status |
|---|---|
| `backend/internal/reservation/appointment_admin_handler.go` | OK |
| `backend/internal/reservation/appointment_admin_repository.go` | FINDING(BE-RC-026) |
| `backend/internal/reservation/appointment_admin_request.go` | OK |
| `backend/internal/reservation/appointment_admin_response.go` | OK |
| `backend/internal/reservation/appointment_admin_service.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/appointment_notification_service.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/availability_slot_merge.go` | OK |
| `backend/internal/reservation/available_dates.go` | OK |
| `backend/internal/reservation/doc.go` | OK |
| `backend/internal/reservation/liff_handler.go` | OK |
| `backend/internal/reservation/liff_request.go` | OK |
| `backend/internal/reservation/liff_response.go` | OK |
| `backend/internal/reservation/liff_service.go` | OK |
| `backend/internal/reservation/liff_service_availability.go` | FINDING(BE-RC-005,BE-RC-025) |
| `backend/internal/reservation/liff_service_availability_business.go` | OK |
| `backend/internal/reservation/liff_service_availability_delegate.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/liff_service_availability_filters.go` | OK |
| `backend/internal/reservation/liff_service_availability_slots.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/liff_service_availability_staff.go` | FINDING(BE-RC-005,BE-RC-025) |
| `backend/internal/reservation/liff_service_availability_time.go` | OK |
| `backend/internal/reservation/liff_service_catalog.go` | FINDING(BE-RC-005,BE-RC-025) |
| `backend/internal/reservation/liff_service_health_card.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/liff_service_reservations.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/liff_validation.go` | OK |
| `backend/internal/reservation/line_reservation_setting_handler.go` | OK |
| `backend/internal/reservation/line_reservation_setting_repository.go` | OK |
| `backend/internal/reservation/line_reservation_setting_request.go` | OK |
| `backend/internal/reservation/line_reservation_setting_response.go` | OK |
| `backend/internal/reservation/line_reservation_setting_service.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/nested_summary_response.go` | OK |
| `backend/internal/reservation/reservation_capacity.go` | OK |
| `backend/internal/reservation/reservation_handler.go` | FINDING(BE-RC-015) |
| `backend/internal/reservation/reservation_intent_repository.go` | OK |
| `backend/internal/reservation/reservation_repository.go` | OK |
| `backend/internal/reservation/reservation_repository_parent_clinic.go` | OK |
| `backend/internal/reservation/reservation_repository_queries.go` | OK |
| `backend/internal/reservation/reservation_repository_slots.go` | OK |
| `backend/internal/reservation/reservation_request.go` | OK |
| `backend/internal/reservation/reservation_response.go` | OK |
| `backend/internal/reservation/reservation_schedule_handler.go` | OK |
| `backend/internal/reservation/reservation_schedule_repository.go` | OK |
| `backend/internal/reservation/reservation_schedule_request.go` | OK |
| `backend/internal/reservation/reservation_schedule_response.go` | OK |
| `backend/internal/reservation/reservation_schedule_service.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/reservation_service.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/reservation_service_update.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/reservation_service_validate.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/reservation_staff_capability_validator.go` | OK |
| `backend/internal/reservation/reservation_staff_handler.go` | OK |
| `backend/internal/reservation/reservation_staff_repository.go` | OK |
| `backend/internal/reservation/reservation_staff_request.go` | OK |
| `backend/internal/reservation/reservation_staff_response.go` | OK |
| `backend/internal/reservation/reservation_staff_service.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/reservation_type_availability_validator.go` | OK |
| `backend/internal/reservation/reservation_type_available_slot_repository.go` | OK |
| `backend/internal/reservation/reservation_type_group_handler.go` | OK |
| `backend/internal/reservation/reservation_type_group_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/reservation/reservation_type_group_request.go` | OK |
| `backend/internal/reservation/reservation_type_group_response.go` | OK |
| `backend/internal/reservation/reservation_type_group_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/reservation/reservation_type_handler.go` | OK |
| `backend/internal/reservation/reservation_type_liff_handler.go` | OK |
| `backend/internal/reservation/reservation_type_liff_repository.go` | FINDING(BE-RC-004,BE-RC-017) |
| `backend/internal/reservation/reservation_type_liff_request.go` | OK |
| `backend/internal/reservation/reservation_type_liff_response.go` | OK |
| `backend/internal/reservation/reservation_type_liff_service.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/reservation_type_occupation_repository.go` | OK |
| `backend/internal/reservation/reservation_type_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/reservation/reservation_type_request.go` | OK |
| `backend/internal/reservation/reservation_type_response.go` | OK |
| `backend/internal/reservation/reservation_type_service.go` | OK |
| `backend/internal/reservation/reservation_type_service_available_slot.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/reservation_type_service_builders.go` | OK |
| `backend/internal/reservation/reservation_type_service_core.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/reservation/reservation_type_service_occupation.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/reservation_type_service_unavailable.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/reservation_type_unavailable_time_repository.go` | OK |
| `backend/internal/reservation/reservation_validators.go` | FINDING(BE-RC-005) |
| `backend/internal/reservation/reservation_validators_create.go` | OK |
| `backend/internal/reservation/response_error.go` | OK |
| `backend/internal/reservation/routes.go` | OK |
| `backend/internal/reservation/service_deps.go` | OK |
| `backend/internal/reservation/staff_affinity.go` | OK |
| `backend/internal/reservation/timeslot_engine.go` | OK |

### `backend/internal/scheduler` (1)

| path | status |
|---|---|
| `backend/internal/scheduler/handler.go` | FINDING(BE-RC-005) |

### `backend/internal/seedbundle` (1)

| path | status |
|---|---|
| `backend/internal/seedbundle/manifest.go` | OK |

### `backend/internal/sharedkernel` (11)

| path | status |
|---|---|
| `backend/internal/sharedkernel/audit_actor.go` | OK |
| `backend/internal/sharedkernel/day_schedule.go` | OK |
| `backend/internal/sharedkernel/doc.go` | OK |
| `backend/internal/sharedkernel/enum_validators.go` | OK |
| `backend/internal/sharedkernel/go_safe.go` | OK |
| `backend/internal/sharedkernel/item_category_resolver.go` | OK |
| `backend/internal/sharedkernel/medical_record_lock.go` | FINDING(BE-RC-005) |
| `backend/internal/sharedkernel/owner_pet_link.go` | OK |
| `backend/internal/sharedkernel/pet_not_deceased.go` | OK |
| `backend/internal/sharedkernel/shift_times.go` | OK |
| `backend/internal/sharedkernel/validators.go` | FINDING(BE-RC-005) |

### `backend/internal/staff` (38)

| path | status |
|---|---|
| `backend/internal/staff/credential_audit.go` | OK |
| `backend/internal/staff/handler.go` | OK |
| `backend/internal/staff/http_binding.go` | OK |
| `backend/internal/staff/occupation_handler.go` | OK |
| `backend/internal/staff/occupation_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/staff/occupation_request.go` | OK |
| `backend/internal/staff/occupation_response.go` | OK |
| `backend/internal/staff/occupation_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/staff/permission_assignment_audit.go` | OK |
| `backend/internal/staff/ports.go` | OK |
| `backend/internal/staff/reservation_staff_update.go` | OK |
| `backend/internal/staff/shift_entry_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/staff/shift_entry_service.go` | FINDING(BE-RC-005) |
| `backend/internal/staff/shift_handler.go` | OK |
| `backend/internal/staff/shift_request.go` | OK |
| `backend/internal/staff/shift_response.go` | OK |
| `backend/internal/staff/shift_template_handler.go` | OK |
| `backend/internal/staff/shift_template_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/staff/shift_template_request.go` | OK |
| `backend/internal/staff/shift_template_response.go` | OK |
| `backend/internal/staff/shift_template_service.go` | FINDING(BE-RC-005) |
| `backend/internal/staff/staff_clinic_assignment_repository.go` | OK |
| `backend/internal/staff/staff_clinic_assignment_service.go` | FINDING(BE-RC-005) |
| `backend/internal/staff/staff_handler.go` | OK |
| `backend/internal/staff/staff_provisioning.go` | FINDING(BE-RC-009) |
| `backend/internal/staff/staff_provisioning_apply.go` | OK |
| `backend/internal/staff/staff_provisioning_repository.go` | OK |
| `backend/internal/staff/staff_provisioning_validate.go` | OK |
| `backend/internal/staff/staff_repository.go` | FINDING(BE-RC-009,BE-RC-015,BE-RC-017) |
| `backend/internal/staff/staff_request.go` | OK |
| `backend/internal/staff/staff_response.go` | OK |
| `backend/internal/staff/staff_service.go` | FINDING(BE-RC-009) |
| `backend/internal/staff/staff_service_account.go` | FINDING(BE-RC-005) |
| `backend/internal/staff/staff_service_builders.go` | OK |
| `backend/internal/staff/staff_service_core.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/staff/staff_service_permissions.go` | FINDING(BE-RC-005) |
| `backend/internal/staff/staff_service_update.go` | OK |
| `backend/internal/staff/validators.go` | OK |

### `backend/internal/testdb` (2)

| path | status |
|---|---|
| `backend/internal/testdb/fixtures.go` | SKIP(テストカーネル) |
| `backend/internal/testdb/testdb.go` | SKIP(テストカーネル) |

### `backend/internal/textsearch` (1)

| path | status |
|---|---|
| `backend/internal/textsearch/textsearch.go` | OK |

### `backend/internal/timeutil` (2)

| path | status |
|---|---|
| `backend/internal/timeutil/format.go` | OK |
| `backend/internal/timeutil/weekday.go` | OK |

### `backend/internal/trimming` (31)

| path | status |
|---|---|
| `backend/internal/trimming/handler.go` | OK |
| `backend/internal/trimming/ports.go` | OK |
| `backend/internal/trimming/query_helpers.go` | OK |
| `backend/internal/trimming/response_summaries.go` | OK |
| `backend/internal/trimming/routes.go` | OK |
| `backend/internal/trimming/trimming_audit.go` | OK |
| `backend/internal/trimming/trimming_course_handler.go` | OK |
| `backend/internal/trimming/trimming_course_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/trimming/trimming_course_request.go` | OK |
| `backend/internal/trimming/trimming_course_response.go` | OK |
| `backend/internal/trimming/trimming_course_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/trimming/trimming_course_type_handler.go` | OK |
| `backend/internal/trimming/trimming_course_type_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/trimming/trimming_course_type_request.go` | OK |
| `backend/internal/trimming/trimming_course_type_response.go` | OK |
| `backend/internal/trimming/trimming_course_type_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/trimming/trimming_handler.go` | OK |
| `backend/internal/trimming/trimming_option_handler.go` | OK |
| `backend/internal/trimming/trimming_option_repository.go` | FINDING(BE-RC-017) |
| `backend/internal/trimming/trimming_option_request.go` | OK |
| `backend/internal/trimming/trimming_option_response.go` | OK |
| `backend/internal/trimming/trimming_option_service.go` | FINDING(BE-RC-004,BE-RC-005) |
| `backend/internal/trimming/trimming_repository.go` | OK |
| `backend/internal/trimming/trimming_request.go` | OK |
| `backend/internal/trimming/trimming_response.go` | OK |
| `backend/internal/trimming/trimming_service.go` | FINDING(BE-RC-005,BE-RC-015) |
| `backend/internal/trimming/trimming_service_create.go` | OK |
| `backend/internal/trimming/trimming_service_mutate.go` | FINDING(BE-RC-005) |
| `backend/internal/trimming/trimming_service_update.go` | OK |
| `backend/internal/trimming/trimming_service_validate.go` | OK |
| `backend/internal/trimming/validators.go` | OK |

---

## 9. 監査メタ

- worktree: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-be-full-audit`
- branch: `refactor/be-rc-full-audit-2026-09`
- claim: `claim/BE-RC-FULL-AUDIT-2026-09`（削除しない）
- runtime `go test` は未実行（docs-only）
- loop-health: 読了 981 / 母集団 981
