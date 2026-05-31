package handler

import "github.com/animal-ekarte/backend/internal/service"

type createReservationStaffRequest struct {
	Name               string   `json:"name"               binding:"required"`
	StaffType          string   `json:"staff_type" binding:"omitempty,oneof=doctor nurse groomer other"`
	ReservationVisible bool     `json:"reservation_visible"`
	ReservationComment string   `json:"reservation_comment"`
	SortOrder          int      `json:"sort_order"`
	ExcludedTypeIDs    []uint64 `json:"excluded_type_ids"`
}

func (r *createReservationStaffRequest) toServiceInput() *service.CreateReservationStaffInput {
	return &service.CreateReservationStaffInput{
		Name:               r.Name,
		StaffType:          r.StaffType,
		ReservationVisible: r.ReservationVisible,
		ReservationComment: r.ReservationComment,
		SortOrder:          r.SortOrder,
		ExcludedTypeIDs:    r.ExcludedTypeIDs,
	}
}

type updateReservationStaffRequest struct {
	Name               *string   `json:"name"`
	StaffType          *string   `json:"staff_type" binding:"omitempty,oneof=doctor nurse groomer other"`
	ReservationVisible *bool     `json:"reservation_visible"`
	ReservationComment *string   `json:"reservation_comment"`
	SortOrder          *int      `json:"sort_order"`
	ExcludedTypeIDs    *[]uint64 `json:"excluded_type_ids"`
}

func (r updateReservationStaffRequest) toServiceInput() *service.UpdateReservationStaffInput {
	return &service.UpdateReservationStaffInput{
		Name:               r.Name,
		StaffType:          r.StaffType,
		ReservationVisible: r.ReservationVisible,
		ReservationComment: r.ReservationComment,
		SortOrder:          r.SortOrder,
		ExcludedTypeIDs:    r.ExcludedTypeIDs,
	}
}

type patchReservationStaffStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type patchReservationStaffSortOrderRequest struct {
	Direction string `json:"direction" binding:"required,oneof=up down"`
}
