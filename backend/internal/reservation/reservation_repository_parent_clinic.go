package reservation

import "gorm.io/gorm"

// reservationRelationsMatchParentClinicSQL は、各 appointment 自身の clinic_id と関連行の
// clinic_id を相関させる。soft-delete 済みの同一 clinic 関連と過去の staff assignment は予約履歴の
// 親行/countを維持するため許容し、現在の関連を応答へ表示するかは Preload 側の条件に委ねる。
// cross-clinic FK と通常予約の Owner/Pet 不一致は、一覧/単件とも親行ごと fail-closed にする。
// tentative はペットに属する意図のため、同医院での飼主移管後も現在の pet owner に基づき表示する。
// created_by は登録時に認可済みの履歴であり、現在の所属で親予約を非表示にしない。
const reservationRelationsMatchParentClinicSQL = `
		EXISTS (
			SELECT 1
			FROM reservation_types scoped_reservation_type
			WHERE scoped_reservation_type.id = appointments.reservation_type_id
			  AND scoped_reservation_type.clinic_id = appointments.clinic_id
			  AND (
				scoped_reservation_type.group_id IS NULL
				OR EXISTS (
					SELECT 1
					FROM reservation_type_groups scoped_reservation_type_group
					WHERE scoped_reservation_type_group.id = scoped_reservation_type.group_id
					  AND scoped_reservation_type_group.clinic_id = appointments.clinic_id
				)
			  )
		)
		AND (
			appointments.owner_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM owners scoped_owner
				WHERE scoped_owner.id = appointments.owner_id
				  AND scoped_owner.clinic_id = appointments.clinic_id
			)
		)
		AND (
			appointments.pet_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM pets scoped_pet
				WHERE scoped_pet.id = appointments.pet_id
				  AND scoped_pet.clinic_id = appointments.clinic_id
				  AND (
					appointments.owner_id IS NULL
					OR appointments.status = 'tentative'
					OR scoped_pet.owner_id = appointments.owner_id
				  )
				  AND EXISTS (
					SELECT 1
					FROM owners scoped_pet_owner
					WHERE scoped_pet_owner.id = scoped_pet.owner_id
					  AND scoped_pet_owner.clinic_id = appointments.clinic_id
				  )
			)
		)
		AND (
			appointments.line_customer_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM line_customers scoped_line_customer
				WHERE scoped_line_customer.id = appointments.line_customer_id
				  AND scoped_line_customer.clinic_id = appointments.clinic_id
			)
		)
		AND (
			appointments.doctor_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM staffs scoped_doctor
				JOIN staff_clinic_assignments scoped_doctor_assignment
				  ON scoped_doctor_assignment.staff_id = scoped_doctor.id
				 AND scoped_doctor_assignment.clinic_id = appointments.clinic_id
				WHERE scoped_doctor.id = appointments.doctor_id
			)
		)
	`

func reservationRelationsMatchParentClinic(q *gorm.DB) *gorm.DB {
	return q.Where(reservationRelationsMatchParentClinicSQL)
}

// A tentative belongs to the pet's current owner; real bookings retain their
// stored owner filter. Correlate both the pet and current owner to the row clinic.
func reservationOwnerFilter(ownerID uint64) func(*gorm.DB) *gorm.DB {
	return func(q *gorm.DB) *gorm.DB {
		return q.Where(`
			(appointments.status <> 'tentative' AND appointments.owner_id = ?)
			OR (appointments.status = 'tentative' AND EXISTS (
				SELECT 1 FROM pets intent_pet
				JOIN owners intent_owner ON intent_owner.id = intent_pet.owner_id
				  AND intent_owner.clinic_id = appointments.clinic_id
				  AND intent_owner.deleted_at IS NULL
				WHERE intent_pet.id = appointments.pet_id
				  AND intent_pet.clinic_id = appointments.clinic_id
				  AND intent_pet.deleted_at IS NULL
				  AND intent_pet.owner_id = ?
			))`, ownerID, ownerID)
	}
}
