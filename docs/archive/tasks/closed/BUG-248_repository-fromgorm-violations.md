# BUG-248: Repository 層で `apperrors.FromGORM` 未使用（15+リポジトリ）

## 概要

プロジェクト規約では Repository 層の GORM エラーは `apperrors.FromGORM(err, "resource", id)` で変換することが必須。
しかし多数のリポジトリが `apperrors.Wrap(err, "message")` を使用しており、
`gorm.ErrRecordNotFound` が適切に 404 に変換されず 500 になるなど、HTTP ステータスの誤変換が発生する。

## 影響範囲

| リポジトリファイル | 行 | メソッド |
|-------------------|-----|---------|
| `accounting_repository.go` | 51,55,75,115,137 | FindAll(Count/Find/Scan), Create, Update後のFirst |
| `refund_repository.go` | 31,42,54 | Create, FindByBillingID, SumByBillingID |
| `billing_item_repository.go` | 49 | FindByBillingID |
| `cage_repository.go` | 36,55,77,92,100 | FindAll, Create, Delete, Reorder |
| `reservation_repository.go` | 54,58,83,94,105,119,130,146,158 | FindAll, Create, UpdateFields, Delete 等 |
| `reservation_schedule_repository.go` | 45,57,69,96,102,115,121,128,141 | FindByMonth, FindBreaks 等 |
| `reservation_course_repository.go` | 38 | FindAll |
| `reservation_admin_repository.go` | 52,66,105 | FindByMonth, FindByDay, FindByCustomerID |
| `reservation_customer_repository.go` | 36 | FindAll |
| `vital_repository.go` | 37,56,67,80 | ListByMedicalRecordID, Create, Update, Delete |
| `medical_record_repository.go` | 118 | CountByPetID |
| `owner_repository.go` | 51,57,78,86,141 | FindAll, FindByEmail, FindByPhone, Update |
| `pet_repository.go` | 52,53 | FindAll (Count/Find) |
| `animal_species_repository.go` | 37 | FindAll |
| `vaccine_repository.go` | 37,67 | FindAll, UpdateFields |
| `vaccination_repository.go` | 50,92 | FindAll, UpdateFields |
| `procedure_repository.go` | 33 | FindAll |
| `service_type_repository.go` | 36,56 | FindAll, Create |
| `merchandise_item_repository.go` | (複数) | FindAll, Create, CountUsage |
| `inquiry_repository.go` | 39,61,88 | UpsertByMedicalRecordID, CountByChiefComplaintCategoryID |

## 修正方針

全箇所で `apperrors.Wrap(err, "...")` → `apperrors.FromGORM(err, "resource", "id")` に機械的に置換する。

### 修正パターン

```go
// 修正前（単一レコード取得）
return nil, apperrors.Wrap(err, "find vital by id")

// 修正後
return nil, apperrors.FromGORM(err, "vital", fmt.Sprintf("%d", id))

// 修正前（リスト取得 — id 不要）
return nil, apperrors.Wrap(err, "find cages")

// 修正後
return nil, apperrors.FromGORM(err, "cage", "")
```

### 注意事項
- `Find`（スライス取得）は GORM が `ErrRecordNotFound` を返さない仕様だが、統一性のため `FromGORM` を使用する
- `Count` / `Scan` も同様に `FromGORM` に統一する

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — エラー処理の統一 (MANDATORY)
> **Repository**: GORM エラーは必ず `apperrors.FromGORM(err, "resource", id)` で変換。

### `.claude/rules/error-handling.md` — Go エラーフロー
> repository/owner_repository.go — `apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))` を使用

## 優先度
**High** — HTTP ステータスコードの誤変換（404 が 500 になる等）により、フロントエンドのエラーハンドリングが正しく動作しない。

## 関連チケット
- BUG-244: バックエンド Go コード規約準拠監査（親チケット）
