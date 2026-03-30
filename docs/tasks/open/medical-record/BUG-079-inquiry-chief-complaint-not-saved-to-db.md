# BUG-079: カルテ問診の主訴・メモが DB に保存されない（GORM FirstOrCreate+Assign バグ）

## 種類
バグ（バックエンド — GORM FirstOrCreate+Assign の同一ポインタ渡しバグ）

## 重要度
高

## 発見日
2026-03-29

## 再現手順

1. カルテ詳細ページ（`/medical-records/:id`）を開く
2. 問診タブで主訴（chief_complaint）と メモ（notes）を入力する
3. 「保存」ボタンをクリックする
4. `PATCH /api/v1/medical-records/:id/inquiries` → HTTP 200 が返る
5. ページをリロードして入力した主訴・メモを確認する

## 期待動作

- 入力した主訴とメモが DB に保存され、リロード後も表示される

## 実際の動作

- API レスポンスは HTTP 200 で成功する
- service 層のログ `"inquiry upserted"` も出力される
- しかし **DB の値が実際には変わらない**（保存されない）
- リロード後に入力した内容が消える

## 原因

`inquiry_service.go`（または対応する service 層）の `Upsert` 処理で、
GORM の `FirstOrCreate` + `Assign` に同一ポインタを渡しているバグ。
GORM が更新対象を正しく識別できず、SELECT 成功後に UPDATE が実行されない。

```go
// ❌ 問題のあるパターン（推定）
var inquiry Inquiry
db.FirstOrCreate(&inquiry, Inquiry{MedicalRecordID: id})
db.Model(&inquiry).Assign(inquiry).Updates(&inquiry)  // 同一ポインタ渡し → 更新されない
```

## 影響範囲

- カルテ問診タブの主訴（`chief_complaint`）フィールドの保存
- カルテ問診タブのメモ（`notes`）フィールドの保存
- 問診エンドポイント全体（`PATCH /api/v1/medical-records/:id/inquiries`）

## 修正方針

`inquiry_service.go` の Upsert ロジックを修正する。
`FirstOrCreate` で既存レコードを取得後、別の変数で更新フィールドを構築して `Updates` を呼ぶ。

```go
// ✅ 修正後パターン
var inquiry Inquiry
db.FirstOrCreate(&inquiry, Inquiry{MedicalRecordID: id})
updates := map[string]any{
    "chief_complaint": input.ChiefComplaint,
    "notes":           input.Notes,
}
db.Model(&inquiry).Updates(updates)
```

## 優先度
高（問診データが一切保存できない。カルテの基本情報として致命的）

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-079（BE） | Backend | inquiry_service.go の FirstOrCreate+Updates バグ修正（同一ポインタ渡し → map[string]any に変更） |
