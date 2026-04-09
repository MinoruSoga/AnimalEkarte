---
name: BUG-243_accounting-detail-usememo-object-dep
description: AccountingDetail.tsx の clinicForDocument useMemo deps に user?.clinic オブジェクト
type: bug
---

# BUG-243: AccountingDetail — `clinicForDocument` useMemo の deps に `user?.clinic` オブジェクト

## 概要

`AccountingDetail.tsx:980` の `clinicForDocument` useMemo が
deps に `user?.clinic`（オブジェクト）を含む。`rerender-dependencies` 違反。

オブジェクトは参照比較のため、`useAuth()` が新しい `user` 参照を返すたびに
useMemo が不要に再実行される可能性がある。

## 現状コード

### `features/accounting/routes/AccountingDetail.tsx:973-980`

```tsx
const clinicForDocument = useMemo(() => {
  const baseClinic = user?.clinic ?? null;
  if (!baseClinic) return null;
  return {
    ...baseClinic,
    invoiceRegistrationNumber,
  };
}, [user?.clinic, invoiceRegistrationNumber]);  // ← user?.clinic はオブジェクト
```

`user?.clinic` はオブジェクト参照。`useAuth()` コンテキストが再レンダーされると
新しい参照が生成され、useMemo が毎回再実行される場合がある。

## 期待する動作

deps に primitive を使い、真に値が変わったときのみ useMemo を再実行する。

## 修正方針

### 方針 A: `user` 全体を deps に（より安定した参照）

```tsx
const clinicForDocument = useMemo(() => {
  const baseClinic = user?.clinic ?? null;
  if (!baseClinic) return null;
  return {
    ...baseClinic,
    invoiceRegistrationNumber,
  };
}, [user, invoiceRegistrationNumber]);
// user オブジェクト全体の参照が安定していれば問題なし
```

### 方針 B: clinic の primitive フィールドを個別に deps に列挙

```tsx
const clinicId   = user?.clinic?.id;
const clinicName = user?.clinic?.name;
// ... 必要なフィールドを列挙

const clinicForDocument = useMemo(() => {
  const baseClinic = user?.clinic ?? null;
  if (!baseClinic) return null;
  return {
    ...baseClinic,
    invoiceRegistrationNumber,
  };
}, [clinicId, clinicName, /* ...primitives */, invoiceRegistrationNumber]);
```

`useAuth()` の実装が `user` オブジェクトを安定参照で返す場合は方針 A で十分。

## 参照実装

`features/hospitalization/hooks/use-hospitalization-form.ts:170-173`（BUG-232 修正箇所）:
```tsx
// rerender-dependencies: hospitalizationData（オブジェクト）の代わりに id（primitive）を deps に使用
}, [hospitalizationData?.id]);
```

## 影響範囲

| ファイル | 行 | 内容 |
|---------|-----|------|
| `features/accounting/routes/AccountingDetail.tsx:973-980` | useMemo deps にオブジェクト | 修正必要 |

## 優先度

**Low** — `useAuth()` が安定した `user` 参照を返す実装であれば実害はないが、
規約準拠として `user?.clinic`（オブジェクト）を deps から除去すべき。

## 関連チケット

- BUG-232: 同一パターン（hospitalization + estimate の useEffect deps）
- BUG-222: useCallback deps にオブジェクト（9箇所）
