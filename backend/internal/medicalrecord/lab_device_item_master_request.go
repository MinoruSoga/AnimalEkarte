package medicalrecord

import (
	"encoding/base64"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type updateLabDeviceItemMasterRequest struct {
	Unit            string  `json:"unit"`
	ExamTypeFieldID *uint64 `json:"exam_type_field_id"`
	IsActive        *bool   `json:"is_active" binding:"required"`
}

func (r updateLabDeviceItemMasterRequest) toServiceInput() UpdateLabDeviceItemMasterInput {
	return UpdateLabDeviceItemMasterInput{
		Unit:            r.Unit,
		ExamTypeFieldID: r.ExamTypeFieldID,
		IsActive:        *r.IsActive,
	}
}

type labDeviceItemMasterResponse struct {
	ID              uint64  `json:"id"`
	SourceType      string  `json:"source_type"`
	DeviceItemCode  string  `json:"device_item_code"`
	Unit            string  `json:"unit"`
	ValueShape      string  `json:"value_shape"`
	ExamTypeFieldID *uint64 `json:"exam_type_field_id"`
	SortOrder       int     `json:"sort_order"`
	IsActive        bool    `json:"is_active"`
}

type labDeviceItemMasterEnsureResponse struct {
	InsertedCount int64                         `json:"inserted_count"`
	Items         []labDeviceItemMasterResponse `json:"items"`
}

func toLabDeviceItemMasterResponse(item *model.LabDeviceItemMaster) labDeviceItemMasterResponse {
	return labDeviceItemMasterResponse{
		ID:              item.ID,
		SourceType:      item.SourceType,
		DeviceItemCode:  item.DeviceItemCode,
		Unit:            item.Unit,
		ValueShape:      item.ValueShape,
		ExamTypeFieldID: item.ExamTypeFieldID,
		SortOrder:       item.SortOrder,
		IsActive:        item.IsActive,
	}
}

type labDeviceFramesRequest struct {
	PayloadBase64 string `json:"payload_base64" binding:"required"`
	DeviceHint    string `json:"device_hint"`
}

type labDeviceWaitRequest struct {
	PetID uint64 `json:"pet_id" binding:"required"`
}

type labDeviceAttachRequest struct {
	PetID uint64 `json:"pet_id" binding:"required"`
}

type labDeviceStationRequest struct {
	SlotsJSON string `json:"slots_json" binding:"required"`
}

type labDeviceWaitResponse struct {
	PetID     uint64    `json:"pet_id"`
	PetName   string    `json:"pet_name"`
	StaffID   uint64    `json:"staff_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type labDeviceStationResponse struct {
	WaitTTLSeconds int    `json:"wait_ttl_seconds"`
	SlotsJSON      string `json:"slots_json"`
}

type labDeviceJobItemResponse struct {
	DeviceItemCode  string  `json:"device_item_code"`
	ValueRaw        string  `json:"value_raw"`
	Unit            string  `json:"unit"`
	Flag            string  `json:"flag"`
	ExamTypeFieldID *uint64 `json:"exam_type_field_id,omitempty"`
	NeedsReview     bool    `json:"needs_review"`
	SortOrder       int     `json:"sort_order"`
}

type labDeviceJobCardResponse struct {
	JobID             uuid.UUID                  `json:"job_id"`
	SourceType        string                     `json:"source_type"`
	DeviceHint        string                     `json:"device_hint"`
	Status            string                     `json:"status"`
	PetID             *uint64                    `json:"pet_id,omitempty"`
	PetName           string                     `json:"pet_name,omitempty"`
	MeasuredAt        *time.Time                 `json:"measured_at,omitempty"`
	ReceivedAt        *time.Time                 `json:"received_at,omitempty"`
	SpecimenIDRaw     string                     `json:"specimen_id_raw"`
	ItemCount         int                        `json:"item_count"`
	UnmappedItemCount int                        `json:"unmapped_item_count"`
	ClockSkew         bool                       `json:"clock_skew"`
	Items             []labDeviceJobItemResponse `json:"items"`
	// ReviewReason は needs_review の原因コード。非 nil のとき UI で理由を表示する（F-1）。
	ReviewReason *string `json:"review_reason,omitempty"`
}

type labDeviceFrameResultResponse struct {
	Duplicate bool                     `json:"duplicate"`
	Job       labDeviceJobCardResponse `json:"job"`
}

type labDeviceReceiveResponse struct {
	Results []labDeviceFrameResultResponse `json:"results"`
}

type labDeviceTodayVisitResponse struct {
	RecordID      uint64 `json:"record_id"`
	PetID         uint64 `json:"pet_id"`
	PetName       string `json:"pet_name"`
	OwnerName     string `json:"owner_name"`
	Species       string `json:"species"`
	DoctorName    string `json:"doctor_name"`
	VisitType     string `json:"visit_type"`
	PetIsDeceased bool   `json:"pet_is_deceased,omitempty"`
}

type labDeviceBoardResponse struct {
	Wait        *labDeviceWaitResponse        `json:"wait"`
	Unlinked    []labDeviceJobCardResponse    `json:"unlinked"`
	Saved       []labDeviceJobCardResponse    `json:"saved"`
	Received    []labDeviceJobCardResponse    `json:"received"`
	TodayVisits []labDeviceTodayVisitResponse `json:"today_visits"`
	Station     labDeviceStationResponse      `json:"station"`
}

type createLabDeviceRequest struct {
	Name       string  `json:"name" binding:"required"`
	SourceType string  `json:"source_type" binding:"required"`
	ExamTypeID *uint64 `json:"exam_type_id"`
	IsActive   *bool   `json:"is_active"`
	SortOrder  int     `json:"sort_order"`
}

func (r createLabDeviceRequest) toServiceInput() CreateLabDeviceInput {
	active := true
	if r.IsActive != nil {
		active = *r.IsActive
	}
	return CreateLabDeviceInput{
		Name:       r.Name,
		SourceType: r.SourceType,
		ExamTypeID: r.ExamTypeID,
		IsActive:   active,
		SortOrder:  r.SortOrder,
	}
}

// is_active / sort_order はポインタ + required。非ポインタだと省略が
// ゼロ値（無効化・並び順 0）として黙って書き込まれる。
type updateLabDeviceRequest struct {
	Name       string  `json:"name" binding:"required"`
	ExamTypeID *uint64 `json:"exam_type_id"`
	IsActive   *bool   `json:"is_active" binding:"required"`
	SortOrder  *int    `json:"sort_order" binding:"required"`
}

func (r updateLabDeviceRequest) toServiceInput() UpdateLabDeviceInput {
	return UpdateLabDeviceInput{
		Name:       r.Name,
		ExamTypeID: r.ExamTypeID,
		IsActive:   *r.IsActive,
		SortOrder:  *r.SortOrder,
	}
}

type saveLabDeviceConfigurationItemRequest struct {
	ID              uint64  `json:"id" binding:"required"`
	Unit            string  `json:"unit"`
	ExamTypeFieldID *uint64 `json:"exam_type_field_id"`
	IsActive        *bool   `json:"is_active" binding:"required"`
}

type saveLabDeviceConfigurationRequest struct {
	Device updateLabDeviceRequest                  `json:"device" binding:"required"`
	Items  []saveLabDeviceConfigurationItemRequest `json:"items" binding:"required"`
}

func (r saveLabDeviceConfigurationRequest) toServiceInput() SaveLabDeviceConfigurationInput {
	items := make([]SaveLabDeviceConfigurationItemInput, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, SaveLabDeviceConfigurationItemInput{
			ID: item.ID,
			UpdateLabDeviceItemMasterInput: UpdateLabDeviceItemMasterInput{
				Unit:            item.Unit,
				ExamTypeFieldID: item.ExamTypeFieldID,
				IsActive:        *item.IsActive,
			},
		})
	}
	return SaveLabDeviceConfigurationInput{Device: r.Device.toServiceInput(), Items: items}
}

type saveLabDeviceConfigurationResponse struct {
	Device labDeviceResponse             `json:"device"`
	Items  []labDeviceItemMasterResponse `json:"items"`
}

type labDeviceResponse struct {
	ID         uint64  `json:"id"`
	SourceType string  `json:"source_type"`
	Name       string  `json:"name"`
	ExamTypeID *uint64 `json:"exam_type_id"`
	IsActive   bool    `json:"is_active"`
	SortOrder  int     `json:"sort_order"`
}

func toLabDeviceResponse(device *model.LabDevice) labDeviceResponse {
	return labDeviceResponse{
		ID:         device.ID,
		SourceType: device.SourceType,
		Name:       device.Name,
		ExamTypeID: device.ExamTypeID,
		IsActive:   device.IsActive,
		SortOrder:  device.SortOrder,
	}
}

func decodeLabDevicePayloadBase64(encoded string) ([]byte, error) {
	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("payload_base64 is invalid")
	}
	if len(payload) == 0 || len(payload) > labDeviceMaxPayloadBytes {
		return nil, apperrors.WrapInvalidInput("invalid_payload")
	}
	return payload, nil
}

func toLabDeviceWaitResponse(wait *LabDeviceWaitView) labDeviceWaitResponse {
	return labDeviceWaitResponse{
		PetID:     wait.PetID,
		PetName:   wait.PetName,
		StaffID:   wait.StaffID,
		ExpiresAt: wait.ExpiresAt,
	}
}

func toLabDeviceStationResponse(station *LabDeviceStationView) labDeviceStationResponse {
	return labDeviceStationResponse{
		WaitTTLSeconds: station.WaitTTLSeconds,
		SlotsJSON:      station.SlotsJSON,
	}
}

func toLabDeviceJobCardResponse(card *LabDeviceJobCard) labDeviceJobCardResponse {
	items := make([]labDeviceJobItemResponse, 0, len(card.Items))
	for i := range card.Items {
		items = append(items, labDeviceJobItemResponse{
			DeviceItemCode:  card.Items[i].DeviceItemCode,
			ValueRaw:        card.Items[i].ValueRaw,
			Unit:            card.Items[i].Unit,
			Flag:            card.Items[i].Flag,
			ExamTypeFieldID: card.Items[i].ExamTypeFieldID,
			NeedsReview:     card.Items[i].NeedsReview,
			SortOrder:       card.Items[i].SortOrder,
		})
	}
	return labDeviceJobCardResponse{
		JobID:             card.JobID,
		SourceType:        string(card.SourceType),
		DeviceHint:        card.DeviceHint,
		Status:            string(card.Status),
		PetID:             card.PetID,
		PetName:           card.PetName,
		MeasuredAt:        card.MeasuredAt,
		ReceivedAt:        card.ReceivedAt,
		SpecimenIDRaw:     card.SpecimenIDRaw,
		ItemCount:         card.ItemCount,
		UnmappedItemCount: card.UnmappedItemCount,
		ClockSkew:         card.ClockSkew,
		Items:             items,
		ReviewReason:      card.ReviewReason,
	}
}

func toLabDeviceReceiveResponse(result *LabDeviceReceiveResult) labDeviceReceiveResponse {
	rows := make([]labDeviceFrameResultResponse, 0, len(result.Results))
	for i := range result.Results {
		card := result.Results[i].Job
		rows = append(rows, labDeviceFrameResultResponse{
			Duplicate: result.Results[i].Duplicate,
			Job:       toLabDeviceJobCardResponse(&card),
		})
	}
	return labDeviceReceiveResponse{Results: rows}
}

func toLabDeviceBoardResponse(board *LabDeviceBoard) labDeviceBoardResponse {
	var wait *labDeviceWaitResponse
	if board.Wait != nil {
		mapped := toLabDeviceWaitResponse(board.Wait)
		wait = &mapped
	}
	return labDeviceBoardResponse{
		Wait:        wait,
		Unlinked:    mapLabDeviceJobCards(board.Unlinked),
		Saved:       mapLabDeviceJobCards(board.Saved),
		Received:    mapLabDeviceJobCards(board.Received),
		TodayVisits: mapLabDeviceTodayVisits(board.TodayVisits),
		Station:     toLabDeviceStationResponse(&board.Station),
	}
}

func mapLabDeviceTodayVisits(visits []LabDeviceTodayVisit) []labDeviceTodayVisitResponse {
	out := make([]labDeviceTodayVisitResponse, 0, len(visits))
	for i := range visits {
		out = append(out, labDeviceTodayVisitResponse{
			RecordID:      visits[i].RecordID,
			PetID:         visits[i].PetID,
			PetName:       visits[i].PetName,
			OwnerName:     visits[i].OwnerName,
			Species:       visits[i].Species,
			DoctorName:    visits[i].DoctorName,
			VisitType:     visits[i].VisitType,
			PetIsDeceased: visits[i].PetIsDeceased,
		})
	}
	return out
}

func mapLabDeviceJobCards(cards []LabDeviceJobCard) []labDeviceJobCardResponse {
	out := make([]labDeviceJobCardResponse, 0, len(cards))
	for i := range cards {
		out = append(out, toLabDeviceJobCardResponse(&cards[i]))
	}
	return out
}
