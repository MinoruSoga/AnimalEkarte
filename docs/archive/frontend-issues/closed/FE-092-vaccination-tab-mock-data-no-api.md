# FE-092: 予防接種タブ - モックデータ使用・API未統合

## 概要
カルテ詳細の「予防接種」タブが `MOCK_HISTORY_ITEMS` ハードコードデータを使用しており、バックエンドAPIと統合されていない。

## 対象ファイル
- `frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx`

## 現象
- 予防接種タブを開くと常に同じモックデータが表示される
- 実際のDBデータは反映されない
- 「接種記録を追加」ボタンが存在するが動作不明（モックのため確認不可）

## 問題箇所
```tsx
// MedicalRecordVaccination.tsx
const MOCK_HISTORY_ITEMS = [...] // ハードコードされたモックデータ
```

## 期待動作
- `GET /api/v1/medical-records/{id}/vaccinations` を呼び出してデータを取得
- 予防接種の追加・編集・削除をAPIと同期

## 優先度
Medium

## 発見日
2026-03-24（機能テスト中）
