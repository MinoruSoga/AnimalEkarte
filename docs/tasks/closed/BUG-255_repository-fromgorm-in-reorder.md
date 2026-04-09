# BUG-255: Repository Reorder/トランザクション内で apperrors.Wrap → FromGORM に統一

## 概要

11個のリポジトリの `Reorder` メソッドおよびトランザクション内部で、GORM エラーに対して
`apperrors.Wrap` を使用している。Repository 層の GORM エラーは `apperrors.FromGORM` を
使用する規約に違反。

## 影響範囲

| ファイル | 行番号 | 現状 |
|---------|--------|------|
| `repository/vaccine_repository.go` | :93 | `apperrors.Wrap(result.Error, "reorder vaccine")` |
| `repository/cage_repository.go` | :93 | `apperrors.Wrap(result.Error, "reorder cage")` |
| `repository/procedure_repository.go` | :107 | `apperrors.Wrap(result.Error, "reorder procedure")` |
| `repository/trimming_master_repository.go` | :103, :197 | `apperrors.Wrap(result.Error, "reorder ...")` |
| `repository/merchandise_item_repository.go` | :105 | `apperrors.Wrap(result.Error, "reorder merchandise item")` |
| `repository/service_type_repository.go` | :102, :110 | `apperrors.Wrap(result.Error/err, "reorder service type")` |
| `repository/staff_repository.go` | :132, :140 | `apperrors.Wrap(result.Error/err, "reorder staff")` |
| `repository/animal_species_repository.go` | :109 | `apperrors.Wrap(result.Error, "reorder animal species")` |
| `repository/reservation_staff_repository.go` | :58, :66, :126, :129, :152, :166 | トランザクション内全箇所 |
| `repository/reservation_course_repository.go` | :113, :116 | SwapSortOrder トランザクション内 |
| `repository/reservation_schedule_repository.go` | :97, :103, :116, :122, :129 | Upsert トランザクション内全箇所 |

## 修正方針

各箇所を以下のパターンで統一:

```go
// Before
return apperrors.Wrap(result.Error, "reorder xxx")

// After
return apperrors.FromGORM(result.Error, "xxx", fmt.Sprintf("%d", id))
```

トランザクション外側のラップ（`return apperrors.Wrap(err, "reorder xxx")`）については、
トランザクション内部で既に `FromGORM` でラップ済みの場合は直接 `return err` で十分。

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> Repository: `apperrors.FromGORM(err, "resource", id)` 使用

## 優先度

**High** — エラーチェーン（`errors.Is`）の一貫性が失われ、ハンドラの HTTP ステータスマッピングが不正確になる。

## 関連チケット

- BUG-248: 同種の問題（第1回監査で修正済み）
- BUG-253: 親チケット
