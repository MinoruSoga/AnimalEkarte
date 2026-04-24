# TASK-046: trimming_master_handler — model.TargetSize 変換がハンドラ層で実行

## 優先度

MEDIUM

---

## 概要

TASK-035 で `shift_handler.go` の `model.ShiftType(req.ShiftType)` ENUM 変換がハンドラ層に漏れている問題を指摘した。同パターンが `trimming_master_handler.go` の `UpdateTrimmingCourse` にも存在する。

ハンドラ層はリクエスト解析と HTTP レスポンスのみを担う。`model` パッケージへの型変換はサービス層の責務である。

---

## 問題: UpdateTrimmingCourse で model.TargetSize 変換をハンドラが実施

### ファイル
`backend/internal/handler/trimming_master_handler.go:95-106`

```go
svcInput := service.UpdateTrimmingCourseInput{
    Name:        req.Name,
    Price:       req.Price,
    IsActive:    req.IsActive,
    Description: req.Description,
    Duration:    req.Duration,
    SortOrder:   req.SortOrder,
}
if req.TargetSize != nil {
    ts := model.TargetSize(*req.TargetSize)  // ← ハンドラが model 型に変換
    svcInput.TargetSize = &ts
}
```

`UpdateTrimmingCourseInput.TargetSize` が `*model.TargetSize` であるため、ハンドラが `model` パッケージに依存せざるを得ない構造になっている。

---

## 修正方針

### 1. service DTO の TargetSize フィールドを `*string` に変更

```go
// service/trimming_master_service.go
type UpdateTrimmingCourseInput struct {
    Name        *string
    Price       *int64
    IsActive    *bool
    Description *string
    TargetSize  *string  // ← model.TargetSize から string に変更
    Duration    *int
    SortOrder   *int
}
```

### 2. service 層で変換

```go
func buildTrimmingCourseUpdateFields(input UpdateTrimmingCourseInput) map[string]any {
    fields := make(map[string]any)
    // ...
    if input.TargetSize != nil {
        fields["target_size"] = model.TargetSize(*input.TargetSize)  // ← service 層で変換
    }
    // ...
}
```

### 3. ハンドラを単純化

```go
// handler/trimming_master_handler.go
svcInput := service.UpdateTrimmingCourseInput{
    Name:        req.Name,
    Price:       req.Price,
    IsActive:    req.IsActive,
    Description: req.Description,
    TargetSize:  req.TargetSize,  // ← *string のまま渡す
    Duration:    req.Duration,
    SortOrder:   req.SortOrder,
}
// if ブロックが不要になる
```

---

## 備考

`CreateTrimmingCourseInput.TargetSize` は `string` 型であり、service 内で `model.TargetSize` に変換している（`trimming_master_service.go:70-73`）。
Update の DTO だけが `*model.TargetSize` を使っており、Create と Update で不統一になっている点も修正対象。
