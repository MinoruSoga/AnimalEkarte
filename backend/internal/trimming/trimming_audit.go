package trimming

import (
	"context"
	"slices"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	trimmingAuditMutationCreateAppointment = "create_appointment"
	trimmingAuditMutationCreateDetail      = "create_detail"
	trimmingAuditMutationUpdate            = "update"
	trimmingAuditMutationDelete            = "delete"
)

func (s *trimmingService) requireAuditTx() error {
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("trimming audit transaction dependency is required")
	}
	return nil
}

func requireTrimmingStaffAuditActor(actorID *uint64) error {
	if actorID == nil || *actorID == 0 {
		return apperrors.WrapInvalidInput("trimming audit actor_id is required")
	}
	return nil
}

func (s *trimmingService) logTrimmingAuditTx(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action string,
	appointmentID uint64,
	mutationPath string,
	oldValue, newValue any,
) error {
	if err := requireTrimmingStaffAuditActor(actorID); err != nil {
		return err
	}
	if err := s.requireAuditTx(); err != nil {
		return err
	}
	resourceID := appointmentID
	if err := s.auditTx.LogEntryTx(ctx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     action,
		Resource:   model.AuditResourceTrimming,
		ResourceID: &resourceID,
		OldValue:   oldValue,
		NewValue:   newValue,
		Metadata:   map[string]any{"mutation_path": mutationPath},
	}); err != nil {
		return apperrors.Wrap(err, "failed to audit trimming mutation")
	}
	return nil
}

func trimmingAuditValue(
	appointment *model.Reservation,
	detail *model.AppointmentTrimmingDetail,
) map[string]any {
	if appointment == nil {
		return nil
	}
	value := map[string]any{
		"appointment_id":      appointment.ID,
		"reservation_type_id": appointment.ReservationTypeID,
		"start_time":          appointment.StartTime,
		"end_time":            appointment.EndTime,
		"owner_id":            trimmingAuditUint64Value(appointment.OwnerID),
		"pet_id":              trimmingAuditUint64Value(appointment.PetID),
		"doctor_id":           trimmingAuditUint64Value(appointment.DoctorID),
		"status":              appointment.Status,
		"reservation_route":   trimmingAuditStringValue(appointment.ReservationRoute),
		"has_detail":          detail != nil,
	}
	if detail == nil {
		return value
	}
	optionIDs := make([]uint64, len(detail.Options))
	for i := range detail.Options {
		optionIDs[i] = detail.Options[i].ID
	}
	slices.Sort(optionIDs)
	value["course_id"] = trimmingAuditUint64Value(detail.CourseID)
	value["style_request"] = detail.StyleRequest
	value["body_weight"] = trimmingAuditFloat64Value(detail.BodyWeight)
	value["bw_unit"] = detail.BWUnit
	value["body_temperature"] = trimmingAuditFloat64Value(detail.BodyTemperature)
	value["used_shampoo"] = detail.UsedShampoo
	value["used_ribbon"] = detail.UsedRibbon
	value["remarks"] = detail.Remarks
	value["style_image"] = detail.StyleImage
	value["completed_image"] = detail.CompletedImage
	value["option_ids"] = optionIDs
	return value
}

func trimmingAuditUint64Value(value *uint64) any {
	if value == nil {
		return nil
	}
	return *value
}

func trimmingAuditFloat64Value(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func trimmingAuditStringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
