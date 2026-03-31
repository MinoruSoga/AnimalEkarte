# BUG-MEDI-003: 鑑別診断（DDx）専用フィールドがない

## 概要
ClinicalPlanSection に鑑別診断（DDx）専用の入力フィールドが存在しない。
現状は `diagnosisDetails` フィールドのみで代用しているが、DDx と確定診断は別管理が必要。

## 期待する動作
- DDx 専用テキストエリアまたは複数入力フィールドを追加
- 鑑別候補のリスト表示・追加・削除

## 実装場所
- `frontend/src/features/medical-records/` の ClinicalPlanSection コンポーネント
- バックエンドモデルに `ddx_notes` フィールド追加が必要な場合はマイグレーション

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md line 2580
- テスト確認日: 2026-03-30
