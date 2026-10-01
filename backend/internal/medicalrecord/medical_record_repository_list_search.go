package medicalrecord

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/textsearch"
)

// applyMedicalRecordSearch はカルテ横断検索の WHERE を組み立てる。
//
// 構造は「一致ID集合を先に索引で計算 → (id, clinic_id) のペア IN で半結合」。
// 各 UNION 腕は自分のテーブルの GIN trigram / FK btree インデックスで駆動されるため、
// medical_records 全行の逐行評価は発生しない（2026-10-01 障害: JOIN+OR の
// 複数テーブル横断 OR はインデックスを殺し、150万行規模を毎検索で走査していた）。
//
// ペア IN は「id 一致 + 同一 clinic 一致」を同時に要求するため、従来の JOIN ON 条件
// （id 相関 + clinic_id 相関 + deleted_at IS NULL）と同じ不変条件を保持する。
// 破損した外部 clinic FK を検索語が復元しない（BUG-454 系）。
// treatments のマスタ腕は (procedure_id, mr_t.clinic_id) ペア IN で、
// 「マスタの clinic = カルテの clinic」の旧条件を保持する（treatment 自体の
// clinic は旧 EXISTS でも未拘束だったため同様に拘束しない）。
// inquiries は旧 JOIN が clinic_id/deleted_at を条件に持たなかったため、
// 腕が返す clinic は JOIN した medical_records 自身のもの（= 条件なしと同義）。
//
// 各列につき畳込み(カナ+空白) translate() 式への ILIKE を1腕だけ生成する。
// 式はリテラル埋め込みで、migrations/015_search_fold_indexes.sql の
// GIN trigram 式インデックスとテキスト一致する必要がある（FoldedExpr 参照）。
func applyMedicalRecordSearch(q *gorm.DB, search string) *gorm.DB {
	qSearch := textsearch.NormalizeQuerySpaces(search)
	if qSearch == "" {
		return q.Where("1 = 0")
	}
	pattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(qSearch)) + "%"
	return q.Where(
		`(medical_records.id, medical_records.clinic_id) IN (`+
			`SELECT mr_no.id, mr_no.clinic_id FROM medical_records mr_no`+
			` WHERE mr_no.deleted_at IS NULL`+
			` AND mr_no.record_no ILIKE ? ESCAPE '\'`+
			` UNION ALL `+
			`SELECT mr_o.id, mr_o.clinic_id FROM medical_records mr_o`+
			` JOIN owners searched_owner`+
			` ON searched_owner.id = mr_o.owner_id`+
			` AND searched_owner.clinic_id = mr_o.clinic_id`+
			` AND searched_owner.deleted_at IS NULL`+
			` WHERE mr_o.deleted_at IS NULL AND (`+
			textsearch.FoldedExpr("searched_owner.name")+` ILIKE ? ESCAPE '\'`+
			` OR `+textsearch.FoldedExpr("searched_owner.name_kana")+` ILIKE ? ESCAPE '\')`+
			` UNION ALL `+
			`SELECT mr_p.id, mr_p.clinic_id FROM medical_records mr_p`+
			` JOIN pets searched_pet`+
			` ON searched_pet.id = mr_p.pet_id`+
			` AND searched_pet.clinic_id = mr_p.clinic_id`+
			` AND searched_pet.deleted_at IS NULL`+
			` WHERE mr_p.deleted_at IS NULL AND (`+
			textsearch.FoldedExpr("searched_pet.name")+` ILIKE ? ESCAPE '\'`+
			` OR `+textsearch.FoldedExpr("searched_pet.name_kana")+` ILIKE ? ESCAPE '\')`+
			` UNION ALL `+
			`SELECT searched_inquiry.medical_record_id, mr_i.clinic_id`+
			` FROM inquiries searched_inquiry`+
			` JOIN medical_records mr_i ON mr_i.id = searched_inquiry.medical_record_id`+
			` WHERE `+textsearch.FoldedExpr("searched_inquiry.chief_complaint")+` ILIKE ? ESCAPE '\'`+
			` UNION ALL `+
			// treatments 腕は旧 EXISTS が treatment 自体に clinic 相関を
			// 課していなかった（medical_record_id 相関のみ、clinic 相関はマスタ側）。
			// mr_t を join してカルテ側の clinic を返すことで同じ条件を維持する。
			`SELECT searched_treatment.medical_record_id, mr_t.clinic_id`+
			` FROM treatments searched_treatment`+
			` JOIN medical_records mr_t ON mr_t.id = searched_treatment.medical_record_id`+
			` WHERE searched_treatment.deleted_at IS NULL AND (`+
			textsearch.FoldedExpr("searched_treatment.content")+` ILIKE ? ESCAPE '\'`+
			` OR `+textsearch.FoldedExpr("searched_treatment.memo")+` ILIKE ? ESCAPE '\'`+
			` OR (searched_treatment.procedure_id, mr_t.clinic_id) IN (`+
			`SELECT searched_procedure.id, searched_procedure.clinic_id FROM procedures searched_procedure`+
			` WHERE searched_procedure.deleted_at IS NULL`+
			` AND `+textsearch.FoldedExpr("searched_procedure.name")+` ILIKE ? ESCAPE '\')`+
			` OR (searched_treatment.medicine_id, mr_t.clinic_id) IN (`+
			`SELECT searched_medicine.id, searched_medicine.clinic_id FROM medicines searched_medicine`+
			` WHERE searched_medicine.deleted_at IS NULL`+
			` AND `+textsearch.FoldedExpr("searched_medicine.name")+` ILIKE ? ESCAPE '\')`+
			` OR (searched_treatment.consultation_id, mr_t.clinic_id) IN (`+
			`SELECT searched_consultation.id, searched_consultation.clinic_id FROM consultations searched_consultation`+
			` WHERE searched_consultation.deleted_at IS NULL`+
			` AND `+textsearch.FoldedExpr("searched_consultation.name")+` ILIKE ? ESCAPE '\')`+
			` OR (searched_treatment.inventory_id, mr_t.clinic_id) IN (`+
			`SELECT searched_inventory.id, searched_inventory.clinic_id FROM inventory_items searched_inventory`+
			` WHERE searched_inventory.deleted_at IS NULL`+
			` AND `+textsearch.FoldedExpr("searched_inventory.name")+` ILIKE ? ESCAPE '\')))`,
		pattern, pattern, pattern, pattern, pattern, pattern,
		pattern, pattern, pattern, pattern, pattern, pattern,
	)
}
