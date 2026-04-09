# BUG-238: rerender-lazy-state-init — VitalsTab EditRow の computed useState

## 概要
`VitalsTab.tsx` の `EditRow` コンポーネント（memo() 適用済み）で、`vital` prop の複数フィールドを変換・結合した初期フォームオブジェクトを `useState` に直接渡している。`new Date(...).toISOString().slice(0, 16)` や `String(vital.xxx)` などの変換処理がレンダーのたびに実行される。lazy initializer `() => (...)` で初回レンダーのみに限定すべきである。

## 現状コード

### `features/medical-records/components/VitalsTab/VitalsTab.tsx:95-106`
```typescript
const EditRow = memo(function EditRow({ vital, onSave, onCancel, isPending }: EditRowProps) {
  // ❌ computed object がレンダーごとに生成される
  const [form, setForm] = useState({
    recorded_at: vital.recorded_at
      ? new Date(vital.recorded_at).toISOString().slice(0, 16)  // Date 生成 + 文字列操作
      : "",
    temperature: vital.temperature != null ? String(vital.temperature) : "",
    heart_rate: vital.heart_rate != null ? String(vital.heart_rate) : "",
    respiratory_rate: vital.respiratory_rate != null ? String(vital.respiratory_rate) : "",
    body_weight: vital.body_weight != null ? String(vital.body_weight) : "",
    weight_unit: vital.weight_unit ?? "Kg",
    note: vital.note ?? "",
  });
```

## 修正方針

初期値生成を関数に抽出し、lazy initializer で囲む。

```typescript
// ✅ 初期化ロジックを分離
function buildInitialForm(vital: VitalRecord) {
  return {
    recorded_at: vital.recorded_at
      ? new Date(vital.recorded_at).toISOString().slice(0, 16)
      : "",
    temperature: vital.temperature != null ? String(vital.temperature) : "",
    heart_rate: vital.heart_rate != null ? String(vital.heart_rate) : "",
    respiratory_rate: vital.respiratory_rate != null ? String(vital.respiratory_rate) : "",
    body_weight: vital.body_weight != null ? String(vital.body_weight) : "",
    weight_unit: vital.weight_unit ?? "Kg",
    note: vital.note ?? "",
  };
}

const EditRow = memo(function EditRow({ vital, onSave, onCancel, isPending }: EditRowProps) {
  // ✅ lazy initializer — 初回マウント時のみ実行
  const [form, setForm] = useState(() => buildInitialForm(vital));
```

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — rerender-lazy-state-init
> 高コストな `useState` 初期化は `useState(() => ...)` lazy 形式を使用

### プロジェクト内参照実装
`features/estimates/hooks/use-estimate-form.ts:51` — `useState(() => buildInitialState(estimate))`

## 優先度
**Low** — `EditRow` は `memo()` 済みで頻繁な再レンダーは起きにくい。しかし `new Date(...)` 生成とひとかたまりのオブジェクト作成がレンダーのたびに走るため、修正してリテラルコストを排除すべき。修正は10分。

## 関連ファイル
- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx:95-106`
