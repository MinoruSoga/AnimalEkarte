package lstep

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// parseOptionalDateOnly は空文字なら (nil, nil) を、YYYY-MM-DD 形式でなければ
// apperrors.WrapInvalidInput を返す（E-1: 4箇所の日付パースブロックの共通化）。
func parseOptionalDateOnly(s, name string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation(time.DateOnly, s, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput(name + " は YYYY-MM-DD 形式で指定してください")
	}
	return &t, nil
}

// parseOptionalNonNegInt は空文字なら (nil, nil) を、0以上の整数でなければ
// apperrors.WrapInvalidInput を返す（E-1: 年齢フィルタ用の int 版）。
func parseOptionalNonNegInt(s, name string) (*int, error) {
	if s == "" {
		return nil, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return nil, apperrors.WrapInvalidInput(name + " は 0 以上の整数で指定してください")
	}
	return &v, nil
}

// parseOptionalNonNegInt64 は空文字なら (nil, nil) を、0以上の整数でなければ
// apperrors.WrapInvalidInput を返す（E-1: 金額・回数フィルタ用の int64 版）。
func parseOptionalNonNegInt64(s, name string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil || v < 0 {
		return nil, apperrors.WrapInvalidInput(name + " は 0 以上の整数で指定してください")
	}
	return &v, nil
}

type checkupSyncPreviewQuery struct {
	CheckupType         string
	Species             string
	CPMStage            string
	LastVisitBefore     string
	LastVisitAfter      string
	MinAgeYears         string
	MaxAgeYears         string
	HasChronicCondition string
	MinTotalAmount      string
	MinAnnualVisitCount string
	LastCheckupBefore   string
	LastCheckupAfter    string
}

func newCheckupSyncPreviewQuery(values url.Values) checkupSyncPreviewQuery {
	return checkupSyncPreviewQuery{
		CheckupType:         values.Get("checkup_type"),
		Species:             values.Get("species"),
		CPMStage:            values.Get("cpm_stage"),
		LastVisitBefore:     values.Get("last_visit_before"),
		LastVisitAfter:      values.Get("last_visit_after"),
		MinAgeYears:         values.Get("min_age_years"),
		MaxAgeYears:         values.Get("max_age_years"),
		HasChronicCondition: values.Get("has_chronic_condition"),
		MinTotalAmount:      values.Get("min_total_amount"),
		MinAnnualVisitCount: values.Get("min_annual_visit_count"),
		LastCheckupBefore:   values.Get("last_checkup_before"),
		LastCheckupAfter:    values.Get("last_checkup_after"),
	}
}

func (q *checkupSyncPreviewQuery) toServiceInput() (*PreviewCheckupSyncInput, error) {
	switch q.CheckupType {
	case "annual", "dental", "blood", "skin", "cancer", "other":
	default:
		return nil, apperrors.WrapInvalidInput("checkup_type は annual/dental/blood/skin/cancer/other のいずれかを指定してください")
	}

	input := &PreviewCheckupSyncInput{
		CheckupType: q.CheckupType,
		Species:     q.Species,
		CPMStage:    q.CPMStage,
	}

	var err error
	if input.LastVisitBefore, err = parseOptionalDateOnly(q.LastVisitBefore, "last_visit_before"); err != nil {
		return nil, err
	}
	if input.LastVisitAfter, err = parseOptionalDateOnly(q.LastVisitAfter, "last_visit_after"); err != nil {
		return nil, err
	}
	if input.MinAgeYears, err = parseOptionalNonNegInt(q.MinAgeYears, "min_age_years"); err != nil {
		return nil, err
	}
	if input.MaxAgeYears, err = parseOptionalNonNegInt(q.MaxAgeYears, "max_age_years"); err != nil {
		return nil, err
	}
	if input.MinAgeYears != nil && input.MaxAgeYears != nil && *input.MinAgeYears > *input.MaxAgeYears {
		return nil, apperrors.WrapInvalidInput("min_age_years は max_age_years 以下で指定してください")
	}
	if q.HasChronicCondition != "" {
		switch q.HasChronicCondition {
		case "true":
			t := true
			input.HasChronicCondition = &t
		case "false":
			f := false
			input.HasChronicCondition = &f
		default:
			return nil, apperrors.WrapInvalidInput("has_chronic_condition は true/false で指定してください")
		}
	}
	if input.MinTotalAmount, err = parseOptionalNonNegInt64(q.MinTotalAmount, "min_total_amount"); err != nil {
		return nil, err
	}
	if input.MinAnnualVisitCount, err = parseOptionalNonNegInt64(q.MinAnnualVisitCount, "min_annual_visit_count"); err != nil {
		return nil, err
	}
	if input.LastCheckupBefore, err = parseOptionalDateOnly(q.LastCheckupBefore, "last_checkup_before"); err != nil {
		return nil, err
	}
	if input.LastCheckupAfter, err = parseOptionalDateOnly(q.LastCheckupAfter, "last_checkup_after"); err != nil {
		return nil, err
	}
	if input.CPMStage != "" {
		switch input.CPMStage {
		case string(CPMStageEncounter),
			string(CPMStageGrowing),
			string(CPMStageCore),
			string(CPMStageSpot),
			string(CPMStageNoah),
			string(CPMStageDormant):
		default:
			return nil, apperrors.WrapInvalidInput("cpm_stage は cpm_encounter/cpm_growing/cpm_core/cpm_spot/cpm_noah/cpm_dormant のいずれかを指定してください")
		}
	}

	return input, nil
}

// checkupSyncRequest は一括タグ付与リクエスト。
type checkupSyncRequest struct {
	CheckupType string   `json:"checkup_type" binding:"required"`
	OwnerIDs    []string `json:"owner_ids"    binding:"required,min=1,max=100"`
	TagName     string   `json:"tag_name"     binding:"required"`
}

func (r checkupSyncRequest) toServiceInput() (CreateCheckupSyncInput, error) {
	if !IsValidManualTagName(r.TagName) {
		return CreateCheckupSyncInput{}, fmt.Errorf("tag_name は英数字・アンダースコア・ハイフンのみ使用可能です（1〜100文字）")
	}

	ownerIDs := make([]uint64, 0, len(r.OwnerIDs))
	for _, s := range r.OwnerIDs {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return CreateCheckupSyncInput{}, fmt.Errorf("owner_ids の値が不正です: %s", s)
		}
		ownerIDs = append(ownerIDs, id)
	}

	return CreateCheckupSyncInput{
		CheckupType: r.CheckupType,
		OwnerIDs:    ownerIDs,
		TagName:     r.TagName,
	}, nil
}
