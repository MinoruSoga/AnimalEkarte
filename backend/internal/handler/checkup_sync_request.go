package handler

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/service"
)

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

func (q *checkupSyncPreviewQuery) toServiceInput() (*service.PreviewCheckupSyncInput, error) {
	input := &service.PreviewCheckupSyncInput{
		CheckupType: q.CheckupType,
		Species:     q.Species,
		CPMStage:    q.CPMStage,
	}

	if q.LastVisitBefore != "" {
		t, err := time.ParseInLocation(time.DateOnly, q.LastVisitBefore, time.Local)
		if err != nil {
			return nil, fmt.Errorf("last_visit_before は YYYY-MM-DD 形式で指定してください")
		}
		input.LastVisitBefore = &t
	}
	if q.LastVisitAfter != "" {
		t, err := time.ParseInLocation(time.DateOnly, q.LastVisitAfter, time.Local)
		if err != nil {
			return nil, fmt.Errorf("last_visit_after は YYYY-MM-DD 形式で指定してください")
		}
		input.LastVisitAfter = &t
	}
	if q.MinAgeYears != "" {
		v, err := strconv.Atoi(q.MinAgeYears)
		if err != nil || v < 0 {
			return nil, fmt.Errorf("min_age_years は 0 以上の整数で指定してください")
		}
		input.MinAgeYears = &v
	}
	if q.MaxAgeYears != "" {
		v, err := strconv.Atoi(q.MaxAgeYears)
		if err != nil || v < 0 {
			return nil, fmt.Errorf("max_age_years は 0 以上の整数で指定してください")
		}
		input.MaxAgeYears = &v
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
			return nil, fmt.Errorf("has_chronic_condition は true/false で指定してください")
		}
	}
	if q.MinTotalAmount != "" {
		v, err := strconv.ParseInt(q.MinTotalAmount, 10, 64)
		if err != nil || v < 0 {
			return nil, fmt.Errorf("min_total_amount は 0 以上の整数で指定してください")
		}
		input.MinTotalAmount = &v
	}
	if q.MinAnnualVisitCount != "" {
		v, err := strconv.ParseInt(q.MinAnnualVisitCount, 10, 64)
		if err != nil || v < 0 {
			return nil, fmt.Errorf("min_annual_visit_count は 0 以上の整数で指定してください")
		}
		input.MinAnnualVisitCount = &v
	}
	if q.LastCheckupBefore != "" {
		t, err := time.ParseInLocation(time.DateOnly, q.LastCheckupBefore, time.Local)
		if err != nil {
			return nil, fmt.Errorf("last_checkup_before は YYYY-MM-DD 形式で指定してください")
		}
		input.LastCheckupBefore = &t
	}
	if q.LastCheckupAfter != "" {
		t, err := time.ParseInLocation(time.DateOnly, q.LastCheckupAfter, time.Local)
		if err != nil {
			return nil, fmt.Errorf("last_checkup_after は YYYY-MM-DD 形式で指定してください")
		}
		input.LastCheckupAfter = &t
	}
	if input.CPMStage != "" {
		switch input.CPMStage {
		case string(service.CPMStageEncounter),
			string(service.CPMStageGrowing),
			string(service.CPMStageCore),
			string(service.CPMStageSpot),
			string(service.CPMStageNoah),
			string(service.CPMStageDormant):
		default:
			return nil, fmt.Errorf("cpm_stage は cpm_encounter/cpm_growing/cpm_core/cpm_spot/cpm_noah/cpm_dormant のいずれかを指定してください")
		}
	}

	return input, nil
}

func (r checkupSyncRequest) toServiceInput() (service.CreateCheckupSyncInput, error) {
	if !tagNamePattern.MatchString(r.TagName) {
		return service.CreateCheckupSyncInput{}, fmt.Errorf("tag_name は英数字・アンダースコア・ハイフンのみ使用可能です（1〜100文字）")
	}

	ownerIDs := make([]uint64, 0, len(r.OwnerIDs))
	for _, s := range r.OwnerIDs {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return service.CreateCheckupSyncInput{}, fmt.Errorf("owner_ids の値が不正です: %s", s)
		}
		ownerIDs = append(ownerIDs, id)
	}

	return service.CreateCheckupSyncInput{
		CheckupType: r.CheckupType,
		OwnerIDs:    ownerIDs,
		TagName:     r.TagName,
	}, nil
}
