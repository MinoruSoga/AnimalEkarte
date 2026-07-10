# BE-refactor-v2.md — backend/ リファクタリング実行計画

- **作成日**: 2026-07-10 / **基準 HEAD**: `54706959`
- **対象読者**: この計画書とリポジトリのコード以外の文脈を持たない実行者（AI/人間）
- **性質**: 全項目が**挙動保存**（ユーザー可視の動作・API レスポンス・DB 書込内容を一切変えない）。挙動変更はすべて「§5 やらないこと」で禁止する。

> **旧 `BE-refactor.md` との関係（重要）**: 旧計画書の「残タスク 18 件」は実態から乖離している。
> G2-1(`1241ae80`)・G2-2(`1f711f55`)・G1-1・G1-4(`65e7f29b`)・G1-3 の大半(`9e83aa4b`)・G1-2 の Phase A〜D(多数の `docs(api)` コミット)は**既にコミット済み**である。
> 本書は 2026-07-10 の現 HEAD に対する全項目再検証（行番号・前提の実測）を経て作成した**唯一の正**であり、旧 `BE-refactor.md` は参照のみ・編集禁止とする。

---

## §1 現状理解（構造マップ）

### 1.1 システム概要

動物病院向け電子カルテ SaaS のバックエンド。Go 1.25 / Gin / GORM / PostgreSQL 18。
モジュール名 `github.com/animal-ekarte/backend`。軽量レイヤードアーキテクチャ **handler → service → repository → model**。
**マルチテナント**: ほぼ全テーブルが `clinic_id` を持ち、クリニック間のデータ分離が最重要インバリアント。

### 1.2 エントリポイント（backend/cmd/）

| バイナリ | 役割 |
|---|---|
| `api` | 本体 HTTP API サーバ。`main.go` で config → logger → JST タイムゾーン → DB → DI → ルータ → graceful shutdown。バッチ goroutine 3 本（no-show/休眠/配信トリガー）も起動 |
| `migrate` | `*.sql` migration + CSV seed バンドルを advisory lock 下で適用。api 起動前に entrypoint.sh から実行される |
| `coverage-ratchet` | CI のカバレッジ下限ゲート |
| `lstep-migrate` / `seed-export` / `seed-old-db` / `stage-import` | 運用 CLI 群（本計画では触らない） |

### 1.3 internal/ パッケージ

| パッケージ | 役割 |
|---|---|
| `handler` (非テスト256) | HTTP 層。`*_request.go`（バインド→`service.XxxInput` 変換）→ `*_response.go`（model→wire DTO 変換）。エラーは全て `RespondError(c, err)`（handler/response.go）経由 |
| `service` (非テスト196) | 業務ロジック。`service.go` の `NewServices()` + `cmd/api/main.go:90-162` の二段階配線（既知の負債・本計画では触らない） |
| `repository` (非テスト105) | データアクセス。共通基盤: `base.go` の `clinicScope`/`clinicScopeIn`/`medicalRecordTenantScope`/`dbOrTx(ctx,db)`（ambient tx 参加）、`helpers.go` の `findByIDScoped`/`updateScopedByID`/`deleteScopedByID`/`reorderByClinicID`。tx 機構は①`transactor.go` の `WithTx`(ctx-txKey 方式) と②`repositories.go` の `Transaction`(repo-swap 方式) の 2 系統が意図的に併存 |
| `model` (非テスト85) | GORM struct（1 テーブル 1 ファイル）。`audit_log.go` に監査タクソノミ定数 |
| `errors` | 全層が `apperrors` エイリアスで import。センチネル + `Wrap*` + `FromGORM`。**bare `return err` 禁止**（wrapcheck が CI ゲート） |
| `apicontract` | `docs/api.yaml` ↔ handler 実装の drift 検査テスト専用パッケージ（route gate + date-format gate） |
| `config` / `dbconn` / `infra` / `logger` / `middleware` / `seedbundle` | 設定 / cmd 系 DSN 共通化 / LINE・Lstep・S3・crypto / slog / 認証・CORS 等 / seed マニフェスト |

各層のコーディング規約は `backend/internal/{handler,service,repository}/CLAUDE.md` と `backend/CODING_RULES.md` に正本がある。**新規テストコードを書く際は必ず該当層の CLAUDE.md を先に読むこと。**

### 1.4 ガード lint テスト（絶対に緑を維持する自己強制インバリアント）

このリポジトリは「規約を人間が守る」のではなく「規約違反で CI が落ちる」設計。以下を壊す変更・allowlist の無断更新は禁止:

| テスト | 固定している内容 |
|---|---|
| `repository/dbortx_inventory_lint_test.go` | `dbOrTx` 参加メソッドの双方向名簿（revert で fail） |
| `repository/preload_clinic_scope_lint_test.go` | clinic-scoped マスタ Preload の `clinic_id` 述語必須 |
| `repository/migration_cascade_lint_test.go` | migrations の `ON DELETE CASCADE` 件数固定 |
| `repository/audit_tx_inventory_lint_test.go` | 臨床 hard-delete の tx 内監査（fail-closed）名簿 |
| `service/master_fk_write_inventory_lint_test.go` | request 由来マスタ FK write の全数名簿（guarded/known-unguarded） |
| `model/schema_drift_test.go` | GORM model ↔ 実 DB スキーマの列存在・型カテゴリ比較（DB 接続必要） |
| `model/audit_taxonomy_exhaustiveness_test.go` | 監査アクション定数の網羅 |
| `apicontract/openapi_route_drift_test.go` | 実装ルート(477) ↔ api.yaml paths の突合。allowlist は stale 検出付き＝**解消したら対応エントリを削除しないと fail する** |
| `apicontract/openapi_date_format_drift_test.go` | `format: date` ↔ Go `time.Time` の drift 22 件 pin |

### 1.5 テスト基盤

- **service 層** = mock リポジトリ、**repository 層** = 実 DB（`ekarte_db_test`）という層別方針。
- repository の DB テストは `ltv_repository_test.go` 内の `setupTestDB(t)`(:1328) / `setupSharedTestSchema(db)`(:1387) を共有。テスト DB が無ければ **CREATE DATABASE で自動作成**される（`db_test.go` 参照）。スキーマは「手書き CREATE TYPE 列挙 + AutoMigrate」で作られ、`migrations/001_init.sql` の複製である（→ R1 の対象）。
- テスト総数: `_test.go` 571 ファイル。カバレッジは CI の coverage-ratchet が下限管理。

### 1.6 検証コマンド規約（Docker 必須・スコープ限定）

- ローカルの `go` / `npm` / `pnpm` 直接実行は禁止。**必ず** `docker compose exec backend go test ...`。
- **フルリポジトリ検証は自動実行禁止**: `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` を実行しない。パッケージ単位（例 `./internal/repository/`）のスコープ限定実行のみ許可。
- `docker compose exec db psql ...`（直接 SQL）は原則禁止。本計画では R1 のテスト DB 再作成 1 コマンドのみ例外として明示する（実行環境が拒否する場合は人間に依頼）。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<変更ディレクトリ>/` が**無出力**であることを確認してからコミットする。

---

## §2 項目 R0: 安全網の構築（最初に必ず実行）

### R0-1. 前提確認

```
docker compose ps            # backend / db が Up であること
git rev-parse HEAD           # 記録する（ロールバックの基準点）
git status --porcelain -- backend/  # 空であること。空でなければ中断して報告
```

`backend/` 外（`.claude/` 等）に未コミット変更が見えても無視してよい。**それらをステージ・コミットしてはならない。**

### R0-2. ベースライン確認（既存テストが安全網。特性テストの新規作成は不要）

以下を順に実行し、**全て `ok`（FAIL 0）であることを確認して結果（パッケージ毎の PASS/FAIL とテスト数）を記録**する。1 つでも FAIL があれば**着手せず中断・報告**（このベースラインが赤なら以降の完了判定が成立しない）:

```
docker compose exec backend go test ./internal/apicontract/ -count=1
docker compose exec backend go test ./internal/config/ -count=1
docker compose exec backend go test ./internal/model/ -count=1        # DB 接続必要
docker compose exec backend go test ./internal/service/ -count=1
docker compose exec backend go test ./internal/repository/ -count=1   # DB 接続必要・数分かかる
```

### R0-3. コミット規約

- **1 項目 = 1 コミット**。メッセージは `<type>(<scope>): <説明> (R<n>)` 形式（type: test/refactor/docs/chore）。
- コミットに `Co-Authored-By` 行を**含めない**。
- **push しない**。コミットはローカルに留め、完了報告でハッシュ一覧を提示する。
- 戻し方の原則: 直前 1 項目の失敗は `git reset --hard HEAD~1`（未 push のため安全）。それより前は該当コミットを `git revert`。

---

## §3 作業項目リスト（この順に実行する）

> 各項目の「完了条件」は**全て満たして初めて完了**。満たせない場合はその項目の変更を破棄（コミットしない）し、SKIP/BLOCKED として理由を記録して**依存されていない次の項目へ進む**（§6 手順 3・4 と同一ルール。依存先が SKIP された項目は連鎖 SKIP）。
> **テストを追加した結果、実装側の実挙動がテストの期待と食い違った場合、実装コードを直してはならない**（挙動保存計画のため）。それは本計画が発見した実バグである — 当該項目を BLOCKED と記録し、失敗テストの内容と実挙動を報告して次の独立項目へ進む。

### R1. テスト DB スキーマの enum 複製を 001_init.sql と同期し、再発防止 lint を新設（旧 G12-2）

- **対象**: `backend/internal/repository/ltv_repository_test.go:1396-1453`（enum 列挙）、同 `:1389-1390`（stale コメント）、新規 `backend/internal/repository/test_schema_enum_parity_test.go`
- **問題**: テスト DB スキーマの手書き enum 複製が正本 `migrations/001_init.sql`（CREATE TYPE 54 型）に対し 50 型しかなく、しかも `item_source`（:1433）に `'trimming'` が欠落している。このため `source='trimming'` を INSERT する統合テストは書いた瞬間に enum 違反で失敗し、R3 が実行不能。:1456-1462 の `DO $$ ... IF NOT EXISTS` ガードのため、**文字列を直しても既存のテスト DB には反映されない**。
- **変更内容**:
  1. `:1433` を `CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization', 'trimming')` に修正（001_init.sql:108 と一致させる）。
  2. 欠落 4 型 `campaign_discount_type` / `checkup_field_type` / `lab_import_job_status` / `lab_import_source_type` を 001_init.sql の定義どおり追加。ただし `checkup_field_type` は `checkup_field_repository_test.go` が自前で DROP+CREATE する real-DDL ヘルパーと衝突しないか先に確認し、衝突するなら追加を見送り、parity lint 側で根拠コメント付き allowlist に載せる。
  3. `:1389-1390` のコメント「001_init.sql の 46 型 + …」を実数（54 型）に修正。
  4. 新規 `test_schema_enum_parity_test.go`: `migration_cascade_lint_test.go` と同じ相対パス読取で `001_init.sql` から `CREATE TYPE <name> AS ENUM (...)` を正規表現抽出（改行・空白を正規化）し、`ltv_repository_test.go` の enum 列挙スライスと**名前→値リストの完全一致**を双方向比較するテスト `TestTestSchemaEnumParity` を追加。差分があれば型名と値差分を出力して fail。意図的除外は根拠コメント付き allowlist 変数で管理。
  5. テスト DB を作り直す（IF NOT EXISTS ガード対策。テスト DB は使い捨てで、次回 `setupTestDB` が自動再作成する）:
     ```
     docker compose exec db psql -U ekarte_user -d ekarte_db -c "DROP DATABASE IF EXISTS ekarte_db_test;"
     ```
     ※実行環境がこのコマンドを拒否する場合は人間に依頼する。`ekarte_db_test` 以外の DB 名を対象にしてはならない。
- **完了条件**:
  - `docker compose exec backend go test ./internal/repository/ -run TestTestSchemaEnumParity -count=1` → PASS
  - `docker compose exec backend go test ./internal/repository/ -count=1` → 全 PASS（DB 再作成後のフル再構築を含めて緑）
- **リスク / 戻し方**: enum 追加により AutoMigrate と手書き DDL の競合が顕在化する可能性 → repository パッケージ全体が緑に戻らない場合は `git reset --hard HEAD~1` + テスト DB 再 DROP で復元。
- **依存**: なし（**R3 より先に必須**）

### R2. 会計完了→予約完了化 SQL の DB 実行テスト追加（旧 G11-2）

- **対象**: 新規 `backend/internal/repository/accounting_complete_appointments_test.go`。テスト対象は `accounting_repository.go:312-352` の `CompleteAccountingAppointments`
- **問題**: 受付ボードの orphan カード残留を防ぐ #77 対策の中核 SQL（JST 日付境界比較 `DATE(start_time AT TIME ZONE 'Asia/Tokyo')` :321-322 と、medical_record_id 経由サブクエリ UPDATE :335-341 の 2 経路）に repository 層テストが 0 件。既存テストは service 層 mock の呼び出し有無（`accounting_service_test.go` の `completeApptsFn`）しか見ておらず、SQL セマンティクスの退行を検出できない。
- **変更内容**: `setupTestDB` を使う新規テストファイルを追加し、`TestAccountingRepository_CompleteAccountingAppointments` 配下にサブテストとして: (1) 同日同ペット status=accounting の予約が completed になる (2) 別日/別ペット/別クリニック/status≠accounting/deleted_at ありは変更されない (3) JST 日付境界: start_time が UTC 前日 15:00(=JST 当日 0:00) は対象、UTC 当日 14:59(=JST 23:59) は対象、UTC 当日 15:00(=JST 翌日 0:00) は対象外 (4) medicalRecordID 経由で status=reserved が完了化される (5) completed/cancelled/no_show は経路(2)で触らない・別クリニックの medical_record は更新されない (6) 戻り値 totalAffected が両経路の合算 (7) ownerID/petID/medicalRecordID が nil の縮退経路。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run TestAccountingRepository_CompleteAccountingAppointments -count=1` → PASS。実装コードの diff が 0 行であること。
- **リスク / 戻し方**: テストが実装の実バグを暴く可能性（特に JST 境界）→ 冒頭の BLOCKED プロトコルに従う。戻しは `git reset --hard HEAD~1`。
- **依存**: R0

### R3. トリミング未請求明細 raw SQL の DB 実行テスト追加（旧 G11-3）

- **対象**: 新規 `backend/internal/repository/billing_item_trimming_test.go`。テスト対象は `billing_item_repository.go:183-279` の `FindUnbilledTrimmingItemsByPetID`（UNION ALL :221・NOT EXISTS :212,:240）と `:283-301` の `CountNonAccountingTrimmingByPetAndDate`
- **問題**: 会計×トリミングの金額起点クエリ（未請求コース/オプションの会計自動取込）が一度も DB で実行されていない。NOT EXISTS の枝別結合条件（course_id vs option_id）、cancelled 請求のみの場合の再取得、UNION 後の並び順が固定されていない。
- **変更内容**: `TestBillingItemRepository_FindUnbilledTrimmingItems` / `TestBillingItemRepository_CountNonAccountingTrimming` を追加: (1) status=accounting のトリミング予約でコース+オプションが結合結果に出る（TrimmingCourseID/TrimmingOptionID の枝別セット確認） (2) 並び順: コース(sort_order=0)→オプション(100+sort_order) 昇順 (3) 有効 billing に同一 appointment_id+course_id の billing_item があると除外 (4) cancelled 請求のみなら再取得対象 (5) price=0/NULL は除外 (6) 別クリニック/別ペット/status≠accounting/カテゴリ≠trimming の除外 (7) Count 側: JST 日付境界と対象 status の判定。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run 'TestBillingItemRepository_(FindUnbilledTrimmingItems|CountNonAccountingTrimming)' -count=1` → PASS。実装 diff 0 行。
- **リスク / 戻し方**: R2 と同じ。戻しは `git reset --hard HEAD~1`。
- **依存**: **R1**（`item_source` enum に `'trimming'` が必要）

### R4. UpdateReservationCapabilities の越境ガード isolation テスト追加（旧 G11-4）

- **対象**: 新規 `backend/internal/repository/reservation_staff_capability_write_clinic_isolation_test.go`。テスト対象は `reservation_staff_repository.go:279-318` の `UpdateReservationCapabilities`（所有権 Count ガード :288）と `:320-330` の `SupportsReservationType`
- **問題**: ガード実装は存在するが動作証明テストが 0 件。兄弟メソッド `UpdateExcludedReservationTypes` には専用テスト（`reservation_staff_exclusion_clinic_isolation_test.go:40`）があり非対称。将来のリファクタで Count 検証が脱落しても CI が緑のまま。※既存の `reservation_staff_capability_preload_clinic_isolation_test.go` は Preload read 側のみで write 経路は未カバー。
- **変更内容**: exclusion 側テスト（:40-80）の構造を踏襲し `TestReservationStaffRepository_UpdateReservationCapabilities_ClinicIsolation` を追加: (1) clinicA 権限で clinicB の type ID → `WrapInvalidInput` エラーかつ junction 未挿入 (2) 自クリニック type ID → 成功 (3) 混在 [A,B] → エラーかつ部分挿入なし（DELETE 先行実行の巻き戻し確認）。同ファイルに `SupportsReservationType` の正/負 1 ケースずつを同乗。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run TestReservationStaffRepository_UpdateReservationCapabilities_ClinicIsolation -count=1` → PASS。実装 diff 0 行。
- **リスク / 戻し方**: 低（既存テストの型どおり）。`git reset --hard HEAD~1`。
- **依存**: R0

### R5. Lstep 購買判定クエリ 2 本の DB 実行テスト追加（旧 G11-5 改）

- **対象**: 新規 `backend/internal/repository/billing_item_lstep_queries_test.go`。テスト対象は `billing_item_repository.go:123-137` の `HasItemByOwnerSince` と `:166-181` の `HasFoodPurchaseByOwnerSince`
- **問題**: FEAT-383 系のフード切れ・再購入リマインド配信のターゲット抽出条件が DB で未検証（service 層 mock のみ）。
- **旧計画からのスコープ変更（重要）**: 旧 G11-5 は `FindOwnersByCategoryPurchaseDate`（:141-164）も対象としていたが、**本計画の再検証で同メソッドは本番呼び出し 0 件の死にコード**と判明した（interface 宣言 + 実装 + テスト mock のみ。Lstep Write API 一時停止(2026-05)由来の意図的休眠の可能性あり）。死にコードにテストを書いても無意味なため**対象外**とし、keep/delete 判断は §4 の要判断リストへ送る。同メソッドの doc コメント（:139 `issued_at`）と SQL（:156 `completed_at`）の乖離も、keep 判断が出るまで放置する。
- **変更内容**: (1) `TestBillingItemRepository_HasItemByOwnerSince`: names 一致 + completed_at>=since で true / since 前のみ・名前不一致・別 owner・別クリニック・soft-deleted billing で false / names 空で false。 (2) `TestBillingItemRepository_HasFoodPurchaseByOwnerSince`: names 指定時は name IN、未指定時は category=food フォールバックの分岐を両方固定。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run 'TestBillingItemRepository_(HasItemByOwnerSince|HasFoodPurchaseByOwnerSince)' -count=1` → PASS。実装 diff 0 行。
- **リスク / 戻し方**: 低。`git reset --hard HEAD~1`。
- **依存**: R0（R1 とは独立だが、同一ファイル群のため R3 の後に実施）

### R6. schema_drift_test の allModels 網羅化（登録 35 型追加 + 網羅性 lint）（旧 G12-1a）

- **対象**: `backend/internal/model/schema_drift_test.go:39-119`（allModels）、新規 `backend/internal/model/all_models_exhaustiveness_test.go`
- **問題**: モデル↔DDL 整合の唯一の機械ゲート `allModels()` が手動列挙 72 型で、`TableName()` 実装 107 型に対し **35 型が未登録**（AuditLog, PaymentSplit, CheckupFieldResult, CheckupTypeField, MedicalRecordAddendum, MedicineDoseParam, PetChronicCondition, PasswordResetToken, TokenBlacklist, LineLinkToken, LineSendLog, SharedFile, TrimmingCourseType, ManualArticle, ManualArticleVersion, Campaign, CampaignTargetCategory, CampaignTargetItem, ClinicIntegration, LabImportJob, LabImportEvent, ReservationTypeAvailableSlot, ReservationTypeOccupation, ReservationTypeUnavailableTime, Lstep 系 11 型）。監査・会計 payment_splits・健診結果という重要テーブルが検査網の外にある。
- **変更内容**:
  1. 新規テスト `TestAllModelsExhaustiveness`: `go/parser` で `internal/model/*.go`（非テスト）を走査し `func (x X) TableName() string` を実装する全 struct 名を列挙、`allModels()` の要素型名（`reflect.TypeOf(m).Elem().Name()`）と**双方向突合**（欠落/余剰で fail）。モデルが存在しない DB テーブル（例 `lstep_migration_progress`）は対象外のまま。先例: `model/audit_taxonomy_exhaustiveness_test.go` の go/ast パターン。
  2. `allModels()` に欠落 35 型を追記。
  3. `docker compose exec backend go test ./internal/model/ -run TestSchemaDrift -count=1` を実行し、新規登録分で drift が RED になった列は**モデル/DDL を直さず**、既存の allowlist 機構（無ければ `knownColumnDrifts` map を新設）に「なぜ許容するか + 別トラック ID」の根拠コメント付きで pin して緑化する（例: `audit_logs.ip_address` は DDL=inet vs model=string の既知乖離。修正は挙動変更のため旧計画 Appendix A X-3 の別トラック）。pin した列は完了報告に全件列挙する。
- **完了条件**: `docker compose exec backend go test ./internal/model/ -run 'TestSchemaDrift|TestAllModelsExhaustiveness' -count=1` → PASS。`internal/model` の非テストファイル diff 0 行。
- **リスク / 戻し方**: 未登録 35 型に予期しない drift が複数眠っている可能性 → pin が **5 列を超えた**場合はスコープ異常として中断・報告（本物のスキーマ問題が隠れている兆候）。戻しは `git reset --hard HEAD~1`。
- **依存**: R0

### R7. schema_drift_test の型カテゴリ厳格化 + nullability 検査追加（旧 G12-1b）

- **対象**: `backend/internal/model/schema_drift_test.go`（`isEnumLike` 許容ブロック :323-329、`pgTypeCategory` 相当の型分類、比較ループ）
- **問題**: (1) `isEnumLike` が PG 組込非デフォルト型（inet 等）vs text の乖離を握り潰す（:325 `if isEnumLike(goCategory) || isEnumLike(dbCategory) { continue }`）。(2) NOT NULL/nullability を一切比較しないため「Go=pointer かつ DB=NOT NULL」という実行時 INSERT 失敗クラスが検査されない。
- **変更内容**:
  1. `inet` 等の PG 組込型を明示カテゴリ化し、`isEnumLike` の許容集合から外す。これで RED になる列（`audit_logs.ip_address` を想定）は R6 と同じ allowlist に根拠コメント付きで pin。
  2. nullability 検査を追加: `migrator.ColumnTypes` の `Nullable()` と Go フィールドの pointer 性を比較し、「Go=pointer かつ DB=NOT NULL(デフォルト無し)」を drift として fail、「DB=nullable かつ Go=非 pointer」は warning リスト（fail しない）として出力。誤検知列は根拠コメント付き allowlist で pin。
- **完了条件**: `docker compose exec backend go test ./internal/model/ -count=1` → 全 PASS（audit_taxonomy 等の既存テスト含む）。pin 一覧を完了報告に列挙。
- **リスク / 戻し方**: nullability 検査の誤検知が大量に出る可能性 → **fail 対象の pin が 20 列を超えたら**検査の設計が実態に合っていない兆候として中断・報告。この項目単体で `git reset --hard HEAD~1` 可能（R6 とは別コミットのため独立に戻せる）。
- **依存**: **R6**（allowlist 機構と網羅登録が前提）

### R8. lab import 補償トランザクション失敗の無言破棄を解消（旧 G13-1）

- **対象**: `backend/internal/service/lab_result_import_service.go:131-132`
- **問題**: PersistBatch 失敗時の補償遷移（job→failed）が `_, _ = s.jobSvc.TransitionStatus(...)` でエラー無言破棄。補償自体が失敗（DB 断・5s タイムアウト）すると job が非終端状態で恒久 stuck するのに観測手段がない。同ファイル :172-177 の終端遷移は slog.ErrorContext + 根拠コメント付きで、非対称。
- **変更内容**（ログ追加のみ・挙動不変）:
  ```go
  // 変更前 (:131-132)
  _, _ = s.jobSvc.TransitionStatus(cleanupCtx, clinicID, jobID, model.LabImportJobStatusFailed,
      TransitionCounts{RowCount: len(inputs), ErrorCode: ptr("context_cancelled"), ErrorMessage: &errMsg})
  // 変更後
  if _, compErr := s.jobSvc.TransitionStatus(cleanupCtx, clinicID, jobID, model.LabImportJobStatusFailed,
      TransitionCounts{RowCount: len(inputs), ErrorCode: ptr("context_cancelled"), ErrorMessage: &errMsg}); compErr != nil {
      slog.ErrorContext(cleanupCtx, "lab result import: failed to transition to failed (compensation)",
          "error", compErr, "job_id", jobID)
      // 主エラーを優先して返す(挙動不変)。job は非終端で残るため jobID から追跡する。
  }
  ```
  併せて `lab_result_import_service_test.go` に「PersistBatch 失敗 + TransitionStatus(failed) も失敗」を注入する mock ケースを追加し、戻り値が従来どおり `"lab import batch interrupted"` の wrap のままであることを固定する。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestLabResultImportService' -count=1` → PASS（新ケース含む）
- **リスク / 戻し方**: 極小（ログ 1 箇所）。`git reset --hard HEAD~1`。
- **依存**: R0

### R9. masterFKWriteAllowlist の accountingService.Update 記録を実態に同期（旧 G14-1）

- **対象**: `backend/internal/service/master_fk_write_inventory_lint_test.go:191`、`backend/internal/service/cross_tenant_master_fk_write_test.go`（テスト追加）
- **問題**: allowlist :191 は `accountingService.Update` の `PaymentMethodID` を `statusKnownUnguarded` と記録するが、実装は `accounting_service_builders.go:38-40` で clinic-scoped 解決値と request 明示値の不一致を `WrapInvalidInput` 拒否しており **guarded 相当**。記録の過小評価は、本当に unguarded な他エントリへの注意を薄める。
- **変更内容**: 先に isolation テストで guard を動作証明し、その後に記録を書き換える（順序厳守）:
  1. `cross_tenant_master_fk_write_test.go` に `TestAccountingService_Update_RejectsForeignPaymentMethodID` を追加（既存 11 テストの構造踏襲）: 別クリニックの payment_method_id を明示指定した Update が invalid-input 拒否されることを確認。
  2. GREEN 化後、allowlist :191 の status を `statusGuarded` に変更し、reason を「resolvePaymentMethodMasterID の mismatch 拒否で validated; test: TestAccountingService_Update_RejectsForeignPaymentMethodID」に書き換え。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestMasterFKWriteInventory|TestAccountingService_Update_RejectsForeignPaymentMethodID' -count=1` → PASS
- **リスク / 戻し方**: テストが guard の穴を発見する可能性（その場合 status 変更は行わず BLOCKED 報告）。`git reset --hard HEAD~1`。
- **依存**: R0

### R10. G7-1 が残した孤児メソッド CountWorkingStaffByReservationTypeID の削除（新規 F2）

- **対象**: `backend/internal/repository/reservation_type_occupation_repository.go:27`（interface 宣言）と `:98`（実装）、および同メソッドを実装するテスト用 mock
- **問題**: G7-1（`755f3e42`）が唯一の本番呼び出し元（`liff_service_availability.go:129`）をバッチ版 `CountWorkingStaffByReservationTypeIDs`（複数形）へ切り替えた結果、単数形が呼び出し元 0 件の死にコードになった。リファクタ自身が作った孤児であり経緯が明確なため、§4 の「要判断」リストとは別に削除してよい。
- **変更内容**: interface 宣言・実装・専用テスト（あれば）・service 層 mock の該当メソッドを削除。**複数形 `...TypeIDs`（バッチ版）は現役なので絶対に触らない。** 削除前に `grep -rn 'CountWorkingStaffByReservationTypeID\b' backend/internal --include='*.go'` で単数形の残存参照が mock/テストのみであることを再確認する。
- **完了条件**: 上記 grep が 0 件（複数形は残存）。`docker compose exec backend go test ./internal/repository/ -run TestReservationTypeOccupation -count=1` と `docker compose exec backend go test ./internal/service/ -run 'TestLiff' -count=1` → PASS。`docker compose exec backend go build ./...` 相当はテスト実行で代替される。
- **リスク / 戻し方**: mock 削除漏れでコンパイルエラー → テスト実行で即検出。`git reset --hard HEAD~1`。
- **依存**: R0

### R11. G8-3 の取りこぼし: 複合日時リテラル 1 件を time.DateTime へ（新規 F4）

- **対象**: `backend/internal/service/lstep_csv_helpers.go:108`
- **問題**: G8-3（日時レイアウトリテラルの stdlib 定数統一）が `"2006-01-02 15:04:05"` を「複合レイアウトのため対象外」と一括りにしたが、この値は Go 1.20+ の `time.DateTime` と完全一致であり、隣接行（:109-110）と異なり定数が存在する。
- **変更内容**: `:108` のリテラルを `time.DateTime` に置換（値同一・挙動不変）。:109-110 の真の複合レイアウトは触らない。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'Csv' -count=1` → PASS
- **リスク / 戻し方**: 極小。`git reset --hard HEAD~1`。
- **依存**: R0

### R12. ConfigureTimeZone の JST 再導出を config.JST 再利用に統一（新規 F5）

- **対象**: `backend/internal/config/timezone.go:23-30`
- **問題**: パッケージ自身が「各呼び出し箇所で再導出せずキャッシュ済み `JST` を使え」とドキュメント（:11-14）した直後の `ConfigureTimeZone()` が `time.LoadLocation` を再実行している。`JST` は package init で panic-fail-fast 済みのため、この関数のエラーパスは到達不能。
- **変更内容**（シグネチャ不変・挙動不変）:
  ```go
  // 変更前
  func ConfigureTimeZone() error {
      loc, err := time.LoadLocation(JapanTimeZone)
      if err != nil {
          return fmt.Errorf("load %s timezone: %w", JapanTimeZone, err)
      }
      time.Local = loc
      return nil
  }
  // 変更後（JST は package init で panic 済みのためエラーパス到達不能）
  func ConfigureTimeZone() error {
      time.Local = JST
      return nil
  }
  ```
  `fmt` import が他で未使用になる場合は import も整理する。
- **完了条件**: `docker compose exec backend go test ./internal/config/ -count=1` → PASS
- **リスク / 戻し方**: 極小。`git reset --hard HEAD~1`。
- **依存**: R0

### R13. G3-3 未完遂分: 業務 repo の Update 系を updateScopedByID へ統合（新規 F1a）

- **対象**（12 サイト・全て `backend/internal/repository/`）: `estimate_repository.go:81` / `checkup_repository.go:149` / `inventory_repository.go:72` / `diagnosis_repository.go:70,202` / `examination_repository.go:111` / `permission_group_repository.go:72` / `reservation_type_liff_repository.go:68` / `reservation_staff_repository.go:87` / `closing_special_period_repository.go:80` / dbOrTx 変種: `shift_entry_repository.go:94` / `reservation_repository.go:170` / `accounting_repository.go:214`
- **問題**: G3-3 は 19 repo を `updateScopedByID`/`deleteScopedByID`（`helpers.go:70,86`）へ統合したがマスタ系で止まっており、ヘルパの doc コメントが「マスタ/業務テーブル」両対応を謳うのに、業務 repo 側に**バイト単位で同型**の手書きボディ（clinicScope + Where(id) + FromGORM + RowsAffected==0→WrapNotFound + 再取得）が残っている。
- **変更内容**: 各サイトを機械的にヘルパ呼び出しへ置換する。ヘルパは `*gorm.DB` を受けるため、dbOrTx 変種は `updateScopedByID(dbOrTx(ctx, r.db), ...)` の形で ambient tx 参加を**維持**する（`dbOrTx` トークンがメソッド本体から消えないため `dbortx_inventory_lint_test.go` も緑のまま）。**サイト毎の必須手順**: 置換前に既存ボディとヘルパの挙動を突合し、①RowsAffected==0 の扱い ②更新後再取得（FindByID）の有無と Preload ③`Select`/`Omit` 句の有無 が完全一致するサイト**のみ**置換する。一つでも差異があるサイトはスキップし、完了報告に「スキップ + 理由」を記載する。
- **除外（触ってはならない）**: `treatment` / `care_plan_item` / `clinical_plan`（サブクエリ隔離 — GORM Joins 制約のコード内文書化済み）、`billing_item`（JOIN スコープ）、`medical_record_repository.go` の Update（draft-status 条件 + Conflict 変換）、`reservation_type_liff` の Delete（FK-conflict 変換）、`pet_chronic_condition_repository.go`（§4 F3 参照 — RowsAffected 検査が無いため置換は挙動変更になる）。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -count=1` → 全 PASS（dbortx lint・isolation テスト群を含む）。置換各サイトの diff がボディ置換のみであること。
- **リスク / 戻し方**: 中。挙動差の見落とし（特に再取得の Preload 有無）が isolation/tx テストで検出されない可能性 → 上記サイト毎突合を厳守。R2〜R5 で追加したテストが accounting/reservation_staff の退行を追加検出する。`git reset --hard HEAD~1`。
- **依存**: R2, R3, R4, R5（保護テストを先に敷く）

### R14. G3-3 未完遂分: 業務 repo の Delete 系を deleteScopedByID へ統合（新規 F1b）

- **対象**（5 サイト）: `estimate_repository.go:95` / `checkup_repository.go:164` / `inventory_repository.go:87` / `medical_record_repository.go:238` / `appointment_admin_repository.go:101`（SoftDelete）
- **問題・変更内容・除外**: R13 と同一の統合・同一の手順・同一の除外リスト。**注意**: `medical_record_repository.go:238` の Delete は `Where("id = ? AND status = ?", id, model.MedicalRecordStatusDraft)` のような追加述語を持つ場合ヘルパ対象外（R13 の突合手順で自動的にスキップ判定される — draft 条件があるのは Update 側だが、Delete 側も置換前に必ず実体を読むこと）。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -count=1` → 全 PASS。
- **リスク / 戻し方**: R13 と同じ。`git reset --hard HEAD~1`。
- **依存**: R13（同一ファイル群の連続変更のため直後に実施）

### R15. api.yaml: Billing スキーマに medical_record ネストプロパティを追記（旧 G1-3 残差）

- **対象**: `backend/docs/api.yaml:1455-1566`（Billing スキーマ）
- **問題**: G1-3（`9e83aa4b`）で total_refunded_amount / payment_splits は追記済みだが、model `accounting.go:87` の `MedicalRecord *MedicalRecord json:"medical_record,omitempty"`（Preload 時のみ直列化）に対応するプロパティだけが未記載のまま残った。
- **変更内容**: Billing スキーマに `medical_record: { $ref または既存 MedicalRecord スキーマ参照, nullable: true, readOnly: true, description: Preload 時のみ }` を追加。既存の payment_splits 追記（:1555-1560）と同じ流儀に合わせる。実装コードは触らない。
- **完了条件**: `docker compose exec backend go test ./internal/apicontract/ -count=1` → PASS（route gate / date-format gate とも）
- **リスク / 戻し方**: YAML 構文ミス → gate テストのパース失敗で即検出。`git reset --hard HEAD~1`。
- **依存**: R0（api.yaml 系は R15→R19 の順で直列実施。並行しない）

### R16. api.yaml: masters GET {id} 系ほか 23 オペレーションを文書化（旧 G1-2 Phase E1）

- **対象**: `backend/docs/api.yaml`、`backend/internal/apicontract/openapi_route_drift_test.go` の `knownMissingFromSpec`
- **問題**: 実装済みルートの api.yaml 未記載が allowlist に 63 件 pin されて残っている（G1-2 Phase A〜D 実施後の残差）。本項目はうち masters 系 23 件を解消する。
- **変更内容**: 以下 23 オペレーションを api.yaml に追記し、**対応する allowlist エントリを削除**する（stale 検出があるため、エントリを消さないとテストが fail する — 完了条件が自己検証になっている）。レスポンス形は対応する `*_response.go` の json タグから起こし、既存スキーマの `$ref` を再利用する:
  - `GET /masters/{animal-species,cages,checkup-types,consultations,diagnosis-names,diagnosis-types,examination-types,hospitalization-plans,insurances,medicines,merchandise-items,occupations,procedures,staffs,trimming-courses,trimming-options,vaccines}/{id}`（17）
  - `GET /masters/checkup-types/{id}/fields`、`GET /masters/diagnosis-names/all`、`GET /masters/staffs/{id}/permission-groups`、`GET /masters/medicines/{id}/dose-params`（4）
  - `PUT /masters/medicines/{id}/dose-params/{species}`、`DELETE /masters/medicines/{id}/dose-params/{species}`（2）
  - allowlist ヘッダーコメント（:54-77）の残差件数の記述も実数に更新する。
- **完了条件**: `docker compose exec backend go test ./internal/apicontract/ -count=1` → PASS
- **リスク / 戻し方**: format: date の新規プロパティを追加すると date-format gate の床値/allowlist に影響しうる → 日付フィールドは既存スキーマの表現をコピーする。`git reset --hard HEAD~1`。
- **依存**: R15

### R17. api.yaml: reservation-staffs / reservations / shifts 系 16 オペレーションを文書化（旧 G1-2 Phase E2）

- **対象**: R16 と同じ 2 ファイル
- **変更内容**: 以下 16 件を追記し allowlist から削除:
  - `/clinics/{clinic_id}/reservation-staffs`: GET, POST / `{staffId}`: PUT, DELETE / `{staffId}/sort-order`: PATCH / `{staffId}/status`: PATCH / `{staffId}/image`: POST / `{staffId}/schedules`: GET / `{staffId}/schedules/{date}`: PUT, DELETE（10）
  - `/clinics/{clinic_id}/reservations`: GET, POST / `{reservationId}`: DELETE（3）
  - `GET /reservations/available-times`、`PATCH /reservations/{id}/reservation-route`、`GET /shifts/on-duty-staffs`（3）
- **完了条件・リスク・戻し方**: R16 と同一。
- **依存**: R16

### R18. api.yaml: checkups / pets / medical-records / hospitalizations / owners 系 9 オペレーションを文書化（旧 G1-2 Phase E3）

- **対象**: R16 と同じ 2 ファイル
- **変更内容**: 以下 9 件を追記し allowlist から削除:
  - `GET /checkups`、`GET /checkups/alerts`、`GET /checkups/field-results`（3）
  - `PATCH /clinics/{clinic_id}/pets/{id}/death`、`DELETE /clinics/{clinic_id}/pets/{id}/death`（2）
  - `PATCH /medical-records/{id}/recommendation-reason`、`POST /medical-records/{id}/images/upload`（2）
  - `POST /hospitalizations/{id}/discharge-with-billing`、`GET /clinics/{clinic_id}/owners/aggregations`（2）
- **完了条件・リスク・戻し方**: R16 と同一。
- **依存**: R17

### R19. api.yaml: lstep 系 + LINE webhook 15 オペレーションを文書化（旧 G1-2 Phase E4・残差ゼロ化）

- **対象**: R16 と同じ 2 ファイル
- **変更内容**: 以下 15 件を追記し allowlist から削除:
  - `/clinics/{clinic_id}/lstep-settings`: GET, PATCH, DELETE / `lstep-settings/test-connection`: POST（4）
  - `/clinics/{clinic_id}/lstep-tag-code-mappings`: GET / `{tag_name}`: PUT（2）
  - `/clinics/{clinic_id}/lstep/checkup-sync`: POST / `checkup-sync/preview`: GET（2）
  - `/clinics/{clinic_id}/lstep/delivery-monitor/{logs,summary}`: GET（2）
  - `/clinics/{clinic_id}/lstep/owners`: GET / `lstep/tag-summary`: GET / `lstep/trigger-priorities`: GET, PATCH（4）
  - `POST /api/line/webhook`（1）— servers prefix `/api/v1` の外にあるため、LIFF ルート（`ee3f238a` で文書化済み）と同じ手法（別 servers エントリまたは絶対パス）で記載する。
  - この時点で allowlist の残存は**意図的 pin のみ**（同一ハンドラの別名 15 件・`/health` 相対規約 1+1 件・`clinics/{id}` パラメータ名差 3+3 件・HTTP verb map）になる。ヘッダーコメントを「残差ゼロ・意図的 pin のみ」に書き換える。
- **完了条件**: `docker compose exec backend go test ./internal/apicontract/ -count=1` → PASS。allowlist の非意図的エントリが 0 件。
- **リスク・戻し方**: R16 と同一。
- **依存**: R18

### R20. プロトタイプ期の虚偽サンプル文書 2 点を削除（旧 G1-5）

- **対象**: `backend/docs/postman-collection.json`、`backend/docs/api-examples.md`
- **問題**: 両ファイルは 2026-01-26 のプロトタイプ期の遺物で、UUID ID（実際は uint64）・API キー認証（実際は Cookie+JWT）・存在しない `/medical-records/paginated`・実在しない `PUT /pets/{id}`（実際は PATCH）を案内する。「実行すると全て失敗する」使用例が正本のふりをして残っている。リポジトリ全体で参照 0 件（旧 BE-refactor.md 内の言及のみ）を検証済み。
- **変更内容**: `git rm backend/docs/postman-collection.json backend/docs/api-examples.md`。docs/README.md は既に 2 ファイル構成（README.md + api.yaml）として書かれているため変更不要。Postman コレクションが将来必要なら api.yaml から自動生成できる。
- **完了条件**: `ls backend/docs/` → `README.md` と `api.yaml` のみ。`grep -rn 'postman-collection\|api-examples' backend/ --include='*.go' --include='*.md' --include='Makefile'` → 0 件（BE-refactor*.md 除く）。
- **リスク / 戻し方**: なし（参照ゼロ確認済み）。`git revert` で完全復元可能。
- **依存**: R0

### R21. backend/README.md と CODING_RULES.md を実態に同期（旧 G1-6）

- **対象**: `backend/README.md`（:8, :44-45, :100-110, :116-244 付近, :310, :336）、`backend/CODING_RULES.md`（:14-67）
- **問題**: オンボーディング文書が現行必須規約と正反対のコードを教えている:
  - README:221 `c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})`（:310 に StatusNotFound 変種も）— P7/P12 で明示禁止のパターン
  - README:234-235 `v1.GET("/owners", h.GetOwners)` — RequirePermission 無し登録（P5 違反）
  - README:132・:336 の `uuid.UUID` 主キー例 — 実モデルは全て uint64
  - README:243 `db.AutoMigrate(...)` を手順として指示 — 実運用は migrations/ SQL 管理
  - README:44-45 存在しない `internal/validation/`、README:109 `PUT /pets/:id`（実装は PATCH）、README:8 `Gin v1.10`（go.mod 実測 v1.12.0）
  - CODING_RULES:16-18 cmd/ を api のみと記載（実際 7 サブディレクトリ）、internal/ 一覧に `apicontract` / `dbconn` / `infra` / `seedbundle` の 4 パッケージが欠落
- **変更内容**（ドキュメントのみ）:
  1. README の「新しい CRUD 機能の追加手順」セクション（教材コード全体）を削除し、`CODING_RULES.md` と `internal/*/CLAUDE.md` への参照 1 段落に置換（教材コードの二重管理を解消）。
  2. README と CODING_RULES のディレクトリツリーを実態に同期（internal/ に apicontract・dbconn・infra・seedbundle を追記、internal/validation を削除、cmd/ 7 サブディレクトリを列挙）。
  3. README のエンドポイント表は「docs/api.yaml 参照」1 行に置換（または PUT→PATCH 等を実態修正）。Gin バージョンを v1.12.0 に修正。
- **完了条件**: `grep -n 'internal/validation\|uuid.UUID\|AutoMigrate\|gin.H{"error"' backend/README.md` → 0 件。`grep -n 'apicontract\|dbconn\|seedbundle' backend/CODING_RULES.md` → 各 1 件以上。ランタイム検証不要（ドキュメントのみ）。
- **リスク / 戻し方**: なし。`git revert`。
- **依存**: R0（R20 の後に実施すると docs/ 構成の記述が確定していて書きやすい）

---

## §4 別トラック（本計画では実行しない・記録のみ）

実行者はこのセクションに**一切手を付けない**。完了報告に「未着手・別トラック」として転記するだけでよい。

| ID | 内容 | 実行しない理由 |
|---|---|---|
| G6-2 | repo 内部 tx 13 ファイルの dbOrTx 化 + tx 規約 CLAUDE.md 追記 | 旧計画 Appendix A の tx 非参加 3 件（挙動変更）修正後でないと同一ファイル競合。BLOCKED |
| G9-1 | main.go 二段階 DI の単一化 | Appendix A `lstep-nilcipher-stale-di`（挙動変更）が前提。BLOCKED |
| X-1〜X-18 | 旧 BE-refactor.md Appendix A の挙動変更 18 件（SanitizeNullBytes バイナリ破壊・tx 非参加・楽観ロック等） | PO/責任者判断を要する挙動変更。本計画は挙動保存のみ |
| H-1 | UpdateStaffGroups の staff_id 単位 DELETE が多施設所属スタッフの他クリニック紐付けを削除しうる | 挙動変更の修正。別チケット推奨（HIGH） |
| F3 | `pet_chronic_condition_repository.go:61,71` の Update/Delete に RowsAffected==0→NotFound 検査が無い（パッケージ内で唯一） | ヘルパ統合するとレース窓の挙動が「無言成功→NotFound」に変わる＝厳密には挙動変更 |
| F6 | 死にコード群の keep/delete 判断: `LstepTagService.BulkAddOwnerTag`（G7-5 が 3 日前に最適化した対象が実は死んでいる）、`SyncPetSpeciesTags`、`SyncSeniorTag`、`SyncHealthcheckTags`/`SyncVaccineDeadlineTag`（plain wrapper）、`FindOwnersByCategoryPurchaseDate`（+ :139 コメントと SQL の issued_at/completed_at 乖離）、`HasReservationByOwnerInRange`、`GetDeliveryHistoryByOwner`、`CountByTag`、`BulkReplaceOwnerTags`、`BulkCreate`(friend snapshot)、`FindByClinicAndDate`(delivery trigger log) | Lstep Write API 一時停止（2026-05）由来の意図的休眠の可能性があり、削除はオーナーの機能ロードマップ判断。盲目的削除禁止 |
| B-2 | Preload read-lint 未登録の 3 マスタ | model に GORM association が無く構文的に登録不能（設計変更が前提） |

## §5 やらないことリスト（禁止事項）

実行者が「善意で」やりがちな逸脱を先回りで禁止する。以下のいずれかが必要だと感じたら、実行せず理由を添えて報告すること。

1. **挙動変更の一切**: バグを見つけても実装コードを直さない。§3 冒頭の BLOCKED プロトコルに従う。テスト側を実挙動に合わせて曲げて緑にするのも禁止（それはバグの隠蔽）。
2. **§4 の別トラック項目への着手**（G6-2 / G9-1 / X 系 / H-1 / F3 / F6 の死にコード削除を含む）。
3. **機能追加・仕様変更・新規エンドポイント追加・DB migration の新規作成/編集**。
4. **依存ライブラリの追加・更新**（go.mod / go.sum を変更しない。kin-openapi 等の validator 追加も不要）。
5. **route drift allowlist の意図的 pin の解消**: 同一ハンドラ別名 15 件・`/health` 1+1 件・`clinics/{id}` パラメータ名差 3+3 件は設計判断済みの pin。api.yaml のパラメータ名を `{clinic_id}` へ書き換える「修正」もしない。
6. **api.yaml の既存 `format: date` プロパティの型・format 変更**（date-format gate の allowlist 22 件が壊れる）。新規追記時は既存スキーマの日付表現をコピーする。
7. **MedicalRecordRepository のサブインターフェース分割**（旧 G2-2 の任意ステップ。YAGNI として不採用が確定済み）。
8. **旧 `BE-refactor.md` の編集**。進捗は本ファイル（BE-refactor-v2.md）の項目見出しに `✅ DONE (commit hash)` を追記する形でのみ記録する。
9. **`.claude/` 配下のファイルのステージ・コミット**（作業ツリーに未コミット変更が見えても無視）。
10. **フルリポジトリ検証の実行**: `go test ./...` / `golangci-lint run ./...` / `gofmt -w ./...` / `pnpm *` / `make codegen` / `docker compose up|down|restart` / DB リセット。例外は R1 の `DROP DATABASE ekarte_db_test` のみ。
11. **push・PR 作成・Issue 投稿・外部サービスへの書き込み**。コミットはローカルに留める。
12. **コミットの squash・rebase・amend**（1 項目 1 コミットの履歴をそのまま残す）。

## §6 実行者への指示文（このままコピペして渡す）

```
あなたはこのリポジトリのバックエンドリファクタリング実行者である。
リポジトリルートの BE-refactor-v2.md が唯一の作業指示書である。以下を厳守せよ。

1. まず BE-refactor-v2.md を全文読む。次に backend/internal/repository/CLAUDE.md、
   backend/internal/service/CLAUDE.md、backend/internal/handler/CLAUDE.md を読む
   （テストコードの流儀・禁止パターンの正本）。
2. §2 の R0（安全網）を最初に実行する。ベースラインが 1 つでも赤なら着手せず報告して終了。
3. §3 の作業項目を R1 から R21 まで番号順に、1 項目ずつ実施する。
   - 1 項目 = 1 コミット。コミット後に次の項目へ進む。並行作業禁止。
   - 各項目の「完了条件」のコマンドを実行し、期待結果を満たさない限りコミットしない。
   - 変更した Go ファイルは gofmt -l（backend コンテナ内）が無出力であることも確認する。
   - 完了条件を満たせない場合: 変更を破棄（git checkout -- <files>）し、その項目を
     SKIP/BLOCKED として理由を記録し、依存されていない次の項目へ進む。
     依存先が SKIP された項目（例: R1 が失敗した場合の R3）は自動的に SKIP。
4. テストが実装の実バグを発見した場合（期待と実挙動の食い違い）:
   実装を直すな。テストを曲げるな。項目を BLOCKED とし、失敗テスト・実挙動・
   証拠(file:line)を記録して次へ進む。
5. §4（別トラック）と §5（やらないこと）に該当する作業は、必要だと感じても実行するな。
6. push するな。コミットはローカルに残す。
7. 全項目終了後、以下を含む完了報告を書く:
   - 項目ごとの DONE/SKIP/BLOCKED とコミットハッシュ
   - R6/R7 で allowlist に pin した列の全件リスト
   - R13/R14 でスキップしたサイトと理由
   - 発見した実バグ（BLOCKED 案件）の一覧
   - ベースラインと完了時点のテスト結果比較
```

---

## 付録: 実行順の依存関係まとめ

```
R0 ─┬─ R1 ──────────── R3 ─┐
    ├─ R2 ─────────────────┤
    ├─ R4 ─────────────────┼─ R13 ─ R14
    ├─ R5 ─────────────────┘
    ├─ R6 ─ R7
    ├─ R8, R9, R10, R11, R12（相互独立・番号順推奨）
    ├─ R15 ─ R16 ─ R17 ─ R18 ─ R19（api.yaml 直列）
    └─ R20 ─ R21（docs 整理）
```

- R1→R3: テスト DB の `item_source` enum に `'trimming'` が必要。
- R2/R3/R4/R5→R13: ヘルパ統合の前に対象ファイル群の保護テストを敷く。
- R6→R7: allowlist 機構と全モデル登録が型検査厳格化の前提。
- R15→…→R19: 同一ファイル（api.yaml + allowlist）の連続編集のため直列。
- R20→R21: docs/ の最終構成を確定させてから README を同期。

