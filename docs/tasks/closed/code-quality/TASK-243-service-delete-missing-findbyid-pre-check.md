# TASK-243: service — Delete メソッドに FindByID 事前確認が欠落（7ドメイン横断）

## 優先度
High

## 対象ファイル
- `backend/internal/service/checkup_type_service.go`
- `backend/internal/service/diagnosis_service.go`（DiagnosisTypeService + DiagnosisNameService）
- `backend/internal/service/vaccine_service.go`
- `backend/internal/service/procedure_service.go`
- `backend/internal/service/cage_service.go`
- `backend/internal/service/permission_group_service.go`
- `backend/internal/service/payment_method_master_service.go`

## 問題概要
上記7ドメインの `Delete` メソッドが FK 依存チェック（`CountUsageBy*`）を行う前に
`FindByID` による存在確認を行っていない。

存在しない id に対して Delete リクエストを送ると、FK 依存カウントが 0 を返し、
本来 404 を返すべきところに実際には 404 が返らず正常削除が実行されるか、
または repository の Delete 呼び出し時点で初めて「レコードなし」が判明する。

規約: **Delete は `FindByID` → FK 依存チェック → `repo.Delete` の順で実行する。**

## 現状コード（代表: checkup_type_service.go）

```go
func (s *checkupTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
    childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)  // ❌ FindByID なし
    if err != nil {
        return apperrors.Wrap(err, "failed to count child checkup types")
    }
    if childCount > 0 {
        return apperrors.WrapConflict("この検査区分は子項目を持つため削除できません")
    }
    // ...
}
```

## 正しい参照実装（chief_complaint_service.go）

```go
func (s *chiefComplaintTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {  // ✅ FindByID 先行
        return apperrors.Wrap(err, "failed to get chief complaint type")
    }
    count, err := s.repo.CountUsageByChiefComplaintTypeID(ctx, clinicID, id)
    // ...
}
```

## 修正対象

| ファイル | メソッド | FK チェック前に FindByID を追加 |
|---------|---------|-------------------------------|
| `checkup_type_service.go` | `Delete` | CountChildrenByParentID の前 |
| `diagnosis_service.go` | `DiagnosisTypeService.Delete` | CountChildrenByParentID の前 |
| `diagnosis_service.go` | `DiagnosisNameService.Delete` | CountClinicalPlansByDiagnosisNameID の前 |
| `vaccine_service.go` | `Delete` | FK チェックの前 |
| `procedure_service.go` | `Delete` | FK チェックの前 |
| `cage_service.go` | `Delete` | CountUsageByCageID の前 |
| `permission_group_service.go` | `Delete` | CountUsageByGroupID の前 |
| `payment_method_master_service.go` | `Delete` | CountUsageByPaymentMethodID の前 |

## 完了条件
- [ ] 上記8メソッドすべての先頭に `FindByID` 呼び出しを追加
- [ ] FindByID が失敗した場合 `apperrors.Wrap(err, "failed to get {entity}")` を返す
- [ ] `go test ./backend/internal/...` がパス
