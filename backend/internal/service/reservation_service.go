package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// allowedReservationRoutes は予約経路の許可値ホワイトリスト（FEAT-381-2 Commit 3）。
var allowedReservationRoutes = map[string]struct{}{
	"line": {}, "phone": {}, "reception": {}, "exam_room": {}, "record_shortcut": {},
}

const colReservationRoute = "reservation_route"
const allowedReservationRoutesMessage = "reservation_route must be one of: line, phone, reception, exam_room, record_shortcut"

// UpdateReservationRouteInput は予約経路更新の入力DTO（FEAT-381-2 Commit 3）。
type UpdateReservationRouteInput struct{ Route string }

// CreateManualReservationInput は管理画面からの予約作成入力 DTO。
type CreateManualReservationInput struct {
	ClinicID          uint64
	StartTime         time.Time
	EndTime           time.Time
	OwnerID           *uint64
	PetID             *uint64
	VisitType         model.VisitType
	ReservationTypeID uint64
	DoctorID          *uint64
	IsDesignated      bool
	Status            model.ReservationStatus
	Notes             string
	Source            model.ReservationSource
	CreatedBy         *uint64
	ReservationRoute  *string
}

// UpdateReservationInput は予約更新のサービス入力 DTO
type UpdateReservationInput struct {
	StartTime         *time.Time
	EndTime           *time.Time
	OwnerID           *uint64
	PetID             *uint64
	VisitType         *model.VisitType
	ReservationTypeID *uint64
	DoctorID          *uint64
	IsDesignated      *bool
	Status            *model.ReservationStatus
	Notes             *string
}

func buildReservationUpdate(input *UpdateReservationInput) map[string]any {
	fields := make(map[string]any)
	if input.StartTime != nil {
		fields["start_time"] = *input.StartTime
	}
	if input.EndTime != nil {
		fields["end_time"] = *input.EndTime
	}
	if input.OwnerID != nil {
		if *input.OwnerID == 0 {
			fields["owner_id"] = nil // 0 は「飼主クリア」として NULL に設定
		} else {
			fields["owner_id"] = *input.OwnerID
		}
	}
	if input.PetID != nil {
		if *input.PetID == 0 {
			fields["pet_id"] = nil // 0 は「ペットクリア」として NULL に設定
		} else {
			fields["pet_id"] = *input.PetID
		}
	}
	if input.VisitType != nil {
		fields["visit_type"] = *input.VisitType
	}
	if input.ReservationTypeID != nil {
		fields["reservation_type_id"] = *input.ReservationTypeID
	}
	if input.DoctorID != nil {
		if *input.DoctorID == 0 {
			fields["doctor_id"] = nil // 0 は「医師未指定に変更」として NULL に設定
		} else {
			fields["doctor_id"] = *input.DoctorID
		}
	}
	if input.IsDesignated != nil {
		fields["is_designated"] = *input.IsDesignated
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	return fields
}

type ReservationService interface {
	// List は指定した複数医院 (#86 拠点横断) の予約一覧を返す。clinicIDs はハンドラ層で所属検証済みであること。
	List(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	// GetByIDForClinics は複数医院スコープで予約を1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error)
	Create(ctx context.Context, input *CreateManualReservationInput) (*model.Reservation, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.Reservation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	UpdateReservationRoute(ctx context.Context, clinicID, id uint64, input UpdateReservationRouteInput) (*model.Reservation, error)
}

type reservationService struct {
	repo                 repository.ReservationRepository
	typeRepo             reservationTypeFinder
	tx                   repository.Transactor
	reservationStaffRepo repository.ReservationStaffRepository
	unavailableTimeRepo  repository.ReservationTypeUnavailableTimeRepository
	availableSlotRepo    repository.ReservationTypeAvailableSlotRepository
}

func NewReservationServiceWithAvailabilityAndType(repo repository.ReservationRepository, typeRepo reservationTypeFinder, tx repository.Transactor, reservationStaffRepo repository.ReservationStaffRepository, unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository, availableSlotRepo ...repository.ReservationTypeAvailableSlotRepository) ReservationService {
	var slotRepo repository.ReservationTypeAvailableSlotRepository
	if len(availableSlotRepo) > 0 {
		slotRepo = availableSlotRepo[0]
	}
	return &reservationService{
		repo:                 repo,
		typeRepo:             typeRepo,
		tx:                   tx,
		reservationStaffRepo: reservationStaffRepo,
		unavailableTimeRepo:  unavailableTimeRepo,
		availableSlotRepo:    slotRepo,
	}
}

// resolveFinalOwnerPet は現在値と PATCH 値を統合した最終 OwnerID/PetID を返す。
// 入力が 0 の場合はクリア（nil）として扱う（DoctorID と同様）。
func resolveFinalOwnerPet(current *model.Reservation, input *UpdateReservationInput) (ownerID, petID *uint64) {
	ownerID = current.OwnerID
	petID = current.PetID
	if input.OwnerID != nil {
		if *input.OwnerID == 0 {
			ownerID = nil
		} else {
			ownerID = input.OwnerID
		}
	}
	if input.PetID != nil {
		if *input.PetID == 0 {
			petID = nil
		} else {
			petID = input.PetID
		}
	}
	return ownerID, petID
}

// validateReservationOwnerPetLinks は sharedkernel.ValidateReservationOwnerPetLinks への
// 既存呼び出し面互換 delegate（実装正本は sharedkernel・BE9-2D ⑤ Batch B 昇格。
// medicalrecord の hospitalization 系と恒久共有のため）。
func validateReservationOwnerPetLinks(ctx context.Context, repo repository.ReservationRepository, clinicID uint64, ownerID, petID *uint64) error {
	return sharedkernel.ValidateReservationOwnerPetLinks(ctx, repo, clinicID, ownerID, petID)
}

func (s *reservationService) List(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicIDs, page, limit, date, startDate, endDate, status, source, petID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list reservations", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list reservations")
	}
	return items, total, nil
}

func (s *reservationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation")
	}
	return result, nil
}

func (s *reservationService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error) {
	result, err := s.repo.FindByIDForClinics(ctx, clinicIDs, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation for clinics", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation for clinics")
	}
	return result, nil
}

func (s *reservationService) Create(ctx context.Context, input *CreateManualReservationInput) (*model.Reservation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	if input.ReservationRoute != nil {
		if _, ok := allowedReservationRoutes[*input.ReservationRoute]; !ok {
			return nil, apperrors.WrapInvalidInput(allowedReservationRoutesMessage)
		}
	}
	if err := reservation.ValidateReservationStaffCapability(ctx, s.reservationStaffRepo, input.ClinicID, input.DoctorID, input.ReservationTypeID); err != nil {
		return nil, err
	}
	enforceBookingConstraints := shouldEnforceReservationBookingConstraints(input.Status, input.ReservationRoute)
	if enforceBookingConstraints {
		if err := reservation.ValidateReservationTypeAvailableTime(ctx, s.unavailableTimeRepo, input.ClinicID, input.ReservationTypeID, input.StartTime, input.EndTime); err != nil {
			return nil, err
		}
	}
	if err := validateTimeRange(input.StartTime, input.EndTime); err != nil {
		return nil, err
	}
	// BE-refactor.md X-14/U6b: クロステナント write 防止。reception/exam_room/record_shortcut
	// 等の shortcut 経路や確定済みステータスでは enforceBookingConstraints=false となり、
	// 下の checkReservationTypeCapacity(容量チェック内の FindByID)がスキップされるため、
	// ReservationTypeID の所有権検証は経路・ステータスに関わらず常にここで行う
	// (reservation_validators.go の typeRepo と同型の nil 許容パターン)。
	if s.typeRepo != nil {
		if _, err := s.typeRepo.FindByID(ctx, input.ClinicID, input.ReservationTypeID); err != nil {
			slog.ErrorContext(ctx, "reservation type not found or belongs to different clinic", "error", err)
			return nil, apperrors.Wrap(err, "failed to verify reservation type ownership")
		}
	}
	reservation := &model.Reservation{
		ClinicID:          input.ClinicID,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		OwnerID:           input.OwnerID,
		PetID:             input.PetID,
		VisitType:         input.VisitType,
		ReservationTypeID: input.ReservationTypeID,
		DoctorID:          input.DoctorID,
		IsDesignated:      input.IsDesignated,
		Status:            input.Status,
		Notes:             input.Notes,
		Source:            input.Source,
		CreatedBy:         input.CreatedBy,
		ReservationRoute:  input.ReservationRoute,
	}

	// SELECT FOR UPDATE + トランザクションで競合を防止
	// LINE予約・電子カルテ予約・管理者手動予約すべてで同一テーブルを使用するため、
	// アプリケーションレベルでの排他制御が必要
	// BE-refactor.md X-9: 空き枠（既存行 0 件）は SELECT FOR UPDATE が何もロックしないため、
	// AcquireBookingLock（clinic 単位 advisory xact lock）で競合チェック～INSERT を直列化する。
	if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := validateReservationOwnerPetLinks(ctx, s.repo, reservation.ClinicID, reservation.OwnerID, reservation.PetID); err != nil {
			return err
		}
		if enforceBookingConstraints {
			if err := s.repo.AcquireBookingLock(ctx, reservation.ClinicID); err != nil {
				return err
			}
			if err := checkSlotConflict(ctx, s.repo, reservation.ClinicID, reservation.DoctorID, reservation.StartTime, reservation.EndTime, nil); err != nil {
				return err
			}
			if err := checkReservationTypeCapacity(ctx, s.repo, s.typeRepo, reservation.ClinicID, reservation.ReservationTypeID, reservation.StartTime, nil); err != nil {
				return err
			}
		}
		return s.repo.Create(ctx, reservation)
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create reservation")
	}

	slog.InfoContext(ctx, "reservation created",
		slog.Uint64("reservation_id", reservation.ID),
		slog.Uint64("clinic_id", reservation.ClinicID))
	return reservation, nil
}

func shouldEnforceReservationBookingConstraints(status model.ReservationStatus, route *string) bool {
	if route != nil {
		switch *route {
		case "reception", "exam_room", "record_shortcut":
			return false
		}
	}
	switch status {
	case model.ReservationStatusCheckedIn,
		model.ReservationStatusInConsultation,
		model.ReservationStatusAccounting,
		model.ReservationStatusCompleted:
		return false
	default:
		return true
	}
}

// validateTimeRange は end_time > start_time を確認する共通バリデーション
// （BE9-2C R①: 実装は sharedkernel.ValidateTimeRange へ昇格・本関数は既存呼び出し面互換の delegate）。
func validateTimeRange(startTime, endTime time.Time) error {
	return sharedkernel.ValidateTimeRange(startTime, endTime)
}

// errNoDoctorsOnDuty は当日の出勤医師が 0 人のためスロット予約不可を示すセンチネルエラー。
// *apperrors.AppError なので RespondError が errors.As で Message を抽出し日本語メッセージを返す。
// LINE パスでは reservation_validators.go が errors.Is でこれを識別し RedirectStep: 4 を返す。
// ※ WrapConflict はパッケージレベル変数として固定ポインタを保持するため errors.Is が機能する。
var errNoDoctorsOnDuty = apperrors.WrapConflict("本日は医師が出勤していないため予約できません")

// checkDoctorSlotConflict は特定医師の時間枠重複をチェックする（SELECT FOR UPDATE）。
func checkDoctorSlotConflict(ctx context.Context, repo repository.ReservationRepository, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) error {
	conflict, err := repo.HasDoctorConflict(ctx, clinicID, doctorID, start, end, excludeID)
	if err != nil {
		return apperrors.Wrap(err, "check doctor slot conflict")
	}
	if conflict {
		return apperrors.WrapConflict("この時間枠は既に予約が入っています")
	}
	return nil
}

// checkCapacitySlotConflict は出勤医師数を上限として時間枠の空き確認をする（SELECT FOR UPDATE）。
// 出勤医師が 0 人の場合は errNoDoctorsOnDuty を返す（LINE パスで RedirectStep を分岐するため）。
func checkCapacitySlotConflict(ctx context.Context, repo repository.ReservationRepository, clinicID uint64, start, end time.Time, excludeID *uint64) error {
	doctorCount, err := repo.CountOnDutyDoctors(ctx, clinicID, start)
	if err != nil {
		return apperrors.Wrap(err, "count on-duty doctors")
	}
	if doctorCount == 0 {
		return errNoDoctorsOnDuty
	}

	conflictCount, err := repo.CountConflicts(ctx, clinicID, start, end, excludeID)
	if err != nil {
		return apperrors.Wrap(err, "count conflicts")
	}
	if conflictCount >= doctorCount {
		return apperrors.WrapConflict("この時間枠は満員です（出勤医師数に達しています）")
	}
	return nil
}

// checkSlotConflict は時間枠の空き・重複をチェックする（SELECT FOR UPDATE）。
//
//   - doctor_id 指定時 → 同一医師の重複のみチェック（別医師は許可）
//   - doctor_id nil 時 → その日の出勤医師数を上限として全予約件数をチェック
//
// excludeID が非 nil の場合、その予約 ID を競合対象から除外する（Update 時の自己競合防止）。
// 競合がある場合は apperrors.ErrConflict ラップエラーを返す。
func checkSlotConflict(ctx context.Context, repo repository.ReservationRepository, clinicID uint64, doctorID *uint64, startTime, endTime time.Time, excludeID *uint64) error {
	if doctorID != nil {
		return checkDoctorSlotConflict(ctx, repo, clinicID, *doctorID, startTime, endTime, excludeID)
	}
	return checkCapacitySlotConflict(ctx, repo, clinicID, startTime, endTime, excludeID)
}

// resolveUpdateParams は現在の予約と更新入力から、競合チェックに使用する時刻・医師 ID を確定する。
// 未指定フィールドは現在値を維持する。DoctorID=0 は NULL（医師未指定）として扱う。
func resolveUpdateParams(current *model.Reservation, input *UpdateReservationInput) (start, end time.Time, doctorID *uint64) {
	start = current.StartTime
	if input.StartTime != nil {
		start = *input.StartTime
	}
	end = current.EndTime
	if input.EndTime != nil {
		end = *input.EndTime
	}
	doctorID = current.DoctorID
	if input.DoctorID != nil {
		if *input.DoctorID == 0 {
			doctorID = nil // 0 は「医師未指定」として NULL 扱い
		} else {
			doctorID = input.DoctorID
		}
	}
	return start, end, doctorID
}

func validateLineReservationCheckedInLink(current *model.Reservation, input *UpdateReservationInput) error {
	if input.Status == nil || *input.Status != model.ReservationStatusCheckedIn {
		return nil
	}
	if current.Source != model.ReservationSourceLine || current.LineCustomerID == nil {
		return nil
	}
	// AUD-001: 0 クリアの意味と揃える（*uint64(0) を non-nil のまま通さない）
	ownerID, petID := resolveFinalOwnerPet(current, input)
	if ownerID == nil || petID == nil {
		return apperrors.WrapInvalidInput("LINE予約を受付済みにする前に飼主とペットの紐付けが必要です")
	}
	return nil
}

// updateWithConflictCheck は SELECT FOR UPDATE + トランザクション内で競合チェック + 予約更新を実行する。
// 時刻・医師変更がある場合にのみ呼び出す。
func (s *reservationService) updateWithConflictCheck(ctx context.Context, clinicID, id uint64, fields map[string]any, input *UpdateReservationInput) (*model.Reservation, error) {
	var result *model.Reservation
	if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		// BE-refactor.md X-9: 空き枠への同時更新（別予約への時刻変更等）もファントムの対象
		// のため、competing な Create と同じ advisory lock で直列化する。
		if err := s.repo.AcquireBookingLock(ctx, clinicID); err != nil {
			return err
		}
		// 現在の予約を行ロックで取得（競合チェックの基準値として使用）
		current, err := s.repo.LockAndFindByID(ctx, clinicID, id)
		if err != nil {
			return err
		}
		finalOwnerID, finalPetID := resolveFinalOwnerPet(current, input)
		if err := validateReservationOwnerPetLinks(ctx, s.repo, clinicID, finalOwnerID, finalPetID); err != nil {
			return err
		}

		resolvedStart, resolvedEnd, resolvedDoctorID := resolveUpdateParams(current, input)
		resolvedReservationTypeID := current.ReservationTypeID
		if input.ReservationTypeID != nil {
			resolvedReservationTypeID = *input.ReservationTypeID
		}

		if input.StartTime != nil || input.EndTime != nil {
			if err := validateTimeRange(resolvedStart, resolvedEnd); err != nil {
				return err
			}
		}

		if err := checkSlotConflict(ctx, s.repo, clinicID, resolvedDoctorID, resolvedStart, resolvedEnd, &id); err != nil {
			return err
		}
		if err := checkReservationTypeCapacity(ctx, s.repo, s.typeRepo, clinicID, resolvedReservationTypeID, resolvedStart, &id); err != nil {
			return err
		}

		updated, err := s.repo.Update(ctx, clinicID, id, fields)
		if err != nil {
			return err
		}
		result = updated
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to update reservation with conflict check", "error", err)
		return nil, apperrors.Wrap(err, "failed to update reservation with conflict check")
	}
	return result, nil
}

func (s *reservationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.Reservation, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	current, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find reservation", "error", err)
		return nil, apperrors.Wrap(err, "failed to find reservation")
	}
	if err := validateLineReservationCheckedInLink(current, input); err != nil {
		return nil, err
	}
	if input.DoctorID != nil || input.ReservationTypeID != nil {
		resolvedDoctorID := current.DoctorID
		if input.DoctorID != nil {
			if *input.DoctorID == 0 {
				resolvedDoctorID = nil
			} else {
				resolvedDoctorID = input.DoctorID
			}
		}
		resolvedReservationTypeID := current.ReservationTypeID
		if input.ReservationTypeID != nil {
			resolvedReservationTypeID = *input.ReservationTypeID
		}
		if err := reservation.ValidateReservationStaffCapability(ctx, s.reservationStaffRepo, clinicID, resolvedDoctorID, resolvedReservationTypeID); err != nil {
			return nil, err
		}
	}
	if input.StartTime != nil || input.EndTime != nil || input.ReservationTypeID != nil {
		resolvedStart, resolvedEnd, _ := resolveUpdateParams(current, input)
		resolvedReservationTypeID := current.ReservationTypeID
		if input.ReservationTypeID != nil {
			resolvedReservationTypeID = *input.ReservationTypeID
		}
		if err := reservation.ValidateReservationTypeAvailableTime(ctx, s.unavailableTimeRepo, clinicID, resolvedReservationTypeID, resolvedStart, resolvedEnd); err != nil {
			return nil, err
		}
	}
	fields := buildReservationUpdate(input)
	// 受付ヘッダー テレメトリ（change-ui.md Phase 2）: checked_in への遷移時点の時刻を記録する。
	// UpdatedAt(autoUpdateTime) は予約編集全般でリセットされ待ち時間算出に流用できないため、
	// 専用カラムへ遷移時刻を都度スタンプする（再受付時は最後の遷移時刻で上書き）。
	if input.Status != nil && *input.Status == model.ReservationStatusCheckedIn && current.Status != model.ReservationStatusCheckedIn {
		fields["checked_in_at"] = time.Now()
	}
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	// 時刻・医師・予約区分の変更がある場合のみ競合チェックが必要
	needsConflictCheck := input.StartTime != nil || input.EndTime != nil || input.DoctorID != nil || input.ReservationTypeID != nil
	needsLinkValidation := input.OwnerID != nil || input.PetID != nil

	var updated *model.Reservation
	switch {
	case needsConflictCheck:
		// 時刻・医師変更あり: SELECT FOR UPDATE + トランザクションで競合を防止（リンク検証も tx 内）
		u, err := s.updateWithConflictCheck(ctx, clinicID, id, fields, input)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update reservation", "error", err)
			return nil, apperrors.Wrap(err, "failed to update reservation")
		}
		updated = u
	case needsLinkValidation:
		// Owner/Pet 変更: 検証と書込みを同一 tx に置く（AUD-001）。
		// 並行更新で古い current を使わないよう、tx 内で行ロックしてから最終状態を検証する。
		var result *model.Reservation
		if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
			locked, err := s.repo.LockAndFindByID(ctx, clinicID, id)
			if err != nil {
				return err
			}
			finalOwnerID, finalPetID := resolveFinalOwnerPet(locked, input)
			if err := validateReservationOwnerPetLinks(ctx, s.repo, clinicID, finalOwnerID, finalPetID); err != nil {
				return err
			}
			u, err := s.repo.Update(ctx, clinicID, id, fields)
			if err != nil {
				return err
			}
			result = u
			return nil
		}); err != nil {
			slog.ErrorContext(ctx, "failed to update reservation", "error", err)
			return nil, apperrors.Wrap(err, "failed to update reservation")
		}
		updated = result
	default:
		// 時刻・医師・リンク変更なし: トランザクション不要
		u, err := s.repo.Update(ctx, clinicID, id, fields)
		if err != nil {
			slog.ErrorContext(ctx, "failed to update reservation", "error", err)
			return nil, apperrors.Wrap(err, "failed to update reservation")
		}
		updated = u
	}

	// Q7: キャンセルされた予約は予約管理から消す（ソフトデリート）。
	// repo.Update 後の FindByID は deleted_at IS NULL のためレコードを返せないので、
	// status=cancelled へ更新（updated 取得）した後に Delete でソフトデリートする。
	if input.Status != nil && *input.Status == model.ReservationStatusCancelled {
		if err := s.repo.Delete(ctx, clinicID, id); err != nil {
			slog.ErrorContext(ctx, "failed to soft-delete cancelled reservation", "error", err)
			return nil, apperrors.Wrap(err, "failed to soft-delete cancelled reservation")
		}
	}

	slog.InfoContext(ctx, "reservation updated",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return updated, nil
}
func (s *reservationService) UpdateReservationRoute(ctx context.Context, clinicID, id uint64, input UpdateReservationRouteInput) (*model.Reservation, error) {
	if input.Route != "" {
		if _, ok := allowedReservationRoutes[input.Route]; !ok {
			return nil, apperrors.WrapInvalidInput(allowedReservationRoutesMessage)
		}
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find reservation")
	}
	var routeValue any
	if input.Route == "" {
		routeValue = nil
	} else {
		routeValue = input.Route
	}
	reservation, err := s.repo.Update(ctx, clinicID, id, map[string]any{colReservationRoute: routeValue})
	if err != nil {
		slog.ErrorContext(ctx, "failed to update reservation_route", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update reservation_route")
	}
	slog.InfoContext(ctx, "reservation_route updated",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.String("route", input.Route))
	return reservation, nil
}

func (s *reservationService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find reservation")
	}
	count, err := s.repo.CountMedicalRecordsByReservationID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check reservation dependencies", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check reservation dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この予約にはカルテが紐付いているため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete reservation", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete reservation")
	}
	slog.InfoContext(ctx, "reservation deleted",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
