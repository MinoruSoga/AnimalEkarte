# backend コード規約チェック結果（2026-09-04 HEAD re-audit）

`backend/` の production `.go` を規約正本に照合し、改善すべき開いた所見だけを残す。コードは修正していない。

- HEAD: `6739b64021607f219f5159f4090f23d1fd6747bb`
- 母集団: `git ls-files 'backend/**/*.go' | grep -v '_test.go$' | grep -v '/_archive/' | sort` → **982**（欠番 0）。`_test.go` と `cmd/_archive` は本番 invariant の母集団外
- 規約正本: `AGENTS.md`、`CLAUDE.md`、`.claude/CLAUDE.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/rules/tdd.md`、`.claude/rules/git-worktree-safety.md`、`.claude/refs/go-gin-backend-review.md`、`.claude/refs/backend-application-invariants.md`、`.claude/refs/error-handling.md`、`.claude/refs/naming-conventions.md`、`.claude/refs/go-language.md`、`.claude/refs/api.md`、`backend/CLAUDE.md`、`backend/CODING_RULES.md`、`backend/migrations/CLAUDE.md`、`backend/.golangci.yml`、`docs/product-philosophy.md`、ADR-001〜008、`docs/architecture/be9-2a-boundary-map.md`、project `.agents/skills` の backend 系（clinic-id-isolation / go-gin-backend / go-security / golang-testing / golang-refactoring / golang-gin-api / gin-api-design / postgres-patterns / database-indexing / security-checklist / scoped-verification-gates / migration-seed-safety）
- 方法: 現行 HEAD で母集団を取り直し、production `.go` をバッチ精読。機械検出で BindJSON / MustBind、`slog.ErrorContext` in `*service*.go`、`err.Error()` Contains、exported `Update(..., map[string]any)`、800 行超を再走査。旧台帳 HEAD `321fe2b8d` / 981 は stale とし正本にしない。旧 DONE は未修正として再掲しない
- 各所見の `path` は `backend/` 起点。行番号は `6739b6402` 時点
- Handler → Service → Repository / Clean Architecture は Go/Gin 公式要件ではない。再導入しない
- カバレッジ行: OK 965 / FINDING 14 / SKIP 3
- 開いた所見: HIGH 0 / MEDIUM 2 / LOW 2
- 2026-09-06 に閉じた: BE-RC-021 / 023 / 034 / 035 / 036
- 2026-09-07 に閉じた: BE-RC-009 / 015

規約矛盾は Truth Source Priority で解決した（キャンペーン BLOCKED にしない）: gosec は `.golangci.yml` enable が正。旧 3 layer は ADR / 現行 tree が正。seed バンドルは ADR-004 / `backend/migrations/CLAUDE.md` の `002_master` が正。

---

## 1. 方法

1. worktree HEAD `6739b64021607f219f5159f4090f23d1fd6747bb` で母集団コマンドを実行（982）。ff-only merge 後。`chief_complaint_repository.go` / `exam_type_repository.go` の原子 DELETE は回帰なし。
2. 規約正本を pre-read。個人 `~/.cursor/rules` は正本外。
3. ディレクトリ構成・残差機械検出・パッケージ精読を並列。ledger writer は1人。
4. 開いた所見だけを HIGH/MEDIUM/LOW に載せ、閉じ済みは §6。カバレッジ表は全 982。

---

## 2. フォルダ／ファイル構成

`backend/internal/` は ADR-006 の 14 target domain（owner / pet / staff / auth / reservation / trimming / medicalrecord / billing / inventory / lstep / clinic / manualarticle / httpapi / identitylink）と命名済み cross-cutting（config / dbconn / middleware / infra / model / timeutil / seedbundle / logger / csvimport / labdeviceagent / authjwt / apperrors / apicontract / lintscan / audit / persistence / scheduler / sharedkernel / textsearch / testdb）で構成される。

`backend/cmd/` は `api`（composition root、22 file）、`migrate`、`csv-import*`、`lab-device-agent`、`lstep-migrate`、`seed-export`、`staff-provision`、`stg-uat-*`、`coverage-ratchet`。`cmd/_archive` は母集団外。

**合格**

| 項目 | 実測 |
|---|---|
| 旧 `internal/handler` `internal/service` `internal/repository` ディレクトリ | 不存在（`test ! -d` 3 つとも成功） |
| domain 内 `handler/` `service/` `repository/` サブパッケージ | 0 |
| production `package util\|common\|misc` | 0。`timeutil` / `sharedkernel` は固有名 |
| 極端な package サイズ | `medicalrecord` 239、`lstep` 133、`model` 94、`billing` 93、`reservation` 84。800 行超 production ファイル 0（最大 718: `staff/staff_repository.go`。次点 `pet/repository.go` 702、`billing/accounting_repository.go` 685、`medicalrecord/lab_device_receive_service.go` 672） |
| `cmd/api` | domain を直接 import する composition root。業務ルールの複製は BE-RC-034 で閉じた |

フォルダ構成の逸脱（新設すべき util/common、layer 再分割）は提案しない。testdb の `truncate.go` は advisory lock + deadlock retry のテストカーネルであり、移動しない。

---

## 3. HIGH（開いた所見のみ）

なし。

---

## 4. MEDIUM（開いた所見のみ）

#### BE-RC-005 [MEDIUM][residual] service の `slog.ErrorContext` と handler `RespondError` が 5xx を二重ログする
- 規約: 同じ error を複数層で重複ログしない。未知 pg だけ request 境界の `c.Error` 重複を例外許容
- 現状: `*service*.go` は **117 files / 722 hits**。5xx は `httpapi/response.go:15-19` が `c.Error` → `middleware/logging.go:50-57` が再記録
- 代表: `billing/insurance_service.go:87,95,122,133,148,160,167,179` → insurance handler `RespondError`、`inventory/inventory_service.go:103,112,138,150,159,173,181`、`audit/service.go:151`、`manualarticle/service.go:46`
- 除外（005 を付けない）: 運用ログ・バッチ best-effort・stream 後（例: `cmd/api/composition_auth_routes.go`、`lstep` tag summary handler、`reservation` appointment notification、`lstep_batch_*.go` / `lstep_delivery_trigger_*.go` / `lstep_health_tag_sync_*.go` / `lstep_tag_sync_*.go`）。`*service.go` ではない handler / middleware / validators
- 改善案: 既知 4xx は service でログしない。5xx は middleware 一本化。新規から増やさない。未触 service の一括削除はしない
- 返却 5xx の `slog.ErrorContext` を外した面に、このスライスで `exam_type` / `hospitalization_plan` / `inquiry_template` / `vaccine` / `procedure` / `diagnosis` / `medicine_dose_param` / `estimate_service_tx` / `checkup` / `prescription` / `vaccination` / `vital_service_update` を追加した
- 運用ログとして残す: `audit_write_failed`、company post-update reload、password_reset mail ops、billing_item campaign suggestion、checkup / prescription / vaccination の tag sync best-effort。`clinic_service.ListClinics` は login 非 admin が失敗を捨てるため診断ログを残す
- 2026-09-07: 返却 5xx の dirty 面は尽きた。残りは examination 系（再開禁止）と `lstep/shared_file_service.go` / `clinical_plan_service.go` / `dose_revalidation.go`（同一 abort バッチ）。運用ログ残置は台帳 OK

#### BE-RC-014 [MEDIUM][residual] pgx encode 判定が `err.Error()` 文字列 Contains
- 対象: `internal/apperrors/errors.go:344` — `isPgxEncodeRangeMessage(err.Error())`（定義 `:371-385`、DEC-34 / BUG-138）
- 規約: `errors.Is` / `As`。pgx v5.10.0 は encode を `fmt.Errorf("unable to encode ...")` のまま出し、typed EncodeError は無い
- 改善案: 新規に同パターンを増やさない。pgx が typed error を出したら `errors.As` へ。LSTEP の同型は BE-RC-033（§6）

---

## 5. LOW（開いた所見のみ）

#### BE-RC-017 [LOW][residual] 同一 package の exported `Update(..., map[string]any)` が常態
- 機械検出 **51 production files**。カバレッジは exported `Update(..., map[string]any)` シグネチャのあるファイルのみ。builder / GORM `Updates` 呼び出し側には付けない
- 先行台帳の偽陰性 2 件（chronic_condition / lab_device_item_master）と cage / care_plan_item / checkup / checkup_type / chief_complaint / consultation / diagnosis_* / exam_type / hospitalization_plan / inquiry_template / vaccine / procedure / medicine_dose_param / occupation / trimming_course_type / trimming_option / trimming_course / payment_method_master / insurance / reservation_type_group / reservation_type / animal_species / reservation_type_liff / shift_template / campaign / closing_special_period / merchandise_item / inventory / company / estimate / billing_item / billing_confirmation / shift_entry / clinic / permission_group / pet / vital / prescription / vaccination / accounting / staff / medicine / hospitalization / treatment_plan / clinical_plan / treatment / medical_record は typed command へ閉じた
- 001/008 の境界（reservation 向け staff、consumer `ClinicRepository`）は typed 済み。回帰なし
- 改善案: 触る repository から unexported `update` にし、外には typed command だけ出す。一括 unexport はしない
- 2026-09-07: exported map Update の閉じた列は medical_record まで。残りは `examination_repository.go`（再開禁止）

#### BE-RC-019 [LOW][residual] `medicalrecord` 本番 239 file の凝集圧
- layer サブパッケージ化は禁止方針どおり避けている。800 行超ファイルなし（package 内最大 `lab_device_receive_service.go` **672** 行。backend 全体最大は `staff/staff_repository.go` **718** 行）
- 800 行超 0 は別の §6 項目であり、019 を閉じる根拠にしない
- カバレッジ代表: `medical_record_repository.go`、`lab_device_receive_service.go:1-672`。239 ファイルすべてには付けない
- 改善案: 分割するなら業務能力（lab / hospitalization）単位。急がない。`handler/service/repository` サブパッケージ化はしない
- 2026-09-06: lab / hospitalization の独立 consumer・変更周期は未成立。019 は BLOCKED のまま。層分割では閉じない

---

## 6. 再検証した合格（未修正として再掲しない）

| 項目 | 根拠 |
|---|---|
| BE-RC-001 typed staff update | `staff/reservation_staff_update.go`、`UpdateForReservation(..., ReservationStaffUpdate)`。回帰なし |
| BE-RC-002 閉集合の `max` | LINE text、refund reason、addendum、OwnerName/Pet Name leftover 等は `max` 維持。列挙残件は 024 で閉じた |
| BE-RC-003 原子 DELETE 3 面 | inventory `DeleteIfUnused`、payment_method / insurance 原子 DELETE。service Count は UX |
| BE-RC-004 列挙マスター | consultation/vaccine/checkup_type/diagnosis_*/chief_complaint/hospitalization_plan/procedure/exam_type/cage/medicine、reservation_type*、trimming_*、merchandise、occupation、permission_group、animal_species、clinic、staff、**inquiry_template**。`inquiry_template_repository.go:74-86` は `DBOrTx` + `ClinicScope` + `id` の原子 DELETE。CountUsage stub `0, nil` は UX（FK 未実装）。medicine/procedure/vaccine Delete も `DBOrTx` + `ClinicScope` + `NOT EXISTS` |
| BE-RC-024 HTTP `max` | clinic/billing/inventory/pet/owner/identitylink の列挙フィールド。name 255 / reason 500 / memo 1000 / search 255 |
| BE-RC-025 LIFF `IsActive` | `liff_service_availability_staff.go` が `IsActive && ReservationVisible`。書込二重防御は残置 |
| BE-RC-026 admin Preload | `appointment_admin_repository.go` の Doctor/CreatedByStaff が `staffAssignedToClinicsCond` + `reservationRelationsMatchParentClinic` |
| BE-RC-027 LineLinkToken | `model/line_link_token.go` の Token は `json:"-"`。`line_reservation_setting.go` の `LineChannelSecret` / `LineAccessToken` も `json:"-"`。回帰なし |
| BE-RC-028 LSTEP fail-closed | nil trigger / nil settingsSvc は error。Find 失敗は slog。intentional skip は slog。ownerIDs 空は対象外のまま |
| BE-RC-030 BE9 コメント | `medicalrecord/pagination.go` / `service_deps.go` を現行 domain に置換。残る `internal/service` 言及は移設履歴 |
| BE-RC-031 lab nil tx | `lab_device_receive_service.go:47-51` の `withTx` は `tx == nil` で fail-closed |
| BE-RC-032 same-tx reload | insurance Update と accounting CorrectCreditPayment の reload を commit 前へ |
| BE-RC-033 LINE classify | `errors.As` + `net.Error.Timeout()` を Contains より先。Contains 増殖禁止 |
| BE-RC-006 列挙 handler `err.Error()` | leftover 閉集合に連結なし。残件 CSV 経路は固定日本語 |
| BE-RC-007 payment_method reload | 同一 `Transaction` で update+reload。保険・訂正は 032 で閉じた |
| BE-RC-008 `UpdateClinic` | consumer `ClinicRepository` から map `Update` 除去 |
| BE-RC-010 `LstepRepository` | exported 0。`LifecycleOwnerRepository` |
| BE-RC-012 wrapcheck `internal/*` wildcard | 廃止済み。残るのは明示 package 列挙（§7） |
| BE-RC-013 vaccination `t.Setenv` | `os.Setenv` 0 |
| BE-RC-016 stale `// Package handler\|service\|repository` | production 0 |
| BE-RC-009 実装側の広い Repository | use-case port + concrete ctor。`permission_group` / provisioning / identitylink / clinic ports / accounting / medical_record / staff。017 map Update は残置。既存の一括分割はしない |
| BE-RC-015 package.Type stutter 代表 | `clinic.Service` / `auth.Service` / `trimming.Service` / `staff.Service` / `staff.Repository` / `pet.Response` / `lstep.SettingsHandler` / `reservation.CRUDHandler`。カバレッジは代表のみ。一括 rename しない |
| BE-RC-021 Config/Load GoDoc | `config/config.go` の exported 代表に GoDoc |
| BE-RC-023 clinic binding `init()` | `RegisterContactBindingValidators` を `NewHandler` / `RegisterClinicRoutes` から呼ぶ。production `init()` なし |
| BE-RC-035 auth 生 TRUNCATE | `current_access_staff_reader_db_test.go` は `testdb.Truncate` |
| BE-RC-036 staff Package コメント | `occupation` / `shift_entry` / `shift_template` / `staff_clinic_assignment` は `// Package staff` |
| BE-RC-034 cmd 業務ルール複製 | clinic pick / 権限表 / cutover 分類 / audit report 作成は owner package。cmd は薄い delegate |
| BE-RC-018 staffs/shift_entries AST gate | `staff/staff_table_write_owner_lint_test.go` 存在 |
| BE-RC-022 `replace_audit_tail` | `internal/service` を正本扱いしない |
| BE-RC-011 見積通常 CRUD 監査 | 意図的 post-commit best-effort。CreateSuccessor のみ fail-closed。再提案しない |
| BE-RC-020 `nested_summary_response.go` | billing / medicalrecord / reservation の意図的コピー。統合しない |
| 旧 3 layer directory | 不存在 |
| ADR-006 DAG / appointments write owner | AST gate 維持。trimming は typed intent |
| `ShouldBind*` | 本番 `.BindJSON(` / `MustBind*` 0 |
| Context struct 保持 / 生 `*gin.Context` goroutine | 未検出 |
| CORS wildcard + credentials | なし |
| AutoMigrate | testdb のみ |
| 本番 800 行超 | 0（最大 `staff/staff_repository.go` 718） |
| staff Preload `deleted_at` only | `lintscan/preload_clinic_scope_lint_test.go:142-159` の `staffExemptAssoc`。名前漏洩のみ・write isolation。reservation admin の 026 は閉じた |
| testdb Truncate | `testdb/truncate.go` は advisory lock + deadlock retry。3 production files は SKIP(test kernel) |
| wrapcheck interface ignore | `ignore-interface-regexps: .+` は意図的。新 ID にしない |

---

## 7. lint / 設定上の既知緩和

現行 `backend/.golangci.yml`（一括解除・設定変更は Out of scope。新 ID にしない）。

- wrapcheck: `internal/*` wildcard は廃止（012 DONE）。`ignore-interface-regexps: .+`（`:122-123`）。ignore-sigs に `.WithTx(`（`:94`）と reservation / trimming validators（`:112-118`）。campaign-touched（staff, reservation, billing, inventory, lstep, medicalrecord, owner, pet, clinic, auth）は wrapcheck 対象のまま。未触 package は `ignore-package-globs`（`:124-151`、`trimming` 含む）。testdb は glob 除外。**一括 wrap キャンペーンはしない**（AppError 二重ラップになる）
- `gocritic` の `hugeParam` / `unnamedResult` / `rangeValCopy` disable（composition/DTO）
- `revive` の `exported` / `package-comments` / `context-as-argument` / `unexported-return` disable
- `contextcheck` は `internal/middleware` と `cmd/api` で除外
- `cmd/csv-import*` / `internal/csvimport` の gocritic/gosec 緩和
- gosec は enable（skills の「未導入」記述より yml が正）

---

## 8. 却下済み・再提案しない（再開条件付き）

- カルテ同日重複に DB unique を採らない（2026-07-27）。再開条件 = 手動作成経路で実害が出た場合のみ
- auto-create に clock seam を導入しない（2026-07-27）。予約日基準が正
- Count→Delete の**一括** retrofit を本監査の実装スコープにしない（CODING_RULES。触る Delete では直す）
- medicalrecord を `handler/service/repository` サブパッケージへ層分割しない
- `map[string]any` の監査 metadata / テスト fixture を禁止しない
- wrapcheck を host の full `golangci-lint run ./...` でエージェントが回さない
- stutter 一括 rename、GoDoc 一括、同一 package map Update 一括 unexport、testdb.Truncate 一括移行
- `nested_summary` 3 package コピーの統合（import cycle / 意図的非統合）
- `util` / `common` / `misc` 新設
- hospitalization / estimate / cash_register の `Doctor`/`CreatedStaff`/`ClosedByStaff` Preload を HIGH として一括是正しない（`staffExemptAssoc`。再開条件 = write isolation を迂回する破損 FK が実害を出したとき。reservation admin の 026 は同 package に assignment 正例があるため対象外）

---

## 9. カバレッジ表（production `.go` 全 982）

各行は `OK` / `FINDING(IDs)` / `SKIP(理由)`。FINDING の ID は開いた所見のみ（DONE 済み ID を未修正として付けない）。015/019/021 は系統的残差のため代表ファイルのみタグ。035 は test-only のため本表に無い。

### `backend/cmd/api` (22)

| path | status |
|---|---|
| `backend/cmd/api/base_routes.go` | OK |
| `backend/cmd/api/batch_scheduler.go` | OK |
| `backend/cmd/api/composition_auth.go` | OK |
| `backend/cmd/api/composition_auth_routes.go` | OK |
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

### `backend/cmd/csv-import` (1)

| path | status |
|---|---|
| `backend/cmd/csv-import/main.go` | OK |

### `backend/cmd/csv-import-failure-rehearsal` (1)

| path | status |
|---|---|
| `backend/cmd/csv-import-failure-rehearsal/main.go` | OK |

### `backend/cmd/csv-import-stg-uat` (1)

| path | status |
|---|---|
| `backend/cmd/csv-import-stg-uat/main.go` | OK |

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
| `backend/internal/audit/service.go` | OK |

### `backend/internal/auth` (34)

| path | status |
|---|---|
| `backend/internal/auth/account_repository.go` | OK |
| `backend/internal/auth/account_service.go` | OK |
| `backend/internal/auth/auth_service.go` | OK |
| `backend/internal/auth/credential_audit.go` | OK |
| `backend/internal/auth/current_access_service.go` | OK |
| `backend/internal/auth/current_access_staff_reader.go` | OK |
| `backend/internal/auth/doc.go` | OK |
| `backend/internal/auth/http_binding.go` | OK |
| `backend/internal/auth/http_password.go` | OK |
| `backend/internal/auth/http_permission.go` | OK |
| `backend/internal/auth/http_response.go` | OK |
| `backend/internal/auth/http_routes.go` | OK |
| `backend/internal/auth/http_session.go` | OK |
| `backend/internal/auth/http_session_cookies.go` | OK |
| `backend/internal/auth/http_session_login.go` | OK |
| `backend/internal/auth/http_session_me.go` | OK |
| `backend/internal/auth/http_session_refresh.go` | OK |
| `backend/internal/auth/http_types.go` | OK |
| `backend/internal/auth/login_failure_audit.go` | OK |
| `backend/internal/auth/login_response_floor.go` | OK |
| `backend/internal/auth/password_reset_persist.go` | OK |
| `backend/internal/auth/password_reset_response_floor.go` | OK |
| `backend/internal/auth/password_reset_service.go` | OK |
| `backend/internal/auth/password_reset_token_repository.go` | OK |
| `backend/internal/auth/password_timing.go` | OK |
| `backend/internal/auth/permission_group_create_request.go` | OK |
| `backend/internal/auth/permission_group_repository.go` | OK |
| `backend/internal/auth/permission_group_service.go` | OK |
| `backend/internal/auth/permission_group_service_mutate.go` | OK |
| `backend/internal/auth/permission_group_service_rules.go` | OK |
| `backend/internal/auth/refresh_family.go` | OK |
| `backend/internal/auth/token_blacklist_repository.go` | OK |
| `backend/internal/auth/token_blacklist_service.go` | OK |
| `backend/internal/auth/token_service.go` | OK |

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
| `backend/internal/billing/accounting_report_service.go` | OK |
| `backend/internal/billing/accounting_report_service_monthly_build.go` | OK |
| `backend/internal/billing/accounting_reports_dto.go` | OK |
| `backend/internal/billing/accounting_repository.go` | OK |
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
| `backend/internal/billing/accounting_request.go` | OK |
| `backend/internal/billing/accounting_response.go` | OK |
| `backend/internal/billing/accounting_service.go` | OK |
| `backend/internal/billing/accounting_service_builders.go` | OK |
| `backend/internal/billing/accounting_service_core.go` | OK |
| `backend/internal/billing/accounting_service_correction.go` | OK |
| `backend/internal/billing/accounting_service_reports.go` | OK |
| `backend/internal/billing/accounting_service_update.go` | OK |
| `backend/internal/billing/allocation.go` | OK |
| `backend/internal/billing/billing_confirmation_handler.go` | OK |
| `backend/internal/billing/billing_confirmation_repository.go` | OK |
| `backend/internal/billing/billing_confirmation_request.go` | OK |
| `backend/internal/billing/billing_confirmation_response.go` | OK |
| `backend/internal/billing/billing_confirmation_service.go` | OK |
| `backend/internal/billing/billing_item_exam.go` | OK |
| `backend/internal/billing/billing_item_handler.go` | OK |
| `backend/internal/billing/billing_item_repository.go` | OK |
| `backend/internal/billing/billing_item_repository_trimming_refs.go` | OK |
| `backend/internal/billing/billing_item_repository_unbilled.go` | OK |
| `backend/internal/billing/billing_item_repository_vaccination.go` | OK |
| `backend/internal/billing/billing_item_repository_vaccination_lock.go` | OK |
| `backend/internal/billing/billing_item_request.go` | OK |
| `backend/internal/billing/billing_item_service.go` | OK |
| `backend/internal/billing/billing_item_service_create.go` | OK |
| `backend/internal/billing/billing_item_service_update.go` | OK |
| `backend/internal/billing/billing_item_unbilled.go` | OK |
| `backend/internal/billing/billing_service.go` | OK |
| `backend/internal/billing/campaign_discount.go` | OK |
| `backend/internal/billing/campaign_handler.go` | OK |
| `backend/internal/billing/campaign_repository.go` | OK |
| `backend/internal/billing/campaign_request.go` | OK |
| `backend/internal/billing/campaign_response.go` | OK |
| `backend/internal/billing/campaign_service.go` | OK |
| `backend/internal/billing/cash_register_close_repository.go` | OK |
| `backend/internal/billing/cash_register_handler.go` | OK |
| `backend/internal/billing/cash_register_request.go` | OK |
| `backend/internal/billing/cash_register_response.go` | OK |
| `backend/internal/billing/cash_register_service.go` | OK |
| `backend/internal/billing/doc.go` | OK |
| `backend/internal/billing/estimate_audit_diff.go` | OK |
| `backend/internal/billing/estimate_handler.go` | OK |
| `backend/internal/billing/estimate_repository.go` | OK |
| `backend/internal/billing/estimate_request.go` | OK |
| `backend/internal/billing/estimate_response.go` | OK |
| `backend/internal/billing/estimate_service.go` | OK |
| `backend/internal/billing/estimate_service_successor.go` | OK |
| `backend/internal/billing/estimate_service_successor_tx.go` | OK |
| `backend/internal/billing/estimate_service_tx.go` | OK |
| `backend/internal/billing/insurance_handler.go` | OK |
| `backend/internal/billing/insurance_repository.go` | OK |
| `backend/internal/billing/insurance_request.go` | OK |
| `backend/internal/billing/insurance_response.go` | OK |
| `backend/internal/billing/insurance_service.go` | OK |
| `backend/internal/billing/list_query_helpers_test_free.go` | OK |
| `backend/internal/billing/nested_summary_response.go` | OK |
| `backend/internal/billing/payment_method_master_handler.go` | OK |
| `backend/internal/billing/payment_method_master_repository.go` | OK |
| `backend/internal/billing/payment_method_master_request.go` | OK |
| `backend/internal/billing/payment_method_master_response.go` | OK |
| `backend/internal/billing/payment_method_master_service.go` | OK |
| `backend/internal/billing/payment_method_scope.go` | OK |
| `backend/internal/billing/refund_handler.go` | OK |
| `backend/internal/billing/refund_repository.go` | OK |
| `backend/internal/billing/refund_request.go` | OK |
| `backend/internal/billing/refund_service.go` | OK |
| `backend/internal/billing/refund_service_tx.go` | OK |
| `backend/internal/billing/routes.go` | OK |
| `backend/internal/billing/service_deps.go` | OK |
| `backend/internal/billing/unpaid_amount.go` | OK |
| `backend/internal/billing/validators_billing_item.go` | OK |
| `backend/internal/billing/validators_insurance.go` | OK |

### `backend/internal/clinic` (25)

| path | status |
|---|---|
| `backend/internal/clinic/clinic_handler.go` | OK |
| `backend/internal/clinic/clinic_holiday_handler.go` | OK |
| `backend/internal/clinic/clinic_holiday_repository.go` | OK |
| `backend/internal/clinic/clinic_holiday_request.go` | OK |
| `backend/internal/clinic/clinic_holiday_response.go` | OK |
| `backend/internal/clinic/clinic_holiday_service.go` | OK |
| `backend/internal/clinic/clinic_repository.go` | OK |
| `backend/internal/clinic/clinic_request.go` | OK |
| `backend/internal/clinic/clinic_response.go` | OK |
| `backend/internal/clinic/clinic_service.go` | OK |
| `backend/internal/clinic/clinic_service_update.go` | OK |
| `backend/internal/clinic/clinic_settings_repository.go` | OK |
| `backend/internal/clinic/closing_settings_handler.go` | OK |
| `backend/internal/clinic/closing_settings_request.go` | OK |
| `backend/internal/clinic/closing_settings_response.go` | OK |
| `backend/internal/clinic/closing_settings_service.go` | OK |
| `backend/internal/clinic/closing_special_period_repository.go` | OK |
| `backend/internal/clinic/company_handler.go` | OK |
| `backend/internal/clinic/company_repository.go` | OK |
| `backend/internal/clinic/company_request.go` | OK |
| `backend/internal/clinic/company_response.go` | OK |
| `backend/internal/clinic/company_service.go` | OK |
| `backend/internal/clinic/handler.go` | OK |
| `backend/internal/clinic/ports.go` | OK |
| `backend/internal/clinic/repositories.go` | OK |

### `backend/internal/config` (2)

| path | status |
|---|---|
| `backend/internal/config/config.go` | OK |
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
| `backend/internal/identitylink/handler.go` | OK |
| `backend/internal/identitylink/repository.go` | OK |
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

### `backend/internal/infra` (7)

| path | status |
|---|---|
| `backend/internal/infra/file_storage.go` | OK |
| `backend/internal/infra/local_file_storage.go` | OK |
| `backend/internal/infra/local_uploader.go` | OK |
| `backend/internal/infra/s3_endpoint.go` | OK |
| `backend/internal/infra/s3_file_storage.go` | OK |
| `backend/internal/infra/s3_uploader.go` | OK |
| `backend/internal/infra/uploader.go` | OK |

### `backend/internal/infra/crypto` (1)

| path | status |
|---|---|
| `backend/internal/infra/crypto/aes_gcm.go` | OK |

### `backend/internal/infra/httpx` (1)

| path | status |
|---|---|
| `backend/internal/infra/httpx/retry.go` | OK |

### `backend/internal/infra/line` (3)

| path | status |
|---|---|
| `backend/internal/infra/line/client.go` | OK |
| `backend/internal/infra/line/errors.go` | OK |
| `backend/internal/infra/line/push.go` | OK |

### `backend/internal/infra/lstep` (6)

| path | status |
|---|---|
| `backend/internal/infra/lstep/breed_codes.go` | OK |
| `backend/internal/infra/lstep/client.go` | OK |
| `backend/internal/infra/lstep/dial.go` | OK |
| `backend/internal/infra/lstep/errors.go` | OK |
| `backend/internal/infra/lstep/tag.go` | OK |
| `backend/internal/infra/lstep/user.go` | OK |

### `backend/internal/infra/smtp` (1)

| path | status |
|---|---|
| `backend/internal/infra/smtp/sender.go` | OK |

### `backend/internal/inventory` (12)

| path | status |
|---|---|
| `backend/internal/inventory/handler.go` | OK |
| `backend/internal/inventory/inventory_handler.go` | OK |
| `backend/internal/inventory/inventory_request.go` | OK |
| `backend/internal/inventory/inventory_response.go` | OK |
| `backend/internal/inventory/inventory_service.go` | OK |
| `backend/internal/inventory/merchandise_item_handler.go` | OK |
| `backend/internal/inventory/merchandise_item_repository.go` | OK |
| `backend/internal/inventory/merchandise_item_request.go` | OK |
| `backend/internal/inventory/merchandise_item_response.go` | OK |
| `backend/internal/inventory/merchandise_item_service.go` | OK |
| `backend/internal/inventory/repository.go` | OK |
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
| `backend/internal/lstep/aggregation_service.go` | OK |
| `backend/internal/lstep/checkup_sync_handler.go` | OK |
| `backend/internal/lstep/checkup_sync_repository.go` | OK |
| `backend/internal/lstep/checkup_sync_repository_preview_sql.go` | OK |
| `backend/internal/lstep/checkup_sync_request.go` | OK |
| `backend/internal/lstep/checkup_sync_response.go` | OK |
| `backend/internal/lstep/checkup_sync_service.go` | OK |
| `backend/internal/lstep/checkup_sync_service_create.go` | OK |
| `backend/internal/lstep/checkup_sync_service_metadata.go` | OK |
| `backend/internal/lstep/checkup_sync_service_preview.go` | OK |
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
| `backend/internal/lstep/line_customer_service.go` | OK |
| `backend/internal/lstep/line_link_handler.go` | OK |
| `backend/internal/lstep/line_link_request.go` | OK |
| `backend/internal/lstep/line_link_service.go` | OK |
| `backend/internal/lstep/line_link_token_repository.go` | OK |
| `backend/internal/lstep/line_messaging_service.go` | OK |
| `backend/internal/lstep/line_send_handler.go` | OK |
| `backend/internal/lstep/line_send_log_repository.go` | OK |
| `backend/internal/lstep/line_send_request.go` | OK |
| `backend/internal/lstep/line_send_response.go` | OK |
| `backend/internal/lstep/line_send_service.go` | OK |
| `backend/internal/lstep/lstep_analytics_handler.go` | OK |
| `backend/internal/lstep/lstep_analytics_request.go` | OK |
| `backend/internal/lstep/lstep_analytics_response.go` | OK |
| `backend/internal/lstep/lstep_analytics_service.go` | OK |
| `backend/internal/lstep/lstep_batch_delivery.go` | OK |
| `backend/internal/lstep/lstep_batch_dormant.go` | OK |
| `backend/internal/lstep/lstep_batch_noshow.go` | OK |
| `backend/internal/lstep/lstep_batch_segmentation.go` | OK |
| `backend/internal/lstep/lstep_batch_service.go` | OK |
| `backend/internal/lstep/lstep_csv_helpers.go` | OK |
| `backend/internal/lstep/lstep_csv_import_concurrency.go` | OK |
| `backend/internal/lstep/lstep_csv_import_handler.go` | OK |
| `backend/internal/lstep/lstep_csv_import_owner_lookup.go` | OK |
| `backend/internal/lstep/lstep_csv_import_prepare.go` | OK |
| `backend/internal/lstep/lstep_csv_import_processing.go` | OK |
| `backend/internal/lstep/lstep_csv_import_repository.go` | OK |
| `backend/internal/lstep/lstep_csv_import_request.go` | OK |
| `backend/internal/lstep/lstep_csv_import_service.go` | OK |
| `backend/internal/lstep/lstep_csv_stream.go` | OK |
| `backend/internal/lstep/lstep_delivery_deps.go` | OK |
| `backend/internal/lstep/lstep_delivery_monitor_handler.go` | OK |
| `backend/internal/lstep/lstep_delivery_monitor_request.go` | OK |
| `backend/internal/lstep/lstep_delivery_monitor_response.go` | OK |
| `backend/internal/lstep/lstep_delivery_monitor_service.go` | OK |
| `backend/internal/lstep/lstep_delivery_trigger_batch.go` | OK |
| `backend/internal/lstep/lstep_delivery_trigger_client.go` | OK |
| `backend/internal/lstep/lstep_delivery_trigger_log_repository.go` | OK |
| `backend/internal/lstep/lstep_delivery_trigger_methods.go` | OK |
| `backend/internal/lstep/lstep_delivery_trigger_service.go` | OK |
| `backend/internal/lstep/lstep_delivery_trigger_state.go` | OK |
| `backend/internal/lstep/lstep_delivery_trigger_suppression.go` | OK |
| `backend/internal/lstep/lstep_friend_attribute_snapshot_repository.go` | OK |
| `backend/internal/lstep/lstep_health_codes.go` | OK |
| `backend/internal/lstep/lstep_health_tag_sync.go` | OK |
| `backend/internal/lstep/lstep_health_tag_sync_batch.go` | OK |
| `backend/internal/lstep/lstep_health_tag_sync_checkup.go` | OK |
| `backend/internal/lstep/lstep_health_tag_sync_food.go` | OK |
| `backend/internal/lstep/lstep_health_tag_sync_prevention.go` | OK |
| `backend/internal/lstep/lstep_health_tag_sync_vaccine.go` | OK |
| `backend/internal/lstep/lstep_lifecycle_deps.go` | OK |
| `backend/internal/lstep/lstep_lifecycle_handler.go` | OK |
| `backend/internal/lstep/lstep_lifecycle_request.go` | OK |
| `backend/internal/lstep/lstep_lifecycle_service.go` | OK |
| `backend/internal/lstep/lstep_settings_connection.go` | OK |
| `backend/internal/lstep/lstep_settings_credentials.go` | OK |
| `backend/internal/lstep/lstep_settings_handler.go` | OK |
| `backend/internal/lstep/lstep_settings_repository.go` | OK |
| `backend/internal/lstep/lstep_settings_request.go` | OK |
| `backend/internal/lstep/lstep_settings_response.go` | OK |
| `backend/internal/lstep/lstep_settings_service.go` | OK |
| `backend/internal/lstep/lstep_settings_thresholds.go` | OK |
| `backend/internal/lstep/lstep_settings_update.go` | OK |
| `backend/internal/lstep/lstep_sync_error_counter_repository.go` | OK |
| `backend/internal/lstep/lstep_sync_settings_repository.go` | OK |
| `backend/internal/lstep/lstep_tag_cache_repository.go` | OK |
| `backend/internal/lstep/lstep_tag_code_mapping_handler.go` | OK |
| `backend/internal/lstep/lstep_tag_code_mapping_repository.go` | OK |
| `backend/internal/lstep/lstep_tag_code_mapping_request.go` | OK |
| `backend/internal/lstep/lstep_tag_code_mapping_response.go` | OK |
| `backend/internal/lstep/lstep_tag_code_mapping_service.go` | OK |
| `backend/internal/lstep/lstep_tag_config_handler.go` | OK |
| `backend/internal/lstep/lstep_tag_config_repository.go` | OK |
| `backend/internal/lstep/lstep_tag_config_request.go` | OK |
| `backend/internal/lstep/lstep_tag_config_response.go` | OK |
| `backend/internal/lstep/lstep_tag_config_service.go` | OK |
| `backend/internal/lstep/lstep_tag_handler.go` | OK |
| `backend/internal/lstep/lstep_tag_request.go` | OK |
| `backend/internal/lstep/lstep_tag_response.go` | OK |
| `backend/internal/lstep/lstep_tag_service.go` | OK |
| `backend/internal/lstep/lstep_tag_summary_handler.go` | OK |
| `backend/internal/lstep/lstep_tag_summary_request.go` | OK |
| `backend/internal/lstep/lstep_tag_summary_response.go` | OK |
| `backend/internal/lstep/lstep_tag_summary_service.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_api.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_care.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_care_checkup.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_care_chronic.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_care_resync.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_pet.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_pet_basic.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_pet_exclusion.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_service.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_shared.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_vaccine.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_visit.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_visit_cpm.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_visit_dormant.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_visit_ltv.go` | OK |
| `backend/internal/lstep/lstep_tag_sync_visit_next.go` | OK |
| `backend/internal/lstep/lstep_trigger_priority_handler.go` | OK |
| `backend/internal/lstep/lstep_trigger_priority_repository.go` | OK |
| `backend/internal/lstep/lstep_trigger_priority_request.go` | OK |
| `backend/internal/lstep/lstep_trigger_priority_service.go` | OK |
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
| `backend/internal/manualarticle/handler.go` | OK |
| `backend/internal/manualarticle/repository.go` | OK |
| `backend/internal/manualarticle/request.go` | OK |
| `backend/internal/manualarticle/response.go` | OK |
| `backend/internal/manualarticle/service.go` | OK |

### `backend/internal/medicalrecord` (239)

| path | status |
|---|---|
| `backend/internal/medicalrecord/audit_diff.go` | OK |
| `backend/internal/medicalrecord/cage_handler.go` | OK |
| `backend/internal/medicalrecord/cage_repository.go` | OK |
| `backend/internal/medicalrecord/cage_request.go` | OK |
| `backend/internal/medicalrecord/cage_response.go` | OK |
| `backend/internal/medicalrecord/cage_service.go` | OK |
| `backend/internal/medicalrecord/care_plan_item_handler.go` | OK |
| `backend/internal/medicalrecord/care_plan_item_repository.go` | OK |
| `backend/internal/medicalrecord/care_plan_item_request.go` | OK |
| `backend/internal/medicalrecord/care_plan_item_response.go` | OK |
| `backend/internal/medicalrecord/care_plan_item_service.go` | OK |
| `backend/internal/medicalrecord/checkup_field_handler.go` | OK |
| `backend/internal/medicalrecord/checkup_field_repository.go` | OK |
| `backend/internal/medicalrecord/checkup_field_request.go` | OK |
| `backend/internal/medicalrecord/checkup_field_response.go` | OK |
| `backend/internal/medicalrecord/checkup_field_result_service.go` | OK |
| `backend/internal/medicalrecord/checkup_handler.go` | OK |
| `backend/internal/medicalrecord/checkup_package_import_apply.go` | OK |
| `backend/internal/medicalrecord/checkup_package_import_apply_tx.go` | OK |
| `backend/internal/medicalrecord/checkup_package_import_handler.go` | OK |
| `backend/internal/medicalrecord/checkup_package_import_service.go` | OK |
| `backend/internal/medicalrecord/checkup_package_manifest.go` | OK |
| `backend/internal/medicalrecord/checkup_repository.go` | OK |
| `backend/internal/medicalrecord/checkup_request.go` | OK |
| `backend/internal/medicalrecord/checkup_response.go` | OK |
| `backend/internal/medicalrecord/checkup_service.go` | OK |
| `backend/internal/medicalrecord/checkup_type_handler.go` | OK |
| `backend/internal/medicalrecord/checkup_type_repository.go` | OK |
| `backend/internal/medicalrecord/checkup_type_request.go` | OK |
| `backend/internal/medicalrecord/checkup_type_response.go` | OK |
| `backend/internal/medicalrecord/checkup_type_service.go` | OK |
| `backend/internal/medicalrecord/chief_complaint_handler.go` | OK |
| `backend/internal/medicalrecord/chief_complaint_repository.go` | OK |
| `backend/internal/medicalrecord/chief_complaint_request.go` | OK |
| `backend/internal/medicalrecord/chief_complaint_response.go` | OK |
| `backend/internal/medicalrecord/chief_complaint_service.go` | OK |
| `backend/internal/medicalrecord/clinical_plan_handler.go` | OK |
| `backend/internal/medicalrecord/clinical_plan_repository.go` | OK |
| `backend/internal/medicalrecord/clinical_plan_request.go` | OK |
| `backend/internal/medicalrecord/clinical_plan_response.go` | OK |
| `backend/internal/medicalrecord/clinical_plan_service.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/clinical_relation_validation.go` | OK |
| `backend/internal/medicalrecord/consultation_handler.go` | OK |
| `backend/internal/medicalrecord/consultation_repository.go` | OK |
| `backend/internal/medicalrecord/consultation_request.go` | OK |
| `backend/internal/medicalrecord/consultation_response.go` | OK |
| `backend/internal/medicalrecord/consultation_service.go` | OK |
| `backend/internal/medicalrecord/daily_record_handler.go` | OK |
| `backend/internal/medicalrecord/daily_record_repository.go` | OK |
| `backend/internal/medicalrecord/daily_record_request.go` | OK |
| `backend/internal/medicalrecord/daily_record_response.go` | OK |
| `backend/internal/medicalrecord/daily_record_service.go` | OK |
| `backend/internal/medicalrecord/diagnosis_handler.go` | OK |
| `backend/internal/medicalrecord/diagnosis_name_repository.go` | OK |
| `backend/internal/medicalrecord/diagnosis_request.go` | OK |
| `backend/internal/medicalrecord/diagnosis_response.go` | OK |
| `backend/internal/medicalrecord/diagnosis_service.go` | OK |
| `backend/internal/medicalrecord/diagnosis_type_repository.go` | OK |
| `backend/internal/medicalrecord/discount_permission.go` | OK |
| `backend/internal/medicalrecord/dose_calc.go` | OK |
| `backend/internal/medicalrecord/dose_revalidation.go` | FINDING(BE-RC-005) |
| `backend/internal/medicalrecord/dose_validators.go` | OK |
| `backend/internal/medicalrecord/exam_reference_range_repository.go` | OK |
| `backend/internal/medicalrecord/exam_result_assessment.go` | OK |
| `backend/internal/medicalrecord/exam_type_field.go` | OK |
| `backend/internal/medicalrecord/exam_type_handler.go` | OK |
| `backend/internal/medicalrecord/exam_type_repository.go` | OK |
| `backend/internal/medicalrecord/exam_type_request.go` | OK |
| `backend/internal/medicalrecord/exam_type_response.go` | OK |
| `backend/internal/medicalrecord/exam_type_service.go` | OK |
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
| `backend/internal/medicalrecord/hospitalization_discharge.go` | OK |
| `backend/internal/medicalrecord/hospitalization_discharge_tx.go` | OK |
| `backend/internal/medicalrecord/hospitalization_handler.go` | OK |
| `backend/internal/medicalrecord/hospitalization_plan_handler.go` | OK |
| `backend/internal/medicalrecord/hospitalization_plan_repository.go` | OK |
| `backend/internal/medicalrecord/hospitalization_plan_request.go` | OK |
| `backend/internal/medicalrecord/hospitalization_plan_response.go` | OK |
| `backend/internal/medicalrecord/hospitalization_plan_service.go` | OK |
| `backend/internal/medicalrecord/hospitalization_repository.go` | OK |
| `backend/internal/medicalrecord/hospitalization_request.go` | OK |
| `backend/internal/medicalrecord/hospitalization_response.go` | OK |
| `backend/internal/medicalrecord/hospitalization_service.go` | OK |
| `backend/internal/medicalrecord/hospitalization_service_create.go` | OK |
| `backend/internal/medicalrecord/inquiry_handler.go` | OK |
| `backend/internal/medicalrecord/inquiry_repository.go` | OK |
| `backend/internal/medicalrecord/inquiry_request.go` | OK |
| `backend/internal/medicalrecord/inquiry_response.go` | OK |
| `backend/internal/medicalrecord/inquiry_service.go` | OK |
| `backend/internal/medicalrecord/inquiry_template_handler.go` | OK |
| `backend/internal/medicalrecord/inquiry_template_repository.go` | OK |
| `backend/internal/medicalrecord/inquiry_template_request.go` | OK |
| `backend/internal/medicalrecord/inquiry_template_response.go` | OK |
| `backend/internal/medicalrecord/inquiry_template_service.go` | OK |
| `backend/internal/medicalrecord/lab_audit_logger.go` | OK |
| `backend/internal/medicalrecord/lab_device_agent_consumer_handler.go` | OK |
| `backend/internal/medicalrecord/lab_device_decode.go` | OK |
| `backend/internal/medicalrecord/lab_device_exam_persist.go` | OK |
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
| `backend/internal/medicalrecord/lab_device_receive_service.go` | FINDING(BE-RC-019) |
| `backend/internal/medicalrecord/lab_device_today_visit.go` | OK |
| `backend/internal/medicalrecord/lab_device_urine.go` | OK |
| `backend/internal/medicalrecord/lab_import_examination_service.go` | OK |
| `backend/internal/medicalrecord/lab_import_examination_write.go` | OK |
| `backend/internal/medicalrecord/lab_import_handler.go` | OK |
| `backend/internal/medicalrecord/lab_import_repository.go` | OK |
| `backend/internal/medicalrecord/lab_import_request.go` | OK |
| `backend/internal/medicalrecord/lab_import_response.go` | OK |
| `backend/internal/medicalrecord/lab_import_revert_service.go` | OK |
| `backend/internal/medicalrecord/lab_import_revert_tx.go` | OK |
| `backend/internal/medicalrecord/lab_import_service.go` | OK |
| `backend/internal/medicalrecord/lab_import_usage_tracker.go` | OK |
| `backend/internal/medicalrecord/lab_report_handler.go` | OK |
| `backend/internal/medicalrecord/lab_report_query_service.go` | OK |
| `backend/internal/medicalrecord/lab_report_response.go` | OK |
| `backend/internal/medicalrecord/lab_result_import_service.go` | OK |
| `backend/internal/medicalrecord/master_validators.go` | OK |
| `backend/internal/medicalrecord/medical_record_addendum_handler.go` | OK |
| `backend/internal/medicalrecord/medical_record_addendum_repository.go` | OK |
| `backend/internal/medicalrecord/medical_record_addendum_request.go` | OK |
| `backend/internal/medicalrecord/medical_record_addendum_response.go` | OK |
| `backend/internal/medicalrecord/medical_record_addendum_service.go` | OK |
| `backend/internal/medicalrecord/medical_record_appointment_context.go` | OK |
| `backend/internal/medicalrecord/medical_record_auto_create.go` | OK |
| `backend/internal/medicalrecord/medical_record_builders.go` | OK |
| `backend/internal/medicalrecord/medical_record_crud.go` | OK |
| `backend/internal/medicalrecord/medical_record_crud_update.go` | OK |
| `backend/internal/medicalrecord/medical_record_delete_conflict.go` | OK |
| `backend/internal/medicalrecord/medical_record_handler.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_handler.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_repository.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_request.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_response.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_service.go` | OK |
| `backend/internal/medicalrecord/medical_record_image_upload_quota.go` | OK |
| `backend/internal/medicalrecord/medical_record_lock.go` | OK |
| `backend/internal/medicalrecord/medical_record_lstep_sync.go` | OK |
| `backend/internal/medicalrecord/medical_record_owner_visit_repository.go` | OK |
| `backend/internal/medicalrecord/medical_record_repository.go` | FINDING(BE-RC-019) |
| `backend/internal/medicalrecord/medical_record_repository_list.go` | OK |
| `backend/internal/medicalrecord/medical_record_repository_list_search.go` | OK |
| `backend/internal/medicalrecord/medical_record_request.go` | OK |
| `backend/internal/medicalrecord/medical_record_response.go` | OK |
| `backend/internal/medicalrecord/medical_record_service.go` | OK |
| `backend/internal/medicalrecord/medical_record_subrecords.go` | OK |
| `backend/internal/medicalrecord/medicine_dose_param_handler.go` | OK |
| `backend/internal/medicalrecord/medicine_dose_param_repository.go` | OK |
| `backend/internal/medicalrecord/medicine_dose_param_request.go` | OK |
| `backend/internal/medicalrecord/medicine_dose_param_response.go` | OK |
| `backend/internal/medicalrecord/medicine_dose_param_service.go` | OK |
| `backend/internal/medicalrecord/medicine_handler.go` | OK |
| `backend/internal/medicalrecord/medicine_repository.go` | OK |
| `backend/internal/medicalrecord/medicine_request.go` | OK |
| `backend/internal/medicalrecord/medicine_response.go` | OK |
| `backend/internal/medicalrecord/medicine_service.go` | OK |
| `backend/internal/medicalrecord/medicine_service_create.go` | OK |
| `backend/internal/medicalrecord/medicine_service_delete.go` | OK |
| `backend/internal/medicalrecord/nested_summary_response.go` | OK |
| `backend/internal/medicalrecord/pagination.go` | OK |
| `backend/internal/medicalrecord/prescription_handler.go` | OK |
| `backend/internal/medicalrecord/prescription_repository.go` | OK |
| `backend/internal/medicalrecord/prescription_request.go` | OK |
| `backend/internal/medicalrecord/prescription_response.go` | OK |
| `backend/internal/medicalrecord/prescription_service.go` | OK |
| `backend/internal/medicalrecord/procedure_handler.go` | OK |
| `backend/internal/medicalrecord/procedure_repository.go` | OK |
| `backend/internal/medicalrecord/procedure_request.go` | OK |
| `backend/internal/medicalrecord/procedure_response.go` | OK |
| `backend/internal/medicalrecord/procedure_service.go` | OK |
| `backend/internal/medicalrecord/query_filter_helpers.go` | OK |
| `backend/internal/medicalrecord/replace_audit_tail.go` | OK |
| `backend/internal/medicalrecord/routes.go` | OK |
| `backend/internal/medicalrecord/routes_hospitalization.go` | OK |
| `backend/internal/medicalrecord/routes_lab.go` | OK |
| `backend/internal/medicalrecord/routes_masters.go` | OK |
| `backend/internal/medicalrecord/routes_records.go` | OK |
| `backend/internal/medicalrecord/service_deps.go` | OK |
| `backend/internal/medicalrecord/to_service_input_error.go` | OK |
| `backend/internal/medicalrecord/treatment_dose_save.go` | OK |
| `backend/internal/medicalrecord/treatment_fields.go` | OK |
| `backend/internal/medicalrecord/treatment_handler.go` | OK |
| `backend/internal/medicalrecord/treatment_master_fk.go` | OK |
| `backend/internal/medicalrecord/treatment_plan_handler.go` | OK |
| `backend/internal/medicalrecord/treatment_plan_repository.go` | OK |
| `backend/internal/medicalrecord/treatment_plan_request.go` | OK |
| `backend/internal/medicalrecord/treatment_plan_response.go` | OK |
| `backend/internal/medicalrecord/treatment_plan_service.go` | OK |
| `backend/internal/medicalrecord/treatment_repository.go` | OK |
| `backend/internal/medicalrecord/treatment_request.go` | OK |
| `backend/internal/medicalrecord/treatment_response.go` | OK |
| `backend/internal/medicalrecord/treatment_service.go` | OK |
| `backend/internal/medicalrecord/treatment_service_tx.go` | OK |
| `backend/internal/medicalrecord/vaccination_handler.go` | OK |
| `backend/internal/medicalrecord/vaccination_repository.go` | OK |
| `backend/internal/medicalrecord/vaccination_request.go` | OK |
| `backend/internal/medicalrecord/vaccination_response.go` | OK |
| `backend/internal/medicalrecord/vaccination_service.go` | OK |
| `backend/internal/medicalrecord/vaccine_handler.go` | OK |
| `backend/internal/medicalrecord/vaccine_repository.go` | OK |
| `backend/internal/medicalrecord/vaccine_request.go` | OK |
| `backend/internal/medicalrecord/vaccine_response.go` | OK |
| `backend/internal/medicalrecord/vaccine_service.go` | OK |
| `backend/internal/medicalrecord/validators.go` | OK |
| `backend/internal/medicalrecord/validators_accounting.go` | OK |
| `backend/internal/medicalrecord/validators_master.go` | OK |
| `backend/internal/medicalrecord/vital_audit.go` | OK |
| `backend/internal/medicalrecord/vital_handler.go` | OK |
| `backend/internal/medicalrecord/vital_repository.go` | OK |
| `backend/internal/medicalrecord/vital_request.go` | OK |
| `backend/internal/medicalrecord/vital_response.go` | OK |
| `backend/internal/medicalrecord/vital_service.go` | OK |
| `backend/internal/medicalrecord/vital_service_create.go` | OK |
| `backend/internal/medicalrecord/vital_service_update.go` | OK |
| `backend/internal/medicalrecord/vital_validation.go` | OK |

### `backend/internal/middleware` (9)

| path | status |
|---|---|
| `backend/internal/middleware/auth.go` | OK |
| `backend/internal/middleware/cors.go` | OK |
| `backend/internal/middleware/csrf.go` | OK |
| `backend/internal/middleware/liff_auth.go` | OK |
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
| `backend/internal/model/line_link_token.go` | OK |
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
| `backend/internal/owner/http_request.go` | OK |
| `backend/internal/owner/http_response.go` | OK |
| `backend/internal/owner/http_routes.go` | OK |
| `backend/internal/owner/ltv_repository.go` | OK |
| `backend/internal/owner/ltv_repository_query.go` | OK |
| `backend/internal/owner/mapper.go` | OK |
| `backend/internal/owner/pet_registration.go` | OK |
| `backend/internal/owner/repository.go` | OK |
| `backend/internal/owner/service.go` | OK |
| `backend/internal/owner/service_builders.go` | OK |
| `backend/internal/owner/service_core.go` | OK |
| `backend/internal/owner/service_delivery.go` | OK |
| `backend/internal/owner/service_line.go` | OK |
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
| `backend/internal/pet/animal_species_repository.go` | OK |
| `backend/internal/pet/animal_species_request.go` | OK |
| `backend/internal/pet/animal_species_response.go` | OK |
| `backend/internal/pet/animal_species_service.go` | OK |
| `backend/internal/pet/chronic_condition_handler.go` | OK |
| `backend/internal/pet/chronic_condition_repository.go` | OK |
| `backend/internal/pet/chronic_condition_request.go` | OK |
| `backend/internal/pet/chronic_condition_service.go` | OK |
| `backend/internal/pet/date.go` | OK |
| `backend/internal/pet/handler.go` | OK |
| `backend/internal/pet/mapper.go` | OK |
| `backend/internal/pet/owner_registration.go` | OK |
| `backend/internal/pet/owner_registration_adapter.go` | OK |
| `backend/internal/pet/pet_handler.go` | OK |
| `backend/internal/pet/pet_owner_handler.go` | OK |
| `backend/internal/pet/pet_owner_repository.go` | OK |
| `backend/internal/pet/pet_owner_request.go` | OK |
| `backend/internal/pet/pet_owner_response.go` | OK |
| `backend/internal/pet/pet_owner_service.go` | OK |
| `backend/internal/pet/pet_request.go` | OK |
| `backend/internal/pet/pet_response.go` | OK |
| `backend/internal/pet/ports.go` | OK |
| `backend/internal/pet/repository.go` | OK |
| `backend/internal/pet/routes.go` | OK |
| `backend/internal/pet/service.go` | OK |
| `backend/internal/pet/validators.go` | OK |

### `backend/internal/reservation` (84)

| path | status |
|---|---|
| `backend/internal/reservation/appointment_admin_handler.go` | OK |
| `backend/internal/reservation/appointment_admin_repository.go` | OK |
| `backend/internal/reservation/appointment_admin_request.go` | OK |
| `backend/internal/reservation/appointment_admin_response.go` | OK |
| `backend/internal/reservation/appointment_admin_service.go` | OK |
| `backend/internal/reservation/appointment_notification_service.go` | OK |
| `backend/internal/reservation/availability_slot_merge.go` | OK |
| `backend/internal/reservation/available_dates.go` | OK |
| `backend/internal/reservation/doc.go` | OK |
| `backend/internal/reservation/liff_handler.go` | OK |
| `backend/internal/reservation/liff_request.go` | OK |
| `backend/internal/reservation/liff_response.go` | OK |
| `backend/internal/reservation/liff_service.go` | OK |
| `backend/internal/reservation/liff_service_availability.go` | OK |
| `backend/internal/reservation/liff_service_availability_business.go` | OK |
| `backend/internal/reservation/liff_service_availability_delegate.go` | OK |
| `backend/internal/reservation/liff_service_availability_filters.go` | OK |
| `backend/internal/reservation/liff_service_availability_slots.go` | OK |
| `backend/internal/reservation/liff_service_availability_staff.go` | OK |
| `backend/internal/reservation/liff_service_availability_time.go` | OK |
| `backend/internal/reservation/liff_service_catalog.go` | OK |
| `backend/internal/reservation/liff_service_health_card.go` | OK |
| `backend/internal/reservation/liff_service_reservations.go` | OK |
| `backend/internal/reservation/liff_validation.go` | OK |
| `backend/internal/reservation/line_reservation_setting_handler.go` | OK |
| `backend/internal/reservation/line_reservation_setting_repository.go` | OK |
| `backend/internal/reservation/line_reservation_setting_request.go` | OK |
| `backend/internal/reservation/line_reservation_setting_response.go` | OK |
| `backend/internal/reservation/line_reservation_setting_service.go` | OK |
| `backend/internal/reservation/nested_summary_response.go` | OK |
| `backend/internal/reservation/reservation_capacity.go` | OK |
| `backend/internal/reservation/reservation_handler.go` | OK |
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
| `backend/internal/reservation/reservation_schedule_service.go` | OK |
| `backend/internal/reservation/reservation_service.go` | OK |
| `backend/internal/reservation/reservation_service_update.go` | OK |
| `backend/internal/reservation/reservation_service_validate.go` | OK |
| `backend/internal/reservation/reservation_staff_capability_validator.go` | OK |
| `backend/internal/reservation/reservation_staff_handler.go` | OK |
| `backend/internal/reservation/reservation_staff_repository.go` | OK |
| `backend/internal/reservation/reservation_staff_request.go` | OK |
| `backend/internal/reservation/reservation_staff_response.go` | OK |
| `backend/internal/reservation/reservation_staff_service.go` | OK |
| `backend/internal/reservation/reservation_type_availability_validator.go` | OK |
| `backend/internal/reservation/reservation_type_available_slot_repository.go` | OK |
| `backend/internal/reservation/reservation_type_group_handler.go` | OK |
| `backend/internal/reservation/reservation_type_group_repository.go` | OK |
| `backend/internal/reservation/reservation_type_group_request.go` | OK |
| `backend/internal/reservation/reservation_type_group_response.go` | OK |
| `backend/internal/reservation/reservation_type_group_service.go` | OK |
| `backend/internal/reservation/reservation_type_handler.go` | OK |
| `backend/internal/reservation/reservation_type_liff_handler.go` | OK |
| `backend/internal/reservation/reservation_type_liff_repository.go` | OK |
| `backend/internal/reservation/reservation_type_liff_request.go` | OK |
| `backend/internal/reservation/reservation_type_liff_response.go` | OK |
| `backend/internal/reservation/reservation_type_liff_service.go` | OK |
| `backend/internal/reservation/reservation_type_occupation_repository.go` | OK |
| `backend/internal/reservation/reservation_type_repository.go` | OK |
| `backend/internal/reservation/reservation_type_request.go` | OK |
| `backend/internal/reservation/reservation_type_response.go` | OK |
| `backend/internal/reservation/reservation_type_service.go` | OK |
| `backend/internal/reservation/reservation_type_service_available_slot.go` | OK |
| `backend/internal/reservation/reservation_type_service_builders.go` | OK |
| `backend/internal/reservation/reservation_type_service_core.go` | OK |
| `backend/internal/reservation/reservation_type_service_occupation.go` | OK |
| `backend/internal/reservation/reservation_type_service_unavailable.go` | OK |
| `backend/internal/reservation/reservation_type_unavailable_time_repository.go` | OK |
| `backend/internal/reservation/reservation_validators.go` | OK |
| `backend/internal/reservation/reservation_validators_create.go` | OK |
| `backend/internal/reservation/response_error.go` | OK |
| `backend/internal/reservation/routes.go` | OK |
| `backend/internal/reservation/service_deps.go` | OK |
| `backend/internal/reservation/staff_affinity.go` | OK |
| `backend/internal/reservation/timeslot_engine.go` | OK |

### `backend/internal/scheduler` (1)

| path | status |
|---|---|
| `backend/internal/scheduler/handler.go` | OK |

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
| `backend/internal/sharedkernel/medical_record_lock.go` | OK |
| `backend/internal/sharedkernel/owner_pet_link.go` | OK |
| `backend/internal/sharedkernel/pet_not_deceased.go` | OK |
| `backend/internal/sharedkernel/shift_times.go` | OK |
| `backend/internal/sharedkernel/validators.go` | OK |

### `backend/internal/staff` (38)

| path | status |
|---|---|
| `backend/internal/staff/credential_audit.go` | OK |
| `backend/internal/staff/handler.go` | OK |
| `backend/internal/staff/http_binding.go` | OK |
| `backend/internal/staff/occupation_handler.go` | OK |
| `backend/internal/staff/occupation_repository.go` | OK |
| `backend/internal/staff/occupation_request.go` | OK |
| `backend/internal/staff/occupation_response.go` | OK |
| `backend/internal/staff/occupation_service.go` | OK |
| `backend/internal/staff/permission_assignment_audit.go` | OK |
| `backend/internal/staff/ports.go` | OK |
| `backend/internal/staff/reservation_staff_update.go` | OK |
| `backend/internal/staff/shift_entry_repository.go` | OK |
| `backend/internal/staff/shift_entry_service.go` | OK |
| `backend/internal/staff/shift_handler.go` | OK |
| `backend/internal/staff/shift_request.go` | OK |
| `backend/internal/staff/shift_response.go` | OK |
| `backend/internal/staff/shift_template_handler.go` | OK |
| `backend/internal/staff/shift_template_repository.go` | OK |
| `backend/internal/staff/shift_template_request.go` | OK |
| `backend/internal/staff/shift_template_response.go` | OK |
| `backend/internal/staff/shift_template_service.go` | OK |
| `backend/internal/staff/staff_clinic_assignment_repository.go` | OK |
| `backend/internal/staff/staff_clinic_assignment_service.go` | OK |
| `backend/internal/staff/staff_handler.go` | OK |
| `backend/internal/staff/staff_provisioning.go` | OK |
| `backend/internal/staff/staff_provisioning_apply.go` | OK |
| `backend/internal/staff/staff_provisioning_repository.go` | OK |
| `backend/internal/staff/staff_provisioning_validate.go` | OK |
| `backend/internal/staff/staff_repository.go` | OK |
| `backend/internal/staff/staff_request.go` | OK |
| `backend/internal/staff/staff_response.go` | OK |
| `backend/internal/staff/staff_service.go` | OK |
| `backend/internal/staff/staff_service_account.go` | OK |
| `backend/internal/staff/staff_service_builders.go` | OK |
| `backend/internal/staff/staff_service_core.go` | OK |
| `backend/internal/staff/staff_service_permissions.go` | OK |
| `backend/internal/staff/staff_service_update.go` | OK |
| `backend/internal/staff/validators.go` | OK |

### `backend/internal/testdb` (3)

| path | status |
|---|---|
| `backend/internal/testdb/fixtures.go` | SKIP(test kernel) |
| `backend/internal/testdb/testdb.go` | SKIP(test kernel) |
| `backend/internal/testdb/truncate.go` | SKIP(test kernel; Truncate advisory lock + deadlock retry) |

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
| `backend/internal/trimming/trimming_course_repository.go` | OK |
| `backend/internal/trimming/trimming_course_request.go` | OK |
| `backend/internal/trimming/trimming_course_response.go` | OK |
| `backend/internal/trimming/trimming_course_service.go` | OK |
| `backend/internal/trimming/trimming_course_type_handler.go` | OK |
| `backend/internal/trimming/trimming_course_type_repository.go` | OK |
| `backend/internal/trimming/trimming_course_type_request.go` | OK |
| `backend/internal/trimming/trimming_course_type_response.go` | OK |
| `backend/internal/trimming/trimming_course_type_service.go` | OK |
| `backend/internal/trimming/trimming_handler.go` | OK |
| `backend/internal/trimming/trimming_option_handler.go` | OK |
| `backend/internal/trimming/trimming_option_repository.go` | OK |
| `backend/internal/trimming/trimming_option_request.go` | OK |
| `backend/internal/trimming/trimming_option_response.go` | OK |
| `backend/internal/trimming/trimming_option_service.go` | OK |
| `backend/internal/trimming/trimming_repository.go` | OK |
| `backend/internal/trimming/trimming_request.go` | OK |
| `backend/internal/trimming/trimming_response.go` | OK |
| `backend/internal/trimming/trimming_service.go` | OK |
| `backend/internal/trimming/trimming_service_create.go` | OK |
| `backend/internal/trimming/trimming_service_mutate.go` | OK |
| `backend/internal/trimming/trimming_service_update.go` | OK |
| `backend/internal/trimming/trimming_service_validate.go` | OK |
| `backend/internal/trimming/validators.go` | OK |

