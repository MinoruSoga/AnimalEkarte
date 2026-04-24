# TASK-235: closing_settings_handler.go — 4エンドポイントが model/DTO を直接返却（response変換なし）

## 優先度
Medium

## 対象ファイル
- `backend/internal/handler/closing_settings_handler.go`

## 問題概要
以下の4エンドポイントが `model.*` または service の内部 DTO を `c.JSON` に直接渡しており、
専用の `to{Entity}Response()` 変換関数を経由していない。

全ハンドラ規約: **model は絶対に handler から直接返却しない。必ず response 変換関数を通す。**

## 現状コード

```go
// 行51 — UpdateClosingSettings: *model.ClinicSettings を直接返却
result, err := h.svc.ClosingSettings.UpdateStandard(...)
c.JSON(http.StatusOK, result)               // ❌

// 行66 — ListSpecialPeriods: []model.ClosingSpecialPeriod を直接返却
periods, err := h.svc.ClosingSettings.ListSpecialPeriods(...)
c.JSON(http.StatusOK, periods)              // ❌

// 行93 — CreateSpecialPeriod: *model.ClosingSpecialPeriod を直接返却
period, err := h.svc.ClosingSettings.CreateSpecialPeriod(...)
c.JSON(http.StatusCreated, period)          // ❌

// 行126 — UpdateSpecialPeriod: *model.ClosingSpecialPeriod を直接返却
period, err := h.svc.ClosingSettings.UpdateSpecialPeriod(...)
c.JSON(http.StatusOK, period)               // ❌
```

正しい参照実装（同ファイル 行26付近 GetClosingSettings）:
```go
resp, err := h.svc.ClosingSettings.Get(...)
c.JSON(http.StatusOK, resp)  // ✅ service DTO を返す（model ではない）
```

## あるべき姿

`closing_settings_response.go` を新規作成し、以下を実装する:

```go
type closingSpecialPeriodResponse struct {
    ID          uint64    `json:"id"`
    ClinicID    uint64    `json:"clinic_id"`
    Name        string    `json:"name"`
    StartDate   string    `json:"start_date"`
    EndDate     string    `json:"end_date"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func toClosingSpecialPeriodResponse(p *model.ClosingSpecialPeriod) closingSpecialPeriodResponse { ... }
```

handler 側:
```go
c.JSON(http.StatusCreated, toClosingSpecialPeriodResponse(period))
```

## 完了条件
- [ ] `closing_settings_response.go` を新規作成
- [ ] 4エンドポイントすべてで `to{Entity}Response()` 経由に変更
- [ ] `go test ./backend/internal/...` がパス
