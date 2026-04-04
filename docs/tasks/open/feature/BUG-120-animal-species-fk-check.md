# BUG-120: 動物種削除時のFK依存チェック未実装

**Priority**: HIGH
**Type**: Bug
**Status**: Open
**Discovered**: 2026-04-04
**Component**: Master Data / Animal Species

---

## 問題概要

動物種マスタの削除時にペット参照の有無をチェックする FK 依存チェックが実装されていない。

### 期待動作
- 動物種を使用中のペットが存在する場合：HTTP 409 Conflict を返す
- エラーメッセージ：「この動物種はペット情報で使用中のため削除できません」

### 現在の動作
- FK チェックなしで DELETE が実行される
- 使用中のデータが存在してもエラーが返されない（データ整合性危機）

---

## 根本原因

### コード検証結果

**backend/internal/service/animal_species_service.go (line 91-93)**
```go
func (s *animalSpeciesService) Delete(ctx context.Context, id uint64) error {
    return s.repo.Delete(ctx, id)  // ❌ FK チェックなし
}
```

**必要な実装**:
1. `backend/internal/repository/pet_repository.go` に `CountPetsByAnimalSpeciesID()` メソッドを追加
2. `animal_species_service.Delete()` に FK 依存チェック追加（参考: `merchandise_item_service.Delete()`）

---

## 影響範囲

| 機能 | 影響 |
|------|------|
| 動物種削除 | ペットが使用中でも削除可能 → データ整合性喪失 |
| テスト報告 | FUNCTIONAL_TEST_REPORT.md 14.16 で「FK依存チェック実装済み」と記録されているが実装されていない |

---

## 参考実装

BUG-109 (物販FK依存チェック) で同じパターンが実装済み:

**backend/internal/service/merchandise_item_service.go**
```go
func (s *merchandiseItemService) Delete(ctx context.Context, clinicID, id uint64) error {
    count, err := s.repo.CountUsageByMerchandiseItemID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check merchandise item dependencies")
    }
    if count > 0 {
        return apperrors.WrapConflict("この物販品目は請求・見積データで使用中のため削除できません")
    }
    return s.repo.Delete(ctx, id)
}
```

---

## 実装方法

1. **Pet Repository に Count メソッド追加**
```go
func (r *petRepository) CountByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error) {
    var count int64
    return count, r.db.WithContext(ctx)
        .Model(&model.Pet{})
        .Where("animal_species_id = ? AND deleted_at IS NULL", speciesID)
        .Count(&count)
        .Error
}
```

2. **Animal Species Service に FK チェック追加**
```go
func (s *animalSpeciesService) Delete(ctx context.Context, id uint64) error {
    count, err := s.petRepo.CountByAnimalSpeciesID(ctx, id)  // petRepo 注入必須
    if err != nil {
        return apperrors.Wrap(err, "failed to check animal species dependencies")
    }
    if count > 0 {
        return apperrors.WrapConflict("この動物種はペット情報で使用中のため削除できません")
    }
    return s.repo.Delete(ctx, id)
}
```

3. **animalSpeciesService に petRepo を注入**
```go
type animalSpeciesService struct {
    repo    repository.AnimalSpeciesRepository
    petRepo repository.PetRepository  // 追加
}

func NewAnimalSpeciesService(repo repository.AnimalSpeciesRepository, petRepo repository.PetRepository) AnimalSpeciesService {
    return &animalSpeciesService{repo: repo, petRepo: petRepo}
}
```

---

## テスト手順

```bash
# 1. テスト用ペット作成（動物種: 犬（ID=1）を使用）
# 2. 動物種削除を試行：DELETE /api/v1/masters/animal-species/1
# 期待結果: HTTP 409 Conflict + メッセージ「この動物種はペット情報で使用中のため削除できません」

# 3. 参照なし動物種削除テスト
# 期待結果: HTTP 204 No Content
```

---

## 関連 Issues

- BUG-109: 物販FK依存チェック実装（参考実装）
- FUNCTIONAL_TEST_REPORT.md 14.16: 動物種削除テストケース

---

**テスト検証済み**: コード実装レビュー確認 (2026-04-04)
**実装担当**: 未割当
