# backend コード規約チェック結果（2026-09-03）

`backend/` 配下（`internal/`・`cmd/`、本番 `.go` を主対象。`*_test.go` は本番の負債を証明する場合のみ。`cmd/_archive` 除外）を以下の規約に照合した結果。

- 規約正本: `backend/CLAUDE.md`、`backend/CODING_RULES.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/go-gin-backend-review.md`、`.claude/refs/backend-application-invariants.md`、`.claude/refs/error-handling.md`、`.claude/refs/naming-conventions.md`、`.claude/refs/go-language.md`、ADR-006
- 方法: (1) 規約項目ごとの機械検出（rg / パッケージ一覧 / lintscan 照合）を全域に実施、(2) 領域別に精読（package 境界・HTTP/Context/error・persistence/security・命名/test の 4 系統）、(3) HIGH 所見は実コードで再検証済み
- 各所見の `path` は `backend/` 起点。行番号は調査時点のもの
- 旧 BE9 台帳（2026-07-24 退役の `BE-refactor.md`）は正本として継承しない。本ファイルは現行ツリーの規約監査である
- Handler → Service → Repository / Clean Architecture は Go/Gin 公式要件ではないため、再導入を推奨しない

---

## 0. サマリー

| 重要度 | 件数 | 概要 |
|---|---|---|
| CRITICAL | 0 | 生きた clinic isolation 漏洩・認可バイパスは未検出 |
| HIGH | 3 | ドメイン境界の generic `map` 更新（1）、HTTP 境界の長さ上限欠落（1 クラスタ）、非原子 Count→Delete（1 クラスタ・既知 residual） |
| MEDIUM | 11 | tx 内 Count→Delete 残差、重複ログ、`err.Error()` 載せ直し、commit 後 reload、広い exported port、lint 無効化 など |
| LOW | 9 | stutter、旧 layer パッケージコメント、GoDoc、巨大 package 凝集 など |

**機械検出で 0 件／ゲート済みを確認した規約（合格）**: 旧 `internal/handler|service|repository` production 0／domain 内 layer 名 subpackage 0／production `util|common|misc` 0／ADR-006 domain import DAG 違反 0／`appointments` owner 外 write 0（AST gate あり）／`staffs`/`shift_entries` の owner 外直接 GORM write 0／本番 `BindJSON`/`MustBind*` 0（`ShouldBind*` + エラー処理）／request-scoped Context の struct 保持 0／`go func` による生 `*gin.Context` 0／package global DI 0／startup `AutoMigrate` 0／CORS `*` + credentials 0／本番 800 行超ファイル 0／receiver `this`/`self` 0／本番 dot import 0／Go 識別子 `Id` 0（`ID` 一貫）。

**既知 residual（CODING_RULES が「一括 retrofit は別作業」と明記）**: master「使用中は削除不可」の非原子 Count→Delete。本監査では現状の残件を列挙する。新規 production と当該 Delete を触る変更では条件付き原子 DELETE に寄せること。

**推奨着手順**: ① BE-RC-001（staff generic update の typed intent 化）→ ② BE-RC-002（高リスク string の `max`）→ ③ BE-RC-003（inventory / 支払方法 / 保険の原子 DELETE）→ ④ BE-RC-006（bind/CSV の固定 message）→ ⑤ BE-RC-005（service の重複 `slog.ErrorContext` を新規から止める）→ ⑥ 触る master Delete で BE-RC-004 を払う。

---

## 1. HIGH

### ドメイン境界

#### BE-RC-001 [HIGH] staff write owner が reservation へ generic `map[string]any` 更新 API を公開している
- 対象:
  - `internal/staff/staff_repository.go:47-48,534-536` — `UpdateForReservation(ctx, clinicID, id, fields map[string]any)` が `fields` をそのまま `Updates`
  - `internal/reservation/reservation_staff_repository.go:24,46-48,159-160` — consumer-side `staffsWriter` が同じ map API を宣言し delegate
- 規約: ADR-006 / CODING_RULES「owner 外へ任意 field を変更できる generic update API を公開しない」。`appointments` は `CompleteForAccounting` / `UpdateForTrimming` 等の typed intent に収束済み
- 現状: HTTP/use case 側は `UpdateReservationStaffInput` → `buildReservationStaffUpdate`（既知キーのみ、`reservation_staff_service.go:19-36,185`）で typed。しかし境界契約は任意 map のままなので、別 caller が任意列を渡せる
- 改善案: `staffsWriter` を typed command（既存 input 相当）にし、map 化は staff package の unexported primitive に閉じる。`appointments` と同じく owner 外へ `map[string]any` を出さない

### HTTP 境界の長さ検証

#### BE-RC-002 [HIGH] 高リスク string に境界 `max` がなく、DB / 外部 API 失敗に長さ検証を委ねている
- 規約: go-gin-backend-guidelines §7 / CODING_RULES HTTP「型・形式・**長さ**・範囲・列挙値を境界で検証」
- 正例: 見積 reject `reason` は `binding:"required,min=1,max=500"`（`billing/estimate_request.go:160`）。会計 confirmation の `return_reason` は 500 文字を handler 側で強制（`billing/billing_confirmation_request.go:13,58`）。飼主の kana/address/email/remarks は `max` あり（`owner/http_request.go:167-179`）
- 対象（優先）:
  | リスク | path | 現状 |
  |---|---|---|
  | LINE 本文・DoS | `internal/lstep/line_send_request.go:17-20` | `Text` / `FileName` / `Purpose` に `max` なし |
  | 締め後会計の理由 | `internal/billing/accounting_request.go:211,341`、`billing_item_request.go:58,71,77` | `post_close_reason` に `max` なし（confirmation は 500 文字強制済み） |
  | 返金理由 | `internal/billing/refund_request.go:15` | `Reason` に `max` なし |
  | クレジット訂正 | `internal/billing/accounting_request.go:297-298` | `Reason` は `required` のみ、`Memo` に `max` なし |
  | 会計メモ | `internal/billing/accounting_request.go:169,203,328` | `Memo` に `max` なし |
  | カルテ追記 | `internal/medicalrecord/medical_record_addendum_request.go:4-5` | `AfterText` / `Reason` が `required` のみ |
  | 入院記録 | `internal/medicalrecord/daily_record_request.go:17,45,66` | `Notes` / `Value` / `Content` に `max` なし |
  | 治療計画 | `internal/medicalrecord/treatment_plan_request.go:4-5` | `TreatmentContent` が `required` のみ、`Memo` に `max` なし |
  | lab revert | `internal/medicalrecord/lab_import_request.go:26` | `Reason` が `required` のみ |
  | 飼主氏名 | `internal/owner/http_request.go:166` | `OwnerName` が `required` のみ（kana は `max=100`） |
  | ペット名 | `internal/pet/pet_request.go:99` | `Name` が `required` のみ（`NameKana` は `max=100`） |
  | キャンペーン名 | `internal/billing/campaign_request.go:12` | `Name` が `required` のみ |
  | 会計明細名 | `internal/billing/accounting_request.go:175` | `Name` が `required` のみ（見積 item は `max=255`） |
  | 見積コメント | `internal/billing/estimate_request.go:59-60,127-128` | `Comment` / `Notes` に `max` なし（`Title` は `max=255`） |
  | LSTEP 理由 | `internal/lstep/lstep_lifecycle_request.go:29,34` | 死亡記録・opt-out の `Reason` に `max` なし |
- 改善案: 同系統の正例に揃える（reason 500、memo 1000、name 255、LINE text は Messaging API 上限）。service 側の再検査は任意、境界 tag を最終防壁にしない

### Persistence（既知 residual）

#### BE-RC-003 [HIGH][既知 residual] 非原子 Count→Delete（tx なし）が在庫・支払・保険に残る
- 規約: CODING_RULES「使用中は削除不可」の正しさは `clinic_id + id` と usage 不在を同一 SQL に束ねた条件付き原子 DELETE。Find→Count→Delete を正しさの根拠にしない。一括 retrofit は別作業、**触る Delete では直す**
- 正例: `internal/billing/estimate_repository.go` の `DeleteIfNotLocked`（早期 Count は UX のみ）
- 対象（再検証済み・tx なし）:
  - `internal/inventory/inventory_service.go:166-183` — Find → `CountUsageByInventoryID` → `Delete`
  - `internal/billing/payment_method_master_service.go:139-163` — システム行チェック後に Count → Delete
  - `internal/billing/insurance_service.go:154-171` — Count → Delete
- 改善案: `DeleteIfNotLocked` 型（`clinic_id + id` + `NOT EXISTS usage`、`RowsAffected==0` → Conflict/NotFound）。早期 Count は UX 用に残してよい

---

## 2. MEDIUM

### Persistence / 監査

#### BE-RC-004 [MEDIUM][既知 residual] その他 master Delete も条件付き原子 DELETE 未達
- 規約: BE-RC-003 と同じ。tx 内 Count→Delete は race 窓が縮小するだけで、規則が求める「同一 SQL」には未到達
- tx なし（臨床マスター含む）: `consultation_service.go`、`vaccine_service.go`、`checkup_type_service.go`、`diagnosis_service.go:346-366`（診断名）、`chief_complaint_service.go`、`hospitalization_plan_service.go`、`inquiry_template_service.go`、`reservation_type_group_service.go` ほか `CountUsage` 後に `Delete`
- tx 内だが条件付き DELETE ではない: merchandise（soft-delete 後に usage 再確認、`merchandise_item_service.go:217-241`）、cage（`FOR UPDATE`→Count→Delete）、procedure / exam_type / trimming course・option・course_type / occupation / animal_species / medicine / permission_group
- 改善案: 当該 Delete を変更する PR で原子 DELETE に置換。一括 sweep は別タスク

#### BE-RC-007 [MEDIUM] Update 成功後の別 statement `FindByID` が成功を失敗へ反転し得る
- 対象: `internal/billing/payment_method_master_repository.go:61-65` — `updateScopedByID` の後に `FindByID`。tx で括られていない
- 規約: invariants「commit 済みの成功を後段 read error で失敗応答へ反転させない」。正例: `owner/repository.go:34-36` の `UpdateAndFind`（reload 失敗で write を rollback）
- 対比: `reservation_staff_service.go:180-216` の reload は `WithTx` 内なので反転しない
- 改善案: 同一 tx で update+reload、`RETURNING`、または reload 失敗時は更新済みエンティティを返す

#### BE-RC-011 [MEDIUM] 見積 approve/reject 監査が commit 後 best-effort（CreateSuccessor は fail-closed）
- 対象: `internal/billing/estimate_service.go:250-251,279-303,471-508` — Create/Update/Delete および status が approved/rejected になる経路が `logEstimateChangeBestEffort`（commit 後、失敗はログのみ）
- 対比: 同ファイル `WithEstimateAuditTx` は後継ドラフト（CreateSuccessor）だけ fail-closed。締め後会計編集と no-show Mark は同一 tx 監査で準拠
- 規約: fail-closed と**定めた** clinical/financial 監査は同一 tx。明示対象の正例は締め後会計であり、見積 approve を含めるかは方針確認が必要。コードコメントは medical_record 同型の意図的 best-effort
- 改善案: approve/reject を CreateSuccessor と同じ `auditTx` 参加にするか、best-effort を ADR/CODING_RULES に「見積の通常 CRUD は対象外」と明記して不整合を消す

### Error / logging

#### BE-RC-005 [MEDIUM] service が `slog.ErrorContext` したうえで handler `RespondError` が 5xx を `c.Error` し、middleware が再ログする
- 規約: guidelines §8 / error-handling.md「同じ error を複数層で重複ログしない。十分な文脈を持つ境界で 1 回」。未知 pg コードだけ request 境界での `c.Error` 重複を例外許容
- 現状: `*_service.go` の `slog.ErrorContext` が広範（billing masters、lstep lifecycle、reservation staff、clinic など）。例: `internal/billing/insurance_service.go` の list/update/delete、`internal/inventory/inventory_service.go:150-160,172-179`
- 5xx は `httpapi/response.go:17-19` が `c.Error(err)` → request logging middleware が再記録
- 改善案: 既知 4xx は service でログしない（return のみ）。5xx は middleware 一本化。新規コードからこの型を増やさない。一括削除はノイズ削減タスク

#### BE-RC-006 [MEDIUM] `WrapInvalidInput(err.Error())` と CSV `+ err.Error()` が内部 message を client へ載せ得る
- 規約: error-handling.md「user message と内部診断を分離」。正例: `httpapi.ParseBindError`（BUG-129 で struct/型名漏洩を遮断、`bind_errors.go:14-34`）。ShouldBind 経路の大半はこれを使う
- 対象:
  - CSV: `internal/lstep/lstep_csv_import_prepare.go:29,44`（`"failed to parse CSV: " + err.Error()`）、同 `lstep_csv_import_processing.go`
  - `toServiceInput` 失敗を `err.Error()` で包む handler 群: `reservation_handler.go:74,160,204,247`、`hospitalization_handler.go:116,147`、`daily_record_handler.go:144,184,224`、`inventory_handler.go:70,102`、`cash_register_handler.go:67`、`liff_handler.go:196`、`chronic_condition_handler.go:90,127`、`checkup_sync_handler.go:63` ほか
- 現状: 多くは日付パース等の固定英語 message で実害は小さい。CSV と encoding/json 由来 error をそのまま載せると内部詳細が漏れ得る。`AppError` を再度 `err.Error()` すると message が二重化する（`apperrors` の `Error()` 形式）
- 改善案: 固定日本語（`"CSVの形式が正しくありません"`、`"日時の形式が正しくありません"`）。既に `error` が `AppError` なら `RespondError(c, err)` をそのまま

#### BE-RC-014 [MEDIUM][residual] pgx encode 判定が `err.Error()` 文字列 Contains
- 対象: `internal/apperrors/errors.go:344-381` — `isPgxEncodeRangeMessage(err.Error())`
- 規約: `errors.Is` / `As`。コメント上 BUG-138 の既知例外（pgx が typed error を出さない）
- 改善案: 新規に同パターンを増やさない。pgx が typed error を出したら `errors.As` へ置換

### Package API

#### BE-RC-008 [MEDIUM] 互換 `ClinicRepository` が read consumer 向けに `Update(map[string]any)` まで露出
- 対象: `internal/clinic/ports.go:29-44` — staff/auth 向け compatibility API に generic `Update`
- 対比: 同ファイルの `clinicServiceRepository` と `PermissionGroupWriter` は狭い
- 改善案: consumer 用は List/Find。`Update(map)` は unexported か typed `UpdateClinic` のみ

#### BE-RC-009 [MEDIUM] 実装側に定義された広い `XxxRepository` interface が常態
- 規約: interface は利用側の最小メソッド。mock のためだけに作らない。実装は concrete を返す
- 最悪サンプル（メソッド数）: `identitylink.Repository`（~28、`identitylink/repository.go:19-61`）、`medicalrecord.MedicalRecordRepository`（~22）、`billing.AccountingRepository`（~22）、`staff.StaffRepository`（18 + cross-domain map Update）、`clinic.ClinicRepository`（互換ワイド）
- 正例: `internal/lstep/composition.go` の consumer-side 最小 port、`clinic.PermissionGroupWriter`
- 改善案: 新規は呼び出し側で切る。既存の一括分割はしない。触る service の依存を狭い interface に置換

#### BE-RC-010 [MEDIUM] provider 側に consumer 名付き `owner.LstepRepository`
- 対象: `internal/owner/repository.go:53-72` — 「LSTEP workflows が消費する」永続化面を owner が export
- 対比: lstep 側 composition は正しく consumer-side
- 改善案: owner から `LstepRepository` を消し、lstep の狭い port だけを DI。composition 用 mega `Repository` も縮小

### Lint / test

#### BE-RC-012 [MEDIUM] `wrapcheck` が `internal/*` 全体で無効
- 対象: `backend/.golangci.yml` の `wrapcheck.ignore-package-globs: github.com/animal-ekarte/backend/internal/*`
- 現状: wrapcheck は有効化されているが、domain 本番コードには実質かからない。`%w` 規約はレビュー依存
- 改善案: ignore を外し、残件を掃くか、ignore を本当に必要な package だけに縮小

#### BE-RC-013 [MEDIUM] テストがプロセス環境変数を `os.Setenv` で汚染し得る
- 対象: `internal/medicalrecord/vaccination_service_test.go:44-45` — `os.Setenv("DB_HOST"/"DB_PORT")` に `t.Setenv` / Cleanup なし
- 規約: test から global state を漏らさない
- 改善案: `t.Setenv`

---

## 3. LOW

#### BE-RC-015 [LOW] package.Type stutter が系統的（推定 ~167 型）
- 例: `clinic.ClinicService`、`auth.AuthService`、`reservation.ReservationHandler`、`trimming.TrimmingService`、`lstep.LstepSettingsHandler`、`pet.PetResponse`
- 規約: export 名に package 名を繰り返さない
- 改善案: 新規/触る面から `Service` / `Handler` / `Repository`。一括 rename は別タスク（gopls Rename + struct tag 再走査）

#### BE-RC-016 [LOW] BE9 移動後も `// Package handler|service|repository` コメントが残る（26 ファイル）
- 例: `internal/staff/staff_handler.go:1`、`staff_repository.go:1`、`staff_service.go:1`。actual package は `staff` / `billing` / `medicalrecord` 等
- 改善案: `// Package staff ...` に直すかファイル先頭の誤コメントを削除。機械置換可

#### BE-RC-017 [LOW] 同一 package 内の exported `Update(..., map[string]any)` が常態
- inventory / billing / medicalrecord / pet / auth permission_group / trimming 等。多くは同 package service だけが呼ぶが、境界を越えやすい
- 改善案: 触る repository から unexported `update` にし、外には typed command だけ出す

#### BE-RC-018 [LOW] `staffs` / `shift_entries` に appointments 相当の AST write-owner gate がない
- 現状: reservation からの write は薄い delegate のみで、owner 外直接 GORM write は未検出
- 改善案: `appointment_write_owner_lint_test.go` を雛形に staff テーブル mutation gate を追加

#### BE-RC-019 [LOW] `medicalrecord` 本番 ~238 file — 単一 package の凝集圧
- layer subpackage 化は禁止方針どおり正しく避けている
- 改善案: 分割するなら業務能力（lab / hospitalization 等）単位。`handler|service|repository` サブパッケージは作らない。急がない

#### BE-RC-020 [LOW] `nested_summary_response.go` が 3 package に意図的コピー
- `internal/billing/`、`internal/medicalrecord/`、`internal/reservation/`（コメントで非統合を明示）
- 改善案: import cycle が無いなら `httpapi` へ。現状維持でもよい

#### BE-RC-021 [LOW] exported GoDoc 欠如は系統的（revive `exported` / `package-comments` が disabled）
- 公開 API 例: `internal/config/config.go` の `Config` / `Load`、各 `XxxService`
- 改善案: 新規 export には GoDoc。一括はしない

#### BE-RC-022 [LOW] `replace_audit_tail.go` のコメントが削除済み `internal/service` を正本扱いしている
- 対象: `internal/medicalrecord/replace_audit_tail.go:11-17`
- 現状: 旧 layer は消滅済みで、この unexported helper が唯一の実装。`auditActorTypeFor` は `sharedkernel.AuditActorTypeFor` の薄い委譲
- 改善案: コメントを現状に直す。helper 自体の移動は急がない

#### BE-RC-023 [LOW] `init()` で gin validator をグローバル登録
- 対象: `internal/clinic/clinic_request.go:18-35`（`jp_email` / `jp_phone` / `jp_postal`）
- 改善案: 現状は実用上問題なし。テスト順副作用が出たら constructor 登録へ

---

## 4. 合格（再提案しない）

| 項目 | 根拠 |
|------|------|
| 旧 3 layer directory | 不存在。`lintscan/package_boundary_gate_test.go` C2 |
| domain 内 `handler/` `service/` `repository/` subpackage | 0。同 C3 |
| production `util` / `common` / `misc` | 0。`timeutil` / `sharedkernel` は固有名 |
| ADR-006 DAG | `domain_import_allowlist_lint_test.go` と一致。`identitylink` → owner/pet の Go import なし |
| appointments write owner | `appointment_write_owner_lint_test.go`。billing/medicalrecord/trimming は typed intent |
| `ShouldBind*` と bind エラー処理 | 本番 `BindJSON`/`MustBind*` 0。二重 response write 未検出 |
| Context 第 1 引数・cancel・goroutine | request-scoped の struct 保持なし。`WithTimeout` は `defer cancel`。`GoSafe` + 標準 Context |
| Route の public / auth 分離 | `auth.RegisterRoutes`、LIFF / LINE webhook は engine 直下に分離 |
| 機密 field | `PasswordHash` / LINE secret 等は `json:"-"`。主要 CRUD は `toXxxResponse` |
| 5xx envelope | `RespondError` は汎用 message、内部は `c.Error` |
| Server lifecycle | `Read/Write/ReadHeader/Idle` timeout、graceful shutdown、release の trusted proxy / HTTPS |
| CORS | allowlist + credentials。wildcard なし |
| AutoMigrate | testdb のみ |
| nested Preload 中間 clinic scope | 確認分は中間 association にも `clinic_id`（例: reservation trimming detail） |
| RowsAffected | `persistence.UpdateScopedByID` が 0 行 → NotFound |
| FOR UPDATE の ambient tx 必須 | clinic/estimate/medical_record 等で fail-closed |
| no-show Mark + 監査 | 同一 `WithTx` |
| 締め後会計監査 | fail-closed（同一 tx） |
| LSTEP バッチの新規 silent swallow | 取得失敗で `return 0, nil` かつ Failed 非計上は未検出。`BatchRunResult.Failed` 計上あり |
| HTTP httptest | RegisterRoutes を持つ domain にゼロ域なし |
| 本番 800 行超 | 0（最大 702: `pet/repository.go`） |
| `cmd/` 複数エントリ | api / migrate / csv-import* / lab-device-agent 等。ガイドラインと整合 |

---

## 5. lint / 設定上の既知緩和（指摘だが急がない）

- `gocritic` の `hugeParam` / `unnamedResult` / `rangeValCopy` は 2026-09-02 に composition/DTO 値渡し由来で disable（`.golangci.yml` コメント）。ポインタ化は別タスク
- `revive` の `exported` / `package-comments` / `context-as-argument` / `unexported-return` は意図的 disable
- `contextcheck` は `internal/middleware` と `cmd/api` で除外（`*gin.Context` が Context を内包する false positive）
- `cmd/csv-import*` / `cmd/seed-old-db` / `internal/csvimport` は gocritic/gosec 緩和（カットオーバー CLI）
- `//nolint` 約 109 行（gosec ~52、errcheck ~26 が中心）

---

## 6. 却下済み・再提案しない（backend/CLAUDE.md より）

- カルテ同日重複に DB unique を採らない（2026-07-27）
- auto-create に clock seam を導入しない（2026-07-27）
- Count→Delete の**一括** retrofit を本監査の実装スコープにしない（CODING_RULES。触る Delete では直す）
- medicalrecord を `handler/service/repository` サブパッケージへ層分割しない
- `map[string]any` の監査 metadata / テスト fixture を禁止しない
- wrapcheck を host の full `golangci-lint run ./...` でエージェントが回さない（`.claude/CLAUDE.md` 禁止コマンド）

---

## 7. 実施記録

キャンペーンブランチ: `refactor/be-rc-2026-09`（base `42510f25c`）。claim: `claim/BE-RC-CAMPAIGN-2026-09`（削除せず残置）。

| ID | Status | 変更ファイル | 検証 |
|---|---|---|---|
| BE-RC-001 | DONE | `staff/reservation_staff_update.go`, `staff/staff_repository.go`, `reservation/reservation_staff_repository.go`, `reservation/reservation_staff_service.go`, isolation/mocks, `staff/staff_clinic_assignment_reservation_race_test.go`（`package staff_test` へ移し import cycle 解消） | `docker compose exec -T backend go test ./internal/staff/...` ok; `./internal/reservation/...` ok。exported `UpdateForReservation` / `staffsWriter` は `ReservationStaffUpdate`。production staff は reservation を import しない |
| BE-RC-002 | DONE | billing/lstep/medicalrecord/owner/pet request ファイルと既存 request テスト | ShouldBindJSON 超過テスト追加。reason 500 / memo-notes-content-AfterText 1000 / name 255 / LINE Text 5000 / FileName+Purpose 255 |
| BE-RC-003 | DONE | `inventory/repository.go` `DeleteIfUnused`; `billing/payment_method_master_repository.go`; `billing/insurance_repository.go` | `./internal/inventory/...` ok。billing は `PaymentMethodMasterRepository_Delete|InsuranceRepository_Delete|PaymentMethodMasterRepository_Update` 緑。package 全体の LTV/CompleteForAccounting 失敗は testdb WIP 由来で本差分外 |
| BE-RC-004 | N/A | — | 他マスター Count→Delete 一括は CODING_RULES residual。003 の 3 面のみ実施 |
| BE-RC-005 | N/A | — | slog 一括削除は別タスク。新規に二重ログを増やしていない |
| BE-RC-006 | DONE | 列挙 handler（reservation/liff/validators/queries, hospitalization/daily_record, inventory, cash_register, chronic_condition, shift, http_password, lstep CSV/checkup_sync） | `WrapInvalidInput(err.Error())` 列挙サイト 0（テストコメント除く）。固定日本語または AppError 素通し |
| BE-RC-007 | DONE | `billing/payment_method_master_repository.go` | 同一 tx の update+reload。reload 失敗で write rollback。テスト追加 |
| BE-RC-008 | DONE | `clinic/ports.go`, `clinic_repository.go`, `clinic_service.go` | 消費者 `ClinicRepository.Update(map)` を `UpdateClinic(*UpdateClinicInput)` に置換。AST gate 追加 |
| BE-RC-009 | N/A | — | fat interface 一括分割はしない。触った面で mega Repository を広げていない |
| BE-RC-010 | DONE | `owner/repository.go` `LifecycleOwnerRepository`; `cmd/api/lstep_adapters.go`; `cmd/lstep-migrate/migrator.go` | provider 側に exported `LstepRepository` なし。cmd は新名を消費 |
| BE-RC-011 | N/A | — | 見積通常 CRUD / approve/reject 監査は意図的 post-commit best-effort。CreateSuccessor のみ fail-closed。product 未決のため実装せず |
| BE-RC-012 | DONE | `backend/.golangci.yml` | `internal/*` ワイルドカード廃止。未触 package を明示 ignore。staff+clinic の scoped golangci は wrapcheck 18（既存 WithTx 素通し）。20 件超ではないため package ignore 再追加はせず、%w 一括掃討もしない |
| BE-RC-013 | DONE | `medicalrecord/vaccination_service_test.go` | `os.Setenv` 0、`t.Setenv` あり。`db_testmain_test.go` は testdb `sync.Once` 用の TestMain 残置 |
| BE-RC-014 | N/A | — | pgx string Contains は BUG-138 既知例外。新規に同パターンを増やしていない |
| BE-RC-015 | N/A | — | stutter 一括 rename は別タスク |
| BE-RC-016 | DONE | allowlist 内 stale `// Package handler\|service\|repository` を実 package 名へ（staff/billing/medicalrecord）。`reservation_type_handler.go` と trimming は allowlist 外のため残置 | allowlist 内 0。allowlist 外は N/A residual |
| BE-RC-017 | N/A | — | 同一 package 内 map Update 一括 unexport はしない。001/008 の境界のみ typed 化 |
| BE-RC-018 | DONE | `reservation/reservation_staff_write_owner_lint_test.go` | map を `UpdateForReservation` に戻すと RED。現状 GREEN（reservation パッケージテスト） |
| BE-RC-019 | N/A | — | medicalrecord 分割はしない。layer サブパッケージ禁止 |
| BE-RC-020 | N/A | — | nested_summary 3 package コピーは意図的現状維持 |
| BE-RC-021 | N/A | — | GoDoc 一括はしない |
| BE-RC-022 | DONE | `medicalrecord/replace_audit_tail.go` | `internal/service` 言及 0。現状パスへ更新 |
| BE-RC-023 | N/A | — | clinic `init()` validators は実用上問題なし。テスト順副作用が出たら constructor 登録 |

Assumptions からの deviation:
- typed staff command のフィールドは prompt 仮定（nameKana/role/username/...）ではなく現行 `buildReservationStaffUpdate` のキー（name, staff_type, reservation_visible, reservation_comment, sort_order）。加えて既存 `PatchStatus` が書いていた `is_active` のみ typed 化。
- W1 は `ReservationStaffRepository` シグネチャ変更に伴い `liff_service_*_test.go` の mock を型合わせ（compile 必須）。
- レーン writer が `claim/BE-RC-*` を追加作成した。キャンペーン claim に加え残置。削除は USER-only。
