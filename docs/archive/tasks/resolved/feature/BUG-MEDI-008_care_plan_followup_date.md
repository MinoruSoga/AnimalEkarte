# BUG-MEDI-008: 治療計画に次回来院指示フィールドがない

## 概要
ClinicalPlanSection に次回来院日・内容の入力フィールドが存在しない。
フォローアップ日もカルテモデルに `follow_up_date` フィールドがなく、ダッシュボード表示も未実装。

## 期待する動作
- 治療計画に「次回来院日」「次回来院内容」フィールドを追加
- `follow_up_date` をバックエンドモデルに追加
- ダッシュボードに近日フォローアップ患者リストを表示

## 実装場所
- `frontend/src/features/medical-records/` の ClinicalPlanSection
- `backend/internal/model/medical_record.go` に `follow_up_date` フィールド追加
- マイグレーション追加

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md lines 2268, 2397
- テスト確認日: 2026-03-30
