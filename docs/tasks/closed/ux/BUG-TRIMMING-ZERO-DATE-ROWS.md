# BUG-TRIMMING-ZERO-DATE-ROWS: トリミング一覧に「0001-01-01」不正行が表示される

## ステータス
🔴 **未修正**

## 優先度
Medium

## 再現手順
1. `/trimming` を開く
2. 一覧最下部を確認

## 症状
日付「0001-01-01」、飼主名・ペット名・種が空欄の行が3件表示される。
担当医カラムにオレンジの警告アイコン（⚠️）が出ている。

## 根本原因
DBに `pet_id = null`, `visit_date = null` のゴミレコードが存在する。
`id = 9, 10, 11` の3件が該当。

確認コマンド:
```sql
SELECT id, pet_id, visit_date FROM trimmings WHERE pet_id IS NULL;
```

Go のトリミング transforms が `null` の `visit_date` に対して
`0001-01-01` (Go の `time.Time` ゼロ値) を返しているため、
フロントエンドで「0001-01-01」と表示される。

## 修正方針
### BE（優先）
- トリミング作成 API で `pet_id` / `visit_date` の NOT NULL バリデーションを強化
- ゴミレコード（id=9,10,11）をDBから削除

### FE
- transforms で `visit_date` が Go のゼロ値（`"0001-01-01"` or `"0001-01-01T00:00:00Z"`）の場合は `""` に変換
  ```ts
  visitDate: d.visit_date && !d.visit_date.startsWith("0001") ? d.visit_date.split("T")[0] : "",
  ```

## 影響範囲
- `/trimming` 一覧ページ
