# FEAT-003: 問診履歴の実API連携未実装

## 概要
カルテの問診履歴（InterviewHistory）がハードコードデータを表示しており、
実際の過去カルテデータと連携していない。

## 該当箇所
- `frontend/src/features/medical-records/` の InterviewHistory コンポーネント
  - `DEFAULT_HISTORY_ITEMS` 等のハードコード定数から表示

## 期待する動作
- 同一ペットの過去カルテ一覧を実APIから取得して表示
- 過去処方薬の参照（アレルギー確認用）
- 問診内容の引用（前回来院から今回へのコピー）
- 受診履歴リストから特定回の詳細表示

## 必要な実装
1. `GET /api/v1/pets/:id/medical-records` エンドポイント（履歴取得）
2. InterviewHistory コンポーネントの実API接続
3. 引用機能（前回問診内容のコピー）

## 優先度
Medium（診療継続性に重要）

## 関連
- FUNCTIONAL_TEST_REPORT.md: 過去処方薬参照・受診履歴 NG 項目（2026-03-28）
