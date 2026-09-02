# リファクタ台帳（BE コード規約準拠）

更新日: 2026-09-02（実装完了）

| 項目 | 値 |
|------|-----|
| **範囲** | backend の **production Go**（`*_test.go` と `cmd/_archive/` は対象外） |
| **目的** | 挙動を変えずに、プロジェクトの Go 規約へ寄せる |
| **ブランチ** | `main`（未コミット。ユーザー依頼で commit しない） |
| **検証** | 変更 package の Docker 経由 scoped `go test`。full `./...` はユーザー手動 |
| **正本** | [go-language.md](.claude/refs/go-language.md) · [go-gin-backend-guidelines.md](.claude/rules/go-gin-backend-guidelines.md) · [error-handling.md](.claude/refs/error-handling.md) |

スキャン（完了時）: `go/ast` の真の if-in-if。`else if` は兄弟。`for`/`switch` は加算しない。

---

## 対象外（意図的にやらなかった）

| 除外 | 理由 |
|------|------|
| Frontend | BE 限定 |
| `GetByID` / HTTP `Get*` の一括リネーム | 既存契約の機械改名は過大 |
| JWT `Keyfunc` の `any`、`[]any` の SQL args、`gin.H` | ライブラリ境界 |
| GORM `Updates(map[string]any)` の struct 化 | **BE-ANY 例外**（下記） |
| ネスト要約 DTO の package 横断統一 | JSON 契約が domain ごとに違う（reservation pet は `danger_level`、billing pet は id+name） |
| `preflightCSVShape` の機械分割 | CSV 状態機械。分割すると挙動を壊しやすい |

---

## 完了したカテゴリ

### BE-CF — early return / 深いネスト

- **BE-CF-001/002**: nest≥4 は **0**。payment graph の completed ガード、auth notifier、会計 unique replay、見積割引、予約検証、シフト、LSTEP CSV、medicine 名衝突、trimming options を flatten。
- **BE-CF-003**: 優先ファイルと 1 箇所ファイルを helper / early return で flatten。完了時 nest≥3 は **0 を目標に潰した**（`csvbundle` の clinic trigger、`clinical_plan` の zero-row、delivery batch の owner map を含む）。

### BE-ERR

- `lab_import_revert_service.go` / `checkup_package_import_service.go` の `err == gorm.ErrRecordNotFound` → `errors.Is`。

### BE-DRY

- `httpapi.ValidateEnum` / `ParseDate` / `ParseFlexibleDate`
- `httpapi/string_query.go`（`OptionalString`、optional uint64/date）
- `medicalrecord/date_parse.go` 削除
- domain の query helper は httpapi へ委譲
- discount は httpapi 正本、medicalrecord は呼ぶだけ
- nested summary は **domain-local JSON 契約**としてコメント修正（型は統合しない）

### BE-DOC

- `.golangci.yml` の死んだ `internal/handler/` 除外を削除
- `httpapi/context.go` の package コメントを現行配置に合わせて書き換え
- production コメントの「handler に残す」未来形を削除

### BE-GO

- `cutover_import.go`: transform panic は `CloseWithError` + 結果送信（COPY がハングするため `GoSafe` 単独は使えない）
- `labdeviceagent/agent.go`: `GoSafe`

### BE-ANY（変換しない。例外を固定する）

GORM `Updates(map[string]any)` は **ゼロ値を省略する**境界 API。同じフィールドを struct の `Updates` にすると `false` / `0` / `""` が「未指定」ではなく「明示ゼロ」になり、PATCH 契約が変わる。

既存の typed ヘルパー（`sharedkernel.SetNullableUint64Field` など）は nullable 列用。`build*Update` が返す map は **persistence 境界の例外**として残す。JSON/API 契約は変えない。

`lstep` の `any(s.repo).(iface)` optional loader は constructor 注入へ上げられるが必須ではない。今回は触っていない。

### BE-SIZE

- **800 行超ファイル: 0**（分割例: identitylink / reservation / trimming / auth session / staff provisioning / billing item / csvimport cutover / medicalrecord routes / permission_group / stg-uat cmd）
- `Complete` は `completeInTx` + items/payments/post-close helper に分割
- `RegisterRoutes` は `routes_*.go` に分割
- 残る 150 行超は SQL 組み立て・CSV 状態機械・巨大バリデータ。追加の機械分割は挙動リスクが分割利益を上回るため、触った関数だけ 50 行を運用する（台帳の当面ルール）

残 150+（参照、今回は関数ごとファイル移動または helper 抽出まで）:

| 行 | 場所 |
|----|------|
| ~232 | `csvimport/cutover_payment_contract.go` `validateCutoverPaymentGraph`（CF-001 の guard 済み。stream コールバックが本体） |
| ~215 | `medicalrecord/checkup_package_import_service.go` `Apply` |
| ~200 | `owner/ltv_repository.go` `FindOwnerLTV`（ヘルパーは `ltv_repository_query.go`） |
| ~184 | `billing/billing_item_repository_vaccination.go` `ValidateVaccinationCreateReference` |
| ~181 | `billing/accounting_repository_reports_close.go` `GetCloseAggregate` |
| ~181 | `trimming/trimming_service_mutate.go` `Update` |
| ~175 | `auth/password_reset_service.go` `ForgotPassword` |
| ~165 | `medicalrecord/lab_import_examination_service.go` `persistExam` |
| ~163 | `staff/staff_service_core.go` `Update` |
| ~162 | `lstep/lstep_csv_stream.go` `preflightCSVShape` |
| ~161 | `reservation/reservation_validators.go` `ValidateAndCreate` |
| ~158 | `billing/accounting_report_service.go` `buildMonthlyReportResponse` |
| ~157 | `medicalrecord/medical_record_repository_list.go` `FindAll` |
| ~156 | `medicalrecord/hospitalization_discharge.go` `DischargeWithBilling` |
| ~155 | `medicalrecord/lab_import_revert_service.go` `Revert` |

---

## 検証（Docker / scoped）

通したもの:

- compile-only: billing, medicalrecord, csvimport, middleware, httpapi, identitylink, reservation, trimming, auth, staff, lstep, labdeviceagent, owner, cmd/stg-uat-*, cmd/staff-provision, cmd/migrate
- `./internal/csvimport` `./internal/httpapi` `./internal/labdeviceagent` `./internal/middleware` `./internal/identitylink`
- `./internal/medicalrecord -run 'TestRegister|TestRoute|Snapshot'`
- `./internal/billing -run 'TestAccountingService_CompleteAccounting|TestEstimateHandler'`

full `go test ./...` は禁止コマンド。複数 package を同時に同じ DB へ流すと TRUNCATE deadlock が出る（以前の一括実行で観測）。package 単位で流すこと。

---

## 完了条件チェック

- 公開 JSON / HTTP status / error code を変えない（意図）
- clinic isolation と write-owner を迂回しない
- `gofmt` / `goimports` 済み
- nest≥4 の新しい `if` を残さない
- GORM map Updates を struct に置換していない
