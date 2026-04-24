# TASK-030: checkup_repository.ListByMedicalRecordID に clinic_id フィルタなし — クロステナントリスク

## 概要

`checkup_repository.go` の `ListByMedicalRecordID` が `medical_record_id = ?` のみで検索し、`clinic_id` フィルタが存在しない。`medical_record_id` はシーケンシャル ID のため他クリニックとのレコードが衝突し得る。加えて `checkup_handler.go` では `extractClinicID` の結果を `_` で破棄しており、service にも clinic_id が渡っていない。

## 優先度

HIGH（マルチテナント分離漏れ・クロステナントデータ閲覧リスク）

## 影響ファイル

| ファイル | 行 | 問題 |
|---------|-----|------|
| `backend/internal/repository/checkup_repository.go` | L64-76 | `ListByMedicalRecordID` に clinic_id フィルタなし |
| `backend/internal/handler/checkup_handler.go` | L16-32 | `extractClinicID` の clinicID を `_` で破棄 |

## 現状コード

```go
// checkup_handler.go:16
_, ok := extractClinicID(c) // clinicID を破棄

// checkup_repository.go:64-76
err := r.db.WithContext(ctx).
    Where("medical_record_id = ?", medicalRecordID). // clinic_id なし
    Find(&checkups).Error
```

## 修正方針

### handler 層
```go
clinicID, ok := extractClinicID(c)
if !ok { return }
// clinicID を service に渡す
```

### repository 層
```go
// シグネチャ変更
ListByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)

// 実装: medical_records JOIN で clinic_id を確認
func (r *checkupRepository) ListByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
    var checkups []model.Checkup
    err := r.db.WithContext(ctx).
        Joins("JOIN medical_records ON medical_records.id = checkups.medical_record_id"+
            " AND medical_records.clinic_id = ?"+
            " AND medical_records.deleted_at IS NULL", clinicID).
        Where("checkups.medical_record_id = ?", medicalRecordID).
        Find(&checkups).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "checkup", "")
    }
    return checkups, nil
}
```

### service 層
```go
ListByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
```

## テスト

- クリニック A の medical_record_id を使って クリニック B の認証で取得を試みたとき、空リストが返ることを確認
