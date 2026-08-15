package reservation

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
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
	// ExcludedCourses is a Stage B compatibility facade derived from capabilities.
	ExcludedCourses []excludedCourseResponse `json:"excluded_courses"`
	// CapableCourses is the affirmative capability surface (TASK-021 Stage B).
	CapableCourses []capableCourseResponse `json:"capable_courses"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

type excludedCourseResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type capableCourseResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func toReservationStaffResponse(
	staff *model.Staff,
	excluded []model.StaffReservationExclusion,
	capable []model.StaffReservationCapability,
) reservationStaffResponse {
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
	capableCourses := make([]capableCourseResponse, 0, len(capable))
	for _, c := range capable {
		if c.ReservationType != nil {
			capableCourses = append(capableCourses, capableCourseResponse{
				ID:   c.ReservationTypeID,
				Name: c.ReservationType.Name,
			})
		} else {
			capableCourses = append(capableCourses, capableCourseResponse{ID: c.ReservationTypeID})
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
		CapableCourses:      capableCourses,
		CreatedAt:           httpapi.LocalTime(staff.CreatedAt),
		UpdatedAt:           httpapi.LocalTime(staff.UpdatedAt),
	}
}
