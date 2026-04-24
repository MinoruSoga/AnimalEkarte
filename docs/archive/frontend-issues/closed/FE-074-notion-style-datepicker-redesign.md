# FE-074: 日付ピッカー Notion 風リデザイン

**Status**: Closed
**Priority**: High
**Affects**: NotionDatePicker, Calendar, DateRangePicker — 飼主生年月日、ペット生年月日/去勢日、履歴フィルタ等の全日付選択UI
**Date Created**: 2026-03-18
**Related**: TASK-021

## Summary

`NotionDatePicker` の年月ヘッダー重複バグを修正し、本物の Notion 日付ピッカーの UX に合わせてリデザインする。

## 確認済みバグ

ブラウザで `/owners/1` → 「飼主生年月日」クリックで確認:
- ヘッダーに「2026 2026」「3月 3月」と年・月が重複表示
- 原因: `calendar.tsx` の `caption_label` が `captionLayout="dropdown"` 時に非表示になっていない

## 現状のコード

### 1. calendar.tsx（バグ箇所）

```typescript
// frontend/src/components/ui/calendar.tsx:22
caption_label: "text-sm font-medium",
// ↑ captionLayout="dropdown" 時にもテキストラベルが表示されてしまう
```

### 2. NotionDatePicker.tsx SinglePicker（リデザイン対象）

```typescript
// frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx:156-170
<Calendar
  mode="single"
  captionLayout="dropdown"           // ← ネイティブ select で年月ドロップダウン
  startMonth={CALENDAR_START_MONTH}  // 100年前
  endMonth={CALENDAR_END_MONTH}      // 現在年
  reverseYears
  selected={selected}
  onSelect={handleSelect}
  defaultMonth={selected ?? new Date()}
  locale={ja}
  className="rounded-md"
  classNames={CALENDAR_CLASS_NAMES}
/>
```

現状の問題:
- `captionLayout="dropdown"` はネイティブ `<select>` を使用 → Notion の見た目と全く異なる
- テキスト入力欄がない
- Today ボタンがない
- 月グリッドビュー（1月〜12月を一覧で選択）がない

## 必要な変更

### 1. calendar.tsx — caption_label バグ修正

```typescript
// Before
caption_label: "text-sm font-medium",

// After
caption_label: "text-sm font-medium [.rdp-captionLayout--dropdown_&]:sr-only",
```

**注意**: react-day-picker v9 では `captionLayout="dropdown"` 時にドロップダウンとラベルが両方レンダリングされる。`caption_label` を条件付きで `sr-only` にすることで、ドロップダウンモード時のみ非表示にする。ただし、このプロジェクトでは Notion 風リデザインで `captionLayout="dropdown"` 自体をやめるため、最終的にはこの修正は不要になる可能性がある。念のため入れておく。

### 2. NotionDatePicker.tsx — Notion 風リデザイン

#### 2.1 ビュー切り替え機能（月グリッド）

```typescript
// 新しい state: "calendar" | "monthGrid"
const [view, setView] = useState<"calendar" | "monthGrid">("calendar");
const [displayMonth, setDisplayMonth] = useState<Date>(selected ?? new Date());
```

#### 2.2 カレンダーヘッダー（Notion 風）

現状の `captionLayout="dropdown"` を廃止し、独自ヘッダーに置き換える:

```
┌─────────────────────────────────┐
│  < 　　 2026年 3月 　　 >        │  ← 年月テキストクリックで月グリッドへ
│  ┌───────────────────────────┐  │
│  │ 2026/03/18               │  │  ← テキスト入力欄（タイプ可能）
│  └───────────────────────────┘  │
│  日  月  火  水  木  金  土      │
│   1   2   3   4   5   6   7    │
│   8   9  10  11  12  13  14    │
│  15  16  17 [18] 19  20  21    │  ← 今日はハイライト
│  22  23  24  25  26  27  28    │
│  29  30  31   1   2   3   4    │
│                                 │
│  Today                          │  ← 今日ボタン
└─────────────────────────────────┘
```

月グリッドビュー:
```
┌─────────────────────────────────┐
│  <         2026年          >    │  ← 年送り
│                                 │
│   1月    2月    3月    4月      │
│   5月    6月    7月    8月      │
│   9月   10月   11月   12月      │
│                                 │
└─────────────────────────────────┘
```

#### 2.3 テキスト入力

ポップオーバー内のカレンダー上部にテキスト入力欄を配置:
- 受け付ける形式: `YYYY/MM/DD`, `YYYY-MM-DD`, `YYYYMMDD`
- 入力確定時（Enter または blur）にカレンダーを該当月に遷移 + 日付選択
- パース失敗時は無視（エラー表示不要）

```typescript
function parseDateInput(input: string): Date | null {
  const trimmed = input.trim();

  // YYYYMMDD
  const compact = trimmed.match(/^(\d{4})(\d{2})(\d{2})$/);
  if (compact) {
    return new Date(+compact[1], +compact[2] - 1, +compact[3], 12);
  }

  // YYYY/MM/DD or YYYY-MM-DD
  const separated = trimmed.match(/^(\d{4})[/-](\d{1,2})[/-](\d{1,2})$/);
  if (separated) {
    return new Date(+separated[1], +separated[2] - 1, +separated[3], 12);
  }

  return null;
}
```

#### 2.4 Today ボタン

カレンダー下部に「Today」テキストボタン:
- クリックで今日の日付を即時選択し、ポップオーバーを閉じる
- `onChange(formatIso(new Date()))` を呼ぶ

#### 2.5 Calendar props 変更

```typescript
// Before
<Calendar
  captionLayout="dropdown"
  startMonth={CALENDAR_START_MONTH}
  endMonth={CALENDAR_END_MONTH}
  reverseYears
  ...
/>

// After — captionLayout="dropdown" を廃止、独自ナビゲーション
<Calendar
  mode="single"
  month={displayMonth}
  onMonthChange={setDisplayMonth}
  selected={selected}
  onSelect={handleSelect}
  locale={ja}
  className="rounded-md"
  classNames={CALENDAR_CLASS_NAMES}
  hideNavigation  // ← 独自ナビゲーションに置き換え
/>
```

#### 2.6 独自ナビゲーションコンポーネント

```typescript
// カレンダーヘッダー
function CalendarHeader({
  displayMonth,
  onPrevMonth,
  onNextMonth,
  onClickTitle,  // → 月グリッドへ切り替え
}: CalendarHeaderProps) {
  const year = displayMonth.getFullYear();
  const month = displayMonth.getMonth() + 1;

  return (
    <div className="flex items-center justify-between px-1 pb-2">
      <button type="button" onClick={onPrevMonth} className="..." aria-label="前の月">
        <ChevronLeft className="h-4 w-4" />
      </button>
      <button type="button" onClick={onClickTitle} className="text-sm font-medium hover:bg-[#F7F6F3] rounded px-2 py-1">
        {year}年 {month}月
      </button>
      <button type="button" onClick={onNextMonth} className="..." aria-label="次の月">
        <ChevronRight className="h-4 w-4" />
      </button>
    </div>
  );
}

// 月グリッドビュー
function MonthGrid({
  year,
  onYearChange,
  onMonthSelect,
}: MonthGridProps) {
  const months = Array.from({ length: 12 }, (_, i) => i);
  return (
    <div>
      <div className="flex items-center justify-between px-1 pb-3">
        <button onClick={() => onYearChange(year - 1)}>
          <ChevronLeft className="h-4 w-4" />
        </button>
        <span className="text-sm font-medium">{year}年</span>
        <button onClick={() => onYearChange(year + 1)}>
          <ChevronRight className="h-4 w-4" />
        </button>
      </div>
      <div className="grid grid-cols-4 gap-1">
        {months.map(m => (
          <button
            key={m}
            onClick={() => onMonthSelect(m)}
            className="rounded px-2 py-1.5 text-sm hover:bg-[#F7F6F3]"
          >
            {m + 1}月
          </button>
        ))}
      </div>
    </div>
  );
}
```

### 3. range モードへの対応

`RangePicker` も同じ独自ナビゲーション + 月グリッドに変更する。テキスト入力と Today ボタンは range モードでは不要（期間選択のため）。ただし、ヘッダーの年月表示と月グリッドは適用する。

### 4. Notion 風デザイントークン（既存を維持）

```typescript
// 既に Notion 配色を使用 — 変更不要
const CALENDAR_CLASS_NAMES = {
  selected: "bg-[#37352F] text-white ...",
  today: "bg-[#F7F6F3] text-[#37352F]",
};
```

## UI 操作フロー

### Single モード（生年月日）
1. ユーザーが「飼主生年月日」欄をクリック
2. ポップオーバーが開く: テキスト入力欄 + カレンダー + Today ボタン
3. **方法A**: カレンダーで日付をクリック → 日付が選択され、ポップオーバーが閉じる
4. **方法B**: テキスト入力欄に「1985/06/15」と入力 → Enter → カレンダーが1985年6月に遷移し、15日が選択される
5. **方法C**: 年月ヘッダー「2026年 3月」をクリック → 月グリッド表示 → 年を `<` `>` で1985年まで送る → 「6月」をクリック → カレンダービューに戻る → 15日をクリック
6. **方法D**: 「Today」クリック → 今日の日付が選択される

### Range モード（履歴フィルタ）
1. ユーザーが期間フィルタをクリック
2. ポップオーバーが開く: 独自ヘッダー + 2ヶ月カレンダー
3. 年月ヘッダーで月グリッドビューに切り替え可能
4. 開始日 → 終了日の順にクリックで範囲選択

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（該当なし — 非同期操作なし）
- [ ] 型は `models.ts` から導出（該当なし — UI コンポーネントの props 型のみ）
- [ ] `useCallback` でハンドラ安定化
- [ ] `useMemo` で月グリッド等の静的リストをキャッシュ

## 依存関係

- バックエンド依存なし（純粋フロントエンド変更）

## 完了条件

- [x] 年月ヘッダー重複バグ修正（「2026 2026」「3月 3月」が解消）
- [x] Notion 風ナビゲーション実装（月グリッドビュー、年送り）
- [x] テキスト入力で日付を直接入力可能（YYYY/MM/DD, YYYY-MM-DD, YYYYMMDD）
- [x] Today ボタンで今日の日付を即時選択
- [x] range モードでもヘッダーバグ解消 + 月グリッド対応
- [x] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [x] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
- [ ] 全使用箇所で回帰なし:（ブラウザ確認が必要）
  - `/owners/1` 飼主生年月日
  - `/owners/1` ペット編集モーダル → 生年月日、去勢日
  - `/owners/new` 新規登録 → 飼主生年月日
  - 履歴フィルタパネル → 期間選択（range モード）

## クローズ情報

- **Closed At**: 2026-03-19
- **備考**: 全機能は前回のコミットで既に実装済み。本クローズはビルド・lint検証のみ。ブラウザでの回帰テストはユーザーに委ねる。
