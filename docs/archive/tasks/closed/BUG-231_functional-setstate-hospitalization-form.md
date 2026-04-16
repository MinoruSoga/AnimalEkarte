# BUG-231: functional setState 未使用（use-hospitalization-form.ts — 3箇所）

## 概要
`features/hospitalization/hooks/use-hospitalization-form.ts` の `addTreatmentPlan`・`removeTreatmentPlan`・`updateTreatmentPlan` で、`treatmentPlans` state を deps に取り込んだ `setTreatmentPlans(直接値)` を使用している。`setState(prev => ...)` 形式を使えば `treatmentPlans` を deps から外せるため、`useCallback` のメモ化が安定し、memo() でラップされた子コンポーネントへの不要な再レンダーを防げる。

## 現状コード

### `features/hospitalization/hooks/use-hospitalization-form.ts:211,215,223`
```typescript
// ❌ treatmentPlans を stale closure 経由で読み込み
const addTreatmentPlan = () => {
  const newPlan: TreatmentPlan = { id: Date.now().toString(), ... };
  setTreatmentPlans([...treatmentPlans, newPlan]);  // line 211
};

const removeTreatmentPlan = (planId: string) => {
  setTreatmentPlans(treatmentPlans.filter(plan => plan.id !== planId));  // line 215
};

const updateTreatmentPlan = (planId: string, field: keyof TreatmentPlan, value: ...) => {
  setTreatmentPlans(
    treatmentPlans.map(plan => { ... })  // line 223
  );
};
```

これらを `useCallback` でメモ化する場合、`treatmentPlans`（配列オブジェクト）を deps に含める必要があり、毎レンダーで `useCallback` が無効化される。

## 修正方針

`prev =>` 形式に変換して `treatmentPlans` を deps から除外する。

```typescript
// ✅ functional setState — treatmentPlans を deps から外せる
const addTreatmentPlan = useCallback(() => {
  const newPlan: TreatmentPlan = { id: Date.now().toString(), ... };
  setTreatmentPlans(prev => [...prev, newPlan]);
}, []); // treatmentPlans 不要

const removeTreatmentPlan = useCallback((planId: string) => {
  setTreatmentPlans(prev => prev.filter(plan => plan.id !== planId));
}, []); // treatmentPlans 不要

const updateTreatmentPlan = useCallback(
  (planId: string, field: keyof TreatmentPlan, value: string | number | boolean) => {
    setTreatmentPlans(prev =>
      prev.map(plan => {
        if (plan.id !== planId) return plan;
        const updated = { ...plan, [field]: value };
        // subtotal 再計算ロジックをそのまま維持
        if (field === "unitPrice" || field === "quantity" || field === "discount") {
          const unitPrice = (field === "unitPrice" ? value : plan.unitPrice) as number;
          const quantity = (field === "quantity" ? value : plan.quantity) as number;
          const discount = (field === "discount" ? value : plan.discount) as number;
          updated.subtotal = Math.round(unitPrice * quantity * (1 - discount / 100));
        }
        return updated;
      })
    );
  },
  []
);
```

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — rerender-functional-setstate
> `useCallback` 内の setState は `prev =>` 形式で state を deps から外す

### プロジェクト内参照実装
`features/owners/routes/OwnerForm.tsx:handleInputChange` — `setFormData(prev => ({ ...prev, [field]: value }))` パターン

## 優先度
**Medium** — `HospitalizationTreatmentTable` が memo() ラップされている場合に効果大。治療プランの追加・更新は頻繁な操作のため対処を推奨。

## 関連ファイル
- `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts:211,215,223`
- `frontend/src/features/hospitalization/components/HospitalizationTreatmentTable.tsx` — 子コンポーネント（memo 有無を確認）
