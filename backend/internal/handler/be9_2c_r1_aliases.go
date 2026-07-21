package handler

import (
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

// BE9-2C R① transitional alias — reservationTypeResponse は internal/reservation へ移動済み。
// 未移行の reservation_response.go（R③）互換のための alias。REMOVE: R③ 移動時。
type reservationTypeResponse = reservation.ReservationTypeResponse

func toReservationTypeResponse(st *model.ReservationType) reservationTypeResponse {
	return reservation.ToReservationTypeResponse(st)
}
