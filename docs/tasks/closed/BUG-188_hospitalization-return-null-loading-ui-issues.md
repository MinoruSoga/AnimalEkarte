# BUG-188: 入院管理の白画面・粗雑なローディング UI 問題（HospitalizationForm・HospitalizationDetail）

## 概要

`HospitalizationForm.tsx` がペット未選択・読み込み中に `return null` で白画面を返す。`HospitalizationDetail.tsx` はローディング時に素の `<div>読み込み中...</div>` を表示しており、プロジェクト標準の `<LoadingFallback />` を使っていない。BUG-163（MedicalRecordForm）・BUG-181（VaccinationForm/TrimmingForm）と同一パターン。

## 再現手順

1. **HospitalizationForm**: `/hospitalization/new?petId=xxx` にアクセスし、ペットデータ読み込み中の画面を確認
   → **結果**: 読み込み中に白画面が表示される（`return null`）

2. **HospitalizationDetail**: `/hospitalization/:id` にアクセスし、データ取得中の表示を確認
   → **結果**: 「読み込み中...」のテキストが素の div として表示される

## 現状コード

### `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:133-134`
```tsx
// ❌ ペット読み込み中・未選択時に白画面
if (!selectedPet && !isEdit && petId) return null;
if (!selectedPet && !isEdit) return null;
```

### `frontend/src/features/hospitalization/routes/HospitalizationDetail.tsx:44`
```tsx
// ❌ 素のテキストでローディング表示
return <div className="p-8 text-center text-gray-500">読み込み中...</div>;
```

**注**: `text-gray-500` も Tailwind プリセット使用（`C.textSecondary` を使うべき）

### 比較: 正しい実装
```tsx
// ✅ MedicalRecords.tsx — 標準 LoadingFallback 使用
import { LoadingFallback } from '@/components/shared/LoadingFallback';

if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback message="データの取得に失敗しました" />;

// ✅ ペット未選択時
if (!selectedPet && !isEdit && petId) return <LoadingFallback />;
if (!selectedPet && !isEdit) {
  return (
    <div className="flex items-center justify-center p-8">
      <p style={{ color: C.textSecondary }}>ペットを選択してください</p>
    </div>
  );
}
```

## 影響範囲

| 対象ファイル | 行番号 | 問題 | 状態 |
|---|---|---|---|
| `features/hospitalization/routes/HospitalizationForm.tsx` | 133-134 | `return null` — 白画面 | 未修正 |
| `features/hospitalization/routes/HospitalizationDetail.tsx` | 44 | 素 div ローディング + `text-gray-500` | 未修正 |

## 修正方針

### 1. `HospitalizationForm.tsx:133-134`
```tsx
// Before
if (!selectedPet && !isEdit && petId) return null;
if (!selectedPet && !isEdit) return null;

// After
if (!selectedPet && !isEdit && petId) return <LoadingFallback />;
if (!selectedPet && !isEdit) {
  return (
    <div className="flex items-center justify-center p-8">
      <p style={{ color: C.textSecondary }}>ペットを選択してください</p>
    </div>
  );
}
```

### 2. `HospitalizationDetail.tsx:44`
```tsx
// Before
return <div className="p-8 text-center text-gray-500">読み込み中...</div>;

// After
return <LoadingFallback />;
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
ローディング中・エラー時は適切なフォールバック UI を表示すること。

## 優先度
**High** — 入院管理フォームの白画面は通常のユーザーフローで発生しうる UX 問題。

## 関連チケット
- BUG-163: MedicalRecordForm の return null（同パターン）
- BUG-181: VaccinationForm/TrimmingForm の return null（同パターン）

## 関連ファイル
- `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx`
- `frontend/src/features/hospitalization/routes/HospitalizationDetail.tsx`
