# BUG-201: バックエンド 複数 Service の Delete メソッドに FK 依存チェック欠如

## 概要

BUG-195 で `inventory_service.go` の FK チェック欠如を報告したが、監査の結果さらに 7 つのサービスの `Delete` メソッドでも同様に FK 依存チェックが欠落していることが判明した。`.claude/CLAUDE.md` の「マスタ削除の FK 依存チェック (MANDATORY)」ルールに違反しており、参照先が存在しない孤立レコードが生じる可能性がある。

## 脆弱性分類
- **影響**: 関連レコードが存在するデータを削除することで、診察記録・予約・処置計画等の参照が孤立し、アプリケーションの NULL 参照エラーや表示不整合を引き起こす。

## 再現手順

### 例: estimate_service（見積）
1. 見積に明細行（estimate_items）が紐づいた状態を作成
2. `DELETE /api/estimates/:id` を実行
3. **結果**: 409 Conflict ではなく 200 OK で削除される（estimate_items が孤立）

### 例: medical_record_service（診察記録）
1. 処置・処方・診断等が紐づいた診察記録を作成
2. `DELETE /api/medical-records/:id` を実行
3. **結果**: 関連レコードのカスケード削除が保証されない or FK 違反でサーバーエラー

### 例: vaccination_service（ワクチン）
1. 診察記録に紐づいたワクチン記録を作成
2. `DELETE /api/vaccinations/:id` を実行
3. **結果**: FK チェックなしで削除される

## 現状コード

### `backend/internal/service/estimate_service.go:119`
```go
// ❌ FK チェックなし
func (s *EstimateService) Delete(ctx context.Context, id uint64) error {
    return s.repo.Delete(ctx, id)
}
```

### `backend/internal/service/inquiry_template_service.go:73`
```go
// ❌ FK チェックなし
func (s *InquiryTemplateService) Delete(ctx context.Context, id uint64) error {
    return s.repo.Delete(ctx, id)
}
```

### `backend/internal/service/medical_record_service.go:136`
```go
// ❌ FK チェックなし
func (s *MedicalRecordService) Delete(ctx context.Context, id uint64) error {
    return s.repo.Delete(ctx, id)
}
```

### `backend/internal/service/reservation_service.go:130`
```go
// ❌ FK チェックなし
func (s *ReservationService) Delete(ctx context.Context, id uint64) error {
    return s.repo.Delete(ctx, id)
}
```

### `backend/internal/service/treatment_plan_service.go:102`
```go
// ❌ FK チェックなし
func (s *TreatmentPlanService) Delete(ctx context.Context, id uint64) error {
    return s.repo.Delete(ctx, id)
}
```

### `backend/internal/service/trimming_master_service.go:145`
```go
// ❌ FK チェックなし（trimmingOptionService.Delete 含む）
func (s *TrimmingMasterService) Delete(ctx context.Context, id uint64) error {
    return s.repo.Delete(ctx, id)
}
```

### `backend/internal/service/vaccination_service.go:123`
```go
// ❌ FK チェックなし
func (s *VaccinationService) Delete(ctx context.Context, id uint64) error {
    return s.repo.Delete(ctx, id)
}
```

### 比較: 正しい実装（参照実装）
```go
// ✅ MANDATORY パターン（CLAUDE.md より）
func (s *InventoryService) Delete(ctx context.Context, id uint64) error {
    count, err := s.repo.CountUsageByInventoryID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check usage")
    }
    if count > 0 {
        return apperrors.WrapConflict("この項目は使用中のため削除できません")
    }
    return s.repo.Delete(ctx, id)
}
```

## 影響範囲

| サービス | Delete メソッド行番号 | 依存リソース | 状態 |
|---|---|---|---|
| `estimate_service.go` | L119 | estimate_items | 未修正 |
| `inquiry_template_service.go` | L73 | 利用実績（问い合わせ） | 未修正 |
| `medical_record_service.go` | L136 | treatments, prescriptions, diagnoses, vaccinations 等 | 未修正 |
| `reservation_service.go` | L130 | billing_records, related notifications | 未修正 |
| `treatment_plan_service.go` | L102 | treatment_plan_items | 未修正 |
| `trimming_master_service.go` | L145 | trimming_records（過去記録） | 未修正 |
| `vaccination_service.go` | L123 | medical_record_vaccinations | 未修正 |

## 修正方針

各サービスについて以下のパターンを適用する。FK の依存関係は実際のスキーマ（`docs/ERD.md`）を参照して CountUsage 関数を正確に実装すること。

### 共通パターン
```go
// Step 1: repository に CountUsageBy[Entity]ID を追加
func (r *[Entity]Repository) CountUsageBy[Entity]ID(ctx context.Context, id uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&[DependentModel]{}).
        Where("[entity]_id = ? AND deleted_at IS NULL", id).
        Count(&count).Error
    return count, err
}

// Step 2: service の Delete に FK チェックを追加
func (s *[Entity]Service) Delete(ctx context.Context, id uint64) error {
    count, err := s.repo.CountUsageBy[Entity]ID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check [entity] usage")
    }
    if count > 0 {
        return apperrors.WrapConflict("この[エンティティ名]は使用中のため削除できません")
    }
    return s.repo.Delete(ctx, id)
}
```

### 優先度別修正順序
1. **medical_record_service** — 診察記録は多数の子レコードを持つ最重要
2. **reservation_service** — 予約は billing と紐づく
3. **vaccination_service** — 診察記録への FK がある
4. **estimate_service** — estimate_items が孤立
5. **treatment_plan_service** — treatment_plan_items が孤立
6. **trimming_master_service** — マスタ削除で過去記録が参照不能に
7. **inquiry_template_service** — 利用実績の確認

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — マスタ削除の FK 依存チェック (MANDATORY)
> マスタ削除時は必ず依存レコードの存在をチェックし、参照がある場合は `apperrors.WrapConflict(...)` で 409 を返す。

### プロジェクト内参照実装
- `backend/internal/service/animal_species_service.go` — `CountUsageByAnimalSpeciesID` の正しい実装パターン

## 優先度
**High** — `medical_record_service` と `reservation_service` の FK チェック欠如はデータ整合性に直結。診察記録の不正削除は患者の医療情報喪失に繋がる。

## 関連チケット
- BUG-195: inventory_service FK チェック欠如・pet/estimate バリデーション欠如（先行チケット）
- BUG-176/177: master repository の clinic_id フィルタ欠如

## 関連ファイル
- `backend/internal/service/estimate_service.go`
- `backend/internal/service/inquiry_template_service.go`
- `backend/internal/service/medical_record_service.go`
- `backend/internal/service/reservation_service.go`
- `backend/internal/service/treatment_plan_service.go`
- `backend/internal/service/trimming_master_service.go`
- `backend/internal/service/vaccination_service.go`
- `backend/internal/apperrors/errors.go`
