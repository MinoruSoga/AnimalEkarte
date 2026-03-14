# 014: 入院日次記録（ケアログ・バイタル・スタッフメモ）UI実装

## 概要
入院詳細画面に日次記録タブを追加する。日付ごとのバイタル・ケアログ・スタッフメモを記録・閲覧できる。日付ナビゲーションで日付を切り替え、その日の記録を時系列（`time` 昇順）で表示する。

## 優先度
high

## 関連APIエンドポイント
- GET `/v1/hospitalizations/{id}/daily-records`
- POST `/v1/hospitalizations/{id}/daily-records`
- GET `/v1/hospitalizations/{id}/daily-records/{date}`
- POST `/v1/hospitalizations/{id}/daily-records/{date}/vitals`
- POST `/v1/hospitalizations/{id}/daily-records/{date}/care-logs`
- POST `/v1/hospitalizations/{id}/daily-records/{date}/staff-notes`

## 関連バックエンドチケット
backend/issues/open/009_hospitalization_daily_records.md（実装済み）

## 実装内容

### API層 (`features/hospitalization/api/`)
`daily-records.ts` を新規作成し以下の TanStack Query hooks を実装する。
- `useDailyRecords(hospitalizationId: string)` — GET 日次記録サマリ一覧
- `useCreateDailyRecord(hospitalizationId: string)` — POST 日付レコード作成
- `useDailyRecord(hospitalizationId: string, date: string)` — GET 特定日の詳細
- `useCreateDailyVital(hospitalizationId: string, date: string)` — POST バイタル追加
- `useCreateCareLog(hospitalizationId: string, date: string)` — POST ケアログ追加
- `useCreateStaffNote(hospitalizationId: string, date: string)` — POST スタッフメモ追加

### コンポーネント (`features/hospitalization/components/`)
`DailyRecordsTab/` ディレクトリを新規作成する。
- `DailyRecordsTab.tsx` — 日付ナビゲーション + 選択日の記録表示
- `DailyDateNav.tsx` — 入院期間内の日付リスト（前日/翌日 ボタンまたはカレンダー選択）
- `DailyVitalsSection.tsx` — バイタル表示・追加フォーム
- `DailyCareLogsSection.tsx` — ケアログ一覧（`time` 昇順）+ 追加フォーム
- `DailyStaffNotesSection.tsx` — スタッフメモ一覧 + 追加フォーム
- `DailyRecordsTab/index.ts` — named export

`care_log_type` のアイコン表示:
- `food` — 食事アイコン
- `excretion` — 排泄アイコン
- `medicine` — 薬アイコン
- `treatment` — 処置アイコン
- `other` — その他アイコン

### ページ/ルート (`features/hospitalization/routes/`)
入院詳細ルートコンポーネント（`HospitalizationDetail.tsx` 相当）のタブ一覧に「日次記録」タブを追加し、`DailyRecordsTab` を組み込む。

### 型定義 (`features/hospitalization/types/`)
`index.ts` に以下を追加する。
```typescript
export type CareLogType = 'food' | 'excretion' | 'medicine' | 'treatment' | 'other';

export interface DailyRecord {
  id: string;
  hospitalization_id: string;
  date: string;  // YYYY-MM-DD
  created_at: string;
  updated_at: string;
}

export interface DailyVital {
  id: string;
  daily_record_id: string;
  time: string;           // HH:MM
  temperature?: number | null;
  heart_rate?: number | null;
  respiratory_rate?: number | null;
  body_weight?: number | null;
  note?: string | null;
}

export interface CareLog {
  id: string;
  daily_record_id: string;
  time: string;
  care_log_type: CareLogType;
  content: string;
  staff_id?: string | null;
}

export interface StaffNote {
  id: string;
  daily_record_id: string;
  time: string;
  note: string;
  staff_id?: string | null;
}
```

## 完了条件
- [ ] 日付ナビゲーションで入院期間内の日付を切り替えられる
- [ ] 選択日のバイタル・ケアログ・スタッフメモが `time` 昇順で表示される
- [ ] バイタル・ケアログ・スタッフメモをそれぞれ追加できる
- [ ] `care_log_type` ごとにアイコンが表示される
- [ ] 選択日に記録がない場合は自動で日次レコードを作成してから追加する
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend npm run build` が通る

## 備考
- バックエンドチケット 009 は実装済みのためフロントエンド単独で進められる
- 日付ナビゲーションは入院の `admission_date`〜`discharge_date`（未退院の場合は当日）の範囲に限定する
