# BUG-198: Estimate・Trimming の LoadingFallback/ErrorFallback 未使用（カスタムスピナー・生テキスト）

## 概要

`features/estimates/` の 3 ファイルと `features/trimming/routes/TrimmingForm.tsx` で、ローディング状態に `<LoadingFallback />` を使わずカスタムスピナー div を使用し、エラー状態に `<ErrorFallback />` を使わず `<div className="... text-red-600">` を使用している。BUG-181/BUG-188 と同パターン。

## 再現手順

1. `/estimates`（見積一覧）を開き、API 取得中の表示を確認
   → **結果**: カスタム回転スピナー（`animate-spin rounded-full`）が表示（`LoadingFallback` 未使用）
2. `/estimates`（見積一覧）で API がエラーを返す状態を再現
   → **結果**: `<div className="p-4 text-red-600">データの取得に失敗しました</div>` が表示
3. `/trimming/:id/edit` を開き、ペット読み込み中を確認
   → **結果**: カスタム div テキスト「読み込み中...」が表示（`LoadingFallback` 未使用）

## 現状コード

### `frontend/src/features/estimates/routes/EstimateList.tsx:225-233`
```tsx
// ❌ カスタムスピナー + 生エラーテキスト
if (isLoading) {
  return (
    <div className="flex justify-center items-center p-8">
      <div className={`inline-block animate-spin rounded-full h-8 w-8 border-b-2 ${C.borderPrimary}`} />
    </div>
  );
}
if (isError) return <div className="p-4 text-red-600">データの取得に失敗しました</div>;
```

### `frontend/src/features/estimates/routes/EstimateDetail.tsx:31-39`
```tsx
// ❌ カスタムスピナー + 生エラーテキスト
if (isLoading) {
  return (
    <div className="flex justify-center items-center p-8">
      <div className={`inline-block animate-spin rounded-full h-8 w-8 border-b-2 ${C.borderPrimary}`} />
    </div>
  );
}
if (isError) return <div className="p-4 text-red-600">データの取得に失敗しました</div>;
```

### `frontend/src/features/estimates/routes/EstimateForm.tsx:284`
```tsx
// ❌ カスタムスピナー
if (isEdit && isLoading) {
  return (
    <div className="flex justify-center items-center p-8">
      <div className={`inline-block animate-spin rounded-full h-8 w-8 border-b-2 ${C.borderPrimary}`} />
    </div>
  );
}
```

### `frontend/src/features/trimming/routes/TrimmingForm.tsx:539,554`
```tsx
// ❌ 生テキストローディング（2箇所）
// L539
return (
  <div className={`flex items-center justify-center p-8 text-base ${C.text50}`}>
    <p>読み込み中...</p>
  </div>
);

// L554
<div className={`px-6 py-12 text-center text-base ${C.text50}`}>読み込み中...</div>
```

### 比較: 正しい実装
```tsx
import { LoadingFallback } from '@/components/shared/LoadingFallback';
import { ErrorFallback } from '@/components/shared/ErrorFallback';

// ✅ MedicalRecords.tsx — 標準パターン
if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback message="データの取得に失敗しました" />;
```

## 影響範囲

| 対象ファイル | 行番号 | 問題 | 状態 |
|---|---|---|---|
| `features/estimates/routes/EstimateList.tsx` | 225-233 | カスタムスピナー + `text-red-600` 生エラーテキスト | 未修正 |
| `features/estimates/routes/EstimateDetail.tsx` | 31-39 | カスタムスピナー + `text-red-600` 生エラーテキスト | 未修正 |
| `features/estimates/routes/EstimateForm.tsx` | 284 | カスタムスピナー | 未修正 |
| `features/trimming/routes/TrimmingForm.tsx` | 539, 554 | 生テキスト「読み込み中...」（2箇所） | 未修正 |

## 修正方針

### Estimate 3ファイル — 統一置換
```tsx
// Before
if (isLoading) {
  return (
    <div className="flex justify-center items-center p-8">
      <div className={`inline-block animate-spin ... ${C.borderPrimary}`} />
    </div>
  );
}
if (isError) return <div className="p-4 text-red-600">データの取得に失敗しました</div>;

// After
if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback message="データの取得に失敗しました" />;
```

### TrimmingForm.tsx:539,554
```tsx
// Before
return (
  <div className={`flex items-center justify-center p-8 text-base ${C.text50}`}>
    <p>読み込み中...</p>
  </div>
);

// After
return <LoadingFallback />;
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
ローディング中・エラー時は適切なフォールバック UI を表示すること。

## 優先度
**High** — エラー表示が `text-red-600` ハードコードかつ非標準コンポーネント使用。見積・トリミングは業務上重要な画面。

## 関連チケット
- BUG-181: VaccinationForm/TrimmingForm の return null（同パターン）
- BUG-188: HospitalizationDetail の素 div ローディング（同パターン）
- BUG-190: AccountingDetail のローディング UI（同パターン）

## 関連ファイル
- `frontend/src/features/estimates/routes/EstimateList.tsx`
- `frontend/src/features/estimates/routes/EstimateDetail.tsx`
- `frontend/src/features/estimates/routes/EstimateForm.tsx`
- `frontend/src/features/trimming/routes/TrimmingForm.tsx`
