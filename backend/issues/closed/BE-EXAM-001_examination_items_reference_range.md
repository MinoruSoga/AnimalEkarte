# BE-EXAM-001: 検査項目テーブル（examination_items）のリファレンスレンジ・異常判定対応

## 概要
`examination_items` テーブルに基準値（参照範囲）と異常フラグのカラムを追加し、
検査結果の異常値判定をバックエンドで行えるようにする。

## 背景
`docs/screens/13-examinations-form.md` では検査項目テーブルとして
「項目名 / 結果 / 単位 / 基準値 / 異常フラグ」の表示が想定されている。
現状の `examination_items` は `name`, `result`, `unit`, `ref`（文字列）のみ。
異常判定はバックエンドで数値比較する必要がある。

## 実装内容

### migration
```sql
-- 基準値を数値で管理できるようにする
ALTER TABLE examination_items
  ADD COLUMN ref_min DECIMAL(10,4),
  ADD COLUMN ref_max DECIMAL(10,4),
  ADD COLUMN is_abnormal BOOLEAN DEFAULT FALSE;
```

### model (`backend/internal/model/examination.go`)
```go
RefMin      *float64 `gorm:"column:ref_min"                    json:"ref_min,omitempty"`
RefMax      *float64 `gorm:"column:ref_max"                    json:"ref_max,omitempty"`
IsAbnormal  bool     `gorm:"column:is_abnormal;default:false"  json:"is_abnormal"`
```

### 自動異常判定ロジック（service 層）
```go
// result が数値かつ ref_min/ref_max が設定されている場合、自動判定
func judgeAbnormal(result string, refMin, refMax *float64) bool {
    val, err := strconv.ParseFloat(result, 64)
    if err != nil || refMin == nil || refMax == nil {
        return false
    }
    return val < *refMin || val > *refMax
}
```

### API
- `POST /api/v1/examinations/:id/items` の request body に `ref_min`, `ref_max` を追加
- `PATCH /api/v1/examinations/:id/items/:item_id` でも更新可能に
- レスポンスに `is_abnormal` を含める

### codegen
モデル変更後に `make codegen` を実行。

## 優先度
Low

## 関連
- 仕様書: `docs/screens/13-examinations-form.md`
- フロントエンド型: `frontend/src/types/index.ts` の `ExaminationItem`
