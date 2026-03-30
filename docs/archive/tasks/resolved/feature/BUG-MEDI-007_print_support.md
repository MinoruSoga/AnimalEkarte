# BUG-MEDI-007: カルテの印刷対応が未実装

## 概要
カルテページに `@media print` スタイルがない（accounting のみ実装済み）。
病院ロゴ・住所の印刷ヘッダー、治療計画の印刷・飼主向け説明書出力が未実装。

## 期待する動作
- `@media print` スタイルを追加してカルテを印刷可能にする
- 印刷ヘッダーに病院ロゴ・住所を表示
- 「印刷」ボタン追加

## 実装場所
- `frontend/src/features/medical-records/` の CSS / スタイリング
- 病院設定から logo/address を取得して印刷ヘッダーに表示

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md lines 2461, 2464, 2399
- テスト確認日: 2026-03-30
