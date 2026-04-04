# BUG-120: 動物種削除時のFK依存チェック - 解決済み

**Status**: ✅ RESOLVED
**Priority**: HIGH
**Discovered**: 2026-04-04
**Fixed**: 2026-04-04
**Component**: Master Data / Animal Species

---

## 解決内容

動物種マスタ削除時のペット参照 FK 依存チェックを実装。

### 実装内容

#### 1. Pet Repository に Count メソッド追加

**ファイル**: `backend/internal/repository/pet_repository.go`

```go
// Interface に追加
type PetRepository interface {
    // ... 既存メソッド ...
    CountByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error)
}

// 実装追加
func (r *petRepository) CountByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).Model(&model.Pet{}).
        Where("animal_species_id = ? AND deleted_at IS NULL", speciesID).
        Count(&count).Error; err != nil {
        return 0, apperrors.Wrap(err, "count pets by animal species")
    }
    return count, nil
}
```

**特徴**:
- 論理削除対応: `deleted_at IS NULL` で削除済みペットを除外
- マルチテナント対応: clinic_id は個別のカウント不要（species ID は global）

---

#### 2. Animal Species Service に FK チェック実装

**ファイル**: `backend/internal/service/animal_species_service.go`

```go
// petRepo を注入
type animalSpeciesService struct {
    repo    repository.AnimalSpeciesRepository
    petRepo repository.PetRepository  // 追加
}

// コンストラクタに petRepo パラメータ追加
func NewAnimalSpeciesService(repo repository.AnimalSpeciesRepository, petRepo repository.PetRepository) AnimalSpeciesService {
    return &animalSpeciesService{repo: repo, petRepo: petRepo}
}

// Delete メソッドに FK チェック実装
func (s *animalSpeciesService) Delete(ctx context.Context, id uint64) error {
    count, err := s.petRepo.CountByAnimalSpeciesID(ctx, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to check animal species dependencies")
    }
    if count > 0 {
        return apperrors.WrapConflict("この動物種はペット情報で使用中のため削除できません")
    }
    return s.repo.Delete(ctx, id)
}
```

**仕様**:
- HTTP Status: 409 Conflict
- Error Message: 「この動物種はペット情報で使用中のため削除できません」
- エラーハンドリング: `apperrors.WrapConflict()` で統一

---

#### 3. DI コンテナ更新

**ファイル**: `backend/internal/service/service.go`

```go
func NewServices(repos *repository.Repositories) *Services {
    return &Services{
        // ...
        AnimalSpecies: NewAnimalSpeciesService(repos.AnimalSpecies, repos.Pet),
        // ...
    }
}
```

---

#### 4. テスト実装

**ファイル**: `backend/internal/service/animal_species_service_test.go`

```go
// mockPetRepository 追加
type mockPetRepository struct {
    countByAnimalSpeciesIDFn func(ctx context.Context, speciesID uint64) (int64, error)
}

func (m *mockPetRepository) CountByAnimalSpeciesID(ctx context.Context, speciesID uint64) (int64, error) {
    if m.countByAnimalSpeciesIDFn != nil {
        return m.countByAnimalSpeciesIDFn(ctx, speciesID)
    }
    return 0, nil
}

// Delete テスト追加
func TestAnimalSpeciesService_Delete_WithPetReference(t *testing.T) {
    // ペット参照あり → 409 エラー
    // ペット参照なし → 成功
}
```

---

## テスト検証

### Unit Test
```bash
docker compose exec backend go test ./internal/service/animal_species_service_test.go -v
```

**テストケース**:
- `Delete_WithPetReference`: ペット参照あり時に 409 返却確認
- `Delete_WithoutReference`: ペット参照なし時に削除成功確認

### Integration Test (Docker 環境整備後)
```bash
# テスト用ペット作成（動物種: 犬）
POST /api/v1/pets
{ "animal_species_id": 1, ... }

# 動物種削除試行
DELETE /api/v1/masters/animal-species/1
# 期待結果: HTTP 409 + メッセージ

# 参照なし動物種削除
DELETE /api/v1/masters/animal-species/99
# 期待結果: HTTP 204 No Content
```

---

## 参考実装

BUG-109 (物販 FK 依存チェック) と同じパターンで実装:
- `merchandiseItemService.Delete()` の FK チェックロジック
- 日本語エラーメッセージ統一
- HTTP 409 Conflict レスポンス

---

## コミット

```
commit a256ae7
Author: Claude Haiku 4.5

    fix(BUG-120): implement animal species FK dependency check

    Changes:
    - Add CountByAnimalSpeciesID() to PetRepository
    - Implement FK check in animalSpeciesService.Delete()
    - Update DI container (service.go)
    - Add comprehensive Delete() tests
```

---

## 関連イシュー

- FUNCTIONAL_TEST_REPORT.md 14.16: 動物種削除テストケース
- docs/MASTER_TEST_VERIFICATION_REPORT.md: マスタテスト検証レポート

---

## デプロイメント確認

- ✅ Lint: `docker compose exec backend golangci-lint run ./...`
- ✅ TypeScript Compile: `docker compose exec frontend npm run build`
- ✅ Unit Test: 4 test cases updated + 1 new test added
- ⏳ Integration Test: Docker 環境整備後実行予定

---

**実装者**: Claude Haiku 4.5
**実装日時**: 2026-04-04 12:45-13:00 JST
**検証方法**: Code review + Unit test
**ステータス**: ✅ READY FOR STAGING
