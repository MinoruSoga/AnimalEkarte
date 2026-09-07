package staff

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestValidateStaffType(t *testing.T) {
	t.Run("empty is allowed for create defaults", func(t *testing.T) {
		require.NoError(t, validateStaffType(""))
	})
	t.Run("accepted enums", func(t *testing.T) {
		for _, st := range []string{
			string(model.StaffTypeDoctor),
			string(model.StaffTypeNurse),
			string(model.StaffTypeTrimmer),
			string(model.StaffTypeResource),
		} {
			require.NoError(t, validateStaffType(st), st)
		}
	})
	t.Run("rejects unknown", func(t *testing.T) {
		err := validateStaffType("receptionist")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid staff_type")
	})
}

func TestBuildStaffUpdate(t *testing.T) {
	t.Run("empty input produces empty field map", func(t *testing.T) {
		fields := buildStaffUpdate(&UpdateStaffInput{})
		assert.Empty(t, fields)
	})

	t.Run("all fields set are mapped to their column names", func(t *testing.T) {
		name := "山田獣医師"
		licenseNumber := "12345"
		occupationID := uint64(2)
		sortOrder := 1
		isActive := true
		staffType := "doctor"
		dispName := "山田先生"
		visible := true
		comment := "コメント"
		imageURL := "https://example.com/staff.png"

		input := &UpdateStaffInput{
			Name:                   &name,
			LicenseNumber:          &licenseNumber,
			OccupationID:           &occupationID,
			SortOrder:              &sortOrder,
			IsActive:               &isActive,
			StaffType:              &staffType,
			ReservationDisplayName: &dispName,
			ReservationVisible:     &visible,
			ReservationComment:     &comment,
			ReservationImageURL:    &imageURL,
		}
		fields := buildStaffUpdate(input)

		assert.Equal(t, name, fields[colStaffName])
		assert.Equal(t, licenseNumber, fields[colStaffLicenseNumber])
		assert.Equal(t, occupationID, fields[colStaffOccupationID])
		assert.Equal(t, sortOrder, fields[colStaffSortOrder])
		assert.Equal(t, isActive, fields[colStaffIsActive])
		assert.Equal(t, staffType, fields[colStaffStaffType])
		assert.Equal(t, dispName, fields[colStaffReservationDisplayName])
		assert.Equal(t, visible, fields[colStaffReservationVisible])
		assert.Equal(t, comment, fields[colStaffReservationComment])
		assert.Equal(t, imageURL, fields[colStaffReservationImageURL])
	})

	t.Run("only Name set: only name column present", func(t *testing.T) {
		name := "佐藤"
		fields := buildStaffUpdate(&UpdateStaffInput{Name: &name})
		assert.Len(t, fields, 1)
		assert.Equal(t, name, fields[colStaffName])
	})
}

// AUS-03: application-layer staff_type rejection on Create/CreateWithAccount/Update paths.
func TestService_CreateUpdate_RejectsInvalidStaffType(t *testing.T) {
	svc := NewService(
		&mockStaffRepository{
			createFn: func(context.Context, *model.Staff) error {
				t.Fatal("repository Create must not run for invalid staff_type")
				return nil
			},
			updateFn: func(context.Context, uint64, uint64, UpdateStaffInput) error {
				t.Fatal("repository Update must not run for invalid staff_type")
				return nil
			},
		},
		&mockAccountForStaff{},
		&mockAssignmentForStaff{},
		&mockReservationForStaff{},
		&mockShiftEntryForStaff{},
		&mockPermissionGroupRepository{},
		&mockResStaffForStaff{},
		nil,
		nil,
		noopTransactor{},
	)

	t.Run("Create", func(t *testing.T) {
		_, err := svc.Create(context.Background(), &CreateStaffInput{
			ClinicID:  1,
			Name:      "staff",
			StaffType: "receptionist",
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("CreateWithAccount", func(t *testing.T) {
		_, err := svc.CreateWithAccount(context.Background(), &CreateStaffWithAccountInput{
			ClinicID:  1,
			Name:      "staff",
			Email:     "staff-type@example.test",
			Password:  "Passw0rd1",
			StaffType: "receptionist",
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("Update", func(t *testing.T) {
		bad := "receptionist"
		_, err := svc.Update(context.Background(), 1, 10, &UpdateStaffInput{StaffType: &bad})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}
