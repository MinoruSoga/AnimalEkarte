# 009: カルテ詳細バイタルタブ実装

## 概要
外来カルテ詳細画面にバイタル記録タブを追加する。体温・心拍数・呼吸数・体重をテーブルで表示し、新規追加・編集・削除ができる。

## 優先度
high

## 関連APIエンドポイント
- GET `/v1/medical-records/{id}/vitals`
- POST `/v1/medical-records/{id}/vitals`
- PATCH `/v1/medical-records/{id}/vitals/{vitalId}`
- DELETE `/v1/medical-records/{id}/vitals/{vitalId}`

## 関連バックエンドチケット
backend/issues/open/004_medical_record_vitals.md（実装済み）

## 実装内容

### API層 (`features/medical-records/api/`)
`vitals.ts` を新規作成し以下の TanStack Query hooks を実装する。
- `useVitals(medicalRecordId: string)` — GET 一覧取得（`recorded_at` 昇順）
- `useCreateVital(medicalRecordId: string)` — POST 新規追加
- `useUpdateVital(medicalRecordId: string)` — PATCH 編集
- `useDeleteVital(medicalRecordId: string)` — DELETE 削除

### コンポーネント (`features/medical-records/components/`)
`VitalsTab/` ディレクトリを新規作成する。
- `VitalsTab.tsx` — バイタル一覧テーブル + インライン追加フォーム
- `VitalsTab/index.ts` — named export

テーブルカラム: `recorded_at`（記録日時）・`temperature`（体温 ℃）・`heart_rate`（心拍数 bpm）・`respiratory_rate`（呼吸数 /min）・`body_weight`（体重 kg）・操作列
全フィールドは nullable（任意入力）。

### ページ/ルート (`features/medical-records/routes/`)
`MedicalRecordDetail.tsx`（または詳細画面相当のルートコンポーネント）のタブ一覧に「バイタル」タブを追加し、`VitalsTab` を組み込む。

### 型定義 (`features/medical-records/types/`)
`index.ts` に以下を追加する。
```typescript
export interface Vital {
  id: string;
  medical_record_id: string;
  recorded_at: string;        // ISO8601
  temperature?: number | null;
  heart_rate?: number | null;
  respiratory_rate?: number | null;
  body_weight?: number | null;
  note?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateVitalInput {
  recorded_at: string;
  temperature?: number | null;
  heart_rate?: number | null;
  respiratory_rate?: number | null;
  body_weight?: number | null;
  note?: string | null;
}
```

## 完了条件
- [ ] バイタル一覧が `recorded_at` 昇順で表示される
- [ ] 新規バイタルを追加できる（フォーム送信後にキャッシュ invalidate）
- [ ] バイタルを行内編集または編集ダイアログで編集できる
- [ ] バイタルを削除できる（確認ダイアログあり）
- [ ] 体温・心拍・呼吸・体重はすべて任意入力（空欄可）
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend pnpm build` が通る

## 備考
- バックエンドチケット 004 は実装済みのためフロントエンド単独で進められる
- 数値フィールドは `Input type="number"` を使い、空送信時は `null` で送る
- タブ組み込み先の詳細画面コンポーネントのパスを事前に確認すること
