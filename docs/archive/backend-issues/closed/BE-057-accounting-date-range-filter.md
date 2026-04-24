# BE-057: 会計管理一覧API に日付範囲フィルタ追加

**Status**: Open
**Priority**: Medium
**Affects**: 会計管理 (`/v1/accountings`)
**Date Created**: 2026-03-25
**Related**: TASK-028, FE-120

## Summary

会計管理一覧API(`GET /v1/accountings`)に`start_date`/`end_date`クエリパラメータを追加し、会計日（`scheduled_date`カラム）での絞り込みを可能にする。

handler → service → repository の3層を修正する。参照実装は`vaccination`の同パターン。

## 現状のコード

```go
// backend/internal/handler/accounting_handler.go:14-54
func (h *Handler) ListAccountings(c *gin.Context) {
    // ... petID, ownerID, status の抽出
    accountings, total, err := h.svc.Accounting.List(c.Request.Context(), clinicID, petID, ownerID, status, page, limit)
    // ← start_date / end_date の処理なし
}
```

```go
// backend/internal/service/accounting_service.go:26
func (s *accountingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, page, limit int) ([]model.Billing, int64, error) {
    return s.repo.FindAll(ctx, clinicID, petID, ownerID, status, page, limit)
    // ← startDate, endDate 引数なし
}
```

```go
// backend/internal/repository/accounting_repository.go:30
func (r *accountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, page, limit int) ([]model.Billing, int64, error) {
    // ← scheduled_date フィルタのwhere句なし
}
```

## 必要な変更

### 1. Handler 変更

```go
// backend/internal/handler/accounting_handler.go
// line 50（status ブロック終了直後、List呼び出しの前）に追加:

startDate, err := parseDateQuery(c, "start_date")
if err != nil {
    RespondError(c, err)
    return
}
endDate, err := parseDateQuery(c, "end_date")
if err != nil {
    RespondError(c, err)
    return
}

accountings, total, err := h.svc.Accounting.List(c.Request.Context(), clinicID, petID, ownerID, status, startDate, endDate, page, limit)
```

### 2. Service Interface + 実装変更

```go
// backend/internal/service/accounting_service.go

// Interface:
type AccountingService interface {
    List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
    // ... 他のメソッドは変更なし
}

// 実装:
func (s *accountingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
    return s.repo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}
```

### 3. Repository Interface + 実装変更

```go
// backend/internal/repository/accounting_repository.go

// Interface:
type AccountingRepository interface {
    FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
    // ... 他のメソッドは変更なし
}

// 実装:
func (r *accountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
    // 既存のクエリ構築に以下を追加:
    if startDate != nil {
        q = q.Where("scheduled_date >= ?", *startDate)
    }
    if endDate != nil {
        q = q.Where("scheduled_date <= ?", *endDate)
    }
    // ...
}
```

**注意**: フィルタ対象カラムは`scheduled_date`（`model.Billing`の`ScheduledDate`フィールド）。`date`ではない。

## APIレスポンス形式（変更なし）

```
GET /v1/accountings?start_date=2026-01-01&end_date=2026-01-31
→ 2026年1月1日〜1月31日のscheduled_dateを持つ会計のみ返す

GET /v1/accountings?status=completed&start_date=2026-01-01
→ ステータスと日付の複合フィルタも動作する
```

## テストファイル更新（必須）

インターフェースのシグネチャが変わるため、モック実装をすべて更新する必要がある。

### `backend/internal/service/accounting_service_test.go`

```go
// line 17: findAllFn 型定義に startDate, endDate を追加
findAllFn func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)

// line 24-25: FindAll メソッドシグネチャを更新
func (m *mockAccountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
    return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

// line 167: テストテーブル内 findAllFn ラムダに _, _ *string を追加
findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _ *string, _, _ *string, _, _ int) ([]model.Billing, int64, error) {

// line 173: svc.List() 呼び出しに nil, nil を追加
billings, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, tt.status, nil, nil, tt.page, tt.limit)
```

※ accounting_handler_test.go は存在しないため、ハンドラ側のモック更新は不要。

## フロントエンド影響

- FE-120 で `getAccountings` に日付パラメータを追加する
- `make codegen` は不要（モデル変更なし）

## 参照実装

vaccination_repository.go の date フィルタパターン（BE-056と同様）。
フィールド名が`date`ではなく`scheduled_date`である点のみ異なる。

## 完了条件

- [ ] `GET /v1/accountings?start_date=2026-01-01&end_date=2026-01-31` が指定期間の会計のみ返す
- [ ] `status`との複合フィルタが正しく動作する
- [ ] パラメータなしの場合は全件返す（既存動作を破壊しない）
- [ ] `accounting_service_test.go` のモックシグネチャ更新済み
- [ ] `go test ./... -v` が通る
- [ ] `golangci-lint run ./...` が通る
