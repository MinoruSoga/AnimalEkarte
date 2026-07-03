package repository

// owner_repository_test.go — OwnerRepository の統合テスト。
// 実 Postgres テスト DB (setupTestDB) に対して実行する。
//
// clinic_id 隔離の基本ケース（FindByID/Update/Delete の isolation）は既に
// owner_pet_clinic_isolation_test.go でカバーされているため、本ファイルではそれ以外の
// メソッド（FindAll, FindByIDForClinics, FindByEmail, FindByPhone, FindByNameAndPhone,
// CreateWithPets, LINE 連携系, CountPetsByOwnerID, FindByIDs）と、DB を伴わない純粋関数
// escapeLike を対象にする。DB セットアップは owner_pet_clinic_isolation_test.go で定義済みの
// setupOwnerPetIsolationTestDB を再利用する（重複定義を避ける）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- escapeLike（純粋関数・DB不要） ----

func TestEscapeLike(t *testing.T) {
	// ltv_repository_test.go の TestEscapeLikePattern と同一の入出力パターンで、
	// 同等のエスケープロジック（\ → %  → _ の順）を持つ escapeLike を検証する。
	assert.Equal(t, `100\%\_\\`, escapeLike(`100%_\`))
	assert.Equal(t, `normal`, escapeLike(`normal`))
}

// ---- FindAll ----

func TestOwnerRepository_FindAll(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA1 := makeOwner(t, db, clinicA, "検索対象飼主タロウ")
	_ = makeOwner(t, db, clinicA, "別の飼主ハナコ")
	_ = makeOwner(t, db, clinicB, "医院Bの飼主")

	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	pet := &model.Pet{ClinicID: clinicA, OwnerID: ownerA1.ID, AnimalSpeciesID: species.ID, Name: "検索対象ペット"}
	require.NoError(t, db.WithContext(ctx).Create(pet).Error)

	t.Run("clinicIDsに一致する飼主のみ返しペットがPreloadされる", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, "")
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		var found *model.Owner
		for i := range got {
			if got[i].ID == ownerA1.ID {
				found = &got[i]
			}
		}
		require.NotNil(t, found)
		require.Len(t, found.Pets, 1)
		assert.Equal(t, pet.ID, found.Pets[0].ID)
		require.NotNil(t, found.Pets[0].AnimalSpecies)
	})

	t.Run("searchで部分一致検索できる", func(t *testing.T) {
		// owner_repository.go の search は NormalizeKana(search) でカタカナ→ひらがな変換した
		// パターンを name ILIKE と name_kana(translate 済み) ILIKE の両方に使う。makeOwner は
		// NameKana を設定しないため、"タロウ"（カタカナ）で検索すると正規化後 "たろう"
		// （ひらがな）になり、name 列の生カタカナ "タロウ" とは文字種が異なり ILIKE が
		// マッチしない。カナ正規化の影響を受けない漢字部分文字列で検索する。
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, "検索対象")
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, ownerA1.ID, got[0].ID)
	})

	t.Run("clinicIDsが空はフェイルセーフで空配列を返す", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{}, 1, 100, "")
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, got)
	})

	t.Run("該当しないクリニックの飼主は含まれない", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, "")
		require.NoError(t, err)
		for _, o := range got {
			assert.Equal(t, clinicA, o.ClinicID)
		}
	})
}

// ---- FindByIDForClinics ----

func TestOwnerRepository_FindByIDForClinics(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB, clinicC = uint64(1), uint64(2), uint64(3)

	owner := makeOwner(t, db, clinicA, "拠点横断飼主")

	t.Run("所属クリニックのいずれかに一致すれば取得できる", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicC}, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, owner.ID, got.ID)
	})

	t.Run("所属していないクリニック集合ではNotFound", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(ctx, []uint64{clinicB, clinicC}, owner.ID)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("空集合はNotFound", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(ctx, []uint64{}, owner.ID)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// ---- FindByEmail / FindByPhone / FindByNameAndPhone ----

func TestOwnerRepository_FindByEmail(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := &model.Owner{ClinicID: clinicA, Name: "メール検索飼主", Email: "search-target@example.com"}
	require.NoError(t, db.WithContext(ctx).Create(owner).Error)

	t.Run("同一クリニックで見つかる", func(t *testing.T) {
		got, err := repo.FindByEmail(ctx, clinicA, "search-target@example.com")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, owner.ID, got.ID)
	})

	t.Run("該当なしはnil,nil", func(t *testing.T) {
		got, err := repo.FindByEmail(ctx, clinicA, "not-exist@example.com")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("別クリニックからはnil,nil", func(t *testing.T) {
		got, err := repo.FindByEmail(ctx, clinicB, "search-target@example.com")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestOwnerRepository_FindByPhone(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := &model.Owner{ClinicID: clinicA, Name: "電話検索飼主", Phone: "090-1234-5678"}
	require.NoError(t, db.WithContext(ctx).Create(owner).Error)

	t.Run("同一クリニックで見つかる", func(t *testing.T) {
		got, err := repo.FindByPhone(ctx, clinicA, "090-1234-5678")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, owner.ID, got.ID)
	})

	t.Run("該当なしはnil,nil", func(t *testing.T) {
		got, err := repo.FindByPhone(ctx, clinicA, "000-0000-0000")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("別クリニックからはnil,nil", func(t *testing.T) {
		got, err := repo.FindByPhone(ctx, clinicB, "090-1234-5678")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestOwnerRepository_FindByNameAndPhone(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	unique := &model.Owner{ClinicID: clinicA, Name: "一意飼主", Phone: "080-1111-1111"}
	require.NoError(t, db.WithContext(ctx).Create(unique).Error)

	t.Run("名前と電話番号が完全一致で1件のみヒット", func(t *testing.T) {
		got, err := repo.FindByNameAndPhone(ctx, clinicA, "一意飼主", "080-1111-1111")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, unique.ID, got.ID)
	})

	t.Run("一致なしはnil,nil", func(t *testing.T) {
		got, err := repo.FindByNameAndPhone(ctx, clinicA, "存在しない飼主", "000-0000-0000")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("前後空白のみの入力はトリム後空扱いでnil,nil", func(t *testing.T) {
		got, err := repo.FindByNameAndPhone(ctx, clinicA, "  ", "080-1111-1111")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("同名同電話番号が複数件ある場合はnil,nil（自動紐付け不可）", func(t *testing.T) {
		dup1 := &model.Owner{ClinicID: clinicA, Name: "重複飼主", Phone: "070-2222-2222"}
		require.NoError(t, db.WithContext(ctx).Create(dup1).Error)
		dup2 := &model.Owner{ClinicID: clinicA, Name: "重複飼主", Phone: "070-2222-2222"}
		require.NoError(t, db.WithContext(ctx).Create(dup2).Error)

		got, err := repo.FindByNameAndPhone(ctx, clinicA, "重複飼主", "070-2222-2222")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

// ---- CreateWithPets ----

func TestOwnerRepository_CreateWithPets(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	species := &model.AnimalSpecies{Name: "猫"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)

	owner := &model.Owner{ClinicID: clinicA, Name: "新規飼主", Email: "new-owner@example.com"}
	pets := []model.Pet{
		{AnimalSpeciesID: species.ID, Name: "ペット1"},
		{AnimalSpeciesID: species.ID, Name: "ペット2"},
	}

	require.NoError(t, repo.CreateWithPets(ctx, owner, pets))
	assert.NotZero(t, owner.ID)
	require.Len(t, owner.Pets, 2)
	for _, p := range owner.Pets {
		assert.Equal(t, owner.ID, p.OwnerID)
		assert.Equal(t, clinicA, p.ClinicID)
	}

	var petCount int64
	require.NoError(t, db.Model(&model.Pet{}).Where("owner_id = ?", owner.ID).Count(&petCount).Error)
	assert.Equal(t, int64(2), petCount)
}

// ---- Delete（成功パス。isolationは owner_pet_clinic_isolation_test.go で別途カバー） ----

func TestOwnerRepository_Delete_SuccessSoftDeletes(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeOwner(t, db, clinicA, "正常削除対象飼主")

	require.NoError(t, repo.Delete(ctx, clinicA, owner.ID))

	got, err := repo.FindByID(ctx, clinicA, owner.ID)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsNotFound(err))

	var raw model.Owner
	require.NoError(t, db.Unscoped().Where("id = ?", owner.ID).First(&raw).Error)
	assert.True(t, raw.DeletedAt.Valid, "deleted_at がセットされているべき")

	t.Run("存在しないIDの削除はNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 99999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// ---- LINE連携系 ----

func TestOwnerRepository_FindByLineUserID(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	lineID := "U-find-by-line"
	owner := makeOwner(t, db, clinicA, "LINE検索飼主")
	require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, owner.ID, &lineID))

	t.Run("同一クリニックで見つかる", func(t *testing.T) {
		got, err := repo.FindByLineUserID(ctx, clinicA, lineID)
		require.NoError(t, err)
		assert.Equal(t, owner.ID, got.ID)
	})

	t.Run("該当なしはNotFound", func(t *testing.T) {
		got, err := repo.FindByLineUserID(ctx, clinicA, "存在しないline_id")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックからは見つからない", func(t *testing.T) {
		got, err := repo.FindByLineUserID(ctx, clinicB, lineID)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestOwnerRepository_FindAllWithLineUserID(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	linked := makeOwner(t, db, clinicA, "連携済み飼主")
	lineID := "U-linked"
	require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, linked.ID, &lineID))
	_ = makeOwner(t, db, clinicA, "未連携飼主")

	otherClinicLinked := makeOwner(t, db, clinicB, "医院B連携済み飼主")
	otherLineID := "U-other-clinic"
	require.NoError(t, repo.UpdateLineUserID(ctx, clinicB, otherClinicLinked.ID, &otherLineID))

	got, err := repo.FindAllWithLineUserID(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, linked.ID, got[0].ID)
}

func TestOwnerRepository_FindAllByLineUserID(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	lineID := "U-cross-clinic"
	ownerA := makeOwner(t, db, clinicA, "医院A飼主")
	require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, ownerA.ID, &lineID))
	ownerB := makeOwner(t, db, clinicB, "医院B飼主")
	require.NoError(t, repo.UpdateLineUserID(ctx, clinicB, ownerB.ID, &lineID))

	got, err := repo.FindAllByLineUserID(ctx, lineID)
	require.NoError(t, err)
	ids := make([]uint64, len(got))
	for i, o := range got {
		ids[i] = o.ID
	}
	assert.ElementsMatch(t, []uint64{ownerA.ID, ownerB.ID}, ids, "clinic_idを問わず横断で全件返す")
}

func TestOwnerRepository_UpdateLineFollowedAt(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeOwner(t, db, clinicA, "LINEフォロー飼主")
	pastBlock := time.Now().Add(-24 * time.Hour)
	require.NoError(t, db.Model(&model.Owner{}).Where("id = ?", owner.ID).Update("line_blocked_at", pastBlock).Error)

	t.Run("同一クリニックからの更新でフォロー日時がセットされブロック日時がリセットされる", func(t *testing.T) {
		followedAt := time.Now()
		require.NoError(t, repo.UpdateLineFollowedAt(ctx, clinicA, owner.ID, followedAt))

		got, err := repo.FindByID(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, got.LineFollowedAt)
		assert.WithinDuration(t, followedAt, *got.LineFollowedAt, time.Second)
		assert.Nil(t, got.LineBlockedAt, "フォロー時にブロック日時はリセットされるべき")
	})

	t.Run("別クリニックからの更新は実データを変更しない（clinic_id述語で対象0件）", func(t *testing.T) {
		other := makeOwner(t, db, clinicA, "変更されないはずの飼主")
		require.NoError(t, repo.UpdateLineFollowedAt(ctx, clinicB, other.ID, time.Now()))

		got, err := repo.FindByID(ctx, clinicA, other.ID)
		require.NoError(t, err)
		assert.Nil(t, got.LineFollowedAt, "別クリニックからの呼び出しでは実際には更新されない")
	})
}

func TestOwnerRepository_UpdateLineBlockedAt(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeOwner(t, db, clinicA, "LINEブロック飼主")
	blockedAt := time.Now()
	require.NoError(t, repo.UpdateLineBlockedAt(ctx, clinicA, owner.ID, blockedAt))

	got, err := repo.FindByID(ctx, clinicA, owner.ID)
	require.NoError(t, err)
	require.NotNil(t, got.LineBlockedAt)
	assert.WithinDuration(t, blockedAt, *got.LineBlockedAt, time.Second)
}

func TestOwnerRepository_UpdateLineUserID(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeOwner(t, db, clinicA, "LINE連携飼主")
	lineID := "U1234567890"

	t.Run("LINE User IDを設定できる", func(t *testing.T) {
		require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, owner.ID, &lineID))
		got, err := repo.FindByID(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, got.LineUserID)
		assert.Equal(t, lineID, *got.LineUserID)
	})

	t.Run("nilを渡すと連携解除される", func(t *testing.T) {
		require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, owner.ID, nil))
		got, err := repo.FindByID(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Nil(t, got.LineUserID)
	})
}

// ---- CountPetsByOwnerID ----

func TestOwnerRepository_CountPetsByOwnerID(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeOwner(t, db, clinicA, "ペット数集計飼主")

	t.Run("0件", func(t *testing.T) {
		count, err := repo.CountPetsByOwnerID(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	pet1 := makeSpeciesAndPet(t, db, clinicA, owner.ID, "ペット1")
	_ = makeSpeciesAndPet(t, db, clinicA, owner.ID, "ペット2")

	t.Run("2件", func(t *testing.T) {
		count, err := repo.CountPetsByOwnerID(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("ソフトデリートされたペットは除外される", func(t *testing.T) {
		petRepo := NewPetRepository(db)
		require.NoError(t, petRepo.Delete(ctx, clinicA, pet1.ID))
		count, err := repo.CountPetsByOwnerID(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックからは0件", func(t *testing.T) {
		count, err := repo.CountPetsByOwnerID(ctx, clinicB, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// ---- FindByIDs ----

func TestOwnerRepository_FindByIDs(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	o1 := makeOwner(t, db, clinicA, "一括取得飼主1")
	o2 := makeOwner(t, db, clinicA, "一括取得飼主2")
	o3 := makeOwner(t, db, clinicB, "医院B飼主")

	t.Run("指定した同一クリニックのIDのみ返す", func(t *testing.T) {
		got, err := repo.FindByIDs(ctx, clinicA, []uint64{o1.ID, o2.ID, o3.ID})
		require.NoError(t, err)
		ids := make([]uint64, len(got))
		for i, o := range got {
			ids[i] = o.ID
		}
		assert.ElementsMatch(t, []uint64{o1.ID, o2.ID}, ids, "別クリニックのIDは除外される")
	})

	t.Run("空スライスはnilを返す", func(t *testing.T) {
		got, err := repo.FindByIDs(ctx, clinicA, []uint64{})
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
