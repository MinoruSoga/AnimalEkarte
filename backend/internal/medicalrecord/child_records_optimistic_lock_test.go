package medicalrecord

// child_records_optimistic_lock_test.go — UAT-R2-EXCLUSIVE-LOCK: 治療・バイタル・処方・接種の
// 子レコード Update に対する楽観的ロック（行 version 列 + expectedVersion CAS）のテスト。
// 対照実装は clinical_plan_repository_optimistic_lock_test.go（BUG-416③）と同型:
//
//	(a) expectedVersion 一致 → 更新成功、version は +1 される
//	(b) expectedVersion 不一致（stale）→ Conflict、書込は反映されない
//	(c) expectedVersion == nil → 従来どおり無条件更新（照合スキップ、後方互換）
//
// 4 経路とも input.Version が repository Update の CAS 述語としてそのまま渡り、
// repo の Conflict が service の返り値に伝播することを mock で検証する。
// vaccination には設計票（docs/work/todo-campaign-20260918/UAT-R2-EXCLUSIVE-LOCK.md
// §最小設計）の draft ガード（medical_record_id != nil のとき lockDraftMedicalRecord）を
// 検証する TestVaccinationService_Update_FinalizedMedicalRecordRejected を併記する。
//
// testdb 注記: version 列は backend/migrations/005_*.sql（未適用・別途ユーザー作業）ではなく、
// model の `gorm:"default:1"` タグ経由の AutoMigrate（testdb.SetupTestDB /
// EnsureAutoMigrated）でテスト DB に供給される。testdb は 001_init.sql を適用しない
// GORM スキーマ double のため、実 DB の version 列は migration 適用後にのみ存在する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- 治療 (treatments) ----

func TestTreatmentRepository_Update_OptimisticLock(t *testing.T) {
	db := setupTreatmentHistoryTestDB(t)
	repo := NewTreatmentRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	makeRecord := func(recordNo string) *model.MedicalRecord {
		t.Helper()
		mr := &model.MedicalRecord{
			ClinicID: clinicA,
			RecordNo: recordNo,
			Date:     time.Now(),
			Status:   model.MedicalRecordStatusDraft,
		}
		require.NoError(t, db.WithContext(ctx).Create(mr).Error)
		return mr
	}

	t.Run("expectedVersion一致で更新成功しversionが+1される", func(t *testing.T) {
		mr := makeRecord("MR-TR-LOCK-001")
		tr := &model.Treatment{MedicalRecordID: mr.ID, ItemType: model.TreatmentItemTypeOther, Content: "更新前"}
		require.NoError(t, repo.Create(ctx, tr))
		require.Equal(t, 1, tr.Version, "新規作成時の初期バージョンは1であるべき")

		expected := 1
		content := "更新後"
		err := repo.Update(ctx, clinicA, tr.ID, UpdateTreatmentInput{Content: &content, Version: &expected})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.Content)
		assert.Equal(t, 2, got.Version, "更新成功時はversionが+1されているべき")
	})

	t.Run("expectedVersionが古い(stale)場合はConflictで書込は反映されない", func(t *testing.T) {
		mr := makeRecord("MR-TR-LOCK-002")
		tr := &model.Treatment{MedicalRecordID: mr.ID, ItemType: model.TreatmentItemTypeOther, Content: "更新前"}
		require.NoError(t, repo.Create(ctx, tr))

		// 先に一度更新してversionを2に進める（他ユーザーの書込みをシミュレート）
		current := 1
		other := "他ユーザーによる更新"
		require.NoError(t, repo.Update(ctx, clinicA, tr.ID, UpdateTreatmentInput{Content: &other, Version: &current}))

		// stale な expectedVersion（1のまま）で更新を試みる
		stale := 1
		conflictContent := "競合更新"
		err := repo.Update(ctx, clinicA, tr.ID, UpdateTreatmentInput{Content: &conflictContent, Version: &stale})

		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "version不一致はConflictであるべき: %v", err)
		assert.Contains(t, err.Error(), "他のユーザーがこの治療を変更しました", "version不一致は専用メッセージであるべき")

		got, err := repo.FindByID(ctx, clinicA, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, "他ユーザーによる更新", got.Content, "stale版の更新は反映されてはならない")
		assert.Equal(t, 2, got.Version)
	})

	t.Run("expectedVersionがnilの場合は実際のversionを問わず無条件更新される(後方互換)", func(t *testing.T) {
		mr := makeRecord("MR-TR-LOCK-003")
		tr := &model.Treatment{MedicalRecordID: mr.ID, ItemType: model.TreatmentItemTypeOther, Content: "更新前"}
		require.NoError(t, repo.Create(ctx, tr))

		// version を明示的に進めておく
		v1 := 1
		mid := "中間更新"
		require.NoError(t, repo.Update(ctx, clinicA, tr.ID, UpdateTreatmentInput{Content: &mid, Version: &v1}))

		// Version=nil（照合スキップ）は実際のversion(2)と無関係に成功するべき
		skip := "照合スキップ更新"
		err := repo.Update(ctx, clinicA, tr.ID, UpdateTreatmentInput{Content: &skip})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, tr.ID)
		require.NoError(t, err)
		assert.Equal(t, "照合スキップ更新", got.Content)
	})
}

func TestTreatmentService_Update_StaleExpectedVersionConflict(t *testing.T) {
	const clinicID = uint64(1)
	var gotVersion *int
	repo := &mockTreatmentRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Treatment, error) {
			return &model.Treatment{ID: 1, MedicalRecordID: 1, Version: 2}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, cmd UpdateTreatmentInput) error {
			gotVersion = cmd.Version
			return apperrors.WrapConflict("他のユーザーがこの治療を変更しました。再読み込みしてください")
		},
	}
	svc := newTreatmentSvc(repo, draftMedicalRecordRepo(), &mockInventoryRepository{}, nil)

	stale := 1
	content := "競合更新"
	updated, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateTreatmentInput{Content: &content, Version: &stale})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "repo の stale Conflict は service の返り値に伝播するべき: %v", err)
	assert.Nil(t, updated)
	require.NotNil(t, gotVersion, "input.Version は repository Update へそのまま渡るべき")
	assert.Equal(t, 1, *gotVersion)
}

// ---- バイタル (vital_records) ----

func TestVitalRepository_Update_OptimisticLock(t *testing.T) {
	db := setupVitalTestDB(t)
	repo := NewVitalRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "バイタル飼主")
	species := makeVitalSpecies(t, db, "犬")
	pet := makeVitalPet(t, db, clinicA, owner.ID, species.ID, "バイタルペット")

	makeVital := func() *model.VitalRecord {
		t.Helper()
		v := &model.VitalRecord{
			ClinicID:   clinicA,
			PetID:      pet.ID,
			RecordedAt: time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC),
			Notes:      "更新前",
		}
		require.NoError(t, repo.Create(ctx, v))
		return v
	}

	t.Run("expectedVersion一致で更新成功しversionが+1される", func(t *testing.T) {
		vital := makeVital()
		require.Equal(t, 1, vital.Version, "新規作成時の初期バージョンは1であるべき")

		expected := 1
		notes := "更新後"
		err := repo.Update(ctx, clinicA, vital.ID, UpdateVitalInput{Notes: &notes, Version: &expected})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, vital.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.Notes)
		assert.Equal(t, 2, got.Version, "更新成功時はversionが+1されているべき")
	})

	t.Run("expectedVersionが古い(stale)場合はConflictで書込は反映されない", func(t *testing.T) {
		vital := makeVital()

		current := 1
		other := "他ユーザーによる更新"
		require.NoError(t, repo.Update(ctx, clinicA, vital.ID, UpdateVitalInput{Notes: &other, Version: &current}))

		stale := 1
		conflictNotes := "競合更新"
		err := repo.Update(ctx, clinicA, vital.ID, UpdateVitalInput{Notes: &conflictNotes, Version: &stale})

		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "version不一致はConflictであるべき: %v", err)
		assert.Contains(t, err.Error(), "他のユーザーがこのバイタルを変更しました", "version不一致は専用メッセージであるべき")

		got, err := repo.FindByID(ctx, clinicA, vital.ID)
		require.NoError(t, err)
		assert.Equal(t, "他ユーザーによる更新", got.Notes, "stale版の更新は反映されてはならない")
		assert.Equal(t, 2, got.Version)
	})

	t.Run("expectedVersionがnilの場合は実際のversionを問わず無条件更新される(後方互換)", func(t *testing.T) {
		vital := makeVital()

		v1 := 1
		mid := "中間更新"
		require.NoError(t, repo.Update(ctx, clinicA, vital.ID, UpdateVitalInput{Notes: &mid, Version: &v1}))

		skip := "照合スキップ更新"
		err := repo.Update(ctx, clinicA, vital.ID, UpdateVitalInput{Notes: &skip})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, vital.ID)
		require.NoError(t, err)
		assert.Equal(t, "照合スキップ更新", got.Notes)
	})
}

func TestVitalService_Update_StaleExpectedVersionConflict(t *testing.T) {
	const clinicID = uint64(1)
	var gotVersion *int
	repo := &mockVitalRepository{
		findByIDFn: func(_ context.Context, _ uint64, _ uint64) (*model.VitalRecord, error) {
			return &model.VitalRecord{ID: 1, ClinicID: clinicID, PetID: 10, MedicalRecordID: ptrUint64(1), Version: 2}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, cmd UpdateVitalInput) error {
			gotVersion = cmd.Version
			return apperrors.WrapConflict("他のユーザーがこのバイタルを変更しました。再読み込みしてください")
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID: 1, ClinicID: clinicID, OwnerID: ptrUint64(100), PetID: ptrUint64(10),
				Status: model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	svc := NewVitalServiceWithRelationValidation(
		repo, mrRepo, okVitalAudit(), validVitalRelations(10, 100), nil, nil, &mockCheckupTransactor{},
	)

	stale := 1
	notes := "競合更新"
	updated, err := svc.Update(context.Background(), clinicID, 1, 1, &UpdateVitalInput{Notes: &notes, Version: &stale})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "repo の stale Conflict は service の返り値に伝播するべき: %v", err)
	assert.Nil(t, updated)
	require.NotNil(t, gotVersion, "input.Version は repository Update へそのまま渡るべき")
	assert.Equal(t, 1, *gotVersion)
}

// ---- 処方 (prescriptions) ----

func TestPrescriptionRepository_Update_OptimisticLock(t *testing.T) {
	db := setupPrescriptionTestDB(t)
	repo := NewPrescriptionRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "処方飼主")
	prescribedAt := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)

	t.Run("expectedVersion一致で更新成功しversionが+1される", func(t *testing.T) {
		rx := makePrescription(t, db, clinicA, owner.ID, nil, prescribedAt)
		require.Equal(t, 1, rx.Version, "新規作成時の初期バージョンは1であるべき")

		expected := 1
		days := 14
		err := repo.Update(ctx, clinicA, rx.ID, UpdatePrescriptionInput{DurationDays: &days, Version: &expected})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, rx.ID)
		require.NoError(t, err)
		assert.Equal(t, 14, got.DurationDays)
		assert.Equal(t, 2, got.Version, "更新成功時はversionが+1されているべき")
	})

	t.Run("expectedVersionが古い(stale)場合はConflictで書込は反映されない", func(t *testing.T) {
		rx := makePrescription(t, db, clinicA, owner.ID, nil, prescribedAt)

		current := 1
		otherDays := 10
		require.NoError(t, repo.Update(ctx, clinicA, rx.ID, UpdatePrescriptionInput{DurationDays: &otherDays, Version: &current}))

		stale := 1
		conflictDays := 30
		err := repo.Update(ctx, clinicA, rx.ID, UpdatePrescriptionInput{DurationDays: &conflictDays, Version: &stale})

		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "version不一致はConflictであるべき: %v", err)
		assert.Contains(t, err.Error(), "他のユーザーがこの処方を変更しました", "version不一致は専用メッセージであるべき")

		got, err := repo.FindByID(ctx, clinicA, rx.ID)
		require.NoError(t, err)
		assert.Equal(t, 10, got.DurationDays, "stale版の更新は反映されてはならない")
		assert.Equal(t, 2, got.Version)
	})

	t.Run("expectedVersionがnilの場合は実際のversionを問わず無条件更新される(後方互換)", func(t *testing.T) {
		rx := makePrescription(t, db, clinicA, owner.ID, nil, prescribedAt)

		v1 := 1
		midDays := 10
		require.NoError(t, repo.Update(ctx, clinicA, rx.ID, UpdatePrescriptionInput{DurationDays: &midDays, Version: &v1}))

		skipDays := 21
		err := repo.Update(ctx, clinicA, rx.ID, UpdatePrescriptionInput{DurationDays: &skipDays})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, rx.ID)
		require.NoError(t, err)
		assert.Equal(t, 21, got.DurationDays)
	})
}

func TestPrescriptionService_Update_StaleExpectedVersionConflict(t *testing.T) {
	const clinicID = uint64(1)
	medicalRecordID := uint64(2)
	var gotVersion *int
	repo := &mockPrescriptionRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Prescription, error) {
			return &model.Prescription{ID: 3, MedicalRecordID: &medicalRecordID, Version: 2}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, cmd UpdatePrescriptionInput) error {
			gotVersion = cmd.Version
			return apperrors.WrapConflict("他のユーザーがこの処方を変更しました。再読み込みしてください")
		},
	}
	medRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewPrescriptionService(repo, medRecordRepo, nil, &mockCheckupTransactor{})

	stale := 1
	days := 14
	updated, err := svc.Update(context.Background(), clinicID, medicalRecordID, 3,
		&UpdatePrescriptionInput{DurationDays: &days, Version: &stale})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "repo の stale Conflict は service の返り値に伝播するべき: %v", err)
	assert.Nil(t, updated)
	require.NotNil(t, gotVersion, "input.Version は repository Update へそのまま渡るべき")
	assert.Equal(t, 1, *gotVersion)
}

// ---- 接種 (vaccinations) ----

func TestVaccinationRepository_Update_OptimisticLock(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "接種飼主")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "接種ペット")
	vaccine := makeVaccineMaster(t, db, clinicA, "混合ワクチン")
	ensureVaccinationTestClinics(t, db, clinicA)

	t.Run("expectedVersion一致で更新成功しversionが+1される", func(t *testing.T) {
		rec := makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine.ID)
		require.Equal(t, 1, rec.Version, "新規作成時の初期バージョンは1であるべき")

		expected := 1
		remarks := "更新後"
		updated, err := repo.Update(ctx, clinicA, rec.ID, UpdateVaccinationInput{Remarks: &remarks, Version: &expected})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "更新後", updated.Remarks)
		assert.Equal(t, 2, updated.Version, "更新成功時はversionが+1されているべき")
	})

	t.Run("expectedVersionが古い(stale)場合はConflictで書込は反映されない", func(t *testing.T) {
		rec := makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine.ID)

		current := 1
		other := "他ユーザーによる更新"
		_, err := repo.Update(ctx, clinicA, rec.ID, UpdateVaccinationInput{Remarks: &other, Version: &current})
		require.NoError(t, err)

		stale := 1
		conflictRemarks := "競合更新"
		updated, err := repo.Update(ctx, clinicA, rec.ID, UpdateVaccinationInput{Remarks: &conflictRemarks, Version: &stale})

		require.Error(t, err)
		assert.Nil(t, updated)
		assert.True(t, apperrors.IsConflict(err), "version不一致はConflictであるべき: %v", err)
		assert.Contains(t, err.Error(), "他のユーザーがこのワクチン接種を変更しました", "version不一致は専用メッセージであるべき")

		got, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, "他ユーザーによる更新", got.Remarks, "stale版の更新は反映されてはならない")
		assert.Equal(t, 2, got.Version)
	})

	t.Run("expectedVersionがnilの場合は実際のversionを問わず無条件更新される(後方互換)", func(t *testing.T) {
		rec := makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine.ID)

		v1 := 1
		mid := "中間更新"
		_, err := repo.Update(ctx, clinicA, rec.ID, UpdateVaccinationInput{Remarks: &mid, Version: &v1})
		require.NoError(t, err)

		skip := "照合スキップ更新"
		updated, err := repo.Update(ctx, clinicA, rec.ID, UpdateVaccinationInput{Remarks: &skip})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "照合スキップ更新", updated.Remarks)
	})
}

func TestVaccinationService_Update_StaleExpectedVersionConflict(t *testing.T) {
	const clinicID = uint64(1)
	const vaccinationID = uint64(1)
	var gotVersion *int
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: vaccinationID, ClinicID: clinicID, VaccineID: 1, Version: 2}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateVaccinationInput) (*model.Vaccination, error) {
			gotVersion = cmd.Version
			return nil, apperrors.WrapConflict("他のユーザーがこのワクチン接種を変更しました。再読み込みしてください")
		},
	}
	svc := NewVaccinationService(
		repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{},
		&permissiveVaccinationMedicalRecordLocker{},
		vaccinationTestTransactor{},
	)

	stale := 1
	remarks := "競合更新"
	updated, err := svc.Update(context.Background(), clinicID, vaccinationID,
		&UpdateVaccinationInput{Remarks: &remarks, Version: &stale})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "repo の stale Conflict は service の返り値に伝播するべき: %v", err)
	assert.Nil(t, updated)
	require.NotNil(t, gotVersion, "input.Version は repository Update へそのまま渡るべき")
	assert.Equal(t, 1, *gotVersion)
}

// TestVaccinationService_Update_FinalizedMedicalRecordRejected は設計票 §最小設計の
// 接種 draft ガードを検証する: medical_record_id != nil の接種更新は、紐付くカルテが
// 確定済みなら Conflict で拒否される（治療/バイタル/処方の lockDraftMedicalRecord と対称化）。
func TestVaccinationService_Update_FinalizedMedicalRecordRejected(t *testing.T) {
	const clinicID = uint64(1)
	const vaccinationID = uint64(1)
	recordID := uint64(30)
	updateCalls := 0
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return &model.Vaccination{
				ID: vaccinationID, ClinicID: clinicID, VaccineID: 1,
				MedicalRecordID: &recordID,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateVaccinationInput) (*model.Vaccination, error) {
			updateCalls++
			return &model.Vaccination{ID: vaccinationID, ClinicID: clinicID}, nil
		},
	}
	locker := &vaccinationMedicalRecordLockerStub{records: map[uint64]*model.MedicalRecord{
		recordID: {
			ID: recordID, ClinicID: clinicID,
			Status: model.MedicalRecordStatusFinalized,
		},
	}}
	svc := NewVaccinationService(
		repo, okVaccineRepo(), nil,
		&permissiveVaccinationRelationVerifier{},
		locker,
		vaccinationTestTransactor{},
	)

	remarks := "確定後の書込"
	updated, err := svc.Update(context.Background(), clinicID, vaccinationID, &UpdateVaccinationInput{Remarks: &remarks})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "確定済みカルテ紐付け接種の更新はConflictであるべき: %v", err)
	assert.Nil(t, updated)
	assert.Zero(t, updateCalls, "確定済みカルテへの書込は repository Update まで到達してはならない")
}
