# TASK-020: 会社設定にインボイス番号（適格請求書発行事業者登録番号）を追加

**作成日**: 2026-03-18
**ステータス**: Closed
**依頼元**: ユーザー

---

## 概要

法人情報（会社設定）画面に「インボイス番号（適格請求書発行事業者登録番号）」の入力・表示フィールドを追加する。`company` テーブルに `invoice_registration_number` カラムを新設し、Backend API・Frontend UI を対応させる。

## 依頼内容（原文）

> 会社設定にてインボイス番号を入力

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | 対象テーブル: company のみか clinics にも必要か？ | company のみ |
| 2 | カラム名 | `invoice_registration_number` |
| 3 | UIラベル | デフォルト採用: 「インボイス番号」（placeholder: `例: T1234567890123`） |
| 4 | バリデーション | デフォルト採用: フロントエンドでヒント表示のみ、厳密バリデーションなし |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | DB + Go モデル + API 全層に `invoice_registration_number` 追加 | BE/DB | BE-045 | - | [x] |
| 2 | CompanySettings UI にインボイス番号フィールド追加 | FE | FE-073 | #1 | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: 法人情報編集画面で「インボイス番号」フィールドに `T1234567890123` を入力し保存すると、読み取り専用ビューに表示される
- [ ] AC-2: 保存後にページをリロードしても値が保持されている（DB に永続化されている）
- [ ] AC-3: フィールドを空欄にして保存した場合、エラーにならず空として保存される
- [ ] AC-4: GET /v1/company のレスポンスに `invoice_registration_number` フィールドが含まれる
- [ ] AC-5: PATCH /v1/company に `invoice_registration_number` を送信すると値が更新される

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 対象テーブル | `company` のみ | インボイス登録番号は法人単位で付与される | `clinics` にも追加 |
| バリデーション | なし（空欄許容） | 登録番号未取得の法人もある。UIヒントで十分 | T + 13桁のフォーマット検証 |

## 影響範囲

### DB
- テーブル: `company` — `invoice_registration_number text NOT NULL DEFAULT ''` カラム追加

### Backend
- `backend/internal/model/company.go` — `InvoiceRegistrationNumber` フィールド追加
- `backend/internal/handler/company_request.go` — `updateCompanyRequest` にフィールド追加
- `backend/internal/handler/company_response.go` — `companyResponse` + `toCompanyResponse()` にフィールド追加
- `backend/internal/handler/company_handler.go` — `UpdateCompany()` で input に渡す
- `backend/internal/service/company_service.go` — `UpdateCompanyInput` + `buildCompanyUpdateFields()` にフィールド追加

### Frontend
- `frontend/src/types/generated/models.ts` — `make codegen` で自動更新
- `frontend/src/features/master/api/company.ts` — `UpdateCompanyRequest`, `transformCompany()` にフィールド追加
- `frontend/src/features/master/routes/CompanySettings.tsx` — `CompanyFormData`, 読み取りビュー, 編集フォームにフィールド追加

## 参照実装

- `frontend/src/features/master/routes/CompanySettings.tsx` — 既存の `registration_number`（法人番号）フィールドの実装パターンをそのまま踏襲

## リスク・懸念事項

特になし。既存フィールドと同じパターンの単純追加。

## 未解決事項

なし

## 実装順序

1. DB マイグレーション（`001_init.sql` に `invoice_registration_number` カラム追加）
2. Go モデル修正 → `make codegen`
3. Backend API 全層（handler → service）にフィールド追加
4. Frontend API transform + 型 + UI フィールド追加

## 関連イシュー

- BE-045: [company テーブルに invoice_registration_number 追加](../../backend/issues/open/BE-045-company-invoice-registration-number.md)
- FE-073: [CompanySettings UI にインボイス番号フィールド追加](../../frontend/issues/open/FE-073-company-invoice-registration-number-ui.md)
