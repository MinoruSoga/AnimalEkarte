# BUG-181: `return null` による白画面問題（VaccinationForm・TrimmingForm・BillingReviewSection）

## 概要

3 つのコンポーネントでデータ未取得・条件不成立時に `return null` を使って早期リターンしており、ユーザーには**真っ白な画面**が表示される。`LoadingFallback` やエラーメッセージを表示すべきところ、何も表示されない UX 問題。BUG-163（MedicalRecordForm）と同パターンの違反。

## 再現手順

1. **VaccinationForm**: `/vaccinations/new` にアクセスし、ペットが選択されていない状態でフォームを開く
   → **結果**: 空白の画面が表示される

2. **TrimmingForm**: `/trimming/new` にアクセスし、ペット読み込み中に表示を確認
   → **結果**: ローディング中に何も表示されない

3. **BillingReviewSection**: 会計詳細で review データが null の状態を確認
   → **結果**: セクション全体が消える

## 期待する動作

- ローディング中は `<LoadingFallback />` を表示
- データなし・エラー時はメッセージまたは `<ErrorFallback />` を表示
- 決して `return null` で何も表示しないまま放置しない

## 現状コード

### `frontend/src/features/vaccinations/routes/VaccinationForm.tsx:143`
```tsx
// ❌ ペット未選択時に何も表示されない
if (!selectedPet && !isEdit) return null;
```

### `frontend/src/features/trimming/routes/TrimmingForm.tsx:536-537`
```tsx
// ❌ ペット読み込み中・未選択時に何も表示されない
if (!selectedPet && mode === "new" && petId) return null;
if (!selectedPet && mode === "new") return null;
```

### `frontend/src/features/medical-records/components/BillingReviewSection.tsx:78-79`
```tsx
// ❌ review が null の場合に何も表示されない
if (!review) return null;
```

### 比較: 正しい実装（参照実装）
```tsx
// ✅ MedicalRecords.tsx:196-197
if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback message="データの取得に失敗しました" />;

// ✅ ペット未選択の場合
if (!selectedPet && !isEdit) {
  return (
    <div className="flex items-center justify-center h-full">
      <p style={{ color: C.textSecondary }}>ペットを選択してください</p>
    </div>
  );
}
```

## 影響範囲

| 対象ファイル | 行番号 | ユーザーへの影響 | 状態 |
|---|---|---|---|
| `features/vaccinations/routes/VaccinationForm.tsx` | 143 | 白画面 / ペット未選択時 | 未修正 |
| `features/trimming/routes/TrimmingForm.tsx` | 536-537 | 白画面 / ペット読み込み中 | 未修正 |
| `features/medical-records/components/BillingReviewSection.tsx` | 78-79 | セクション消失 / review なし時 | 未修正 |

## 修正方針

### 1. `VaccinationForm.tsx:143`
```tsx
import { C } from '@/lib/design-tokens';

// Before
if (!selectedPet && !isEdit) return null;

// After
if (!selectedPet && !isEdit) {
  return (
    <div className="flex items-center justify-center p-8">
      <p style={{ color: C.textSecondary }} className="text-sm">
        ペットを選択してください
      </p>
    </div>
  );
}
```

### 2. `TrimmingForm.tsx:536-537`
```tsx
// Before
if (!selectedPet && mode === "new" && petId) return null;
if (!selectedPet && mode === "new") return null;

// After
if (!selectedPet && mode === "new" && petId) {
  return <LoadingFallback />; // petId があるがペットが読み込まれていない場合
}
if (!selectedPet && mode === "new") {
  return (
    <div className="flex items-center justify-center p-8">
      <p style={{ color: C.textSecondary }} className="text-sm">ペットを選択してください</p>
    </div>
  );
}
```

### 3. `BillingReviewSection.tsx:78-79`
```tsx
// Before
if (!review) return null;

// After
if (!review) {
  return (
    <div className="p-4 text-sm" style={{ color: C.textSecondary }}>
      会計確認情報がありません
    </div>
  );
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
ローディング中・エラー時は適切なフォールバック UI を表示すること。`return null` でユーザーに何も見せないのは UX 上の問題。

### プロジェクト内参照実装
- `features/medical-records/routes/MedicalRecords.tsx:196-197` — LoadingFallback / ErrorFallback の正しい使用例

## 優先度
**High** — ユーザーが白画面を見ることになる直接的な UX 問題。VaccinationForm と TrimmingForm は通常のユーザーフローで発生する。

## 関連チケット
- BUG-163: MedicalRecordForm の return null（同パターン）
- BUG-178: ローディング/エラー状態未処理

## 関連ファイル
- `frontend/src/features/vaccinations/routes/VaccinationForm.tsx:143`
- `frontend/src/features/trimming/routes/TrimmingForm.tsx:536-537`
- `frontend/src/features/medical-records/components/BillingReviewSection.tsx:78-79`
