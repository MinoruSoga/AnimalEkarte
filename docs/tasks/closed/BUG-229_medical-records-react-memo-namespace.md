# BUG-229: `React.memo()` 名前空間プレフィックス使用（medical-records 12 件 + HistoryFilterPanel 1 件）

## 概要

`features/medical-records/components/` 配下の 12 ファイルおよび `components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx` の計 13 ファイルが `import React from "react"` + `React.memo()` の形式を使用している。プロジェクト規約（参照実装 `features/owners/`）では `import { memo } from "react"` で名前付きインポートし、`memo()` を直接呼び出すことが標準。`React.` 名前空間プレフィックスは不統一であり、不要な `React` デフォルトインポートも含む。

## 再現手順

（ランタイム動作は変わらないが、コードの一貫性違反として確認可能）

1. 各ファイルを開き `React.memo(` を検索する
2. **結果**: 12 ファイルで `React.memo(function ...)` パターンが使用されている
3. **期待**: `memo(function ...)` + `import { memo } from "react"`

## 期待する動作

- すべての `memo()` 呼び出しは `import { memo } from "react"` で名前付きインポートし、`React.` プレフィックスなしで使用する

## 現状コード

### 影響を受ける 13 ファイル（抜粋）

```
components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx:31
features/medical-records/components/ExaminationGroup.tsx:30
features/medical-records/components/ExaminationFilter.tsx:24
features/medical-records/components/ImageGalleryFilter.tsx:29
features/medical-records/components/VaccinationHistory.tsx:21
features/medical-records/components/EstimateForm.tsx:15
features/medical-records/components/TreatmentDetailedSummary.tsx:21
features/medical-records/components/DiagnosisHeaderDiagnosis.tsx:31
features/medical-records/components/DiagnosisHeaderPhysicalExam.tsx:19
features/medical-records/components/DiagnosisHeaderChiefComplaint.tsx:16
features/medical-records/components/StaffSelectionModal.tsx:14
features/medical-records/components/ImageGalleryGroup.tsx:28
features/medical-records/components/DiagnosisHeader.tsx:28
```

#### `ExaminationGroup.tsx:1-3,30`（代表例）
```tsx
// ❌ 現状: React デフォルトインポート + React.memo()
import React from "react";
// ...
export const ExaminationGroup = React.memo(function ExaminationGroup({
```

### 比較: 正しい実装（`features/owners/routes/OwnerForm.tsx:2,52`）
```tsx
// ✅ 正しい: 名前付きインポート + memo()
import { useState, lazy, Suspense, memo, useCallback, useEffect } from "react";
// ...
const MembershipTypeButtons = memo(function MembershipTypeButtons({
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx:31` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/ExaminationGroup.tsx:30` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/ExaminationFilter.tsx:24` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/ImageGalleryFilter.tsx:29` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/VaccinationHistory.tsx:21` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/EstimateForm.tsx:15` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/TreatmentDetailedSummary.tsx:21` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/DiagnosisHeaderDiagnosis.tsx:31` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/DiagnosisHeaderPhysicalExam.tsx:19` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/DiagnosisHeaderChiefComplaint.tsx:16` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/StaffSelectionModal.tsx:14` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/ImageGalleryGroup.tsx:28` | `React.memo()` 使用 | 未修正 |
| `features/medical-records/components/DiagnosisHeader.tsx:28` | `React.memo()` 使用 | 未修正 |

## 修正方針

各ファイルについて以下の 2 ステップを適用する。

### ステップ 1: React デフォルトインポートを削除し `memo` を名前付きインポートに追加

```tsx
// Before
import React from "react";

// After（他の named import があれば追記）
import { memo } from "react";
// または既存の named import に memo を追加
import { useCallback, memo } from "react";
```

### ステップ 2: `React.memo(` を `memo(` に置換

```tsx
// Before
export const ExaminationGroup = React.memo(function ExaminationGroup({

// After
export const ExaminationGroup = memo(function ExaminationGroup({
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — `Shared Component memo()`
> `DataTable`, `NotionFilter`, `Pagination`, `SidePeekPanel` は `memo()` 適用済み。新規共有コンポーネントも同様に適用すること。

### `.claude/rules/typescript-react.md` — React 19 Patterns
> コンポーネントは関数宣言で定義（`FC`型は使わない）

`React.memo()` ではなく `memo()` を直接使うのが React 19 のコーディングスタイル。デフォルトインポート `import React` は React 17 以前の慣習。

### プロジェクト内参照実装
- `frontend/src/features/owners/routes/OwnerForm.tsx:2,52,89,223` — `memo()` の正しい使用例（名前付きインポート）

## 優先度
**Low** — ランタイム動作は同一。コードスタイルの一貫性の問題。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/components/shared/HistoryFilterPanel/HistoryFilterPanel.tsx:31`
- `frontend/src/features/medical-records/components/` — 対象 12 ファイル
- `frontend/src/features/owners/routes/OwnerForm.tsx` — 参照実装
