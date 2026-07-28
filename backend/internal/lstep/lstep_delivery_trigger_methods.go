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
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFirstVisitFollowUp3D, model.TriggerTypeFirstVisitFollowUp3D, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerFirstVisitFollowUp7D(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, -7)
	ownerIDs, err := s.medRecordRepo.FindOwnersByFirstVisitDate(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger first_visit_followup_7d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by first visit date")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFirstVisitFollowUp7D, model.TriggerTypeFirstVisitFollowUp7D, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerNextVisitReminder(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.medRecordRepo.FindOwnersByNextVisitRecommended(ctx, clinicID, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger next_visit_reminder: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by next visit recommended date")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeNextVisitReminder, model.TriggerTypeNextVisitReminder, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerVaccineDeadline60(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, 60)
	ownerIDs, err := s.vaccinationRepo.FindOwnersByVaccineDeadline(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger vaccine_deadline_60d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by vaccine deadline")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeVaccineDeadline60, model.TriggerTypeVaccineDeadline60, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerVaccineDeadline30(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	target := asOf.AddDate(0, 0, 30)
	ownerIDs, err := s.vaccinationRepo.FindOwnersByVaccineDeadline(ctx, clinicID, target)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger vaccine_deadline_30d: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by vaccine deadline")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeVaccineDeadline30, model.TriggerTypeVaccineDeadline30, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerBirthdayMessage(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.petRepo.FindOwnersByPetBirthday(ctx, clinicID, int(asOf.Month()), asOf.Day())
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger birthday_message: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by pet birthday")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeBirthdayMessage, model.TriggerTypeBirthdayMessage, asOf)
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
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention180, model.TriggerTypeDormantPrevention180, asOf)
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
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention210, model.TriggerTypeDormantPrevention210, asOf)
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
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention240, model.TriggerTypeDormantPrevention240, asOf)
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
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeDormantPrevention365, model.TriggerTypeDormantPrevention365, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerFilariaAlert(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.tagCacheRepo.FindOwnerIDsByTag(ctx, clinicID, tagFilariaAlert)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger filaria_alert: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by filaria tag")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFilariaAlert, model.TriggerTypeFilariaAlert, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerFleaTickAlert(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.tagCacheRepo.FindOwnerIDsByTag(ctx, clinicID, tagFleaTickAlert)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger flea_tick_alert: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by flea tick tag")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFleaTickAlert, model.TriggerTypeFleaTickAlert, asOf)
}

func (s *lstepDeliveryTriggerService) TriggerFoodRefillReminder(ctx context.Context, clinicID uint64, asOf time.Time) (int, []error) {
	ownerIDs, err := s.tagCacheRepo.FindOwnerIDsByTag(ctx, clinicID, tagFoodRefill)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger food_refill_reminder: find owners error", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by food refill tag")}
	}
	return s.runBatch(ctx, clinicID, ownerIDs, model.TriggerTypeFoodRefillReminder, model.TriggerTypeFoodRefillReminder, asOf)
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
	_, err = s.processSingleOwner(ctx, client, clinicID, ownerID, model.TriggerTypeFirstVisitWelcome, model.TriggerTypeFirstVisitWelcome, time.Now())
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
	_, err = s.processSingleOwner(ctx, client, clinicID, ownerID, model.TriggerTypeCheckupFollowUp, model.TriggerTypeCheckupFollowUp, time.Now())
	return err
}
