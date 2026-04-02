# 010: カルテ詳細治療明細タブ実装

## 概要
外来カルテ詳細画面に治療明細タブを追加する。診療・処置・薬品ごとの明細を管理し、合計金額を自動計算して会計に連携する。

## 優先度
high

## 関連APIエンドポイント
- GET `/v1/medical-records/{id}/treatments`
- POST `/v1/medical-records/{id}/treatments`
- PATCH `/v1/medical-records/{id}/treatments/{treatmentId}`
- DELETE `/v1/medical-records/{id}/treatments/{treatmentId}`
- PUT `/v1/medical-records/{id}/treatments`（並び替え bulk update）

## 関連バックエンドチケット
backend/issues/open/005_medical_record_treatments.md（実装済み）

## 実装内容

### API層 (`features/medical-records/api/`)
`treatments.ts` を新規作成し以下の TanStack Query hooks を実装する。
- `useTreatments(medicalRecordId: string)` — GET 一覧取得（`sort_order` 昇順）
- `useCreateTreatment(medicalRecordId: string)` — POST 新規追加
- `useUpdateTreatment(medicalRecordId: string)` — PATCH 編集
- `useDeleteTreatment(medicalRecordId: string)` — DELETE 削除
- `useReorderTreatments(medicalRecordId: string)` — PUT 並び替え

### コンポーネント (`features/medical-records/components/`)
`TreatmentsTab/` ディレクトリを新規作成する。
- `TreatmentsTab.tsx` — 明細リスト + フッターに合計金額表示 + 追加フォーム
- `TreatmentRow.tsx` — 各明細行（インライン編集対応）
- `TreatmentsTab/index.ts` — named export

`item_type` 別にグループ表示またはバッジ表示を行う（`consultation` / `procedure` / `medicine` / `other`）。
`selected` フラグでチェックボックスによる有効/無効の切り替えを行う。
合計金額は `unit_price × quantity - discount_amount` をクライアント側で計算して表示する。
`insurance` フラグはバッジまたはアイコンで表示する。

### ページ/ルート (`features/medical-records/routes/`)
カルテ詳細ルートコンポーネントのタブ一覧に「治療明細」タブを追加し、`TreatmentsTab` を組み込む。

### 型定義 (`features/medical-records/types/`)
`index.ts` に以下を追加する。
```typescript
export type TreatmentItemType = 'consultation' | 'procedure' | 'medicine' | 'other';

export interface Treatment {
  id: string;
  medical_record_id: string;
  item_type: TreatmentItemType;
  item_id?: string | null;
  name: string;
  unit_price: number;
  quantity: number;
  discount_amount: number;
  insurance: boolean;
  selected: boolean;
  sort_order: number;
  note?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateTreatmentInput {
  item_type: TreatmentItemType;
  item_id?: string | null;
  name: string;
  unit_price: number;
  quantity: number;
  discount_amount?: number;
  insurance?: boolean;
  selected?: boolean;
  sort_order?: number;
  note?: string | null;
}
```

## 完了条件
- [ ] 治療明細一覧が `sort_order` 昇順で表示される
- [ ] 明細を追加・編集・削除できる
- [ ] 並び替え（PUT bulk update）が動作する
- [ ] 合計金額（`unit_price × quantity - discount_amount`）がリアルタイムで自動計算される
- [ ] `insurance` フラグが各行に表示される
- [ ] `selected` フラグの切り替えができる
- [ ] `item_type` 別の表示区分けが実装されている
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend npm run build` が通る

## 備考
- バックエンドチケット 005 は実装済みのためフロントエンド単独で進められる
- 並び替えは drag-and-drop または上下ボタンのいずれかで実装する（Figmaデザイン確認後に決定）
- 合計金額計算はフロントエンド側で行うが、バックエンドの返却値とも突合する
