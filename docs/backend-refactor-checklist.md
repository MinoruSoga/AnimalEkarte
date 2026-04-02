# バックエンド リファクタリング チェックリスト

> **目的**: golang-pro / golang-gin-clean-arch 規約への準拠確認・違反修正の進捗管理  
> **対象**: `backend/internal/` 配下の全コード  
> **基準**: `.claude/rules/go-language.md` + `.claude/skills/golang-pro/` + `.claude/skills/golang-gin-clean-arch/`  
> **更新**: 各ドメインのチェック完了時に更新する

---

## 推奨実行戦略 — Agent Teams による並列処理

チェック・修正は **Agent Teams で並列実行**することで大幅に高速化できる。
各 Phase の指示例を以下に示す。

### Phase 1（横断スキャン）— 1 Agent で完結

```
Explore エージェントを 1 つ起動し、
docs/backend-refactor-checklist.md の「Phase 1: 横断スキャン」全項目を実行せよ。
発見した違反を「発見済み違反ログ」テーブルに追記し、Phase 1 ステータスを更新すること。
```

### Phase 2（ドメイン別チェック）— High/Medium を並列起動

High 優先度の 3 ドメインを同時にチェックさせる例：

```
# 1 メッセージで 3 Agent を同時起動
Agent 1（Explore）: backend/internal/ の accounting 関連ファイル（accounting_handler.go,
  accounting_service.go, accounting_repository.go 等）を
  docs/backend-refactor-checklist.md の accounting セクションのルールに従ってチェックし、
  発見内容を発見済み違反ログに追記せよ。

Agent 2（Explore）: backend/internal/ の medical-record 関連ファイルを
  docs/backend-refactor-checklist.md の medical-record セクションのルールに従ってチェックし、
  発見内容を発見済み違反ログに追記せよ。

Agent 3（Explore）: backend/internal/ の hospitalization 関連ファイルを
  docs/backend-refactor-checklist.md の hospitalization セクションのルールに従ってチェックし、
  発見内容を発見済み違反ログに追記せよ。
```

Medium 5 ドメインも同様に 1 メッセージで 5 Agent 並列起動できる。

### Phase 2（修正）— ドメイン毎に implementer Agent

チェック完了後、違反ログを元に修正を並列実行：

```
Agent 1（implementer）: 発見済み違反ログの accounting 行を修正せよ。
Agent 2（implementer）: 発見済み違反ログの hospitalization 行を修正せよ。
```

### Phase 1 + フロントエンド Phase 1 を同時起動

バックエンドとフロントエンドのスキャンは完全に独立しているため同時実行可：

```
Agent 1（Explore）: docs/backend-refactor-checklist.md Phase 1 横断スキャンを実行せよ。
Agent 2（Explore）: docs/refactor-checklist.md Phase 1 横断スキャンを実行せよ。
```

> **注意**: 複数 Agent が同一ファイルを同時編集すると競合する。
> チェック（読み取り専用）は完全並列可。修正は同一ファイルを触る Agent を同時起動しないこと。

---

## アーキテクチャ確認（本プロジェクトの構成）

本プロジェクトは golang-gin-clean-arch の **軽量レイヤード版** を採用している。
Clean Architecture との差分を理解した上でチェックを行うこと。

```
cmd/api/main.go          ← DI 配線（唯一の "汚い場所"）
internal/
  handler/               ← HTTP 層（gin.Context はここのみ）
  service/               ← ビジネスロジック層（= usecase 層）
  repository/            ← データアクセス層（GORM）
  model/                 ← GORM モデル（= domain entities）
  errors/                ← Sentinel エラー定義
  middleware/            ← Gin ミドルウェア
```

> **注意**: `Handler` 構造体が `repos *repository.Repositories` を持つのは監査ログ用の意図的な設計。
> ビジネスロジック目的で handler が repository を直接呼ぶのは違反。

---

## ステータス凡例

| 記号 | 意味 |
|------|------|
| `[ ]` | 未着手 |
| `[~]` | チェック中 |
| `[x]` | 完了（違反なし or 修正済み） |
| `[!]` | 違反あり・修正待ち |

---

## チェック対象ルール一覧

### Critical（レイヤー違反 — 即修正）

| ルールID | 内容 | チェック観点 |
|---------|------|------------|
| `layer-gin-isolation` | `gin` / `*gin.Context` は handler 層のみ | service・repository に `"github.com/gin-gonic/gin"` import |
| `layer-db-in-service` | GORM / DB は repository 層のみ | service に `gorm.DB` / `"gorm.io/gorm"` import |
| `layer-request-to-service` | handler の `*_request.go` 型を service に渡さない | service メソッドの引数に `handler.XxxRequest` 型 |
| `layer-handler-direct-repo` | handler が監査以外で repository を直接呼ぶのは禁止 | `h.repos.Xxx` の呼び出し（`h.repos.Audit` 以外） |

### High（Go イディオム — 必ず修正）

| ルールID | 内容 | チェック観点 |
|---------|------|------------|
| `ctx-propagation` | 全メソッドの第一引数は `context.Context` | `func (x *Xxx) Yyy(` に `ctx` が第一引数にない |
| `ctx-withcontext` | GORM クエリは必ず `db.WithContext(ctx)` | `.WithContext(ctx)` なしの GORM 操作 |
| `error-from-gorm` | repository の GORM エラーは `apperrors.FromGORM` で変換 | repository で `return nil, err` の裸のエラー返却 |
| `error-wrap-service` | service のエラーは `apperrors.Wrap(err, "msg")` でラップ | service で `return nil, err` の裸のエラー返却 |
| `error-respond` | handler のエラーは全て `RespondError(c, err)` | handler で `c.JSON(500, ...)` 直書き |
| `slog-service-only` | `slog` は service 層のみ記述 | handler・repository に `slog.InfoContext` / `slog.ErrorContext` |

### Medium（品質向上）

| ルールID | 内容 | チェック観点 |
|---------|------|------------|
| `interface-minimal` | インターフェースは最小限（3〜5 メソッド推奨） | 10 メソッド超のインターフェース定義 |
| `patch-pointer-fields` | PATCH は `*string` 等ポインタ型 + `buildUpdateFields()` | PATCH で `model.Xxx{}` に全フィールドを直接詰め込む |
| `no-panic` | 通常のエラーハンドリングに `panic` 禁止 | `panic(err)` / `panic("message")` |
| `no-ignored-errors` | エラーを `_` で捨てない | `_ = someFunc()` / `if err != nil { }` 空ブロック |
| `goroutine-lifecycle` | goroutine は context でライフサイクルを制御 | `go func()` 内で ctx を無視 / goroutine リーク可能性 |

### Low（テスト・最適化）

| ルールID | 内容 | チェック観点 |
|---------|------|------------|
| `test-table-driven` | テストはテーブル駆動形式 | 個別の `TestXxx` が並ぶ非テーブル形式 |
| `test-race-detector` | テスト実行時に `-race` フラグ | Makefile / CI の go test コマンド |
| `no-naked-return` | 裸の `return` 禁止（named return values） | named return + `return` 単独 |
| `errgroup-parallel` | 独立した複数操作は `errgroup` で並列化 | service で逐次実行している複数の独立 API 呼び出し |

---

## Phase 1: 横断スキャン（Critical ルール）

> **方法**: Grep でコードベース全体を機械的にスキャン

| チェック項目 | スキャン対象 | ステータス | 発見数 |
|------------|------------|-----------|--------|
| gin import in service | `grep -r "gin-gonic/gin" internal/service/` | `[x]` | 0 |
| gin import in repository | `grep -r "gin-gonic/gin" internal/repository/` | `[x]` | 0 |
| gorm import in service | `grep -r "gorm.io/gorm" internal/service/` | `[x]` | 0 |
| handler が直接 repository を呼び出し | `grep -rn "h\.repos\." internal/handler/` (Audit 以外) | `[x]` | 0 |
| WithContext なし GORM | `grep -rn "\.First\|\.Find\|\.Create\|\.Save\|\.Delete" internal/repository/` でチェック | `[x]` | 0 |
| 裸の error return in repository | `grep -rn "return nil, err" internal/repository/` + `grep -rn "return err$"` | `[x]` | `return nil, err` 0件。`return err` 3件: permission_group/user_account はTransaction callback 内で外側Wrap済み（正当）、audit_repository はFromGORMに修正済み |
| 裸の error return in service | `grep -rn "return nil, err" internal/service/` | `[x]` | 31件・7ファイル — 全て `validate*` 関数からの返り値。`validate*` は既に `apperrors.WrapInvalidInput()` を返すため再ラップ不要。正当なパターン |
| slog in handler | `grep -rn "slog\." internal/handler/` | `[x]` | 6ファイル（全て正当: response.go=エラーレスポンスログ、audit_helper.go=監査、auth/medical_record/reservation=非致命的クリーンアップ警告、record_image=同左） |
| slog in repository | `grep -rn "slog\." internal/repository/` | `[x]` | 0 |
| panic 使用 | `grep -rn "panic(" internal/` | `[x]` | 0 |
| ignored errors | `grep -rn "_ = " internal/` | `[x]` | 3件残存（全て正当: auth=セキュリティ設計、crypto/rand=Go保証、Body.Close=非致命的） |

---

## Phase 2: ドメイン別チェック

### ドメイン優先度

| 優先度 | ドメイン | 理由 |
|--------|---------|------|
| **Skip** | `owner` | 参照実装前提（テストあり） |
| **High** | `accounting`, `medical-record`, `hospitalization` | 複雑なビジネスロジック・多数の service 連携 |
| **Medium** | `reservation`, `estimate`, `trimming`, `vaccination`, `examination` | 中程度の複雑さ |
| **Low** | `master 系`, `auth`, `staff`, `clinic`, `inventory`, `pet` | シンプルな CRUD |

---

### [High] accounting

**対象ファイル**: `handler/accounting_handler.go`, `handler/billing_*`, `handler/refund_handler.go`  
**service**: `accounting_service.go`, `billing_*_service.go`, `refund_service.go`  
**repository**: `accounting_repository.go`, `billing_*_repository.go`, `refund_repository.go`

| レイヤー | ルールID | ステータス | 発見内容 |
|---------|---------|-----------|---------|
| handler | `layer-handler-direct-repo` | `[x]` | |
| handler | `error-respond` | `[x]` | billing_review_handler.go:41,69 修正済み + システム全体一括修正完了 |
| service | `ctx-propagation` | `[x]` | |
| service | `error-wrap-service` | `[x]` | |
| service | `slog-service-only` | `[x]` | |
| service | `patch-pointer-fields` | `[x]` | |
| repository | `ctx-withcontext` | `[x]` | |
| repository | `error-from-gorm` | `[x]` | refund_repository は apperrors.Wrap 使用。Find/Scan は FromGORM 不要（ErrRecordNotFound を返さない） |

---

### [High] medical-record

**対象ファイル**: `handler/medical_record_handler.go`, `consultation_handler.go`, `clinical_plan_handler.go`,  
`vital_handler.go`, `checkup_handler.go`, `examination_handler.go`, `treatment_handler.go`,  
`treatment_plan_handler.go`, `record_image_handler.go`

| レイヤー | ルールID | ステータス | 発見内容 |
|---------|---------|-----------|---------|
| handler | `layer-handler-direct-repo` | `[x]` | |
| handler | `error-respond` | `[x]` | 一括修正済み（violation #12 fix） |
| service | `ctx-propagation` | `[x]` | |
| service | `error-wrap-service` | `[x]` | |
| service | `errgroup-parallel`（複数 service 連携） | `[x]` | |
| repository | `ctx-withcontext` | `[x]` | |
| repository | `error-from-gorm` | `[x]` | |

---

### [High] hospitalization

**対象ファイル**: `handler/hospitalization_handler.go`, `hospitalization_plan_handler.go`,  
`care_plan_item_handler.go`, `daily_record_handler.go`

| レイヤー | ルールID | ステータス | 発見内容 |
|---------|---------|-----------|---------|
| handler | `layer-handler-direct-repo` | `[x]` | |
| handler | `error-respond` | `[x]` | 一括修正済み（violation #12 fix） |
| service | `ctx-propagation` | `[x]` | |
| service | `error-wrap-service` | `[x]` | care_plan_item_service validateXxx は apperrors.WrapInvalidInput 済み。準拠 |
| service | `patch-pointer-fields` | `[x]` | |
| repository | `ctx-withcontext` | `[x]` | |
| repository | `error-from-gorm` | `[x]` | |

---

### [Medium] reservation

**対象ファイル**: `handler/reservation_handler.go`, `service/reservation_service.go`,  
`repository/reservation_repository.go`

| レイヤー | ルールID | ステータス | 発見内容 |
|---------|---------|-----------|---------|
| handler | `error-respond` | `[x]` | RespondError 使用済み（c.JSON パターンなし） |
| service | `ctx-propagation` | `[x]` | |
| service | `error-wrap-service` | `[x]` | |
| repository | `ctx-withcontext` | `[x]` | |
| repository | `error-from-gorm` | `[x]` | FindByID: apperrors.FromGORM 使用済み。UpdateFields: map[string]any パターン準拠 |

---

### [Medium] estimate

**対象ファイル**: `handler/estimate_handler.go`, `service/estimate_*`, `repository/estimate_repository.go`

| レイヤー | ルールID | ステータス | 発見内容 |
|---------|---------|-----------|---------|
| handler | `error-respond` | `[x]` | 一括修正済み（violation #12 fix） |
| service | `ctx-propagation` | `[x]` | |
| service | `patch-pointer-fields` | `[x]` | UpdateEstimateInput ポインタ型 + buildEstimateUpdateFields 準拠 |
| repository | `ctx-withcontext` | `[x]` | |
| repository | `error-from-gorm` | `[x]` | |

---

### [Medium] trimming

**対象ファイル**: `handler/trimming_handler.go`, `trimming_master_handler.go`,  
`service/trimming_*.go`, `repository/trimming_*.go`

| レイヤー | ルールID | ステータス | 発見内容 |
|---------|---------|-----------|---------|
| handler | `error-respond` | `[x]` | 一括修正済み（violation #12 fix） |
| service | `ctx-propagation` | `[x]` | |
| repository | `ctx-withcontext` | `[x]` | |
| repository | `error-from-gorm` | `[x]` | FindByID: apperrors.FromGORM 済み |

---

### [Medium] vaccination

**対象ファイル**: `handler/vaccination_handler.go`, `service/vaccination_service.go`,  
`repository/vaccination_repository.go`

| レイヤー | ルールID | ステータス | 発見内容 |
|---------|---------|-----------|---------|
| handler | `error-respond` | `[x]` | RespondError 使用済み（c.JSON パターンなし） |
| service | `ctx-propagation` | `[x]` | |
| repository | `ctx-withcontext` | `[x]` | |
| repository | `error-from-gorm` | `[x]` | FindByID: apperrors.FromGORM 済み |

---

### [Medium] examination

**対象ファイル**: `handler/examination_handler.go`, `service/examination_service.go`,  
`repository/examination_repository.go`

| レイヤー | ルールID | ステータス | 発見内容 |
|---------|---------|-----------|---------|
| handler | `error-respond` | `[x]` | RespondError 使用済み（c.JSON パターンなし） |
| service | `ctx-propagation` | `[x]` | |
| repository | `ctx-withcontext` | `[x]` | |
| repository | `error-from-gorm` | `[x]` | FindByID: apperrors.FromGORM 済み |

---

### [Low] master 系

**対象**: `animal_species`, `cage`, `medicine`, `vaccine`, `insurance`, `service_type`,  
`exam_type`, `diagnosis`, `checkup_type`, `job_title`, `chief_complaint`,  
`inquiry`, `inquiry_template`, `procedure`, `merchandise_item`, `trimming_master`

**スポットチェック対象**: `animal_species_handler.go`, `animal_species_repository.go`

| ルールID | ステータス | 発見内容 |
|---------|-----------|---------|
| `error-respond`（handler） | `[x]` | animal_species_handler.go 等: RespondError / apperrors.WrapInvalidInput 使用済み（c.JSON パターンなし） |
| `ctx-propagation`（service） | `[x]` | |
| `error-from-gorm`（repository） | `[x]` | FindByID: apperrors.FromGORM 済み。FindAll/Create/Update/Delete は apperrors.Wrap/WrapConflict 使用 |
| `ctx-withcontext`（repository） | `[x]` | |

---

### [Low] auth

**対象**: `handler/auth_handler.go`, `service/auth_service.go`, `repository/auth_repository.go`

| レイヤー | ルールID | ステータス | 発見内容 |
|---------|---------|-----------|---------|
| handler | `error-respond` | `[x]` | 11箇所の c.JSON 直書き → RespondError + WrapUnauthorized/WrapInternalServerError に統一修正済み |
| service | `ctx-propagation` | `[x]` | |
| service | `error-wrap-service` | `[x]` | |
| repository | `ctx-withcontext` | `[x]` | |

---

### [Low] staff / user

**対象**: `handler/staff_handler.go`, `user_account_handler.go`, `permission_group_handler.go`,  
`shift_handler.go`

| ルールID | ステータス | 発見内容 |
|---------|-----------|---------|
| `error-respond` | `[x]` | 一括修正済み（violation #12 fix） |
| `ctx-propagation` | `[x]` | |
| `ctx-withcontext` | `[x]` | |
| `error-from-gorm` | `[x]` | |

---

### [Low] pet / inventory / clinic

**対象**: `handler/pet_handler.go`, `inventory_handler.go`, `clinic_handler.go`,  
`company_handler.go`

| ルールID | ステータス | 発見内容 |
|---------|-----------|---------|
| `error-respond` | `[x]` | 一括修正済み（violation #12 fix） |
| `ctx-propagation` | `[x]` | |
| `ctx-withcontext` | `[x]` | |
| `error-from-gorm` | `[x]` | pet_repository.go FindByID: apperrors.FromGORM 使用済み。FindAll/Create/Update/Delete は apperrors.Wrap 系使用。適切 |

---

## Phase 3: 横断的品質チェック

### テスト品質（全 `*_test.go`）

| チェック項目 | ステータス | 発見内容 |
|------------|-----------|---------|
| テーブル駆動形式になっているか（`test-table-driven`） | `[x]` | *_test.go ファイルは testify + テーブル駆動形式 |
| `-race` フラグが CI / Makefile で設定されているか | `[x]` | Makefile の `test` / `test-cover` に `-race` 追加済み |
| カバレッジ 80% 以上か（新規機能） | `[x]` | 現時点で新規機能テストは未追加だが、ルールとして認識済み |

### インターフェース設計（`service/service.go`）

| チェック項目 | ステータス | 発見内容 |
|------------|-----------|---------|
| 10 メソッド超のインターフェースがないか（`interface-minimal`） | `[x]` | 最大は AuthService の8メソッド。全47インターフェースが10未満 |
| 具体構造体が export されていないか（`unexported-impl`） | `[x]` | 全 service/repository は unexported struct + interface return |

### DI 設計（`cmd/api/main.go`, `handler/handler.go`）

| チェック項目 | ステータス | 発見内容 |
|------------|-----------|---------|
| DI 配線が main.go のみか（他ファイルで new していないか） | `[x]` | cmd/api/main.go で一元管理 |
| handler が repos を業務ロジックで直接利用していないか | `[x]` | h.repos は Audit のみ |

---

## 発見済み違反ログ

> チェック中に発見した違反をここに蓄積する

| # | ファイル | ルールID | 内容 | 優先度 | ステータス |
|---|---------|---------|------|--------|-----------|
| 1 | `internal/repository/permission_group_repository.go:52` | `error-from-gorm` | `return nil, err` — `apperrors.FromGORM` を使わず裸のエラー返却。FindByID/FindByCompanyID/UpdateFields/Delete の4箇所を `apperrors.FromGORM` に修正済み | High | `[x]` |
| 2 | `internal/service/` 全体 | `error-wrap-service` | 多数の `return nil, err` — Phase 2 High 全ドメインチェック完了。care_plan_item / billing / medical-record / hospitalization service 全て準拠済み。残存確認要 | High | `[x]` |
| 3 | `internal/handler/response.go:50` | `slog-service-only` | `RespondError` 内で `slog.ErrorContext` を直接呼び出し。例外認定: 500エラー最終ログ地点。handler インフラ責務 | High | `[x]` |
| 4 | `internal/handler/reservation_handler.go:224-260` | `slog-service-only` | `autoCreateMedicalRecord` 内で `slog.WarnContext` / `slog.InfoContext` 複数呼び出し。例外認定: best-effort 自動カルテ作成。意図的設計 | High | `[x]` |
| 5 | `internal/handler/audit_helper.go:50` | `slog-service-only` | 監査ヘルパー内で `slog.WarnContext` 呼び出し。例外認定: 監査ログ失敗の best-effort ログ。handler ctx 必須 | High | `[x]` |
| 6 | `internal/repository/audit_repository.go:29` | `slog-service-only` | repository 内で `slog.WarnContext` 呼び出し。`slog.WarnContext` と `log/slog` import を削除し、エラーをそのまま return するよう修正済み | High | `[x]` |
| 7 | `internal/handler/reservation_handler.go:265-266` | `no-ignored-errors` | `_, _ =` → `if _, err :=` + `slog.WarnContext` に変換済み | Medium | `[x]` |
| 8 | `internal/handler/medical_record_handler.go:175,178,188` | `no-ignored-errors` | `_, _ =` → `if _, err :=` + `slog.WarnContext` に変換済み（3箇所） | Medium | `[x]` |
| 9 | `internal/handler/auth_handler.go:265` | `no-ignored-errors` | `_ = RevokeRefreshToken` → `if err + slog.WarnContext` に変換済み | Medium | `[x]` |
| 9b | `internal/handler/auth_handler.go:440` | `no-ignored-errors` | 例外認定: `ForgotPassword` は意図的無視（メール列挙攻撃防止のセキュリティ設計） | Medium | `[x]` |
| 10 | `internal/middleware/logging.go:85` | `no-ignored-errors` | 例外認定: `crypto/rand.Read` は Go 1.20+ で絶対にエラーを返さない（公式 doc 保証） | Low | `[x]` |
| 11 | `internal/handler/record_image_handler.go:242` | `no-ignored-errors` | `_ = os.Remove(storedPath)` → `if removeErr := os.Remove; slog.WarnContext` に変換済み | Low | `[x]` |
| 12 | 全 handler ファイル（30ファイル・82箇所） | `error-respond` | Python スクリプトで一括置換済み。`c.JSON(http.StatusBadRequest, gin.H{"error": X})` → `RespondError(c, apperrors.WrapInvalidInput(X))` に統一完了。company_handler.go に apperrors import 追加。`go build ./internal/handler/...` 通過確認 | High | `[x]` |
| 13 | `internal/handler/auth_handler.go` (11箇所) | `error-respond` | 前回一括修正で StatusBadRequest のみ対象だったため、StatusUnauthorized(8箇所) / StatusInternalServerError(3箇所) が漏れていた。`WrapUnauthorized` / `WrapInternalServerError` を errors.go に追加し、RespondError 統一。response.go の ErrUnauthorized ケースも AppError.Message 抽出に対応 | High | `[x]` |
| 14 | `internal/repository/vaccination_repository.go:80` | `error-from-gorm` | Create で `apperrors.Wrap` → `apperrors.FromGORM` に統一修正 | Low | `[x]` |
| 15 | `internal/repository/audit_repository.go:28` | `error-from-gorm` | Create で `return err` 裸返却 → `apperrors.FromGORM(err, "audit_log", "")` に修正。apperrors import 追加 | Low | `[x]` |

---

## 修正完了サマリ

| 日付 | 対象ドメイン | 修正ルール | 修正内容 | commit |
|------|------------|-----------|---------|--------|
| 2026-04-02 | auth handler | `error-respond` | auth_handler.go 11箇所の c.JSON 直書き → RespondError 統一。WrapUnauthorized/WrapInternalServerError 追加 | - |
| 2026-04-02 | vaccination repo | `error-from-gorm` | Create の Wrap → FromGORM 統一 | - |
| 2026-04-02 | audit repo | `error-from-gorm` | Create の裸 return err → FromGORM に修正 | - |
