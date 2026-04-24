# 012: カルテ健診記録タブ実装

## 概要
外来カルテ詳細画面に健診記録タブを追加する。健診種別・健診日・次回予定日・結果を管理する。

## 優先度
medium

## 関連APIエンドポイント
- GET `/v1/medical-records/{id}/checkups`
- POST `/v1/medical-records/{id}/checkups`
- PATCH `/v1/medical-records/{id}/checkups/{checkupId}`
- DELETE `/v1/medical-records/{id}/checkups/{checkupId}`

## 関連バックエンドチケット
backend/issues/open/016_checkup_crud.md（未実装）

## 実装内容

### API層 (`features/medical-records/api/`)
`checkups.ts` を新規作成し以下の TanStack Query hooks を実装する。
- `useCheckups(medicalRecordId: string)` — GET 一覧取得（`date` 昇順）
- `useCreateCheckup(medicalRecordId: string)` — POST 新規追加
- `useUpdateCheckup(medicalRecordId: string)` — PATCH 編集
- `useDeleteCheckup(medicalRecordId: string)` — DELETE 削除

健診種別マスタ（`checkup_types`）の取得には既存の `features/master` API を参照すること。

### コンポーネント (`features/medical-records/components/`)
`CheckupsTab/` ディレクトリを新規作成する。
- `CheckupsTab.tsx` — 健診記録一覧テーブル + 追加フォーム
- `CheckupsTab/index.ts` — named export

テーブルカラム: `date`（健診日）・`checkup_type`（健診種別）・`next_scheduled_date`（次回予定日）・`result`（結果）・操作列
健診種別は `Select` コンポーネントで `checkup_types` マスタから選択する。

### ページ/ルート (`features/medical-records/routes/`)
カルテ詳細ルートコンポーネントのタブ一覧に「健診」タブを追加し、`CheckupsTab` を組み込む。

### 型定義 (`features/medical-records/types/`)
`index.ts` に以下を追加する。
```typescript
export interface Checkup {
  id: string;
  medical_record_id: string;
  checkup_type_id: string;
  checkup_type_name?: string;
  date: string;              // YYYY-MM-DD
  next_scheduled_date?: string | null;
  result?: string | null;
  note?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateCheckupInput {
  checkup_type_id: string;
  date: string;
  next_scheduled_date?: string | null;
  result?: string | null;
  note?: string | null;
}
```

## 完了条件
- [ ] 健診記録一覧が `date` 昇順で表示される
- [ ] 健診種別を `checkup_types` マスタのセレクトで選択できる
- [ ] 健診記録を追加・編集・削除できる
- [ ] 次回予定日を設定・表示できる
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend pnpm build` が通る

## 備考
- バックエンドチケット 016 が未実装のため、バックエンド実装完了を待ってから着手すること
- バックエンド実装完了後、API レスポンス構造を `backend/docs/api.yaml` で確認してから型定義を確定させること
