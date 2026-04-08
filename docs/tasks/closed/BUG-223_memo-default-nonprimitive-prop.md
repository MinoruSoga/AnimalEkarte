# BUG-223: memo() コンポーネントのデフォルト非プリミティブ prop が毎レンダーで新参照を生成

## 概要

`memo()` でラップされた `DailyRecordSection` が `plans = []` というデフォルト値を持つ。
`[]` はレンダーごとに新しい配列インスタンスを生成するため、親から `plans` を渡さない場合でも
`memo()` の浅い比較が毎回 false となり、メモ化が完全に無効化される。

## 現状コード

### `features/hospitalization/components/DailyRecord/DailyRecordSection.tsx:24`

```typescript
export const DailyRecordSection = memo(function DailyRecordSection({
  records,
  plans = [],        // ❌ [] は毎レンダーで新参照 → memo が機能しない
  onAddVital,
  onAddLog,
}: DailyRecordSectionProps) {
```

## 比較: 正しい実装（プロジェクト内参照実装）

```typescript
// features/owners/routes/OwnerForm.tsx — 非プリミティブ default はモジュール定数で宣言
const PET_TABLE_HEADER = (
  <SortableDataTableRow className="cursor-default select-none" clickable={false}>
    ...
  </SortableDataTableRow>
);
// 同様に空配列は:
const EMPTY_PLANS: CarePlan[] = [];

export const DailyRecordSection = memo(function DailyRecordSection({
  records,
  plans = EMPTY_PLANS,  // ✅ モジュール定数 = 常に同じ参照
  onAddVital,
  onAddLog,
}: DailyRecordSectionProps) {
```

## 修正方針

### `features/hospitalization/components/DailyRecord/DailyRecordSection.tsx`

```typescript
// ① モジュール定数を追加（コンポーネント定義の外）
import type { CarePlan } from "../../types";  // 既存の型を使用

const EMPTY_PLANS: CarePlan[] = [];

// ② props のデフォルト値をモジュール定数に変更
export const DailyRecordSection = memo(function DailyRecordSection({
  records,
  plans = EMPTY_PLANS,  // ✅ 常に同じ参照
  onAddVital,
  onAddLog,
}: DailyRecordSectionProps) {
```

## 影響範囲

| ファイル | 行 | 内容 |
|---------|-----|------|
| `features/hospitalization/components/DailyRecord/DailyRecordSection.tsx` | 24 | `plans = []` → `EMPTY_PLANS` |

## 準拠すべきプロジェクト規約・ベストプラクティス

### Vercel React Best Practices — `rerender-memo-with-default-value`
> Hoist default non-primitive props to module-level constants.
> Default props like `[] {}` create new references each render, defeating `memo()`.

### プロジェクト内参照実装
`features/owners/routes/OwnerForm.tsx` — `PET_TABLE_HEADER` をモジュール定数として宣言

## 優先度

**Low** — 修正コストが非常に小さい（2行の変更）。`memo()` の有効性向上。

## 関連チケット

なし
