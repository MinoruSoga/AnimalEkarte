# BUG-106: 削除失敗時のエラーメッセージが generic（期待するユーザー向けメッセージと異なる）

## 概要

FK制約チェックが修正されたサービス種別・スタッフの削除で 409 Conflict を返すようになったが、
フロントエンドのエラートーストが「〇〇の削除に失敗しました」という generic なメッセージになっている。
BUG-030 の期待する動作「このデータは他のレコードに使用されています」が表示されていない。

## 症状

- `/settings/service-type` で依存データありのサービス種別削除実行
  - BE: DELETE → 409 Conflict（正しい）
  - FE トースト: **「診療サービスの削除に失敗しました」**（generic）

- `/settings/staff` で依存データありのスタッフ削除実行
  - BE: DELETE → 409 Conflict（正しい）
  - FE トースト: **「スタッフの削除に失敗しました」**（generic）

## 期待する動作

409 Conflict の場合にユーザーがアクションを理解できるメッセージを表示：

> 「このデータは他のレコードに使用されているため削除できません」

## 根本原因

FE の削除 mutation/ハンドラが 409 と他のエラーを区別せず、どちらも同じ generic トーストを表示している。

```typescript
// 現状（例）
onError: () => {
  toast.error("スタッフの削除に失敗しました");  // ← 409 でも他のエラーでも同じ
}

// 修正後
onError: (error) => {
  if (isConflictError(error)) {
    toast.error("このデータは他のレコードに使用されているため削除できません");
  } else {
    toast.error("スタッフの削除に失敗しました");
  }
}
```

## 影響ファイル

- `frontend/src/features/master/api/delete-staff.ts`（または delete mutation hooks）
- `frontend/src/features/master/api/delete-service-type.ts`
- 削除 mutation を持つ全マスタ API ファイル

## 優先度

Medium（UX改善）

## 関連

- BUG-030（FK制約チェックの BE 修正は完了。FE 側のエラー表示が不十分）
- BUG-103, BUG-104, BUG-105（同系統の他マスタも同様に修正要）
- テスト確認日: 2026-04-01（ローカル環境）
