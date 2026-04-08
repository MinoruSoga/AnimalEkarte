# BUG-224: useEffect で derived state を同期している（2箇所）

## 概要

他の state/props から導出できる値を `useEffect` 経由で別 state に同期している箇所が 2件ある。
`useEffect` を使うと毎回余分なレンダリングサイクルが発生する。
React の原則は「レンダリング中に導出できるものは effect で同期しない」。

## 現状コード

### 1. `features/owners/hooks/use-owner-form.ts:236-239`

```typescript
// ❌ formState.fieldErrors を別の state にコピーするだけ
useEffect(() => {
  setManualErrors(formState.fieldErrors || {});
}, [formState.fieldErrors, formState.timestamp]);

// 使い方
const manualErrors = /* state */ Record<string, string>;
```

`manualErrors` は常に `formState.fieldErrors || {}` の写しであり、
独立した state として持つ必要がない。effect が実行されるたびに余分なレンダーが発生する。

なお `manualErrors` はローカルにも変更される（line 392: `setManualErrors(prev => ...)`）。
フォームフィールドのエラーをクリアする操作があるため、単純な derived value ではなく
「サーバーエラーを初期値とし、ローカルで上書き可能な状態」という設計意図がある。

### 2. `features/inventory/hooks/use-inventory-form.ts:33-37`

```typescript
// ❌ existingItem.category を category state に同期するだけ
useEffect(() => {
  if (existingItem?.category) {
    setCategory(existingItem.category as InventoryItem["category"]);
  }
}, [existingItem]);
```

`category` は `existingItem.category` から導出できるにもかかわらず、
別 state として持ち、effect で同期している。
`existingItem` が変わるたびに余分なレンダーが 2回発生する（effect → setState → rerender）。

## 比較: 正しい実装（プロジェクト内参照実装）

```typescript
// features/owners/hooks/use-owner-form.ts と同ファイル内の
// inline 同期パターン（React推奨）:
// useEffect の代わりに、前のレンダーの値を ref に保存して render 中に同期

// パターン A: レンダー中に直接導出（副作用なしで済む場合）
const manualErrors = formState.fieldErrors || {};

// パターン B: ローカル変更が必要な場合は useReducer で state machine を作る
// または、previous value パターン:
const [prevFieldErrors, setPrevFieldErrors] = useState(formState.fieldErrors);
const [manualErrors, setManualErrors] = useState(formState.fieldErrors || {});
if (prevFieldErrors !== formState.fieldErrors) {
  setPrevFieldErrors(formState.fieldErrors);
  setManualErrors(formState.fieldErrors || {});
}
// これで同一レンダー内に同期でき、余分なレンダーサイクルが不要になる
```

## 修正方針

### 1. `features/inventory/hooks/use-inventory-form.ts`（シンプルなケース）

`category` はローカルに変更されず `existingItem.category` のみから決まるため、
`useState` を廃止して直接導出する:

```typescript
// Before:
const [category, setCategory] = useState<InventoryItem["category"]>("medicine");
useEffect(() => {
  if (existingItem?.category) {
    setCategory(existingItem.category as InventoryItem["category"]);
  }
}, [existingItem]);

// After:
const category = (existingItem?.category ?? "medicine") as InventoryItem["category"];
// setCategory は削除、useEffect も削除
// formAction 内では category 変数を直接使用（すでにそうなっている）
```

### 2. `features/owners/hooks/use-owner-form.ts`（ローカル変更が必要なケース）

`manualErrors` はサーバーエラー受け取り後にローカルでクリアできる設計のため、
React の "previous value" パターンで effect を排除:

```typescript
// Before:
const [manualErrors, setManualErrors] = useState<Record<string, string>>({});
useEffect(() => {
  setManualErrors(formState.fieldErrors || {});
}, [formState.fieldErrors, formState.timestamp]);

// After:
const [prevTimestamp, setPrevTimestamp] = useState(formState.timestamp);
const [manualErrors, setManualErrors] = useState<Record<string, string>>(
  formState.fieldErrors || {}
);
// レンダー中に同期（effect 不要）
if (prevTimestamp !== formState.timestamp) {
  setPrevTimestamp(formState.timestamp);
  setManualErrors(formState.fieldErrors || {});
}
```

## 影響範囲

| ファイル | 行 | 問題 |
|---------|-----|------|
| `features/owners/hooks/use-owner-form.ts` | 236-239 | formState.fieldErrors → manualErrors の effect 同期 |
| `features/inventory/hooks/use-inventory-form.ts` | 33-37 | existingItem.category → category の effect 同期 |

## 準拠すべきプロジェクト規約・ベストプラクティス

### Vercel React Best Practices — `rerender-derived-state-no-effect`
> Don't use useEffect to sync state from props or other state.
> Derive it during render instead to avoid unnecessary re-render cycles.

### React 公式ドキュメント
> "You don't need Effects for transforming data for rendering." — react.dev/learn/you-might-not-need-an-effect

## 優先度

**Low** — 機能的な問題はない。余分なレンダーが 1-2 回発生するが、ユーザーへの影響は軽微。
`inventory` の修正は簡単（2行削除）。`owners` はより慎重な変更が必要。
