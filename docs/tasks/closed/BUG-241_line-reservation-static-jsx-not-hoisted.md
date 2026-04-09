---
name: BUG-241_line-reservation-static-jsx-not-hoisted
description: line-reservation 4ファイルで static SelectItem JSX がコンポーネント内で毎レンダー生成
type: bug
---

# BUG-241: line-reservation feature の static SelectItem JSX が未巻き上げ（4ファイル・6箇所）

## 概要

`frontend/src/features/line-reservation/` の 4 ファイルで、静的な `<SelectItem>` JSX が
コンポーネント内（または `Object.entries().map()` 呼び出し）で毎レンダー生成されている。
`rendering-hoist-jsx` 違反。コンポーネントが再レンダーされるたびに不要な JSX オブジェクト生成が発生する。

## 現状コード

### `LineReservationSettings.tsx:170-173, 191-194`
```tsx
// SettingsForm コンポーネント内（毎レンダーで生成）
<SelectContent>
  <SelectItem value="minimize_gaps">空き時間を最小化</SelectItem>
  <SelectItem value="allow_gaps">空き時間を許容</SelectItem>
</SelectContent>
// ...
<SelectContent>
  <SelectItem value="first_available">最初の空き</SelectItem>
  <SelectItem value="top_priority">優先度最上位</SelectItem>
</SelectContent>
```

### `LineReservationCourses.tsx:176-181`
```tsx
// CourseFormDialog コンポーネント内（毎レンダーで .map() 実行）
<SelectContent>
  {Object.entries(DAY_OPTION_LABELS).map(([value, label]) => (
    <SelectItem key={value} value={value}>
      {label}
    </SelectItem>
  ))}
</SelectContent>
```

### `LineReservationStaffs.tsx:132-135`
```tsx
// StaffFormDialog コンポーネント内（毎レンダーで .map() 実行）
<SelectContent>
  {Object.entries(STAFF_TYPE_LABELS).map(([value, label]) => (
    <SelectItem key={value} value={value}>
      {label}
    </SelectItem>
  ))}
</SelectContent>
```

### `LineStaffSchedule.tsx:124-128`
```tsx
// ScheduleEditDialog コンポーネント内（毎レンダーで .map() 実行）
<SelectContent>
  {Object.entries(SHIFT_COLORS).map(([value, config]) => (
    <SelectItem key={value} value={value}>
      {config.label}
    </SelectItem>
  ))}
</SelectContent>
```

## 期待する動作

静的な JSX はモジュールスコープの定数として巻き上げ、コンポーネント外で一度だけ生成する。

## 修正方針

各ファイルでモジュール定数を追加し、コンポーネント内の JSX 生成を置き換える。

### `LineReservationSettings.tsx`
```tsx
// モジュールトップに追加
const TIME_SLOT_MODE_ITEMS = (
  <>
    <SelectItem value="minimize_gaps">空き時間を最小化</SelectItem>
    <SelectItem value="allow_gaps">空き時間を許容</SelectItem>
  </>
);

const NO_STAFF_MODE_ITEMS = (
  <>
    <SelectItem value="first_available">最初の空き</SelectItem>
    <SelectItem value="top_priority">優先度最上位</SelectItem>
  </>
);

// SettingsForm 内で使用
<SelectContent>{TIME_SLOT_MODE_ITEMS}</SelectContent>
// ...
<SelectContent>{NO_STAFF_MODE_ITEMS}</SelectContent>
```

### `LineReservationCourses.tsx`
```tsx
// モジュールトップに追加（DAY_OPTION_LABELS 定義の直後）
const DAY_OPTION_SELECT_ITEMS = Object.entries(DAY_OPTION_LABELS).map(([value, label]) => (
  <SelectItem key={value} value={value}>
    {label}
  </SelectItem>
));

// CourseFormDialog 内で使用
<SelectContent>{DAY_OPTION_SELECT_ITEMS}</SelectContent>
```

### `LineReservationStaffs.tsx`
```tsx
// モジュールトップに追加（STAFF_TYPE_LABELS 定義の直後）
const STAFF_TYPE_SELECT_ITEMS = Object.entries(STAFF_TYPE_LABELS).map(([value, label]) => (
  <SelectItem key={value} value={value}>
    {label}
  </SelectItem>
));

// StaffFormDialog 内で使用
<SelectContent>{STAFF_TYPE_SELECT_ITEMS}</SelectContent>
```

### `LineStaffSchedule.tsx`
```tsx
// モジュールトップに追加（SHIFT_COLORS 定義の直後）
const SHIFT_TYPE_SELECT_ITEMS = Object.entries(SHIFT_COLORS).map(([value, config]) => (
  <SelectItem key={value} value={value}>
    {config.label}
  </SelectItem>
));

// ScheduleEditDialog 内で使用
<SelectContent>{SHIFT_TYPE_SELECT_ITEMS}</SelectContent>
```

## 参照実装

`frontend/src/features/estimates/routes/EstimateForm.tsx:37-42`
```tsx
// rendering-hoist-jsx: ステータス SelectItem リストは静的なのでモジュール定数に巻き上げ
const STATUS_SELECT_ITEMS = STATUS_OPTIONS.map(opt => (
  <SelectItem key={opt.value} value={opt.value}>
    {opt.label}
  </SelectItem>
));
```

`frontend/src/features/owners/components/` でも同パターンが正しく実装済み。

## 影響範囲

| ファイル | 箇所 | SelectItem 数 |
|---------|------|--------------|
| `features/line-reservation/routes/LineReservationSettings.tsx:170-173` | timeSlotMode Select | 2 |
| `features/line-reservation/routes/LineReservationSettings.tsx:191-194` | noStaffMode Select | 2 |
| `features/line-reservation/routes/LineReservationCourses.tsx:176-181` | dayOption Select | 4 |
| `features/line-reservation/routes/LineReservationStaffs.tsx:132-135` | staffType Select | 3 |
| `features/line-reservation/routes/LineStaffSchedule.tsx:124-128` | shiftType Select | 5 |

合計: 5箇所 / 4ファイル / 16 SelectItems

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — `rendering-hoist-jsx`
> コンポーネント外の静的 JSX（Select 選択肢など）はモジュール定数に巻き上げ

### `frontend/CLAUDE.md` — パフォーマンスパターン
> `rendering-hoist-jsx`: コンポーネント外の静的 JSX（Select 選択肢など）はモジュール定数に巻き上げ

## 優先度

**Low** — パフォーマンス上の悪影響は軽微（コンポーネントが memo() されていないため memoization 破壊はないが、不要なオブジェクト生成を排除すべき）

## 関連チケット

- BUG-227: 他 feature での同一パターン（4ファイル）— line-reservation は新規追加 feature のため別票
