package pet

// checkup_history.go — EMR-197-01: GET /v1/pets?checkup_history= のフィルタ実装。
//
// 列挙値はスプレッドシート要件「直近 N 年以内に健診を受けた犬を検索したい」に対応する:
//   within_1y|within_2y|within_3y    live 健診が JST 当日起点 N 年（包含境界）内に存在
//   not_within_1y|not_within_2y|not_within_3y … 同窓内の live 健診が無い（履歴なし含む）
//   none                             live 健診履歴が一切無い
// 未指定（空文字）はフィルタ無しとして一切の述語を張らず、列挙外の値は invalid input。

import (
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
)

// CheckupHistoryFilter は GET /v1/pets?checkup_history= の列挙値。
// 空文字は「フィルタ無し」（未指定）を表し、applyPetCheckupHistoryFilter が述語を張らない。
type CheckupHistoryFilter string

const (
	CheckupHistoryWithin1Y    CheckupHistoryFilter = "within_1y"
	CheckupHistoryWithin2Y    CheckupHistoryFilter = "within_2y"
	CheckupHistoryWithin3Y    CheckupHistoryFilter = "within_3y"
	CheckupHistoryNotWithin1Y CheckupHistoryFilter = "not_within_1y"
	CheckupHistoryNotWithin2Y CheckupHistoryFilter = "not_within_2y"
	CheckupHistoryNotWithin3Y CheckupHistoryFilter = "not_within_3y"
	CheckupHistoryNone        CheckupHistoryFilter = "none"
)

// parseCheckupHistoryFilter はクエリ値を request boundary で列挙値へ検証する。
// 空文字（未指定）はフィルタ無しとして受理し、列挙外の値は invalid input にする
// （未知の値を黙って無視すると 400 化の契約に反する）。
func parseCheckupHistoryFilter(value string) (CheckupHistoryFilter, error) {
	switch CheckupHistoryFilter(value) {
	case "":
		return "", nil
	case CheckupHistoryWithin1Y, CheckupHistoryWithin2Y, CheckupHistoryWithin3Y,
		CheckupHistoryNotWithin1Y, CheckupHistoryNotWithin2Y, CheckupHistoryNotWithin3Y,
		CheckupHistoryNone:
		return CheckupHistoryFilter(value), nil
	default:
		return "", apperrors.WrapInvalidInput("invalid checkup_history")
	}
}

// petCheckupHistoryCorrelation は pets 行に対する live 健診の存在を評価する
// EXISTS 相関述語（JOIN ではなく EXISTS を使い、複数健診を持つペット行の重複を防ぐ）。
// ペット解決は checkups.pet_id 直接参照か、live かつ clinic 一致の medical_records.pet_id
// 経由（medical_record fallback join）のいずれかで行う。
// clinic_id parity は c.clinic_id = pets.clinic_id と m.clinic_id = c.clinic_id の二重条件で維持し、
// soft-deleted 行は c/m 双方の deleted_at IS NULL で除外する。
const petCheckupHistoryCorrelation = `EXISTS (
		SELECT 1
		  FROM checkups c
		  LEFT JOIN medical_records m
		    ON m.id = c.medical_record_id
		   AND m.deleted_at IS NULL
		   AND m.clinic_id = c.clinic_id
		 WHERE c.deleted_at IS NULL
		   AND c.clinic_id = pets.clinic_id
		   AND (c.pet_id = pets.id OR m.pet_id = pets.id)`

// applyPetCheckupHistoryFilter は checkup_history フィルタを count/list 両クエリへ
// 同一の述語で適用する（total とページ結果の一貫性）。窓は [JST 当日−N 年, JST 当日]
// の両端包含（「N 年以内に受診」は未来日付の受診予定を含まない）。下限・上限とも
// 暦日を CAST(? AS date) としてパラメータバインドする — 列挙値そのものはコード側で
// 分岐するため SQL へ値展開しない。
func applyPetCheckupHistoryFilter(q *gorm.DB, filter CheckupHistoryFilter, now time.Time) *gorm.DB {
	if filter == CheckupHistoryNone {
		// live 健診履歴が一切無いペット（窓を持たない否定形）。
		return q.Where("NOT " + petCheckupHistoryCorrelation + ")")
	}
	years, negative := checkupHistoryWindowYears(filter)
	if years == 0 {
		return q
	}
	lowerBound := checkupHistoryWindowStart(now, years).Format("2006-01-02")
	upperBound := now.In(config.JST).Format("2006-01-02")
	windowed := petCheckupHistoryCorrelation +
		"\n\t\t   AND c.date >= CAST(? AS date)" +
		"\n\t\t   AND c.date <= CAST(? AS date))"
	if negative {
		return q.Where("NOT "+windowed, lowerBound, upperBound)
	}
	return q.Where(windowed, lowerBound, upperBound)
}

// checkupHistoryWindowYears は within_*/not_within_* を (N 年, 否定フラグ) へ解決する。
// none / 空文字（フィルタ無し）は (0, false) を返す。
func checkupHistoryWindowYears(filter CheckupHistoryFilter) (years int, negative bool) {
	switch filter {
	case CheckupHistoryWithin1Y:
		return 1, false
	case CheckupHistoryWithin2Y:
		return 2, false
	case CheckupHistoryWithin3Y:
		return 3, false
	case CheckupHistoryNotWithin1Y:
		return 1, true
	case CheckupHistoryNotWithin2Y:
		return 2, true
	case CheckupHistoryNotWithin3Y:
		return 3, true
	}
	return 0, false
}

// checkupHistoryWindowStart は JST 当日から N 年前の暦日（包含下限）を返す。
// 「2年以内に受診」は当日起点ちょうど2年前の日を含む（境界包含）契約。
func checkupHistoryWindowStart(now time.Time, years int) time.Time {
	today := now.In(config.JST)
	return time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, config.JST).AddDate(-years, 0, 0)
}
