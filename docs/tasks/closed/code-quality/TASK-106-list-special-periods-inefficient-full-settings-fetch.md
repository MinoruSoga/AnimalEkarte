# TASK-106: `ListSpecialPeriods` ハンドラで不要な設定データを取得している

## 優先度

**Medium** — 不要な DB クエリが発生しており、責務分離も崩れている。

---

## 概要

`closing_settings_handler.go` の `ListSpecialPeriods` ハンドラが
`ClosingSettings.Get(ctx, clinicID)` を呼び出しているが、このメソッドは
`clinic_settings` テーブルと `closing_special_periods` テーブルの両方をクエリして
`ClosingSettingsResponse`（settings + special_periods）を返す。

`ListSpecialPeriods` は special_periods のみを必要としているため、
settings の DB クエリが毎回無駄に実行されている。

---

## 問題箇所

### `handler/closing_settings_handler.go:87-93`

```go
// GET /v1/closing-settings/special-periods
func (h *Handler) ListSpecialPeriods(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    // ❌ Get() は settings + special_periods の両方を取得するが、
    // special_periods しか使わない
    resp, err := h.svc.ClosingSettings.Get(c.Request.Context(), clinicID)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, resp.SpecialPeriods)  // settings は捨てる
}
```

---

## 修正方針

`ClosingSettingsService` インターフェースに `ListSpecialPeriods` メソッドを追加し、
`periodRepo.FindAll()` のみを呼ぶ専用の実装を提供する。

### 1. `service/closing_settings_service.go` — インターフェース拡張

```go
// ✅ インターフェースに追加
type ClosingSettingsService interface {
    Get(ctx context.Context, clinicID uint64) (*ClosingSettingsResponse, error)
    ListSpecialPeriods(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error)  // 追加
    UpdateStandard(ctx context.Context, clinicID uint64, input UpdateClinicSettingsInput) (*model.ClinicSettings, error)
    CreateSpecialPeriod(ctx context.Context, clinicID uint64, input CreateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
    UpdateSpecialPeriod(ctx context.Context, clinicID, id uint64, input UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error)
    DeleteSpecialPeriod(ctx context.Context, clinicID, id uint64) error
    ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*DaySchedule, error)
}

// ✅ 実装追加
func (s *closingSettingsService) ListSpecialPeriods(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
    periods, err := s.periodRepo.FindAll(ctx, clinicID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to list special periods")
    }
    return periods, nil
}
```

### 2. `handler/closing_settings_handler.go:83-94`

```go
// ✅ 修正後
func (h *Handler) ListSpecialPeriods(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    periods, err := h.svc.ClosingSettings.ListSpecialPeriods(c.Request.Context(), clinicID)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, periods)
}
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `handler/closing_settings_handler.go:87-93` | ListSpecialPeriods | ❌ 不要な settings 取得 |
| `service/closing_settings_service.go` | ClosingSettingsService インターフェース | 追加必要 |
| `service/closing_settings_service.go` | closingSettingsService 実装 | 追加必要 |

---

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — バックエンド・アーキテクチャ規約

> handler → service → repository の軽量レイヤードを徹底

handler は必要なデータだけを取得するサービスメソッドを呼ぶべきであり、
不要なデータを取得して破棄するのは責務分離違反。

### `.claude/rules/performance-rules.md` — データベース最適化

> 必要カラムのみ取得

settings テーブルへの無駄なクエリを排除することで API のレスポンスタイムを改善できる。
