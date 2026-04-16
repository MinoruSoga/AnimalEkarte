# BUG-235: TreatmentSearchDialog が CarePlanDialog 内で static import かつ常時レンダー

## 概要
`CarePlanDialog.tsx` は `CarePlanSection.tsx` から `lazy()` でロードされているため初期バンドルへの影響はない。しかし `TreatmentSearchDialog`（266行）が CarePlanDialog 内に static import され、`open={false}` のまま常時 DOM に存在している。`open` prop で制御する Radix Dialog パターンは、利点として内部状態の保持があるが、コンポーネントが常に初期化されバンドルサイズが CarePlanDialog チャンクに加算される。

## 現状コード

### `features/hospitalization/components/CarePlan/CarePlanDialog.tsx:17,241-245`
```typescript
// 静的 import — CarePlanDialog チャンクに同梱される
import { TreatmentSearchDialog } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

// 常時マウント (open={false} でも DOM に存在)
<TreatmentSearchDialog
    open={isSearchOpen}
    onOpenChange={setIsSearchOpen}
    onSelect={handleSelectMaster}
/>
```

### 参考: CarePlanSection.tsx:16-17 では CarePlanDialog 自体は正しく lazy
```typescript
const CarePlanDialog = lazy(() =>
  import("@/features/hospitalization/components/CarePlan/CarePlanDialog").then(...)
);
```

## 影響

- `CarePlanDialog` チャンクに `TreatmentSearchDialog` の 266 行が同梱される
- 検索ダイアログが一度も使われないセッションでも初期化コストが発生
- **BUG-225（StaffSelectionModal）と同様のパターン**

## 修正方針

条件付きレンダーに変更し、lazy で遅延ロードする。

```typescript
import { lazy, Suspense } from "react";

const TreatmentSearchDialog = lazy(() =>
  import("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog").then((m) => ({
    default: m.TreatmentSearchDialog,
  }))
);

// 条件付きレンダーに変更
{isSearchOpen ? (
  <Suspense fallback={null}>
    <TreatmentSearchDialog
        open={isSearchOpen}
        onOpenChange={setIsSearchOpen}
        onSelect={handleSelectMaster}
    />
  </Suspense>
) : null}
```

**注意**: `open` prop パターンから条件レンダーに切り替えることで、ダイアログが閉じるたびにアンマウントされる。TreatmentSearchDialog 内部に保持すべき状態がある場合は `open` prop パターンを維持してよい（その場合は static import のままで問題なし）。

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — bundle-dynamic-imports
> 重いモーダル・ダイアログは `lazy()` + `Suspense` で遅延ロード

### プロジェクト内参照実装
`features/owners/routes/OwnerForm.tsx` — `const PetEditModal = lazy(...)` + 条件レンダーパターン

## 優先度
**Low** — `CarePlanDialog` 自体はすでに lazy ロード済み。TreatmentSearchDialog は CarePlanDialog チャンク内のみに影響する。治療プラン検索を使うユーザーには常に必要なコードのため、実質的な初回ロード削減効果は限定的。

## 関連チケット
- BUG-225: StaffSelectionModal が lazy() でロードされていない（同パターン）

## 関連ファイル
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanDialog.tsx:17,241-245`
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanSection.tsx:16-17`
- `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx`
