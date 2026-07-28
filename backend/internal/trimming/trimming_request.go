package trimming

import (
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type listTrimmingQuery struct {
	PetID     string
	OwnerID   string
	StartDate string
	EndDate   string
}

func newListTrimmingQuery(values url.Values) listTrimmingQuery {
	return listTrimmingQuery{
		PetID:     values.Get("pet_id"),
		OwnerID:   values.Get("owner_id"),
		StartDate: values.Get("start_date"),
		EndDate:   values.Get("end_date"),
	}
}

type listTrimmingFilters struct {
	PetID     *uint64
	OwnerID   *uint64
	StartDate *string
	EndDate   *string
}

func (q listTrimmingQuery) toServiceFilters() (listTrimmingFilters, error) {
	petID, err := parseOptionalUintQueryFilter(q.PetID, "pet_id")
	if err != nil {
		return listTrimmingFilters{}, err
	}
	ownerID, err := parseOptionalUintQueryFilter(q.OwnerID, "owner_id")
	if err != nil {
		return listTrimmingFilters{}, err
	}
	startDate, err := parseOptionalDateQueryFilter(q.StartDate, "start_date")
	if err != nil {
		return listTrimmingFilters{}, err
	}
	endDate, err := parseOptionalDateQueryFilter(q.EndDate, "end_date")
	if err != nil {
		return listTrimmingFilters{}, err
	}
	return listTrimmingFilters{
		PetID:     petID,
		OwnerID:   ownerID,
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

// createTrimmingRequest はトリミング予約作成のバインド struct（BE-119: appointments ベース）
type createTrimmingRequest struct {
	AppointmentID     *uint64    `json:"appointment_id"`
	ReservationTypeID uint64     `json:"reservation_type_id" binding:"required"`
	StartTime         *time.Time `json:"start_time"`
	EndTime           *time.Time `json:"end_time"`
	PetID             *uint64    `json:"pet_id"               binding:"required"`
	StaffID           *uint64    `json:"staff_id"`
	Status            string     `json:"status"              binding:"omitempty,oneof=confirmed pending cancelled checked_in in_consultation accounting"`
	ReservationRoute  string     `json:"reservation_route"   binding:"omitempty,oneof=line phone reception exam_room record_shortcut"`
	// トリミング詳細
	CourseID       *uint64  `json:"course_id"`
	StyleRequest   string   `json:"style_request" binding:"omitempty,max=2000"`
	BW             *float64 `json:"bw"`
	BWUnit         string   `json:"bw_unit" binding:"omitempty,oneof=Kg g"`
	BT             *float64 `json:"bt"`
	UsedShampoo    string   `json:"used_shampoo" binding:"omitempty,max=2000"`
	UsedRibbon     string   `json:"used_ribbon" binding:"omitempty,max=2000"`
	Remarks        string   `json:"remarks" binding:"omitempty,max=2000"`
	StyleImage     string   `json:"style_image" binding:"omitempty,max=2048"`
	CompletedImage string   `json:"completed_image" binding:"omitempty,max=2048"`
	OptionIDs      []uint64 `json:"option_ids" binding:"max=50,dive"`
}

func (r *createTrimmingRequest) toServiceInput() *CreateTrimmingInput {
	input := &CreateTrimmingInput{
		AppointmentID:     r.AppointmentID,
		ReservationTypeID: r.ReservationTypeID,
		PetID:             r.PetID,
		StaffID:           r.StaffID,
		CourseID:          r.CourseID,
		StyleRequest:      r.StyleRequest,
		BodyWeight:        r.BW,
		BodyTemperature:   r.BT,
		UsedShampoo:       r.UsedShampoo,
		UsedRibbon:        r.UsedRibbon,
		Remarks:           r.Remarks,
		StyleImage:        r.StyleImage,
		CompletedImage:    r.CompletedImage,
		OptionIDs:         r.OptionIDs,
	}
	if r.StartTime != nil {
		input.StartTime = *r.StartTime
	}
	if r.EndTime != nil {
		input.EndTime = *r.EndTime
	}
	if r.Status != "" {
		input.Status = model.ReservationStatus(r.Status)
	}
	if r.ReservationRoute != "" {
		input.ReservationRoute = &r.ReservationRoute
	}
	if r.BWUnit != "" {
		input.BWUnit = model.BodyWeightUnit(r.BWUnit)
	}
	return input
}

// updateTrimmingRequest はトリミング予約更新のバインド struct（BE-119）
type updateTrimmingRequest struct {
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	PetID     *uint64    `json:"pet_id"`
	StaffID   *uint64    `json:"staff_id"`
	Status    *string    `json:"status"      binding:"omitempty,oneof=confirmed pending cancelled checked_in in_consultation accounting"`
	// トリミング詳細
	CourseID       *uint64  `json:"course_id"`
	StyleRequest   *string  `json:"style_request" binding:"omitempty,max=2000"`
	BW             *float64 `json:"bw"`
	BWUnit         *string  `json:"bw_unit" binding:"omitempty,oneof=Kg g"`
	BT             *float64 `json:"bt"`
	UsedShampoo    *string  `json:"used_shampoo" binding:"omitempty,max=2000"`
	UsedRibbon     *string  `json:"used_ribbon" binding:"omitempty,max=2000"`
	Remarks        *string  `json:"remarks" binding:"omitempty,max=2000"`
	StyleImage     *string  `json:"style_image" binding:"omitempty,max=2048"`
	CompletedImage *string  `json:"completed_image" binding:"omitempty,max=2048"`
	// nil = 変更なし、non-nil（空スライス含む）= 全置換
	OptionIDs *[]uint64 `json:"option_ids" binding:"omitempty,max=50,dive"`
}

func (r *updateTrimmingRequest) toServiceInput() *UpdateTrimmingInput {
	input := &UpdateTrimmingInput{
		StartTime:       r.StartTime,
		EndTime:         r.EndTime,
		PetID:           r.PetID,
		StaffID:         r.StaffID,
		CourseID:        r.CourseID,
		StyleRequest:    r.StyleRequest,
		BodyWeight:      r.BW,
		BodyTemperature: r.BT,
		UsedShampoo:     r.UsedShampoo,
		UsedRibbon:      r.UsedRibbon,
		Remarks:         r.Remarks,
		StyleImage:      r.StyleImage,
		CompletedImage:  r.CompletedImage,
		OptionIDs:       r.OptionIDs,
	}
	if r.Status != nil {
		status := model.ReservationStatus(*r.Status)
		input.Status = &status
	}
	if r.BWUnit != nil {
		unit := model.BodyWeightUnit(*r.BWUnit)
		input.BWUnit = &unit
	}
	return input
}
