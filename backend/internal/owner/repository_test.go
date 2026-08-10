package owner

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
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

// ---- escapeLike（純粋関数・DB不要） ----

func TestEscapeLike(t *testing.T) {
	// ltv_repository_test.go の TestEscapeLikePattern と同一の入出力パターンで、
	// 同等のエスケープロジック（\ → %  → _ の順）を持つ escapeLike を検証する。
	assert.Equal(t, `100\%\_\\`, textsearch.EscapeLike(`100%_\`))
	assert.Equal(t, `normal`, textsearch.EscapeLike(`normal`))
}

// ---- FindAll ----

func TestOwnerRepository_FindAll(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA1 := makeTestOwner(t, db, clinicA, "検索対象飼主タロウ")
	_ = makeTestOwner(t, db, clinicA, "別の飼主ハナコ")
	_ = makeTestOwner(t, db, clinicB, "医院Bの飼主")

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
		// パターンを name ILIKE と name_kana(translate 済み) ILIKE の両方に使う。makeTestOwner は
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
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB, clinicC = uint64(1), uint64(2), uint64(3)

	owner := makeTestOwner(t, db, clinicA, "拠点横断飼主")

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
	repo := newTestRepository(db)
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
	repo := newTestRepository(db)
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
	repo := newTestRepository(db)
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
		// POC-06: after uk_owners_clinic_phone, concurrent phone dups are rejected at DB.
		// Keep multi-match defense only for environments without the index (AutoMigrate tests).
		require.NoError(t, db.Exec(`DROP INDEX IF EXISTS uk_owners_clinic_phone`).Error)
		dup1 := &model.Owner{ClinicID: clinicA, Name: "重複飼主", Phone: "070-2222-2222"}
		require.NoError(t, db.WithContext(ctx).Create(dup1).Error)
		dup2 := &model.Owner{ClinicID: clinicA, Name: "重複飼主", Phone: "070-2222-2222"}
		require.NoError(t, db.WithContext(ctx).Create(dup2).Error)

		got, err := repo.FindByNameAndPhone(ctx, clinicA, "重複飼主", "070-2222-2222")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestOwnerRepository_PhoneUniqueConstraint(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	// Match production migration 007 (AutoMigrate does not create partial unique indexes).
	require.NoError(t, db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uk_owners_clinic_phone
		ON owners (clinic_id, phone)
		WHERE deleted_at IS NULL AND phone <> ''
	`).Error)
	t.Cleanup(func() {
		_ = db.Exec(`DROP INDEX IF EXISTS uk_owners_clinic_phone`).Error
	})

	first := &model.Owner{ClinicID: clinicA, Name: "電話一意A", Phone: "090-3333-3333"}
	require.NoError(t, repo.CreateWithPets(ctx, first, nil))

	dup := &model.Owner{ClinicID: clinicA, Name: "電話一意B", Phone: "090-3333-3333"}
	err := repo.CreateWithPets(ctx, dup, nil)
	require.Error(t, err)
	assert.True(t, apperrors.IsAlreadyExists(err), "expected AlreadyExists, got %v", err)
	// BUG-019: natural Japanese message on phone unique constraint
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "この電話番号はすでに登録されています", appErr.Message)
	assert.NotContains(t, appErr.Message, "already exists")

	// Empty phones may coexist (partial index excludes '').
	empty1 := &model.Owner{ClinicID: clinicA, Name: "空電話1", Phone: ""}
	empty2 := &model.Owner{ClinicID: clinicA, Name: "空電話2", Phone: ""}
	require.NoError(t, repo.CreateWithPets(ctx, empty1, nil))
	require.NoError(t, repo.CreateWithPets(ctx, empty2, nil))
}

// ---- CreateWithPets ----

func TestOwnerRepository_CreateWithPets(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
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
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "正常削除対象飼主")

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
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	lineID := "U-find-by-line"
	owner := makeTestOwner(t, db, clinicA, "LINE検索飼主")
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
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	linked := makeTestOwner(t, db, clinicA, "連携済み飼主")
	lineID := "U-linked"
	require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, linked.ID, &lineID))
	_ = makeTestOwner(t, db, clinicA, "未連携飼主")

	otherClinicLinked := makeTestOwner(t, db, clinicB, "医院B連携済み飼主")
	otherLineID := "U-other-clinic"
	require.NoError(t, repo.UpdateLineUserID(ctx, clinicB, otherClinicLinked.ID, &otherLineID))

	got, err := repo.FindAllWithLineUserID(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, linked.ID, got[0].ID)
}

// ---- FindAllWithLineUserIDCursor（PERF-FOLLOWUP-02: カーソルページネーション） ----

// seedLineLinkedOwners は clinicID に line_user_id 設定済みの飼主を count 件一括作成する
// （FindAllWithLineUserIDCursor のページ境界テスト用）。
func seedLineLinkedOwners(t *testing.T, db *gorm.DB, clinicID uint64, count int) {
	t.Helper()
	owners := make([]model.Owner, count)
	for i := range owners {
		lineID := fmt.Sprintf("U-cursor-%d-%d", clinicID, i)
		owners[i] = model.Owner{ClinicID: clinicID, Name: fmt.Sprintf("カーソル飼主%d", i), LineUserID: &lineID}
	}
	require.NoError(t, db.WithContext(context.Background()).CreateInBatches(owners, 200).Error)
}

func TestOwnerRepository_FindAllWithLineUserIDCursor_ExactlyOnePage(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	const pageSize = 500

	seedLineLinkedOwners(t, db, clinicA, pageSize)

	page1, err := repo.FindAllWithLineUserIDCursor(ctx, clinicA, 0, pageSize)
	require.NoError(t, err)
	require.Len(t, page1, pageSize, "ちょうど pageSize 件は1ページに収まる")

	afterID := page1[len(page1)-1].ID
	page2, err := repo.FindAllWithLineUserIDCursor(ctx, clinicA, afterID, pageSize)
	require.NoError(t, err)
	assert.Empty(t, page2, "全件消化後の次ページ取得は空を返す")

	// カーソルは id 昇順で重複・欠落がないこと。
	ids := make(map[uint64]bool, len(page1))
	prev := uint64(0)
	for _, o := range page1 {
		assert.False(t, ids[o.ID], "重複 id が含まれてはいけない")
		ids[o.ID] = true
		assert.Greater(t, o.ID, prev, "id は昇順でなければならない")
		prev = o.ID
	}
}

func TestOwnerRepository_FindAllWithLineUserIDCursor_TwoPages(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	const pageSize = 500
	const total = pageSize + 1

	seedLineLinkedOwners(t, db, clinicA, total)

	page1, err := repo.FindAllWithLineUserIDCursor(ctx, clinicA, 0, pageSize)
	require.NoError(t, err)
	require.Len(t, page1, pageSize)

	afterID := page1[len(page1)-1].ID
	page2, err := repo.FindAllWithLineUserIDCursor(ctx, clinicA, afterID, pageSize)
	require.NoError(t, err)
	require.Len(t, page2, 1, "501 件目は2ページ目に1件だけ現れる")

	// 重複・欠落なし: 全体件数が total と一致し、id の重複がない。
	seen := make(map[uint64]bool, total)
	for _, o := range append(page1, page2...) {
		assert.False(t, seen[o.ID], "重複 id が含まれてはいけない")
		seen[o.ID] = true
	}
	assert.Len(t, seen, total)

	page3, err := repo.FindAllWithLineUserIDCursor(ctx, clinicA, page2[len(page2)-1].ID, pageSize)
	require.NoError(t, err)
	assert.Empty(t, page3, "全件消化後の次ページ取得は空を返す")
}

func TestOwnerRepository_UpdateLineFollowedAt(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "LINEフォロー飼主")
	lineUserID := "U-follow-owner"
	require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, owner.ID, &lineUserID))
	pastBlock := time.Now().Add(-24 * time.Hour)
	require.NoError(t, db.Model(&model.Owner{}).Where("id = ?", owner.ID).Update("line_blocked_at", pastBlock).Error)

	t.Run("同一クリニックからの更新でフォロー日時がセットされブロック日時がリセットされる", func(t *testing.T) {
		followedAt := time.Now()
		updated, err := repo.UpdateLineFollowedAt(ctx, clinicA, owner.ID, lineUserID, followedAt)
		require.NoError(t, err)
		assert.True(t, updated)

		got, err := repo.FindByID(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, got.LineFollowedAt)
		assert.WithinDuration(t, followedAt, *got.LineFollowedAt, time.Second)
		assert.Nil(t, got.LineBlockedAt, "フォロー時にブロック日時はリセットされるべき")
	})

	t.Run("別クリニックからの更新は実データを変更しない（clinic_id述語で対象0件）", func(t *testing.T) {
		other := makeTestOwner(t, db, clinicA, "変更されないはずの飼主")
		otherLineUserID := "U-follow-other"
		require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, other.ID, &otherLineUserID))
		updated, err := repo.UpdateLineFollowedAt(ctx, clinicB, other.ID, otherLineUserID, time.Now())
		require.NoError(t, err)
		assert.False(t, updated)

		got, err := repo.FindByID(ctx, clinicA, other.ID)
		require.NoError(t, err)
		assert.Nil(t, got.LineFollowedAt, "別クリニックからの呼び出しでは実際には更新されない")
	})

	t.Run("再連携後の古いLINE User IDイベントは更新しない", func(t *testing.T) {
		other := makeTestOwner(t, db, clinicA, "再連携済み飼主")
		currentLineUserID := "U-current-follow"
		require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, other.ID, &currentLineUserID))

		updated, err := repo.UpdateLineFollowedAt(ctx, clinicA, other.ID, "U-stale-follow", time.Now())

		require.NoError(t, err)
		assert.False(t, updated)
		got, findErr := repo.FindByID(ctx, clinicA, other.ID)
		require.NoError(t, findErr)
		assert.Nil(t, got.LineFollowedAt)
	})
}

func TestOwnerRepository_UpdateLineBlockedAt(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックからブロック日時を更新できる", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "LINEブロック飼主")
		lineUserID := "U-block-owner"
		require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, owner.ID, &lineUserID))
		blockedAt := time.Now()
		updated, err := repo.UpdateLineBlockedAt(ctx, clinicA, owner.ID, lineUserID, blockedAt)
		require.NoError(t, err)
		assert.True(t, updated)

		got, err := repo.FindByID(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, got.LineBlockedAt)
		assert.WithinDuration(t, blockedAt, *got.LineBlockedAt, time.Second)
	})

	t.Run("別クリニックからの更新はNotFoundになる", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "変更されないブロック飼主")
		lineUserID := "U-block-other"
		require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, owner.ID, &lineUserID))

		updated, err := repo.UpdateLineBlockedAt(ctx, clinicB, owner.ID, lineUserID, time.Now())

		require.NoError(t, err)
		assert.False(t, updated)
		got, findErr := repo.FindByID(ctx, clinicA, owner.ID)
		require.NoError(t, findErr)
		assert.Nil(t, got.LineBlockedAt)
	})

	t.Run("再連携後の古いLINE User IDイベントは更新しない", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "再連携済みブロック飼主")
		currentLineUserID := "U-current-block"
		require.NoError(t, repo.UpdateLineUserID(ctx, clinicA, owner.ID, &currentLineUserID))

		updated, err := repo.UpdateLineBlockedAt(ctx, clinicA, owner.ID, "U-stale-block", time.Now())

		require.NoError(t, err)
		assert.False(t, updated)
		got, findErr := repo.FindByID(ctx, clinicA, owner.ID)
		require.NoError(t, findErr)
		assert.Nil(t, got.LineBlockedAt)
	})
}

func TestOwnerRepository_LineWebhookEventOrdering(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	base := time.Date(2026, time.July, 24, 9, 0, 0, 0, time.UTC)
	older := base.Add(time.Minute)
	newer := base.Add(2 * time.Minute)

	type eventStep struct {
		eventType  string
		eventAt    time.Time
		wantUpdate bool
	}
	tests := []struct {
		name          string
		steps         []eventStep
		wantFollowed  *time.Time
		wantBlockedAt *time.Time
	}{
		{
			name: "newer unfollow rejects an older follow",
			steps: []eventStep{
				{eventType: "unfollow", eventAt: newer, wantUpdate: true},
				{eventType: "follow", eventAt: older, wantUpdate: false},
			},
			wantBlockedAt: &newer,
		},
		{
			name: "newer follow rejects an older unfollow",
			steps: []eventStep{
				{eventType: "follow", eventAt: newer, wantUpdate: true},
				{eventType: "unfollow", eventAt: older, wantUpdate: false},
			},
			wantFollowed: &newer,
		},
		{
			name: "duplicate follow is an idempotent no-op",
			steps: []eventStep{
				{eventType: "follow", eventAt: newer, wantUpdate: true},
				{eventType: "follow", eventAt: newer, wantUpdate: false},
			},
			wantFollowed: &newer,
		},
		{
			name: "same timestamp unfollow wins and duplicates are no-op",
			steps: []eventStep{
				{eventType: "follow", eventAt: newer, wantUpdate: true},
				{eventType: "unfollow", eventAt: newer, wantUpdate: true},
				{eventType: "follow", eventAt: newer, wantUpdate: false},
				{eventType: "unfollow", eventAt: newer, wantUpdate: false},
			},
			wantFollowed:  &newer,
			wantBlockedAt: &newer,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner := makeTestOwner(t, db, clinicID, fmt.Sprintf("LINE順序テスト%d", i))
			lineUserID := fmt.Sprintf("U-ordering-%d", i)
			require.NoError(t, repo.UpdateLineUserID(ctx, clinicID, owner.ID, &lineUserID))

			for _, step := range tt.steps {
				var (
					updated bool
					err     error
				)
				switch step.eventType {
				case "follow":
					updated, err = repo.UpdateLineFollowedAt(
						ctx,
						clinicID,
						owner.ID,
						lineUserID,
						step.eventAt,
					)
				case "unfollow":
					updated, err = repo.UpdateLineBlockedAt(
						ctx,
						clinicID,
						owner.ID,
						lineUserID,
						step.eventAt,
					)
				default:
					t.Fatalf("unsupported event type: %s", step.eventType)
				}
				require.NoError(t, err)
				assert.Equal(t, step.wantUpdate, updated)
			}

			got, err := repo.FindByID(ctx, clinicID, owner.ID)
			require.NoError(t, err)
			if tt.wantFollowed == nil {
				assert.Nil(t, got.LineFollowedAt)
			} else {
				require.NotNil(t, got.LineFollowedAt)
				assert.WithinDuration(t, *tt.wantFollowed, *got.LineFollowedAt, time.Millisecond)
			}
			if tt.wantBlockedAt == nil {
				assert.Nil(t, got.LineBlockedAt)
			} else {
				require.NotNil(t, got.LineBlockedAt)
				assert.WithinDuration(t, *tt.wantBlockedAt, *got.LineBlockedAt, time.Millisecond)
			}
		})
	}
}

func TestOwnerRepository_UpdateLineUserID(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "LINE連携飼主")
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
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "ペット数集計飼主")

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
		require.NoError(t, db.WithContext(ctx).
			Model(&model.Pet{}).
			Where("clinic_id = ? AND id = ?", clinicA, pet1.ID).
			Delete(&model.Pet{}).Error)
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
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	o1 := makeTestOwner(t, db, clinicA, "一括取得飼主1")
	o2 := makeTestOwner(t, db, clinicA, "一括取得飼主2")
	o3 := makeTestOwner(t, db, clinicB, "医院B飼主")

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
