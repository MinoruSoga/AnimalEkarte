package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type reservationStaffResponse struct {
	ID                  uint64                   `json:"id"`
	Name                string                   `json:"name"`
	IsActive            bool                     `json:"is_active"`
	SortOrder           int                      `json:"sort_order"`
	StaffType           string                   `json:"staff_type"`
	ReservationVisible  bool                     `json:"reservation_visible"`
	ReservationComment  string                   `json:"reservation_comment"`
	ReservationImageURL string                   `json:"reservation_image_url"`
	ExcludedCourses     []excludedCourseResponse `json:"excluded_courses"`
	CreatedAt           time.Time                `json:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at"`
}

type excludedCourseResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func toReservationStaffResponse(staff *model.Staff, excluded []model.StaffExcludedReservationCategory) reservationStaffResponse {
	courses := make([]excludedCourseResponse, 0, len(excluded))
	for _, e := range excluded {
		if e.ReservationType != nil {
			courses = append(courses, excludedCourseResponse{
				ID:   e.ReservationTypeID,
				Name: e.ReservationType.Name,
			})
		} else {
			courses = append(courses, excludedCourseResponse{ID: e.ReservationTypeID})
		}
	}
	return reservationStaffResponse{
		ID:                  staff.ID,
		Name:                staff.Name,
		IsActive:            staff.IsActive,
		SortOrder:           staff.SortOrder,
		StaffType:           string(staff.StaffType),
		ReservationVisible:  staff.ReservationVisible,
		ReservationComment:  staff.ReservationComment,
		ReservationImageURL: staff.ReservationImageURL,
		ExcludedCourses:     courses,
		CreatedAt:           staff.CreatedAt,
		UpdatedAt:           staff.UpdatedAt,
	}
}
