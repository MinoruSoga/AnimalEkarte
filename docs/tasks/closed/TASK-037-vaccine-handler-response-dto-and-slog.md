# TASK-037: vaccine ドメイン — Response DTO 未使用 & slog 不備 3件

## 優先度

HIGH（Response DTO）/ MEDIUM（slog）

---

## 問題 1: vaccine_handler の全レスポンスが生モデルを返している（Response DTO なし）

### ファイル
`backend/internal/handler/vaccine_handler.go:31, 49, 84, 123`

### 問題
GetVaccine / ListVaccines / CreateVaccine / UpdateVaccine の全エンドポイントで生モデルを `c.JSON` に渡している。

```go
// L31 GetVaccine
c.JSON(http.StatusOK, vaccine)                     // *model.Vaccine を直接返す

// L49 ListVaccines
c.JSON(http.StatusOK, vaccines)                    // []model.Vaccine を直接返す

// L84 CreateVaccine
c.JSON(http.StatusCreated, vaccine)                // *model.Vaccine を直接返す

// L123 UpdateVaccine
c.JSON(http.StatusOK, vaccine)                     // *model.Vaccine を直接返す
```

medicine / insurance 等は `toXxxResponse()` 経由でレスポンスを返しており、DB カラム追加時の API コントラクト意図しない変更リスクがある。

### 修正案
`handler/vaccine_response.go` に `toVaccineResponse` を追加し、全レスポンスで使用する。

```go
type vaccineResponse struct {
    ID          uint64  `json:"id"`
    Name        string  `json:"name"`
    Price       *int64  `json:"price"`
    IsActive    bool    `json:"is_active"`
    Description string  `json:"description"`
    Species     *string `json:"species"`
    Interval    string  `json:"interval"`
    ParentID    *uint64 `json:"parent_id"`
    SortOrder   int     `json:"sort_order"`
    CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
}

func toVaccineResponse(v *model.Vaccine) vaccineResponse { ... }

// handler 側
c.JSON(http.StatusOK, toVaccineResponse(vaccine))
c.JSON(http.StatusOK, mapSlice(vaccines, toVaccineResponse))
```

---

## 問題 2: vaccine_service の Create/Update slog に clinic_id が欠落

### ファイル
`backend/internal/service/vaccine_service.go:71, 83`

### 問題
```go
// L71 Create
slog.InfoContext(ctx, "vaccine created", slog.Uint64("vaccine_id", vaccine.ID))
// → clinic_id なし

// L83 Update
slog.InfoContext(ctx, "vaccine updated", slog.Uint64("vaccine_id", id))
// → clinic_id なし

// L141 Delete（こちらは正しい）
slog.InfoContext(ctx, "vaccine deleted", slog.Uint64("vaccine_id", id), slog.Uint64("clinic_id", clinicID))
```

マルチテナント環境でのデバッグ・監査が困難になる。Delete には clinic_id があるが Create/Update にない非対称。

### 修正案
```go
slog.InfoContext(ctx, "vaccine created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("vaccine_id", vaccine.ID))

slog.InfoContext(ctx, "vaccine updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("vaccine_id", id))
```

---

## 問題 3: vaccine_service の Reorder に slog.InfoContext なし

### ファイル
`backend/internal/service/vaccine_service.go:145-153`

### 問題
```go
func (s *vaccineService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if len(ids) == 0 {
        return apperrors.WrapInvalidInput("ids must not be empty")
    }
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return apperrors.Wrap(err, "failed to reorder vaccines")
    }
    return nil  // slog なし
}
```

Create / Update / Delete には slog があるが Reorder のみ欠落。TASK-027/029 で他ドメインの同パターンを修正済みだが vaccine は未対応。

### 修正案
```go
slog.InfoContext(ctx, "vaccines reordered",
    slog.Uint64("clinic_id", clinicID),
    slog.Int("count", len(ids)))
return nil
```
