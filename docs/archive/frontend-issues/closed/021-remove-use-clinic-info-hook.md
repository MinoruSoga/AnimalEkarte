# 021: use-clinic-info hook を削除し /me の clinic 情報を使う

## 概要

`src/hooks/use-clinic-info.ts` を削除し、`/me` レスポンスの `clinic` フィールドを
`authStore` 経由で参照するよう変更する。

## 背景

現状 `use-clinic-info.ts` は `GET /v1/clinics`（全件取得）を叩いて先頭1件を医院情報として使っている。
これは無駄なAPIコールであり、アーキテクチャ上も shared 層が features 層のエンドポイントと
重複するという問題を抱えている。

backend/issues/open/001-add-clinic-to-me-endpoint.md の対応完了後に着手する。

## 修正内容

### 1. authStore に clinic 情報を追加

```typescript
// src/stores/authStore.ts
interface AuthState {
  user: User | null;
  clinic: Clinic | null;  // 追加
}
```

### 2. 使用箇所を authStore 参照に変更

| ファイル | 変更内容 |
|---------|---------|
| `src/components/shared/Layout/Sidebar.tsx` | `useClinicInfo()` → `useAuthStore(s => s.clinic)` |
| `src/features/hospital-settings/routes/ClinicSettings.tsx` | 同上 |
| `src/features/accounting/components/AccountingDocument.tsx` | 同上 |

### 3. use-clinic-info.ts を削除

`src/hooks/use-clinic-info.ts` を削除する。

## 前提条件

- backend/issues/open/001-add-clinic-to-me-endpoint.md が完了していること
- `/me` レスポンスに `clinic` フィールドが含まれていること

## 優先度

medium

## 関連イシュー

- backend/issues/open/001-add-clinic-to-me-endpoint.md
