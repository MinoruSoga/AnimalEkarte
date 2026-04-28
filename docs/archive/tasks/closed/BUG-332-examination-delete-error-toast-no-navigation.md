# BUG-332: 検査削除後にエラートーストが表示されページ遷移しない

**Status**: CLOSED  
**Priority**: Medium  
**Discovery**: 機能テスト Section 5 検査管理 (2026-04-12)

## 概要

`/examinations/:id` で削除確認ダイアログを承認した後、「サーバーエラーが発生しました。しばらく経ってから再度お試しください。」エラートーストが表示され、一覧ページへの遷移が行われない。手動で API を呼び出すと 204 が返るため、API 自体は正常。フロントエンドの削除フックが `onSuccess` でなく `onError` パスに落ちている。

## 再現手順

1. 管理者アカウントでログイン
2. `/examinations` で既存の検査記録を開く
3. 削除ボタンをクリック → 確認ダイアログが開く
4. 「削除する」ボタンをクリック
5. **結果**: 「サーバーエラーが発生しました」エラートーストが表示される。ページは `/examinations/:id` にとどまる
6. **期待**: 「検査記録を削除しました」成功トーストが表示され、`/examinations` 一覧に遷移する

## 現状コード

### `frontend/src/features/examinations/api/delete-examination.ts:5-21`
```ts
export const deleteExamination = async (id: string): Promise<void> => {
  await axios.delete(`/v1/examinations/${id}`);
};

export const useDeleteExamination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteExamination,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["examinations"] });
    },
    onError: (error) => {
      handleApiError(error, "削除");   // ← ここが発火している
    },
  });
};
```

### `frontend/src/features/examinations/hooks/use-examination-form.ts:150-160`
```ts
const handleDelete = useCallback((onSuccess?: () => void) => {
  if (!isEdit || !id) return;
  startDeleteTransition(() => {
    deleteMutation.mutate(id, {
      onSuccess: () => {
        toast.success("検査記録を削除しました");
        onSuccess?.();   // ← navigate() が渡されているがここに到達しない
      },
    });
  });
}, [isEdit, id, deleteMutation, startDeleteTransition]);
```

### `frontend/src/features/examinations/routes/ExaminationForm.tsx:334-338`
```tsx
const handleDeleteConfirm = useCallback(() => {
  markClean();
  handleDelete(() => navigate(paths.examinations.getHref()));
  setIsDeleteConfirmOpen(false);
}, [markClean, handleDelete, navigate]);
```

## 調査済み事項

| 確認項目 | 結果 |
|---------|------|
| バックエンド DELETE エンドポイント | `DELETE /api/v1/examinations/:id` — 204 返却、実装は正常 |
| 手動 `fetch('/api/v1/examinations/4', {method:'DELETE'})` | 204 — Vite プロキシ経由では正常動作 |
| CORS 設定 (`docker-compose.yml`) | `http://localhost:3003` は許可済み |
| axios baseURL | `http://localhost:8080/api` — 直接バックエンド接続（Vite プロキシ非経由） |
| 認証クッキー | `withCredentials: true` で自動送信 |

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/examinations/api/delete-examination.ts:5` | `deleteExamination` 関数 | 要調査 |
| `frontend/src/features/examinations/hooks/use-examination-form.ts:150` | `handleDelete` コールバック | 要調査 |

## 修正方針

### 調査必要事項（まず確認すること）

ブラウザの Network タブで削除実行時のリクエスト/レスポンスを確認する：
1. `DELETE http://localhost:8080/api/v1/examinations/:id` のステータスコードを確認
2. レスポンスヘッダー `Access-Control-Allow-Origin` が含まれているか確認
3. エラーの場合、エラーメッセージを確認

### 方針 1: Vite プロキシ経由に変更

`deleteExamination` で使用する URL を相対パスに変更して Vite プロキシ経由にする。

他の API 関数との一貫性の観点では、全 axios 呼び出しが `baseURL: http://localhost:8080/api` を使っているため、この特定の関数だけ変更するのは不整合になる可能性がある。

### 方針 2: バックエンドエラー原因特定

DELETE リクエスト時のバックエンドログを確認し、エラーが発生しているのかどうか（400, 403, 404, 500 のいずれか）を特定する。エラーが返っているなら `handleApiError` が正しくエラートーストを表示している。

### 方針 3: invalidateQueries のエラー伝播確認

`queryClient.invalidateQueries` でエラーが発生し、それが mutation の `onError` を呼び出している可能性は低いが、確認が必要。

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> ```typescript
> // ✅ MANDATE: すべての catch ブロックで handleApiError を使用
> try {
>   await api.updateOwner(id, data);
> } catch (error) {
>   handleApiError(error, "オーナーの更新");
> }
> ```

`onError` に `handleApiError` を使うのは正しいが、成功すべきリクエストが `onError` に落ちていることが問題。

### プロジェクト内参照実装
- `frontend/src/features/owners/api/delete-owner.ts` — 正常動作する削除フックのパターン

## 優先度

**Medium** — ユーザーが削除後に手動でページを更新するまで操作の成否が不明になるが、データの整合性には影響しない可能性がある（削除が実際に成功しているかどうかが不確実）。

## 関連チケット
- BUG-328: 検査編集フォームで Select が「選択してください」を表示（同一 Section 5）

## 関連ファイル
- `frontend/src/features/examinations/api/delete-examination.ts:5-21` — 修正対象
- `frontend/src/features/examinations/hooks/use-examination-form.ts:150-160` — 要確認
- `frontend/src/features/examinations/routes/ExaminationForm.tsx:334-338` — 要確認
- `frontend/src/lib/axios.ts:51-57` — baseURL 設定
- `backend/internal/handler/examination_handler.go:191-205` — バックエンド実装（正常）
