package reservation

// timeslot_availability.go — EMR-170: 院内予約フォーム向け 〇△✕ 空き状況評価。
// GenerateTimeSlots（空き枠のみ列挙）を拡張し、全対象スタッフが埋まった満員枠も含めて
// 各枠の残り受入数（remaining）と3段階ステータスを算出する。

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SlotVacancyStatus は1予約枠の空き状況を表す（EMR-170 〇△✕表示）。
type SlotVacancyStatus string

const (
	// SlotVacancyAvailable は 〇（空きあり）。
	SlotVacancyAvailable SlotVacancyStatus = "available"
	// SlotVacancyLow は △（残りわずか — remaining が lowVacancyRemaining 以下）。
	SlotVacancyLow SlotVacancyStatus = "low"
	// SlotVacancyFull は ✕（満員 — remaining が 0）。
	SlotVacancyFull SlotVacancyStatus = "full"
)

// lowVacancyRemaining は「残りわずか」（△）とみなす残り枠数の上限。
// EMR-170 spec 決定: 残り1枠のみを△とする（この枠を取ると埋まることを示す）。
const lowVacancyRemaining = 1

// SlotCapacity は1予約枠と、それを受け付けられる対象スタッフ数を表す。
// GenerateTimeSlots が「1名以上空きのある枠」のみ返すのに対し、GenerateSlotCapacities は
// 全対象スタッフが埋まった満員枠も FreeStaffCount=0 で含める。
type SlotCapacity struct {
	TimeSlot
	// FreeStaffCount はこの枠を受け付けられる対象スタッフの数。
	FreeStaffCount int
	// StaffBound はスタッフ勤務グリッド由来の枠か。予約可能枠テーブル（加算モード）由来の
	// 追加枠はスタッフ制約を受けないため false を使う（mergeSlotCapacities で設定）。
	StaffBound bool
}

// TimeSlotAvailability は1予約枠の空き状況評価結果（EMR-170）。
type TimeSlotAvailability struct {
	TimeSlot
	Status SlotVacancyStatus
	// Remaining は受け入れ可能な残り枠数 = min(空きスタッフ数, 区分定員残)。
	// スタッフ制約も区分定員もない枠は上限なしのため nil。
	Remaining *int
}

// GenerateSlotCapacities は対象スタッフ全員分の勤務内枠を対象集合として列挙し、
// 各枠を受け付けられるスタッフ数（FreeStaffCount）を付けて返す。
// GenerateTimeSlots は「1名以上空きのある枠」の和集合を返すのに対し、本関数は
// 全員が埋まった枠も FreeStaffCount=0 として含める。
// minimize_gaps モードでは予約占有による区間分割で free 側だけに現れる開始時刻が
// 生じうるため、free 側の枠も対象集合に含める（その枠は必ず FreeStaffCount>=1）。
func GenerateSlotCapacities(input *TimeSlotsInput) ([]SlotCapacity, error) {
	candidates := make(map[string]TimeSlot)
	freeSets := make([]map[string]struct{}, 0, len(input.Staffs))
	for i := range input.Staffs {
		potential, free, err := staffWorkAndFreeIntervals(input, &input.Staffs[i])
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to generate time slots")
		}
		potSlots, err := generateSlotsFromIntervals(input, potential)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to generate time slots")
		}
		for _, s := range potSlots {
			candidates[s.StartTime] = s
		}
		freeSlots, err := generateSlotsFromIntervals(input, free)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to generate time slots")
		}
		freeSet := make(map[string]struct{}, len(freeSlots))
		for _, s := range freeSlots {
			freeSet[s.StartTime] = struct{}{}
			if _, ok := candidates[s.StartTime]; !ok {
				candidates[s.StartTime] = s
			}
		}
		freeSets = append(freeSets, freeSet)
	}

	result := make([]SlotCapacity, 0, len(candidates))
	for start, slot := range candidates {
		freeCount := 0
		for _, fs := range freeSets {
			if _, ok := fs[start]; ok {
				freeCount++
			}
		}
		result = append(result, SlotCapacity{TimeSlot: slot, FreeStaffCount: freeCount, StaffBound: true})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime < result[j].StartTime
	})
	return result, nil
}

// rateSlotCapacities は予約可能枠テーブルの加算と区分定員（max_concurrent）の残数を
// 反映して各枠の空き状況を確定する。mergeAndFilterGeneratedSlots と同じ適用順序を保持する。
func (s *liffService) rateSlotCapacities(
	ctx context.Context,
	clinicID, typeID uint64,
	date time.Time,
	course *model.ReservationType,
	caps []SlotCapacity,
) ([]TimeSlotAvailability, error) {
	if s.availableSlotRepo != nil {
		availableSlots, err := s.availableSlotRepo.FindAll(ctx, clinicID, typeID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get available slots")
		}
		if HasActiveAvailableSlots(availableSlots) {
			caps = mergeSlotCapacities(caps, availableSlots, date, course.DurationMinutes)
		}
	}

	var taken map[string]int64
	if course.MaxConcurrent != nil && len(caps) > 0 {
		slotTimes := make([]TimeSlot, len(caps))
		for i := range caps {
			slotTimes[i] = caps[i].TimeSlot
		}
		counts, err := countReservationsAtSlotStarts(ctx, slotTimes, s.reservationRepo, clinicID, typeID, date)
		if err != nil {
			return nil, err
		}
		taken = counts
	}

	result := make([]TimeSlotAvailability, len(caps))
	for i := range caps {
		result[i] = rateSlotCapacity(caps[i], course.MaxConcurrent, taken)
	}
	return result, nil
}

// mergeSlotCapacities は予約可能枠テーブルの時刻をスタッフグリッド生成枠へ加算する。
// MergeAvailableTimeSlots と同じ加算モードの挙動を保持する:
//   - グリッドに無い時刻はスタッフ制約なしの枠（StaffBound=false）として追加する
//   - グリッドに存在するが全スタッフが埋まった枠（FreeStaffCount=0）は、従来の空き枠一覧
//     でも加算枠として復活していたため StaffBound=false に置き換える
//   - グリッドで空きがある枠は従来どおり重複追加しない
func mergeSlotCapacities(
	caps []SlotCapacity,
	availableSlots []model.ReservationTypeAvailableSlot,
	date time.Time,
	durationMinutes int,
) []SlotCapacity {
	applicableSlots := FilterApplicableAvailableSlots(availableSlots, date)
	if len(applicableSlots) == 0 {
		return caps
	}
	indexByStart := make(map[string]int, len(caps))
	for i := range caps {
		indexByStart[caps[i].StartTime] = i
	}
	result := make([]SlotCapacity, len(caps), len(caps)+len(applicableSlots))
	copy(result, caps)
	for i := range applicableSlots {
		startNorm := strings.ReplaceAll(applicableSlots[i].StartTime, ":", "")
		if idx, exists := indexByStart[startNorm]; exists {
			if result[idx].StaffBound && result[idx].FreeStaffCount == 0 {
				result[idx].StaffBound = false
			}
			continue
		}
		startMin, err := MinutesSinceMidnight(startNorm)
		if err != nil {
			continue
		}
		indexByStart[startNorm] = len(result)
		result = append(result, SlotCapacity{
			TimeSlot:   TimeSlot{StartTime: startNorm, EndTime: minutesToHHMM(startMin + durationMinutes)},
			StaffBound: false,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime < result[j].StartTime
	})
	return result
}

// rateSlotCapacity は1枠の残り受入数と空き状況を確定する。
// remaining = min(空きスタッフ数, 区分定員残)。どちらの制約も無い枠は Remaining=nil（上限なし）
// で常に 〇（available）。max_concurrent を超過した枠は残数を0に丸める。
func rateSlotCapacity(slotCap SlotCapacity, maxConcurrent *int, taken map[string]int64) TimeSlotAvailability {
	remaining := -1 // -1 = 上限なし
	if slotCap.StaffBound {
		remaining = slotCap.FreeStaffCount
	}
	if maxConcurrent != nil {
		typeRemaining := *maxConcurrent - int(taken[slotCap.StartTime])
		if typeRemaining < 0 {
			typeRemaining = 0
		}
		if remaining < 0 || typeRemaining < remaining {
			remaining = typeRemaining
		}
	}

	status := SlotVacancyAvailable
	var remainingPtr *int
	if remaining >= 0 {
		r := remaining
		remainingPtr = &r
		switch {
		case remaining == 0:
			status = SlotVacancyFull
		case remaining <= lowVacancyRemaining:
			status = SlotVacancyLow
		}
	}
	return TimeSlotAvailability{TimeSlot: slotCap.TimeSlot, Status: status, Remaining: remainingPtr}
}
