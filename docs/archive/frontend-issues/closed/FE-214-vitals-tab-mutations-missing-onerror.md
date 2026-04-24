# FE-214: VitalsTab の create/update/delete mutation に onError がない

## 概要

`VitalsTab.tsx` の useMutation 呼び出し（作成・更新・削除）に `onError` コールバックが設定されていない。
API エラー発生時にユーザーへの通知が行われず、エラーが握り潰される。

## 問題コード

### `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx`

```tsx
// Line 347-353: createMutation — onError なし
createMutation.mutate(payload, {
  onSuccess: () => {
    // ...成功処理
  },
  // onError なし → バイタル追加失敗がサイレントに握り潰される
});

// Line 364-372: updateMutation — onError なし
updateMutation.mutate(payload, {
  onSuccess: () => {
    // ...成功処理
  },
  // onError なし → バイタル更新失敗がサイレントに握り潰される
});

// Line 383-388: deleteMutation — onError なし
deleteMutation.mutate(vitalId, {
  onSuccess: () => {
    // ...成功処理
  },
  // onError なし → バイタル削除失敗がサイレントに握り潰される
});
```

## 修正方針

各 `mutate()` 呼び出しに `onError` を追加する。

```tsx
// After: 各 mutate に onError 追加
createMutation.mutate(payload, {
  onSuccess: () => { ... },
  onError: (error) => handleApiError(error, "バイタル追加"),
});

updateMutation.mutate(payload, {
  onSuccess: () => { ... },
  onError: (error) => handleApiError(error, "バイタル更新"),
});

deleteMutation.mutate(vitalId, {
  onSuccess: () => { ... },
  onError: (error) => handleApiError(error, "バイタル削除"),
});
```

あるいは useMutation 定義側の `onError` に移動してもよい（一元化推奨）。

## 追加: デザイントークン違反

同ファイル `VitalsTab.tsx` 内に以下の違反も存在する：

- **Line 195, 560**: `bg-gray-50 hover:bg-gray-100`（体重単位トグルボタン）
  → `C.bgLight`, `C.hoverBgLight` に変更

## 影響範囲

| 対象 | 問題 | 状態 |
|------|------|------|
| `VitalsTab.tsx:347-353` | createMutation に onError なし | 要修正 |
| `VitalsTab.tsx:364-372` | updateMutation に onError なし | 要修正 |
| `VitalsTab.tsx:383-388` | deleteMutation に onError なし | 要修正 |
| `VitalsTab.tsx:195,560` | bg-gray-50, hover:bg-gray-100 デザイントークン違反 | 要修正 |

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックおよび `onError` コールバックで `handleApiError(error, "コンテキスト")` を呼び出す。

### プロジェクト内参照実装
- `frontend/src/features/vaccinations/api/create-vaccination.ts` — `onError: (error) => handleApiError(error, "...")` で正しく実装

## 優先度
**High** — バイタル追加・更新・削除の失敗がユーザーに通知されない。

## 関連ファイル
- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx`
- `frontend/src/lib/handle-api-error.ts`
