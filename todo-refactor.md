# リファクタ台帳（BE コード規約準拠）

更新日: 2026-09-02（100–109行の公開関数抽出まで完了）

| 項目 | 値 |
|------|-----|
| **範囲** | backend の **production Go**（`*_test.go` と `cmd/_archive/` は対象外） |
| **目的** | 挙動を変えずに、プロジェクトの Go 規約へ寄せる |
| **ブランチ** | `main`。第1弾 `739207a1f`（ファイル分割）。第2弾 `8fa232d09`（150行超関数）。第3弾 `e4f47fbba`（140行超関数）。第4弾 `b94f1dd7e`（130行超関数）。第5弾 `56df7c429`（120行超関数）。第6弾 `994055a95`（110–119行の公開関数）。第7弾は続くコミット（残110行と100–109行の公開関数） |
| **検証** | 変更 package の Docker 経由 scoped `go test`。full `./...` はユーザー手動 |
| **正本** | [go-language.md](.claude/refs/go-language.md) · [go-gin-backend-guidelines.md](.claude/rules/go-gin-backend-guidelines.md) · [error-handling.md](.claude/refs/error-handling.md) |

スキャン: 関数行数は `func ` の開き `{` から対応する `}` まで（brace 対応）。`AuditLog.TableName` は後続 const ブロックで過大カウントされる偽陽性。`testdb.SetupIsolatedTestDB` はテスト基盤のため対象外。

---

## 対象外（意図的にやらなかった）

| 除外 | 理由 |
|------|-----|
| Frontend | BE 限定 |
| `GetByID` / HTTP `Get*` の一括リネーム | 既存契約の機械改名は過大 |
| JWT `Keyfunc` の `any`、`[]any` の SQL args、`gin.H` | ライブラリ境界 |
| GORM `Updates(map[string]any)` の struct 化 | **BE-ANY 例外**（下記） |
| ネスト要約 DTO の package 横断統一 | JSON 契約が domain ごとに違う（reservation pet は `danger_level`、billing pet は id+name） |
| `preflightCSVShape` の機械分割 | CSV 状態機械。分割すると挙動を壊しやすい。残る唯一の実 150 行超関数 |
| `updateExaminationInTx` の再分割 | 第3弾で切り出した tx helper（122行）。これ以上割ると Update 契約が読みにくくなる |
| 直前スライスで切り出した tx helper の再分割 | applyCheckupPackageInTx / revertLabImportInTx / createLineReservationInTx / dischargeWithBillingInTx / createTrimmingDetailForExistingInTx |
| `ForgotPassword` の再分割 | 応答時間フロアと mail reservation の契約が残っている。persist/dispatch は第3弾で抽出済み |
| `consumeResetTokenInTx` / `updateTrimmingInTx` / `setClinicAssignmentsInTx` の再分割 | 直前スライスの tx helper。再分割しない |
| `Config.Validate` / `RegisterRoutes` / LSTEP tag batch / 会計 SQL | env 列挙・route 表・バッチ同期・Raw SQL。分割しても契約が読みにくくなる |
| `cmd/lstep-migrate` の `run` | CLI 起動手順。processOwner 抽出で十分 |
| 95–99 行まで機械分割 | 規約ゲートは 150。50 行ルールは個人コーディングスタイルであり台帳ゲートではない |

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

既存の typed ヘルパー（`sharedkernel.SetNullableUint64Field` など）は nullable 列用。`build*Update` が返す map は **persistence 境界の例外**として残す。JSON/API 契約は変えない。`BuildClinicUpdate` / `buildMedicineUpdate` は map のまま、フィールド群だけ helper に分けた。

`lstep` の `any(s.repo).(iface)` optional loader は constructor 注入へ上げられるが必須ではない。今回は触っていない。

### BE-SIZE

- **800 行超ファイル: 0**
- **150 行超の実関数: `preflightCSVShape` のみ**（意図的残置）
- **140 行超の実関数: 0**（`preflightCSVShape` と AuditLog 偽陽性を除く）
- **130 行超の実関数: 0**（`preflightCSVShape` と AuditLog 偽陽性を除く）
- **120 行超の実関数: `updateExaminationInTx` のみ**（第3弾 helper。意図的残置）
- 第2弾の抽出例: payment graph loader、LTV SQL 定数、締め集計クエリ、月次 daily map、カルテ検索 WHERE、予防接種 lock/quote、LINE 予約 create、trimming/staff Update、退院会計、lab persist/revert、健診パッケージ types/fields、ForgotPassword persist
- 第3弾の抽出例: trimming Create / existing-detail、pet identity group create、ForgotPassword mail dispatch、billing item ambient create、examination Update、月次レポート scan、クレジット訂正、complete digest、medicine update フィールド群
- 第4弾の抽出例: staff provision validator、lab job item resolution、LSTEP checkup preview SQL、会計 Update tx、入院 Create tx、健診パッケージ apply tx、予防接種 lock 分割、lstep-migrate CLI。dbortx inventory のファイル分割後パスも追随
- 第5弾の抽出例: treatment Create/Update tx、vital Update tx、予約からの自動カルテ create tx、見積 successor tx、検査 replace helper、clinic assignment tx、synthetic failure injection、complete replay/header、健診 manifest types/fields
- 第6弾の抽出例: カルテ Update tx、owner identity group create tx、空き日評価、検査 revision persist/project、password reset consume tx、refresh identity/clinic、credential audit clinic pick、見積 Create/Update tx、lab device receive tx
- 第7弾の抽出例: `/me` DTO 合成、login clinic scope/audit、薬剤 Create tx、ペット一覧 query/search、LSTEP CSV spool/header/actor/tx、lstep-migrate owner sync、lab exam persist write、LIFF 空き時間 filter

---

## 検証（Docker / scoped）

第1弾: compile-only 多数 package。csvimport / httpapi / labdeviceagent / middleware / identitylink。medicalrecord の Register/Route/Snapshot。billing Complete/Estimate。reservation / trimming / auth / staff は package 単位 PASS。

第2弾:

- compile-only: reservation, owner, csvimport, billing, medicalrecord, trimming, staff, auth, clinic
- `./internal/csvimport` `./internal/owner` `./internal/clinic` `./internal/billing`
- `./internal/medicalrecord -skip TestLabDeviceConnectivityDocsMatchRuntime`（欠ける hex-skills 文書。本差分と無関係）
- `./internal/reservation` `./internal/trimming` `./internal/staff` `./internal/auth`

staff の `TestStaffCredentialMutationAuditSourceContract` は `staff_service_update.go` も読むように更新した。

第3弾:

- compile-only: trimming, identitylink, auth, billing, medicalrecord
- `./internal/trimming` `./internal/identitylink` `./internal/auth` `./internal/billing`
- `./internal/medicalrecord -skip TestLabDeviceConnectivityDocsMatchRuntime`

第4弾:

- compile-only: staff, medicalrecord, lstep, billing, lintscan, `./cmd/lstep-migrate`
- `./internal/lintscan -skip TestMigrationCascadeInventory_NoUnreviewedCascade`（未適用 migration の CASCADE レビュー。本差分と無関係。`claim/DB-INIT-SCHEMA-HARDENING` 側）
- `./internal/staff` `./internal/lstep` `./internal/billing`
- `./internal/medicalrecord -skip TestLabDeviceConnectivityDocsMatchRuntime`

第5弾:

- compile-only / `go vet`: medicalrecord, billing, staff, csvimport
- `./internal/staff` `./internal/csvimport` `./internal/billing`
- `./internal/medicalrecord -skip TestLabDeviceConnectivityDocsMatchRuntime`

第6弾:

- compile-only / `go vet`: medicalrecord, billing, identitylink, reservation, auth, cmd/api
- `./internal/identitylink` `./internal/auth` `./internal/reservation` `./internal/billing` `./cmd/api`
- `./internal/medicalrecord -skip TestLabDeviceConnectivityDocsMatchRuntime`

第7弾:

- compile-only / `go vet`: auth, medicalrecord, pet, lstep, reservation, cmd/lstep-migrate
- `./internal/auth` `./internal/pet` `./internal/lstep` `./internal/reservation`
- `./internal/medicalrecord -skip TestLabDeviceConnectivityDocsMatchRuntime`
- `./cmd/lstep-migrate`（テストファイルなし）

full `go test ./...` は禁止コマンド。複数 package を同時に同じ DB へ流すと TRUNCATE deadlock が出る。package 単位で流すこと。

---

## 完了条件チェック

- 公開 JSON / HTTP status / error code を変えない（意図）
- clinic isolation と write-owner を迂回しない
- `gofmt` / `goimports` 済み
- nest≥4 の新しい `if` を残さない
- GORM map Updates を struct に置換していない
- 150 行超は `preflightCSVShape` 以外ゼロ（実関数）
- 140 行超の生産関数は `preflightCSVShape` 以外ゼロ
- 130 行超の生産関数は `preflightCSVShape` 以外ゼロ
- 120 行超の生産関数は `preflightCSVShape` と `updateExaminationInTx` 以外ゼロ
- 110 行超の公開関数は直前スライスの tx helper と ForgotPassword を除きゼロ
- 100 行超の公開関数は直前スライスの tx helper・ForgotPassword・Config.Validate・RegisterRoutes・LSTEP tag batch・会計 SQL・lstep-migrate `run` を除きゼロ
