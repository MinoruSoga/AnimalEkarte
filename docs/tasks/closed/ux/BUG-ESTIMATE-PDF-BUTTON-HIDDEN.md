# BUG-ESTIMATE-PDF-BUTTON-HIDDEN: カルテ見積書タブの「PDF出力」ボタンが固定フッターに隠れる

## ステータス
✅ **修正済み**（コミット `6f847a12` fix(ux): 4件のUXバグ修正 — pb-10 → pb-24 に変更済み）

## 優先度
Medium

## 再現手順
1. `/medical-records/:id` → 「見積書」タブを開く
2. ページ末尾にスクロール
3. 「PDF出力」ボタンが右下フッター（バイタル記録・印刷・保存）に隠れていることを確認

## 症状
「PDF出力」ボタンが `fixed bottom-6 right-6 z-50` の固定フッターと重なり、
ボタンの大部分が隠れてクリック不能になる。

実測値:
- PDF出力ボタン: `bottom = 944px`, `top = 904px`
- 固定フッター: `top = 923px`, `bottom = 967px`
- → 約20px 分が重なっている

## 根本原因
`MedicalRecordEstimate.tsx` のスクロールコンテナに
`pb-10` (40px) が指定されているが、固定フッターの高さ (44px) + bottom-6 (24px) = 68px
に対して不足している。

```tsx
// MedicalRecordEstimate.tsx
<div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-10 pr-1">
```

## 修正方針
`pb-10` → `pb-24` (96px) に変更することで固定フッターとの重なりを解消する。

```tsx
<div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-24 pr-1">
```

## 影響ファイル
`frontend/src/features/medical-records/components/MedicalRecordEstimate.tsx`
