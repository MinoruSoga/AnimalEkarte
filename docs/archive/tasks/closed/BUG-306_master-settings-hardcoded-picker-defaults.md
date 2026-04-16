# BUG-306: master設定 — カラーピッカーデフォルト値ハードコード

## 概要

`PermissionGroupSettings.tsx` と `ServiceTypeSettings.tsx` の初期状態にHexカラーがハードコードされている。`PALETTE` に定数を追加して参照させる。

## 影響ファイル

| ファイル | 違反箇所 |
|---------|---------|
| `frontend/src/features/master/routes/PermissionGroupSettings.tsx` | line 85 |
| `frontend/src/features/master/routes/ServiceTypeSettings.tsx` | line 56 |

## 違反箇所と修正

### design-tokens.ts に追加
```ts
// PALETTE に追加
pickerDefaultGray: "#6B7280",   // Tailwind gray-500 — 権限グループのデフォルトカラー
pickerDefaultBlue: "#3B82F6",   // Tailwind blue-500 — サービス種別のデフォルトカラー
```

### PermissionGroupSettings.tsx
```tsx
// Before (line 85)
color: item?.color ?? "#6B7280",

// After
color: item?.color ?? PALETTE.pickerDefaultGray,
```

### ServiceTypeSettings.tsx
```tsx
// Before (line 56)
color: item?.color ?? "#3B82F6",

// After
color: item?.color ?? PALETTE.pickerDefaultBlue,
```

## 適用ルール

- デザイントークン規約: Hexカラー直接指定禁止。`PALETTE.*` を使用すること

## ステータス

✅ 修正済み
