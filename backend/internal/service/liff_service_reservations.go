package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateReservation は予約を確定する。staffID=0 の場合は no_staff_mode に従って自動割当する。
func (s *liffService) CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Reservation, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation setting", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation setting")
	}
	input.ClinicID = clinicID
	input.CustomerID = customerID
	input.Settings = setting

	// TASK-RES-025: 指名なし委譲ロジック
	if input.StaffID == 0 {
		date, err := toDateTime(input.Date, input.StartTime)
		if err == nil {
			assignedID, err := s.delegateStaff(ctx, clinicID, input.ReservationTypeID, setting.NoStaffMode, date, input.StartTime, input.EndTime)
			if err == nil && assignedID != 0 {
				input.StaffID = assignedID
			}
		}
	}

	appt, err := s.validators.ValidateAndCreate(ctx, input)
	if err != nil {
		slog.ErrorContext(ctx, "failed to validate and create appointment", "error", err)
		return nil, apperrors.Wrap(err, "failed to validate and create appointment")
	}

	// BE-120: category=trimming の場合、トリミング詳細を作成（best-effort）
	if input.TrimmingCourseID != nil {
		detail := &model.AppointmentTrimmingDetail{
			ClinicID:      clinicID,
			AppointmentID: appt.ID,
			CourseID:      input.TrimmingCourseID,
			StyleRequest:  input.TrimmingStyleRequest,
		}
		if err := s.trimmingDetailRepo.Create(ctx, detail); err != nil {
			slog.WarnContext(ctx, "failed to create trimming detail (best-effort)", "error", err, "appointment_id", appt.ID)
		} else if len(input.TrimmingOptionIDs) > 0 {
			if err := s.trimmingDetailRepo.SetOptions(ctx, appt.ID, input.TrimmingOptionIDs); err != nil {
				slog.WarnContext(ctx, "failed to set trimming options (best-effort)", "error", err, "appointment_id", appt.ID)
			}
		}
	}

	// 顧客の追加フィールドを更新（プロフィール自動保存）
	if len(input.CustomerFields) > 0 && string(input.CustomerFields) != "{}" {
		if err := s.customerRepo.UpdateAdditionalFields(ctx, clinicID, customerID, input.CustomerFields); err != nil {
			slog.WarnContext(ctx, "failed to update customer additional fields (best-effort)", "error", err)
		}
	}

	// 自動オーナー紐付け: customer_fields の氏名+電話番号で owners を検索し、1件一致で自動リンク
	s.tryAutoLinkOwner(ctx, clinicID, customerID, input.CustomerFields)
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

// GetMyReservations は顧客自身の予約一覧を返す。
func (s *liffService) GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error) {
	items, err := s.adminRepo.FindAllByCustomerID(ctx, clinicID, customerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get my reservations", "error", err)
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
	if err != nil || customer == nil || customer.OwnerID == nil {
		return
	}

	fields := map[string]any{
		"owner_id": *customer.OwnerID,
	}
	if petID := resolveReservationPetID(customer, customerFields); petID != nil {
		fields["pet_id"] = *petID
	}

	updated, err := s.reservationRepo.Update(ctx, clinicID, appt.ID, fields)
	if err != nil {
		slog.WarnContext(ctx, "failed to attach owner/pet to line reservation (best-effort)", "error", err)
		return
	}
	*appt = *updated
}

// ---- 内部ヘルパー ----

func resolveReservationPetID(customer *model.LineCustomer, customerFields []byte) *uint64 {
	if customer == nil || customer.Owner == nil || len(customer.Owner.Pets) == 0 {
		return nil
	}
	if len(customer.Owner.Pets) == 1 {
		id := customer.Owner.Pets[0].ID
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
	for i := range customer.Owner.Pets {
		if strings.TrimSpace(customer.Owner.Pets[i].Name) == wantName {
			id := customer.Owner.Pets[i].ID
			return &id
		}
	}
	return nil
}

// tryAutoLinkOwner は予約顧客の氏名+電話番号で owners テーブルを検索し、
// 1件だけ一致した場合に line_customers.owner_id を自動紐付けする。
// best-effort: 失敗しても予約処理は中断しない。
func (s *liffService) tryAutoLinkOwner(ctx context.Context, clinicID, customerID uint64, customerFields []byte) {
	if s.ownerRepo == nil {
		return
	}

	// 既に紐付け済みならスキップ
	customer, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
	if err != nil || customer == nil || customer.OwnerID != nil {
		return
	}

	// customer_fields から氏名と電話番号を抽出
	var fields struct {
		CustomerName string `json:"customer_name"`
		Phone        string `json:"phone"`
		OwnerName    string `json:"owner_name"`
	}
	if len(customerFields) == 0 {
		return
	}
	if err := json.Unmarshal(customerFields, &fields); err != nil {
		return
	}

	// owner_name を優先、なければ customer_name を使用
	name := fields.OwnerName
	if name == "" {
		name = fields.CustomerName
	}
	phone := fields.Phone
	if name == "" || phone == "" {
		return
	}

	// owners テーブルで氏名+電話番号の完全一致検索（1件のみ返す）
	owner, err := s.ownerRepo.FindByNameAndPhone(ctx, clinicID, name, phone)
	if err != nil {
		slog.WarnContext(ctx, "auto-link owner lookup failed (best-effort)", "error", err)
		return
	}
	if owner == nil {
		return // 0件 or 複数件 → 紐付けしない
	}

	// 自動紐付け実行
	if err := s.customerRepo.UpdateOwnerLink(ctx, clinicID, customerID, &owner.ID); err != nil {
		slog.WarnContext(ctx, "auto-link owner update failed (best-effort)", "error", err)
		return
	}
	slog.InfoContext(ctx, "auto-linked LINE customer to owner",
		"customer_id", customerID, "owner_id", owner.ID, "name", name)
}
