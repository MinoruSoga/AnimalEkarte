# FE-051: master ドメイン — 旧STI型の削除・整理

**親タスク**: [TASK-012](../../docs/tasks/open/TASK-012-models-ts-type-migration-by-domain.md)
**Status**: Open
**Priority**: Medium
**Date Created**: 2026-03-18

## Summary

master feature は12ファイルで models.ts を使用済みだが、旧 master_items STI 時代の `CreateMasterItemRequest`/`UpdateMasterItemRequest` が残存している。専用マスタテーブルに移行済みのため削除・整理が必要。

## 現状

```typescript
// features/master/api/get-master-items.ts
// 旧STI型: CreateMasterItemRequest, UpdateMasterItemRequest が残存
// 専用マスタ: Consultation, ExaminationType, Procedure 等は models.ts import 済み
```

## 必要な変更

1. `CreateMasterItemRequest` / `UpdateMasterItemRequest` を削除
2. 各専用マスタの Request 型が models.ts から導出されているか確認
3. 使用箇所があれば各専用マスタの型に置き換え

## 完了条件

- [ ] 旧STI型（CreateMasterItemRequest, UpdateMasterItemRequest）が削除されている
- [ ] 全専用マスタが models.ts から型を使用
- [ ] `npm run build` 成功・型エラーなし
