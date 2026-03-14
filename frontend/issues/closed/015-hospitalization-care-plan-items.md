# 015: 入院ケアプラン項目 UI 実装

## 概要
入院詳細画面にケアプラン項目タブを追加する。投薬・処置・ケア・指導のプランを管理し、timing（朝/昼/夜）の複数選択・ステータス管理ができる。

## 優先度
high

## 関連APIエンドポイント
- GET `/v1/hospitalizations/{id}/care-plan-items`
- POST `/v1/hospitalizations/{id}/care-plan-items`
- PATCH `/v1/hospitalizations/{id}/care-plan-items/{itemId}`
- DELETE `/v1/hospitalizations/{id}/care-plan-items/{itemId}`

## 関連バックエンドチケット
backend/issues/open/010_care_plan_items.md（実装済み）

## 実装内容

### API層 (`features/hospitalization/api/`)
`care-plan-items.ts` を新規作成し以下の TanStack Query hooks を実装する。
- `useCarePlanItems(hospitalizationId: string)` — GET 一覧取得（`sort_order` 昇順）
- `useCreateCarePlanItem(hospitalizationId: string)` — POST 新規追加
- `useUpdateCarePlanItem(hospitalizationId: string)` — PATCH 編集
- `useDeleteCarePlanItem(hospitalizationId: string)` — DELETE 削除

### コンポーネント (`features/hospitalization/components/`)
`CarePlanTab/` ディレクトリを新規作成する。
- `CarePlanTab.tsx` — ケアプラン項目リスト + 追加フォーム
- `CarePlanItemRow.tsx` — 各プラン行（timing チェックボックス・ステータスバッジ・操作ボタン）
- `CarePlanTab/index.ts` — named export

`timing` フィールドの仕様（複数選択チェックボックス）:
- `morning`（朝）
- `noon`（昼）
- `night`（夜）
バックエンドへは選択済み値の配列として送信する。

`type` 別アイコン表示:
- `medicine` — 薬アイコン
- `procedure` — 処置アイコン
- `care` — ケアアイコン
- `other` — その他アイコン

`status` バッジ:
- `active` — 「実施中」（blue）
- `completed` — 「完了」（green）
- `discontinued` — 「中止」（gray）

### ページ/ルート (`features/hospitalization/routes/`)
入院詳細ルートコンポーネントのタブ一覧に「ケアプラン」タブを追加し、`CarePlanTab` を組み込む。

### 型定義 (`features/hospitalization/types/`)
`index.ts` に以下を追加する。
```typescript
export type CarePlanItemType = 'medicine' | 'procedure' | 'care' | 'other';
export type CarePlanItemStatus = 'active' | 'completed' | 'discontinued';
export type CarePlanTiming = 'morning' | 'noon' | 'night';

export interface CarePlanItem {
  id: string;
  hospitalization_id: string;
  type: CarePlanItemType;
  name: string;
  timing: CarePlanTiming[];
  status: CarePlanItemStatus;
  sort_order: number;
  start_date?: string | null;
  end_date?: string | null;
  note?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateCarePlanItemInput {
  type: CarePlanItemType;
  name: string;
  timing: CarePlanTiming[];
  status?: CarePlanItemStatus;
  sort_order?: number;
  start_date?: string | null;
  end_date?: string | null;
  note?: string | null;
}
```

## 完了条件
- [ ] ケアプラン項目一覧が `sort_order` 昇順で表示される
- [ ] 項目を追加・編集・削除できる
- [ ] `timing` の複数選択（朝/昼/夜）が正しく保存・表示される
- [ ] `type` 別アイコンが表示される
- [ ] `status` バッジが表示され、ステータス変更ができる
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend npm run build` が通る

## 備考
- バックエンドチケット 010 は実装済みのためフロントエンド単独で進められる
- `timing` は配列型のため、バックエンドが JSON 配列として返すか文字列 CSV で返すかを `backend/docs/api.yaml` で確認してから transforms を実装すること
