package lstep

// EMR-330: 非 tentative 配信トリガーのイベントソース適格性再読フック。
// candidate 抽出後に対象が失われていれば送信しない。全ファミリーで
// capability 欠落・再読エラーは fail-closed（送信せず error を返す）。
// tentative 経路は deliveryRevalidationHook（EMR-308）を使い、本ファイルの
// フックは使わない。

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// aliveFirstVisitEligibility は初診フォローアップ系の再読。
// 生存ペット由来カルテの初診日が targetDate 暦日のままかを owner 単位で確認する。
// 「他の生存ペットがいる」だけでは契約を満たさない —— 初診イベント由来の
// 生存ペット行が残っていることが必須。
func (s *lstepDeliveryTriggerService) aliveFirstVisitEligibility(clinicID uint64, targetDate time.Time) deliveryEligibilityFn {
	return func(ctx context.Context, ownerID uint64) (bool, string, error) {
		if s.medRecordRepo == nil {
			return false, "", apperrors.WrapInternalServerError("medical record repository is not configured")
		}
		ok, err := s.medRecordRepo.OwnerHasAliveFirstVisitOnDate(ctx, clinicID, ownerID, targetDate)
		if err != nil {
			return false, "", apperrors.Wrap(err, "failed to re-read alive first-visit source")
		}
		if !ok {
			return false, "stale_first_visit_source", nil
		}
		return true, "", nil
	}
}

// aliveLastVisitEligibility は休眠予防系の再読。
// 生存ペット由来カルテの最終来院日が targetDate 暦日のままかを owner 単位で確認する。
func (s *lstepDeliveryTriggerService) aliveLastVisitEligibility(clinicID uint64, targetDate time.Time) deliveryEligibilityFn {
	return func(ctx context.Context, ownerID uint64) (bool, string, error) {
		if s.medRecordRepo == nil {
			return false, "", apperrors.WrapInternalServerError("medical record repository is not configured")
		}
		ok, err := s.medRecordRepo.OwnerHasAliveLastVisitOnDate(ctx, clinicID, ownerID, targetDate)
		if err != nil {
			return false, "", apperrors.Wrap(err, "failed to re-read alive last-visit source")
		}
		if !ok {
			return false, "stale_last_visit_source", nil
		}
		return true, "", nil
	}
}

// vaccineDeadlineEligibility はワクチン期限系の再読。
// 生存ペットの next_date が targetDate のワクチンを飼い主が持つかを確認する。
func (s *lstepDeliveryTriggerService) vaccineDeadlineEligibility(clinicID uint64, targetDate time.Time) deliveryEligibilityFn {
	return func(ctx context.Context, ownerID uint64) (bool, string, error) {
		if s.vaccinationRepo == nil {
			return false, "", apperrors.WrapInternalServerError("vaccination repository is not configured")
		}
		ok, err := s.vaccinationRepo.OwnerHasAliveVaccineDeadline(ctx, clinicID, ownerID, targetDate)
		if err != nil {
			return false, "", apperrors.Wrap(err, "failed to re-read alive vaccine deadline source")
		}
		if !ok {
			return false, "stale_vaccine_deadline_source", nil
		}
		return true, "", nil
	}
}

// petBirthdayEligibility は誕生日メッセージ系の再読。
// 指定月日が誕生日の生存ペットを飼い主が持つかを確認する。
func (s *lstepDeliveryTriggerService) petBirthdayEligibility(clinicID uint64, month, day int) deliveryEligibilityFn {
	return func(ctx context.Context, ownerID uint64) (bool, string, error) {
		if s.petRepo == nil {
			return false, "", apperrors.WrapInternalServerError("pet repository is not configured")
		}
		ok, err := s.petRepo.OwnerHasAlivePetBirthday(ctx, clinicID, ownerID, month, day)
		if err != nil {
			return false, "", apperrors.Wrap(err, "failed to re-read alive birthday source")
		}
		if !ok {
			return false, "stale_birthday_source", nil
		}
		return true, "", nil
	}
}

// livingOwnerEligibility はタグ由来などペットイベント識別子を持たない
// トリガーの scoped living-owner 再検証。
// 全ペット死亡なら送信しない。owner レベルのタグを特定ペットへ帰属させない。
func (s *lstepDeliveryTriggerService) livingOwnerEligibility(clinicID uint64) deliveryEligibilityFn {
	return func(ctx context.Context, ownerID uint64) (bool, string, error) {
		if s.petRepo == nil {
			return false, "", apperrors.WrapInternalServerError("pet repository is not configured")
		}
		living, err := s.petRepo.CountLivingByOwner(ctx, clinicID, ownerID)
		if err != nil {
			return false, "", apperrors.Wrap(err, "failed to re-read living pet count")
		}
		if living == 0 {
			return false, "all_pets_deceased", nil
		}
		return true, "", nil
	}
}

// eventPetEligibility は初診ウェルカムのイベントペット再読。
// 起票イベントのペットが生存・当該飼い主所有・同医院であることを確認する。
// 履歴上の初診判定（CountByOwnerID）は呼び出し側の責務で、ここでは触れない。
func (s *lstepDeliveryTriggerService) eventPetEligibility(clinicID, ownerID, petID uint64) deliveryEligibilityFn {
	return func(ctx context.Context, _ uint64) (bool, string, error) {
		if s.petRepo == nil {
			return false, "", apperrors.WrapInternalServerError("pet repository is not configured")
		}
		pet, err := s.petRepo.FindByID(ctx, clinicID, petID)
		if err != nil {
			return false, "", apperrors.Wrap(err, "failed to re-read welcome event pet")
		}
		if pet == nil || pet.OwnerID != ownerID || sharedkernel.IsPetDeceased(pet) {
			return false, "event_pet_ineligible", nil
		}
		return true, "", nil
	}
}
