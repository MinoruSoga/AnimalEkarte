# FE-208: ClinicalPlanSection の診断カテゴリ/病名フィールドが自由入力テキストで ID を破損させる

## 概要

`ClinicalPlanSection.tsx` の診断カテゴリと診断病名のフィールドが `<input type="text">` で実装されている。
表示値としてカテゴリ名（文字列）を表示するが、ユーザーが入力した内容が
`diagnosisCategoryId` / `diagnosisNameId` の state に保存され、
`Number(diagnosisCategoryId)` で数値変換されて API に送信される。

ユーザーが名前テキストを編集すると `Number("消化器疾患") === NaN` となり、API に `null` または `NaN` が送信される。
**診断データが無音で破損する。**

## 再現手順

1. 任意のカルテを開く → 「診察/治療プラン」タブ
2. 「診察所見・診断・治療方針」セクションを確認
3. 「診断カテゴリ」フィールドに任意のテキストを入力（例: "犬"）
4. カルテを保存
5. **結果**: `diagnosis_category_id: NaN` または `null` が API に送信され、診断カテゴリが消える

## 現状コード

### `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx:88-119`

```tsx
{/* 診断カテゴリ */}
<input
  type="text"
  value={
    data?.diagnosis_category
      ? data.diagnosis_category.name  // ← カテゴリ"名"を表示
      : diagnosisCategoryId           // ← IDを表示（型不整合）
  }
  onChange={(e) => setDiagnosisCategoryId(e.target.value)}  // ← ユーザー入力をIDとして保存
  placeholder="カテゴリを選択"
  disabled={!canEdit}
/>

{/* 診断病名 */}
<input
  type="text"
  value={
    data?.diagnosis_name
      ? data.diagnosis_name.name  // ← 病名"名"を表示
      : diagnosisNameId           // ← IDを表示（型不整合）
  }
  onChange={(e) => setDiagnosisNameId(e.target.value)}  // ← ユーザー入力をIDとして保存
  placeholder="病名を選択"
  disabled={!canEdit}
/>
```

### 保存処理 (`ClinicalPlanSection.tsx:44-54`)
```tsx
const handleSave = useCallback(async (): Promise<void> => {
  const input: UpdateClinicalPlanInput = {
    diagnosis_category_id: diagnosisCategoryId
      ? Number(diagnosisCategoryId)  // ← "消化器疾患" → NaN
      : null,
    diagnosis_name_id: diagnosisNameId
      ? Number(diagnosisNameId)      // ← "感染症" → NaN
      : null,
  };
  await updateMutation.mutateAsync(input);
}, [...]);
```

## 根本原因

診断カテゴリと診断病名は**マスタ参照フィールド**（ID で管理）なのに、
自由入力 `<input type="text">` で実装されている。

- 表示: 名前（文字列）を見せる
- 保存: ID（数値）を送る

この2つが1つの `<input>` に混在しており整合性がない。

## 影響範囲

| 対象 | 行番号 | 状態 |
|------|--------|------|
| `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx` | 88-119 | 要修正 |
| `frontend/src/features/medical-records/api/clinical-plan.ts` | UpdateClinicalPlanInput 型 | 要確認 |

## 修正方針

診断カテゴリと診断病名フィールドを **`MasterSelectTrigger` + `MasterSelectModal`** パターンに置き換える。
選択時は ID を state に保存し、表示は選択した項目の名前を `MasterSelectTrigger` が担う。

```tsx
// After: MasterSelectTrigger + MasterSelectModal パターン
import { MasterSelectModal } from "@/components/shared/MasterSelectModal/MasterSelectModal";
import { MasterSelectTrigger } from "@/components/shared/MasterSelectModal/MasterSelectTrigger";
import { useGetAllDiagnosisCategories } from "@/features/master/api/diagnosis-categories";

const { data: diagnosisCategories = [] } = useGetAllDiagnosisCategories();

// state: ID を数値で管理
const [diagnosisCategoryId, setDiagnosisCategoryId] = useState<number | null>(null);

// 選択済みカテゴリ名を表示
const selectedCategory = diagnosisCategories.find(c => c.id === diagnosisCategoryId);

<MasterSelectTrigger
  selectedItem={selectedCategory ? { name: selectedCategory.name } : undefined}
  placeholder="カテゴリを選択"
  onClick={() => setIsCategoryModalOpen(true)}
  disabled={!canEdit}
/>
<MasterSelectModal
  open={isCategoryModalOpen}
  onOpenChange={setIsCategoryModalOpen}
  title="診断カテゴリ選択"
  items={diagnosisCategories}
  onSelect={(item) => setDiagnosisCategoryId(Number(item.id))}
  selectedValue={diagnosisCategoryId}
  matchBy="id"
/>
```

**saveに渡す値は数値 ID のみ。名前文字列は API に送らない。**

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — 型安全性最優先
> TypeScript で `any` を禁止し、厳格な型定義を行う。
> マスタ参照フィールドは ID で管理し、表示名と保存値を明確に分離すること。

### プロジェクト内参照実装
- `frontend/src/features/trimming/routes/TrimmingForm.tsx:95-102` — `MasterSelectTrigger` + `MasterSelectModal` の正しい実装

## 優先度
**High** — ユーザーが診断カテゴリ/病名を編集すると診断データが無音で消去される。
データ破損バグ。

## 関連ファイル
- `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx:88-119`
- `frontend/src/components/shared/MasterSelectModal/MasterSelectModal.tsx`
- `frontend/src/components/shared/MasterSelectModal/MasterSelectTrigger.tsx`
- `frontend/src/features/master/api/diagnosis-categories.ts` — API（要確認）
