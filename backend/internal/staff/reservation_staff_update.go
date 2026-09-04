package staff

// ReservationStaffUpdate is the staff-owned typed write command for
// reservation-originated staffs column updates (BE-RC-001).
//
// Field-update keys match reservation.buildReservationStaffUpdate
// (name, staff_type, reservation_visible, reservation_comment, sort_order).
// IsActive is the PatchStatus write (staffs.is_active) that already used the
// same clinic-scoped reservation_staff UPDATE; it is not emitted by
// buildReservationStaffUpdate. Arbitrary map[string]any is not part of this
// contract. Column map conversion stays unexported in this package.
type ReservationStaffUpdate struct {
	Name               *string
	StaffType          *string
	ReservationVisible *bool
	ReservationComment *string
	SortOrder          *int
	IsActive           *bool
}

const (
	reservationStaffColName               = "name"
	reservationStaffColStaffType          = "staff_type"
	reservationStaffColReservationVisible = "reservation_visible"
	reservationStaffColReservationComment = "reservation_comment"
	reservationStaffColSortOrder          = "sort_order"
	reservationStaffColIsActive           = "is_active"
)

func reservationStaffUpdateFields(cmd ReservationStaffUpdate) map[string]any {
	fields := make(map[string]any)
	if cmd.Name != nil {
		fields[reservationStaffColName] = *cmd.Name
	}
	if cmd.StaffType != nil {
		fields[reservationStaffColStaffType] = *cmd.StaffType
	}
	if cmd.ReservationVisible != nil {
		fields[reservationStaffColReservationVisible] = *cmd.ReservationVisible
	}
	if cmd.ReservationComment != nil {
		fields[reservationStaffColReservationComment] = *cmd.ReservationComment
	}
	if cmd.SortOrder != nil {
		fields[reservationStaffColSortOrder] = *cmd.SortOrder
	}
	if cmd.IsActive != nil {
		fields[reservationStaffColIsActive] = *cmd.IsActive
	}
	return fields
}
