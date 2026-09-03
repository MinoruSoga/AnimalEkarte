package staff

// ReservationStaffUpdate is the staff-owned typed write command for
// reservation-originated staffs column updates (BE-RC-001).
//
// Fields match the keys already mapped by reservation.buildReservationStaffUpdate
// (name, staff_type, reservation_visible, reservation_comment, sort_order).
// Arbitrary map[string]any is not part of this contract. Column map conversion
// stays unexported in this package.
//
// W1 switches exported UpdateForReservation and reservation staffsWriter onto
// this type. Phase 0 freezes the command so concurrent lanes share one shape.
type ReservationStaffUpdate struct {
	Name               *string
	StaffType          *string
	ReservationVisible *bool
	ReservationComment *string
	SortOrder          *int
}

const (
	reservationStaffColName               = "name"
	reservationStaffColStaffType          = "staff_type"
	reservationStaffColReservationVisible = "reservation_visible"
	reservationStaffColReservationComment = "reservation_comment"
	reservationStaffColSortOrder          = "sort_order"
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
	return fields
}
