package service

import "github.com/animal-ekarte/backend/internal/reservation"

// BE9-2C R③ transitional aliases — timeslot engine 系の型は internal/reservation へ移動済み。
// 残留 consumer（liff 系=R⑤・appointment 系=R④）互換のための型 alias。REMOVE: R④/R⑤移動時。
type (
	TimeSlot               = reservation.TimeSlot
	BusinessHours          = reservation.BusinessHours
	BreakPeriod            = reservation.BreakPeriod
	AvailableDateResult    = reservation.AvailableDateResult
	AvailableDatesSettings = reservation.AvailableDatesSettings
	StaffSlotInput         = reservation.StaffSlotInput
	TimeSlotsInput         = reservation.TimeSlotsInput
	StaffScheduleOverride  = reservation.StaffScheduleOverride
	ExistingReservation    = reservation.ExistingReservation
	AvailableDatesInput    = reservation.AvailableDatesInput
	BookingWindow          = reservation.BookingWindow

	CreateReservationInput       = reservation.CreateReservationInput
	CreateManualReservationInput = reservation.CreateManualReservationInput
	UpdateReservationInput       = reservation.UpdateReservationInput
	UpdateReservationRouteInput  = reservation.UpdateReservationRouteInput
	ReservationValidators        = reservation.ReservationValidators
	ReservationLimitError        = reservation.ReservationLimitError
	ReservationService           = reservation.ReservationService
)

// reservationTypeFinder は appointment/liff 系（R④/R⑤）残留 consumer 用の view alias。REMOVE: R④⑤移動時。
type reservationTypeFinder = reservation.ReservationTypeFinderView
