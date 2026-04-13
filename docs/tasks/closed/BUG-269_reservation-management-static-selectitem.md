# BUG-269: ReservationManagement の CALENDAR_VIEW_SELECT_ITEMS をモジュール定数に巻き上げ未実施

## 概要
`ReservationManagement` コンポーネント内でカレンダー表示切替の `<SelectItem>` JSX を `CALENDAR_VIEW_VALUES.map()` でインライン生成している。`CALENDAR_VIEW_VALUES`（モジュール定数）と `getCalendarViewLabel`（純粋関数）のみに依存する静的な JSX であるため、コンポーネント再レンダー時も同一の出力になる。モジュール定数に巻き上げることで毎レンダーの JSX オブジェクト生成を回避できる。

## 現状コード

### `frontend/src/features/reservations/routes/ReservationManagement.tsx:233-240`
```typescript
<SelectContent>
  {CALENDAR_VIEW_VALUES.map((v) => (
    <SelectItem key={v} value={v}>
      {getCalendarViewLabel(v)}
    </SelectItem>
  ))}
</SelectContent>
```

- `CALENDAR_VIEW_VALUES`: `"../types"` からインポートするモジュール定数
- `getCalendarViewLabel`: `"@/utils/status-helpers"` からインポートする純粋関数
- どちらも props/state に依存しない → JSX 出力は常に同一

## 修正方針

### `frontend/src/features/reservations/routes/ReservationManagement.tsx`

```typescript
// インポート群の後、コンポーネント定義の前に追加
// rendering-hoist-jsx: 静的 SelectItem JSX をモジュール定数に巻き上げ
const CALENDAR_VIEW_SELECT_ITEMS = CALENDAR_VIEW_VALUES.map((v) => (
  <SelectItem key={v} value={v}>
    {getCalendarViewLabel(v)}
  </SelectItem>
));
```

```typescript
// コンポーネント内 SelectContent を修正
<SelectContent>
  {CALENDAR_VIEW_SELECT_ITEMS}
</SelectContent>
```

## 影響範囲

| ファイル | 行番号 | 内容 |
|---------|-------|------|
| `frontend/src/features/reservations/routes/ReservationManagement.tsx` | 233-239 | SelectItem インライン生成 |

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — rendering-hoist-jsx
> コンポーネント外の静的 JSX（Select 選択肢など）はモジュール定数に巻き上げ

### プロジェクト内参照実装
- `features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:32-38` — `SHIFT_TYPE_OPTIONS` を正しくモジュール定数化
- `features/line-reservation/routes/LineReservationStaffs.tsx:43-47` — `STAFF_TYPE_SELECT_ITEMS`（BUG-241 で修正済み）
- `features/line-reservation/routes/LineStaffSchedule.tsx:47-51` — `SHIFT_TYPE_SELECT_ITEMS`（BUG-241 で修正済み）

## 優先度
**Low** — 機能への影響はないが毎レンダーで不要な JSX オブジェクトを生成している。修正コスト 5分。

## 関連チケット
- BUG-227: 静的 SelectItem 未巻き上げ（同一パターン、別ドメイン）
- BUG-241: CLOSED — line-reservation 4ファイルのモジュール定数化

## 関連ファイル
- `frontend/src/features/reservations/routes/ReservationManagement.tsx:233-239` — 違反箇所
- `frontend/src/features/reservations/types/index.ts` — `CALENDAR_VIEW_VALUES` 定義
- `frontend/src/utils/status-helpers.ts` — `getCalendarViewLabel` 定義
