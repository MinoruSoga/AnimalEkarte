package medicalrecord

// preload_followup_clinic_isolation_test.go
// クロステナント READ IDOR remediation follow-up — (a) 既修正だが専用テストの無かった3 repo の
// master Preload clinic 隔離回帰テスト（examination / trimming / reservation FindAllByCategory）。
//
// 各テストは別クリニックのマスタを指す FK を植え付け、clinic_id スコープで
// 別クリニックのマスタが応答に混入しないことを検証する。必須マスタを指す
// Examination は、現在の relation scope により汚染行そのものを fail-closed で除外する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// --- (a1) examination: ExaminationType ---

func makePreloadExamRec(t *testing.T, db *gorm.DB, clinicID, medicalRecordID, petID, doctorID, examTypeID uint64) uint64 {
	t.Helper()
	mrid := medicalRecordID
	pid := petID
	did := doctorID
	e := &model.Examination{
		ClinicID:        clinicID,
		MedicalRecordID: &mrid,
		PetID:           &pid,
		DoctorID:        &did,
		ExamTypeID:      examTypeID,
		Date:            time.Now(),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(e).Error)
	return e.ID
}

func TestExaminationRepository_FindByID_ExaminationTypePreloadClinicIsolation(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.ExaminationType{},
		&model.Examination{},
		&model.ExamResult{},
	))
	db.Exec("TRUNCATE TABLE exam_results CASCADE")
	db.Exec("TRUNCATE TABLE exams CASCADE")
	db.Exec("TRUNCATE TABLE exam_types CASCADE")
	db.Exec("TRUNCATE TABLE staff_clinic_assignments CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	petA, mrA, doctorA := makeClinicScopedClinicalReadParents(t, db, clinicA, "検査")
	typeB := &model.ExaminationType{ClinicID: clinicB, Name: "医院Bの検査種別"}
	require.NoError(t, db.WithContext(ctx).Create(typeB).Error)
	typeA := &model.ExaminationType{ClinicID: clinicA, Name: "医院Aの検査種別"}
	require.NoError(t, db.WithContext(ctx).Create(typeA).Error)

	crossID := makePreloadExamRec(t, db, clinicA, mrA.ID, petA.ID, doctorA.ID, typeB.ID)
	legitID := makePreloadExamRec(t, db, clinicA, mrA.ID, petA.ID, doctorA.ID, typeA.ID)

	tests := []struct {
		name         string
		id           uint64
		wantNotFound bool
		wantTypeID   uint64
	}{
		{
			name:         "別クリニックの必須検査種別を指す行は取得対象外",
			id:           crossID,
			wantNotFound: true,
		},
		{
			name:       "同一クリニックの検査種別と整合した患者医師関係を取得",
			id:         legitID,
			wantTypeID: typeA.ID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindByID(ctx, clinicA, tt.id)
			if tt.wantNotFound {
				require.Error(t, err)
				assert.True(t, apperrors.IsNotFound(err))
				assert.Nil(t, got, "別クリニックの検査種別を参照する行を返してはならない")
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got.ExaminationType, "同一クリニックの検査種別は Preload されるべき")
			assert.Equal(t, tt.wantTypeID, got.ExaminationType.ID)
			require.NotNil(t, got.Pet)
			require.NotNil(t, got.Pet.Owner)
			require.NotNil(t, got.Doctor)
			assert.Equal(t, doctorA.ID, got.Doctor.ID)
		})
	}
}
