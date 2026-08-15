package lstep

// line_customer_repository_test.go — LineCustomerRepository 統合テスト。
//
// 保護する不変条件:
//   - FindAll / FindByID は clinic_id でテナント隔離される。
//   - FindByID は Owner（clinic_id 一致 AND deleted_at IS NULL）を preload する。
//     LineCustomerService.LinkOwner は service 層で ownerRepo.FindByID による所属クリニック検証を
//     行うが（FE-refactor.md 残件 3 対応）、本テストは repository 単体の Preload 防御を独立して
//     固定する — service ガードを経由しない直接呼び出し（不正データ・将来の呼び出し元）でも
//     Owner は preload されず nil にフォールバックすることを保証する（GetLiffProfile/GetHealthCard
//     のクロステナント露出防止の多層防御）。
//   - FindOrCreateByLineUserID は未登録なら作成し、既存なら既存行を返す（重複作成しない）。
//   - UpdateAdditionalFields / UpdateOwnerLink は clinicScope に一致しない行を更新しない。
//   - UpdateOwnerLink は対象なしで NotFound を返す。

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupLineCustomerTestDB は line_customers テーブルと関連テーブルを整備する。
func setupLineCustomerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.LineCustomer{}))
	db.Exec("TRUNCATE TABLE line_customers CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

// makeLineCustomer はテスト用 LineCustomer を作成して返す。
func makeLineCustomer(t *testing.T, db *gorm.DB, clinicID uint64, lineUserID, displayName string) *model.LineCustomer {
	t.Helper()
	c := &model.LineCustomer{
		ClinicID:         clinicID,
		LineUserID:       lineUserID,
		DisplayName:      displayName,
		AdditionalFields: []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(c).Error)
	return c
}

func TestLineCustomerRepository_FindAll(t *testing.T) {
	db := setupLineCustomerTestDB(t)
	repo := NewLineCustomerRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	makeLineCustomer(t, db, clinicA, "lineA1", "Aさん1")
	makeLineCustomer(t, db, clinicA, "lineA2", "Aさん2")
	makeLineCustomer(t, db, clinicB, "lineB1", "Bさん1")

	t.Run("returns only same-clinic customers", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 2)
		for _, c := range got {
			assert.Equal(t, clinicA, c.ClinicID)
		}
	})

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "lineB1", got[0].LineUserID)
	})

	t.Run("empty clinic returns empty slice not nil", func(t *testing.T) {
		got, err := repo.FindAll(ctx, uint64(999))
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})
}

// G2F-05 / SOLO-16: FindAll is safety-capped (full page/limit API needs service allowlist).
func TestLineCustomerRepository_FindAll_RespectsSafetyCap(t *testing.T) {
	db := setupLineCustomerTestDB(t)
	repo := NewLineCustomerRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	for i := 0; i < lineCustomerListMax+5; i++ {
		makeLineCustomer(t, db, clinicID, fmt.Sprintf("cap-line-%04d", i), fmt.Sprintf("cap-%d", i))
	}

	got, err := repo.FindAll(ctx, clinicID)
	require.NoError(t, err)
	require.Len(t, got, lineCustomerListMax)
	assert.LessOrEqual(t, len(got), lineCustomerListMax)
}

func TestLineCustomerRepository_FindByID(t *testing.T) {
	db := setupLineCustomerTestDB(t)
	repo := NewLineCustomerRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := testdb.MakeTestOwner(t, db, clinicA, "飼主リンク")
	c := makeLineCustomer(t, db, clinicA, "lineA1", "Aさん1")
	require.NoError(t, db.Model(&model.LineCustomer{}).Where("id = ?", c.ID).Update("owner_id", owner.ID).Error)

	t.Run("returns customer with Owner preloaded", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		require.NotNil(t, got.Owner)
		assert.Equal(t, owner.ID, got.Owner.ID)
	})

	t.Run("does not preload a foreign-clinic pet from a contaminated owner relation", func(t *testing.T) {
		foreignPet := makeSpeciesAndPet(t, db, clinicB, owner.ID, "他院ペット")

		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		require.NotNil(t, got.Owner)
		for i := range got.Owner.Pets {
			assert.NotEqual(t, foreignPet.ID, got.Owner.Pets[i].ID)
			assert.Equal(t, clinicA, got.Owner.Pets[i].ClinicID)
		}
	})

	t.Run("not found for nonexistent id returns NotFound error", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("clinic isolation: different clinic cannot fetch", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, c.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("soft-deleted owner is not preloaded", func(t *testing.T) {
		deletedOwner := testdb.MakeTestOwner(t, db, clinicA, "削除済み飼主")
		c2 := makeLineCustomer(t, db, clinicA, "lineA-deletedowner", "Aさん3")
		require.NoError(t, db.Model(&model.LineCustomer{}).Where("id = ?", c2.ID).Update("owner_id", deletedOwner.ID).Error)
		require.NoError(t, db.Delete(deletedOwner).Error)

		got, err := repo.FindByID(ctx, clinicA, c2.ID)
		require.NoError(t, err)
		assert.Nil(t, got.Owner, "ソフトデリート済みの Owner は preload されない")
	})

	t.Run("cross-clinic owner_id linkage does not leak the other clinic's Owner", func(t *testing.T) {
		// LineCustomerService.LinkOwner は service 層で他クリニックの ownerID を NotFound
		// として拒否し、UpdateOwnerLink には到達しない(FE-refactor.md 残件 3 対応)。
		// 本テストは repository 単体の防御を独立検証するため、service ガードを経由せず
		// UpdateOwnerLink を直接叩いて不正データ相当の状況を再現し、read 側(Preload の
		// clinic_id 述語)が単独でもクロステナント漏洩を防いでいることを確認する。
		otherClinicOwner := testdb.MakeTestOwner(t, db, clinicB, "他院の飼主")
		c3 := makeLineCustomer(t, db, clinicA, "lineA-crossclinic", "Aさん4")
		require.NoError(t, repo.UpdateOwnerLink(ctx, clinicA, c3.ID, &otherClinicOwner.ID))

		got, err := repo.FindByID(ctx, clinicA, c3.ID)
		require.NoError(t, err)
		assert.Nil(t, got.Owner, "他クリニックの owner_id は preload されず nil にフォールバックすること（IDOR 防止）")
	})
}

func TestLineCustomerRepository_FindOrCreateByLineUserID(t *testing.T) {
	db := setupLineCustomerTestDB(t)
	repo := NewLineCustomerRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)

	t.Run("creates a new customer when not found", func(t *testing.T) {
		got, err := repo.FindOrCreateByLineUserID(ctx, clinicA, "new-line-user", "新規さん")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.NotZero(t, got.ID)
		assert.Equal(t, "新規さん", got.DisplayName)
	})

	t.Run("returns existing customer without creating a duplicate", func(t *testing.T) {
		first, err := repo.FindOrCreateByLineUserID(ctx, clinicA, "existing-line-user", "既存さん")
		require.NoError(t, err)

		second, err := repo.FindOrCreateByLineUserID(ctx, clinicA, "existing-line-user", "別名を渡しても既存が返る")
		require.NoError(t, err)
		assert.Equal(t, first.ID, second.ID)
		assert.Equal(t, "既存さん", second.DisplayName, "既存行を返すのでDisplayNameは変わらない")

		var count int64
		require.NoError(t, db.Model(&model.LineCustomer{}).Where("line_user_id = ?", "existing-line-user").Count(&count).Error)
		assert.Equal(t, int64(1), count, "重複作成されない")
	})

	t.Run("same line_user_id in different clinic creates a separate row", func(t *testing.T) {
		const clinicB = uint64(2)
		_, err := repo.FindOrCreateByLineUserID(ctx, clinicA, "shared-line-user", "clinicA")
		require.NoError(t, err)
		gotB, err := repo.FindOrCreateByLineUserID(ctx, clinicB, "shared-line-user", "clinicB")
		require.NoError(t, err)
		assert.Equal(t, clinicB, gotB.ClinicID)

		var count int64
		require.NoError(t, db.Model(&model.LineCustomer{}).Where("line_user_id = ?", "shared-line-user").Count(&count).Error)
		assert.Equal(t, int64(2), count)
	})
}

func TestLineCustomerRepository_UpdateAdditionalFields(t *testing.T) {
	db := setupLineCustomerTestDB(t)
	repo := NewLineCustomerRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	c := makeLineCustomer(t, db, clinicA, "lineA1", "Aさん1")

	t.Run("updates fields for matching clinic", func(t *testing.T) {
		require.NoError(t, repo.UpdateAdditionalFields(ctx, clinicA, c.ID, []byte(`{"foo":"bar"}`)))

		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.JSONEq(t, `{"foo":"bar"}`, string(got.AdditionalFields))
	})

	t.Run("clinic isolation: wrong clinic does not update and does not error", func(t *testing.T) {
		err := repo.UpdateAdditionalFields(ctx, clinicB, c.ID, []byte(`{"hacked":true}`))
		require.NoError(t, err, "0件更新でもエラーにはならない実装")

		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.JSONEq(t, `{"foo":"bar"}`, string(got.AdditionalFields), "別クリニックからの更新は反映されない")
	})
}

func TestLineCustomerRepository_UpdateOwnerLink(t *testing.T) {
	db := setupLineCustomerTestDB(t)
	repo := NewLineCustomerRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := testdb.MakeTestOwner(t, db, clinicA, "紐付け対象飼主")
	c := makeLineCustomer(t, db, clinicA, "lineA1", "Aさん1")

	t.Run("links owner successfully", func(t *testing.T) {
		require.NoError(t, repo.UpdateOwnerLink(ctx, clinicA, c.ID, &owner.ID))

		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		require.NotNil(t, got.OwnerID)
		assert.Equal(t, owner.ID, *got.OwnerID)
	})

	t.Run("unlinks owner with nil", func(t *testing.T) {
		require.NoError(t, repo.UpdateOwnerLink(ctx, clinicA, c.ID, nil))

		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.Nil(t, got.OwnerID)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		err := repo.UpdateOwnerLink(ctx, clinicA, uint64(999999), &owner.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation: wrong clinic returns NotFound", func(t *testing.T) {
		err := repo.UpdateOwnerLink(ctx, clinicB, c.ID, &owner.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
