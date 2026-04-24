# FE-073: CompanySettings UI にインボイス番号フィールド追加

**Status**: Closed
**Priority**: Medium
**Affects**: 法人情報設定画面（CompanySettings）
**Date Created**: 2026-03-18
**Related**: TASK-020, BE-045

## Summary

法人情報（CompanySettings）画面の読み取りビューと編集フォームに「インボイス番号」フィールドを追加する。既存の「法人番号」フィールドと同じ PropertyRow パターンで実装する。

## 現状のコード

### API transform

```typescript
// frontend/src/features/master/api/company.ts:27-43
function transformCompany(data: ModelCompany) {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    postalCode: data.postal_code,
    address: data.address,
    phoneNumber: data.phone_number,
    faxNumber: data.fax_number,
    email: data.email,
    website: data.website,
    directorName: data.director_name,
    registrationNumber: data.registration_number,
    logoUrl: data.logo_url,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}
```

### API Request 型

```typescript
// frontend/src/features/master/api/company.ts:10-21
export interface UpdateCompanyRequest {
  name?: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  email?: string;
  website?: string;
  director_name?: string;
  registration_number?: string;
  logo_url?: string;
}
```

### フォーム状態

```typescript
// frontend/src/features/master/routes/CompanySettings.tsx:27-37
interface CompanyFormData {
  name: string;
  postal_code: string;
  address: string;
  phone_number: string;
  fax_number: string;
  email: string;
  website: string;
  director_name: string;
  registration_number: string;
}
```

### 読み取りビュー（法人番号の直後に追加する）

```typescript
// frontend/src/features/master/routes/CompanySettings.tsx:216-220
<PropertyRow label="法人番号">
  <span className={`text-base ${company.registrationNumber ? C.text : C.text30}`}>
    {company.registrationNumber || "空"}
  </span>
</PropertyRow>
```

### 編集フォーム（法人番号の直後に追加する）

```typescript
// frontend/src/features/master/routes/CompanySettings.tsx:351-361
<PropertyRow label="法人番号">
  <input
    type="text"
    className={PROP_INPUT_CLASS}
    value={formData.registration_number}
    onChange={(e) =>
      setFormData((prev) => ({ ...prev, registration_number: e.target.value }))
    }
    placeholder="例: 1234567890123"
  />
</PropertyRow>
```

## 必要な変更

### 1. API transform に追加

```typescript
// frontend/src/features/master/api/company.ts — transformCompany() に追加
invoiceRegistrationNumber: data.invoice_registration_number,
```

### 2. API Request 型に追加

```typescript
// frontend/src/features/master/api/company.ts — UpdateCompanyRequest に追加
invoice_registration_number?: string;
```

### 3. CompanyFormData に追加

```typescript
// frontend/src/features/master/routes/CompanySettings.tsx — CompanyFormData に追加
invoice_registration_number: string;

// DEFAULT_FORM_DATA にも追加
invoice_registration_number: "",
```

### 4. useEffect のフォーム初期化に追加

```typescript
// frontend/src/features/master/routes/CompanySettings.tsx:66-77 — setFormData 内に追加
invoice_registration_number: company.invoiceRegistrationNumber,
```

### 5. handleCloseEdit のリセットに追加

```typescript
// frontend/src/features/master/routes/CompanySettings.tsx:87-97 — setFormData 内に追加
invoice_registration_number: company.invoiceRegistrationNumber,
```

### 6. handleSave のリクエストに追加

```typescript
// frontend/src/features/master/routes/CompanySettings.tsx:107-117 — req に追加
invoice_registration_number: formData.invoice_registration_number || undefined,
```

### 7. 読み取りビューに PropertyRow 追加

法人番号の PropertyRow の直後（220行目の後）に追加:

```typescript
<PropertyRow label="インボイス番号">
  <span className={`text-base font-mono ${company.invoiceRegistrationNumber ? C.text : C.text30}`}>
    {company.invoiceRegistrationNumber || "空"}
  </span>
</PropertyRow>
```

### 8. 編集フォームに PropertyRow 追加

法人番号の PropertyRow の直後（361行目の後）に追加:

```typescript
<PropertyRow label="インボイス番号">
  <input
    type="text"
    className={PROP_INPUT_CLASS}
    value={formData.invoice_registration_number}
    onChange={(e) =>
      setFormData((prev) => ({ ...prev, invoice_registration_number: e.target.value }))
    }
    placeholder="例: T1234567890123"
  />
</PropertyRow>
```

## UI 操作フロー

1. ユーザーが設定 → 法人情報画面を開く
2. 「編集」ボタンをクリックし右サイドピークを開く
3. 「インボイス番号」フィールドに `T1234567890123` を入力
4. 「保存」ボタンをクリック
5. 左の読み取りビューに入力した値が表示される

## プロジェクトルール遵守チェック

- [x] `any` 型なし
- [x] `FC` / `forwardRef` なし
- [x] barrel index 経由 import なし
- [x] 条件レンダー `? ... : null`（`&&` 禁止）
- [x] 型は `models.ts` から導出（手書き interface なし — `CompanyFormData` はUI専用型のため許容）
- [x] 既存パターン（PropertyRow）を踏襲

## 依存関係

**BE-045 がブロッカー** — BE-045 の完了 + `make codegen` 実行が必須。それまで実装を開始しないこと。

- BE-045 が先に完了している必要がある（`invoice_registration_number` が API レスポンスに含まれる）
- `make codegen` で `models.ts` の `Company` 型が更新されている必要がある

## 前提条件チェック（実装開始前に確認）

- [ ] BE-045 が `closed/` に移動済み
- [ ] `make codegen` 実行済み
- [ ] `models.ts` の `Company` 型に `invoice_registration_number` フィールドが存在する
- [ ] `docker compose exec frontend pnpm build` で型エラーなし

## 完了条件

- [ ] 型エラーなし（`pnpm build` パス）
- [ ] ESLint エラーなし（`pnpm lint` パス）
- [ ] 法人情報の読み取りビューに「インボイス番号」が表示される
- [ ] 編集フォームで「インボイス番号」を入力・保存できる
- [ ] 保存後リロードしても値が保持される
- [ ] 空欄で保存してもエラーにならない

## クローズ情報

- **Closed At**: 2026-03-18
- **変更ファイル**:
  - `frontend/src/features/master/api/company.ts` — `transformCompany()` に `invoiceRegistrationNumber` 追加、`UpdateCompanyRequest` に `invoice_registration_number` 追加
  - `frontend/src/features/master/routes/CompanySettings.tsx` — `CompanyFormData` + `DEFAULT_FORM_DATA` + useEffect初期化 + handleCloseEditリセット + handleSaveリクエスト + 読み取りビューPropertyRow + 編集フォームPropertyRow にインボイス番号フィールド追加
