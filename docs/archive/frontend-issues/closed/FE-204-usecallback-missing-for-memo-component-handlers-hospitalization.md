# FE-204: 入院管理で memo コンポーネントへ渡すハンドラが useCallback で安定化されていない

## 概要

`frontend/src/features/hospitalization/` 配下で、`memo()` で囲まれたコンポーネントに
props として渡すハンドラ関数が `useCallback` で安定化されていない。
memo の効果が無効化され、不要な再レンダリングが発生している。

## 問題箇所

### 1. `use-hospitalization-list.ts:56` — `handleNavigateToForm`
```ts
// Before: useCallback なし
const handleNavigateToForm = (id?: string) => {
  navigate(id ? paths.hospitalization.edit.getHref(id) : paths.hospitalization.new.getHref());
};

// 渡し先: HospitalizationBoard(memo) → CageCard(memo)
// → handleNavigateToForm が毎レンダリングで新しい参照になるため CageCard が再レンダリング

// After
const handleNavigateToForm = useCallback((id?: string) => {
  navigate(id ? paths.hospitalization.edit.getHref(id) : paths.hospitalization.new.getHref());
}, [navigate]);
```

### 2. `CarePlanSection.tsx:35-43` — `handleOpenEdit`
```tsx
// Before: useCallback なし
const handleOpenEdit = (plan: CarePlanItem) => {
  setEditTarget(plan);
  setIsEditOpen(true);
};
// 渡し先: useMemo 内で CarePlanItemRow(memo) に props で渡している
// → useMemo の deps に handleOpenEdit が含まれるため、毎回再計算される

// After
const handleOpenEdit = useCallback((plan: CarePlanItem) => {
  setEditTarget(plan);
  setIsEditOpen(true);
}, []); // setEditTarget, setIsEditOpen は安定参照なので deps 不要
```

### 3. `HospitalizationBoard.tsx:173-181` — `onNavigateToForm`, `onMovePet` の pass-through
```tsx
// Before: props を直接 CageCard(memo) に pass-through している
{cages.map((cage) => (
  <CageCard
    key={cage.id}
    cage={cage}
    onNavigateToForm={onNavigateToForm}  // 親から渡された関数をそのまま渡す
    onMovePet={onMovePet}               // 同上
  />
))}

// → HospitalizationBoard 呼び出し元で useCallback が使われていない場合、
//   毎レンダリングで CageCard に新しい参照が渡り memo が無効化される
//
// HospitalizationBoard 呼び出し元（use-hospitalization-list.ts）で
// useCallback を適用すること（上記 case 1 と組み合わせ）
```

## 影響範囲

| 対象 | 行番号 | 状態 |
|------|--------|------|
| `frontend/src/features/hospitalization/hooks/use-hospitalization-list.ts` | 56 | 要修正 |
| `frontend/src/features/hospitalization/components/CarePlan/CarePlanSection.tsx` | 35-43 | 要修正 |
| `frontend/src/features/hospitalization/components/HospitalizationBoard.tsx` | 173-181 | 要確認 |

## 修正方針

上記3箇所の関数を `useCallback` でラップし、deps を適切に設定する。
`useCallback` deps 内にオブジェクトは入れず、primitive か安定参照のみ使用すること。

```ts
// rerender-functional-setstate パターン適用
const handleOpenEdit = useCallback((plan: CarePlanItem) => {
  setEditTarget(prev => plan);   // prev 形式でなくてよいが、setter は安定
  setIsEditOpen(true);
}, []);
```

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — rerender-memo
> 独立した大きいセクションは `memo()` で囲む。**必ず props ハンドラを `useCallback` で安定化すること。**

### `.claude/rules/code-style.md` — rerender-dependencies
> `useCallback` deps にオブジェクトを入れない — primitive を抽出して使う。

## 優先度
**Low** — 機能的障害はない。パフォーマンス最適化。

## 関連チケット
- FE-203: 共有コンポーネントに memo() 未適用

## 関連ファイル
- `frontend/src/features/hospitalization/hooks/use-hospitalization-list.ts:56`
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanSection.tsx:35-43`
- `frontend/src/features/hospitalization/components/HospitalizationBoard.tsx:173-181`
