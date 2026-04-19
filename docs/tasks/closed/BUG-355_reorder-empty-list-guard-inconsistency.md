# BUG-355: 7 つの Reorder メソッドに `len(ids) == 0` ガードが欠落（一貫性違反）

## 概要

19 個の `Reorder` メソッドのうち 12 個は `len(ids) == 0` で `WrapInvalidInput` を返すガード節を持つが、
残り 7 個はガードなしでリポジトリに直接委譲している。空リストを渡すと無駄なトランザクションが発生する。

## ガードなしのメソッド（7件）

| ファイル | メソッド |
|---------|---------|
| `trimming_master_service.go:80` | `trimmingCourseService.Reorder` |
| `trimming_master_service.go:192` | `trimmingOptionService.Reorder` |
| `occupation_service.go:102` | `occupationService.Reorder` |
| `insurance_service.go:80` | `insuranceService.Reorder` |
| `hospitalization_plan_service.go:84` | `hospitalizationPlanService.Reorder` |
| `permission_group_service.go:115` | `permissionGroupService.Reorder` |
| `reservation_type_group_service.go:123` | `reservationTypeGroupService.Reorder` |

## ガードありのメソッド（12件・正しいパターン）

`merchandise_item`, `consultation`, `reservation_type`, `staff`, `procedure`, `shift_template`,
`cage`, `vaccine`, `animal_species`, `medicine`, `diagnosis_type`, `diagnosis_name`,
`checkup_type`, `exam_type`

## 修正内容

各メソッドの先頭に追加:
```go
if len(ids) == 0 {
    return apperrors.WrapInvalidInput("ids must not be empty")
}
```

## 優先度

**LOW** — 機能バグではなく一貫性の問題。空リストで無駄なトランザクションが発生するが、データ破損のリスクはない。
