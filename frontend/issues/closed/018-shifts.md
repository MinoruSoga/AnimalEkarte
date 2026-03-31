# 018: スタッフシフト管理画面実装

## 概要
スタッフのシフト（勤務区分）を月単位カレンダーで管理する画面を実装する。シフト種別ごとに色分け表示し、スタッフフィルタで絞り込みができる。

## 優先度
low

## 関連APIエンドポイント
- GET `/v1/shifts?date=YYYY-MM&staff_id=xxx`
- POST `/v1/shifts`
- PATCH `/v1/shifts/{id}`
- DELETE `/v1/shifts/{id}`

## 関連バックエンドチケット
なし（バックエンド実装状況を確認してから着手すること）

## 実装内容

### API層 (`features/shifts/api/`)
以下のファイルを新規作成する。
- `get-shifts.ts` — `useShifts(params: { date: string; staff_id?: string })` hook
- `create-shift.ts` — `useCreateShift()` mutation hook
- `update-shift.ts` — `useUpdateShift()` mutation hook
- `delete-shift.ts` — `useDeleteShift()` mutation hook
- `types.ts` — API リクエスト/レスポンス型
- `transforms.ts` — バックエンド ↔ フロントエンド変換

`date` パラメータは `YYYY-MM` 形式で月を指定する。

### コンポーネント (`features/shifts/components/`)
- `ShiftCalendar/` — 月単位カレンダービュー（スタッフ × 日付のグリッド）
- `ShiftCell/` — 各セルのシフト表示（色分けバッジ）
- `ShiftFormDialog/` — シフト追加・編集ダイアログ

`shift_type` の色分け表示:
- `full_day` — 「全日」（blue）
- `morning` — 「午前」（green）
- `afternoon` — 「午後」（teal）
- `night` — 「夜勤」（purple）
- `day_off` — 「休日」（gray）
- `holiday` — 「祝日」（red）

### ページ/ルート (`features/shifts/routes/`)
- `ShiftCalendar.tsx` — シフトカレンダーページ（月ナビゲーション + スタッフフィルタ + カレンダービュー）

### 型定義 (`features/shifts/types/`)
`index.ts` を新規作成し以下を定義する。
```typescript
export type ShiftType =
  | 'full_day'
  | 'morning'
  | 'afternoon'
  | 'night'
  | 'day_off'
  | 'holiday';

export interface Shift {
  id: string;
  staff_id: string;
  staff_name?: string;
  date: string;        // YYYY-MM-DD
  shift_type: ShiftType;
  start_time?: string | null;  // HH:MM
  end_time?: string | null;    // HH:MM
  note?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateShiftInput {
  staff_id: string;
  date: string;
  shift_type: ShiftType;
  start_time?: string | null;
  end_time?: string | null;
  note?: string | null;
}
```

## 完了条件
- [ ] 月単位カレンダーでスタッフ別シフトが表示される
- [ ] 月ナビゲーション（前月/翌月）が動作する
- [ ] スタッフフィルタ（`staff_id` クエリパラメータ）が動作する
- [ ] シフトを追加・編集・削除できる
- [ ] `shift_type` ごとに色分けバッジが表示される
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend npm run build` が通る

## 備考
- バックエンド API の実装状況を事前に `backend/docs/api.yaml` で確認すること
- `features/shifts/` は新規 feature ディレクトリとして作成する
- ルーターへの追加は `app/router.tsx` で行うこと
- カレンダーの描画は外部ライブラリ（`react-big-calendar` 等）を使わず、CSS Grid で自前実装することを推奨する（依存追加が必要な場合は事前に確認すること）
