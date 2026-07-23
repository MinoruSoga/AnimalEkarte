package service

import (
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/trimming"
)

// BE9-2E compatibility aliases; remove with the service aggregator in BE9-2F.
type CreateTrimmingInput = trimming.CreateTrimmingInput
type UpdateTrimmingInput = trimming.UpdateTrimmingInput
type TrimmingService = trimming.TrimmingService

func NewTrimmingService(
	reservationRepo trimming.TrimmingReservationRepository,
	reservationType repository.ReservationTypeRepository,
	reservationStaff repository.ReservationStaffRepository,
	unavailableTime repository.ReservationTypeUnavailableTimeRepository,
	_ repository.ReservationTypeAvailableSlotRepository,
	trimmingDetail repository.AppointmentTrimmingDetailRepository,
	trimmingCourseRepo repository.TrimmingCourseRepository,
	trimmingOptionRepo repository.TrimmingOptionRepository,
	transactor repository.Transactor,
	auditTx AuditTxLogger,
) TrimmingService {
	var trimmingAudit trimming.AuditTxLogger
	if auditTx != nil {
		trimmingAudit = trimmingAuditTxAdapter{inner: auditTx}
	}
	return trimming.NewTrimmingServiceWithAudit(
		reservationRepo,
		reservationType,
		reservationStaff,
		unavailableTime,
		trimmingDetail,
		trimmingCourseRepo,
		trimmingOptionRepo,
		transactor,
		trimmingAudit,
	)
}
