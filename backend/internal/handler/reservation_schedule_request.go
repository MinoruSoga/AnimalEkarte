package handler

type breakInputRequest struct {
	Start string `json:"start" binding:"required"`
	End   string `json:"end"   binding:"required"`
}

type upsertReservationScheduleRequest struct {
	ShiftType string              `json:"shift_type" binding:"required,oneof=full morning afternoon off paid_leave"`
	WorkStart *string             `json:"work_start"`
	WorkEnd   *string             `json:"work_end"`
	Breaks    []breakInputRequest `json:"breaks"`
}
