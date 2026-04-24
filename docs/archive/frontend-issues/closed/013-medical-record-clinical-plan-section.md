# 013: カルテ診察所見・診断・治療方針セクション実装

## 概要
カルテ詳細画面の診察タブに身体検査所見・診断カテゴリ・診断病名・治療方針を入力・表示するセクションを実装する。1カルテにつき1レコード（GET で取得し PATCH で更新）。診断カテゴリ変更時は診断病名のセレクトを連動リセットする。

## 優先度
high

## 関連APIエンドポイント
- GET `/v1/medical-records/{id}/clinical-plans`（GetOrCreate — 常に1件返る）
- PATCH `/v1/medical-records/{id}/clinical-plans/{planId}`

## 関連バックエンドチケット
backend/issues/open/017_clinical_plan_crud.md（未実装）

## 実装内容

### API層 (`features/medical-records/api/`)
`clinical-plan.ts` を新規作成し以下の TanStack Query hooks を実装する。
- `useClinicalPlan(medicalRecordId: string)` — GET（または GetOrCreate）1件取得
- `useUpdateClinicalPlan(medicalRecordId: string)` — PATCH 更新

### コンポーネント (`features/medical-records/components/`)
`ClinicalPlanSection/` ディレクトリを新規作成する。
- `ClinicalPlanSection.tsx` — 所見・診断・治療方針の入力フォームセクション
- `ClinicalPlanSection/index.ts` — named export

フィールド仕様:
- `physical_findings`（身体検査所見）— `Textarea`（大型テキスト入力）
- `diagnosis_category_id`（診断カテゴリ）— `Select`、変更時に `diagnosis_id` をリセット
- `diagnosis_id`（診断病名）— `Select`、選択中のカテゴリでフィルタ
- `treatment_plan`（治療方針）— `Textarea`（大型テキスト入力）

保存方式: debounce 500ms による自動保存、または明示的「保存」ボタン（Figma 確認後に決定）。

### ページ/ルート (`features/medical-records/routes/`)
カルテ詳細ルートコンポーネントの診察タブ内に `ClinicalPlanSection` を組み込む。カルテ開封時にデータを自動ロードする。

### 型定義 (`features/medical-records/types/`)
`index.ts` に以下を追加する。
```typescript
export interface ClinicalPlan {
  id: string;
  medical_record_id: string;
  physical_findings?: string | null;
  diagnosis_category_id?: string | null;
  diagnosis_id?: string | null;
  treatment_plan?: string | null;
  created_at: string;
  updated_at: string;
}

export interface UpdateClinicalPlanInput {
  physical_findings?: string | null;
  diagnosis_category_id?: string | null;
  diagnosis_id?: string | null;
  treatment_plan?: string | null;
}
```

## 完了条件
- [ ] カルテ開封時に所見・診断・治療方針が自動読み込みされる
- [ ] 自動保存（debounce 500ms）または明示的保存ボタンでデータが永続化される
- [ ] 診断カテゴリ変更時に診断病名セレクトがリセットされる
- [ ] 診断病名はカテゴリに連動してフィルタされる
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend pnpm build` が通る

## 備考
- バックエンドチケット 017 が未実装のため、バックエンド実装完了を待ってから着手すること
- 診断カテゴリ・診断病名のマスタ取得 API は既存の `features/master` を参照すること
- debounce 実装は `@/hooks/useDebounce.ts` の共有 hook を使用すること
