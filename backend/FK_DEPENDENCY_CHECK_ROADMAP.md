# FK Dependency Check Implementation Roadmap

## Overview
本プロジェクトではマスタデータ削除時に外部キー依存性をチェックし、参照レコードがある場合は 409 Conflict で拒否する仕組みを実装します。

## 実装パターン

### 1. Repository 層
- CountUsageByXxxID メソッドを定義（インターフェース）
- 参照元テーブルをJOINで合計件数をカウント
- 論理削除（deleted_at IS NULL）を考慮

```go
func (r *xxxRepository) CountUsageByXxxID(ctx context.Context, xxxID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.ReferencingTable{}).
        Where("xxx_id = ? AND deleted_at IS NULL", xxxID).
        Count(&count).Error
    if err != nil {
        return 0, apperrors.Wrap(err, "count xxx usage")
    }
    return count, nil
}
```

### 2. Service 層
- Delete メソッドで削除前に依存チェック
- 参照がある場合は apperrors.WrapConflict() で 409 返却

```go
func (s *xxxService) Delete(ctx context.Context, clinicID, id uint64) error {
    count, err := s.repo.CountUsageByXxxID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "check xxx dependencies")
    }
    if count > 0 {
        return apperrors.WrapConflict("このアイテムは使用中のため削除できません")
    }
    return s.repo.Delete(ctx, clinicID, id)
}
```

### 3. Handler 層
- 既存の RespondError() で統一レスポンス → 409 ステータス自動応答

---

## 実装対象（優先度順）

### TIER 1 - HIGH（本日実装推奨）
| # | Entity | 参照元テーブル | Repository メソッド | 状態 |
|----|--------|-----------|------------|------|
| 1 | MerchandiseItem | billing_items, estimate_items | CountUsageByMerchandiseItemID | ✅ リポジトリメソッド実装済み |
| 2 | Vaccine | vaccinations | CountUsageByVaccineID | ⏳ 実装待ち |
| 3 | Medicine | treatments | CountUsageByMedicineID | ⏳ 実装待ち |

### TIER 2 - MEDIUM（来週実装推奨）
| # | Entity | 参照元テーブル | Repository メソッド |
|----|--------|-----------|------------|
| 4 | Consultation | treatments | CountUsageByConsultationID |
| 5 | Procedure | treatments | CountUsageByProcedureID |
| 6 | ExamType | exams | CountUsageByExamTypeID |
| 7 | CheckupType | checkups | CountUsageByCheckupTypeID |
| 8 | JobTitle | staff | CountUsageByJobTitleID |
| 9 | PermissionGroup | user_permission_groups | CountUsageByPermissionGroupID |
| 10 | Cage | animals | CountUsageByCanvaseID |
| 11 | DiagnosisCategory | diagnoses | CountUsageByDiagnosisCategoryID |
| 12 | ClinicalPlan | (なし - 削除制限なし) | (不要) |

### TIER 3 - LOW（オプション・ドメインロジック検証後）
- Accounting (削除制限なし)
- BillingItem (削除制限なし)
- CarePlanItem (削除制限なし)
- Clinic (削除制限なし)
- EstimateItem (削除制限なし)
- HospitalizationPlan (削除制限なし)
- Inventory (削除制限なし)

---

## 実装チェックリスト（TIER 1）

### MerchandiseItem
- [x] Repository: CountUsageByMerchandiseItemID 実装済み
- [ ] Service: Delete に FK チェック追加
- [ ] テスト: merchandise_item_repository_test.go を確認・更新
- [ ] 統合テスト: DELETE /api/merchandise-items/:id → 409 応答確認

### Vaccine
- [ ] Repository: CountUsageByVaccineID メソッド追加
- [ ] インターフェース定義（VaccineRepository）
- [ ] Service: Delete に FK チェック追加
- [ ] テスト: vaccine_repository_test.go を作成
- [ ] 統合テスト: DELETE /api/vaccines/:id → 409 応答確認

### Medicine
- [ ] Repository: CountUsageByMedicineID メソッド追加
- [ ] インターフェース定義（MedicineRepository）
- [ ] Service: Delete に FK チェック追加
- [ ] テスト: medicine_repository_test.go を作成
- [ ] 統合テスト: DELETE /api/medicines/:id → 409 応答確認

---

## テスト戦略

### ユニットテスト（Repository）
```go
func TestMedicineRepository_CountUsageByMedicineID(t *testing.T) {
    tests := []struct {
        name       string
        medicineID uint64
        want       int64
        wantErr    bool
    }{
        {"no usage", 999, 0, false},
        {"used in 2 treatments", 1, 2, false},
    }
    // ...
}
```

### ユニットテスト（Service）
```go
func TestMedicineService_Delete_WithUsage(t *testing.T) {
    // Mock: repo.CountUsageByMedicineID → 2 を返す
    // Assert: Service.Delete → ErrConflict を返す
}
```

### 統合テスト（Handler）
```bash
curl -X DELETE /api/medicines/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Clinic-ID: 1"

# Expected: 409 Conflict
# {
#   "error": "このアイテムは使用中のため削除できません"
# }
```

---

## 実装時の注意点

### 1. 論理削除対応
すべての COUNT クエリで `WHERE deleted_at IS NULL` を含める。

### 2. マルチテナント対応
clinic_id フィルタを含める（必要な場合）。

```go
// ✅ Clinic ID フィルタ
WHERE xxx.clinic_id = ? AND xxx_id = ? AND deleted_at IS NULL

// ❌ 非テナント対応
WHERE xxx_id = ?
```

### 3. エラーメッセージ
ユーザーフレンドリーで、参照先テーブルを明記。

```
❌ "cannot delete record"
✅ "このアイテムは◯◯で使用中のため削除できません"
```

### 4. パフォーマンス
COUNT クエリは複合インデックスで最適化。

```sql
CREATE INDEX idx_treatments_medicine_id
  ON treatments(clinic_id, medicine_id)
  WHERE deleted_at IS NULL;
```

---

## 進捗トラッキング

実装完了時に以下を更新：
1. `CRITICAL-BUGS-TRACKING.md` - 41 TODO → X件実装完了
2. 各Repository test ファイル - CountUsageBy テスト追加
3. 本 Roadmap - 実装状況を DONE に更新

---

**Last Updated**: 2026-04-04 15:25 JST
**Owner**: Claude Haiku 4.5
**Status**: In Progress (TIER 1 ready for implementation)
