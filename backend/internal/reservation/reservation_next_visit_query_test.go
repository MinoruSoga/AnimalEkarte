package reservation

// EMR-308: next-visit appointment finder（reservation_next_visit_query.go）の
// クエリセマンティクス検証。LSTEP の next_visit_* タグ同期と tentative
// リマインダー配信が依存するため、clinic スコープ・active 判定・JST 暦日の
// 3 点を実 DB で確認する（source of truth は appointments のみ）。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// makeNextVisitQueryReservation は owner/status/start_time/intended_date を
// 自由に指定できる appointment 生成ヘルパー（本ファイル専用）。
func makeNextVisitQueryReservation(
	t *testing.T,
	db *gorm.DB,
	clinicID, reservationTypeID uint64,
	ownerID *uint64,
	status model.ReservationStatus,
	start time.Time,
	intendedDate *time.Time,
) *model.Reservation {
	t.Helper()
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		OwnerID:           ownerID,
		ReservationTypeID: reservationTypeID,
		VisitType:         model.VisitTypeRevisit,
		Status:            status,
		Source:            model.ReservationSourceManual,
		CustomerFields:    json.RawMessage(`{}`),
		IntendedDate:      intendedDate,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}

// jstMidnight は t と同じ JST 暦日の 00:00 を返す（intended_date 用）。
func jstMidnight(t time.Time) time.Time {
	j := t.In(config.JST)
	return time.Date(j.Year(), j.Month(), j.Day(), 0, 0, 0, 0, config.JST)
}

// makeNextVisitQueryPet は生存/死亡を指定できる pet 生成ヘルパー（本ファイル専用）。
// 死亡状態は status と deceased_at を独立に指定する（EMR-290/BUG-407 OR 契約の
// 両方のフィールドを検証できるようにするため）。
func makeNextVisitQueryPet(
	t *testing.T,
	db *gorm.DB,
	clinicID, ownerID uint64,
	name string,
	status model.PetStatus,
	deceasedAt *time.Time,
) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "次回検索種_" + name}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: species.ID,
		Name:            name,
		Status:          status,
		DeceasedAt:      deceasedAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

// makeNextVisitQueryPetReservation は pet 紐付け appointment 生成ヘルパー
// （本ファイル専用）。基本ヘルパーで作成後に pet_id を付与する。
func makeNextVisitQueryPetReservation(
	t *testing.T,
	db *gorm.DB,
	clinicID, reservationTypeID uint64,
	ownerID, petID *uint64,
	status model.ReservationStatus,
	start time.Time,
	intendedDate *time.Time,
) *model.Reservation {
	t.Helper()
	res := makeNextVisitQueryReservation(t, db, clinicID, reservationTypeID, ownerID, status, start, intendedDate)
	require.NoError(t, db.WithContext(context.Background()).Model(res).Update("pet_id", petID).Error)
	res.PetID = petID
	return res
}

func TestReservationRepository_FindNextAppointmentForOwner(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	rtA := makeReservationType(t, db, clinicA)
	rtB := makeReservationType(t, db, clinicB)
	ownerA := testdb.MakeTestOwner(t, db, clinicA, "次回来院検索ownerA")
	ownerA2 := testdb.MakeTestOwner(t, db, clinicA, "次回来院検索ownerA2")

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	inTwoDays := now.AddDate(0, 0, 2)
	inThreeDays := now.AddDate(0, 0, 3)
	inTenDays := now.AddDate(0, 0, 10)
	intendedThreeDays := jstMidnight(inThreeDays)

	t.Run("returns nil when owner has no appointments", func(t *testing.T) {
		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, ownerA2.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("confirmed appointment is selected via start_time", func(t *testing.T) {
		want := makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &ownerA.ID,
			model.ReservationStatusConfirmed, inTwoDays, nil)

		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, ownerA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, want.ID, got.ID)
	})

	t.Run("tentative appointment is selected via intended_date", func(t *testing.T) {
		want := makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &ownerA2.ID,
			model.ReservationStatusTentative, inThreeDays, &intendedThreeDays)

		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, ownerA2.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, want.ID, got.ID)
		assert.Equal(t, model.ReservationStatusTentative, got.Status)
	})

	t.Run("earliest semantic date wins across tentative and confirmed", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回来院検索混在owner")
		intendedTomorrow := jstMidnight(now.AddDate(0, 0, 1))
		// confirmed の start_time は3日後、tentative の intended_date は明日。
		makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &owner.ID,
			model.ReservationStatusConfirmed, inThreeDays, nil)
		earlier := makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &owner.ID,
			model.ReservationStatusTentative, intendedTomorrow, &intendedTomorrow)

		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, earlier.ID, got.ID)
	})

	t.Run("past appointment is excluded", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回来院検索過去owner")
		makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &owner.ID,
			model.ReservationStatusConfirmed, yesterday, nil)

		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("terminal statuses are excluded", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回来院検索終端owner")
		for _, status := range []model.ReservationStatus{
			model.ReservationStatusCancelled,
			model.ReservationStatusNoShow,
			model.ReservationStatusCompleted,
		} {
			makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &owner.ID, status, inTenDays, nil)
		}

		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("soft-deleted rows are excluded", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回来院検索削除owner")
		deleted := makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &owner.ID,
			model.ReservationStatusConfirmed, inTwoDays, nil)
		require.NoError(t, db.Delete(deleted).Error)

		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("in-progress appointment on/after today counts as next visit", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回来院検索進行owner")
		want := makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &owner.ID,
			model.ReservationStatusCheckedIn, now, nil)

		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, want.ID, got.ID)
	})

	t.Run("clinic scope: other clinic appointment for same owner id is ignored", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回来院検索他院owner")
		// clinicB の予約は owner_id が同じでも clinicA の検索結果に混ざらない。
		makeNextVisitQueryReservation(t, db, clinicB, rtB.ID, &owner.ID,
			model.ReservationStatusConfirmed, inTwoDays, nil)

		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("deceased pet appointment is skipped in favor of living pet appointment", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回来院検索死亡pet_owner")
		deadPet := makeNextVisitQueryPet(t, db, clinicA, owner.ID, "亡きポチ", model.PetStatusDeceased, &now)
		livePet := makeNextVisitQueryPet(t, db, clinicA, owner.ID, "元気なタマ", model.PetStatusAlive, nil)
		// 死亡ペットの confirmed が早くても skip され、生存ペットの予約が選ばれる。
		makeNextVisitQueryPetReservation(t, db, clinicA, rtA.ID, &owner.ID, &deadPet.ID,
			model.ReservationStatusConfirmed, inTwoDays, nil)
		want := makeNextVisitQueryPetReservation(t, db, clinicA, rtA.ID, &owner.ID, &livePet.ID,
			model.ReservationStatusConfirmed, inThreeDays, nil)

		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, want.ID, got.ID)
	})

	t.Run("owner with only a deceased pet's tentative has no next visit", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "次回来院検索死亡のみowner")
		// deceased_at のみ（status='alive'）でも OR 契約で死亡扱い。
		deadPet := makeNextVisitQueryPet(t, db, clinicA, owner.ID, "亡きミケ", model.PetStatusAlive, &now)
		makeNextVisitQueryPetReservation(t, db, clinicA, rtA.ID, &owner.ID, &deadPet.ID,
			model.ReservationStatusTentative, inThreeDays, &intendedThreeDays)

		got, err := repo.FindNextAppointmentForOwner(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestReservationRepository_FindOwnerIDsWithTentativeIntendedDate(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	rtA := makeReservationType(t, db, clinicA)
	rtB := makeReservationType(t, db, clinicB)

	now := time.Now()
	target := jstMidnight(now.AddDate(0, 0, 5))
	otherDay := jstMidnight(now.AddDate(0, 0, 6))

	matching := testdb.MakeTestOwner(t, db, clinicA, "tentative対象owner")
	wrongDay := testdb.MakeTestOwner(t, db, clinicA, "tentative別日owner")
	cancelled := testdb.MakeTestOwner(t, db, clinicA, "tentative取消owner")
	deleted := testdb.MakeTestOwner(t, db, clinicA, "tentative削除owner")
	promoted := testdb.MakeTestOwner(t, db, clinicA, "実予約昇格owner")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "tentative他院owner")

	makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &matching.ID,
		model.ReservationStatusTentative, target, &target)
	makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &wrongDay.ID,
		model.ReservationStatusTentative, otherDay, &otherDay)
	makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &cancelled.ID,
		model.ReservationStatusCancelled, target, &target)
	softDeleted := makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &deleted.ID,
		model.ReservationStatusTentative, target, &target)
	require.NoError(t, db.Delete(softDeleted).Error)
	// 昇格済み（status は tentative ではない）行は intended_date 一致でも対象外。
	makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &promoted.ID,
		model.ReservationStatusConfirmed, target, &target)
	// 他院クリニックの tentative は clinicA クエリに混ざらない。
	makeNextVisitQueryReservation(t, db, clinicB, rtB.ID, &ownerB.ID,
		model.ReservationStatusTentative, target, &target)
	// owner 未紐付けの tentative は対象外。
	makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, nil,
		model.ReservationStatusTentative, target, &target)

	t.Run("returns only owners with active tentative on the target JST date", func(t *testing.T) {
		got, err := repo.FindOwnerIDsWithTentativeIntendedDate(ctx, clinicA, target)
		require.NoError(t, err)
		assert.Equal(t, []uint64{matching.ID}, got)
	})

	t.Run("clinic isolation: foreign clinic rows are not returned", func(t *testing.T) {
		got, err := repo.FindOwnerIDsWithTentativeIntendedDate(ctx, clinicB, target)
		require.NoError(t, err)
		assert.Equal(t, []uint64{ownerB.ID}, got)
		assert.NotContains(t, got, matching.ID)
	})

	t.Run("no candidates returns empty slice without error", func(t *testing.T) {
		got, err := repo.FindOwnerIDsWithTentativeIntendedDate(ctx, clinicA, jstMidnight(now.AddDate(0, 0, 30)))
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("deceased pet's tentative is excluded from candidates", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "tentative死亡pet_owner")
		deadPet := makeNextVisitQueryPet(t, db, clinicA, owner.ID, "亡きクロ", model.PetStatusDeceased, &now)
		makeNextVisitQueryPetReservation(t, db, clinicA, rtA.ID, &owner.ID, &deadPet.ID,
			model.ReservationStatusTentative, target, &target)

		got, err := repo.FindOwnerIDsWithTentativeIntendedDate(ctx, clinicA, target)
		require.NoError(t, err)
		assert.NotContains(t, got, owner.ID)
	})

	t.Run("owner with only a deceased pet's tentative produces no candidate", func(t *testing.T) {
		owner := testdb.MakeTestOwner(t, db, clinicA, "tentative死亡のみowner")
		// deceased_at のみ（status='alive'）でも OR 契約で死亡扱い。
		deadPet := makeNextVisitQueryPet(t, db, clinicA, owner.ID, "亡きハナ", model.PetStatusAlive, &now)
		otherTarget := jstMidnight(now.AddDate(0, 0, 9))
		makeNextVisitQueryPetReservation(t, db, clinicA, rtA.ID, &owner.ID, &deadPet.ID,
			model.ReservationStatusTentative, otherTarget, &otherTarget)

		got, err := repo.FindOwnerIDsWithTentativeIntendedDate(ctx, clinicA, otherTarget)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestReservationRepository_FindTentativeByOwnerAndIntendedDate(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	rtA := makeReservationType(t, db, clinicA)

	now := time.Now()
	target := jstMidnight(now.AddDate(0, 0, 7))
	otherDay := jstMidnight(now.AddDate(0, 0, 8))

	owner := testdb.MakeTestOwner(t, db, clinicA, "再検証owner")
	tentative := makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &owner.ID,
		model.ReservationStatusTentative, target, &target)

	t.Run("returns the matching active tentative", func(t *testing.T) {
		got, err := repo.FindTentativeByOwnerAndIntendedDate(ctx, clinicA, owner.ID, target)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, tentative.ID, got.ID)
	})

	t.Run("cancelled tentative is stale and returns nil", func(t *testing.T) {
		staleOwner := testdb.MakeTestOwner(t, db, clinicA, "再検証取消owner")
		makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &staleOwner.ID,
			model.ReservationStatusCancelled, target, &target)

		got, err := repo.FindTentativeByOwnerAndIntendedDate(ctx, clinicA, staleOwner.ID, target)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("promoted reservation is stale and returns nil", func(t *testing.T) {
		promotedOwner := testdb.MakeTestOwner(t, db, clinicA, "再検証昇格owner")
		makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &promotedOwner.ID,
			model.ReservationStatusConfirmed, target, &target)

		got, err := repo.FindTentativeByOwnerAndIntendedDate(ctx, clinicA, promotedOwner.ID, target)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("soft-deleted tentative is stale and returns nil", func(t *testing.T) {
		deletedOwner := testdb.MakeTestOwner(t, db, clinicA, "再検証削除owner")
		res := makeNextVisitQueryReservation(t, db, clinicA, rtA.ID, &deletedOwner.ID,
			model.ReservationStatusTentative, target, &target)
		require.NoError(t, db.Delete(res).Error)

		got, err := repo.FindTentativeByOwnerAndIntendedDate(ctx, clinicA, deletedOwner.ID, target)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("date mismatch returns nil", func(t *testing.T) {
		got, err := repo.FindTentativeByOwnerAndIntendedDate(ctx, clinicA, owner.ID, otherDay)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("owner mismatch returns nil", func(t *testing.T) {
		other := testdb.MakeTestOwner(t, db, clinicA, "再検証別owner")
		got, err := repo.FindTentativeByOwnerAndIntendedDate(ctx, clinicA, other.ID, target)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("clinic mismatch returns nil", func(t *testing.T) {
		got, err := repo.FindTentativeByOwnerAndIntendedDate(ctx, clinicB, owner.ID, target)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("deceased pet tentative is stale and returns nil", func(t *testing.T) {
		// 配信直前再検証でも死亡ペットの tentative は stale 扱い（→ excluded）。
		staleOwner := testdb.MakeTestOwner(t, db, clinicA, "再検証死亡pet_owner")
		deadPet := makeNextVisitQueryPet(t, db, clinicA, staleOwner.ID, "亡きブチ", model.PetStatusDeceased, &now)
		makeNextVisitQueryPetReservation(t, db, clinicA, rtA.ID, &staleOwner.ID, &deadPet.ID,
			model.ReservationStatusTentative, target, &target)

		got, err := repo.FindTentativeByOwnerAndIntendedDate(ctx, clinicA, staleOwner.ID, target)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
