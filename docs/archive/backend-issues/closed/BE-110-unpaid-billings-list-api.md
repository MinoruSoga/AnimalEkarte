# BE-110: 未納 billings 一覧 API（飼主単位集約 / 会計単位）

**Status**: Closed (2026-04-14, commit 8fcd1382)
**Priority**: High
**Affects**: `accounting_repository.go`, `accounting_service.go`, `accounting_handler.go`
**Date Created**: 2026-04-14
**Related**: BUG-370, FE-248

## Summary

基準日時点で未納（`status=waiting` かつ `scheduled_date < 基準日`）の billings を、飼主単位集約 / 会計単位の 2 形式で返す API を新設する。

## 現状のコード

`backend/internal/repository/accounting_repository.go:31-87` の `FindAll` は status / 日付フィルタを持つが、「未納に特化した飼主単位集約」のメソッドは存在しない。

```go
// backend/internal/model/accounting.go:17-24
type BillingStatus string
const (
    BillingStatusWaiting   BillingStatus = "waiting"
    BillingStatusCompleted BillingStatus = "completed"
    BillingStatusCancelled BillingStatus = "cancelled"
    BillingStatusPending   BillingStatus = "pending"
)
```

## 必要な変更

### 1. DB マイグレーション
**なし**（既存スキーマで対応）

### 2. Model 変更
**なし**

### 3. Repository 変更

```go
// backend/internal/repository/accounting_repository.go

// 会計単位の未納一覧
type UnpaidBillingFilter struct {
    AsOf  time.Time // 基準日（この日より前の scheduled_date が対象）
    Page  int
    Limit int
}

func (r *accountingRepository) FindUnpaidByBilling(
    ctx context.Context, clinicID uint64, filter UnpaidBillingFilter,
) ([]model.Billing, int64, error) {
    var billings []model.Billing
    var total int64

    q := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Where("status = ?", model.BillingStatusWaiting).
        Where("scheduled_date < ?", filter.AsOf).
        Where("deleted_at IS NULL")

    if err := q.Model(&model.Billing{}).Count(&total).Error; err != nil {
        return nil, 0, apperrors.FromGORM(err, "billings", "")
    }

    err := q.Preload("Owner").Preload("Pet").
        Order("scheduled_date ASC"). // 経過日数の降順 = 古い順
        Limit(filter.Limit).Offset((filter.Page - 1) * filter.Limit).
        Find(&billings).Error
    if err != nil {
        return nil, 0, apperrors.FromGORM(err, "billings", "")
    }
    return billings, total, nil
}

// 飼主単位の集約結果型
type UnpaidByOwnerRow struct {
    OwnerID         uint64    `gorm:"column:owner_id"`
    OwnerName       string    `gorm:"column:owner_name"`
    BillingCount    int64     `gorm:"column:billing_count"`
    TotalAmount     int64     `gorm:"column:total_amount"`
    OldestDate      time.Time `gorm:"column:oldest_date"`
    LatestDate      time.Time `gorm:"column:latest_date"`
}

func (r *accountingRepository) FindUnpaidByOwner(
    ctx context.Context, clinicID uint64, filter UnpaidBillingFilter,
) ([]UnpaidByOwnerRow, int64, error) {
    var rows []UnpaidByOwnerRow
    var total int64

    base := r.db.WithContext(ctx).
        Table("billings").
        Joins("INNER JOIN owners ON owners.id = billings.owner_id AND owners.deleted_at IS NULL").
        Where("billings.clinic_id = ?", clinicID).
        Where("billings.status = ?", model.BillingStatusWaiting).
        Where("billings.scheduled_date < ?", filter.AsOf).
        Where("billings.deleted_at IS NULL").
        Where("billings.owner_id IS NOT NULL").
        Group("billings.owner_id, owners.name")

    // 件数（owner 数）
    if err := r.db.WithContext(ctx).
        Raw("SELECT COUNT(*) FROM (?) sub", base.Select("billings.owner_id")).
        Scan(&total).Error; err != nil {
        return nil, 0, apperrors.FromGORM(err, "billings", "")
    }

    err := base.
        Select(`
            billings.owner_id AS owner_id,
            owners.name AS owner_name,
            COUNT(billings.id) AS billing_count,
            SUM(billings.total_amount) AS total_amount,
            MIN(billings.scheduled_date) AS oldest_date,
            MAX(billings.scheduled_date) AS latest_date
        `).
        Order("oldest_date ASC"). // 最古日付の昇順 = 経過日数の降順
        Limit(filter.Limit).Offset((filter.Page - 1) * filter.Limit).
        Scan(&rows).Error
    if err != nil {
        return nil, 0, apperrors.FromGORM(err, "billings", "")
    }
    return rows, total, nil
}

// 売掛金総額（フィルタ対象全件のサマリー）
type UnpaidSummary struct {
    TotalAmount  int64
    BillingCount int64
    OwnerCount   int64
}

func (r *accountingRepository) UnpaidSummary(
    ctx context.Context, clinicID uint64, asOf time.Time,
) (*UnpaidSummary, error) {
    var s UnpaidSummary
    err := r.db.WithContext(ctx).
        Table("billings").
        Where("clinic_id = ?", clinicID).
        Where("status = ?", model.BillingStatusWaiting).
        Where("scheduled_date < ?", asOf).
        Where("deleted_at IS NULL").
        Select(`
            COALESCE(SUM(total_amount), 0) AS total_amount,
            COUNT(*) AS billing_count,
            COUNT(DISTINCT owner_id) FILTER (WHERE owner_id IS NOT NULL) AS owner_count
        `).
        Scan(&s).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "billings", "")
    }
    return &s, nil
}
```

### 4. Service 変更

```go
// backend/internal/service/accounting_service.go

type ListUnpaidInput struct {
    AsOf    time.Time
    GroupBy string // "owner" or "billing"
    Page    int
    Limit   int
}

type ListUnpaidByOwnerOutput struct {
    Rows    []repository.UnpaidByOwnerRow
    Total   int64
    Summary *repository.UnpaidSummary
}

type ListUnpaidByBillingOutput struct {
    Billings []model.Billing
    Total    int64
    Summary  *repository.UnpaidSummary
}

func (s *accountingService) ListUnpaidByBilling(
    ctx context.Context, clinicID uint64, input ListUnpaidInput,
) (*ListUnpaidByBillingOutput, error) {
    summary, err := s.repo.UnpaidSummary(ctx, clinicID, input.AsOf)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get unpaid summary")
    }
    billings, total, err := s.repo.FindUnpaidByBilling(ctx, clinicID, repository.UnpaidBillingFilter{
        AsOf: input.AsOf, Page: input.Page, Limit: input.Limit,
    })
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to list unpaid billings")
    }
    return &ListUnpaidByBillingOutput{Billings: billings, Total: total, Summary: summary}, nil
}

func (s *accountingService) ListUnpaidByOwner(
    ctx context.Context, clinicID uint64, input ListUnpaidInput,
) (*ListUnpaidByOwnerOutput, error) {
    summary, err := s.repo.UnpaidSummary(ctx, clinicID, input.AsOf)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get unpaid summary")
    }
    rows, total, err := s.repo.FindUnpaidByOwner(ctx, clinicID, repository.UnpaidBillingFilter{
        AsOf: input.AsOf, Page: input.Page, Limit: input.Limit,
    })
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to list unpaid by owner")
    }
    return &ListUnpaidByOwnerOutput{Rows: rows, Total: total, Summary: summary}, nil
}
```

### 5. Handler 変更

```go
// backend/internal/handler/accounting_handler.go

// ListUnpaidBillings godoc
// GET /api/clinics/:id/unpaid-billings?as_of=YYYY-MM-DD&group_by=owner|billing&page=1&limit=20
func (h *Handler) ListUnpaidBillings(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }

    asOf, err := parseDateQuery(c, "as_of")
    if err != nil {
        RespondError(c, err)
        return
    }
    if asOf == nil {
        today := time.Now().In(time.Local).Truncate(24 * time.Hour)
        asOf = &today
    }

    groupBy := c.DefaultQuery("group_by", "billing")
    if groupBy != "owner" && groupBy != "billing" {
        RespondError(c, apperrors.WrapInvalidInput("group_by must be 'owner' or 'billing'"))
        return
    }

    page, limit, err := parsePagination(c)
    if err != nil {
        RespondError(c, err)
        return
    }

    input := service.ListUnpaidInput{AsOf: *asOf, GroupBy: groupBy, Page: page, Limit: limit}

    if groupBy == "owner" {
        out, err := h.svc.Accounting.ListUnpaidByOwner(c.Request.Context(), clinicID, input)
        if err != nil {
            RespondError(c, err)
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "data":    toUnpaidByOwnerResponse(out.Rows),
            "summary": toUnpaidSummaryResponse(out.Summary),
            "total":   out.Total,
            "page":    page,
            "limit":   limit,
        })
        return
    }

    out, err := h.svc.Accounting.ListUnpaidByBilling(c.Request.Context(), clinicID, input)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "data":    toAccountingResponseList(out.Billings),
        "summary": toUnpaidSummaryResponse(out.Summary),
        "total":   out.Total,
        "page":    page,
        "limit":   limit,
    })
}
```

### 6. Response 変更

```go
// backend/internal/handler/accounting_response.go に追加

type UnpaidByOwnerResponse struct {
    OwnerID      uint64    `json:"owner_id"`
    OwnerName    string    `json:"owner_name"`
    BillingCount int64     `json:"billing_count"`
    TotalAmount  int64     `json:"total_amount"`
    OldestDate   time.Time `json:"oldest_date"`
    LatestDate   time.Time `json:"latest_date"`
    ElapsedDays  int       `json:"elapsed_days"` // 今日 - oldest_date
}

type UnpaidSummaryResponse struct {
    TotalAmount  int64 `json:"total_amount"`
    BillingCount int64 `json:"billing_count"`
    OwnerCount   int64 `json:"owner_count"`
}

func toUnpaidByOwnerResponse(rows []repository.UnpaidByOwnerRow) []UnpaidByOwnerResponse {
    today := time.Now().In(time.Local).Truncate(24 * time.Hour)
    out := make([]UnpaidByOwnerResponse, 0, len(rows))
    for _, r := range rows {
        out = append(out, UnpaidByOwnerResponse{
            OwnerID:      r.OwnerID,
            OwnerName:    r.OwnerName,
            BillingCount: r.BillingCount,
            TotalAmount:  r.TotalAmount,
            OldestDate:   r.OldestDate,
            LatestDate:   r.LatestDate,
            ElapsedDays:  int(today.Sub(r.OldestDate).Hours() / 24),
        })
    }
    return out
}

func toUnpaidSummaryResponse(s *repository.UnpaidSummary) UnpaidSummaryResponse {
    if s == nil {
        return UnpaidSummaryResponse{}
    }
    return UnpaidSummaryResponse{
        TotalAmount:  s.TotalAmount,
        BillingCount: s.BillingCount,
        OwnerCount:   s.OwnerCount,
    }
}
```

### 7. ルート追加

```go
// backend/cmd/api/main.go の accounting ルートグループに追加
clinics.GET("/unpaid-billings", h.ListUnpaidBillings)
```

## API レスポンス形式

```json
// GET /api/clinics/1/unpaid-billings?as_of=2026-04-30&group_by=owner&page=1&limit=20
{
  "data": [
    {
      "owner_id": 123,
      "owner_name": "山田太郎",
      "billing_count": 3,
      "total_amount": 25000,
      "oldest_date": "2026-02-10T00:00:00+09:00",
      "latest_date": "2026-04-05T00:00:00+09:00",
      "elapsed_days": 64
    }
  ],
  "summary": {
    "total_amount": 1234567,
    "billing_count": 45,
    "owner_count": 23
  },
  "total": 23,
  "page": 1,
  "limit": 20
}
```

## フロントエンド影響

- `make codegen` で `models.ts` に新規型は追加されない（response 型は handler 側のみ）
- FE-248 で対応が必要

## 完了条件

- [ ] DB マイグレーションなし
- [ ] 3層実装（repository → service → handler）完了
- [ ] ルート登録完了
- [ ] テストケース追加（テーブル駆動: `group_by=owner` / `group_by=billing` / `as_of` 未指定 / 不正値）
- [ ] `clinic_id` マルチテナント絞り込みが正しく機能
- [ ] `deleted_at IS NULL` が両クエリで適用されている
- [ ] `cancelled` / `completed` の billing が混入しないこと
- [ ] `go test ./... -race` パス
- [ ] `golangci-lint run` パス
