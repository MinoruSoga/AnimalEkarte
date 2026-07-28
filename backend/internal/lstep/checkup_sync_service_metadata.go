package lstep

import "time"

// buildCheckupSyncPreviewMetadata は PreviewCheckupSync の監査メタデータを生成する（ISSUE-010）。
// 受入条件の必須キー（total_count / eligible_count / line_linked_count / opt_out_count /
// no_living_pet_count / species / last_visit_before / last_visit_after）に加え、
// ISSUE-009 で追加された拡張フィルタ条件も含めて永続化する。
// 日付は YYYY-MM-DD 文字列に正規化する（タイムゾーンによる比較揺れを防ぐ）。
func buildCheckupSyncPreviewMetadata(
	input *PreviewCheckupSyncInput,
	totalCount, eligibleCount, lineLinkedCount, optOutCount, noLivingPetCount int,
) map[string]any {
	filter := map[string]any{
		"checkup_type": input.CheckupType,
		"species":      input.Species,
	}
	if input.LastVisitBefore != nil {
		filter["last_visit_before"] = input.LastVisitBefore.Format(time.DateOnly)
	}
	if input.LastVisitAfter != nil {
		filter["last_visit_after"] = input.LastVisitAfter.Format(time.DateOnly)
	}
	if input.MinAgeYears != nil {
		filter["min_age_years"] = *input.MinAgeYears
	}
	if input.MaxAgeYears != nil {
		filter["max_age_years"] = *input.MaxAgeYears
	}
	if input.HasChronicCondition != nil {
		filter["has_chronic_condition"] = *input.HasChronicCondition
	}
	if input.CPMStage != "" {
		filter["cpm_stage"] = input.CPMStage
	}
	if input.MinTotalAmount != nil {
		filter["min_total_amount"] = *input.MinTotalAmount
	}
	if input.MinAnnualVisitCount != nil {
		filter["min_annual_visit_count"] = *input.MinAnnualVisitCount
	}
	if input.LastCheckupBefore != nil {
		filter["last_checkup_before"] = input.LastCheckupBefore.Format(time.DateOnly)
	}
	if input.LastCheckupAfter != nil {
		filter["last_checkup_after"] = input.LastCheckupAfter.Format(time.DateOnly)
	}
	return map[string]any{
		"operation":           "checkup_sync_preview",
		"filter":              filter,
		"total_count":         totalCount,
		"eligible_count":      eligibleCount,
		"line_linked_count":   lineLinkedCount,
		"opt_out_count":       optOutCount,
		"no_living_pet_count": noLivingPetCount,
	}
}
