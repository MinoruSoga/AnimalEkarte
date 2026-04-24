# FE-094: 画像タブ - モックデータ使用・アップロード未実装

## 概要
カルテ詳細の「画像」タブがモックデータを使用しており、バックエンドAPIと統合されていない。「画像アップロード」機能も未実装。

## 対象ファイル
- `frontend/src/features/medical-records/components/MedicalRecordImage.tsx`（推定）

## 現象
1. 画像タブはモックデータを表示（APIコールなし）
2. 「画像アップロード」ボタン: 機能しない（未実装）
3. 実際のDB画像データは表示されない

## 期待動作
- `GET /api/v1/medical-records/{id}/images` でデータ取得
- `POST /api/v1/medical-records/{id}/images` でアップロード
- ファイル選択 → プレビュー → アップロード処理

## 優先度
Medium

## 発見日
2026-03-24（機能テスト中）
