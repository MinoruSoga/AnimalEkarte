# BUG-283: deprecated `use-master-items` hook — 6ファイルで本番使用中

## 概要
`hooks/use-master-items.ts` は `@deprecated` マーカーが付与されているにもかかわらず、6つの本番ファイルから依然としてimportされている。このラッパーは `@/features/master` への移行を促すためのshimだが、移行が完了していない。

## 再現手順
1. `cat frontend/src/hooks/use-master-items.ts` → `@deprecated` コメントを確認
2. `grep -r "use-master-items" frontend/src --include="*.tsx" --include="*.ts" -l` → 7ファイル（うち1件はmaster/index.ts自身の再エクスポート）

**deprecated wrapper の内容:**
```typescript
// frontend/src/hooks/use-master-items.ts
/**
 * @deprecated Import directly from @/features/master
 * This re-export exists for backward compatibility.
 */
export { useMasterItems } from "@/features/master";
```

## 期待する動作
```typescript
// 正しい import
import { useMasterItems } from "@/features/master";
```

## 現状コード

### 6ファイルの違反 import
```typescript
// ❌ deprecated wrapper 経由
import { useMasterItems } from "@/hooks/use-master-items";
```

### 比較: 正しい実装
```typescript
// ✅ features/master 直接
import { useMasterItems } from "@/features/master";
```

## 影響範囲

| 対象ファイル | 詳細 | 状態 |
|------------|------|------|
| `features/trimming/routes/TrimmingForm.tsx` | `@/hooks/use-master-items` import | 要修正 |
| `features/examinations/routes/ExaminationForm.tsx` | `@/hooks/use-master-items` import | 要修正 |
| `features/hospitalization/hooks/use-hospitalization-list.ts` | `@/hooks/use-master-items` import | 要修正 |
| `features/hospitalization/routes/HospitalizationForm.tsx` | `@/hooks/use-master-items` import | 要修正 |
| `components/shared/ReservationFormModal/ReservationFormFields.tsx` | `@/hooks/use-master-items` import | 要修正 |
| `hooks/use-staff-validation.ts` | `@/hooks/use-master-items` import | 要修正 |

## 修正方針

### 1. 全6ファイルの import を変更

**修正前:**
```typescript
import { useMasterItems } from "@/hooks/use-master-items";
```

**修正後:**
```typescript
import { useMasterItems } from "@/features/master";
```

対象ファイル（全6件）:
- `frontend/src/features/trimming/routes/TrimmingForm.tsx`
- `frontend/src/features/examinations/routes/ExaminationForm.tsx`
- `frontend/src/features/hospitalization/hooks/use-hospitalization-list.ts`
- `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx`
- `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx`
- `frontend/src/hooks/use-staff-validation.ts`

### 2. deprecated wrapper 削除

全6ファイルの修正が完了したら:
```bash
rm frontend/src/hooks/use-master-items.ts
```

削除前に参照ゼロを確認:
```bash
grep -r "use-master-items" frontend/src --include="*.tsx" --include="*.ts"
# 0件 → 削除安全
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Feature Indexing
> **Deep imports from features**: 外部からの利用は Feature の `index.ts` を経由すること。

`@/hooks/use-master-items` は `@/features/master` の機能を横断的に使う迂回路であり、Feature Indexing の趣旨に反する。

### プロジェクト内参照実装
- `features/owners/` — `useMasterItems` を `@/features/master` から直接importする正しいパターン

## 優先度
**Medium** — 動作への影響はないが、deprecated コードが6ファイルで現役使用中という矛盾した状態が継続している。deprecated marker の意味がなくなる。

## 関連チケット
- BUG-282: SidePeek barrel index.ts 欠落

## 関連ファイル
- `frontend/src/hooks/use-master-items.ts:1-5` — deprecated wrapper（移行完了後に削除）
- `frontend/src/features/master/index.ts` — 移行先（`useMasterItems` を re-export済み）
