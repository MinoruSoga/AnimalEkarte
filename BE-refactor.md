# BE-refactor.md — バックエンド リファクタリング計画書（第7期）

- **作成日**: 2026-07-13
- **基準コミット**: `9aeee96d`(main)。行番号はずれたら**シンボル名で再特定**する。
- **性格**: 本書は実行計画の正本。判断できない事態は**中断して報告**。本書とコード以外の文脈を前提にしない。
- **別台帳**(本書と重複させない): PERF・SEED残 = `BE_todo.md` / 任意検証 = `BE-pending.md` / 既知バグ skip 台帳 = `BE_todo.md` 末尾（#236 で修正予定、**本書の対象外**）
- **進捗**: 対応済 12 / 22。残存 BE7-13〜BE7-21（部分対応: BE7-15）。棚卸し基準 HEAD `e424de5c`（2026-07-14）。各項目完了時に見出しの `[ ]` を `[x]` に更新してよい（同一コミットに本書を含めてよい）。

---

## 1. 現状理解（実行者への文脈共有）

### 1.1 全体像

動物病院向け電子カルテ（EMR）のバックエンド。Go 1.25 / Gin / GORM / PostgreSQL 18、マルチテナント（全データ `clinic_id` 分離）。

- **エントリポイント**: `cmd/api/main.go` — config → logger → DB → repositories → services（単一段階DI）→ handler → cron バッチ3本（no-show 毎時 / dormant 毎日02:00 / delivery trigger 毎時）→ Gin ルーター → graceful shutdown。
- **デプロイ**: `backend/worker/index.ts` は Cloudflare Workers の薄い TS プロキシ（Container 内の Go バイナリへ全フォワード、ビジネスロジックなし）。**Go の監査対象ではない**。
- **レイヤ**: `internal/handler`（`*_request.go`/`*_response.go`、P5-P18） → `internal/service`（`validators_*.go`、P1-P17、slog はこの層のみ） → `internal/repository`（P2-P16、`clinicScope`/`dbOrTx`） → `internal/model`。層別規約は各層の `CLAUDE.md` が正本。
- **エラー**: `internal/errors`（apperrors、`FromGORM` で GORM/pgx→AppError 変換）→ handler の `RespondError` が唯一の応答経路。
- **tx 参加は2系統**: ctx-txKey 方式（`dbOrTx`）と repo-swap 方式（`Repositories.Transaction`）。`r.db.WithContext(ctx).Transaction(...)` は ambient tx 非参加バグパターンで**禁止**。

### 1.2 機械強制ゲート（壊してはならない）

CI で以下が gating 済み。**該当層に触る項目は完了条件でゲートのローカル実行を要求する**:

| ゲート | ローカル実行コマンド |
|---|---|
| P3.1 Preload clinic-scope lint | `docker compose exec backend go test ./internal/repository/ -run TestPreloadClinicScope -count=1` |
| master-FK write inventory | `docker compose exec backend go test ./internal/service/ -run TestMasterFKWriteInventory -count=1` |
| 監査 tx inventory | `docker compose exec backend go test ./internal/repository/ -run TestClinicalResultAuditTxInventory -count=1` |
| dbOrTx inventory | `docker compose exec backend go test ./internal/repository/ -run TestDBOrTxInventory -count=1` |
| ルート snapshot | `docker compose exec backend go test ./internal/handler/ -run TestRouteSnapshot -count=1`（run 名は実ファイル `handler_routes_snapshot_test.go` で確認） |
| OpenAPI date-format / route drift | `docker compose exec backend go test ./internal/apicontract/ -count=1` |
| coverage-ratchet | CI のみ（baseline 91.3%）。**テストの削除は原則禁止**（BE7-14 の orphan 削除に伴う対応テスト削除のみ例外、カバレッジは分母ごと減るため低下しない） |

**含意**: clinic_id 隔離・ルート登録漏れ・API ルート drift は機械強制済み。本計画はゲートが**見ない**領域（フィールドレベル API 乖離・重複・デッドメソッド・エラー処理の穴・層内の責務混在）に絞る。

### 1.3 命名・規約の到達点

P16（repository メソッド命名）/ P17（service Input 命名）/ レシーバ名 / ファイル名 snake_case は今回の全数監査で**違反0件**（過去期で収束済み）。命名系の項目は本計画にない。

---

## 2. 実行者が守る規約

### 2.1 検証コマンド（Docker スコープ限定）

- テストは **`docker compose exec backend go test ./internal/<pkg>/ -run <TestName> -count=1`**。フル `go test ./...` / `golangci-lint run ./...` / `gofmt -w ./...` は**自動実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/`（および `./cmd/api/`）が**無出力**であることを確認してからコミット。
- lint が必要な場合はスコープ限定: `docker compose run --rm --no-deps --entrypoint golangci-lint backend run ./internal/<pkg>/...`。

### 2.2 Git

- main 直作業。**push / PR 禁止**。`Co-Authored-By` を入れない。
- **1項目=1コミット**。メッセージは `fix(backend):` / `refactor(backend):` / `test(backend):` / `docs(backend):`（各項目に指定あり）。
- `git add` は**ファイル指定のみ**（`git add -A` / `git add .` 禁止）。

### 2.3 dirty ファイル（触らない）

- `FE-refactor.md`（FE 第6期計画書・別ワークストリーム）は**変更・コミット・stash 禁止**。
- 例外: 本書 `BE-refactor.md` の進捗欄更新のみ可。

---

## 3. BE7-0 [x] 安全網の構築（最初に実行）

1. `git status --porcelain` の dirty が `FE-refactor.md` ＋本書のみであることを確認。想定外があれば**中断して報告**。
2. `git rev-parse --short HEAD` を記録（戻し先）。
3. §1.2 の機械ゲート6本をローカル実行し、**全PASS を記録**。1本でも赤なら着手せず中断報告。
4. **各項目に着手する直前**、その項目の完了条件のテストコマンドを**変更前に1回実行**する。変更前から赤の場合はその項目をスキップし、理由を最終報告に含める。
5. BE7-0 でのコミットは不要。以後の戻し方は原則 `git revert <該当コミット>` または `git checkout -- <対象ファイル>`（未コミット時）。

---

## 4. 作業項目（実行順）

### Phase 1 — fix: 実バグとエラー処理の穴

#### BE7-1 [x] fix: 会計明細カテゴリ4値が service 層で不正拒否される

- **棚卸し（2026-07-14）**: **完了** — `TestValidateItemCategory` は model の12定数＋`unknown`/`invalid_category` を検証。`validateItemCategory` は12値を許可。
- **対象**: `internal/service/validators_accounting.go`（`validateItemCategory`）
- **問題**: `internal/model/accounting.go` の `ItemCategory` は12定数、handler の binding（`internal/handler/accounting_request.go` の oneof）も12値を許可するが、`validateItemCategory` は8値のみで **`vaccine` / `trimming` / `hotel` / `training` が「明細カテゴリの値が不正です」で拒否される**。`billing_item_service.go` の `CreateItem` から無条件で呼ばれる到達可能な実バグ（4カテゴリ追加時の service 側更新漏れ）。
- **どう変える**（テスト先行）:
  1. `internal/service/validators_accounting_test.go` を新設（なければ）し、table-driven テストを追加: `テスト名: TestValidateItemCategory` — model の12定数全てが nil を返すこと＋不正値 `"unknown"` がエラーになること。この時点で4値分が **RED**。
  2. `validateItemCategory` の case に `model.ItemCategoryVaccine, model.ItemCategoryTrimming, model.ItemCategoryHotel, model.ItemCategoryTraining` を追加して GREEN。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestValidateItemCategory|TestBillingItem' -count=1` 全PASS。
- **リスク / 戻し方**: 低（許可集合の拡大のみ。縮小ではないため既存データに影響なし）。`git revert`。
- **コミット**: `fix(backend): 会計明細カテゴリvaccine/trimming/hotel/trainingのservice層拒否を修正`
- **依存**: BE7-0

#### BE7-2 [x] fix: 全クリニックバッチで「全滅クリニック」の監査記録とエラー内容が欠落する

- **対象**: `internal/service/lstep_batch_service.go:97-114`（`runBatchAllClinics`）
- **問題**: (a) `perClinic` が返す `errs` の**中身が一切ログに出ない**（`error_count` の数字のみ）。(b) `count == 0 && len(errs) > 0`（そのクリニックの処理が全滅）の場合、`if count > 0` ガードで audit_logs への記録が**丸ごとスキップ**される。no-show / dormant / delivery-trigger の cron 3本すべてがこの経路で、あるクリニックのバッチ全滅が運用から観測不能。
- **どう変える**:

```go
count, errs := perClinic(ctx, clinic.ID)
if len(errs) > 0 {
    // 先頭3件のエラー本文を残す（全件はログ肥大のため打ち切り、件数は error_count が持つ）
    sample := errs
    if len(sample) > 3 {
        sample = sample[:3]
    }
    msgs := make([]string, 0, len(sample))
    for _, e := range sample {
        msgs = append(msgs, e.Error())
    }
    slog.ErrorContext(ctx, label+": partial errors",
        "clinic_id", clinic.ID, "error_count", len(errs), "errors", msgs)
}
if count > 0 || len(errs) > 0 {   // 全滅クリニックも監査に残す
    ...既存の audit 記録（meta の processed_count/error_count は既存のまま）...
}
```

  `count > 0` 専用だった `slog.InfoContext(... syncedSuffix ...)` は `count > 0` 条件のまま残す（成功ログの意味を変えない）。
  テスト追加（`lstep_batch_service` の既存テストファイルに、既存のモック流儀で）: `テスト名: 全滅クリニックでも監査ログが記録されエラー内容がログに出る` — `perClinic` が `(0, []error{...})` を返すケースで audit 呼び出しを assert。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestRunBatchAllClinics|TestLstepBatch' -count=1` 全PASS（run 名は既存テストに合わせて調整可、ただし新ケースを含むこと）。
- **リスク / 戻し方**: 低（ログ・監査記録の追加のみ。業務処理は不変）。`git revert`。
- **コミット**: `fix(backend): バッチ全滅クリニックの監査記録欠落とエラー内容非表示を修正`
- **依存**: BE7-0

#### BE7-3 [x] fix: Lステップ疎通確認の復号エラーが無言で握りつぶされる

- **対象**: `internal/service/lstep_settings_connection.go:23`（`val, _ := s.decrypt(...)`）
- **問題**: 同じ `s.decrypt` を呼ぶ兄弟2箇所（`lstep_settings_credentials.go:35` / `lstep_settings_service.go:177`）は `decErr` を `slog.ErrorContext` でログしてから空文字フォールバックするが、ここだけ無言。暗号キーのローテーション後や暗号文破損時、**疎通確認（診断機能）自体が原因を隠して**空文字キーでテストし失敗する。
- **どう変える**: 兄弟2箇所と**同型**に変更（`val, decErr := s.decrypt(...)` → `if decErr != nil { slog.ErrorContext(...) }` → 空文字フォールバック維持）。ログの属性名・文言は `lstep_settings_credentials.go:35` 側の実装をそのまま踏襲する。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run TestLstepSettings -count=1` 全PASS。`rg -n "s\.decrypt" internal/service/lstep_settings_*.go` で `_ :=` 破棄が0件。
- **リスク / 戻し方**: なし（ログ追加のみ）。`git revert`。
- **コミット**: `fix(backend): Lステップ疎通確認の復号エラー無言破棄をログ付きフォールバックに統一`
- **依存**: BE7-0

#### BE7-4 [x] fix: LIFF 空き日程の休診設定パースが fail-open（壊れた設定＝休診日なし扱い）

- **棚卸し（2026-07-14）**: **部分対応** — `ParseAvailableDatesSettings` / `TestParseAvailableDatesSettings` は存在するが、unmarshal 失敗時は依然 fail-open（nil 代入・error 常に nil）。既存テストも fail-open を期待。caller のエラー伝播も未達。本文は残す。
- **対象**: `internal/service/available_dates.go`（`ParseAvailableDatesSettings`）と唯一の本番呼び出し元 `internal/service/liff_service_availability.go`
- **問題**: `closed_weekdays` / `closed_dates` の JSON unmarshal 失敗を `nil` 代入で握りつぶし「休診日なし」として返す（**fail-open**）。関数シグネチャは `error` を返せる形なのに常に nil。実際の予約作成は `reservation_validators.go` 側で fail-closed 済みのためデータ破損はしないが、LINE 予約画面が**休診日を予約可能として表示**する UX バグ経路であり、過去に CRITICAL 修正した fail-open クラスの残存兄弟。
- **どう変える**:
  1. unmarshal 失敗時に `AvailableDatesSettings{}, apperrors.WrapInvalidInput(...)`（文言例: `"休診設定の解析に失敗しました"`、元エラーを wrap）を返す（fail-closed）。
  2. `liff_service_availability.go` の呼び出し元でエラーを伝播（この層の既存エラー伝播流儀に合わせる。LIFF 側は既存の RespondError 経路でエラー応答になる）。
  3. テスト追加: `テスト名: 壊れたclosed_weekdays JSONはエラーを返す`（available_dates の既存テストファイルに）。既存の正常系テストは無変更で通ること。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestParseAvailableDates|TestLiffServiceAvailability' -count=1` 全PASS（run 名は既存に合わせ調整可）。
- **リスク / 戻し方**: 中（挙動変更: 設定破損時に空き日程APIがエラーを返すようになる。正常データでは不変）。既存テストが落ちる場合は呼び出し元の期待を確認し、**仕様疑義があれば中断して報告**。`git revert`。
- **コミット**: `fix(backend): LIFF空き日程の休診設定パースをfail-openからfail-closedに変更`
- **依存**: BE7-0

#### BE7-5 [x] fix: panic recovery 漏れの goroutine 2箇所

- **対象**: `cmd/api/main.go:192` 付近（`server.ListenAndServe()` の goroutine）/ `internal/middleware/rate_limit.go:32`（`go s.cleanupLoop(ctx)`）
- **問題**: 第6期 A-1 で業務系 goroutine は `goSafe`（`internal/service/go_safe.go`、service パッケージ私有）と `runScheduledIteration` で保護されたが、この2本は対象外のまま。特に rate_limit の cleanupLoop は panic すると**プロセスは生きたまま cleanup だけ永久停止**し、limiter マップがリークし続ける（検知が遅い障害クラス）。
- **どう変える**: `goSafe` は service パッケージ私有のため import しない。各所に**インラインの defer-recover** を最小追加する:

```go
// cmd/api/main.go — ListenAndServe goroutine の先頭に
go func() {
    defer func() {
        if r := recover(); r != nil {
            logger.Error("server goroutine panic", slog.Any("panic", r))
        }
    }()
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed { /* 既存 */ }
}()

// internal/middleware/rate_limit.go — cleanupLoop のループ本体を per-iteration で保護
// （evict 1回分を safeEvict() に包み、recover 後もループが継続すること）
func (s *RateLimitStore) safeEvict() {
    defer func() {
        if r := recover(); r != nil { /* slog.Error で panic 内容を記録 */ }
    }()
    s.evict()
}
```

  （実際のフィールド名・関数名・ticker 構造は現物に合わせる。recover 後もループが継続することが要件。）
- **完了条件**: `docker compose exec backend go test ./internal/middleware/ -run TestRateLimit -count=1` 全PASS。`docker compose exec backend go build ./cmd/api/` 成功。`rg -n "go func|go s\." cmd/api/main.go internal/middleware/rate_limit.go` の各 goroutine に recover 経路があることを目視確認。
- **リスク / 戻し方**: 低（正常系は素通り）。`git revert`。
- **コミット**: `fix(backend): ListenAndServeとrate-limit cleanupのgoroutineにpanic recoveryを追加`
- **依存**: BE7-0

### Phase 2 — refactor: 重複の統合

#### BE7-6 [x] refactor: repository ページネーション定型文（約18ファイル）を共有 Scope に集約

- **対象**: `internal/repository/base.go`（ヘルパ新設）＋ `Offset((page-1)*limit).Limit(limit)` イディオムを手書きする repository 群（`owner` / `pet` / `vaccination` / `medical_record` / `treatment` / `examination` / `checkup` / `estimate` / `diagnosis` / `hospitalization` / `inventory` / `medicine` / `procedure` / `reservation` / `staff` / `vaccine` / `accounting` / `accounting_repository_unpaid` の各 `*_repository.go` — 着手時に `rg -n "Offset\(\(" internal/repository/` で全数再列挙すること）
- **問題**: 共有ヘルパが repository 層に存在せず、同一定型文が約18ファイルに手書き重複。
- **どう変える**: `base.go` に GORM Scope を新設し、全サイトを機械置換:

```go
// paginate は 1-origin の page と limit を Offset/Limit に変換する共有 Scope。
// page/limit の値正規化（下限・上限）は service 層 normalizePagination の責務であり、ここでは行わない。
func paginate(page, limit int) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Offset((page - 1) * limit).Limit(limit)
    }
}
// 使用側: .Scopes(clinicScope(clinicID), paginate(page, limit))
```

  計算式が `(page-1)*limit` と**厳密一致するサイトのみ**置換する。異なる計算（cursor 方式等）は触らない。
- **完了条件**: `rg -n "Offset\(\(page" internal/repository/` → 0件。`docker compose exec backend go test ./internal/repository/ -run 'TestOwnerRepository|TestReservationRepository|TestAccountingRepository' -count=1` 全PASS ＋ §1.2 の repository 系ゲート3本（Preload / 監査tx / dbOrTx）全PASS。
- **リスク / 戻し方**: 低（同一式への委譲）。`git revert`。
- **コミット**: `refactor(backend): repositoryページネーション定型文をpaginate Scopeへ集約`
- **依存**: BE7-0

#### BE7-7 [x] refactor: junction テーブル「検証→削除→一括insert」3実装の統合とバッチ挿入の統一

- **対象**: `internal/repository/reservation_staff_repository.go:211-264,293-336`（`UpdateExcludedReservationTypes` / `UpdateReservationCapabilities`）と `internal/repository/permission_group_repository.go:201-245`（`UpdateStaffGroups`）
- **問題**: 3関数が同型実装（コード自身が「対称」と相互参照コメント済み）なのに未抽出。さらに `UpdateStaffGroups` のみ `tx.CreateInBatches(rows, 100)` で他2つは無制限 `tx.Create(&items)`（バインド上限防御の非対称）。サブクエリ内の `.Where("clinic_id = ?")` 手書きが2箇所（`permission_group_repository.go:225` / `reservation_staff_repository.go:240`）で `clinicScope` をバイパス。
- **どう変える**: `base.go`（または `junction_helpers.go` 新設）に共通ヘルパを抽出し、3関数を委譲に書き換える。差分（対象モデル型・FK列名・検証クエリ）はコールバック/ジェネリクスで注入。insert は **3関数とも `CreateInBatches(rows, 100)` に統一**。`clinicScope` バイパス2箇所は `.Scopes(clinicScope(clinicID))` に置換。3関数の外部シグネチャ・エラー内容は不変。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run 'TestReservationStaffRepository|TestPermissionGroupRepository' -count=1` 全PASS ＋ repository 系ゲート3本 全PASS。**注意**: `reservation_staff_repository_test.go:166` の skip テスト（#236 台帳）には触れない。
- **リスク / 戻し方**: 中（tx 内ロジックの抽出。既存テストがゲート）。`git revert`。
- **コミット**: `refactor(backend): junction置換3実装を共通ヘルパへ統合しバッチ挿入を統一`
- **依存**: BE7-6（base.go への変更を直列化）

#### BE7-8 [x] refactor: シフト休憩置換の19行重複をヘルパ統合

- **対象**: `internal/repository/shift_entry_repository.go:111-130`（`ReplaceBreaks`）/ `internal/repository/shift_template_repository.go:92-111`（`UpdateBreaks`）
- **問題**: 型と FK 名のみ異なるほぼ完全一致の19行コピペ。
- **どう変える**: BE7-7 で作った共通ヘルパ（またはジェネリクス関数）で両者を委譲に書き換え。外部シグネチャ不変。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run 'TestShiftEntryRepository|TestShiftTemplateRepository' -count=1` 全PASS。
- **リスク / 戻し方**: 低。`git revert`。
- **コミット**: `refactor(backend): シフト休憩置換の重複実装を共通ヘルパへ統合`
- **依存**: BE7-7

#### BE7-9 [x] refactor: 日本語曜日配列の重複を timeutil へ集約

- **対象**: `internal/service/appointment_notification_service.go:346`（`weekdaysJP`）/ `internal/service/accounting_report_service.go:189`（`weekdayJP`、関数内ローカル）
- **問題**: `[...]string{"日","月","火","水","木","金","土"}` が2箇所で完全一致重複。
- **どう変える**: `internal/timeutil` に `WeekdayJP(t time.Time) string`（または公開配列 `WeekdaysJP`）を追加し、2箇所を置換。timeutil に単体テストを1本追加（`テスト名: TestWeekdayJP` — 日曜と土曜の境界）。
- **完了条件**: `docker compose exec backend go test ./internal/timeutil/ -run TestWeekdayJP -count=1` と `./internal/service/ -run 'TestAppointmentNotification|TestAccountingReport' -count=1` 全PASS。`rg -n '"日", "月"' internal/service/` → 0件。
- **リスク / 戻し方**: 低。`git revert`。
- **コミット**: `refactor(backend): 日本語曜日配列の重複をtimeutilへ集約`
- **依存**: BE7-0

#### BE7-10 [x] refactor: 割引率バリデーションのインライン再実装2箇所を共有ヘルパへ

- **対象**: `internal/service/treatment_service.go:199,363` — `internal/service/validators_owner.go:11-16` の `validateDiscountRate`（0-100 検証）と同一ルールをインライン再実装
- **問題**: 共有ヘルパが既にあるのに使っていない。ルール変更時にドリフトする。
- **どう変える**: 2箇所を `validateDiscountRate(...)` 呼び出しに置換。**エラーメッセージが変わる場合は既存テストの期待値を確認し、メッセージ差分があるなら実装側をヘルパへ寄せた上でテスト期待値を更新**（メッセージ統一が本項目の意図に含まれる）。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run TestTreatment -count=1` 全PASS。
- **リスク / 戻し方**: 低。`git revert`。
- **コミット**: `refactor(backend): 割引率バリデーションをvalidateDiscountRateへ統一`
- **依存**: BE7-0

#### BE7-11 [x] refactor: 犬/猫種別判定の3箇所コピペをヘルパ抽出

- **対象**: `internal/service/lstep_tag_sync_pet.go:33-36` / `lstep_tag_sync_pet_basic.go:103-105` / `lstep_health_tag_sync_prevention.go:46` — `strings.Contains(name, "犬"/"猫")` の部分一致判定を独立実装
- **問題**: 3箇所コピペ。なお `dose_calc.go:29-41` の `doseSpeciesAliases`（完全一致・投薬安全の fail-closed 設計）とは**契約が意図的に異なる**ため統合しない。
- **どう変える**: service パッケージ内に `isDogSpeciesName(name string) bool` / `isCatSpeciesName(name string) bool` を新設（実装は現行の `strings.Contains` のまま）し、3箇所を置換。ヘルパのコメントに「**部分一致（マーケティングタグ用途）。投薬計算の doseSpeciesAliases（完全一致・fail-closed）とは契約が異なり、統合してはならない**」と明記。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestLstepTagSync|TestLstepHealthTag' -count=1` 全PASS。`rg -n 'strings.Contains\(.*"犬"' internal/service/` のヒットがヘルパ定義1箇所のみ。
- **リスク / 戻し方**: 低（挙動同一）。`git revert`。
- **コミット**: `refactor(backend): 犬猫種別判定の3箇所コピペをisDog/isCatSpeciesNameへ抽出`
- **依存**: BE7-0

#### BE7-12 [x] refactor: LINE / Lステップ API URL の重複定義を一元化

- **対象**:
  - LINE push URL 完全重複: `internal/service/line_messaging_service.go:15`（`lineMessagingAPIURL`）と `internal/infra/line/client.go:16`（`pushEndpoint`）
  - LINE verify URL: `internal/middleware/liff_auth.go:131`（`var lineVerifyURL`、テスト差し替え可能設計）に対し `internal/service/line_link_service.go:290` が同一文字列を生リテラルで再ハードコード。`internal/service/lstep_settings_connection.go:47` にも `https://api.line.me` 素ホスト。
  - lstep base URL fallback 重複: `internal/service/lstep_settings_credentials.go:44` / `lstep_settings_connection.go:33`
- **問題**: 外部エンドポイントの二重管理。片方だけ変更されると静かに乖離する。テスト差し替え可能な変数があるのに再ハードコードした側は差し替え不能。
- **どう変える**: **依存方向が全層から安全な `internal/infra/line` / `internal/infra/lstep` を正とする**:
  1. `infra/line` に `const PushEndpoint` / `const VerifyEndpoint`（既存 `pushEndpoint` を公開化）を置き、`line_messaging_service.go` / `line_link_service.go` / `lstep_settings_connection.go` はそこから参照（service→infra import は既存方向で合法）。
  2. `middleware/liff_auth.go` の `var lineVerifyURL` は初期値を `line.VerifyEndpoint` にする（テスト差し替え可能性は維持）。
  3. `infra/lstep` に `const DefaultBaseURL` を置き、`lstep_settings_credentials.go` / `lstep_settings_connection.go` の fallback をそこへ委譲。
  値は全て現行文字列のまま（**挙動変更なし**）。
- **完了条件**: `rg -n '"https://api\.line\.me' internal/ --glob '!*_test.go'` のヒットが `infra/line` 内の const 定義のみ。`rg -n '"https://api\.lstep\.jp"' internal/ --glob '!*_test.go'` のヒットが `infra/lstep` 内のみ。`docker compose exec backend go test ./internal/service/ -run 'TestLineMessaging|TestLineLink|TestLstepSettings' -count=1` と `./internal/middleware/ -run TestLiffAuth -count=1` 全PASS。
- **リスク / 戻し方**: 低（値不変の参照統一）。`git revert`。
- **コミット**: `refactor(backend): LINE/LステップAPI URLの重複定義をinfra層constへ一元化`
- **依存**: BE7-0

#### BE7-13 [ ] refactor: バッチページサイズ定数500の重複統合

- **対象**: `internal/service/lstep_batch_dormant.go:13`（`dormantBatchPageSize = 500`）/ `internal/service/lstep_health_tag_sync_batch.go:12`（`healthPreventionBatchPageSize = 500`）
- **問題**: 同一値・同一目的（PERF-FOLLOWUP-02 カーソルページネーション）の定数が2つの名前で重複定義。
- **どう変える**: service パッケージ内で `lstepBatchPageSize = 500` に統合し（配置は `lstep_batch_service.go` 冒頭）、両ファイルの参照を置換。旧定数は削除。
- **完了条件**: `rg -n "BatchPageSize" internal/service/` のヒットが統合後の定義＋参照のみ。`docker compose exec backend go test ./internal/service/ -run 'TestLstepBatch|TestLstepHealthTag' -count=1` 全PASS。
- **リスク / 戻し方**: なし。`git revert`。
- **コミット**: `refactor(backend): Lステップバッチのページサイズ定数を統合`
- **依存**: BE7-0

### Phase 3 — refactor: デッドコード削除

#### BE7-14 [ ] refactor: 到達不能な orphan メソッド5件の削除

- **対象**（interface 宣言＋実装＋対応テストのセットで削除）:
  1. `internal/service/checkup_service.go:130` `CheckupService.GetByID` — handler に単票 GET ルートなし、呼び出しゼロ
  2. `internal/service/lstep_csv_import_service.go:202` `LstepCsvImportService.GetByID` — 同上
  3. `internal/repository/lab_import_repository.go:23,96` `FindByClinic` — service 呼び出しゼロ
  4. `internal/repository/lstep_delivery_trigger_log_repository.go:51,210` `FindAllByOwnerAndDateRange` — 同上
  5. `internal/repository/lstep_tag_code_mapping_repository.go:19,63` `Update` — 所有 service は削除→再作成方式のみ使用
- **問題**: いずれも production 参照ゼロ（テストのみ）の到達不能コード。git 履歴に残るため復元可能。
- **どう変える**: 削除前に**1件ずつ** `rg -n "<メソッド名>" internal cmd --glob '!*_test.go'` で production 参照0件を再確認（1件でもヒットしたら**残して報告**）。interface 宣言・実装・そのメソッド専用のテストケースを削除。**注意**: `AuditService.Log`（`audit_service.go:150`）は同ファイル内の `LogXxx` 群から内部利用されており**削除禁止**（§6 判断待ち）。
- **完了条件**: `docker compose exec backend go build ./...` 成功（ビルドのみフル可）。`docker compose exec backend go test ./internal/service/ -run 'TestCheckup|TestLstepCsvImport' -count=1` と `./internal/repository/ -run 'TestLabImport|TestLstepDeliveryTriggerLog|TestLstepTagCodeMapping' -count=1` 全PASS。
- **リスク / 戻し方**: 低（未到達コードの削除。coverage は分母ごと減るため ratchet に影響しない）。`git revert`。
- **コミット**: `refactor(backend): 到達不能なorphanメソッド5件を削除`
- **依存**: BE7-0

### Phase 4 — test: テスト層の重複整理

#### BE7-15 [ ] test: 重複モック5クラスタ（15型）の集約

- **棚卸し（2026-07-14）**: **部分対応** — handler の `mockAuditService` は `internal/handler/mocks_audit_test.go` に統合済み（F-4）。service 層の AuditService / PermissionGroupRepository / OccupationRepository / ReservationTypeOccupationRepository クラスタと `mocks_shared_test.go` 集約は未達。本文は残す。
- **対象**: service/handler テストに散在する同名同役割モック — `AuditService`×5 / `AuditRepository`×2 / `PermissionGroupRepository`×4 / `OccupationRepository`×2 / `ReservationTypeOccupationRepository`×2（着手時に `rg -n "type mock" internal/service internal/handler --glob '*_test.go'` で全数再列挙して突合）
- **問題**: 直近10日で積み上がった生きた重複。能力差もある（例: `mockAuditRepository` は last-value capture、`mockTreatmentAuditRepository` は append-list — multi-call 検証が片方では不可能）。
- **どう変える**: パッケージごとに共有モックファイル（例: `internal/service/mocks_shared_test.go`）へ**最も高機能な実装**（append-list 記録＋メソッド別 error 注入）を1つずつ集約し、各テストは embed ＋必要時 override で利用。**テストの assert 自体は変更しない**（モック差し替えのみ）。1クラスタずつ置換し、都度対象テストを回す。
- **完了条件**: 対象クラスタの重複定義が各1つに減っていること（rg 再列挙で確認）。`docker compose exec backend go test ./internal/service/ -run 'TestAudit|TestTreatment|TestPermissionGroup|TestOccupation' -count=1` 全PASS。
- **リスク / 戻し方**: 中（テスト大量修正。落ちたら該当クラスタのみ revert し残りは維持してよい — その場合は分割コミットに変更し報告）。
- **コミット**: `test(backend): 重複モック5クラスタを共有モックへ集約`
- **依存**: BE7-0

#### BE7-16 [ ] test: repository テストの makeOwner 系ヘルパ3重複を統合

- **対象**: `internal/repository/vital_repository_test.go:39`（`makeVitalOwner`）/ `accounting_repository_unpaid_test.go:36`（`makeOwner`）/ `lstep_tag_cache_repository_test.go:41`（`makeTagCacheOwner`）
- **問題**: 完全一致の owner 生成ヘルパが3ファイルに重複。
- **どう変える**: 既存のテスト共有ファイル（`db_setup_test.go` など package 共通の *_test.go）に `makeTestOwner` として1本化し、3箇所を置換・旧ヘルパ削除。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run 'TestVital|TestAccountingRepositoryUnpaid|TestLstepTagCache' -count=1` 全PASS。
- **リスク / 戻し方**: 低。`git revert`。
- **コミット**: `test(backend): repositoryテストのowner生成ヘルパを統合`
- **依存**: BE7-0

### Phase 5 — docs: API 契約の乖離是正

#### BE7-17 [ ] docs: api.yaml のフィールドレベル乖離4クラスタを実装に同期

- **対象**: `backend/docs/api.yaml`（フィールドレベルは apicontract ゲートの対象外＝手動同期が必要な領域。**実装が正、spec を直す**）:
  1. **ShiftEntry**（api.yaml:3258-3298 / 実装 `internal/handler/shift_response.go:18-32`）: `id`/`clinic_id`/`staff_id` を `type: integer, format: int64` → **`type: string`** に修正（実装は `strconv.FormatUint` で string 直列化）。`staff_name`（string）を追加。spec のみに存在する幽霊 `note`（単数）を削除（実装は `notes` のみ）。
  2. **LineCustomer**（api.yaml:5898-5924 / 実装 `internal/handler/line_customer_response.go:18`）: `owner_name`（string）を追加。
  3. **ReservationAppointment**（api.yaml:1199-1290 / 実装 `internal/handler/reservation_response.go`）: 未文書化5フィールド `owner` / `line_customer_id` / `reservation_route` / `actual_reservation_at` / `created_by_staff` を実装の json 型に合わせて追加。spec のみの幽霊 `deleted_at`（1262-1266行）を削除。
  4. 各修正時、実装側 struct の json タグと1フィールドずつ突合してから書く（推測で型を書かない）。
- **完了条件**: `docker compose exec backend go test ./internal/apicontract/ -count=1` 全PASS（date-format / route drift ゲートを壊していないこと）。コード変更なし。
- **リスク / 戻し方**: なし（ドキュメントのみ。FE の型は tygo が Go から直接生成するため api.yaml 変更の実行時影響はない）。`git revert`。
- **コミット**: `docs(backend): api.yamlのShiftEntry/LineCustomer/ReservationAppointmentフィールド乖離を実装に同期`
- **依存**: BE7-0

### Phase 6 — refactor: 構造（M〜L規模、最後に実施）

#### BE7-18 [ ] refactor: billing_item CreateItem の3ブロック抽出

- **対象**: `internal/service/billing_item_service.go:167`（`CreateItem`、114行）
- **問題**: (a) L168-179 入力値バリデーション、(b) L182-200 クロステナント所有権検証、(c) L203-229 category/taxType/taxRate/source のデフォルト解決 — 相互依存のない3ブロックが1関数に同居。
- **どう変える**: 同ファイル内に `validateCreateBillingItemInput` / `validateBillingItemOwnership` / `resolveBillingItemDefaults` の3つの非公開関数として**そのまま持ち出す**（ロジック不変・tx 外の純粋処理のみ）。P13 のファイル内定義順序（const→buildFunc→…→methods）を守る。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run TestBillingItem -count=1` 全PASS（既存テスト無変更）＋ master-FK write inventory ゲート PASS（所有権検証コードを移すため、allowlist の更新が必要ならその差分も同一コミットに含め報告）。
- **リスク / 戻し方**: 低〜中。`git revert`。
- **コミット**: `refactor(backend): billing_item CreateItemの検証・所有権・デフォルト解決を関数抽出`
- **依存**: BE7-1（同ファイル近傍を触るため直列化）

#### BE7-19 [ ] refactor: trimming の appointment 差分構築を buildXxxUpdateFields パターンへ

- **対象**: `internal/service/trimming_service.go:290`（`createDetailForExistingAppointment`、108行）の L336-361 — apptFields 差分構築が tx クロージャ内にインライン
- **問題**: 他サービスで確立済みの `buildXxxUpdateFields` パターン（P13）を使わず tx 内にベタ書き。
- **どう変える**: `buildTrimmingAppointmentUpdateFields(...) map[string]any` としてトップレベル非公開関数に抽出（ロジック不変）。tx クロージャは呼び出しに置換。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run TestTrimming -count=1` 全PASS（既存テスト無変更）。
- **リスク / 戻し方**: 低。`git revert`。
- **コミット**: `refactor(backend): trimmingのappointment差分構築をbuildUpdateFieldsパターンへ抽出`
- **依存**: BE7-0

#### BE7-20 [ ] refactor: TokenService の抽出（Auth 構造是正・第1段）

- **対象**: `internal/handler/auth_session.go:138`（`issueAuthCookies` 内の JWT 生成・署名）/ `internal/handler/auth_handler.go:181`（`RefreshToken` 内の JWT パース・署名検証・blacklist 照合）
- **問題**: Auth ドメインには service 層が存在せず（`internal/service/` に `validators_auth.go` のみ）、JWT 発行/検証というセキュリティクリティカルなロジックが handler 層に直書きされている。service 層の機械強制規約（P8/P11 等）が適用されない構造的迂回。
- **どう変える**（第1段はトークンのみ。**挙動・Cookie 属性・トークン形式は1ビットも変えない**）:
  1. `internal/service/token_service.go` を新設し、`TokenService` interface（`IssueAccessToken` / `IssueRefreshToken` / `VerifyRefreshToken`（blacklist 照合込み）等 — メソッド分割は現 handler コードの自然な切れ目に従う）と実装を定義。JWT 生成・パース・blacklist 照合コードを**そのまま移す**。
  2. `service.NewServices` に組み込み、handler は `h.svc.Token` を呼ぶだけにする。Cookie の Set 自体（http 層の関心）は handler に残す。
  3. 既存の handler auth 系テスト（`internal/handler/` の auth 系 *_test.go）は**無変更で全PASS**が絶対条件（挙動固定ゲート）。service 側に token の単体テストを新規追加（`テスト名: TestTokenService_IssueAndVerify` — 発行→検証往復、改竄検知、blacklist 拒否）。
- **完了条件**: `docker compose exec backend go test ./internal/handler/ -run TestAuth -count=1`（無変更で）全PASS ＋ `./internal/service/ -run TestTokenService -count=1` 全PASS ＋ ルート snapshot ゲート PASS。
- **リスク / 戻し方**: 中〜高（認証コア。既存テスト無変更 PASS が満たせない場合は**即 revert して中断報告** — テスト書き換えで通すことを禁止）。
- **コミット**: `refactor(backend): JWT発行・検証をTokenServiceへ抽出しhandler層から分離`
- **依存**: BE7-0〜BE7-17 完了後に着手（認証を触る前に他項目で green 実績を積む）

#### BE7-21 [ ] refactor: AuthService の抽出（Auth 構造是正・第2段）

- **対象**: `internal/handler/auth_session.go:79`（`authenticateUser` — bcrypt 照合）/ `:109`（`resolveClinicInfo`）/ `:126`（`resolveSystemAdminMainClinicID`）/ `internal/handler/auth_me_response.go:125`（`calculateEffectivePermissions`）
- **問題**: BE7-20 と同根。ログイン認証・クリニック解決・実効権限計算のビジネスロジックが handler 層に居る。
- **どう変える**: `internal/service/auth_service.go` を新設し、上記4関数のロジックを `AuthService`（`Authenticate` / `ResolveClinicInfo` / `EffectivePermissions` 等）へ**そのまま移す**（bcrypt 比較・フォールバック分岐・権限計算の1行も変えない）。handler は bind → `h.svc.Auth.*` → response 変換のみに縮小。監査ログ呼び出しの位置（現状 handler／middleware にあるもの）は**動かさない**（audit tx ゲートに触れないため）。既存 handler テスト無変更 PASS が絶対条件。service 側に `TestAuthService_Authenticate`（正しいパスワード/誤り/無効スタッフ）を新規追加。
- **完了条件**: `docker compose exec backend go test ./internal/handler/ -run 'TestAuth|TestMe' -count=1`（無変更で）全PASS ＋ `./internal/service/ -run TestAuthService -count=1` 全PASS ＋ ルート snapshot ゲート PASS。
- **リスク / 戻し方**: 高（ログイン全経路）。満たせなければ**即 revert・中断報告**。
- **コミット**: `refactor(backend): 認証・権限計算ロジックをAuthServiceへ抽出`
- **依存**: BE7-20

---

## 5. やらないこと（実行者はこれらに手を出さない）

1. **#236 の既知バグ修正**（クロステナント Staff 削除 / ClinicSettings 列名不一致 / IsNotFound 恒偽）— Issue 側で管理。skip テストの解除も禁止。
2. **cron 未配線バッチ3本**（`RunLTVTopPercentSyncAllClinics` / `RunVisitDormantSyncAllClinics` / `RunHealthPreventionTagSyncAllClinics`、FEAT-377/379）の**配線も削除も禁止**。Lステップ Write API 一時停止の経緯があり PO 判断待ち（§6）。
3. アップロード許可拡張子の乖離（`model/shared_file.go` は GIF なし vs `medical_record_image_request.go` は GIF あり）の同期 — 業務意図の可能性、PO 確認待ち（§6）。
4. accounting service の repository 集計型パススルーの DTO 化 — 変換層追加の設計判断が必要（§6）。
5. duration リテラル（`24*time.Hour` 等）の一括定数化・`INTERVAL '365 days'` SQL 3箇所の共通化 — 「同一値」でも意味の同一性が保証できず、機械統合は誤りを生む。
6. `internal/middleware/response.go` の `respondError` と handler `RespondError` の統一（X-17 で意図的見送り済み）。
7. `AuditService.Log` の interface 露出整理・`middleware/liff_auth.go` の repository 直接呼び出しの service 経由化 — 要精査（§6）。
8. `middleware/auth.go` `checkStaffActive` の fail-open 変更 — コメントで意図的と文書化済み（§6 で再検討提起のみ）。
9. apicontract ゲートのフィールドレベル突合への拡張 — 大工事、次期（§6）。
10. 機能追加・仕様変更・API パス変更・migration / seed / DB スキーマ変更・依存ライブラリ変更。
11. `api.yaml` の変更は BE7-17 に明記した範囲のみ。ルート追加・削除はしない。
12. テストの削除・skip 化（BE7-14 の orphan 対応テスト削除のみ例外）。
13. push / PR / フル `go test ./...` / フル lint / `gofmt -w` の自動実行。
14. `FE-refactor.md`・frontend/ 配下に触らない。

---

## 6. 次期監査への引き継ぎ・PO 判断待ち（実行者は無視してよい）

- **[PO判断] cron 未配線3本（FEAT-377/379）**: 実装済み・テスト済みだが `cmd/api/main.go` に未登録＝本番未稼働。Lステップ Write API 一時停止（2026-05）との整合を確認の上、「配線して稼働」か「削除」かを決める。
- **[PO確認] 画像アップロード拡張子の乖離**: 共有ファイル（GIF 不可）vs カルテ画像（GIF 可）。意図的なら両所にコメント1行で文書化。
- **[設計判断] accounting service の repository 型パススルー**（`GetMonthlyUnpaidCarryover` 等4メソッド）: service DTO を挟むか現状維持か。
- **[要精査] `middleware/liff_auth.go:56,116`**: `LineCustomerRepository.FindOrCreateByLineUserID` を middleware から直接呼んでいる（層違反の疑い。構造監査では lookup インターフェース経由とも見える — 両者の実態を精査）。
- **[再検討] `checkStaffActive` の fail-open**（DB 一時障害時に認証を通す）: 文書化済みの意図的挙動だが、過去の CRITICAL fail-open 修正クラスと同型。権限系レビュー時に再検討。
- **[次期] apicontract のフィールドレベルゲート**: 今回の BE7-17 クラスの乖離を機械検出する拡張。ShiftEntry の型不一致クラスの再発防止として価値が高い。
- **[次期] `examination_service.go:337` ReplaceItems（101行）/ medicine・reservation・cash_register の80行台関数**: 今回未精査。次期の god-function 走査対象。
- **[次期] `lstep_csv_import_service.go` が自前 `s.db`（gorm 直 import）を持つ妥当性**。
- **[次期] `reservation_type_handler.go:142-153` の weekly/specific 相互依存バリデーションの service 移動**（LOW）。
- **[記録] 旧監査数値の訂正**: 「死にモック30」は現存せず（全数走査で未使用0）。「モック6重複」の実態は5クラスタ15型（BE7-15 が正）。
- **[記録] `AuthService`/`TokenService` 抽出（BE7-20/21）完了後**、`validators_auth.go` との統合と P8/P11 準拠の付与を次期で検討（今回は「移すだけ」に限定している）。

---

## 7. 実行者への指示文（このままコピペして渡す）

```
あなたは AnimalEkarte のバックエンド実行者です。BE-refactor.md（第7期）を実行してください。
（棚卸し 2026-07-14: 対応済 9。残存 BE7-10〜BE7-21。部分対応 BE7-15 は完了条件を満たしてから [x]。）

1. backend/CLAUDE.md、必要に応じて各層 CLAUDE.md（handler/service/repository）、および BE-refactor.md 全文を読む。本書とコード以外の文脈は存在しない前提で作業する。
2. BE7-0（安全網）から着手し、以後 BE7-1 → BE7-21 を番号順に、1項目ずつ実施する。
3. 各項目は「変更 → 完了条件のコマンドを実行 → 全PASS確認 → gofmt -l 無出力確認 → 指定メッセージでコミット（git add はファイル指定のみ）」の順。完了条件を満たせない場合は、その項目を revert し、中断して状況を報告する。勝手な代替実装をしない。
4. 1項目=1コミット。複数項目をまとめない。本書の進捗欄 [x] 更新は同一コミットに含めてよい。
5. repository 層に触れた項目では §1.2 の機械ゲート（Preload / 監査tx / dbOrTx）を、handler のルート近傍に触れた項目ではルート snapshot を、必ずローカル実行する。
6. §5「やらないこと」を厳守。特に #236 のバグ修正・skip 解除・cron 未配線3本には触れない。
7. BE7-20/21（Auth 構造是正）は既存 handler テストを 1 文字も変えずに全PASS させること。満たせなければ即 revert・中断報告。テストの書き換えで通すことを禁止する。
8. push・PR・フル go test ./...・フル lint は実行しない。全項目完了後、以下をユーザーに提示して手動実行を依頼する:
   $ docker compose exec backend go test ./...
   $ docker compose exec backend golangci-lint run ./...
9. 最後に、各項目のコミットハッシュ一覧と、中断・スキップした項目があればその理由を報告する。
```
