# TASK-074: exam_type / checkup_type レスポンスフィールド欠落

## 優先度

HIGH

---

## 概要

`exam_type_response.go` と `checkup_type_response.go` のレスポンス構造体に
モデルが持つ重要フィールドが欠落している。
クライアントが親子関係や固有フィールドを参照できない。

---

## exam_type_response.go — ParentID 欠落

### 問題

```go
// ❌ 現状: ParentID が response struct に存在しない
type examTypeResponse struct {
    ID          uint64                 `json:"id"`
    ClinicID    uint64                 `json:"clinic_id"`
    Name        string                 `json:"name"`
    Price       *int64                 `json:"price,omitempty"`
    IsActive    bool                   `json:"is_active"`
    Description string                 `json:"description"`
    // ParentID がない
    SortOrder   int                    `json:"sort_order"`
    Items       []examTypeItemResponse `json:"items,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

### 修正後

```go
// ✅ ParentID 追加
type examTypeResponse struct {
    ID          uint64                 `json:"id"`
    ClinicID    uint64                 `json:"clinic_id"`
    Name        string                 `json:"name"`
    Price       *int64                 `json:"price,omitempty"`
    IsActive    bool                   `json:"is_active"`
    Description string                 `json:"description"`
    ParentID    *uint64                `json:"parent_id,omitempty"`  // 追加
    SortOrder   int                    `json:"sort_order"`
    Items       []examTypeItemResponse `json:"items,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

func toExamTypeResponse(et *model.ExaminationType) examTypeResponse {
    return examTypeResponse{
        // ...
        ParentID: et.ParentID,  // 追加
        // ...
    }
}
```

---

## checkup_type_response.go — ParentID / Interval / TargetAge 欠落

### 問題

```go
// ❌ 現状: ParentID, Interval, TargetAge が欠落
type checkupTypeResponse struct {
    ID          uint64    `json:"id"`
    ClinicID    uint64    `json:"clinic_id"`
    Name        string    `json:"name"`
    Price       *int64    `json:"price,omitempty"`
    IsActive    bool      `json:"is_active"`
    Description string    `json:"description"`
    SortOrder   int       `json:"sort_order"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    // ParentID, Interval, TargetAge がない
}
```

### 修正後

```go
// ✅ 3フィールド追加
type checkupTypeResponse struct {
    ID          uint64    `json:"id"`
    ClinicID    uint64    `json:"clinic_id"`
    Name        string    `json:"name"`
    Price       *int64    `json:"price,omitempty"`
    IsActive    bool      `json:"is_active"`
    Description string    `json:"description"`
    ParentID    *uint64   `json:"parent_id,omitempty"`   // 追加
    SortOrder   int       `json:"sort_order"`
    Interval    string    `json:"interval"`               // 追加
    TargetAge   string    `json:"target_age"`             // 追加
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func toCheckupTypeResponse(ct *model.CheckupType) checkupTypeResponse {
    return checkupTypeResponse{
        // ...
        ParentID:  ct.ParentID,   // 追加
        Interval:  ct.Interval,   // 追加
        TargetAge: ct.TargetAge,  // 追加
        // ...
    }
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `exam_type_response.go` | `ParentID *uint64` 追加（struct + 変換関数） |
| `checkup_type_response.go` | `ParentID *uint64`, `Interval string`, `TargetAge string` 追加（struct + 変換関数） |
