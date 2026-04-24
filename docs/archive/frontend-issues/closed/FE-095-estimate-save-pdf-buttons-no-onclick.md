# FE-095: 見積書タブ - フォーム内「保存」「PDF出力」ボタンのonClick未実装

## 概要
カルテ詳細の「見積書」タブ内のフォームに「保存」「PDF出力」ボタンが存在するが、両方とも `onClick` ハンドラが未実装。見積書データはAPIに保存されない。

## 対象ファイル
- `frontend/src/features/medical-records/components/MedicalRecordEstimate.tsx`

## 現象
1. 見積書フォームの「保存」ボタン: `onClick` なし → クリックしても何も起きない（または誤動作）
2. 見積書フォームの「PDF出力」ボタン: `onClick` なし → 機能しない
3. フローティングバーとフォーム内ボタンのz-index重複問題（see also FE-096）

## 問題箇所
```tsx
// MedicalRecordEstimate.tsx (lines 120-134)
<Button>保存</Button>    // onClick なし
<Button>PDF出力</Button> // onClick なし
```

## 期待動作
- 「保存」: `POST/PUT /api/v1/estimates` で見積書データを保存
- 「PDF出力」: 見積書PDFを生成してダウンロード
- 見積書件名、明細行、コメント、備考が全て保存される

## 関連
- FE-096: フローティングバーz-index重複バグ

## 優先度
High（見積書機能が完全に使えない）

## 発見日
2026-03-24（機能テスト中）
