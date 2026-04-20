# CODE-QUALITY-225: Repository Count メソッド命名パターン不統一

## 概要

マスタ削除前の依存チェックに使用する Count メソッドの命名パターンが
`CountUsageBy*` / `CountRecordsBy*` / `CountBy*` / `CountStaff*` と混在しており、
同じ用途のメソッドが異なる命名になっている。

## 不統一の実態

| メソッド名 | ファイル | 用途 |
|-----------|---------|------|
| `CountUsageByCheckupTypeID` | checkup_type_repository.go | 削除前の参照チェック |
| `CountUsageByExamTypeID` | exam_type_repository.go | 削除前の参照チェック |
| `CountUsageByMedicineID` | medicine_repository.go | 削除前の参照チェック |
| `CountUsageByProcedureID` | procedure_repository.go | 削除前の参照チェック |
| `CountUsageByVaccineID` | vaccine_repository.go | 削除前の参照チェック |
| `CountUsageByInquiryTemplateID` | inquiry_template_repository.go | 削除前の参照チェック |
| `CountUsageByReservationTypeID` | reservation_type_repository.go | 削除前の参照チェック |
| `CountUsageByChiefComplaintTypeID` | chief_complaint_repository.go | 削除前の参照チェック |
| `CountRecordsByCourseID` | trimming_course_repository.go | 削除前の参照チェック ← 命名が異なる |
| `CountRecordsByOptionID` | trimming_option_repository.go | 削除前の参照チェック ← 命名が異なる |
| `CountRecordsByCageID` | cage_repository.go | 削除前の参照チェック ← 命名が異なる |
| `CountStaffsByGroupID` | permission_group_repository.go | 削除前の参照チェック ← 命名が異なる |
| `CountStaffsByOccupationID` | occupation_repository.go | 削除前の参照チェック ← 命名が異なる |
| `CountPetsByInsuranceID` | insurance_repository.go | 削除前の参照チェック ← 命名が異なる |
| `CountByAnimalSpeciesID` | pet_repository.go | 削除前の参照チェック ← 命名が異なる |
| `CountByChiefComplaintTypeID` | inquiry_repository.go（内部） | 削除前の参照チェック ← 命名が異なる |

また「親子関係チェック」も命名が混在:
| `CountChildrenByParentID` | exam_type, checkup_type, medicine 等 | 子マスタ存在確認 |
| `CountNamesByCategoryID` | diagnosis_repository.go | 子マスタ存在確認 ← 命名が異なる |

## 推奨命名規則

| 用途 | 推奨命名パターン | 例 |
|------|----------------|-----|
| 削除前の参照チェック（業務レコード） | `CountUsageByXxxID` | `CountUsageByMedicineID` |
| 子マスタの存在チェック（親子関係） | `CountChildrenByParentID` | `CountChildrenByParentID` |
| 子マスタの存在チェック（意味的に「子」ではない場合） | `CountUsageByXxxID` も可 | — |

## 修正対象

`CountUsageBy*` に統一すべきメソッド:

| 現在の名前 | 修正後 | ファイル |
|----------|--------|---------|
| `CountRecordsByCourseID` | `CountUsageByCourseID` | trimming_course_repository.go |
| `CountRecordsByOptionID` | `CountUsageByOptionID` | trimming_option_repository.go |
| `CountRecordsByCageID` | `CountUsageByCageID` | cage_repository.go |
| `CountStaffsByGroupID` | `CountUsageByGroupID` | permission_group_repository.go |
| `CountStaffsByOccupationID` | `CountUsageByOccupationID` | occupation_repository.go |
| `CountPetsByInsuranceID` | `CountUsageByInsuranceID` | insurance_repository.go |
| `CountByAnimalSpeciesID` | `CountUsageByAnimalSpeciesID` | pet_repository.go |
| `CountNamesByCategoryID` | `CountChildrenByParentID` | diagnosis_repository.go |

## 修正時の注意

メソッド名変更は:
1. interface 定義（`type XxxRepository interface`）
2. 実装（`func (r *xxxRepository) Count...`）
3. service 側の呼び出し箇所
4. テストのモック定義・呼び出し箇所

の4箇所を同時に変更すること。

## 優先度

MEDIUM — 機能上の問題はないが、新規開発者がパターンを参照する際に
「どのメソッド名を使えばいいか」迷いが生じる。コードベース全体で統一すべき。
