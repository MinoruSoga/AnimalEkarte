# BE-032: サービス層テスト未作成（21 サービス）

## 問題
42 サービスのうち 21 件（50%）にテストファイルが存在しない。

## テスト未作成サービス一覧
1. `animal_species_service.go`
2. `billing_review_service.go`
3. `care_plan_item_service.go`
4. `checkup_service.go`
5. `chief_complaint_category_service.go`
6. `clinical_plan_service.go`
7. `company_service.go`
8. `consultation_service.go`
9. `daily_record_service.go`
10. `estimate_service.go`
11. `hospitalization_plan_service.go`
12. `inquiry_template_service.go`
13. `job_title_service.go`
14. `procedure_service.go`
15. `record_image_service.go`
16. `shift_entry_service.go`
17. `shift_service.go`
18. `treatment_plan_service.go`
19. `treatment_service.go`
20. `user_account_service.go`
21. `vital_service.go`

## テスト済みサービス（参照実装）
- `cage_service_test.go` — CRUD + バリデーション + エラーケース網羅
- `staff_service_test.go` — パスワードハッシュ・更新フィールド選択テスト
- `owner_service_test.go` — ペット連携ロジックのテスト

## 修正方針
1. ビジネスロジックが複雑なものから優先（treatment, clinical_plan, daily_record）
2. 既存の `cage_service_test.go` パターンに従い mock repository を使用
3. 正常系・エラー系・エッジケースを各サービスに追加

## 優先度
HIGH（リグレッション防止）
