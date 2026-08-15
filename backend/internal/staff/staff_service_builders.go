package staff

import (
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colStaffName                   = "name"
	colStaffLicenseNumber          = "license_number"
	colStaffOccupationID           = "occupation_id"
	colStaffSortOrder              = "sort_order"
	colStaffIsActive               = "is_active"
	colStaffStaffType              = "staff_type"
	colStaffReservationDisplayName = "reservation_display_name"
	colStaffReservationVisible     = "reservation_visible"
	colStaffReservationComment     = "reservation_comment"
	colStaffReservationImageURL    = "reservation_image_url"
)

// validateStaffType enforces doctor/nurse/trimmer/resource at application boundary (AUS-03).
// Empty is allowed (defaults applied by Create paths).
func validateStaffType(staffType string) error {
	if staffType == "" {
		return nil
	}
	switch model.StaffType(staffType) {
	case model.StaffTypeDoctor, model.StaffTypeNurse, model.StaffTypeTrimmer, model.StaffTypeResource:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid staff_type: %s", staffType))
	}
}

func buildStaffUpdate(input *UpdateStaffInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields[colStaffName] = *input.Name
	}
	if input.LicenseNumber != nil {
		fields[colStaffLicenseNumber] = *input.LicenseNumber
	}
	if input.OccupationID != nil {
		fields[colStaffOccupationID] = *input.OccupationID
	}
	if input.SortOrder != nil {
		fields[colStaffSortOrder] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields[colStaffIsActive] = *input.IsActive
	}
	if input.StaffType != nil {
		fields[colStaffStaffType] = *input.StaffType
	}
	if input.ReservationDisplayName != nil {
		fields[colStaffReservationDisplayName] = *input.ReservationDisplayName
	}
	if input.ReservationVisible != nil {
		fields[colStaffReservationVisible] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields[colStaffReservationComment] = *input.ReservationComment
	}
	if input.ReservationImageURL != nil {
		fields[colStaffReservationImageURL] = *input.ReservationImageURL
	}
	return fields
}
