# BUG-392: exam_type_response.examTypeItemResponse に Unit フィールドが欠落

## 概要
`exam_type_response.go` の `examTypeItemResponse` 構造体が `model.ExamTypeField` の `Unit` フィールドを含んでいない。DB モデルには `Unit string` が定義されており `not null;default:''` 制約があるが、API レスポンスでは常に欠落する。クライアント側で検査項目の単位（例: "mg/dL", "kg"）を表示できない。

## 再現手順
1. 検査種別（exam_type）を取得する `GET /masters/exam-types/:id`
2. レスポンスの `items[]` オブジェクトを確認する
3. **結果**: `unit` フィールドが欠落している

## 期待する動作
`items[]` の各オブジェクトに `unit` フィールドが含まれる:
```json
{
  "id": 1,
  "exam_type_id": 10,
  "name": "体重",
  "inspection_value": "",
  "normal_value": "3.0-5.0",
  "unit": "kg",
  "sort_order": 0,
  "created_at": "..."
}
```

## 現状コード

### `backend/internal/handler/exam_type_response.go:9-17`
```go
type examTypeItemResponse struct {
    ID              uint64    `json:"id"`
    ExamTypeID      uint64    `json:"exam_type_id"`
    Name            string    `json:"name"`
    InspectionValue string    `json:"inspection_value"`
    NormalValue     string    `json:"normal_value"`
    // Unit フィールドが存在しない ← 欠落
    SortOrder       int       `json:"sort_order"`
    CreatedAt       time.Time `json:"created_at"`
}
```

### `backend/internal/model/examination_type.go:ExamTypeField`
```go
type ExamTypeField struct {
    ID              uint64    `json:"id"`
    ExamTypeID      uint64    `json:"exam_type_id"`
    Name            string    `json:"name"`
    InspectionValue string    `json:"inspection_value"`
    NormalValue     string    `json:"normal_value"`
    Unit            string    `gorm:"not null;default:''"  json:"unit"`  // ← モデルには存在
    SortOrder       int       `json:"sort_order"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

## 修正方針

### 1. `backend/internal/handler/exam_type_response.go:9-17` — Unit フィールド追加
```go
type examTypeItemResponse struct {
    ID              uint64    `json:"id"`
    ExamTypeID      uint64    `json:"exam_type_id"`
    Name            string    `json:"name"`
    InspectionValue string    `json:"inspection_value"`
    NormalValue     string    `json:"normal_value"`
    Unit            string    `json:"unit"`             // ← 追加
    SortOrder       int       `json:"sort_order"`
    CreatedAt       time.Time `json:"created_at"`
}
```

### 2. `toExamTypeItemResponse` 関数（33-43行目）— フィールドマッピング追加
```go
func toExamTypeItemResponse(item *model.ExamTypeField) examTypeItemResponse {
    return examTypeItemResponse{
        ID:              item.ID,
        ExamTypeID:      item.ExamTypeID,
        Name:            item.Name,
        InspectionValue: item.InspectionValue,
        NormalValue:     item.NormalValue,
        Unit:            item.Unit,                    // ← 追加
        SortOrder:       item.SortOrder,
        CreatedAt:       item.CreatedAt,
    }
}
```

## 優先度
**Medium** — 機能上は動作しているが、クライアントが検査項目の単位情報を取得できない API 契約の欠落。検査結果表示 UI に影響する。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/exam_type_response.go:9-43` — 修正対象
- `backend/internal/model/examination_type.go:ExamTypeField` — モデル定義（正）
