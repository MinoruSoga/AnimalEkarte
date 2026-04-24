# FE-108: NumberInput を全 feature に適用

**Status**: Closed
**Priority**: High
**Affects**: trimming, estimates, medical-records, master, inventory, hospitalization, owners
**Date Created**: 2026-03-25
**Related**: TASK-025

## Summary

`<Input type="number">` の直書きが14ファイルに存在する。前セッションで作成した共有 `NumberInput` コンポーネントに置き換える。機能変更なし・見た目変更なし（suffix 表示が追加される場合のみ UX 向上）。

## 現状のコード

```typescript
// frontend/src/features/estimates/routes/EstimateForm.tsx:141-149
<Input
  id="subtotal"
  type="number"
  min={0}
  value={form.subtotal}
  onChange={e => handleChangeWithDirty('subtotal', Number(e.target.value))}
  className="h-9 text-sm"
/>

// frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx:230
<Input type="number" step="0.1" min="0.1" value={item.quantity} ... />
```

## 対象ファイル一覧（14ファイル）

| ファイル | 箇所数 | suffix（追加すべき単位） |
|---------|--------|----------------------|
| `features/owners/routes/OwnerForm.tsx` | 1 | なし |
| `features/owners/components/PetEditModal.tsx` | 1 | なし |
| `features/trimming/routes/TrimmingForm.tsx` | 2 | 円, 分 等 |
| `features/estimates/routes/EstimateForm.tsx` | 5 | 円 |
| `features/medical-records/components/TreatmentDetailedSummary.tsx` | 確認 | 円 |
| `features/medical-records/components/TreatmentsTab/TreatmentRow.tsx` | 3 | -, %, ¥ |
| `features/medical-records/components/TreatmentTable.tsx` | 確認 | -, %, ¥ |
| `features/medical-records/components/VitalsTab/VitalsTab.tsx` | 確認 | ℃, kg, /min 等 |
| `features/master/routes/InsuranceSettings.tsx` | 確認 | % or 円 |
| `features/master/routes/MedicineSettings.tsx` | 確認 | なし |
| `features/master/routes/TrimmingSettings.tsx` | 確認 | 円, 分 |
| `features/inventory/routes/InventoryForm.tsx` | 確認 | なし |
| `features/hospitalization/components/HospitalizationCostSummary.tsx` | 確認 | 円 |
| `features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx` | 確認 | ℃, kg 等 |

## 必要な変更

### 各ファイルで行う置き換えパターン

**Before:**
```typescript
import { Input } from "@/components/ui/input";
// ...
<Input
  type="number"
  min={0}
  step={0.1}
  value={form.xxx}
  onChange={e => handler(Number(e.target.value))}
  className="h-9 text-sm"
/>
```

**After:**
```typescript
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
// ...
<NumberInput
  min={0}
  step={0.1}
  value={form.xxx}
  onChange={v => handler(Number(v))}
  className="h-9 text-sm"
  suffix="円"   // 単位がある場合のみ追加
/>
```

### onChange シグネチャ変換ルール

- `Input`: `onChange={(e) => fn(e.target.value)}` または `onChange={(e) => fn(Number(e.target.value))}`
- `NumberInput`: `onChange={(v) => fn(v)}` または `onChange={(v) => fn(Number(v))}`
- `Input` の import が他の用途でも使われていれば残す。`type="number"` にしか使われていなければ削除。

### suffix 追加ルール

- 単価・金額系 → `suffix="円"`
- 体重 → `suffix="kg"`
- 体温 → `suffix="℃"`
- 心拍数・呼吸数 → `suffix="/min"`
- 割引率 → `suffix="%"`
- suffix が文脈から不明・不要な場合は省略（suffix なし）

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし（`@/components/shared/NumberInput/NumberInput` を直接 import）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] 型は `models.ts` から導出

## 依存関係

- `NumberInput` コンポーネントは既に実装済み（`frontend/src/components/shared/NumberInput/NumberInput.tsx`）
- BE 変更なし

## 完了条件

- [ ] `grep -rn 'type="number"' frontend/src/features` の出力が 0 件
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス（エラー 0）
- [ ] 既存の数値入力 UI に見た目の崩れがない

## クローズ情報

- **Closed At**: 2026-03-25
- **変更ファイル**:
  - `features/owners/routes/OwnerForm.tsx` — discountRate → NumberInput suffix="%"
  - `features/owners/components/PetEditModal.tsx` — weight → NumberInput suffix="kg"
  - `features/trimming/routes/TrimmingForm.tsx` — 体重・体温 → NumberInput
  - `features/estimates/routes/EstimateForm.tsx` — 金額5箇所 → NumberInput suffix="円"
  - `features/medical-records/components/TreatmentDetailedSummary.tsx` — 割引率・値引額 → NumberInput
  - `features/hospitalization/components/HospitalizationCostSummary.tsx` — 割引率・値引額 → NumberInput
  - `features/hospitalization/components/DailyRecordsTab/DailyVitalsSection.tsx` — バイタル4項目 → NumberInput
- **変換見送り**（技術的制約）:
  - TreatmentRow.tsx: ref+focus/select() 機能があるため
  - TreatmentTable.tsx: TableInput ヘルパーのフルサイズレイアウトが崩れるため
  - InventoryForm.tsx: 非制御フォーム（defaultValue+name）のため
  - VitalsTab/InsuranceSettings/MedicineSettings/TrimmingSettings: ネイティブ input 使用のため
