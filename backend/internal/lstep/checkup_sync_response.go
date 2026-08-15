package lstep

import (
	"strconv"
	"time"
)

// checkupSyncPreviewOwnerResponse はプレビュー一覧の1件レスポンス（ISSUE-005: 除外理由対応 / ISSUE-009: 追加フィルタ表示）。
type checkupSyncPreviewOwnerResponse struct {
	OwnerID         string   `json:"owner_id"`
	OwnerName       string   `json:"owner_name"`
	PetNames        []string `json:"pet_names"`
	LastVisitDate   *string  `json:"last_visit_date"`
	HasLine         bool     `json:"has_line"`
	IsOptOut        bool     `json:"is_opt_out"`
	HasLivingPet    bool     `json:"has_living_pet"`
	ExclusionReason *string  `json:"exclusion_reason"`
	CurrentTags     []string `json:"current_tags"`

	// ISSUE-009: フィルタ確認用のメタ情報（additive）
	MinPetAgeYears      *int    `json:"min_pet_age_years"`
	MaxPetAgeYears      *int    `json:"max_pet_age_years"`
	HasChronicCondition bool    `json:"has_chronic_condition"`
	CPMStage            string  `json:"cpm_stage"`
	TotalAmount         int64   `json:"total_amount"`
	AnnualVisitCount    int64   `json:"annual_visit_count"`
	LastCheckupDate     *string `json:"last_checkup_date"`
}

// checkupSyncPreviewResponse はプレビューレスポンス（ISSUE-005: 除外サマリー対応）。
type checkupSyncPreviewResponse struct {
	Owners           []checkupSyncPreviewOwnerResponse `json:"owners"`
	TotalCount       int                               `json:"total_count"`
	EligibleCount    int                               `json:"eligible_count"`
	LineLinkedCount  int                               `json:"line_linked_count"`
	OptOutCount      int                               `json:"opt_out_count"`
	NoLivingPetCount int                               `json:"no_living_pet_count"`
}

// checkupSyncResultResponse は一括タグ付与レスポンス。
type checkupSyncResultResponse struct {
	SuccessCount   int      `json:"success_count"`
	SkippedCount   int      `json:"skipped_count"`
	FailedCount    int      `json:"failed_count"`
	FailedOwnerIDs []string `json:"failed_owner_ids"`
}

func toCheckupSyncPreviewOwnerResponse(o *CheckupSyncPreviewOwner) checkupSyncPreviewOwnerResponse {
	r := checkupSyncPreviewOwnerResponse{
		OwnerID:             strconv.FormatUint(o.OwnerID, 10),
		OwnerName:           o.OwnerName,
		PetNames:            o.PetNames,
		HasLine:             o.HasLine,
		IsOptOut:            o.IsOptOut,
		HasLivingPet:        o.HasLivingPet,
		ExclusionReason:     o.ExclusionReason,
		CurrentTags:         o.CurrentTags,
		MinPetAgeYears:      o.MinPetAgeYears,
		MaxPetAgeYears:      o.MaxPetAgeYears,
		HasChronicCondition: o.HasChronicCondition,
		CPMStage:            o.CPMStage,
		TotalAmount:         o.TotalAmount,
		AnnualVisitCount:    o.AnnualVisitCount,
	}
	if o.LastVisitDate != nil {
		s := o.LastVisitDate.In(time.Local).Format(time.DateOnly)
		r.LastVisitDate = &s
	}
	if o.LastCheckupDate != nil {
		s := o.LastCheckupDate.In(time.Local).Format(time.DateOnly)
		r.LastCheckupDate = &s
	}
	return r
}

func toCheckupSyncPreviewResponse(result *PreviewCheckupSyncResult) checkupSyncPreviewResponse {
	owners := make([]checkupSyncPreviewOwnerResponse, 0, len(result.Owners))
	for i := range result.Owners {
		owners = append(owners, toCheckupSyncPreviewOwnerResponse(&result.Owners[i]))
	}
	return checkupSyncPreviewResponse{
		Owners:           owners,
		TotalCount:       result.TotalCount,
		EligibleCount:    result.EligibleCount,
		LineLinkedCount:  result.LineLinkedCount,
		OptOutCount:      result.OptOutCount,
		NoLivingPetCount: result.NoLivingPetCount,
	}
}
