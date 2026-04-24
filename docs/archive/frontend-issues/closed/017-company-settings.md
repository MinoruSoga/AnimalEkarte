# 017: 法人情報設定画面実装

## 概要
運営法人（会社）情報の閲覧・編集画面を設定ページに追加する。法人名・郵便番号・住所・電話番号・メールアドレスを管理する。

## 優先度
low

## 関連APIエンドポイント
- GET `/v1/company`
- PATCH `/v1/company`

## 関連バックエンドチケット
なし（バックエンド実装状況を確認してから着手すること）

## 実装内容

### API層 (`features/master/api/`)
`company.ts` を新規作成し以下の TanStack Query hooks を実装する。
- `useCompany()` — GET 法人情報取得
- `useUpdateCompany()` — PATCH 法人情報更新

### コンポーネント (`features/master/components/`)
既存の設定コンポーネントパターン（`StaffSettings` 等）に倣い、法人情報フォームセクションを実装する。
フィールド: `company_name`（法人名）・`postal_code`（郵便番号）・`address`（住所）・`phone`（電話番号）・`email`（メールアドレス）

### ページ/ルート (`features/master/routes/`)
既存の `Settings.tsx` に「法人情報」セクションを追加するか、専用の `CompanySettings.tsx` を新規作成してルート（`/settings/company`）に追加する。既存設定ページのレイアウトパターンに合わせること。

### 型定義 (`features/master/api/types.ts`)
既存の `types.ts` に以下を追加する。
```typescript
export interface Company {
  id: string;
  company_name: string;
  postal_code?: string | null;
  address?: string | null;
  phone?: string | null;
  email?: string | null;
  created_at: string;
  updated_at: string;
}

export interface UpdateCompanyInput {
  company_name?: string;
  postal_code?: string | null;
  address?: string | null;
  phone?: string | null;
  email?: string | null;
}
```

## 完了条件
- [ ] 設定ページから法人情報セクション（または `/settings/company` ルート）にアクセスできる
- [ ] 法人情報（法人名・郵便番号・住所・電話・メール）を閲覧できる
- [ ] 法人情報を編集・保存できる
- [ ] 保存成功時にトースト通知が表示される
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend pnpm build` が通る

## 備考
- バックエンド API の実装状況を事前に `backend/docs/api.yaml` で確認すること
- 設定ページへの統合方針（既存 `Settings.tsx` に追加 vs 新規ルート）は既存コードを確認してから決定すること
- `features/master/api/` に既存の API ファイルが複数あるため、命名・パターンを揃えること
