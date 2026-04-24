# FE-238: use-trimming-form・use-vaccination-form の catch で toast.error 直呼び

## 概要

2つのフォームフックで、`useActionState` の catch ブロックが
`handleApiError` を使わずに `toast.error()` を直接呼び出している。
プロジェクト規約「すべての catch ブロックで `handleApiError` を呼び出す」に違反。

## 問題箇所

### `frontend/src/features/trimming/hooks/use-trimming-form.ts:174-176`

```ts
// Before
} catch {
  toast.error("保存に失敗しました");  // ← handleApiError でない
}

// After
} catch (error) {
  handleApiError(error, "トリミング記録の保存");
}
```

### `frontend/src/features/vaccinations/hooks/use-vaccination-form.ts:198-200`

```ts
// Before
} catch {
  toast.error("保存に失敗しました");  // ← handleApiError でない
}

// After
} catch (error) {
  handleApiError(error, "ワクチン接種記録の保存");
}
```

## 影響

`handleApiError` を使わないと：
- HTTP ステータス別メッセージ分岐が機能しない（401/403/404/409/500）
- バックエンドの `{"message": "..."}` が無視される
- 「保存に失敗しました」という汎用メッセージのみが表示される

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックで `handleApiError(error, "コンテキスト")` を呼び出す。

### 関連チケット
- FE-218: 同種問題（OwnersList, CompanySettings）
- FE-235: 同種問題（use-examination-form）

## 優先度
**Medium** — API エラー詳細がユーザーに正確に伝わらない。

## 関連ファイル
- `frontend/src/features/trimming/hooks/use-trimming-form.ts`
- `frontend/src/features/vaccinations/hooks/use-vaccination-form.ts`
- `frontend/src/lib/handle-api-error.ts`
