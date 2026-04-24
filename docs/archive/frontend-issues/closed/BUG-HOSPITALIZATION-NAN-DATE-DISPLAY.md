# BUG-HOSPITALIZATION-NAN-DATE-DISPLAY

## 概要
入院詳細ページのデイリーカルテタブで日付が「NaN年NaN月NaN日（undefined）」と表示される。

## 優先度
HIGH

## 再現手順
1. 入院中（退院日未設定）の入院レコードの詳細ページを開く (`/hospitalization/:id`)
2. 「デイリーカルテ」タブを選択
3. DailyDateNav の日付表示が「NaN年NaN月NaN日（undefined）」になる

## 根本原因

### 原因チェーン
1. `transforms.ts` の `formatDate(null)` が `"-"` (文字列) を返す
2. `HospitalizationTabbedView.tsx:29` と `HospitalizationExpandedView.tsx:28` で:
   ```ts
   const dischargeDate = hospitalization.endDate || new Date().toISOString().split("T")[0];
   ```
   `endDate` が `"-"` (truthy) のため、`today` ではなく `"-"` が `dischargeDate` に入る
3. `DailyDateNav` が `"-"` を受け取り `new Date("-")` を実行 → `NaN`
4. 日付フォーマット関数が NaN を処理して「NaN年NaN月NaN日（undefined）」と表示

### 影響ファイル
- `frontend/src/features/hospitalization/api/transforms.ts` — `formatDate(null)` が `"-"` を返す
- `frontend/src/features/hospitalization/components/HospitalizationTabbedView.tsx:29`
- `frontend/src/features/hospitalization/components/HospitalizationExpandedView.tsx:28`

## 修正方針

### Option A（推奨）: `endDate` の null チェックを強化
```ts
// HospitalizationTabbedView.tsx & HospitalizationExpandedView.tsx
const dischargeDate = (hospitalization.endDate && hospitalization.endDate !== "-")
  ? hospitalization.endDate
  : new Date().toISOString().split("T")[0];
```

### Option B: `transforms.ts` で `formatDate(null)` を `null` のまま返す
`formatDate(null)` が `"-"` ではなく `null | undefined` を返すよう変更し、呼び出し側で表示用の変換を行う（影響範囲が広いため慎重に）

Option A のほうが影響範囲が小さく安全。

## 確認方法
- 退院日未設定の入院レコードの詳細ページを開く
- デイリーカルテタブで正常な今日の日付が表示されること
- 退院日が設定済みの場合、その日付が正しく表示されること
