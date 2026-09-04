package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// Unbilled candidate queries for BillingItemRepository.
// Split from billing_item_repository.go (ARCH-A4-billing S2) — behavior unchanged.

func (r *billingItemRepository) FindUnbilledVaccinationItemsByPetID(
	ctx context.Context,
	clinicID, petID uint64,
) ([]model.BillingItem, int, error) {
	type vaccinationBillingRow struct {
		VaccinationID uint64
		VaccineID     *uint64
		Name          *string
		UnitPrice     *int64
	}
	rows := make([]vaccinationBillingRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			vaccination.id AS vaccination_id,
			vaccine.id AS vaccine_id,
			vaccine.name AS name,
			vaccine.price AS unit_price
		FROM vaccinations AS vaccination
		JOIN pets AS pet
		  ON pet.id = vaccination.pet_id
		 AND pet.clinic_id = vaccination.clinic_id
		 AND pet.deleted_at IS NULL
		JOIN owners AS owner
		  ON owner.id = pet.owner_id
		 AND owner.clinic_id = vaccination.clinic_id
		 AND owner.deleted_at IS NULL
		LEFT JOIN vaccines AS vaccine
		  ON vaccine.id = vaccination.vaccine_id
		 AND vaccine.clinic_id = vaccination.clinic_id
		 AND vaccine.deleted_at IS NULL
		WHERE vaccination.clinic_id = ?
		  AND vaccination.pet_id = ?
		  AND vaccination.deleted_at IS NULL
		  AND vaccination.medical_record_id IS NOT NULL
		  AND EXISTS (
		      SELECT 1
		      FROM medical_records AS medical_record
		      WHERE medical_record.id = vaccination.medical_record_id
		        AND medical_record.clinic_id = vaccination.clinic_id
		        AND medical_record.pet_id = vaccination.pet_id
		        AND medical_record.deleted_at IS NULL
		  )
		  AND EXISTS (
		      SELECT 1
		      FROM billing_confirmations AS billing_confirmation
		      WHERE billing_confirmation.medical_record_id = vaccination.medical_record_id
		        AND billing_confirmation.status = 'confirmed'
		  )
		  AND NOT EXISTS (
		      SELECT 1
		      FROM billing_items AS billing_item
		      WHERE billing_item.vaccination_id = vaccination.id
		        AND billing_item.clinic_id = vaccination.clinic_id
		        AND billing_item.deleted_at IS NULL
		  )
		ORDER BY vaccination.date ASC, vaccination.id ASC
	`, clinicID, petID).Scan(&rows).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "vaccination", fmt.Sprintf("pet:%d", petID))
	}

	items := make([]model.BillingItem, 0, len(rows))
	unbillableCount := 0
	for i := range rows {
		row := &rows[i]
		if row.VaccineID == nil ||
			row.Name == nil ||
			strings.TrimSpace(*row.Name) == "" ||
			row.UnitPrice == nil ||
			*row.UnitPrice < 0 {
			// BUG-013: data-quality は skip + count（typed warning 入力）。infra ではない。
			unbillableCount++
			continue
		}
		vaccinationID := row.VaccinationID
		items = append(items, model.BillingItem{
			ID:            vaccinationID,
			Category:      model.ItemCategoryVaccine,
			Name:          *row.Name,
			UnitPrice:     *row.UnitPrice,
			Quantity:      1,
			TaxType:       model.TaxTypeExcluded,
			TaxRate:       sharedkernel.DefaultTaxRate,
			Source:        model.ItemSourceMedicalRecord,
			VaccinationID: &vaccinationID,
		})
	}
	return items, unbillableCount, nil
}

func (r *billingItemRepository) FindUnbilledExamItemsByPetID(
	ctx context.Context,
	clinicID, petID uint64,
) ([]model.BillingItem, int, error) {
	type examBillingRow struct {
		ExamID        uint64
		ExamTypeID    *uint64
		Name          *string
		UnitPrice     *int64
		MedicalRecord uint64
	}
	rows := make([]examBillingRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			exam.id AS exam_id,
			exam_type.id AS exam_type_id,
			exam_type.name AS name,
			exam_type.price AS unit_price,
			exam.medical_record_id AS medical_record
		FROM exams AS exam
		JOIN pets AS pet
		  ON pet.id = exam.pet_id
		 AND pet.clinic_id = exam.clinic_id
		 AND pet.deleted_at IS NULL
		JOIN medical_records AS medical_record
		  ON medical_record.id = exam.medical_record_id
		 AND medical_record.clinic_id = exam.clinic_id
		 AND medical_record.pet_id = exam.pet_id
		 AND medical_record.deleted_at IS NULL
		JOIN billing_confirmations AS billing_confirmation
		  ON billing_confirmation.medical_record_id = exam.medical_record_id
		 AND billing_confirmation.status = 'confirmed'
		LEFT JOIN exam_types AS exam_type
		  ON exam_type.id = exam.exam_type_id
		 AND exam_type.clinic_id = exam.clinic_id
		 AND exam_type.deleted_at IS NULL
		WHERE exam.clinic_id = ?
		  AND exam.pet_id = ?
		  AND exam.deleted_at IS NULL
		  AND exam.medical_record_id IS NOT NULL
		  AND NOT EXISTS (
		      SELECT 1
		      FROM billing_items AS billing_item
		      WHERE billing_item.exam_id = exam.id
		        AND billing_item.clinic_id = exam.clinic_id
		        AND billing_item.deleted_at IS NULL
		  )
		ORDER BY exam.date ASC, exam.id ASC
	`, clinicID, petID).Scan(&rows).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "exam", fmt.Sprintf("pet:%d", petID))
	}

	items := make([]model.BillingItem, 0, len(rows))
	unbillableCount := 0
	for i := range rows {
		row := &rows[i]
		if row.ExamTypeID == nil ||
			row.Name == nil ||
			strings.TrimSpace(*row.Name) == "" ||
			row.UnitPrice == nil ||
			*row.UnitPrice < 0 {
			unbillableCount++
			continue
		}
		examID := row.ExamID
		medicalRecordID := row.MedicalRecord
		items = append(items, model.BillingItem{
			ID:              examID,
			Category:        model.ItemCategoryTest,
			Name:            *row.Name,
			UnitPrice:       *row.UnitPrice,
			Quantity:        1,
			TaxType:         model.TaxTypeExcluded,
			TaxRate:         sharedkernel.DefaultTaxRate,
			Source:          model.ItemSourceMedicalRecord,
			ExamID:          &examID,
			MedicalRecordID: &medicalRecordID,
		})
	}
	return items, unbillableCount, nil
}

func (r *billingItemRepository) FindUnbilledTrimmingItemsByPetID(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error) {
	type row struct {
		AppointmentID    uint64
		OriginID         uint64
		Name             string
		UnitPrice        int64
		SortOrder        int
		TrimmingCourseID *uint64
		TrimmingOptionID *uint64
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			a.id AS appointment_id,
			tc.id AS origin_id,
			tc.name AS name,
			COALESCE(tc.price, 0)::bigint AS unit_price,
			0 AS sort_order,
			tc.id AS trimming_course_id,
			NULL::bigint AS trimming_option_id
		FROM appointment_trimming_details atd
		JOIN appointments a ON a.id = atd.appointment_id AND atd.clinic_id = a.clinic_id AND a.deleted_at IS NULL
		JOIN reservation_types rt ON rt.id = a.reservation_type_id AND rt.clinic_id = a.clinic_id AND rt.deleted_at IS NULL
		JOIN trimming_courses tc ON tc.id = atd.course_id AND tc.clinic_id = a.clinic_id AND tc.deleted_at IS NULL
		WHERE a.clinic_id = ?
		  AND a.pet_id = ?
		  AND EXISTS (SELECT 1 FROM pets p WHERE p.id = a.pet_id AND p.clinic_id = a.clinic_id)
		  AND a.status = ?
		  AND rt.category = ?
		  AND COALESCE(tc.price, 0) > 0
		  AND NOT EXISTS (
		      SELECT 1
		      FROM billing_items bi
		      JOIN billings b ON b.id = bi.billing_id AND b.clinic_id = a.clinic_id AND b.deleted_at IS NULL
		      WHERE bi.appointment_id = a.id
		        AND bi.trimming_course_id = tc.id
		        AND bi.deleted_at IS NULL
		        AND b.status != ?
		  )
		UNION ALL
		SELECT
			a.id AS appointment_id,
			topt.id AS origin_id,
			topt.name AS name,
			COALESCE(topt.price, 0)::bigint AS unit_price,
			100 + COALESCE(ato.sort_order, 0) AS sort_order,
			NULL::bigint AS trimming_course_id,
			topt.id AS trimming_option_id
		FROM appointment_trimming_details atd
		JOIN appointments a ON a.id = atd.appointment_id AND atd.clinic_id = a.clinic_id AND a.deleted_at IS NULL
		JOIN reservation_types rt ON rt.id = a.reservation_type_id AND rt.clinic_id = a.clinic_id AND rt.deleted_at IS NULL
		JOIN appointment_trimming_options ato ON ato.appointment_id = a.id
		JOIN trimming_options topt ON topt.id = ato.option_id AND topt.clinic_id = a.clinic_id AND topt.deleted_at IS NULL
		WHERE a.clinic_id = ?
		  AND a.pet_id = ?
		  AND EXISTS (SELECT 1 FROM pets p WHERE p.id = a.pet_id AND p.clinic_id = a.clinic_id)
		  AND a.status = ?
		  AND rt.category = ?
		  AND COALESCE(topt.price, 0) > 0
		  AND NOT EXISTS (
		      SELECT 1
		      FROM billing_items bi
		      JOIN billings b ON b.id = bi.billing_id AND b.clinic_id = a.clinic_id AND b.deleted_at IS NULL
		      WHERE bi.appointment_id = a.id
		        AND bi.trimming_option_id = topt.id
		        AND bi.deleted_at IS NULL
		        AND b.status != ?
		  )
		ORDER BY appointment_id ASC, sort_order ASC, origin_id ASC
	`,
		clinicID, petID, model.ReservationStatusAccounting, model.ReservationTypeCategoryTrimming, model.BillingStatusCancelled,
		clinicID, petID, model.ReservationStatusAccounting, model.ReservationTypeCategoryTrimming, model.BillingStatusCancelled,
	).Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("clinic=%d pet=%d trimming", clinicID, petID))
	}

	items := make([]model.BillingItem, 0, len(rows))
	for i, row := range rows {
		appointmentID := row.AppointmentID
		items = append(items, model.BillingItem{
			ID:        uint64(i + 1),
			BillingID: 0,
			Category: sharedkernel.ResolveItemCategory(sharedkernel.ItemCategoryResolverInput{
				Source: model.ItemSourceTrimming,
			}),
			Name:                  row.Name,
			UnitPrice:             row.UnitPrice,
			Quantity:              1,
			TaxType:               model.TaxTypeExcluded,
			TaxRate:               0.10,
			IsInsuranceApplicable: false,
			Source:                model.ItemSourceTrimming,
			AppointmentID:         &appointmentID,
			TrimmingCourseID:      row.TrimmingCourseID,
			TrimmingOptionID:      row.TrimmingOptionID,
			SortOrder:             row.SortOrder,
		})
	}
	return items, nil
}

// CountNonAccountingTrimmingByPetAndDate は同日同ペットの「未会計対象化」トリミング appointment 件数を返す(#77)。
// トリミング予約区分で status が accounting/completed/cancelled でない = まだ会計待ちに進んでいない取り残し候補。
func (r *billingItemRepository) CountNonAccountingTrimmingByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Reservation{}).
		Joins("JOIN reservation_types rt ON rt.id = appointments.reservation_type_id AND rt.clinic_id = appointments.clinic_id AND rt.deleted_at IS NULL").
		Where("appointments.clinic_id = ? AND appointments.pet_id = ? AND appointments.deleted_at IS NULL", clinicID, petID).
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = appointments.pet_id AND p.clinic_id = appointments.clinic_id)").
		Where("rt.category = ?", model.ReservationTypeCategoryTrimming).
		Where("appointments.status NOT IN ?", []model.ReservationStatus{
			model.ReservationStatusAccounting,
			model.ReservationStatusCompleted,
			model.ReservationStatusCancelled,
		}).
		Where("DATE(appointments.start_time AT TIME ZONE 'Asia/Tokyo') = DATE(? AT TIME ZONE 'Asia/Tokyo')", date).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "appointment", fmt.Sprintf("clinic=%d pet=%d trimming", clinicID, petID))
	}
	return count, nil
}
