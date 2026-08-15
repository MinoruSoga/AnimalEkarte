package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// livingPetFinder は Create 成功系 fixture 用の生存ペット stub。
func livingPetFinder() petFinder {
	return &mockPetRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id}, nil
		},
	}
}

// draftMedicalRecordWithPet は画像 Create の親 draft カルテ（PetID 付き）を返す。
func draftMedicalRecordWithPet(clinicID, recordID, petID uint64) *model.MedicalRecord {
	return &model.MedicalRecord{
		ID:       recordID,
		ClinicID: clinicID,
		PetID:    &petID,
		Status:   model.MedicalRecordStatusDraft,
	}
}

// ---- MedicalRecordImage モック ----

type mockMedicalRecordImageRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error)
	findByIDFn              func(ctx context.Context, clinicID, imageID uint64) (*model.MedicalRecordImage, error)
	createFn                func(ctx context.Context, image *model.MedicalRecordImage) error
	deleteFn                func(ctx context.Context, clinicID, imageID uint64) error
}

func (m *mockMedicalRecordImageRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockMedicalRecordImageRepository) FindByID(ctx context.Context, clinicID, imageID uint64) (*model.MedicalRecordImage, error) {
	return m.findByIDFn(ctx, clinicID, imageID)
}

func (m *mockMedicalRecordImageRepository) Create(ctx context.Context, image *model.MedicalRecordImage) error {
	return m.createFn(ctx, image)
}

func (m *mockMedicalRecordImageRepository) Delete(ctx context.Context, clinicID, imageID uint64) error {
	return m.deleteFn(ctx, clinicID, imageID)
}

type clinicalStaffLockerStub struct {
	staff *model.Staff
	err   error
	calls int
}

func (s *clinicalStaffLockerStub) LockActiveByIDForShare(_ context.Context, id uint64) (*model.Staff, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	if s.staff == nil {
		return nil, nil
	}
	staff := *s.staff
	if staff.ID == 0 {
		staff.ID = id
	}
	return &staff, nil
}

type clinicalStaffAssignmentLockerStub struct {
	assignment *model.StaffClinicAssignment
	err        error
	calls      int
}

func (s *clinicalStaffAssignmentLockerStub) LockActiveByStaffAndClinic(
	_ context.Context,
	staffID, clinicID uint64,
) (*model.StaffClinicAssignment, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	if s.assignment == nil {
		return nil, nil
	}
	assignment := *s.assignment
	if assignment.StaffID == 0 {
		assignment.StaffID = staffID
	}
	if assignment.ClinicID == 0 {
		assignment.ClinicID = clinicID
	}
	return &assignment, nil
}

type medicalRecordImageExaminationFinderStub struct {
	exam  *model.Examination
	err   error
	calls int
}

func (s *medicalRecordImageExaminationFinderStub) FindByID(
	_ context.Context,
	_, _ uint64,
) (*model.Examination, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	if s.exam == nil {
		return nil, nil
	}
	exam := *s.exam
	return &exam, nil
}

// ---- Tests ----

func TestMedicalRecordImageService_List(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		repoImages      []model.MedicalRecordImage
		repoErr         error
		wantLen         int
		wantErr         bool
	}{
		{
			name:            "returns record images for medical record",
			medicalRecordID: 1,
			repoImages: []model.MedicalRecordImage{
				{ID: 1, MedicalRecordID: 1, ImageType: model.MedicalImageTypeXray, FileName: "xray1.jpg"},
				{ID: 2, MedicalRecordID: 1, ImageType: model.MedicalImageTypeEcho, FileName: "ultrasound1.jpg"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:            "returns empty list when no images exist",
			medicalRecordID: 999,
			repoImages:      []model.MedicalRecordImage{},
			repoErr:         nil,
			wantLen:         0,
			wantErr:         false,
		},
		{
			name:            "propagates repository error",
			medicalRecordID: 1,
			repoImages:      nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordImageRepository{
				listByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.MedicalRecordImage, error) {
					return tt.repoImages, tt.repoErr
				},
			}
			medRecRepo := &mockMedicalRecordRepository{}
			svc := NewMedicalRecordImageService(repo, medRecRepo, &mockCheckupTransactor{})

			images, err := svc.List(context.Background(), 1, tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, images, tt.wantLen)
			}
		})
	}
}

func TestMedicalRecordImageService_Create(t *testing.T) {
	takenAt := time.Now()
	examID := uint64(1)
	staffID := uint64(5)

	tests := []struct {
		name            string
		medicalRecordID uint64
		input           *CreateMedicalRecordImageInput
		repoErr         error
		wantErr         bool
	}{
		{
			name:            "creates record image with explicit image type",
			medicalRecordID: 1,
			input: &CreateMedicalRecordImageInput{
				ImageURL:     "https://example.com/image1.jpg",
				ThumbnailURL: "https://example.com/thumb1.jpg",
				FileName:     "image1.jpg",
				FileSize:     1024000,
				MimeType:     "image/jpeg",
				ImageType:    model.MedicalImageTypeXray,
				Description:  "X-ray of chest",
				TakenAt:      &takenAt,
				ExamID:       &examID,
				StaffID:      &staffID,
				SortOrder:    1,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "creates record image with default image type when empty",
			medicalRecordID: 1,
			input: &CreateMedicalRecordImageInput{
				ImageURL:     "https://example.com/image2.jpg",
				ThumbnailURL: "https://example.com/thumb2.jpg",
				FileName:     "image2.jpg",
				FileSize:     2048000,
				MimeType:     "image/jpeg",
				ImageType:    "", // Will default to Other
				Description:  "Generic image",
				SortOrder:    2,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "returns error on invalid image type",
			medicalRecordID: 1,
			input: &CreateMedicalRecordImageInput{
				ImageURL:  "https://example.com/image3.jpg",
				FileName:  "image3.jpg",
				ImageType: model.MedicalImageType("invalid_type"),
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:            "returns error when repository fails",
			medicalRecordID: 1,
			input: &CreateMedicalRecordImageInput{
				ImageURL:  "https://example.com/image4.jpg",
				FileName:  "image4.jpg",
				ImageType: model.MedicalImageTypeEcho,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	const petID = uint64(7)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordImageRepository{
				createFn: func(_ context.Context, _ *model.MedicalRecordImage) error {
					return tt.repoErr
				},
			}
			medRecRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					// stub: always return a valid draft parent record (LockByIDForUpdate delegates to FindByID)
					return draftMedicalRecordWithPet(clinicID, id, petID), nil
				},
			}
			svc := NewMedicalRecordImageServiceWithRelationValidation(
				repo,
				medRecRepo,
				livingPetFinder(),
				&medicalRecordImageExaminationFinderStub{exam: &model.Examination{
					ID: examID, ClinicID: 1, MedicalRecordID: ptrUint64(tt.medicalRecordID),
				}},
				&clinicalStaffLockerStub{staff: &model.Staff{ID: staffID, IsActive: true}},
				&clinicalStaffAssignmentLockerStub{
					assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: 1},
				},
				&mockCheckupTransactor{},
			)

			image, err := svc.Create(context.Background(), 1, tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, image)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, image)
				assert.Equal(t, tt.medicalRecordID, image.MedicalRecordID)
			}
		})
	}
}

func TestMedicalRecordImageService_Delete(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		imageID         uint64
		repoImage       *model.MedicalRecordImage
		findByIDErr     error
		deleteErr       error
		wantErr         bool
	}{
		{
			name:            "deletes record image successfully",
			medicalRecordID: 1,
			imageID:         1,
			repoImage: &model.MedicalRecordImage{
				ID:              1,
				MedicalRecordID: 1,
				FileName:        "image1.jpg",
			},
			findByIDErr: nil,
			deleteErr:   nil,
			wantErr:     false,
		},
		{
			name:            "returns not found when image does not belong to medical record",
			medicalRecordID: 1,
			imageID:         999,
			repoImage: &model.MedicalRecordImage{
				ID:              999,
				MedicalRecordID: 2, // Different medical record
				FileName:        "image999.jpg",
			},
			findByIDErr: nil,
			deleteErr:   nil,
			wantErr:     true,
		},
		{
			name:            "returns error when image not found",
			medicalRecordID: 1,
			imageID:         999,
			repoImage:       nil,
			findByIDErr:     errors.New("not found"),
			deleteErr:       nil,
			wantErr:         true,
		},
		{
			name:            "returns error when delete fails",
			medicalRecordID: 1,
			imageID:         1,
			repoImage: &model.MedicalRecordImage{
				ID:              1,
				MedicalRecordID: 1,
			},
			findByIDErr: nil,
			deleteErr:   errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordImageRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecordImage, error) {
					return tt.repoImage, tt.findByIDErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			medRecRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					// stub: always return a valid draft parent record (LockByIDForUpdate delegates to FindByID)
					return &model.MedicalRecord{ID: tt.medicalRecordID, Status: model.MedicalRecordStatusDraft}, nil
				},
			}
			svc := NewMedicalRecordImageService(repo, medRecRepo, &mockCheckupTransactor{})

			err := svc.Delete(context.Background(), 1, tt.medicalRecordID, tt.imageID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMedicalRecordImageService_FinalizedGuard は SD-2 回帰テスト:
// 確定済み(finalized)カルテへの画像追加・削除が拒否されることを検証する
// （examination_service.go の同型テストが先例、backend/internal/service/CLAUDE.md
// 「カルテ子エンティティ書込」不変条件）。
func TestMedicalRecordImageService_FinalizedGuard(t *testing.T) {
	t.Run("Create: returns conflict when parent medical record is finalized", func(t *testing.T) {
		repo := &mockMedicalRecordImageRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecordImage) error {
				t.Fatal("repo.Create must not be called when parent medical record is finalized")
				return nil
			},
		}
		medRecRepo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: 1, Status: model.MedicalRecordStatusFinalized}, nil
			},
		}
		svc := NewMedicalRecordImageService(repo, medRecRepo, &mockCheckupTransactor{})

		image, err := svc.Create(context.Background(), 1, 1, &CreateMedicalRecordImageInput{
			ImageURL:  "https://example.com/image.jpg",
			FileName:  "image.jpg",
			ImageType: model.MedicalImageTypeXray,
		})

		assert.Error(t, err)
		assert.Nil(t, image)
	})

	t.Run("Delete: returns conflict when parent medical record is finalized", func(t *testing.T) {
		repo := &mockMedicalRecordImageRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecordImage, error) {
				return &model.MedicalRecordImage{ID: 1, MedicalRecordID: 1, FileName: "image.jpg"}, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				t.Fatal("repo.Delete must not be called when parent medical record is finalized")
				return nil
			},
		}
		medRecRepo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: 1, Status: model.MedicalRecordStatusFinalized}, nil
			},
		}
		svc := NewMedicalRecordImageService(repo, medRecRepo, &mockCheckupTransactor{})

		err := svc.Delete(context.Background(), 1, 1, 1)

		assert.Error(t, err)
	})
}

func TestMedicalRecordImageService_CreateRejectsInvalidRequestRelations(t *testing.T) {
	const (
		clinicID      = uint64(1)
		medicalRecord = uint64(10)
		otherRecord   = uint64(11)
		examinationID = uint64(20)
		staffID       = uint64(30)
	)

	tests := []struct {
		name       string
		exam       *model.Examination
		staff      *model.Staff
		assignment *model.StaffClinicAssignment
		wantErr    bool
	}{
		{
			name: "accepts active same-clinic relations",
			exam: &model.Examination{
				ID: examinationID, ClinicID: clinicID, MedicalRecordID: ptrUint64(medicalRecord),
			},
			staff:      &model.Staff{ID: staffID, IsActive: true},
			assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID},
		},
		{
			name: "rejects examination belonging to another medical record",
			exam: &model.Examination{
				ID: examinationID, ClinicID: clinicID, MedicalRecordID: ptrUint64(otherRecord),
			},
			staff:      &model.Staff{ID: staffID, IsActive: true},
			assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID},
			wantErr:    true,
		},
		{
			name: "rejects examination belonging to another clinic",
			exam: &model.Examination{
				ID: examinationID, ClinicID: clinicID + 1, MedicalRecordID: ptrUint64(medicalRecord),
			},
			staff:      &model.Staff{ID: staffID, IsActive: true},
			assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID},
			wantErr:    true,
		},
		{
			name: "rejects inactive staff",
			exam: &model.Examination{
				ID: examinationID, ClinicID: clinicID, MedicalRecordID: ptrUint64(medicalRecord),
			},
			staff:      &model.Staff{ID: staffID, IsActive: false},
			assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID},
			wantErr:    true,
		},
		{
			name: "rejects staff without requested-clinic assignment",
			exam: &model.Examination{
				ID: examinationID, ClinicID: clinicID, MedicalRecordID: ptrUint64(medicalRecord),
			},
			staff:      &model.Staff{ID: staffID, IsActive: true},
			assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID + 1},
			wantErr:    true,
		},
	}

	const petID = uint64(40)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createCalls := 0
			repo := &mockMedicalRecordImageRepository{
				createFn: func(_ context.Context, image *model.MedicalRecordImage) error {
					createCalls++
					image.ID = 1
					return nil
				},
			}
			medRecRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, gotClinicID, gotRecordID uint64) (*model.MedicalRecord, error) {
					return draftMedicalRecordWithPet(gotClinicID, gotRecordID, petID), nil
				},
			}
			svc := NewMedicalRecordImageServiceWithRelationValidation(
				repo,
				medRecRepo,
				livingPetFinder(),
				&medicalRecordImageExaminationFinderStub{exam: tt.exam},
				&clinicalStaffLockerStub{staff: tt.staff},
				&clinicalStaffAssignmentLockerStub{assignment: tt.assignment},
				&mockCheckupTransactor{},
			)

			got, err := svc.Create(context.Background(), clinicID, medicalRecord, &CreateMedicalRecordImageInput{
				ImageURL: "https://example.com/image.jpg",
				ExamID:   ptrUint64(examinationID),
				StaffID:  ptrUint64(staffID),
			})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Zero(t, createCalls)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, 1, createCalls)
		})
	}
}

func TestMedicalRecordImageService_CreateFailsClosedWithoutRelationDependencies(t *testing.T) {
	createCalls := 0
	const petID = uint64(40)
	repo := &mockMedicalRecordImageRepository{
		createFn: func(_ context.Context, _ *model.MedicalRecordImage) error {
			createCalls++
			return nil
		},
	}
	medRecRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			return draftMedicalRecordWithPet(clinicID, id, petID), nil
		},
	}
	// pets 依存なしの簡易 constructor — 死亡検証で fail-closed になる。
	svc := NewMedicalRecordImageService(repo, medRecRepo, &mockCheckupTransactor{})

	got, err := svc.Create(context.Background(), 1, 10, &CreateMedicalRecordImageInput{
		ImageURL: "https://example.com/image.jpg",
		ExamID:   ptrUint64(20),
	})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Zero(t, createCalls)
}

// TestMedicalRecordImageService_CreateRejectsDeceasedPet は SEC-CS-F14:
// 親カルテのペットが死亡している場合、画像 Create を拒否し repo.Create を呼ばない。
func TestMedicalRecordImageService_CreateRejectsDeceasedPet(t *testing.T) {
	const (
		clinicID      = uint64(1)
		medicalRecord = uint64(10)
		petID         = uint64(40)
	)
	deceasedAt := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	createCalls := 0
	repo := &mockMedicalRecordImageRepository{
		createFn: func(_ context.Context, _ *model.MedicalRecordImage) error {
			createCalls++
			t.Fatal("repo.Create must not be called for a deceased pet")
			return nil
		},
	}
	medRecRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, gotClinicID, gotRecordID uint64) (*model.MedicalRecord, error) {
			return draftMedicalRecordWithPet(gotClinicID, gotRecordID, petID), nil
		},
	}
	pets := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, DeceasedAt: &deceasedAt}, nil
		},
	}
	svc := NewMedicalRecordImageServiceWithRelationValidation(
		repo,
		medRecRepo,
		pets,
		&medicalRecordImageExaminationFinderStub{},
		&clinicalStaffLockerStub{},
		&clinicalStaffAssignmentLockerStub{},
		&mockCheckupTransactor{},
	)

	got, err := svc.Create(context.Background(), clinicID, medicalRecord, &CreateMedicalRecordImageInput{
		ImageURL:  "https://example.com/image.jpg",
		FileName:  "image.jpg",
		ImageType: model.MedicalImageTypeXray,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, got)
	assert.Zero(t, createCalls)
}
