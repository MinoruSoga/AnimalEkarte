# TASK-014: フロントエンド MEDIUM 違反 3件の修正

## 概要

フロントエンド監査で検出された MEDIUM 優先度の違反 3件をまとめて対応する。

## 優先度

MEDIUM

---

## 違反 1: `VitalsTab.sortedVitals` が `useMemo` なしで毎回ソート

### ファイル
`frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx:315-320`

### 問題
```typescript
// 現状: レンダリングのたびに新配列生成 + ソート実行
const sortedVitals: Vital[] = vitals ? [...vitals].sort((a, b) =>
  new Date(a.recorded_at).getTime() - new Date(b.recorded_at).getTime()
) : [];
```

`VitalsTab` 自体は `memo()` 適用済みだが、内部で不要な計算が走る。

### 修正案
```typescript
const sortedVitals = useMemo(
  () => vitals ? [...vitals].sort((a, b) =>
    new Date(a.recorded_at).getTime() - new Date(b.recorded_at).getTime()
  ) : [],
  [vitals]
);
```

---

## 違反 2: `useInventory` フックの命名が規約に不適合

### ファイル
`frontend/src/features/inventory/hooks/use-inventory.ts:31`

### 問題
API データフェッチ + transform + フィルタリングを内包するフックの命名が `useInventory`。規約ではデータ取得系フックは `useGet` / `useFetch` 等の動詞プレフィックスを推奨している。`useGetInventoryItems` が既に存在しており、命名の重複感がある。

### 修正案
```typescript
// Before: hooks/use-inventory.ts
export function useInventory() { ... }

// After: hooks/use-inventory-list.ts
export function useInventoryList() { ... }
```

`index.ts` の re-export も合わせて更新すること。

---

## 違反 3: `auth/hooks/use-auth.tsx` の Fast Refresh 警告

### ファイル
`frontend/src/features/auth/hooks/use-auth.tsx`

### 問題
`react-refresh/only-export-components` ESLint 警告が発生。`AuthContext` と `AuthProvider` と `useAuth` フックを同一ファイルから export しているため、HMR（Hot Module Replacement）の精度が低下する。

### 修正案
```
auth/hooks/
├── use-auth.tsx         ← useAuth フックのみ（既存）
├── auth-context.ts      ← AuthContext の型定義・作成
└── auth-provider.tsx    ← AuthProvider コンポーネント
```

または、ESLint 設定で `use-auth.tsx` を `react-refresh/only-export-components` の例外に追加する（影響範囲が小さければこちらでも可）:
```js
// eslint.config.js
{ files: ["src/features/auth/hooks/use-auth.tsx"], rules: { "react-refresh/only-export-components": "off" } }
```
