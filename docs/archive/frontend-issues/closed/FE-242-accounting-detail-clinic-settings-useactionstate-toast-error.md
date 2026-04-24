# FE-242: AccountingDetail・ClinicMasterSettings の useActionState catch で toast.error 直呼び

## 概要

2つのルートコンポーネントで `useActionState` の catch ブロックが
`handleApiError` を使わずに `toast.error()` を直接呼び出している。

## 問題箇所

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:207-209付近`

```ts
// Before: useActionState 内の catch で toast 直呼び
} catch {
  toast.error("保存に失敗しました");
}

// After
} catch (error) {
  handleApiError(error, "会計情報の保存");
}
```

### `frontend/src/features/hospital-settings/routes/ClinicMasterSettings.tsx:207-209付近`

```ts
// Before: 同様
} catch {
  toast.error("保存に失敗しました");
}

// After
} catch (error) {
  handleApiError(error, "クリニック設定の保存");
}
```

## 影響

`handleApiError` を使わないと：
- HTTP 409（重複エラー）でも「保存に失敗しました」という汎用メッセージのみ表示
- バックエンドの `{"message": "クリニック名が既に使用されています"}` 等が失われる

## 備考

FE-210（AccountingDetail の isError 未チェック + handleUpdateItemTax）は別の問題。
本チケットは `useActionState` 内の catch の問題のみを対象とする。

FE-219（hospital-settings の clinics.ts mutation の onError 欠落）は API フックの問題。
本チケットはルートコンポーネントの useActionState の問題。

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックで `handleApiError(error, "コンテキスト")` を呼び出す。

## 優先度
**Medium** — 会計情報・クリニック設定の保存失敗時にエラー詳細が伝わらない。

## 関連ファイル
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
- `frontend/src/features/hospital-settings/routes/ClinicMasterSettings.tsx`
- `frontend/src/lib/handle-api-error.ts`
