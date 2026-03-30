# BUG-MEDI-009: カルテ保存後にダッシュボードの診療中カードが更新されない

## 概要
カルテの保存処理にアポイントメント status 更新の呼び出しがない。
カルテを保存・完了しても、ダッシュボードの「診療中」カンバンから患者が消えない。

## 期待する動作
- カルテ完了（保存）時にアポイントメント status を「完了」に更新
- ダッシュボードの「診療中」カードから自動的に消える

## 実装場所
- `frontend/src/features/medical-records/` のカルテ保存ロジック
- カルテ保存成功後に `PATCH /api/v1/appointments/:id` で status を更新

## 優先度
Medium

## 関連
- FUNCTIONAL_TEST_REPORT.md line 2619
- テスト確認日: 2026-03-30
