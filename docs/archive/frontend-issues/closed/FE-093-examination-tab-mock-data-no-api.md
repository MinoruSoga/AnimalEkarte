# FE-093: 検査タブ - モックデータ使用・API未統合・ボタン未実装

## 概要
カルテ詳細の「検査」タブが `MOCK_EXAM_GROUPS` ハードコードデータを使用しており、バックエンドAPIと統合されていない。またインタラクティブ機能も未実装。

## 対象ファイル
- `frontend/src/features/medical-records/components/MedicalRecordExamination.tsx`

## 現象
1. 検査タブは `MOCK_EXAM_GROUPS` 定数からデータを表示（APIコールなし）
2. 「詳細を表示」ボタン: `onClick` ハンドラなし（クリックしても何も起きない）
3. 「検査取り込み」ボタン: `onClick` ハンドラなし

## 問題箇所
```tsx
// MedicalRecordExamination.tsx
const MOCK_EXAM_GROUPS = [...] // ハードコードモック

// ExaminationGroup component
<Button>詳細を表示</Button>    // onClick なし
<Button>検査取り込み</Button>  // onClick なし
```

## 期待動作
- `GET /api/v1/medical-records/{id}/examinations` を呼び出してデータを取得
- 「詳細を表示」: 検査詳細モーダルを開く
- 「検査取り込み」: 外部検査システムからデータを取り込む機能

## 優先度
Medium

## 発見日
2026-03-24（機能テスト中）
