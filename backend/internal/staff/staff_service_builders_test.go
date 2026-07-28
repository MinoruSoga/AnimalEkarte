package staff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
