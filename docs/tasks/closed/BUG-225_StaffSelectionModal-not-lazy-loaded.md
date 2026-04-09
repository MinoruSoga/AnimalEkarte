# BUG-225: StaffSelectionModal が lazy() でロードされていない

## 概要
`StaffSelectionModal` は条件付きで表示される重いモーダルコンポーネントだが、`lazy()` + `Suspense` で遅延ロードされていない。初回バンドルに含まれ、不要なJSを初期ロード時に読み込んでいる。

## 現状コード

### 使用箇所（複数）
```typescript
// ❌ 静的 import — 初回バンドルに含まれる
import { StaffSelectionModal } from "@/components/shared/StaffSelectionModal";

// 使用
{isStaffModalOpen ? (
  <StaffSelectionModal
    open={isStaffModalOpen}
    onClose={() => setIsStaffModalOpen(false)}
    onSelect={handleStaffSelect}
  />
) : null}
```

## 修正方針

`lazy()` + `Suspense` で遅延ロードする。

```typescript
// ✅ lazy import
import { lazy, Suspense } from "react";

const StaffSelectionModal = lazy(() =>
  import("@/components/shared/StaffSelectionModal").then(m => ({
    default: m.StaffSelectionModal,
  }))
);

// 使用
{isStaffModalOpen ? (
  <Suspense fallback={null}>
    <StaffSelectionModal
      open={isStaffModalOpen}
      onClose={() => setIsStaffModalOpen(false)}
      onSelect={handleStaffSelect}
    />
  </Suspense>
) : null}
```

## 影響範囲

| ファイル | 役割 |
|---------|------|
| StaffSelectionModal を使用する全コンポーネント | 静的 import を lazy に変更 |

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — bundle-dynamic-imports
> 重いモーダル・ダイアログは `lazy()` + `Suspense` で遅延ロード

### プロジェクト内参照実装
`features/owners/routes/OwnerForm.tsx` — `const PetEditModal = lazy(...)` パターン

## 優先度
**Low** — バンドルサイズ最適化。機能的影響なし。修正は10分。

## 関連ファイル
- `frontend/src/components/shared/StaffSelectionModal/` — コンポーネント本体
- StaffSelectionModal を import している全ファイル
