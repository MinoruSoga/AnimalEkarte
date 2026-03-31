---
status: closed
closed_at: 2026-03-16
---

# [master] DiagnosisSettings: lucide-react バレルインポート（tree-shaking 阻害）

## 優先度
中

## 種別
パフォーマンス・コード品質

## 対象ファイル
`frontend/src/features/master/routes/DiagnosisSettings.tsx`

## 問題

L20-24 で `lucide-react` のバレルインポートを使用している。

```tsx
// 現状（バレルインポート — 禁止）
import { Plus, FolderTree, ClipboardList } from "lucide-react";
```

`TreatmentPlanMaster.tsx` が直接 ESM パスからインポートしているのと対照的であり、
プロジェクトの `bundle-barrel-imports` ルール（「直接ファイルから import」）に違反している。

```tsx
// 正しいパターン（TreatmentPlanMaster.tsx の実装）
import Plus from "lucide-react/dist/esm/icons/plus";
import FolderTree from "lucide-react/dist/esm/icons/folder-tree";
import ClipboardList from "lucide-react/dist/esm/icons/clipboard-list";
```

## 修正方針
バレルインポートを各アイコンの直接 ESM パスインポートに変更する。

## 完了条件
- [x] `from "lucide-react"` のインポートが削除されている
- [x] 各アイコンが `lucide-react/dist/esm/icons/xxx` から直接インポートされている
- [x] ビルドエラーなし
