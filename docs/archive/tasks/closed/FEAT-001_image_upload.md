# FEAT-001: 画像アップロード機能未実装

## 概要
カルテの「画像」タブにおける画像アップロード機能が未実装。
アップロードボタンに `onClick` ハンドラが存在しない。

## 該当箇所
- `frontend/src/features/medical-records/` の画像タブコンポーネント
  - `ImageGalleryFilter.tsx`: アップロードボタンに onClick なし
  - `ImageGalleryGroup.tsx`: img要素なし、labelテキスト表示のみ

## 期待する動作
- PNG/JPG/PDF ファイルを選択してアップロードできる
- 複数ファイル同時アップロード対応
- アップロード後にサムネイル表示
- 拡大表示（ライトボックス）
- 10MB 超ファイルの制限とエラー表示
- 画像削除ボタン

## 必要な実装
1. Backend: `/api/v1/medical-records/:id/images` エンドポイント（POST/GET/DELETE）
2. Frontend: ファイル選択 → S3/ローカルアップロード → サムネイル表示

## 優先度
Medium

## 関連
- FUNCTIONAL_TEST_REPORT.md: 画像アップロード関連 NG 項目多数（2026-03-28）
