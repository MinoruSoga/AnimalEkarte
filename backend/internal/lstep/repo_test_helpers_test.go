package lstep

// repo_test_helpers_test.go — internal/repository の test helper の最小限の複製
// （package 跨ぎ import 不能のため・他 domain 先例）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// makeSpeciesAndPet — B④でbillingへ移動したaccounting_repository_unpaid_test.go由来のcarrier複製（appointment_admin用・lstep/appointment移行時に解消）。
// makeSpeciesAndPet はテスト用の AnimalSpecies と Pet を作成して返す。
func makeSpeciesAndPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: species.ID,
		Name:            petName,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}
