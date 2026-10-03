package medicalrecord

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/textsearch"
)

// EMR-245: 表示列フィルタ（飼主名・ペット名・主訴）。Search の横断 OR 検索とは
// 独立した条件で、呼出側で AND 結合される。
//
// 構造は EXISTS の半結合: 各述語は対応テーブル側の folded 式インデックス
// （migrations/015_search_fold_indexes.sql の GIN trigram）で一致行を先に絞り、
// medical_records へ相関させる。JOIN+OR による複数テーブル横断 OR は
// インデックスを殺して全行逐行評価化するため使わない（2026-10-01 障害の教訓）。
//
// owners / pets は id + clinic_id 相関 + deleted_at IS NULL で、従来 JOIN /
// 検索腕と同じ不変条件を保持する（壊れた外部 clinic FK を検索語が復元しない）。
// inquiries は clinic_id / deleted_at 列を持たないため medical_record_id 相関のみ
// （clinic スコープは外側の medical_records.clinic_id IN が担保）—
// applyMedicalRecordSearch の主訴腕と同じ判断。
//
// パターンは textsearch.NormalizeQuerySpaces → NormalizeKana → EscapeLike で
// 組み立て、パラメータ化された ILIKE ? ESCAPE '\' にバインドする。

// medicalRecordColumnLikePattern は表示列フィルタ値を folded ILIKE パターンに
// 正規化する。空白のみなど正規化後に空になる入力は ok=false（呼出側は 1=0 で
// 「ヒットしない条件」にする: 空語を無条件へ捨てると一覧が全件露出し得る）。
func medicalRecordColumnLikePattern(value string) (pattern string, ok bool) {
	normalized := textsearch.NormalizeQuerySpaces(value)
	if normalized == "" {
		return "", false
	}
	return "%" + textsearch.EscapeLike(textsearch.NormalizeKana(normalized)) + "%", true
}

// applyMedicalRecordOwnerNameFilter は飼主名（name / name_kana）の部分一致条件。
func applyMedicalRecordOwnerNameFilter(q *gorm.DB, value string) *gorm.DB {
	pattern, ok := medicalRecordColumnLikePattern(value)
	if !ok {
		return q.Where("1 = 0")
	}
	return q.Where(
		`EXISTS (`+
			`SELECT 1 FROM owners filtered_owner`+
			` WHERE filtered_owner.id = medical_records.owner_id`+
			` AND filtered_owner.clinic_id = medical_records.clinic_id`+
			` AND filtered_owner.deleted_at IS NULL`+
			` AND (`+textsearch.FoldedExpr("filtered_owner.name")+` ILIKE ? ESCAPE '\'`+
			` OR `+textsearch.FoldedExpr("filtered_owner.name_kana")+` ILIKE ? ESCAPE '\'))`,
		pattern, pattern,
	)
}

// applyMedicalRecordPetNameFilter はペット名（name / name_kana）の部分一致条件。
func applyMedicalRecordPetNameFilter(q *gorm.DB, value string) *gorm.DB {
	pattern, ok := medicalRecordColumnLikePattern(value)
	if !ok {
		return q.Where("1 = 0")
	}
	return q.Where(
		`EXISTS (`+
			`SELECT 1 FROM pets filtered_pet`+
			` WHERE filtered_pet.id = medical_records.pet_id`+
			` AND filtered_pet.clinic_id = medical_records.clinic_id`+
			` AND filtered_pet.deleted_at IS NULL`+
			` AND (`+textsearch.FoldedExpr("filtered_pet.name")+` ILIKE ? ESCAPE '\'`+
			` OR `+textsearch.FoldedExpr("filtered_pet.name_kana")+` ILIKE ? ESCAPE '\'))`,
		pattern, pattern,
	)
}

// applyMedicalRecordChiefComplaintFilter は主訴（inquiries.chief_complaint）の
// 部分一致条件。
func applyMedicalRecordChiefComplaintFilter(q *gorm.DB, value string) *gorm.DB {
	pattern, ok := medicalRecordColumnLikePattern(value)
	if !ok {
		return q.Where("1 = 0")
	}
	return q.Where(
		`EXISTS (`+
			`SELECT 1 FROM inquiries filtered_inquiry`+
			` WHERE filtered_inquiry.medical_record_id = medical_records.id`+
			` AND `+textsearch.FoldedExpr("filtered_inquiry.chief_complaint")+` ILIKE ? ESCAPE '\')`,
		pattern,
	)
}
