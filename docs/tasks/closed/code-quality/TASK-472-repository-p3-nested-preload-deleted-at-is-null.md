---
title: Repository P3 Nested Preload deleted_at IS NULL 欠落
issue: #472
priority: HIGH
status: open
area: repository
pattern: P3
---

## 概要

リポジトリの Preload 時に、ネストされた関連エンティティの soft-delete 条件（`deleted_at IS NULL`）が欠落しているケースが 4 件検出されました。

### パターン
- **P3 違反**：Preload 第2引数に `"deleted_at IS NULL"` がない、または ネストされた関連エンティティが条件対象外

### 違反ファイル一覧

| ファイル | 行番号 | 現在の Preload | 問題 |
|---------|--------|--------------|------|
| examination_repository.go | 58 | `Preload("Pet.Owner", "deleted_at IS NULL")` | Owner（soft-delete）が nested に含まれるが、条件が Pet の直下のみ |
| reservation_repository.go | 107 | `Preload("Pet.AnimalSpecies")` | AnimalSpecies は soft-delete entity だが条件なし |
| owner_repository.go | 72 | `Preload("Pets.Insurance", "deleted_at IS NULL")` | Insurance（soft-delete）が nested に含まれるが、条件が Pets のみ |
| accounting_repository.go | 141 | `Preload("Refunds.RefundedByStaff", "deleted_at IS NULL")` | RefundedByStaff（Staff = soft-delete）が nested に含まれるが、条件が Refunds のみ |

## 修正方法

GORM の nested Preload では以下のいずれかのパターンで対応：

1. **個別 Preload を分ける**（推奨）
   ```go
   db.Preload("Pet", "deleted_at IS NULL").
      Preload("Pet.Owner", "deleted_at IS NULL").
      Find(...)
   ```

2. **Join + condition（複雑な場合）**
   ```go
   db.Joins("JOIN pets ON ...").
      Joins("JOIN owners ON ... AND owners.deleted_at IS NULL").
      Find(...)
   ```

## テスト

修正後、以下の確認を実施：
- [ ] ロジカルデリート済みエンティティが Preload に含まれないこと
- [ ] 既存データの取得が正常に動作すること
- [ ] リポジトリテストが全件パス

## 参考

- Pattern: P3 (Preload deleted_at IS NULL)
- 関連タスク: TASK-468 (P2 CountUsage)、TASK-469 (P4 Upsert clinicScope)
