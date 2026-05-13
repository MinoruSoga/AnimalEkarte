package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// extractTagCodes は mappings から code_type が一致するコード一覧を返す（FEAT-379）。
func extractTagCodes(mappings []*model.LstepTagCodeMapping, codeType string) []string {
	var codes []string
	for _, m := range mappings {
		if m.CodeType == codeType {
			codes = append(codes, []string(m.Codes)...)
		}
	}
	return codes
}

// strSet はスライスを O(1) 検索用マップに変換する。
func strSet(ss []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

// hasVaccineDeadlineSoon は vaccinations の中に NextDate が now から days 日以内のものがあれば true を返す。
// pure function — モックなしでテスト可能。
func hasVaccineDeadlineSoon(vaccinations []model.Vaccination, now time.Time, days int) bool { //nolint:unparam // days is configurable per clinic in future
	deadline := now.AddDate(0, 0, days)
	for i := range vaccinations {
		nd := vaccinations[i].NextDate
		if nd == nil {
			continue
		}
		// now <= NextDate <= deadline
		if !nd.Before(now) && !nd.After(deadline) {
			return true
		}
	}
	return false
}

// SyncHealthcheckTags は健診履歴に基づき HLTH_健診あり / HLTH_健診未受診 を同期する（FEAT-379）。
func (s *lstepTagSyncService) SyncHealthcheckTags(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, HlthHealthcheckDoneTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for healthcheck", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	checkupCodes := extractTagCodes(mappings, model.CodeTypeCheckupType)
	if len(checkupCodes) == 0 {
		return nil
	}

	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for healthcheck tags", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	since := time.Now().AddDate(0, 0, -HealthPreventionLookbackDays)
	checkups, err := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkups for healthcheck tags", "error", err)
		return apperrors.Wrap(err, "failed to find checkups")
	}

	codeSet := strSet(checkupCodes)
	hasHealthcheck := false
	var lastCheckupDate time.Time
	for i := range checkups {
		if checkups[i].Date.Before(since) {
			continue
		}
		if checkups[i].CheckupType == nil {
			continue
		}
		if _, ok := codeSet[checkups[i].CheckupType.Name]; ok {
			hasHealthcheck = true
			if checkups[i].Date.After(lastCheckupDate) {
				lastCheckupDate = checkups[i].Date
			}
		}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if hasHealthcheck {
		if addErr := client.AddTag(ctx, lineUserID, HlthHealthcheckDoneTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add healthcheck done tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add healthcheck done tag")
		}
		_ = s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, HlthHealthcheckDoneTag, "auto", fmt.Sprintf("最終健診: %s", lastCheckupDate.Format("2006-01-02")))
		if delErr := client.RemoveTag(ctx, lineUserID, HlthHealthcheckNeverTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove healthcheck never tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, HlthHealthcheckNeverTag)
		}
	} else {
		if addErr := client.AddTag(ctx, lineUserID, HlthHealthcheckNeverTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add healthcheck never tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add healthcheck never tag")
		}
		_ = s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, HlthHealthcheckNeverTag, "auto", "")
		if delErr := client.RemoveTag(ctx, lineUserID, HlthHealthcheckDoneTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove healthcheck done tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, HlthHealthcheckDoneTag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncAnnual4CheckupTag は年2回以上来院かつ健診履歴がある飼い主に HLTH_年4回候補 を付与する（FEAT-379）。
func (s *lstepTagSyncService) SyncAnnual4CheckupTag(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, HlthHealthcheckDoneTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for annual4checkup", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	checkupCodes := extractTagCodes(mappings, model.CodeTypeCheckupType)
	if len(checkupCodes) == 0 {
		return nil
	}

	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for annual4checkup tag", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	since := time.Now().AddDate(0, 0, -HealthPreventionLookbackDays)
	checkups, err := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkups for annual4checkup tag", "error", err)
		return apperrors.Wrap(err, "failed to find checkups")
	}

	codeSet := strSet(checkupCodes)
	hasHealthcheck := false
	for i := range checkups {
		if checkups[i].Date.Before(since) {
			continue
		}
		if checkups[i].CheckupType == nil {
			continue
		}
		if _, ok := codeSet[checkups[i].CheckupType.Name]; ok {
			hasHealthcheck = true
			break
		}
	}

	visitSummary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary for annual4checkup tag", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	qualified := hasHealthcheck && visitSummary.AnnualCount >= 2

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if qualified {
		if addErr := client.AddTag(ctx, lineUserID, HlthAnnual4CheckupTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add annual4checkup tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add annual4checkup tag")
		}
		_ = s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, HlthAnnual4CheckupTag, "auto", "")
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, HlthAnnual4CheckupTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove annual4checkup tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, HlthAnnual4CheckupTag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncVaccineDeadlineTag はワクチン次回予定日が VaccineDeadlineDays 以内なら
// PREV_ワクチン期限 を付与し、それ以外は解除する（FEAT-379）。
func (s *lstepTagSyncService) SyncVaccineDeadlineTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for vaccine deadline tag", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	vaccinations, err := s.vacRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find vaccinations for deadline tag", "error", err)
		return apperrors.Wrap(err, "failed to find vaccinations")
	}

	now := time.Now()
	deadline := now.AddDate(0, 0, VaccineDeadlineDays)
	deadlineSoon := false
	var earliestNextDate *time.Time
	for i := range vaccinations {
		nd := vaccinations[i].NextDate
		if nd == nil {
			continue
		}
		if !nd.Before(now) && !nd.After(deadline) {
			deadlineSoon = true
			if earliestNextDate == nil || nd.Before(*earliestNextDate) {
				copied := *nd
				earliestNextDate = &copied
			}
		}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if deadlineSoon {
		if addErr := client.AddTag(ctx, lineUserID, PrevVaccineDeadlineTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add vaccine deadline tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add vaccine deadline tag")
		}
		vaccineReason := ""
		if earliestNextDate != nil {
			vaccineReason = fmt.Sprintf("次回期限: %s", earliestNextDate.Format("2006-01-02"))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, PrevVaccineDeadlineTag, "auto", vaccineReason); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert vaccine deadline tag cache", "error", cacheErr)
		}
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, PrevVaccineDeadlineTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove vaccine deadline tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, PrevVaccineDeadlineTag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncFilariaTag はフィラリア検査・予防薬処方に基づき PREV_フィラリア未完了 を同期する（FEAT-379）。
// 犬を飼育する飼い主のみ対象。検査・処方どちらかが完了していればタグを解除する。
func (s *lstepTagSyncService) SyncFilariaTag(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, PrevFilariaTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for filaria tag", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	testCodes := extractTagCodes(mappings, model.CodeTypeCheckupType)
	rxCodes := extractTagCodes(mappings, model.CodeTypePrescription)
	if len(testCodes) == 0 && len(rxCodes) == 0 {
		return nil
	}

	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for filaria tag", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	// 犬を保有する飼い主のみ対象
	pets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find pets for filaria tag", "error", err)
		return apperrors.Wrap(err, "failed to find pets")
	}
	hasDog := false
	for i := range pets {
		if pets[i].AnimalSpecies != nil && strings.Contains(pets[i].AnimalSpecies.Name, "犬") {
			hasDog = true
			break
		}
	}
	if !hasDog {
		return nil
	}

	since := time.Now().AddDate(0, 0, -HealthPreventionLookbackDays)

	testDone := false
	if len(testCodes) > 0 {
		checkups, chkErr := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
		if chkErr != nil {
			slog.ErrorContext(ctx, "failed to find checkups for filaria tag", "error", chkErr)
			return apperrors.Wrap(chkErr, "failed to find checkups")
		}
		codeSet := strSet(testCodes)
		for i := range checkups {
			if checkups[i].Date.Before(since) {
				continue
			}
			if checkups[i].CheckupType == nil {
				continue
			}
			if _, ok := codeSet[checkups[i].CheckupType.Name]; ok {
				testDone = true
				break
			}
		}
	}

	rxDone := false
	if len(rxCodes) > 0 && s.billingItemRepo != nil {
		var rxErr error
		rxDone, rxErr = s.billingItemRepo.HasItemByOwnerSince(ctx, clinicID, ownerID, since, rxCodes)
		if rxErr != nil {
			slog.ErrorContext(ctx, "failed to check prescription for filaria tag", "error", rxErr)
			return apperrors.Wrap(rxErr, "failed to check prescription")
		}
	}

	// 検査・処方どちらか完了していれば「完了」→ タグ解除
	incomplete := !testDone && !rxDone

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if incomplete {
		if addErr := client.AddTag(ctx, lineUserID, PrevFilariaTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add filaria tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add filaria tag")
		}
		_ = s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, PrevFilariaTag, "auto", "未処方")
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, PrevFilariaTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove filaria tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, PrevFilariaTag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncFleaTickTag はノミ・マダニ駆除薬処方に基づき PREV_ノミダニ対象 を同期する（FEAT-379）。
// 処方実績がなければタグを付与し、あれば解除する。
func (s *lstepTagSyncService) SyncFleaTickTag(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil || s.billingItemRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, PrevFleaTickTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for flea tick tag", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	rxCodes := extractTagCodes(mappings, model.CodeTypePrescription)
	if len(rxCodes) == 0 {
		return nil
	}

	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for flea tick tag", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	since := time.Now().AddDate(0, 0, -HealthPreventionLookbackDays)
	hasRx, err := s.billingItemRepo.HasItemByOwnerSince(ctx, clinicID, ownerID, since, rxCodes)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check billing items for flea tick tag", "error", err)
		return apperrors.Wrap(err, "failed to check billing items")
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if !hasRx {
		if addErr := client.AddTag(ctx, lineUserID, PrevFleaTickTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add flea tick tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add flea tick tag")
		}
		_ = s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, PrevFleaTickTag, "auto", "未処方")
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, PrevFleaTickTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove flea tick tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, PrevFleaTickTag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncFoodPurchaseTag はフード購入履歴に基づき LTV_フード購入あり を同期する（FEAT-379）。
// codes が空の場合は category='food' でフォールバック検索する。
func (s *lstepTagSyncService) SyncFoodPurchaseTag(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil || s.billingItemRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, LtvFoodPurchaseTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for food purchase tag", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	// itemCodes が空でも HasFoodPurchaseByOwnerSince は category='food' にフォールバック
	itemCodes := extractTagCodes(mappings, model.CodeTypeMerchandiseItem)

	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for food purchase tag", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	since := time.Now().AddDate(0, 0, -HealthPreventionLookbackDays)
	hasPurchase, err := s.billingItemRepo.HasFoodPurchaseByOwnerSince(ctx, clinicID, ownerID, since, itemCodes)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check food purchase for tag", "error", err)
		return apperrors.Wrap(err, "failed to check food purchase")
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if hasPurchase {
		if addErr := client.AddTag(ctx, lineUserID, LtvFoodPurchaseTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add food purchase tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add food purchase tag")
		}
		_ = s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, LtvFoodPurchaseTag, "auto", "購入済")
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, LtvFoodPurchaseTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove food purchase tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, LtvFoodPurchaseTag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncSuppPurchaseTag はサプリ購入履歴に基づき LTV_サプリ購入あり を同期する（FEAT-385-supp）。
// SuppPurchaseCodes が空の場合は noop。category フォールバックなし。
func (s *lstepTagSyncService) SyncSuppPurchaseTag(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil || s.billingItemRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, LtvSuppPurchaseTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for supp purchase tag", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	itemCodes := extractTagCodes(mappings, model.CodeTypeMerchandiseItem)
	if len(itemCodes) == 0 {
		// カテゴリフォールバックなし — codes 未設定なら noop
		return nil
	}

	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for supp purchase tag", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	since := time.Now().AddDate(0, 0, -HealthPreventionLookbackDays)
	hasPurchase, err := s.billingItemRepo.HasSuppPurchaseByOwnerSince(ctx, clinicID, ownerID, since, itemCodes)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check supp purchase for tag", "error", err)
		return apperrors.Wrap(err, "failed to check supp purchase")
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if hasPurchase {
		if addErr := client.AddTag(ctx, lineUserID, LtvSuppPurchaseTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add supp purchase tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add supp purchase tag")
		}
		_ = s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, LtvSuppPurchaseTag, "auto", "購入済")
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, LtvSuppPurchaseTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove supp purchase tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, LtvSuppPurchaseTag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncHealthPreventionTagsForClinic は指定クリニックの全飼い主に対して
// 健診・予防・物販タグを一括同期する（FEAT-379 バッチエントリポイント）。
func (s *lstepTagSyncService) SyncHealthPreventionTagsForClinic(ctx context.Context, clinicID uint64) (int, []error) {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return 0, []error{apperrors.Wrap(err, "failed to check lstep sync enabled")}
	} else if skip {
		return 0, nil
	}

	owners, err := s.ownerRepo.FindAllWithLineUserID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "health-prevention batch: failed to find owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners with line user id")}
	}

	var errs []error
	count := 0
	for i := range owners {
		ownerID := owners[i].ID
		syncFns := []struct {
			name string
			fn   func() error
		}{
			{"SyncHealthcheckTags", func() error { return s.SyncHealthcheckTags(ctx, clinicID, ownerID) }},
			{"SyncAnnual4CheckupTag", func() error { return s.SyncAnnual4CheckupTag(ctx, clinicID, ownerID) }},
			{"SyncVaccineDeadlineTag", func() error { return s.SyncVaccineDeadlineTag(ctx, clinicID, ownerID) }},
			{"SyncFilariaTag", func() error { return s.SyncFilariaTag(ctx, clinicID, ownerID) }},
			{"SyncFleaTickTag", func() error { return s.SyncFleaTickTag(ctx, clinicID, ownerID) }},
			{"SyncFoodPurchaseTag", func() error { return s.SyncFoodPurchaseTag(ctx, clinicID, ownerID) }},
			{"SyncSuppPurchaseTag", func() error { return s.SyncSuppPurchaseTag(ctx, clinicID, ownerID) }},
		}
		ownerFailed := false
		for _, sf := range syncFns {
			if syncErr := sf.fn(); syncErr != nil {
				slog.ErrorContext(ctx, "health-prevention batch: sync failed",
					"clinic_id", clinicID, "owner_id", ownerID, "method", sf.name, "error", syncErr)
				errs = append(errs, apperrors.Wrap(syncErr, sf.name))
				ownerFailed = true
			}
		}
		if !ownerFailed {
			count++
		}
	}
	return count, errs
}
