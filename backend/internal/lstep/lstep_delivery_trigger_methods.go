package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// maxTentativeReminderLockAttempts は tentative リマインダー直列化で
// candidate 集合が拡大し続けた場合の再ロック試行上限（Codex #22 再指摘）。
// tentative の新規追加が収束しない異常 churn は送信しないで表面化する。
const maxTentativeReminderLockAttempts = 4

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
	// Codex: 再検証 read と外部送信の間に mutation が割り込めるため、tentative
	// mutation と同じ (clinic,pet) advisory lock を保持する tx 内で再読み込み→
	// 送信を直列化する（WithTentativeReminderSerialization）。
	// Codex #22: 同一 owner の複数 pet が対象日の tentative を持ち得るため、
	// candidate 集合の全 pet をロックする（1件だけロックして owner+date で再検索
	// すると、ロック外の別 pet の intent で再検証を通過し得る）。再検証も
	// ロック中の全 candidate pet を対象に行う。
	// Codex #22 再指摘（PR #621 review #4）: 初回候補の列挙〜ロック取得の間に
	// 新しい pet の tentative が commit されると、ロック済み集合外の intent が
	// 再検索に混ざり送信根拠に化け得る。再検索がロック済み集合を超える pet を
	// 返した場合は送信せず、拡大後の集合でロックを取り直して再検証する
	// （attempt 上限で churn には fail-closed）。
	revalidate := func(ctx context.Context, ownerID uint64, send func(context.Context) error) (bool, string, error) {
		for attempt := 0; ; attempt++ {
			candidates, err := s.appointmentRepo.FindTentativesByOwnerAndIntendedDate(ctx, clinicID, ownerID, target)
			if err != nil {
				return false, "", apperrors.Wrap(err, "failed to revalidate tentative appointment")
			}
			petIDs := make([]uint64, 0, len(candidates))
			for _, c := range candidates {
				if c.PetID != nil {
					petIDs = append(petIDs, *c.PetID)
				}
			}
			if len(petIDs) == 0 {
				return false, "stale_tentative_appointment", nil
			}
			locked := make(map[uint64]struct{}, len(petIDs))
			for _, petID := range petIDs {
				locked[petID] = struct{}{}
			}
			stillValid := false
			setGrewBeyondLock := false
			serErr := s.appointmentRepo.WithTentativeReminderSerialization(ctx, clinicID, petIDs, func(txCtx context.Context) error {
				// ロック取得時点のコミット済み状態で再検証 —— 同時実行中の mutation は
				// ここで待機し、コミット済みなら再読み込みに反映される。同一 owner+date
				// の active tentative が1件でも残っていれば owner 宛て送信は正当。
				current, err := s.appointmentRepo.FindTentativesByOwnerAndIntendedDate(txCtx, clinicID, ownerID, target)
				if err != nil {
					return apperrors.Wrap(err, "failed to revalidate tentative appointment")
				}
				// 再検索がロック済み集合外の pet を含む場合、その pet の mutation は
				// 直列化されていないため送信根拠に採用しない。ロックを取り直す。
				for _, c := range current {
					if c.PetID == nil {
						setGrewBeyondLock = true // pet 未解決行は直列化不能
						continue
					}
					if _, ok := locked[*c.PetID]; !ok {
						setGrewBeyondLock = true
					}
				}
				if setGrewBeyondLock || len(current) == 0 {
					return nil // 送信せず tx のみ commit（stale または再ロックへ）
				}
				// 外部送信・ログ書込は元 ctx（tx 外）で実行 —— 送信自体はロック保持中の
				// 直列化境界内にあり、送信後のログ更新が tx ロールバックで巻き戻るのを防ぐ。
				if err := send(ctx); err != nil {
					return err
				}
				stillValid = true
				return nil
			})
			if serErr != nil {
				return false, "", serErr
			}
			if setGrewBeyondLock {
				if attempt+1 >= maxTentativeReminderLockAttempts {
					// 集合が拡大し続ける異常 churn —— 未送信のままエラーで表面化する。
					return false, "", apperrors.WrapInternalServerError(
						"tentative reminder candidate set kept growing during serialization")
				}
				continue
			}
			if !stillValid {
				return false, "stale_tentative_appointment", nil
			}
			return true, "", nil
		}
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
