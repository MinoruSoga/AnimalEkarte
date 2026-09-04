package billing

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func validateTrimmingBillingItemRefs(
	tx *gorm.DB,
	clinicID uint64,
	treatmentID, trimmingCourseID, trimmingOptionID, effectiveAppointmentID *uint64,
	billingRef billingItemBillingReference,
	medicalRecordRef *billingItemMedicalRecordReference,
) error {
	if effectiveAppointmentID != nil {
		var appointmentRef billingItemAppointmentReference
		if err := tx.
			Table("appointments").
			Select("owner_id", "pet_id").
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", *effectiveAppointmentID, clinicID).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Take(&appointmentRef).Error; err != nil {
			return apperrors.FromGORM(err, "appointment", fmt.Sprintf("%d", *effectiveAppointmentID))
		}
		if !sameOptionalBillingReference(billingRef.OwnerID, appointmentRef.OwnerID) ||
			!sameOptionalBillingReference(billingRef.PetID, appointmentRef.PetID) {
			return invalidBillingItemReferenceCombination()
		}
		enforceMedicalAppointment := (medicalRecordRef != nil && treatmentID != nil) ||
			(medicalRecordRef != nil && trimmingCourseID == nil && trimmingOptionID == nil)
		if enforceMedicalAppointment &&
			(medicalRecordRef.AppointmentID == nil || *medicalRecordRef.AppointmentID != *effectiveAppointmentID) {
			return invalidBillingItemReferenceCombination()
		}
	}
	if (trimmingCourseID != nil || trimmingOptionID != nil) && effectiveAppointmentID == nil {
		return invalidBillingItemReferenceCombination()
	}
	if trimmingCourseID != nil {
		var id uint64
		if err := tx.
			Table("appointment_trimming_details").
			Select("trimming_courses.id").
			Joins("JOIN appointments ON appointments.id = appointment_trimming_details.appointment_id AND appointments.clinic_id = appointment_trimming_details.clinic_id").
			Joins("JOIN trimming_courses ON trimming_courses.id = appointment_trimming_details.course_id AND trimming_courses.clinic_id = appointment_trimming_details.clinic_id AND trimming_courses.deleted_at IS NULL").
			Where("appointment_trimming_details.appointment_id = ? AND appointment_trimming_details.clinic_id = ? AND appointment_trimming_details.course_id = ?", *effectiveAppointmentID, clinicID, *trimmingCourseID).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Take(&id).Error; err != nil {
			return apperrors.FromGORM(err, "trimming_course", fmt.Sprintf("%d", *trimmingCourseID))
		}
	}
	if trimmingOptionID != nil {
		var id uint64
		if err := tx.
			Table("appointment_trimming_options AS ato").
			Select("topt.id").
			Joins("JOIN appointments AS a ON a.id = ato.appointment_id AND a.clinic_id = ? AND a.deleted_at IS NULL", clinicID).
			Joins("JOIN trimming_options AS topt ON topt.id = ato.option_id AND topt.clinic_id = a.clinic_id AND topt.deleted_at IS NULL").
			Where("ato.appointment_id = ? AND ato.option_id = ?", *effectiveAppointmentID, *trimmingOptionID).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Take(&id).Error; err != nil {
			return apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", *trimmingOptionID))
		}
	}
	return nil
}
