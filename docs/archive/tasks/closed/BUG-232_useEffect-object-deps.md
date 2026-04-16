# BUG-232: useEffect deps にオブジェクトを指定（hospitalization + estimate — 2件）

## 概要
`useEffect` の依存配列にオブジェクト型の値を指定している箇所が2件ある。オブジェクトは毎レンダーで新しい参照を生成するため、useEffect が意図より頻繁に再実行される可能性がある。また ESLint の `react-hooks/exhaustive-deps` を意図せず回避するコメントが付いている箇所もある。

## 現状コード

### `features/hospitalization/hooks/use-hospitalization-form.ts:135-168`
```typescript
// ❌ hospitalizationData（オブジェクト）を deps に指定
useEffect(() => {
  if (!hospitalizationData) return;
  setFormData(prev => ({
    ...prev,
    hospitalizationType: hospitalizationData.hospitalization_type === "hospitalization"
      ? "入院" : "ホテル",
    // ... 多数のフィールド
  }));
  if (hospitalizationData.pet && hospitalizationData.owner_id) {
    setSelectedPets([{ id: String(hospitalizationData.pet_id), ... }]);
  }
}, [hospitalizationData, setSelectedPets]); // hospitalizationData はオブジェクト
```

### `features/estimates/hooks/use-estimate-form.ts:54-59`
```typescript
// ❌ estimate（オブジェクト）を deps に指定
useEffect(() => {
  if (estimate) {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setForm(buildInitialState(estimate));
  }
}, [estimate]); // estimate はオブジェクト
```

## 影響・リスク評価

- **hospitalization**: `hospitalizationData` は `useGetHospitalizationRaw(id)` の戻り値。React Query はキャッシュヒット時に同一参照を返すため、実害は軽微。ただし `id` を deps に使った方が意図が明確。
- **estimate**: `estimate` は親から渡される props。親が再レンダーするたびに新規オブジェクトが渡される可能性があり、フォームの意図しないリセットが発生しうる。

## 修正方針

### `use-hospitalization-form.ts` — id を deps に使う
```typescript
const hospitalizationId = id; // primitive

useEffect(() => {
  if (!hospitalizationData) return;
  setFormData(prev => ({ ...prev, /* ... */ }));
  // ...
}, [hospitalizationId]); // オブジェクトではなく id で制御
// ※ setFormData は安定参照なので deps 省略可
```

### `use-estimate-form.ts` — estimate?.id を deps に使う
```typescript
const estimateId = estimate?.id;

useEffect(() => {
  if (estimate) {
    setForm(buildInitialState(estimate));
  }
}, [estimateId]); // id が変わった時のみ再初期化
```

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — rerender-dependencies
> `useCallback` / `useEffect` deps にオブジェクトを入れない — primitive を抽出して使う

## 優先度
**Medium** — `use-estimate-form.ts` は意図しないフォームリセットのリスクがある。先に対処すべき。

## 関連ファイル
- `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts:135-168`
- `frontend/src/features/estimates/hooks/use-estimate-form.ts:54-59`
