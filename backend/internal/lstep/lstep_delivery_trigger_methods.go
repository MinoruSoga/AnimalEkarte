package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepDeliveryTriggerService) TriggerFirstVisitFollowUp3D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, -3)
	ownerIDs, err := s.medRecordRepo.FindOwnersByFirstVisitDate(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger first_visit_followup_3d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by first visit date")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFirstVisitFollowUp3D, model.TriggerTypeFirstVisitFollowUp3D, asOf, nil)
}

func (s *lstepDeliveryTriggerService) TriggerFirstVisitFollowUp7D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, -7)
	ownerIDs, err := s.medRecordRepo.FindOwnersByFirstVisitDate(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger first_visit_followup_7d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by first visit date")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFirstVisitFollowUp7D, model.TriggerTypeFirstVisitFollowUp7D, asOf, nil)
}

// TriggerNextVisitTentativeReminder3D は intended_date が asOf+3日（JST 暦日）の
// active tentative を持つ飼い主へ3日前リマインダーを配信する（EMR-308）。
func (s *lstepDeliveryTriggerService) TriggerNextVisitTentativeReminder3D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return s.runTentativeReminderBatch(ctx, clinicID, asOf, 3, model.TriggerTypeNextVisitTentativeReminder3d)
}

// TriggerNextVisitTentativeReminder0D は intended_date が asOf 当日（JST 暦日）の
// active tentative を持つ飼い主へ当日リマインダーを配信する（EMR-308）。
// 3日前ステージとは trigger_type が異なるため冪等・抑制は独立に機能する。
func (s *lstepDeliveryTriggerService) TriggerNextVisitTentativeReminder0D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	return s.runTentativeReminderBatch(ctx, clinicID, asOf, 0, model.TriggerTypeNextVisitTentativeReminder0d)
}

// runTentativeReminderBatch は tentative リマインダーの共通実行部。
// intended_date が asOf+daysBefore（JST 暦日）の active tentative を candidate とし、
// 送信直前に同一 intended_date の active tentative が残っているかを再検証する
// （取消・完了・削除・実予約昇格・日付変更された stale candidate は送信せず
// excluded ログを残す）。stage 識別子は triggerType 文字列そのもの。
func (s *lstepDeliveryTriggerService) runTentativeReminderBatch(
	ctx context.Context,
	clinicID uint64,
	asOf time.Time,
	daysBefore int,
	triggerType string,
) (int, []error) {
	if s.appointmentRepo == nil {
		// tentative リマインダーは appointment 状態が source of truth —— repo 未注入は
		// 構成不全であり、黙ってスキップすると配信漏れが検知できないため fail-closed
		// （lstep_batch_noshow.go の未注入依存と同じ契約）。
		err := apperrors.WrapInternalServerError("tentative reminder appointment repository is not configured")
		slog.ErrorContext(ctx, "delivery trigger tentative reminder: appointment repository is not configured",
			"clinic_id", clinicID, "trigger", triggerType)
		return 0, []error{err}
	}
	target := asOf.AddDate(0, 0, daysBefore)
	ownerIDs, err := s.appointmentRepo.FindOwnerIDsWithTentativeIntendedDate(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger tentative reminder: find owners error", "clinic_id", clinicID, "trigger", triggerType, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by tentative intended date")}
	}
	// EMR-308: 送信直前の stale 再検証。candidate 抽出時点以降に tentative が
	// 取消・完了・削除・実予約昇格・日付変更されていれば送信せず除外ログを残す。
	revalidate := func(ctx context.Context, ownerID uint64) (bool, string, error) {
		candidate, err := s.appointmentRepo.FindTentativeByOwnerAndIntendedDate(ctx, clinicID, ownerID, target)
		if err != nil {
			return false, "", apperrors.Wrap(err, "failed to revalidate tentative appointment")
		}
		return candidate != nil, "stale_tentative_appointment", nil
	}
	return s.runBatch(ctx, clinicID, ownerIDs, triggerType, triggerType, asOf, revalidate)
}

func (s *lstepDeliveryTriggerService) TriggerVaccineDeadline60(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, 60)
	ownerIDs, err := s.vaccinationRepo.FindOwnersByVaccineDeadline(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger vaccine_deadline_60d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by vaccine deadline")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeVaccineDeadline60, model.TriggerTypeVaccineDeadline60, asOf, nil)
}

func (s *lstepDeliveryTriggerService) TriggerVaccineDeadline30(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, 30)
	ownerIDs, err := s.vaccinationRepo.FindOwnersByVaccineDeadline(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger vaccine_deadline_30d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by vaccine deadline")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeVaccineDeadline30, model.TriggerTypeVaccineDeadline30, asOf, nil)
}

func (s *lstepDeliveryTriggerService) TriggerBirthdayMessage(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.petRepo.FindOwnersByPetBirthday(ctx, clinicID, int(asOf.Month()), asOf.Day())
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger birthday_message: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by pet birthday")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeBirthdayMessage, model.TriggerTypeBirthdayMessage, asOf, nil)
}

func (s *lstepDeliveryTriggerService) TriggerDormantPrevention180(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	thresholds, err := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_180d: get thresholds error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to get dormant thresholds")}
	}
	ownerIDs, err := s.medRecordRepo.FindOwnersByLastVisitDays(ctx, clinicID, thresholds.Stage180, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_180d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by last visit days")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention180, model.TriggerTypeDormantPrevention180, asOf, nil)
}

func (s *lstepDeliveryTriggerService) TriggerDormantPrevention210(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	thresholds, err := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_210d: get thresholds error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to get dormant thresholds")}
	}
	ownerIDs, err := s.medRecordRepo.FindOwnersByLastVisitDays(ctx, clinicID, thresholds.Stage210, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_210d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by last visit days")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention210, model.TriggerTypeDormantPrevention210, asOf, nil)
}

func (s *lstepDeliveryTriggerService) TriggerDormantPrevention240(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	thresholds, err := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_240d: get thresholds error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to get dormant thresholds")}
	}
	ownerIDs, err := s.medRecordRepo.FindOwnersByLastVisitDays(ctx, clinicID, thresholds.Stage240, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_240d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by last visit days")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention240, model.TriggerTypeDormantPrevention240, asOf, nil)
}

func (s *lstepDeliveryTriggerService) TriggerDormantPrevention365(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	thresholds, err := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_365d: get thresholds error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to get dormant thresholds")}
	}
	ownerIDs, err := s.medRecordRepo.FindOwnersByLastVisitDays(ctx, clinicID, thresholds.Stage365, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger dormant_prevention_365d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by last visit days")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention365, model.TriggerTypeDormantPrevention365, asOf, nil)
}

func (s *lstepDeliveryTriggerService) TriggerFilariaAlert(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.tagCacheRepo.FindOwnerIDsByTag(ctx, clinicID, tagFilariaAlert)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger filaria_alert: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by filaria tag")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFilariaAlert, model.TriggerTypeFilariaAlert, asOf, nil)
}

func (s *lstepDeliveryTriggerService) TriggerFleaTickAlert(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.tagCacheRepo.FindOwnerIDsByTag(ctx, clinicID, tagFleaTickAlert)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger flea_tick_alert: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by flea tick tag")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFleaTickAlert, model.TriggerTypeFleaTickAlert, asOf, nil)
}

func (s *lstepDeliveryTriggerService) TriggerFoodRefillReminder(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.tagCacheRepo.FindOwnerIDsByTag(ctx, clinicID, tagFoodRefill)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger food_refill_reminder: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by food refill tag")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFoodRefillReminder, model.TriggerTypeFoodRefillReminder, asOf, nil)
}

// TriggerFirstVisitWelcome は初診完了直後に呼び出されるイベント駆動トリガー（FEAT-383 Phase 2）。
func (s *lstepDeliveryTriggerService) TriggerFirstVisitWelcome(ctx context.Context, clinicID, ownerID uint64) error {
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger first_visit_welcome: failed to build lstep client", "clinic_id", clinicID, "owner_id", ownerID, "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}
	_, err = s.processSingleOwner(ctx, client, clinicID, ownerID, model.TriggerTypeFirstVisitWelcome, model.TriggerTypeFirstVisitWelcome, time.Now(), nil)
	return err
}

// TriggerCheckupFollowUp は健診作成直後に呼び出されるイベント駆動トリガー（FEAT-383 Phase 2）。
func (s *lstepDeliveryTriggerService) TriggerCheckupFollowUp(ctx context.Context, clinicID, ownerID uint64) error {
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger checkup_followup: failed to build lstep client", "clinic_id", clinicID, "owner_id", ownerID, "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}
	_, err = s.processSingleOwner(ctx, client, clinicID, ownerID, model.TriggerTypeCheckupFollowUp, model.TriggerTypeCheckupFollowUp, time.Now(), nil)
	return err
}
