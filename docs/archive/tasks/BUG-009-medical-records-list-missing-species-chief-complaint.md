# BUG-009: カルテ一覧の「種別」「主訴」カラムが空欄

## 種類
バグ（バックエンド API 不備）

## 発見日
2026-03-23

## 再現手順
1. カルテ一覧 `/medical-records` を開く
2. 「種別」「主訴」カラムを確認

## 期待動作
- 種別: 「犬」「猫」等、ペットの動物種が表示される
- 主訴: カルテに登録された主訴テキストが表示される

## 実際の動作
- 種別: 空欄（データなし）
- 主訴: 空欄（データなし）

## 根本原因
`GET /v1/medical-records` のバックエンド実装で以下の Preload が欠落している：

```go
// 欠落している Preload
db.Preload("Pet.AnimalSpecies")  // → 種別
db.Preload("Inquiry")            // → 主訴
```

## 影響範囲
- カルテ一覧ページ全体で種別・主訴が表示されない
- 重症度: **中**（一覧の視認性が著しく低下）

## 修正方針
`backend/internal/repository/medical_record_repository.go` の
カルテ一覧取得クエリに以下を追加：

```go
db.Preload("Pet").
   Preload("Pet.AnimalSpecies").
   Preload("Inquiry").
   Where("clinic_id = ?", clinicID).
   Find(&records)
```

また `GET /v1/medical-records` レスポンス型（`*_response.go`）に
`species_name` / `chief_complaint` フィールドを追加してフロントに返す。

## 関連
- バックエンドイシュー: `backend/issues/open/` に BE-052 として起票すること
  （既存の BE-052 ファイルが `backend/issues/open/` にあれば流用可）
