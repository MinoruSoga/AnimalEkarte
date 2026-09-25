package reservation

// availability_slot_merge.go — BE9-2C R③: liff_service_availability.go（R⑤）から
// timeslot engine 系 2 関数を前倒し移動（TimeSlot/MinutesSinceMidnight 等 engine 内部に依存する
// 実質 engine 関数のため。liff 側は reservation. 修飾で消費）。

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// MergeAvailableTimeSlots は営業時間から生成されたスロットに、
// 予約可能枠テーブルに登録された時刻を加算して返す（ホワイトリストではなく加算モード）。
// 既に営業時間に含まれている時刻は重複追加しない。
func MergeAvailableTimeSlots(slots []TimeSlot, availableSlots []model.ReservationTypeAvailableSlot, date time.Time, durationMinutes int) []TimeSlot {
	applicableSlots := FilterApplicableAvailableSlots(availableSlots, date)
	if len(applicableSlots) == 0 {
		return slots
	}
	existingStarts := make(map[string]struct{}, len(slots))
	for _, s := range slots {
		existingStarts[s.StartTime] = struct{}{}
	}
	result := make([]TimeSlot, len(slots), len(slots)+len(applicableSlots))
	copy(result, slots)
	for i := range applicableSlots {
		startNorm := strings.ReplaceAll(applicableSlots[i].StartTime, ":", "")
		if _, exists := existingStarts[startNorm]; exists {
			continue
		}
		startMin, err := MinutesSinceMidnight(startNorm)
		if err != nil {
			continue
		}
		result = append(result, TimeSlot{
			StartTime: startNorm,
			EndTime:   minutesToHHMM(startMin + durationMinutes),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime < result[j].StartTime
	})
	return result
}

// FilterSlotsByCapacity は max_concurrent に達したスロットを除外する。
// BE-refactor.md R2-4 (D8): repo が reservationTypeCapacityBatchCounter を実装していれば
// （本番の *reservationRepository は実装する）その日の全スロットの件数を1クエリで取得する。
// 実装していない場合（一部テストモック等）は従来どおり per-slot で問い合わせる（挙動保存）。
func FilterSlotsByCapacity(
	ctx context.Context,
	slots []TimeSlot,
	repo reservationTypeCapacityCounter,
	clinicID, typeID uint64,
	date time.Time,
	maxConcurrent int,
) ([]TimeSlot, error) {
	counts, err := countReservationsAtSlotStarts(ctx, slots, repo, clinicID, typeID, date)
	if err != nil {
		return nil, err
	}
	result := make([]TimeSlot, 0, len(slots))
	for _, slot := range slots {
		count, ok := counts[slot.StartTime]
		if !ok {
			continue // 不正な形式のスロットは結果から除外（既存挙動）
		}
		if count < int64(maxConcurrent) {
			result = append(result, slot)
		}
	}
	return result, nil
}

// countReservationsAtSlotStarts は各スロット開始時刻（"HHMM"）の当日予約数を返す。
// 戻り値のキーは slot.StartTime（不正な HHMM のスロットはキーに含めない）。
// repo が reservationTypeCapacityBatchCounter を実装していれば1クエリで取得し、
// 実装していない場合は per-slot で問い合わせる（FilterSlotsByCapacity 既存挙動の踏襲）。
// EMR-170 の空き状況評価（rateSlotCapacities）と同一カウント経路を共有するために抽出。
func countReservationsAtSlotStarts(
	ctx context.Context,
	slots []TimeSlot,
	repo reservationTypeCapacityCounter,
	clinicID, typeID uint64,
	date time.Time,
) (map[string]int64, error) {
	dateJST := date.In(config.JST)

	startByHHMM := make(map[string]time.Time, len(slots))
	for _, slot := range slots {
		startMin, err := MinutesSinceMidnight(slot.StartTime)
		if err != nil {
			continue // 不正な形式のスロットは結果から除外（既存挙動）
		}
		startByHHMM[slot.StartTime] = time.Date(
			dateJST.Year(), dateJST.Month(), dateJST.Day(),
			startMin/60, startMin%60, 0, 0, config.JST,
		)
	}

	counts := make(map[string]int64, len(startByHHMM))
	if batchRepo, ok := repo.(reservationTypeCapacityBatchCounter); ok {
		startTimes := make([]time.Time, 0, len(startByHHMM))
		for _, startTime := range startByHHMM {
			startTimes = append(startTimes, startTime)
		}
		byUnix, err := batchRepo.CountByTypeAndStartTimes(ctx, clinicID, typeID, startTimes, nil)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to count reservations")
		}
		for hhmm, startTime := range startByHHMM {
			counts[hhmm] = byUnix[startTime.Unix()]
		}
		return counts, nil
	}

	for hhmm, startTime := range startByHHMM {
		count, err := repo.CountByTypeAndStartTime(ctx, clinicID, typeID, startTime, nil)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to count reservations")
		}
		counts[hhmm] = count
	}
	return counts, nil
}
