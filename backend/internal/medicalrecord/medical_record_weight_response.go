package medicalrecord

// EMR-356: 問診画面/履歴の体重投影。
// 正本は引き続き vital_records。本投影は「このカルテに相関する体重測定」の
// 表示用スナップショットのみを運び、新しい体重 SoT・専用エンドポイントは作らない。
// wire 上は常に weight_measurements を出力する: []=「測定記録なし」、
// キー欠落=旧 API（投影未提供）として FE 側が区別できるようにする。

import (
	"sort"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// WeightMeasurementResponse は vital_records の体重測定を表す表示投影。
// id/value/unit/recorded_at のみ（患者本体の属性や他バイタルは含めない）。
type WeightMeasurementResponse struct {
	ID         uint64                `json:"id"`
	Value      model.ExactNumeric    `json:"value"`
	Unit       *model.BodyWeightUnit `json:"unit"`
	RecordedAt time.Time             `json:"recorded_at"`
}

// projectWeightMeasurements は親カルテの Vitals から体重投影を組み立てる。
// preload 側の相関に加えて、clinic/pet/record の完全一致・weight 非 NULL・
// daily_record（入院/日次）非由来・非削除を防衛的に再確認する。
// 返り値は常に非 nil（測定なしは [] として wire に出す）。
func projectWeightMeasurements(r *model.MedicalRecord) []WeightMeasurementResponse {
	measurements := make([]WeightMeasurementResponse, 0, len(r.Vitals))
	for _, v := range r.Vitals {
		if v.Weight == nil || v.DeletedAt.Valid {
			continue
		}
		if v.DailyRecordID != nil || v.MedicalRecordID == nil || *v.MedicalRecordID != r.ID {
			continue
		}
		if v.ClinicID != r.ClinicID || r.PetID == nil || v.PetID != *r.PetID {
			continue
		}
		measurements = append(measurements, WeightMeasurementResponse{
			ID:         v.ID,
			Value:      *v.Weight,
			Unit:       v.WeightUnit,
			RecordedAt: httpapi.LocalTime(v.RecordedAt),
		})
	}
	// stable: 測定日時 → id 昇順（同時刻測定の並びを決定的にする）
	sort.SliceStable(measurements, func(i, j int) bool {
		if !measurements[i].RecordedAt.Equal(measurements[j].RecordedAt) {
			return measurements[i].RecordedAt.Before(measurements[j].RecordedAt)
		}
		return measurements[i].ID < measurements[j].ID
	})
	return measurements
}
