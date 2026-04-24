# BUG-LINE-006: 管理画面の予約削除が不能（BUG-LINE-005 の派生）

## 概要

Admin の `DELETE /api/v1/clinics/:id/reservations/:id` が BUG-LINE-005 の `:id` 重複により、実際の予約レコードを削除できない。

## 再現

```javascript
// 実在する予約 id=15 を削除
fetch('/api/v1/clinics/3/reservations/15', {
  method: 'DELETE', credentials: 'include'
});
// → 404 Not Found  (id=3 = clinic_id を reservation_id として検索、存在しないため 404)
```

## 影響

- 管理画面から LINE 予約含め既存予約をハード削除できない
- 結合テスト中の破損データをクリーンアップできない（staging 上に id=15 が残存中）

## 回避策

- 予約は DELETE ではなく status を cancelled に変更する運用に変更する
- または BUG-LINE-005 の修正で一括解消

## 関連

- BUG-LINE-005 を解消すれば自動的に直る
