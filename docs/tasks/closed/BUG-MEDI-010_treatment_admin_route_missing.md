# BUG-MEDI-010: 薬品追加後の投与方法フィールドと自動フォーカスが未実装

## 概要
TreatmentsTab で薬品を選択・追加した後、投与方法フィールドが存在しない。
追加後に数量・投与方法フィールドへの自動フォーカスもない。

## 期待する動作
- 薬品選択後、投与方法（経口・注射・外用など）のフィールドを表示
- 追加後に数量フィールドへ自動フォーカス

## 実装場所
- `frontend/src/features/medical-records/` の TreatmentsTab コンポーネント
- 投与方法フィールド追加と `useRef` / `focus()` による自動フォーカス

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md line 2078
- テスト確認日: 2026-03-30
