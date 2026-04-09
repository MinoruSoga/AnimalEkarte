# BUG-223: DailyRecordSection の `plans = []` デフォルト値が memo を無効化

## 概要
`DailyRecordSection` コンポーネントで props のデフォルト値として `plans = []` を使用している。関数コンポーネントのデフォルト引数はレンダーごとに新しいリテラルを生成するため、`memo()` のシャロー比較が常に不一致となり、メモ化が完全に無効化される。

## 現状コード

### `features/hospitalization/components/DailyRecordSection.tsx`
```typescript
// ❌ plans = [] はレンダーごとに新しい配列参照 → memo 無効化
export const DailyRecordSection = memo(function DailyRecordSection({
  date,
  plans = [],   // ← ここが問題
  onUpdate,
}: Props) {
  // ...
});
```

## 修正方針

デフォルト値をモジュールレベルの定数に巻き上げる。

### `features/hospitalization/components/DailyRecordSection.tsx`
```typescript
// ✅ モジュールスコープに定数として定義
const EMPTY_PLANS: CarePlan[] = [];

export const DailyRecordSection = memo(function DailyRecordSection({
  date,
  plans = EMPTY_PLANS,   // ← 常に同一参照
  onUpdate,
}: Props) {
  // ...
});
```

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — rerender-memo-with-default-value
> `memo()` に渡すハンドラは `useCallback` で安定化すること。非 primitive デフォルト値は必ずモジュール定数に巻き上げる。

### プロジェクト内参照実装
`frontend/CODING_RULES.md` Section 12 — `rendering-hoist-jsx` パターン（モジュール定数への巻き上げ）

## 優先度
**Low** — memo の有効化のみ。機能的影響なし。修正は5分。

## 関連ファイル
- `frontend/src/features/hospitalization/components/DailyRecordSection.tsx`
