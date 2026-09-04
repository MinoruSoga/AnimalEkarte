package medicalrecord

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/textsearch"
)

func applyMedicalRecordSearch(q *gorm.DB, search string) *gorm.DB {
	qSearch := textsearch.NormalizeQuerySpaces(search)
	if qSearch == "" {
		return q.Where("1 = 0")
	}
	rawPattern := "%" + textsearch.EscapeLike(qSearch) + "%"
	normalizedPattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(qSearch)) + "%"
	return q.Where(
		`(medical_records.record_no ILIKE ? ESCAPE '\'`+
			` OR owners.name ILIKE ? ESCAPE '\'`+
			` OR translate(owners.name, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR translate(owners.name, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR translate(owners.name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR pets.name ILIKE ? ESCAPE '\'`+
			` OR translate(pets.name, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR translate(pets.name, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR translate(pets.name_kana, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR inquiries.chief_complaint ILIKE ? ESCAPE '\'`+
			` OR EXISTS (`+
			`SELECT 1 FROM treatments searched_treatment`+
			` WHERE searched_treatment.medical_record_id = medical_records.id`+
			` AND searched_treatment.deleted_at IS NULL`+
			` AND (`+
			`translate(searched_treatment.content, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR translate(searched_treatment.memo, ?, ?) ILIKE ? ESCAPE '\'`+
			` OR EXISTS (`+
			`SELECT 1 FROM procedures searched_procedure`+
			` WHERE searched_procedure.id = searched_treatment.procedure_id`+
			` AND searched_procedure.clinic_id = medical_records.clinic_id`+
			` AND searched_procedure.deleted_at IS NULL`+
			` AND translate(searched_procedure.name, ?, ?) ILIKE ? ESCAPE '\')`+
			` OR EXISTS (`+
			`SELECT 1 FROM medicines searched_medicine`+
			` WHERE searched_medicine.id = searched_treatment.medicine_id`+
			` AND searched_medicine.clinic_id = medical_records.clinic_id`+
			` AND searched_medicine.deleted_at IS NULL`+
			` AND translate(searched_medicine.name, ?, ?) ILIKE ? ESCAPE '\')`+
			` OR EXISTS (`+
			`SELECT 1 FROM consultations searched_consultation`+
			` WHERE searched_consultation.id = searched_treatment.consultation_id`+
			` AND searched_consultation.clinic_id = medical_records.clinic_id`+
			` AND searched_consultation.deleted_at IS NULL`+
			` AND translate(searched_consultation.name, ?, ?) ILIKE ? ESCAPE '\')`+
			` OR EXISTS (`+
			`SELECT 1 FROM inventory_items searched_inventory`+
			` WHERE searched_inventory.id = searched_treatment.inventory_id`+
			` AND searched_inventory.clinic_id = medical_records.clinic_id`+
			` AND searched_inventory.deleted_at IS NULL`+
			` AND translate(searched_inventory.name, ?, ?) ILIKE ? ESCAPE '\'))))`,
		normalizedPattern,
		rawPattern,
		textsearch.SpaceSourceChars, textsearch.SpaceTargetChars, rawPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		rawPattern,
		textsearch.SpaceSourceChars, textsearch.SpaceTargetChars, rawPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		normalizedPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
		textsearch.KanaAndSpaceSourceChars, textsearch.KanaAndSpaceTargetChars, normalizedPattern,
	)
}
