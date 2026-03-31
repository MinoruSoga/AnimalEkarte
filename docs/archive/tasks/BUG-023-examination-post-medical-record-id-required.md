# BUG-023: 検査登録フォームから `POST /api/v1/examinations` が 400 エラー（medical_record_id 必須判定）

## 種類
バグ（バックエンドバリデーション過剰 + エラーメッセージタイポ）

## 重要度
高

## 発見日
2026-03-28

## 再現手順

1. 検査管理メニュー（`/examinations`）から「新規登録」をクリック
2. 検査項目・結果値を入力して「保存」をクリック
3. POST リクエストが送信される

## 期待動作

- `medical_record_id` を指定しない（null）スタンドアロン検査として HTTP 201 で登録される
- 検査管理メニューからカルテ紐付けなしで検査を独立登録できる

## 実際の動作

```json
POST /api/v1/examinations
{"medical_record_id": null, ...}
→ HTTP 400
{"error": "medical_record_i_d is required"}
```

- バックエンドが `medical_record_id` を必須扱いにしている
- エラー文字列に `_i_d` というタイポあり（`medical_record_id` ではなく `medical_record_i_d`）

## 原因

`POST /api/v1/examinations` のリクエストバインディングで `binding:"required"` が付与されている。
スタンドアロン検査（カルテ紐付けなし）はシステム仕様上サポートされるべきだが、
バリデーションが必須扱いになっている。

## 影響範囲

- 検査管理メニューからの独立した検査登録フローが完全に機能しない
- 検査一覧・結果値表示がすべて BUG-023 依存のためテスト不可

## 修正方針

1. `examination_handler.go` の `CreateExamination` リクエストバインディングで `medical_record_id` を `omitempty` に変更
2. エラー文字列のタイポ `medical_record_i_d` → `medical_record_id` を修正

## 優先度
高（検査機能全体が利用不可）

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-023（BE） | Backend | `medical_record_id` を任意フィールドに変更・エラーメッセージタイポ修正 |
