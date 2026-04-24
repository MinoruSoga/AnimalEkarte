# FE-203: 共有コンポーネント群に memo() が未適用

## 概要

`frontend/src/components/shared/` 配下の複数の中〜大型共有コンポーネントに `memo()` が適用されていない。
プロジェクト規約「新規共有コンポーネントも `memo()` を適用すること」に違反。
頻繁に使われる共有コンポーネントが memo なしのため、不要な再レンダリングが発生している。

## 影響ファイル

| ファイルパス | コンポーネント名 | 使用箇所 | 優先度 |
|------------|--------------|---------|--------|
| `components/shared/PageLayout/PageLayout.tsx` | `PageLayout` | ほぼ全ページ | 高 |
| `components/shared/MasterSelectModal/MasterSelectModal.tsx` | `MasterSelectModal` | trimming, 各マスタ選択 | 高 |
| `components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx` | `TreatmentSearchDialog` | カルテ・入院 3箇所 | 高 |
| `components/shared/ConfirmDialog/ConfirmDialog.tsx` | `ConfirmDialog` | 削除確認 多数 | 中 |
| `components/shared/PatientInfoCard/PatientInfoCard.tsx` | `PatientInfoCard` | カルテ・診察 | 中 |
| `components/shared/ReservationFormModal/ReservationFormModal.tsx` | `ReservationFormModal` | 予約フォーム | 中 |
| `components/shared/PetSelection/PetSelection.tsx` | `PetSelection` | ペット選択画面 | 中 |

## 現状コード

```tsx
// MasterSelectModal.tsx (現状) — memo なし
export function MasterSelectModal({ open, ... }: MasterSelectModalProps) {
  return <Dialog ...>...</Dialog>;
}

// TreatmentSearchDialog.tsx (現状) — memo なし
export function TreatmentSearchDialog({ open, ... }: TreatmentSearchDialogProps) {
  return <CommandDialog ...>...</CommandDialog>;
}

// PageLayout.tsx (現状) — memo なし
export function PageLayout({ children, title, ... }: PageLayoutProps) {
  return <div ...>...</div>;
}
```

## 修正方針

各ファイルで `memo()` でラップする。

```tsx
// MasterSelectModal.tsx — After
import { memo } from "react";

export const MasterSelectModal = memo(function MasterSelectModal({
  open,
  onOpenChange,
  title,
  items,
  onSelect,
  selectedValue,
  searchPlaceholder = "検索...",
  matchBy = "name",
}: MasterSelectModalProps) {
  // ... 既存実装
});
```

```tsx
// TreatmentSearchDialog.tsx — After
export const TreatmentSearchDialog = memo(function TreatmentSearchDialog({
  open,
  onOpenChange,
  onSelect,
}: TreatmentSearchDialogProps) {
  // ... 既存実装
});
```

```tsx
// PageLayout.tsx — After
export const PageLayout = memo(function PageLayout({
  children,
  title,
  // ... props
}: PageLayoutProps) {
  // ... 既存実装
});
```

同様に `ConfirmDialog`, `PatientInfoCard`, `ReservationFormModal`, `PetSelection` も適用。

## 注意事項

- `memo()` 適用後、props として渡す関数ハンドラは呼び出し元で `useCallback` で安定化が必要
- `Layout.tsx` と `Sidebar.tsx` はアプリルートに1回配置されるため対象外でよい

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Shared Component memo()
> `DataTable`, `NotionFilter`, `Pagination`, `SidePeekPanel` は `memo()` 適用済み。
> **新規共有コンポーネントも同様に適用すること。**

### `.claude/rules/code-style.md` — Performance Rules
> 独立した大きいセクションは `memo()` で囲む。必ず props ハンドラを `useCallback` で安定化すること。

### プロジェクト内参照実装
- `frontend/src/components/shared/DataTable/DataTable.tsx` — `memo()` 適用済み参照

## 優先度
**Medium** — 機能的障害はないが、全ページで不要な再レンダリングが発生する可能性がある。

## 関連チケット
- FE-204: useCallback 未使用（memo コンポーネントへのハンドラ）

## 関連ファイル
- `frontend/src/components/shared/PageLayout/PageLayout.tsx`
- `frontend/src/components/shared/MasterSelectModal/MasterSelectModal.tsx`
- `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx`
- `frontend/src/components/shared/ConfirmDialog/ConfirmDialog.tsx`
- `frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx`
- `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx`
- `frontend/src/components/shared/PetSelection/PetSelection.tsx`
