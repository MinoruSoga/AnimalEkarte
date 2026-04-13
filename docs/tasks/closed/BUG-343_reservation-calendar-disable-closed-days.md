# BUG-343: 予約フォームのカレンダーで定休日・休業日を選択不可にする

## 概要
予約登録フォームの日付選択カレンダーに、診療所の定休日・臨時休業日（`clinic_holidays` テーブル）の disabled 制御が実装されていない。ユーザーが休業日を選択して予約登録しようとしても、フロントエンド側で弾かれず、後続の時間枠チェックで初めてエラーになる（または登録できてしまう）UX 上の問題がある。

## 再現手順
1. `admin@example.com` / `password` でログイン
2. シフト管理 → 定休日として任意の日付を登録する（例: 翌月の特定日）
3. 予約カレンダー → 「予約登録」ボタンをクリック
4. 日付選択カレンダーを開き、手順2で登録した定休日を選択する
5. **結果**: 定休日でも選択・入力が可能。disabled になっていない。

## 期待する動作
- `clinic_holidays` に登録された日付はカレンダー上で disabled（選択不可・グレーアウト）表示になる
- 過去日付は既存の制御どおり選択不可
- disabled な日付にホバーした場合、ツールチップで理由（例: 「定休日」）を表示する（任意）

## 現状コード

### `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:96-114`
```tsx
// Calendar コンポーネント: disabled prop が設定されていない
<Popover>
  <PopoverTrigger asChild>
    <Button variant="outline">
      {formData.start ? format(formData.start, "yyyy/MM/dd") : "日付を選択"}
    </Button>
  </PopoverTrigger>
  <PopoverContent>
    <Calendar
      mode="single"
      selected={selectedDate}
      onSelect={handleDateSelect}
      // disabled={(date) => isPast(date)} のような制御なし
    />
  </PopoverContent>
</Popover>
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// frontend/src/features/shifts/components/ShiftCalendar/ShiftCalendar.tsx:94-95
// ShiftCalendar では holidaySet を使って disabled 表示を実現している
const holidaySet = useMemo(
  () => new Set(holidays.map((h) => h.date)),
  [holidays]
);
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:96-114` | Calendar の disabled prop 追加 | 未修正 |
| `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx` | `useGetClinicHolidays` フック呼び出し追加 + props 追加 | 未修正 |
| `frontend/src/features/shifts/api/clinic-holidays.ts:20-24` | 既存 API フック `useGetClinicHolidays(yearMonth)` — 再利用可能 | 実装済み |
| `backend/internal/service/clinic_holiday_service.go` | `List(ctx, clinicID, yearMonth)` — 既存実装 | 実装済み |

## 修正方針

### 1. ReservationFormModal で年月変動に応じて定休日を取得
`frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx`
```tsx
import { useGetClinicHolidays } from '@/features/shifts';

// カレンダーで表示中の年月を state で管理
const [calendarMonth, setCalendarMonth] = useState(() => format(new Date(), "yyyy-MM"));
const { data: clinicHolidays = [] } = useGetClinicHolidays(calendarMonth);

const holidayDates = useMemo(
  () => new Set(clinicHolidays.map((h) => h.date)),
  [clinicHolidays]
);
```

### 2. Calendar に disabled 関数を渡す
`frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:96-114`
```tsx
<Calendar
  mode="single"
  selected={selectedDate}
  onSelect={handleDateSelect}
  onMonthChange={(month) => onMonthChange(format(month, "yyyy-MM"))}
  disabled={(date) => {
    if (isBefore(date, startOfDay(new Date()))) return true; // 過去日
    return holidayDates.has(format(date, "yyyy-MM-dd"));     // 定休日
  }}
/>
```

### 3. props に holidayDates と onMonthChange を追加
`ReservationFormFieldsProps` に `holidayDates: Set<string>` と `onMonthChange: (yearMonth: string) => void` を追加する。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — `js-cache-function-results`
> API 由来の JSX リスト生成は `useMemo([list])` でキャッシュ

`holidayDates` の Set 生成は `useMemo` でキャッシュする。

### `.claude/rules/code-style.md` — `rerender-memo`
> 独立した大きいセクションは `memo()` で囲む。必ず props ハンドラを `useCallback` で安定化すること。

`onMonthChange` は `useCallback` で安定化する。

### `.claude/rules/accessibility-rules.md` — キーボード操作対応
> `Tab` フォーカス順序を管理

disabled な日付が Tab フォーカスを受け取らないように shadcn/ui Calendar の仕様に従って実装する。

### プロジェクト内参照実装
- `frontend/src/features/shifts/components/ShiftCalendar/ShiftCalendar.tsx:94-95` — `holidaySet` による disabled パターン

## 優先度
**High** — 休業日に予約が入る可能性があり、運用上の問題が発生しうる。定休日 API は既に実装されており、フロント側の接続コストは低い。

## 関連チケット
- BUG-344: 担当者選択の出勤フィルタリング（同一フォームの関連改善）

## 関連ファイル
- `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:96-114` — Calendar コンポーネント
- `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx` — フォームルート（holiday データ取得を追加する場所）
- `frontend/src/features/shifts/api/clinic-holidays.ts:20-24` — `useGetClinicHolidays` フック
- `backend/internal/service/clinic_holiday_service.go` — 定休日 Service（参照のみ）
