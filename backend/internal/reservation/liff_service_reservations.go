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

	// EMR-308: tentative 同一行昇格のため、作成前に紐付け先 owner/pet を解決する
	// （best-effort）。解決できない顧客は nil のまま作成し、従来どおり作成後の
	// attach に委ねる。
	s.resolveReservationOwnerPet(ctx, clinicID, customerID, input)

	appt, previousOwnerID, err := s.validators.ValidateAndCreate(ctx, input)
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

	// EMR-308: tentative 昇格・最終 owner 反映を含む作成完了後に next_visit タグを
	// 再同期する（best-effort・未注入は no-op。owner 未解決なら helper 側で skip）。
	syncNextVisitTagsBestEffort(ctx, s.nextVisitTagSync, clinicID, previousOwnerID, appt.OwnerID)

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
// EMR-520: 一覧の pet ペアは認可パスを1クエリで解決する。取得失敗は一覧全体のエラー
// とする（静かに null へ落とさない）。未解決の予約は PetPairs に載らず両方 null。
func (s *liffService) GetMyReservations(ctx context.Context, clinicID, customerID uint64) (*LiffMyReservationsResult, error) {
	items, err := s.adminRepo.FindAllByCustomerID(ctx, clinicID, customerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get my reservations")
	}
	ids := make([]uint64, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
	}
	pairs, err := s.adminRepo.FindLiffPetPairs(ctx, clinicID, customerID, ids)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get my reservation pet pairs")
	}
	return &LiffMyReservationsResult{Items: items, PetPairs: pairs}, nil
}

// CancelReservation は予約をキャンセルする。
func (s *liffService) CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error {
	// Phase 6: キャンセル通知に必要な関連情報を取得する。
	var apptForNotify *model.Reservation
	if s.notifier != nil {
		var err error
		apptForNotify, err = s.adminRepo.FindByIDForNotify(ctx, clinicID, reservationID)
		if err != nil {
			slog.WarnContext(ctx, "failed to find appointment for notification (best-effort)", "error", err)
		}
	}

	cancelledOwnerID, err := s.adminRepo.CancelByID(ctx, clinicID, customerID, reservationID)
	if err != nil {
		return apperrors.Wrap(err, "failed to cancel reservation")
	}

	// EMR-308: 通常の LIFF handler では取消 UPDATE の commit 後に同期する。
	// 通知用の古い snapshot ではなく、実際に取り消した行の飼主を使用する。
	syncNextVisitTagBestEffort(ctx, s.nextVisitTagSync, clinicID, cancelledOwnerID)

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

// resolveReservationOwnerPet は EMR-308 の tentative 同一行昇格のため、予約作成前に
// 紐付け先 owner/pet を input.OwnerID / input.PetID へ best-effort で解決する。
// SEC-CS2-F02: name+phone による line_customers.owner_id 自動紐付けは行わず、
// 既にトークン/スタッフ経路で紐付け済みの顧客だけを対象にする。未紐付け・解決失敗・
// 死亡ペット再確認に失敗した場合は nil のまま返り、作成後の attach に委ねる。
func (s *liffService) resolveReservationOwnerPet(
	ctx context.Context,
	clinicID, customerID uint64,
	input *CreateReservationInput,
) {
	customer, err := s.findCustomerWithOwnerSync(ctx, clinicID, customerID)
	if err != nil || customer == nil || customer.OwnerID == nil || customer.Owner == nil {
		return
	}
	ownerID := customer.Owner.ID
	input.OwnerID = &ownerID
	petID := resolveReservationPetID(customer, input.CustomerFields)
	if petID == nil || s.reservationRepo == nil {
		return
	}
	// EMR-235: 作成へ pet_id を載せる前に DB で死亡状態を再確認する（fail-closed）。
	if err := ValidateReservationPetNotDeceased(ctx, s.reservationRepo, clinicID, petID); err != nil {
		slog.WarnContext(ctx, "skipping pre-resolved pet_id: deceased re-check failed (best-effort)", "error", err)
		return
	}
	input.PetID = petID
}

// tryAttachReservationOwnerPet は line_customer の owner 紐付け状態に応じて、
// 予約へ owner_id / pet_id を best-effort で反映する。
// EMR-308: 作成前解決済み（appt.OwnerID 非nil）の場合は再 attach しない。
func (s *liffService) tryAttachReservationOwnerPet(
	ctx context.Context,
	clinicID, customerID uint64,
	appt *model.Reservation,
	customerFields []byte,
) {
	if s.reservationRepo == nil || appt == nil {
		return
	}
	// EMR-308: 作成時に owner を解決済み（pre-resolve または tentative 昇格で維持）
	// の場合、作成後 attach は不要。
	if appt.OwnerID != nil {
		return
	}

	customer, err := s.findCustomerWithOwnerSync(ctx, clinicID, customerID)
	if err != nil || customer == nil || customer.OwnerID == nil || customer.Owner == nil {
		return
	}

	fields := map[string]any{
		"owner_id": customer.Owner.ID,
	}
	if petID := resolveReservationPetID(customer, customerFields); petID != nil {
		// EMR-235: preload スナップショットはロックなしで、登録と attach の間に
		// 死亡が記録されうるため、pet_id 書込前に DB で再確認する（fail-closed）。
		if err := ValidateReservationPetNotDeceased(ctx, s.reservationRepo, clinicID, petID); err != nil {
			slog.WarnContext(ctx, "skipping pet_id attach: pet deceased re-check failed (best-effort)", "error", err)
		} else {
			fields["pet_id"] = *petID
		}
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

// livingPets は生存ペットだけを返す（#261 P0 / EMR-290）。
// 死亡契約は status=deceased または deceased_at 非NULL（sharedkernel.ValidatePetNotDeceased と同じ
// OR 契約）。旧データは status=deceased でも deceased_at IS NULL の行があり、deceased_at だけを
// 見ると死亡個体が予約の pet_id 候補に混入する。
func livingPets(pets []model.Pet) []model.Pet {
	out := make([]model.Pet, 0, len(pets))
	for i := range pets {
		if pets[i].Status == model.PetStatusDeceased || pets[i].DeceasedAt != nil {
			continue
		}
		out = append(out, pets[i])
	}
	return out
}
