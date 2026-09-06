package staff

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestService_UpdateRequiresAuthorityOverEveryActiveAssignment(t *testing.T) {
	const (
		targetStaffID = uint64(7)
		clinicA       = uint64(10)
		clinicB       = uint64(20)
	)
	name := "global profile update"

	tests := []struct {
		name                string
		assignmentClinicIDs []uint64
		authorizedClinicIDs []uint64
		isSystemAdmin       bool
		wantForbidden       bool
	}{
		{
			name:                "rejects clinic A editor when target is also assigned to clinic B",
			assignmentClinicIDs: []uint64{clinicA, clinicB},
			authorizedClinicIDs: []uint64{clinicA},
			wantForbidden:       true,
		},
		{
			name:                "rejects non-system editor even when assigned to every target clinic",
			assignmentClinicIDs: []uint64{clinicA, clinicB},
			authorizedClinicIDs: []uint64{clinicA, clinicB},
			wantForbidden:       true,
		},
		{
			name:                "allows editor for a single-clinic target",
			assignmentClinicIDs: []uint64{clinicA},
			authorizedClinicIDs: []uint64{clinicA},
		},
		{
			name:                "allows system administrator without enumerating clinic ids",
			assignmentClinicIDs: []uint64{clinicA, clinicB},
			isSystemAdmin:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updated := false
			repo := &coreMockStaffRepository{
				lockInClinicFn: func(
					_ context.Context,
					clinicID, staffID uint64,
				) (*model.Staff, error) {
					assert.Equal(t, clinicA, clinicID)
					assert.Equal(t, targetStaffID, staffID)
					return &model.Staff{ID: staffID, ClinicID: clinicA}, nil
				},
				updateFn: func(
					_ context.Context,
					clinicID, staffID uint64,
					_ UpdateStaffInput,
				) error {
					assert.Equal(t, clinicA, clinicID)
					assert.Equal(t, targetStaffID, staffID)
					updated = true
					return nil
				},
			}
			assignments := &coreMockStaffClinicAssignmentRepository{
				lockActiveFn: func(
					_ context.Context,
					staffID uint64,
				) ([]model.StaffClinicAssignment, error) {
					assert.Equal(t, targetStaffID, staffID)
					assignments := make(
						[]model.StaffClinicAssignment,
						0,
						len(test.assignmentClinicIDs),
					)
					for _, clinicID := range test.assignmentClinicIDs {
						assignments = append(
							assignments,
							model.StaffClinicAssignment{
								StaffID:  staffID,
								ClinicID: clinicID,
							},
						)
					}
					return assignments, nil
				},
				lockStaffAndClinicFn: func(
					context.Context,
					uint64,
					uint64,
				) (*model.StaffClinicAssignment, error) {
					t.Fatal("global-impact update must lock every assignment in one operation")
					return nil, nil
				},
			}
			service := newCoreService(
				repo,
				&coreMockAccountRepository{},
				assignments,
				&coreMockReservationQueryRepository{},
				&coreMockShiftEntryRepository{},
				&coreFakeTransactor{},
			)

			result, err := service.Update(
				context.Background(),
				clinicA,
				targetStaffID,
				&UpdateStaffInput{
					Name:                &name,
					AuthorizedClinicIDs: test.authorizedClinicIDs,
					IsSystemAdmin:       test.isSystemAdmin,
				},
			)

			if test.wantForbidden {
				require.Error(t, err)
				assert.True(t, errors.Is(err, apperrors.ErrForbidden))
				assert.Nil(t, result)
				assert.False(t, updated)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, updated)
		})
	}
}

func TestService_UpdateLocksAssignmentsInsideMutationTransaction(t *testing.T) {
	type transactionMarker struct{}

	const (
		targetStaffID = uint64(7)
		clinicID      = uint64(10)
	)
	name := "authorized update"
	transaction := passwordUpdateTransactorFunc(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(context.WithValue(ctx, transactionMarker{}, true))
		},
	)
	assignments := &coreMockStaffClinicAssignmentRepository{
		lockActiveFn: func(
			ctx context.Context,
			staffID uint64,
		) ([]model.StaffClinicAssignment, error) {
			assert.Equal(t, true, ctx.Value(transactionMarker{}))
			return []model.StaffClinicAssignment{
				{StaffID: staffID, ClinicID: clinicID},
			}, nil
		},
	}
	service := newCoreService(
		&coreMockStaffRepository{},
		&coreMockAccountRepository{},
		assignments,
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		transaction,
	)

	_, err := service.Update(
		context.Background(),
		clinicID,
		targetStaffID,
		&UpdateStaffInput{
			Name:                &name,
			AuthorizedClinicIDs: []uint64{clinicID},
		},
	)

	require.NoError(t, err)
}

func TestService_UpdateAssignmentLockFailurePreventsMutation(t *testing.T) {
	lockError := errors.New("assignment lock failed")
	updated := false
	service := newCoreService(
		&coreMockStaffRepository{
			updateFn: func(context.Context, uint64, uint64, UpdateStaffInput) error {
				updated = true
				return nil
			},
		},
		&coreMockAccountRepository{},
		&coreMockStaffClinicAssignmentRepository{
			lockActiveFn: func(
				context.Context,
				uint64,
			) ([]model.StaffClinicAssignment, error) {
				return nil, lockError
			},
		},
		&coreMockReservationQueryRepository{},
		&coreMockShiftEntryRepository{},
		&coreFakeTransactor{},
	)
	name := "blocked update"

	result, err := service.Update(
		context.Background(),
		10,
		7,
		&UpdateStaffInput{
			Name:                &name,
			AuthorizedClinicIDs: []uint64{10},
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, lockError)
	assert.Nil(t, result)
	assert.False(t, updated)
}

func TestAuthorizeGlobalStaffUpdateRejectsInvalidAssignmentState(t *testing.T) {
	const (
		staffID  = uint64(7)
		clinicID = uint64(10)
	)
	tests := []struct {
		name          string
		assignments   []model.StaffClinicAssignment
		isSystemAdmin bool
		wantCode      string
	}{
		{
			name:     "no active assignments",
			wantCode: "NOT_FOUND",
		},
		{
			name: "assignment belongs to another staff",
			assignments: []model.StaffClinicAssignment{
				{StaffID: staffID + 1, ClinicID: clinicID},
			},
			wantCode: "INTERNAL",
		},
		{
			name: "assignment has zero clinic id",
			assignments: []model.StaffClinicAssignment{
				{StaffID: staffID},
			},
			wantCode: "INTERNAL",
		},
		{
			name: "assignment is soft deleted",
			assignments: []model.StaffClinicAssignment{
				{
					StaffID:  staffID,
					ClinicID: clinicID,
					DeletedAt: gorm.DeletedAt{
						Valid: true,
					},
				},
			},
			wantCode: "INTERNAL",
		},
		{
			name: "current clinic assignment is absent for non-admin",
			assignments: []model.StaffClinicAssignment{
				{StaffID: staffID, ClinicID: clinicID + 1},
			},
			wantCode: "NOT_FOUND",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := authorizeGlobalStaffUpdate(
				staffID,
				clinicID,
				test.assignments,
				nil,
				test.isSystemAdmin,
			)

			require.Error(t, err)
			var appError *apperrors.AppError
			require.ErrorAs(t, err, &appError)
			assert.Equal(t, test.wantCode, appError.Code)
		})
	}
}

type passwordUpdateTransactorFunc func(
	context.Context,
	func(context.Context) error,
) error

func (f passwordUpdateTransactorFunc) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	return f(ctx, fn)
}

func authorizedStaffUpdate(input *UpdateStaffInput, clinicIDs ...uint64) *UpdateStaffInput {
	if input == nil {
		return nil
	}
	authorizedInput := *input
	authorizedInput.AuthorizedClinicIDs = append([]uint64(nil), clinicIDs...)
	return &authorizedInput
}

func testStaffCredentialAudit(
	clinicID, targetStaffID uint64,
) *CredentialMutationAudit {
	return &CredentialMutationAudit{
		ClinicID:      clinicID,
		ActorStaffID:  999,
		TargetStaffID: targetStaffID,
	}
}

func TestValidatePasswordUsesRuneMinimumAndBcryptByteMaximum(t *testing.T) {
	t.Run("rejects seven unicode runes even when byte length exceeds eight", func(t *testing.T) {
		err := validatePassword("あいうえおか1")

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("accepts eight unicode runes within bcrypt byte maximum", func(t *testing.T) {
		require.NoError(t, validatePassword("あいうえおかき1"))
	})

	t.Run("rejects more than seventy two bytes", func(t *testing.T) {
		err := validatePassword("Abcdefgh1" + strings.Repeat("x", 64))

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}
