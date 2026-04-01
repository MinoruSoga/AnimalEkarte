# BUG-117: 主訴マスタカテゴリ削除 → FK チェックなし (204) — inquiries が孤立

## 概要
`DELETE /api/v1/masters/chief-complaint-categories/:id` が 204 を返し、
`inquiries.chief_complaint_category_id` で参照されていても削除成功してしまう。
削除後にその ID を参照しているカルテ問診レコードが孤立する。

## 再現手順
1. `/settings/interview/chief-complaint` を開く
2. カルテ問診で主訴として使用されているカテゴリ（例: 食欲不振 id=1）の編集パネルを開く
3. 削除ボタン → 確認ダイアログ「削除する」をクリック
4. → HTTP 204 で削除成功。`inquiries` テーブルの `chief_complaint_category_id` が孤立。

## 実確認
- ローカル確認: 2026-04-01
- `DELETE http://localhost:8080/api/v1/masters/chief-complaint-categories/1` → **204**
- 「食欲不振」が一覧から消え、5件→5件（削除前6件）

## 期待動作
- `inquiries.chief_complaint_category_id` で参照されている場合 → **409 Conflict**
- エラーメッセージ: 「この主訴は問診記録で使用中のため削除できません」
- 参照なしの場合のみ 204 で削除成功

## 影響範囲
- `inquiries` テーブルの `chief_complaint_category_id` 外部キー参照が孤立
- カルテ Tab1 の主訴表示が壊れる可能性

## 類似バグ
- BUG-113: 診断病名マスタ削除 FK チェックなし（修正済み）と同根パターン

## 担当
Backend: `internal/handler/master_handler.go`（DeleteChiefComplaintCategory）
→ service 層で `inquiries` テーブルの参照チェックを追加する
