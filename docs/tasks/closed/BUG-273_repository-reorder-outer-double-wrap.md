# BUG-273: Repository Reorder/Transaction 外側二重ラップ

## 概要

Reorder / SetRules / SetStaffGroups メソッドのトランザクション外側で `apperrors.Wrap(err, "...")` が使われており、
内側で既に `FromGORM` / `WrapInvalidInput` 済みのエラーを二重ラップしている。
`errors.Is()` チェーンが壊れるリスクがあるため `return err` に統一すべき。

## 影響範囲

| ファイル | 行 | 現状コード |
|----------|-----|----------|
| `checkup_type_repository.go` | 102 | `return apperrors.Wrap(err, "reorder checkup type")` |
| `consultation_repository.go` | 102 | `return apperrors.Wrap(err, "reorder consultation")` |
| `diagnosis_repository.go` | 130 | `return apperrors.Wrap(err, "reorder diagnosis category")` |
| `diagnosis_repository.go` | 273 | `return apperrors.Wrap(err, "reorder diagnosis name")` |
| `exam_type_repository.go` | 102 | `return apperrors.Wrap(err, "reorder examination type")` |
| `insurance_repository.go` | 112 | `return apperrors.Wrap(err, "reorder insurance")` |
| `medicine_repository.go` | 139 | `return apperrors.Wrap(err, "reorder medicine")` |
| `occupation_repository.go` | 117 | `return apperrors.Wrap(err, "reorder occupation")` |
| `permission_group_repository.go` | 126 | `return apperrors.Wrap(err, "failed to set permission group rules")` |
| `permission_group_repository.go` | 213 | `return apperrors.Wrap(err, "failed to set staff permission groups")` |
| `permission_group_repository.go` | 235 | `return apperrors.Wrap(err, "reorder permission group")` |

**合計: 11箇所 / 8ファイル**

## 修正方針

全箇所を `return err` に変更。内側で既に適切なエラー型が設定されているため、外側で再ラップは不要。

## 優先度

**Medium** — errors.Is() チェーンが壊れるリスクはあるが、現状の呼び出し元では直接的な影響は軽微。

## 関連チケット

- BUG-255: Repository Reorder 内側 FromGORM（第2回監査で内側を修正済み）
- BUG-270: 第4回監査 親チケット
