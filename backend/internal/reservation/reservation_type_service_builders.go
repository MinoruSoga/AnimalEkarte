package reservation

import (
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	colReservationTypeName                = "name"
	colReservationTypeColor               = "color"
	colReservationTypeIsActive            = "is_active"
	colReservationTypeDescription         = "description"
	colReservationTypeSortOrder           = "sort_order"
	colReservationTypeCategory            = "category"
	colReservationTypeReservationDispName = "reservation_display_name"
	colReservationTypeDurationMinutes     = "duration_minutes"
	colReservationTypeMaxConcurrent       = "max_concurrent"
	colReservationTypeShortName           = "short_name"
	colReservationTypeShowShortName       = "show_short_name"
	colReservationTypeReservationVisible  = "reservation_visible"
	colReservationTypeReservationComment  = "reservation_comment"
	colReservationTypeReservationDayOpt   = "reservation_day_option"
	colReservationTypeIsInternal          = "is_internal"
	colReservationTypeReservationImageURL = "reservation_image_url"
	colReservationTypeGroupID             = "group_id"
	colReservationTypeParentID            = "parent_id"
)

// buildReservationTypeUpdate は UpdateReservationTypeInput から nil でないフィールドのみ map に変換する
func buildReservationTypeUpdate(input *UpdateReservationTypeInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colReservationTypeName] = *input.Name
	}
	if input.Color != nil {
		fields[colReservationTypeColor] = *input.Color
	}
	if input.IsActive != nil {
		fields[colReservationTypeIsActive] = *input.IsActive
	}
	if input.Description != nil {
		fields[colReservationTypeDescription] = *input.Description
	}
	if input.SortOrder != nil {
		fields[colReservationTypeSortOrder] = *input.SortOrder
	}
	if input.Category != nil {
		fields[colReservationTypeCategory] = model.ReservationTypeCategory(*input.Category)
	}
	if input.ReservationDisplayName != nil {
		fields[colReservationTypeReservationDispName] = *input.ReservationDisplayName
	}
	if input.DurationMinutes != nil {
		fields[colReservationTypeDurationMinutes] = *input.DurationMinutes
	}
	if input.ClearMaxConcurrent {
		fields[colReservationTypeMaxConcurrent] = nil
	} else if input.MaxConcurrent != nil {
		fields[colReservationTypeMaxConcurrent] = *input.MaxConcurrent
	}
	if input.ShortName != nil {
		fields[colReservationTypeShortName] = *input.ShortName
	}
	if input.ShowShortName != nil {
		fields[colReservationTypeShowShortName] = *input.ShowShortName
	}
	if input.ReservationVisible != nil {
		fields[colReservationTypeReservationVisible] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields[colReservationTypeReservationComment] = *input.ReservationComment
	}
	if input.ReservationImageURL != nil {
		fields[colReservationTypeReservationImageURL] = *input.ReservationImageURL
	}
	if input.ReservationDayOption != nil {
		fields[colReservationTypeReservationDayOpt] = *input.ReservationDayOption
	}
	if input.IsInternal != nil {
		fields[colReservationTypeIsInternal] = *input.IsInternal
	}
	sharedkernel.SetNullableUint64Field(fields, colReservationTypeGroupID, input.ClearGroupID, input.GroupID)
	sharedkernel.SetNullableUint64Field(fields, colReservationTypeParentID, input.ClearParentID, input.ParentID)
	return fields
}
