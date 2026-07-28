package reservation

// staff_preload_clinic_isolation_test.go
// クロステナント READ IDOR remediation follow-up — (d) Staff(Doctor)preload の多医院所属対応。
//
// staff は staff_clinic_assignments による多医院所属のため、staffs.clinic_id（主所属）単純スコープでは
// 共有医師を誤って隠す。reservation の Doctor/CreatedByStaff は assignment-EXISTS でスコープし、
//   (i)  共有医師（主所属が別 clinic でも当該 clinic に配属済み）は表示される（非破壊）
//   (ii) 当該 clinic に未配属のスタッフを指す予約は親行ごと fail-closed になる
//   (iii)#86 認可集合に staff の配属 clinic が含まれても、予約自身の clinic と相関しなければ
//        親行ごと fail-closed になる
// を満たすこと。注: 既往カルテ等の履歴 preload は退職スタッフ名の表示を保つため意図的に scope しない
// （repository/CLAUDE.md 参照）。本テストは現在/未来データである reservation のみを対象とする。

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// seedClinicsForFK は staff_clinic_assignments → clinics FK を満たすため clinics を seed する。
func seedClinicsForFK(t *testing.T, db *gorm.DB, clinicIDs ...uint64) {
	t.Helper()
	company := &model.Company{Name: "テスト法人(staff preload)"}
	require.NoError(t, db.WithContext(context.Background()).Create(company).Error)
	for _, cid := range clinicIDs {
		c := &model.Clinic{ID: cid, CompanyID: company.ID, Name: fmt.Sprintf("テスト医院%d", cid)}
		require.NoError(t, db.WithContext(context.Background()).Clauses(clause.OnConflict{DoNothing: true}).Create(c).Error)
	}
}

func makeStaffClinicAssignment(t *testing.T, db *gorm.DB, staffID, clinicID uint64) {
	t.Helper()
	a := &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}
	require.NoError(t, db.WithContext(context.Background()).Create(a).Error)
}

func makeReservationWithDoctor(t *testing.T, db *gorm.DB, clinicID, reservationTypeID, doctorID uint64) *model.Reservation {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Minute)
	did := doctorID
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         now,
		EndTime:           now.Add(15 * time.Minute),
		ReservationTypeID: reservationTypeID,
		DoctorID:          &did,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}

func TestReservationRepository_DoctorPreload_MultiClinicStaffIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	db.Exec("TRUNCATE TABLE appointments CASCADE")
	db.Exec("TRUNCATE TABLE staff_clinic_assignments CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{}, &model.ReservationType{}, &model.Reservation{}))
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	seedClinicsForFK(t, db, clinicA, clinicB)
	rtA := makeReservationType(t, db, clinicA)

	// 共有医師: 主所属は clinic B だが clinic A に配属済み。
	staffShared := makeDoctor(t, db, clinicB, "共有医師")
	makeStaffClinicAssignment(t, db, staffShared.ID, clinicA)
	makeStaffClinicAssignment(t, db, staffShared.ID, clinicB)

	// B 単独医師: clinic B のみ配属。clinic A の予約に紐づけても表示されてはならない。
	staffBOnly := makeDoctor(t, db, clinicB, "B単独医師")
	makeStaffClinicAssignment(t, db, staffBOnly.ID, clinicB)

	resShared := makeReservationWithDoctor(t, db, clinicA, rtA.ID, staffShared.ID)
	resBOnly := makeReservationWithDoctor(t, db, clinicA, rtA.ID, staffBOnly.ID) // 別テナント単独スタッフFKを植え付け

	// (i) 共有医師は clinic A に配属済みなので表示される（多医院所属の非破壊）
	gotShared, err := repo.FindByID(ctx, clinicA, resShared.ID)
	require.NoError(t, err)
	require.NotNil(t, gotShared.Doctor, "主所属が B でも A 配属の共有医師は表示されるべき")
	assert.Equal(t, staffShared.ID, gotShared.Doctor.ID)

	// (ii) clinic A 未配属のスタッフを指す予約は親行ごと fail-closed になる
	gotBOnly, err := repo.FindByID(ctx, clinicA, resBOnly.ID)
	require.Error(t, err)
	assert.Nil(t, gotBOnly)
	assert.True(t, apperrors.IsNotFound(err), "clinic A に未配属のスタッフを指す予約は NotFound にする: %v", err)

	// (iii) #86: 認可集合 [A,B] でも、clinic A の予約と clinic B 単独医師は相関しない
	gotForBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, resBOnly.ID)
	require.Error(t, err)
	assert.Nil(t, gotForBoth)
	assert.True(t, apperrors.IsNotFound(err), "複数医院認可でも関連は予約自身の clinic と相関させる: %v", err)
}
