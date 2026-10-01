package medicalrecord

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/textsearch"
)

// applyMedicalRecordSearch はカルテ横断検索の WHERE を組み立てる。
// 検索対象はカルテ番号・飼主名/カナ・ペット名/カナ・主訴の4条件
// （EMR-244: 治療内容・メモ・治療マスタ名の腕は除外）。
//
// 構造は「一致ID集合を先に索引で計算 → (id, clinic_id) のペア IN で半結合」。
// 各 UNION 腕は自分のテーブルの GIN trigram / FK btree インデックスで駆動されるため、
// medical_records 全行の逐行評価は発生しない（2026-10-01 障害: JOIN+OR の
// 複数テーブル横断 OR はインデックスを殺し、150万行規模を毎検索で走査していた）。
//
// ペア IN は「id 一致 + 同一 clinic 一致」を同時に要求するため、従来の JOIN ON 条件
// （id 相関 + clinic_id 相関 + deleted_at IS NULL）と同じ不変条件を保持する。
// 破損した外部 clinic FK を検索語が復元しない（BUG-454 系）。
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
			` WHERE `+textsearch.FoldedExpr("searched_inquiry.chief_complaint")+` ILIKE ? ESCAPE '\')`,
		pattern, pattern, pattern, pattern, pattern, pattern,
	)
}
