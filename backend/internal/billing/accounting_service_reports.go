package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetPeriodUnpaidCarryover は指定期間の未納繰越（期間前繰越・期間内未納・期末繰越）を
// 飼主+ペット単位で返す。EMR-188: startDate/endDate は handler の期間クエリで検証済み。
func (s *accountingService) GetPeriodUnpaidCarryover(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]PeriodUnpaidOwnerPet, int64, PeriodUnpaidSummary, error) {
	items, total, summary, err := s.repo.FindPeriodUnpaidCarryover(ctx, clinicID, startDate, endDate, page, limit)
	if err != nil {
		return nil, 0, summary, apperrors.Wrap(err, "failed to get period unpaid carryover")
	}
	return items, total, summary, nil
}

// #120: start_date/end_date 2引数
func (s *accountingService) ListUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error) {
	result, total, err := s.repo.FindUnpaidByBilling(ctx, clinicID, startDate, endDate, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list unpaid billings")
	}
	return result, total, nil
}

// #120: start_date/end_date 2引数（飼主単位集約）
func (s *accountingService) ListUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
	result, total, summary, err := s.repo.FindUnpaidByOwner(ctx, clinicID, startDate, endDate, page, limit)
	if err != nil {
		return nil, 0, summary, apperrors.Wrap(err, "failed to list unpaid by owner")
	}
	return result, total, summary, nil
}

// GetOwnerUnpaidBalance は会計画面表示用に飼主の未納残高を返す。#182
func (s *accountingService) GetOwnerUnpaidBalance(ctx context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error) {
	if ownerID == 0 {
		return OwnerUnpaidBalance{}, apperrors.WrapInvalidInput("owner_id is required")
	}
	result, err := s.repo.SumUnpaidByOwner(ctx, clinicID, ownerID)
	if err != nil {
		return OwnerUnpaidBalance{}, apperrors.Wrap(err, "failed to get owner unpaid balance")
	}
	return result, nil
}

// Cancel は会計のキャンセルを受け付けない。held for EMR-505（cancel の slice。inventory §7 の cancel 行）。
// 旧実装は billings.status を cancelled へ更新し監査ログを書いていたが、status 列は廃止された（001_init.sql:10889）。
// 後継は claim_state='cancelled'、理由が必須の billing_change_events(intent='cancel')、請求元の claim 解放
// （billing_source_bindings.claim_released_*）、再請求リンクを同一 TX で書く操作で、cancel の操作 family（EMR-489）も未登録のため、
// 書き込み前に固定エラー（422）で拒否する。成功を装わない。
// 読取だけは従来どおり行う: 会計の存在確認（clinic スコープ。なければ 404）と、キャンセル済みの二重キャンセル（409。AC-12）。
func (s *accountingService) Cancel(ctx context.Context, clinicID, id uint64, _ *uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find accounting for cancel")
	}
	if billingClaimState(existing) == billingClaimStateCancelled {
		return apperrors.WrapConflict("既にキャンセル済みの会計です")
	}
	return apperrors.WrapFinanceConsumerPending("会計のキャンセルは、理由を記録する変更イベントと請求元の claim 解放が未接続のため利用できません")
}

// parseDailySummaryDate は dateStr（空なら今日）を JST でパースする共通ヘルパー。
func parseDailySummaryDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		dateStr = time.Now().In(time.Local).Format(time.DateOnly)
	}
	date, err := time.ParseInLocation(time.DateOnly, dateStr, config.JST)
	if err != nil {
		return time.Time{}, apperrors.WrapInvalidInput("date must be YYYY-MM-DD")
	}
	return date, nil
}

// GetDailySummary は指定日のレジ締め集計を返す。BUG-368
func (s *accountingService) GetDailySummary(ctx context.Context, clinicID uint64, dateStr string) (*DailySummaryResult, error) {
	date, err := parseDailySummaryDate(dateStr)
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetDailySummary(ctx, clinicID, date)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get daily summary")
	}
	return result, nil
}

// GetDailySummaryForClinics は複数医院の拠点別日次集計を返す (#86 段階3 論点4=2)。
func (s *accountingService) GetDailySummaryForClinics(ctx context.Context, clinicIDs []uint64, dateStr string) ([]ClinicDailySummary, error) {
	date, err := parseDailySummaryDate(dateStr)
	if err != nil {
		return nil, err
	}
	results := make([]ClinicDailySummary, 0, len(clinicIDs))
	for _, clinicID := range clinicIDs {
		r, rerr := s.repo.GetDailySummary(ctx, clinicID, date)
		if rerr != nil {
			return nil, apperrors.Wrap(rerr, "failed to get daily summary for clinics")
		}
		results = append(results, ClinicDailySummary{ClinicID: clinicID, Summary: r})
	}
	return results, nil
}
