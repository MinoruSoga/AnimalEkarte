package reservation

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateReservation は予約を確定する。staffID=0 の場合は no_staff_mode に従って自動割当する。
func (s *liffService) CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Reservation, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation setting")
	}
	input.ClinicID = clinicID
	input.CustomerID = customerID
	input.Settings = setting

	// TASK-RES-025: 指名なし委譲ロジック
	// RSV-09 / DEC-35: keep best-effort auto-delegate, but never discard errors silently.
	if input.StaffID == 0 {
		s.tryAutoDelegateStaff(ctx, clinicID, input, setting)
	}

	appt, err := s.validators.ValidateAndCreate(ctx, input)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to validate and create appointment")
	}

	// 顧客の追加フィールドを更新（プロフィール自動保存）
	if len(input.CustomerFields) > 0 && string(input.CustomerFields) != "{}" {
		if err := s.customerRepo.UpdateAdditionalFields(ctx, clinicID, customerID, input.CustomerFields); err != nil {
			slog.WarnContext(ctx, "failed to update customer additional fields (best-effort)", "error", err)
		}
	}

	// SEC-CS2-F02: name+phone による line_customers.owner_id 自動紐付けは行わない。
	// 既にトークン/スタッフ経路で紐付け済みの顧客のみ、予約へ owner/pet を反映する。
	s.tryAttachReservationOwnerPet(ctx, clinicID, customerID, appt, input.CustomerFields)

	// Phase 6: 予約確定通知（LINE + メール）fire-and-forget
	if s.notifier != nil {
		// 通知メッセージ用にリレーションをロード（enriched が nil の場合は元の appt を使う）
		notifyAppt := appt
		if enriched, err := s.adminRepo.FindByIDForNotify(ctx, clinicID, appt.ID); err == nil && enriched != nil {
			notifyAppt = enriched
		}
		// best-effort: 通知失敗は予約成否に影響させない
		customer, custErr := s.customerRepo.FindByID(ctx, clinicID, customerID)
		if custErr != nil {
			slog.WarnContext(ctx, "failed to get customer for notification (best-effort)", "error", custErr)
		}
		s.notifier.NotifyCreated(ctx, notifyAppt, customer)
	}

	return appt, nil
}

func (s *liffService) tryAutoDelegateStaff(
	ctx context.Context,
	clinicID uint64,
	input *CreateReservationInput,
	setting *model.LineReservationSetting,
) {
	date, err := ToDateTime(input.Date, input.StartTime)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse date for staff auto-delegate (best-effort)", "error", err)
		return
	}
	assignedID, err := s.delegateStaff(ctx, clinicID, input.ReservationTypeID, setting.NoStaffMode, date, input.StartTime, input.EndTime)
	if err != nil {
		slog.WarnContext(ctx, "failed to auto-delegate staff for LIFF reservation (best-effort)", "error", err)
		return
	}
	if assignedID != 0 {
		input.StaffID = assignedID
	}
}

// GetMyReservations は顧客自身の予約一覧を返す。
func (s *liffService) GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error) {
	items, err := s.adminRepo.FindAllByCustomerID(ctx, clinicID, customerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get my reservations")
	}
	return items, nil
}

// CancelReservation は予約をキャンセルする。
func (s *liffService) CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error {
	// Phase 6: キャンセル通知のために事前にアポを取得する
	var apptForNotify *model.Reservation
	if s.notifier != nil {
		var err error
		apptForNotify, err = s.adminRepo.FindByIDForNotify(ctx, clinicID, reservationID)
		if err != nil {
			slog.WarnContext(ctx, "failed to find appointment for notification (best-effort)", "error", err)
		}
	}

	if err := s.adminRepo.CancelByID(ctx, clinicID, customerID, reservationID); err != nil {
		return apperrors.Wrap(err, "failed to cancel reservation")
	}

	if s.medicalRecord != nil {
		s.medicalRecord.DeleteDraftFromReservation(ctx, clinicID, reservationID)
	}

	// Phase 6: キャンセル通知（LINE + メール）fire-and-forget
	if s.notifier != nil && apptForNotify != nil {
		// best-effort: 通知失敗は予約キャンセル成否に影響させない
		customer, custErr := s.customerRepo.FindByID(ctx, clinicID, customerID)
		if custErr != nil {
			slog.WarnContext(ctx, "failed to get customer for cancel notification (best-effort)", "error", custErr)
		}
		s.notifier.NotifyCancelled(ctx, apptForNotify, customer)
	}

	return nil
}

// tryAttachReservationOwnerPet は line_customer の owner 紐付け状態に応じて、
// 予約へ owner_id / pet_id を best-effort で反映する。
func (s *liffService) tryAttachReservationOwnerPet(
	ctx context.Context,
	clinicID, customerID uint64,
	appt *model.Reservation,
	customerFields []byte,
) {
	if s.reservationRepo == nil || appt == nil {
		return
	}

	customer, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
	if err != nil || customer == nil || customer.OwnerID == nil || customer.Owner == nil {
		return
	}

	fields := map[string]any{
		"owner_id": customer.Owner.ID,
	}
	if petID := resolveReservationPetID(customer, customerFields); petID != nil {
		fields["pet_id"] = *petID
	}

	updated, err := s.reservationRepo.update(ctx, clinicID, appt.ID, fields)
	if err != nil {
		slog.WarnContext(ctx, "failed to attach owner/pet to line reservation (best-effort)", "error", err)
		return
	}
	*appt = *updated
}

// ---- 内部ヘルパー ----

func resolveReservationPetID(customer *model.LineCustomer, customerFields []byte) *uint64 {
	if customer == nil || customer.Owner == nil {
		return nil
	}
	// #261 P0: 死亡ペットは予約へ pet_id を付けない（health card 表示除外と揃える fail-closed）。
	living := livingPets(customer.Owner.Pets)
	if len(living) == 0 {
		return nil
	}
	if len(living) == 1 {
		id := living[0].ID
		return &id
	}

	var fields struct {
		Pets []struct {
			Name string `json:"name"`
		} `json:"pets"`
	}
	if len(customerFields) == 0 || json.Unmarshal(customerFields, &fields) != nil || len(fields.Pets) == 0 {
		return nil
	}

	wantName := strings.TrimSpace(fields.Pets[0].Name)
	if wantName == "" {
		return nil
	}
	for i := range living {
		if strings.TrimSpace(living[i].Name) == wantName {
			id := living[i].ID
			return &id
		}
	}
	return nil
}

// livingPets は deceased_at が nil のペットだけを返す（#261 P0）。
func livingPets(pets []model.Pet) []model.Pet {
	out := make([]model.Pet, 0, len(pets))
	for i := range pets {
		if pets[i].DeceasedAt == nil {
			out = append(out, pets[i])
		}
	}
	return out
}
