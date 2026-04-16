# BUG-349: BulkUpdateTreatments がクリニック境界チェックなしで他クリニックのデータを書き換え可能

## 概要

`PUT /medical-records/:id/treatments` エンドポイントの `BulkUpdateSortOrder` 処理で、
`clinicID` によるテナント境界チェックが欠落している。
認証済みの任意クリニックユーザーが、別クリニックの treatment の `sort_order` を書き換えることができる。

## 影響範囲

- **エンドポイント**: `PUT /api/clinics/:clinicId/medical-records/:id/treatments`
- **攻撃条件**: 有効な JWT を持つ認証済みユーザーであれば、treatment ID さえわかれば任意の treatment を書き換え可能
- **影響**: データ完全性の侵害（他クリニックの診療記録内の処置順序改ざん）

## 根本原因

ハンドラが `clinicID` を明示的に破棄し、サービス・リポジトリ全層でテナント検証が行われていない。

```go
// backend/internal/handler/treatment_handler.go:177（バグ箇所）
_ = clinicID // BulkUpdateSortOrder does not need clinic-scoped queries per item
// ↑ コメントは間違い。clinic スコープは必要。
```

```go
// backend/internal/service/treatment_service.go:253-271（clinicID を受け取らない）
func (s *treatmentService) BulkUpdateSortOrder(ctx context.Context, medicalRecordID uint64, input *BulkUpdateTreatmentsInput) error {
    // medicalRecordID が clinicID に属するかチェックなし
    // 直接 repo.BulkUpdateSortOrder に委譲
}
```

```go
// backend/internal/repository/treatment_repository.go:101-115（id のみで UPDATE）
result := tx.Model(&model.Treatment{}).
    Where("id = ? AND deleted_at IS NULL", u.ID). // clinic_id チェックなし
    Update("sort_order", u.SortOrder)
```

## 修正方針

**方法 A（推奨）: サービス層で medicalRecordID の clinic 所属を検証**

```go
// service/treatment_service.go
func (s *treatmentService) BulkUpdateSortOrder(ctx context.Context, clinicID, medicalRecordID uint64, input *BulkUpdateTreatmentsInput) error {
    // medicalRecordID が clinicID に属するか確認
    if _, err := s.repos.MedicalRecord.FindByID(ctx, clinicID, medicalRecordID); err != nil {
        return apperrors.Wrap(err, "medical record not found")
    }
    // 以降は変更なし
}
```

**方法 B: リポジトリ層で clinic 境界を適用**

`BulkUpdateSortOrder` のシグネチャに `clinicID` を追加し、サブクエリでテナント確認:
```go
func (r *treatmentRepository) BulkUpdateSortOrder(ctx context.Context, clinicID uint64, updates []TreatmentSortUpdate) error {
    // ...
    result := tx.Model(&model.Treatment{}).
        Where("id = ? AND deleted_at IS NULL"+
            " AND medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)",
            u.ID, clinicID).
        Update("sort_order", u.SortOrder)
}
```

方法 A の方が既存のパターン（サービス層で FindByID を事前に呼ぶ）と一貫性が高い。

## 修正が必要なファイル

1. `backend/internal/handler/treatment_handler.go:177` — `clinicID` を `BulkUpdateSortOrder` に渡す
2. `backend/internal/service/treatment_service.go:253` — シグネチャに `clinicID` 追加、事前 FindByID で検証
3. `backend/internal/service/treatment_service.go:74` — インターフェースにも `clinicID` を追加

## 優先度

**HIGH（セキュリティ）** — 認証済みユーザーによるクロステナントデータ操作が可能。sort_order のみの書き換えで影響は限定的だが、テナント分離の原則を侵害している。
