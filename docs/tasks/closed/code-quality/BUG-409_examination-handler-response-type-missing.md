# BUG-409: examination_handler がモデルを直接返却している（Response型ラップ欠落）

## 概要

`examination_handler.go` の Get/Create/Update エンドポイントが、他のマスタハンドラと異なり、
`toExaminationResponse()` でラップせず `*model.Examination` を直接 `c.JSON()` に渡している。
これにより API レスポンス仕様が不統一になり、不要なフィールドが露出するリスクがある。

## 問題箇所

```go
// examination_handler.go:86
c.JSON(http.StatusOK, exam)   // ← *model.Examination 直接返却

// examination_handler.go:133
c.JSON(http.StatusCreated, exam)

// examination_handler.go:189
c.JSON(http.StatusOK, exam)
```

## 期待する実装

他のマスタハンドラと同様に Response 型でラップすること。

```go
// 他のマスタハンドラの標準パターン（animal_species_handler.go:21）
c.JSON(http.StatusOK, mapSlice(species, toAnimalSpeciesResponse))

// medicine_handler.go:51
c.JSON(http.StatusOK, toMedicineResponse(medicine))
```

## 修正方針

1. `backend/internal/handler/examination_response.go` を新規作成
2. `toExaminationResponse(exam *model.Examination) examinationResponse` 関数を実装
   - 必要フィールドのみを明示的にマッピング
   - `model.Examination` の内部フィールドを外部に露出しない
3. `examination_handler.go` の3箇所を `toExaminationResponse()` 呼び出しに変更

## 影響ファイル

- `backend/internal/handler/examination_handler.go` — 行 86, 133, 189
- `backend/internal/handler/examination_response.go` — 新規作成

## 優先度

**Medium** — API 仕様不統一。不要フィールド露出リスク。

## 関連チケット

- BUG-395（マスタハンドラテスト未実装）
