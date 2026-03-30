# BUG-MEDI-006: 来院回数・最終来院日の表示がない

## 概要
カルテ一覧・詳細画面に同一ペットの来院回数と最終来院日の表示フィールドがない。
来院頻度グラフ（月別）も未実装。

## 期待する動作
- ペットの過去来院回数をカルテ詳細または PatientInfoCard に表示
- 最終来院日を表示
- 来院頻度グラフ（月別）表示

## 実装場所
- `frontend/src/features/medical-records/` のカルテ詳細コンポーネント
- バックエンドで集計クエリ（`COUNT(*)` / `MAX(visit_date)`）を追加

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md lines 2512-2514
- テスト確認日: 2026-03-30
