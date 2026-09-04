package reservation

import "gorm.io/gorm"

// reservationRelationsMatchParentClinicSQL は、各 appointment 自身の clinic_id と関連行の
// clinic_id を相関させる。soft-delete 済みの同一 clinic 関連と過去の staff assignment は予約履歴の
// 親行/countを維持するため許容し、現在の関連を応答へ表示するかは Preload 側の条件に委ねる。
// cross-clinic FK と Owner/Pet 不一致だけは、一覧/単件のどちらでも親行ごと fail-closed にする。
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
		AND (
			appointments.created_by IS NULL
			OR EXISTS (
				SELECT 1
				FROM staffs scoped_creator
				JOIN staff_clinic_assignments scoped_creator_assignment
				  ON scoped_creator_assignment.staff_id = scoped_creator.id
				 AND scoped_creator_assignment.clinic_id = appointments.clinic_id
				WHERE scoped_creator.id = appointments.created_by
			)
		)
	`

func reservationRelationsMatchParentClinic(q *gorm.DB) *gorm.DB {
	return q.Where(reservationRelationsMatchParentClinicSQL)
}
