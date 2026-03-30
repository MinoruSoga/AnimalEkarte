# BUG-068: 治療数量0でAPI直接送信すると500エラー

## 概要
治療タブの数量フィールドに0を入力してAPI（PATCH）を直接送信すると500 Internal Server Errorが返る。
フロントエンドは min=0.1 制約で0入力を防ぐが、バックエンドは400でなく500を返す。
また `parseFloat("0")||1` パターンで0が1にフォールバックされエラーメッセージもなし。

## 再現手順
1. `PATCH /api/v1/treatments/:id` に `{"quantity": 0}` を送信
2. → HTTP 500 Internal Server Error（本来は400であるべき）

## 期待する動作
- バックエンド: quantity ≤ 0 の場合 400 Bad Request + "quantity must be greater than 0"
- フロントエンド: 0入力時に明示的エラーメッセージ表示（min制約だけでなくエラーメッセージも）

## 実装場所
- `backend/internal/service/treatment_service.go` または `handler/treatment_handler.go`
- フロントエンド: `frontend/src/features/medical-records/` の TreatmentRow コンポーネント

## 優先度
High（500エラーは常に不正）

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-068
- テスト確認日: 2026-03-30
