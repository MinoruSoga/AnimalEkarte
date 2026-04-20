# CODE-QUALITY-234: service テスト — Update nil input ケース欠落（複数ドメイン）

## 概要

複数のサービステストで、`Update` に `input == nil` を渡した場合のテストケースが欠落している。
サービス実装では `ErrMsgInputNotNil` をチェックしているが、
テストで検証されていないためリグレッションを検知できない。

---

## 対象ファイルと欠落ケース

| ファイル | 対象メソッド | 欠落ケース |
|---------|------------|-----------|
| `backend/internal/service/exam_type_service_test.go:255-324` | `TestExamTypeService_Update` | `{name: "nil input → error", input: nil, wantErr: true}` |
| `backend/internal/service/chief_complaint_service_test.go:232-310` | `TestChiefComplaintTypeService_Update` | `{name: "nil input → error", input: nil, wantErr: true}` |
| `backend/internal/service/trimming_course_service_test.go` | `TestTrimmingCourseService_Update` | `{name: "nil input → error", input: nil, wantErr: true}` |
| `backend/internal/service/trimming_option_service_test.go` | `TestTrimmingOptionService_Update` | `{name: "nil input → error", input: nil, wantErr: true}` |
| `backend/internal/service/cage_service_test.go` | `TestCageService_Update` | `{name: "nil input → error", input: nil, wantErr: true}` |
| `backend/internal/service/medicine_service_test.go` | `TestMedicineService_Update` | `{name: "nil input → error", input: nil, wantErr: true}` |
| `backend/internal/service/procedure_service_test.go` | `TestProcedureService_Update` | `{name: "nil input → error", input: nil, wantErr: true}` |
| `backend/internal/service/insurance_service_test.go:263` | `TestInsuranceService_Update` | `{name: "nil input → error", input: nil, wantErr: true}` |
| `backend/internal/service/permission_group_service_test.go:159` | `TestPermissionGroupService_Update` | `{name: "nil input → error", input: nil, wantErr: true}` |
| `backend/internal/service/checkup_type_service_test.go:174-243` | `TestCheckupTypeService_Create` | `{name: "empty name → error", input: &Create...{Name: ""}, wantErr: true}` |

---

## テストケース例（medicine_service を参考に）

```go
// medicine_service_test.go パターン（正しい実装）から
{
    name: "nil input returns error",
    setup: func(m *mockMedicineRepository) {
        // FindByID を呼ばないため mock 不要
    },
    clinicID: 1,
    id:       1,
    input:    nil,
    wantErr:  true,
},
```

---

## 注意事項

- `chief_complaint_service.go` と `insurance_service.go` は `input == nil` チェック自体が欠落（CODE-QUALITY-233）。
  そのドメインでは実装修正後にテストを追加する。
- その他のドメインはサービス実装は正しく、テストが欠けているのみ。

---

## 優先度

MEDIUM — 実装バグではなくテストカバレッジ欠落。
`input == nil` チェックのリグレッションを検知できないリスク。
