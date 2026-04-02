# FE-025: 日付ピッカーにドロップダウンナビゲーション追加

**Status**: Closed
**Priority**: High
**Affects**: NotionDatePicker（全日付入力フィールド）
**Date Created**: 2026-03-18
**Related**: TASK-019

## Summary

`NotionDatePicker` の Calendar に `captionLayout="dropdown"` を追加し、年・月をドロップダウンで選択可能にする。現在は月送りボタン（Chevron）のみで、過去の日付（特に生年月日）の入力が非常に不便。

## 現状のコード

### calendar.tsx（カレンダー基盤コンポーネント）

```typescript
// frontend/src/components/ui/calendar.tsx:14-69
<DayPicker
  showOutsideDays={showOutsideDays}
  className={cn("p-3", className)}
  classNames={{
    months: "flex flex-col sm:flex-row gap-2",
    month: "flex flex-col gap-4",
    month_caption: "flex justify-center pt-1 relative items-center w-full",
    caption_label: "text-sm font-medium",
    nav: "flex items-center gap-1",
    button_previous: cn(
      buttonVariants({ variant: "outline" }),
      "size-10 bg-transparent p-0 opacity-50 hover:opacity-100 absolute left-1",
    ),
    button_next: cn(
      buttonVariants({ variant: "outline" }),
      "size-10 bg-transparent p-0 opacity-50 hover:opacity-100 absolute right-1",
    ),
    // ... 以下略
  }}
  // captionLayout 未設定 → デフォルト（月送りボタンのみ）
  {...props}
/>
```

### NotionDatePicker.tsx（SinglePicker）

```typescript
// frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx:153-161
<Calendar
  mode="single"
  selected={selected}
  onSelect={handleSelect}
  defaultMonth={selected ?? new Date()}
  locale={ja}
  className="rounded-md"
  classNames={CALENDAR_CLASS_NAMES}
/>
// captionLayout, startMonth, endMonth 未設定
```

### NotionDatePicker.tsx（RangePicker）

```typescript
// frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx:233-243
<Calendar
  mode="range"
  selected={dateRange}
  onSelect={handleSelect}
  numberOfMonths={2}
  defaultMonth={range.from ?? new Date()}
  locale={ja}
  className="rounded-md"
  classNames={CALENDAR_CLASS_NAMES}
/>
// captionLayout, startMonth, endMonth 未設定
```

## 必要な変更

### 1. calendar.tsx — ドロップダウン用 classNames 追加

react-day-picker v9 の UI enum から、ドロップダウン関連の classNames は以下の5つ:

```typescript
// frontend/src/components/ui/calendar.tsx
// classNames オブジェクトに以下を追加:
dropdowns: "flex items-center gap-2",
dropdown_root: "relative",
dropdown: "appearance-none rounded-md border border-input bg-background px-2 py-1.5 text-sm font-medium focus:outline-none focus:ring-1 focus:ring-ring cursor-pointer",
months_dropdown: "",
years_dropdown: "",
```

**UI enum → className マッピング:**
| UI enum key | className key | 適用先 |
|-------------|--------------|--------|
| `Dropdowns` | `dropdowns` | 月・年ドロップダウンのコンテナ |
| `DropdownRoot` | `dropdown_root` | 各ドロップダウンのラッパー |
| `Dropdown` | `dropdown` | `<select>` 要素そのもの |
| `MonthsDropdown` | `months_dropdown` | 月ドロップダウン固有スタイル |
| `YearsDropdown` | `years_dropdown` | 年ドロップダウン固有スタイル |

**注意**: `captionLayout="dropdown"` を使う場合、react-day-picker v9 は `caption_label` を非表示にし、代わりに `dropdowns` コンテナ内に `<select>` 要素（月・年）を表示する。

### 2. NotionDatePicker.tsx — captionLayout + startMonth/endMonth 追加

```typescript
// frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx

// ファイル先頭に定数を追加（モジュールレベル・静的JSX巻き上げルール準拠）:
const CALENDAR_START_MONTH = new Date(new Date().getFullYear() - 100, 0);
const CALENDAR_END_MONTH = new Date(new Date().getFullYear(), 11);

// SinglePicker の Calendar に props 追加:
<Calendar
  mode="single"
  captionLayout="dropdown"
  startMonth={CALENDAR_START_MONTH}
  endMonth={CALENDAR_END_MONTH}
  reverseYears
  selected={selected}
  onSelect={handleSelect}
  defaultMonth={selected ?? new Date()}
  locale={ja}
  className="rounded-md"
  classNames={CALENDAR_CLASS_NAMES}
/>

// RangePicker の Calendar にも同様に props 追加:
<Calendar
  mode="range"
  captionLayout="dropdown"
  startMonth={CALENDAR_START_MONTH}
  endMonth={CALENDAR_END_MONTH}
  reverseYears
  selected={dateRange}
  onSelect={handleSelect}
  numberOfMonths={2}
  defaultMonth={range.from ?? new Date()}
  locale={ja}
  className="rounded-md"
  classNames={CALENDAR_CLASS_NAMES}
/>
```

### 3. calendar.tsx — nav 関連スタイルの調整

`captionLayout="dropdown"` ではナビゲーションボタン（ChevronLeft/Right）が非表示になる。`button_previous` / `button_next` の absolute 配置がドロップダウンと競合しないことを確認。ドロップダウン使用時は nav 自体が表示されないため問題ない。

### 4. NotionDatePicker.tsx — reverseYears prop 追加

年ドロップダウンの表示順を新しい年→古い年の順にするため、`reverseYears` prop を両方の Calendar に追加する。生年月日入力では近年から選択するケースが大半のため。

## UI 操作フロー

1. ユーザーが日付入力フィールドをクリック
2. Popover でカレンダーが開く
3. カレンダー上部に**年ドロップダウン**と**月ドロップダウン**が表示される
4. 年ドロップダウンをクリック → 1926年〜2026年の年リストが表示される
5. 月ドロップダウンをクリック → 1月〜12月のリストが表示される
6. 年・月を選択するとカレンダーが即座にその月に遷移
7. 日付をクリックして確定

## プロジェクトルール遵守チェック

- [x] `any` 型なし
- [x] `FC` / `forwardRef` なし
- [x] barrel index 経由 import なし
- [x] 条件レンダー `? ... : null`（`&&` 禁止）
- [x] 型は `models.ts` から導出（この変更ではモデル型変更なし）
- [x] 静的値はモジュール定数に巻き上げ（`CALENDAR_START_MONTH`, `CALENDAR_END_MONTH`）

## 依存関係

- なし（フロントエンドのみ、Backend変更不要）

## 完了条件

- [ ] `captionLayout="dropdown"` が SinglePicker・RangePicker 両方に適用
- [ ] `reverseYears` が両方に適用（年が新しい順で表示される）
- [ ] 年ドロップダウン: 現在年〜100年前（新しい順）
- [ ] 月ドロップダウン: 1月〜12月
- [ ] ドロップダウン用 classNames 5つ（dropdowns, dropdown, dropdown_root, months_dropdown, years_dropdown）が calendar.tsx に追加
- [ ] 全7箇所の日付ピッカーで動作確認
- [ ] ドロップダウンのスタイルが Notion 風デザインと調和
- [ ] 型エラーなし（`npm run build` パス）
- [ ] ESLint エラーなし（`npm run lint` パス）
- [ ] 既存の日付選択・クリア機能が壊れていない

## クローズ情報

- **Closed At**: 2026-03-18
- **変更ファイル**:
  - `frontend/src/components/ui/calendar.tsx` — ドロップダウン用 classNames 5つ追加（dropdowns, dropdown_root, dropdown, months_dropdown, years_dropdown）
  - `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx` — captionLayout="dropdown" + startMonth/endMonth + reverseYears を SinglePicker・RangePicker 両方に追加。CALENDAR_START_MONTH/CALENDAR_END_MONTH 定数追加
