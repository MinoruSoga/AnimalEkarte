package staff

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupStaffOccupationWriteRaceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Account{},
		&model.Occupation{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE staff_clinic_assignments, staffs, accounts, occupations CASCADE",
	).Error)
	return db
}

func newStaffOccupationRaceAccountStore(db *gorm.DB) StaffAccountStore {
	return staffResetInvalidationAccountStore{
		AccountRepository: auth.NewAccountRepository(db),
		resetTokens:       auth.NewPasswordResetTokenRepository(db),
	}
}

func makeStaffOccupationRaceClinic(
	t *testing.T,
	db *gorm.DB,
	name string,
) (*model.Clinic, *model.Occupation) {
	t.Helper()
	company := &model.Company{Name: name + "法人"}
	require.NoError(t, db.Create(company).Error)
	clinic := &model.Clinic{CompanyID: company.ID, Name: name + "医院", IsActive: true}
	require.NoError(t, db.Create(clinic).Error)
	occupation := &model.Occupation{
		ClinicID: clinic.ID, Name: name + "職種", IsActive: true,
	}
	require.NoError(t, db.Create(occupation).Error)
	return clinic, occupation
}

type coordinatedOccupationRepository struct {
	OccupationRepository
	shareLocked     chan struct{}
	releaseShare    chan struct{}
	updateAttempted chan struct{}
	shareOnce       sync.Once
	updateOnce      sync.Once
}

func (r *coordinatedOccupationRepository) LockActiveByIDForShare(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Occupation, error) {
	occupation, err := r.OccupationRepository.LockActiveByIDForShare(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	r.shareOnce.Do(func() {
		close(r.shareLocked)
		<-r.releaseShare
	})
	return occupation, nil
}

func (r *coordinatedOccupationRepository) LockActiveByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.Occupation, error) {
	r.updateOnce.Do(func() { close(r.updateAttempted) })
	return r.OccupationRepository.LockActiveByIDForUpdate(ctx, clinicID, id)
}

type staffOccupationMutationResult struct {
	staff *model.Staff
	err   error
}

func awaitStaffOccupationMutation(
	t *testing.T,
	result <-chan staffOccupationMutationResult,
) staffOccupationMutationResult {
	t.Helper()
	select {
	case got := <-result:
		return got
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for staff occupation mutation")
		return staffOccupationMutationResult{}
	}
}

func makeUnassignedOccupationRaceStaff(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
) *model.Staff {
	t.Helper()
	staff := &model.Staff{
		ClinicID: clinicID, Name: "更新対象スタッフ",
		IsActive: true, StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(staff).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: staff.ID, ClinicID: clinicID, IsMain: true,
	}).Error)
	return staff
}

func TestStaffOccupationWrites_SerializeAgainstOccupationDeleteDatabase(t *testing.T) {
	tests := []struct {
		name      string
		operation string
	}{
		{name: "Create", operation: "create"},
		{name: "CreateWithAccount", operation: "create_with_account"},
		{name: "Update", operation: "update"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupStaffOccupationWriteRaceDB(t)
			clinic, occupation := makeStaffOccupationRaceClinic(t, db, tt.name)
			repo := &coordinatedOccupationRepository{
				OccupationRepository: NewOccupationRepository(db),
				shareLocked:          make(chan struct{}),
				releaseShare:         make(chan struct{}),
				updateAttempted:      make(chan struct{}),
			}
			staffService := NewService(
				NewRepository(db),
				newStaffOccupationRaceAccountStore(db),
				NewStaffClinicAssignmentRepository(db),
				nil, nil, nil, nil,
				repo,
				nil,
				persistence.NewTransactor(db),
			)
			occupationService := NewOccupationService(repo)
			var existing *model.Staff
			if tt.operation == "update" {
				existing = makeUnassignedOccupationRaceStaff(t, db, clinic.ID)
			}

			mutationResult := make(chan staffOccupationMutationResult, 1)
			go func() {
				var result *model.Staff
				var err error
				switch tt.operation {
				case "create":
					result, err = staffService.Create(context.Background(), &CreateStaffInput{
						ClinicID: clinic.ID, Name: "新規スタッフ", OccupationID: &occupation.ID,
					})
				case "create_with_account":
					result, err = staffService.CreateWithAccount(
						context.Background(),
						&CreateStaffWithAccountInput{
							ClinicID:     clinic.ID,
							Name:         "アカウント付きスタッフ",
							Email:        fmt.Sprintf("occupation-race-%d@example.com", occupation.ID),
							Password:     "password123",
							OccupationID: &occupation.ID,
						},
					)
				case "update":
					result, err = staffService.Update(
						context.Background(),
						clinic.ID,
						existing.ID,
						&UpdateStaffInput{
							OccupationID:        &occupation.ID,
							AuthorizedClinicIDs: []uint64{clinic.ID},
						},
					)
				}
				mutationResult <- staffOccupationMutationResult{staff: result, err: err}
			}()
			awaitStaffTestSignal(t, repo.shareLocked, "occupation SHARE lock")

			deleteResult := make(chan error, 1)
			go func() {
				deleteResult <- occupationService.Delete(
					context.Background(),
					clinic.ID,
					occupation.ID,
				)
			}()
			awaitStaffTestSignal(t, repo.updateAttempted, "occupation UPDATE lock attempt")
			close(repo.releaseShare)

			mutation := awaitStaffOccupationMutation(t, mutationResult)
			require.NoError(t, mutation.err)
			require.NotNil(t, mutation.staff)
			deleteErr := awaitStaffTestError(t, deleteResult, "occupation delete")
			require.Error(t, deleteErr)
			assert.True(t, apperrors.IsConflict(deleteErr), "unexpected error: %v", deleteErr)

			var reloaded model.Staff
			require.NoError(t, db.First(&reloaded, mutation.staff.ID).Error)
			require.NotNil(t, reloaded.OccupationID)
			assert.Equal(t, occupation.ID, *reloaded.OccupationID)
			var activeOccupationCount int64
			require.NoError(t, db.Model(&model.Occupation{}).
				Where("id = ?", occupation.ID).Count(&activeOccupationCount).Error)
			assert.Equal(t, int64(1), activeOccupationCount)
		})
	}
}
