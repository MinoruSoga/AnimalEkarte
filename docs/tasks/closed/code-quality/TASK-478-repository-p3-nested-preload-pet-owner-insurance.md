---
title: Repository P3 Nested Preload Pet.Owner/Insurance/AnimalSpecies deleted_at IS NULL 欠落
issue: '#478'
priority: HIGH
status: open
area: repository
pattern: P3
---

## 概要

リポジトリの Preload 時に、ネストされた関連エンティティの soft-delete 条件（`deleted_at IS NULL`）が欠落しているケースが追加で 8 件検出されました。TASK-472 に続く P3 違反です。

### パターン
- **P3 違反**：Preload の第2引数に `"deleted_at IS NULL"` がない、または nested エンティティに条件が適用されていない

### 違反ファイル一覧

| ファイル | 行番号 | 現在の Preload | 問題 |
|---------|--------|--------------|------|
| examination_repository.go | 58, 70 | `Preload("Pet.Owner", "deleted_at IS NULL")` | Owner（soft-delete）の nested 条件が不完全 |
| hospitalization_repository.go | 56, 68 | `Preload("Pet.AnimalSpecies", "deleted_at IS NULL")` | AnimalSpecies nested 条件の確認必要 |
| medical_record_repository.go | 52, 68 | `Preload("Pet.AnimalSpecies", "deleted_at IS NULL")` | AnimalSpecies nested 条件の確認必要 |
| owner_repository.go | 62, 72 | `Preload("Pets.Insurance", "deleted_at IS NULL")` | Insurance（soft-delete）の nested 条件が不完全 |
| reservation_repository.go | 94, 106, 328 | `Preload("Pet.Owner")` と `Preload("Pet.Owner", "deleted_at IS NULL")` | Owner nested に条件が完全でない（3 メソッド） |
| pet_repository.go | 62, 73 | `Preload("AnimalSpecies")` | AnimalSpecies の deleted_at 条件が不明確 |
| accounting_repository.go | 101, 141, 186 | `Preload("Payments.PaidByStaff", "deleted_at IS NULL")` | Staff nested（PaidByStaff）の条件が不完全（3 メソッド） |
| reservation_staff_repository.go | 172, 188 | `Preload("ReservationType", "deleted_at IS NULL")` | ReservationType nested verification 必要 |

## 修正方法

GORM の nested Preload では以下のいずれかのパターンで対応：

1. **個別 Preload を分ける（推奨）**
   ```go
   // examination_repository.go (L58 例)
   db.Preload("Pet", "deleted_at IS NULL").
      Preload("Pet.Owner", "deleted_at IS NULL").
      Find(...)
   
   // owner_repository.go (L62 例)
   db.Preload("Pets", "deleted_at IS NULL").
      Preload("Pets.Insurance", "deleted_at IS NULL").
      Find(...)
   
   // accounting_repository.go (L101 例)
   db.Preload("Payments", "deleted_at IS NULL").
      Preload("Payments.PaidByStaff", "deleted_at IS NULL").
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
- [ ] ロジカルデリート済みエンティティが nested Preload に含まれないこと
- [ ] 既存データの取得が正常に動作すること
- [ ] リポジトリテストが全件パス
- [ ] 特に Pet.Owner, Pets.Insurance, Payments.PaidByStaff の nested chain を確認

## 参考

- Pattern: P3 (Preload deleted_at IS NULL - Nested)
- 関連タスク: TASK-472 (P3 base violations)
