# BUG-237: rerender-lazy-state-init — use-hospitalization-form.ts の treatmentPlans 初期値

## 概要
`use-hospitalization-form.ts` で `useState` に 2 つのオブジェクトリテラルからなる配列を直接渡している。`useState` は初回レンダー時にしか初期値を使わないが、オブジェクトリテラルはレンダーのたびに評価・生成される。`() =>` の lazy 形式にすることで初回以外の無駄なアロケーションを排除できる。

## 現状コード

### `features/hospitalization/hooks/use-hospitalization-form.ts:54-77`
```typescript
// ❌ 配列・オブジェクトリテラルがレンダーごとに生成される
const [treatmentPlans, setTreatmentPlans] = useState<TreatmentPlan[]>([
  {
    id: "1",
    treatmentContent: "adm rate",
    memo: "入院料1日分",
    insurance: true,
    unitPrice: 990,
    quantity: 1,
    discount: 0,
    discountAmount: 0,
    subtotal: 990,
  },
  {
    id: "2",
    treatmentContent: "PCG/SC ~15kg",
    memo: "",
    insurance: false,
    unitPrice: 990,
    quantity: 1,
    discount: 0,
    discountAmount: 0,
    subtotal: 990,
  },
]);
```

## 修正方針

初期値をモジュールレベルの定数に抽出するか、lazy initializer で囲む。

```typescript
// ✅ 方法1: モジュール定数（最も効果的）
const DEFAULT_TREATMENT_PLANS: TreatmentPlan[] = [
  {
    id: "1",
    treatmentContent: "adm rate",
    memo: "入院料1日分",
    insurance: true,
    unitPrice: 990,
    quantity: 1,
    discount: 0,
    discountAmount: 0,
    subtotal: 990,
  },
  {
    id: "2",
    treatmentContent: "PCG/SC ~15kg",
    memo: "",
    insurance: false,
    unitPrice: 990,
    quantity: 1,
    discount: 0,
    discountAmount: 0,
    subtotal: 990,
  },
];

const [treatmentPlans, setTreatmentPlans] = useState<TreatmentPlan[]>(DEFAULT_TREATMENT_PLANS);

// ✅ 方法2: lazy initializer（モジュール定数が難しい場合）
const [treatmentPlans, setTreatmentPlans] = useState<TreatmentPlan[]>(() => [
  { id: "1", treatmentContent: "adm rate", ... },
  { id: "2", treatmentContent: "PCG/SC ~15kg", ... },
]);
```

**方法1推奨**: モジュール定数にすれば全レンダーで同一参照を共有でき、GC 圧力を完全に排除できる。

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — rerender-lazy-state-init
> 高コストな `useState` 初期化は `useState(() => ...)` lazy 形式を使用

### プロジェクト内参照実装
`features/estimates/hooks/use-estimate-form.ts:51` — `useState<EstimateFormState>(() => buildInitialState(estimate))` のように lazy 形式を使用

## 優先度
**Medium** — オブジェクト生成コストは軽微だが、入院フォームは頻繁に表示される画面のため対処推奨。修正は5分。

## 関連チケット
- BUG-231: use-hospitalization-form.ts functional setState 未使用（同ファイル）
- BUG-232: use-hospitalization-form.ts useEffect object deps（同ファイル）

## 関連ファイル
- `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts:54-77`
