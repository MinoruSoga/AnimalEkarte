# FE-124: テーブルセルのテキストサイズ・パディング統一

**Status**: Open
**Priority**: High
**Affects**: 検査管理一覧、カルテ管理一覧
**Date Created**: 2026-03-26
**Related**: TASK-030, FE-125

## Summary

全一覧ページのテーブルセルは `text-base` + `py-2.5` が標準だが、検査管理は全セルが `text-sm`、カルテ管理の主訴セルも `text-sm` になっており視覚的不統一が発生している。また大多数のページがセル padding を `py-2` でハードコードしており、デザイントークン `STYLE.tableCell`（`py-2.5`）と乖離している。

## 現状のコード

### 検査管理（全セルが `text-sm`）

```typescript
// frontend/src/features/examinations/routes/Examinations.tsx:229-244
<TableCell className="font-mono text-sm text-[#37352F] py-2">{r.date}</TableCell>
<TableCell className="text-sm text-[#37352F] py-2">{r.ownerName}</TableCell>
<TableCell className="text-sm text-[#37352F] py-2">{r.petName}</TableCell>
<TableCell className="text-sm font-medium text-[#37352F] py-2">{r.testType}</TableCell>
<TableCell className="text-sm text-muted-foreground truncate max-w-[200px] py-2">
  {r.resultSummary || "-"}
</TableCell>
<TableCell className="text-sm text-[#37352F] py-2">{r.doctor}</TableCell>
```

### カルテ管理（主訴のみ `text-sm`）

```typescript
// frontend/src/features/medical-records/routes/MedicalRecords.tsx:212
<TableCell className={`text-sm ${C.text} max-w-[200px] truncate py-2`} title={r.chiefComplaint}>
  {r.chiefComplaint}
</TableCell>
```

### デザイントークン定義

```typescript
// frontend/src/lib/design-tokens.ts
tableCell:     `text-base ${C.text} py-2.5`,   // 正規定義
tableCellMono: `font-mono text-base ${C.text} py-2.5`,
```

### 他ページの標準（参考）

```typescript
// VaccinationList.tsx, TrimmingList.tsx, Accounting.tsx, CheckupsList.tsx など
<TableCell className="text-base text-[#37352F] py-2">{r.ownerName}</TableCell>
// ↑ text-base は正しいが py-2 はデザイントークンより -0.5（py-2.5 が正規）
```

## 必要な変更

### 1. Examinations.tsx — `text-sm` を `text-base` に、`py-2` を `py-2.5` に変更

```typescript
// Before
<TableCell className="font-mono text-sm text-[#37352F] py-2">{r.date}</TableCell>
<TableCell className="text-sm text-[#37352F] py-2">{r.ownerName}</TableCell>
<TableCell className="text-sm text-[#37352F] py-2">{r.petName}</TableCell>
<TableCell className="text-sm font-medium text-[#37352F] py-2">{r.testType}</TableCell>
<TableCell className="text-sm text-muted-foreground truncate max-w-[200px] py-2">
<TableCell className="text-sm text-[#37352F] py-2">{r.doctor}</TableCell>

// After
<TableCell className="font-mono text-base text-[#37352F] py-2.5">{r.date}</TableCell>
<TableCell className="text-base text-[#37352F] py-2.5">{r.ownerName}</TableCell>
<TableCell className="text-base text-[#37352F] py-2.5">{r.petName}</TableCell>
<TableCell className="text-base font-medium text-[#37352F] py-2.5">{r.testType}</TableCell>
<TableCell className="text-base text-muted-foreground truncate max-w-[200px] py-2.5">
<TableCell className="text-base text-[#37352F] py-2.5">{r.doctor}</TableCell>
```

StatusBadge セル（ステータス・操作）の `py-2` も `py-2.5` に：
```typescript
// Before
<TableCell className="py-2">  // ステータスセル
<TableCell className="text-right py-2">  // 操作セル

// After
<TableCell className="py-2.5">
<TableCell className="text-right py-2.5">
```

### 2. MedicalRecords.tsx — 主訴セルの `text-sm` を `text-base` に

```typescript
// Before
<TableCell className={`text-sm ${C.text} max-w-[200px] truncate py-2`} title={r.chiefComplaint}>

// After
<TableCell className={`text-base ${C.text} max-w-[200px] truncate py-2.5`} title={r.chiefComplaint}>
```

その他の `py-2` セルも `py-2.5` に統一（`STYLE.tableCell` 使用セルはすでに `py-2.5` のため変更不要）：
```typescript
// py-2 → py-2.5 対象（STYLE.tableCell/tableCellMono 以外のセル）
// 関連セル（line 215）、ステータスセル（line 241）、操作セル（line 246）
```

## プロジェクトルール遵守チェック

- [x] `any` 型なし
- [x] `FC` / `forwardRef` なし
- [x] barrel index 経由 import なし
- [x] 条件レンダー `? ... : null`
- [x] 型は `models.ts` から導出

## 完了条件

- [ ] Examinations.tsx の全 TableCell が `text-base` を使用している
- [ ] Examinations.tsx の全 TableCell が `py-2.5` を使用している
- [ ] MedicalRecords.tsx の主訴セルが `text-base` になっている
- [ ] MedicalRecords.tsx の `py-2` ハードコードセルが `py-2.5` になっている
- [ ] `pnpm lint` パス（ESLint エラーなし）
- [ ] ビルド成功（`pnpm build`）
