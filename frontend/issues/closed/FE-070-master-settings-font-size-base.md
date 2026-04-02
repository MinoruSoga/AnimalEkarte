# FE-070: マスタ設定ページのフォントサイズ最低 text-base 化

**Status**: Open
**Priority**: High
**Affects**: master feature 全ページ（15ファイル）
**Date Created**: 2026-03-18
**Related**: TASK-017, FE-067

## Summary

マスタ設定ページ内でハードコードされた `text-xs` / `text-sm` を `text-base` に置換する。15ファイルに渡る広範囲の変更。

## 現状のコード

### 単純なテーブルセル系（text-sm → text-base）

以下のファイルは TableCell 内の `text-sm` をハードコードしている：

```typescript
// InsuranceSettings.tsx:88-90
<TableCell className={`font-medium text-sm ${C.text}`}>{item.name}</TableCell>
<TableCell className={`text-sm text-center ${C.text}`}>{item.coverageRate > 0 ? `${item.coverageRate}%` : "-"}</TableCell>
<TableCell className={`text-sm ${C.text70}`}>{item.contactPhone || "-"}</TableCell>

// StaffSettings.tsx:148-149
<TableCell className={`font-medium text-sm ${C.text}`}>{item.name}</TableCell>
<TableCell className={`text-sm ${C.text}`}>{STAFF_ROLE_LABELS[item.staffRole]}</TableCell>

// JobTitleSettings.tsx:103-104
// AnimalSpeciesSettings.tsx:80
// ChiefComplaintSettings.tsx:76-77
// HospitalizationSettings.tsx:94-97
// ServiceTypeSettings.tsx:102, 108
// InterviewTemplateSettings.tsx:87-88
// DiagnosisSettings.tsx:300, 303, 391, 394
```

### TrimmingSettings.tsx（複雑）

```typescript
// :98 text-xs バッジ
<span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-sm text-xs bg-[#DDEDEA] text-[#0F7B6C]">
// :104 text-xs バッジ
<span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-sm text-xs bg-[#E3E2E0] text-[#37352F]/60">
// :199, 383 text-sm 通貨記号
<span className={`text-sm ${C.text65} select-none`}>¥</span>
// :203, 387 text-sm 入力
className={`w-32 bg-transparent text-sm ${C.text} outline-none ...`}
// :277-286, 461-470 text-sm テーブルセル
// :636 text-sm タブ
className={`h-9 border-b-2 border-b-transparent px-4 text-sm ${C.text60} ...`}
```

### MedicineSettings.tsx（複雑）

```typescript
// :104 定数
const SELECT_TRIGGER_FULL = `h-[30px] text-sm bg-transparent ${C.text} ...`;
// :157 text-sm
<TableCell className={`${STYLE.tableCell} w-[100px] text-center text-sm`}>
// :213, 216, 219, 285, 307, 310 text-sm 各種
// :349, 761, 764 text-xs
// :860 text-sm
```

### CageSettings.tsx

```typescript
// :81-84 text-sm（リスト表示行）
<div className={`flex-1 min-w-0 text-sm font-medium ${C.text} px-3`}>{cage.name}</div>
<div className="w-[100px] shrink-0 text-sm text-[#37352F]/65">{CAGE_TYPE_LABELS[cage.cageType]}</div>
// :161-164 text-sm テーブルセル
// :178 text-sm 追加ボタン
```

### CompanySettings.tsx

```typescript
// :21 定数
const PROP_INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} ...`;
// :143 text-sm ボタン
// :152 text-sm ローディング
// :182-217 text-sm プロパティ値（8箇所）
// :224 text-sm
// :238 text-xs
```

### MasterSettingsIndex.tsx

```typescript
// :118 text-sm ラベル
<div className={`text-sm font-medium ${C.text} leading-tight`}>{label}</div>
// :119 text-xs 説明
<div className={`text-xs ${C.text45} mt-0.5 truncate`}>{description}</div>
// :122 text-xs カウント
<span className={`text-xs ${C.text40} tabular-nums shrink-0`}>
// :199 text-sm
// :209 text-xs
```

### TreatmentPlanMaster.tsx

```typescript
// :226, 230, 401, 408 text-sm
// :403 text-xs（子項目カウント）
// :745 text-sm タブ
```

### DiagnosisSettings.tsx

```typescript
// :300, 303, 391, 394 text-sm テーブルセル
// :557 text-sm タブ
```

## 必要な変更

### 一括ルール

| 変更前 | 変更後 |
|--------|--------|
| `text-xs` | `text-base` |
| `text-sm` | `text-base` |

### ファイル別変更箇所数

| ファイル | text-sm | text-xs | 合計 |
|---------|---------|---------|------|
| InsuranceSettings.tsx | 3 | 0 | 3 |
| StaffSettings.tsx | 2 | 0 | 2 |
| JobTitleSettings.tsx | 2 | 0 | 2 |
| AnimalSpeciesSettings.tsx | 1 | 0 | 1 |
| ChiefComplaintSettings.tsx | 2 | 0 | 2 |
| HospitalizationSettings.tsx | 4 | 0 | 4 |
| ServiceTypeSettings.tsx | 2 | 0 | 2 |
| InterviewTemplateSettings.tsx | 2 | 0 | 2 |
| DiagnosisSettings.tsx | 4 | 0 | 4 + タブ1 |
| TrimmingSettings.tsx | 10 | 2 | 12 + タブ1 |
| MedicineSettings.tsx | 8 | 3 | 11 |
| CageSettings.tsx | 6 | 0 | 6 |
| CompanySettings.tsx | 11 | 1 | 12 |
| MasterSettingsIndex.tsx | 2 | 3 | 5 |
| TreatmentPlanMaster.tsx | 4 | 1 | 5 + タブ1 |
| **合計** | **63** | **10** | **~74** |

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] 型は `models.ts` から導出

## 依存関係

- FE-067 が先に完了していること（STYLE トークン経由の変更が先に反映される前提）

## 完了条件

- [ ] 上記 15 ファイル内に `text-xs` / `text-sm`（テキストサイズ）が残っていない
- [ ] `npm run build` パス
- [ ] `npm run lint` パス
- [ ] 全マスタ設定ページのレイアウトが崩れていない
- [ ] TrimmingSettings のバッジ（text-xs → text-base）の見た目が適切
