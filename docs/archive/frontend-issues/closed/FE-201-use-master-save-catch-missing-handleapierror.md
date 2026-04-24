# FE-201: use-master-save.ts の catch ブロックで handleApiError を未使用

## 概要

`use-master-save.ts` の `onSuccess` コールバック内の catch ブロックが
`handleApiError` を呼ばず `toast.error()` のみを使用している。
プロジェクト規約ではすべての catch ブロックで `handleApiError` を呼び出すことが必須。

## 現状コード

### `frontend/src/features/master/hooks/use-master-save.ts:60-73`（更新時）
```ts
onSuccess: async (savedData) => {
  try {
    const saved = savedData as T;
    if (onSuccess) {
      await onSuccess(saved, data);  // ユーザー提供の onSuccess を await
    }
    toast.success("更新しました");
    crudHandleClose();
  } catch (error) {
    toast.error("保存に失敗しました");  // ← handleApiError 未使用
  }
},
onError: () => toast.error("更新に失敗しました"),
```

### `frontend/src/features/master/hooks/use-master-save.ts:77-90`（作成時）
```ts
onSuccess: async (savedData) => {
  try {
    const saved = savedData as T;
    if (onSuccess) {
      await onSuccess(saved, data);
    }
    toast.success("登録しました");
    crudHandleClose();
  } catch (error) {
    toast.error("保存に失敗しました");  // ← handleApiError 未使用
  }
},
onError: () => toast.error("登録に失敗しました"),
```

## 問題

`onSuccess` 内で `await onSuccess(saved, data)` が例外を投げた場合（例: 追加処理の API コールに失敗した場合）、
`handleApiError` が呼ばれないため:
1. エラーの詳細が握り潰される（型・ステータスコード・メッセージが捨てられる）
2. API エラー（401, 403, 409 等）が `toast.error("保存に失敗しました")` という汎用メッセージに統一されてしまう

## 影響範囲

| 対象 | 行番号 | 状態 |
|------|--------|------|
| `frontend/src/features/master/hooks/use-master-save.ts` | 68-70（更新） | 要修正 |
| `frontend/src/features/master/hooks/use-master-save.ts` | 85-87（作成） | 要修正 |

## 修正方針

### `use-master-save.ts:68-70` と `:85-87`
```ts
// Before
} catch (error) {
  toast.error("保存に失敗しました");
}

// After
} catch (error) {
  handleApiError(error, "保存後処理");
}
```

import 追加が必要:
```ts
import { handleApiError } from "@/lib/handle-api-error";
```

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md` — Frontend エラーハンドリング
> すべての `catch` ブロックで `handleApiError` を呼び出す。
> `toast.error()` の直接使用は禁止。

### プロジェクト内参照実装
- `frontend/src/features/owners/hooks/use-owner-form.ts` — `handleApiError(error, "オーナーの更新")` で正しく実装

## 優先度
**Medium** — マスタ更新後の onSuccess 処理が失敗した場合のエラー詳細が失われる。

## 関連ファイル
- `frontend/src/features/master/hooks/use-master-save.ts:60-90` — 要修正
- `frontend/src/lib/handle-api-error.ts` — handleApiError 定義
