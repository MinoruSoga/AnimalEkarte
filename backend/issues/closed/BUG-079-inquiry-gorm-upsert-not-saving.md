# BE: BUG-079 問診フィールドがDBに保存されない（GORMバグ）

## 概要

PATCH `/api/v1/medical-records/:id/inquiries` で主訴・メモを送信するとサービス層で
"inquiry upserted" ログが出るが、DBの値が実際には変わらない。
GORMの `FirstOrCreate` + `Assign` に同一ポインタを渡すバグが原因。

## 再現手順

1. カルテの問診タブで主訴を入力して保存
2. ページをリロード
3. → 入力した内容が消えている（DBに保存されていない）

## 原因

`backend/internal/service/` の inquiry upsert 処理で GORM の `FirstOrCreate` + `Assign` に
同一ポインタを渡しているため、実際の UPDATE が発生していない。

## 期待する動作

- 主訴・メモの変更が DB に正しく保存される
- 主訴フィールドの上限（1000文字）を超えた場合は 400 Bad Request を返す

## 実装場所

- `backend/internal/service/` の inquiry/medical record 関連サービス
- GORM の upsert ロジックを修正（`Save` または `Updates` を使用）

```go
// ❌ 問題のあるパターン
db.Attrs(&existing).FirstOrCreate(&existing, condition)

// ✅ 修正例
var existing Inquiry
result := db.Where(condition).First(&existing)
if result.Error == gorm.ErrRecordNotFound {
    existing = Inquiry{...newData}
    db.Create(&existing)
} else {
    db.Model(&existing).Updates(newData)
}
```

## 優先度

Critical（データが保存されない）

## 関連

- `docs/tasks/open/crash/BUG-079_inquiry_not_saved.md`
- FUNCTIONAL_TEST_REPORT.md BUG-079
