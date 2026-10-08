package reservation

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *reservationService) updateWithConflictCheck(ctx context.Context, clinicID, id uint64, fields map[string]any, input *UpdateReservationInput) (*model.Reservation, error) {
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("reservation transaction dependency is required")
	}
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
		if err := ValidateReservationOwnerPetLinksWithRepo(ctx, s.repo, clinicID, finalOwnerID, finalPetID); err != nil {
			return err
		}
		// #261 P0: 死亡拒否は pet 付け替え/新規紐付け時のみ（既存死亡ペット予約の時刻変更等は許可）。
		if input.PetID != nil && *input.PetID != 0 {
			if err := ValidateReservationPetNotDeceased(ctx, s.repo, clinicID, input.PetID); err != nil {
				return err
			}
		}
		if err := validateInConsultationHasMedicalRecord(ctx, s.repo, clinicID, id, current, input); err != nil {
			return err
		}

		resolvedStart, resolvedEnd, resolvedDoctorID := resolveUpdateParams(current, input)
		resolvedReservationTypeID := current.ReservationTypeID
		if input.ReservationTypeID != nil {
			resolvedReservationTypeID = *input.ReservationTypeID
		}
		if input.DoctorID != nil || input.ReservationTypeID != nil {
			if err := ValidateReservationStaffCapability(ctx, s.reservationStaffRepo, clinicID, resolvedDoctorID, resolvedReservationTypeID); err != nil {
				return err
			}
		}

		if input.StartTime != nil || input.EndTime != nil {
			if err := validateTimeRange(resolvedStart, resolvedEnd); err != nil {
				return err
			}
		}

		startOrEndChanged := !resolvedStart.Equal(current.StartTime) || !resolvedEnd.Equal(current.EndTime)
		if startOrEndChanged &&
			shouldEnforceClosedDayConstraintOnUpdate(current.Status, current.ReservationRoute) {
			// Legacy test/compatibility constructors may omit LINE settings; production
			// composition injects them and update must enforce the same closed days as create.
			if err := s.validateCreateClosedDaysIfConfigured(ctx, clinicID, resolvedStart); err != nil {
				return err
			}
			if err := validateClinicHoliday(ctx, s.holidayFinder, clinicID, resolvedStart); err != nil {
				return err
			}
		}

		if err := CheckSlotConflict(ctx, s.repo, clinicID, resolvedDoctorID, resolvedStart, resolvedEnd, &id); err != nil {
			return err
		}
		if err := CheckReservationTypeCapacity(ctx, s.repo, s.typeRepo, clinicID, resolvedReservationTypeID, resolvedStart, &id); err != nil {
			return err
		}

		updated, err := s.repo.update(ctx, clinicID, id, fields)
		if err != nil {
			return err
		}
		result = updated
		return nil
	}); err != nil {
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
		return nil, apperrors.Wrap(err, "failed to find reservation")
	}
	if current == nil {
		return nil, apperrors.WrapNotFound("reservation", strconv.FormatUint(id, 10))
	}
	// EMR-308: tentative（次回予定）は sentinel 時刻（JST 00:00-00:15）を持つ日付のみの
	// 意図行で、booking 検証（時間枠・容量・休診日）を前提としない。汎用 PATCH で
	// pending/confirmed 等へ昇格させるとその検証を通らず sentinel のまま実予約へ混入する
	// ため、許可する遷移は checked_in（当日受付。sentinel は実受付時刻へ置換）と
	// cancelled（終了）のみ。確定は POST confirm-tentative 専用経路が担当し、status 以外
	// のフィールド更新も tentative 行には適用しない。
	if current.Status == model.ReservationStatusTentative {
		if input.Status == nil {
			return nil, apperrors.WrapInvalidInput("仮予約（次回予定）の内容変更はできません。次回予定の設定・確定・終了の操作を使用してください")
		}
		switch *input.Status {
		case model.ReservationStatusCheckedIn, model.ReservationStatusCancelled:
		default:
			return nil, apperrors.WrapInvalidInput("仮予約（次回予定）は予約確定操作でのみ本予約へ変更できます")
		}
		// tentative 遷移は status のみ許可 — owner/pet/区分/schedule の同時書換は
		// 下流の link・区分・capability・競合検証を経由せず CAS に流れ込むため拒否する。
		if input.StartTime != nil || input.EndTime != nil || input.OwnerID != nil ||
			input.PetID != nil || input.VisitType != nil || input.ReservationTypeID != nil ||
			input.DoctorID != nil || input.IsDesignated != nil || input.Notes != nil {
			return nil, apperrors.WrapInvalidInput("仮予約（次回予定）の受付・終了では status 以外のフィールドを変更できません")
		}
	}
	if err := validateLineReservationCheckedInLink(current, input); err != nil {
		return nil, err
	}
	resolvedStart, resolvedEnd, _ := resolveUpdateParams(current, input)
	if input.StartTime != nil || input.EndTime != nil {
		if err := validateTimeRange(resolvedStart, resolvedEnd); err != nil {
			return nil, err
		}
	}
	needsConflictCheck := shouldReevaluateReservationBookingConstraintsOnUpdate(current, input)
	if needsConflictCheck {
		resolvedReservationTypeID := resolveUpdateReservationTypeID(current, input)
		if err := ValidateReservationTypeAvailableTime(ctx, s.unavailableTimeRepo, clinicID, resolvedReservationTypeID, resolvedStart, resolvedEnd); err != nil {
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
	// EMR-308: tentative（record_intent）行の当日受付は sentinel 時刻（JST 00:00-00:15）を
	// 実受付時刻に置き換え、同じ行を checked_in へ昇格する。general placeholder 区分・
	// intended_date・source_medical_record_id の provenance は維持し新規行は作らない。
	// tentative PATCH は status-only に制限済み（上記ゲート）のため input の時刻は常に nil。
	// status-only PATCH では schedule 不変扱いのため枠競合・休診日チェックは発火しない。
	if input.Status != nil && *input.Status == model.ReservationStatusCheckedIn &&
		current.Status == model.ReservationStatusTentative {
		now := time.Now()
		fields["start_time"] = now
		fields["end_time"] = now.Add(nextVisitIntentDurationMinutes * time.Minute)
	}
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	// EMR-308: tentative の終端遷移（checked_in/cancelled）は CAS 専用経路へ。
	// 冒頭の FindByID はロックを取らないため、汎用 repo.update（status 述語なし）で
	// 書くと、並行 ConfirmTentative が先に pending/confirmed + 実枠時刻で commit
	// した行を checked_in/cancelled+論理削除で上書きし得る（TOCTOU）。
	if current.Status == model.ReservationStatusTentative {
		return s.updateTentativeStatusTransition(ctx, clinicID, id, current, input, fields)
	}

	needsLinkValidation := input.OwnerID != nil || input.PetID != nil
	needsCapabilityCheck := reservationStaffCapabilityRequiresRevalidation(current, input)
	needsKarteCheck := isTransitioningToInConsultation(current, input)
	isCancel := input.Status != nil && *input.Status == model.ReservationStatusCancelled

	// EMR-74: 汎用 Update は trimming 予約の in_consultation 遷移を InvalidInput で拒否する。
	// FindByID は ReservationType を clinic スコープで preload 済みのため current で判定できる。
	// tx 内のカルテ存在チェックより先に評価し、cancel/conflict-check 両経路をカバーする。
	if err := validateGenericUpdateRejectsTrimmingInConsultation(current, input); err != nil {
		return nil, err
	}

	// RSV-06 / X-06: cancel = status update + soft delete as one business graph.
	// Q7: soft-delete after status=cancelled so FindByID (deleted_at IS NULL) can still return
	// the cancelled row from the update path before Delete; both must share one transaction.
	if isCancel {
		if s.tx == nil {
			return nil, apperrors.WrapInternalServerError("reservation transaction dependency is required")
		}
		var updated *model.Reservation
		if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
			u, err := s.applyReservationUpdate(txCtx, clinicID, id, fields, input, needsConflictCheck, needsLinkValidation, needsCapabilityCheck, needsKarteCheck)
			if err != nil {
				return err
			}
			if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
				return apperrors.Wrap(err, "failed to soft-delete cancelled reservation")
			}
			updated = u
			return nil
		}); err != nil {
			return nil, apperrors.Wrap(err, "failed to cancel reservation")
		}
		// EMR-308: cancel（status 更新+soft delete）で次回来院が消えるため新旧 owner 双方を
		// post-commit 再同期する（所有者付け替え時は旧側のタグも取り残さない）。
		syncNextVisitTagsBestEffort(ctx, s.nextVisitTagSync, clinicID, current.OwnerID, updated.OwnerID)
		slog.InfoContext(ctx, "reservation updated",
			slog.Uint64("reservation_id", id),
			slog.Uint64("clinic_id", clinicID))
		return updated, nil
	}

	updated, err := s.applyReservationUpdate(ctx, clinicID, id, fields, input, needsConflictCheck, needsLinkValidation, needsCapabilityCheck, needsKarteCheck)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update reservation")
	}

	// EMR-308: 日時・ステータス・所有者変更で次回来院が変わり得るため post-commit で
	// 新旧 owner を再同期する（同一 owner は helper 内で重複排除される）。
	syncNextVisitTagsBestEffort(ctx, s.nextVisitTagSync, clinicID, current.OwnerID, updated.OwnerID)
	slog.InfoContext(ctx, "reservation updated",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return updated, nil
}

// updateTentativeStatusTransition は tentative 行の終端遷移を status CAS で書く。
// 呼び出し時点で input.Status は checked_in/cancelled に絞られている（Update 冒頭の
// tentative ゲート）。
// checked_in: 実来院への昇格なので ConfirmTentative/Create と同じ死亡ペット境界を
// 通す（#261。status-only PATCH は付け替え時の死亡チェックを通らないためここで拒否）。
// cancelled: 汎用 cancel と同じ status update + soft delete の graph を tx で取る。
func (s *reservationService) updateTentativeStatusTransition(
	ctx context.Context,
	clinicID, id uint64,
	current *model.Reservation,
	input *UpdateReservationInput,
	fields map[string]any,
) (*model.Reservation, error) {
	if *input.Status == model.ReservationStatusCheckedIn {
		if err := ValidateReservationPetNotDeceased(ctx, s.repo, clinicID, current.PetID); err != nil {
			return nil, err
		}
		updated, err := s.repo.updateTentativeStatus(ctx, clinicID, id, fields)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to check in tentative reservation")
		}
		syncNextVisitTagsBestEffort(ctx, s.nextVisitTagSync, clinicID, current.OwnerID, updated.OwnerID)
		return updated, nil
	}
	// cancelled: status=cancelled を刻んでから論理削除（RSV-06/Q7 と同じ graph）。
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("reservation transaction dependency is required")
	}
	var updated *model.Reservation
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		u, err := s.repo.updateTentativeStatus(txCtx, clinicID, id, fields)
		if err != nil {
			return err
		}
		if err := s.repo.Delete(txCtx, clinicID, id); err != nil {
			return apperrors.Wrap(err, "failed to soft-delete cancelled tentative")
		}
		updated = u
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to cancel tentative reservation")
	}
	syncNextVisitTagsBestEffort(ctx, s.nextVisitTagSync, clinicID, current.OwnerID, updated.OwnerID)
	return updated, nil
}

// applyReservationUpdate applies field updates with the existing conflict/link branching.
// Caller supplies ambient transaction when atomic multi-write is required (e.g. cancel).
func (s *reservationService) applyReservationUpdate(
	ctx context.Context,
	clinicID, id uint64,
	fields map[string]any,
	input *UpdateReservationInput,
	needsConflictCheck, needsLinkValidation, needsCapabilityCheck, needsKarteCheck bool,
) (*model.Reservation, error) {
	switch {
	case needsConflictCheck:
		// 時刻・医師変更あり: SELECT FOR UPDATE + トランザクションで競合を防止（リンク検証も tx 内）
		// Capability already runs inside updateWithConflictCheck; do not divert this path.
		return s.updateWithConflictCheck(ctx, clinicID, id, fields, input)
	case needsLinkValidation || needsCapabilityCheck || needsKarteCheck:
		// Owner/Pet・担当者/区分・in_consultation カルテ確認: 行ロック後に検証し同一 tx で書く。
		// Capability SHARE locks must be held through commit (ReservationStaffWriteGuard).
		// Status-only in_consultation does not take AcquireBookingLock.
		if s.tx == nil {
			return nil, apperrors.WrapInternalServerError("reservation transaction dependency is required")
		}
		var result *model.Reservation
		if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
			locked, err := s.repo.LockAndFindByID(ctx, clinicID, id)
			if err != nil {
				return err
			}
			if needsLinkValidation {
				finalOwnerID, finalPetID := resolveFinalOwnerPet(locked, input)
				if err := ValidateReservationOwnerPetLinksWithRepo(ctx, s.repo, clinicID, finalOwnerID, finalPetID); err != nil {
					return err
				}
				// #261 P0: pet 付け替え/新規紐付け時のみ死亡拒否（既存関連のままの更新は許可）。
				if err := validateReservationPetNotDeceasedOnRelink(ctx, s.repo, clinicID, input.PetID); err != nil {
					return err
				}
			}
			if needsCapabilityCheck {
				_, _, resolvedDoctorID := resolveUpdateParams(locked, input)
				resolvedReservationTypeID := resolveUpdateReservationTypeID(locked, input)
				if err := ValidateReservationStaffCapability(ctx, s.reservationStaffRepo, clinicID, resolvedDoctorID, resolvedReservationTypeID); err != nil {
					return err
				}
			}
			if err := validateInConsultationHasMedicalRecord(ctx, s.repo, clinicID, id, locked, input); err != nil {
				return err
			}
			u, err := s.repo.update(ctx, clinicID, id, fields)
			if err != nil {
				return err
			}
			result = u
			return nil
		}); err != nil {
			return nil, err
		}
		return result, nil
	default:
		return s.repo.update(ctx, clinicID, id, fields)
	}
}
func (s *reservationService) UpdateReservationRoute(ctx context.Context, clinicID, id uint64, input UpdateReservationRouteInput) (*model.Reservation, error) {
	// EMR-308: record_intent は tentative（次回予定）の provenance 専用値で、経路更新 API
	// からの設定も除去も許可しない（provenance を越境させない）。
	if input.Route == nextVisitIntentRoute {
		return nil, apperrors.WrapInvalidInput("record_intent は次回予定専用の予約経路のため指定できません")
	}
	if input.Route != "" {
		if _, ok := AllowedReservationRoutes[input.Route]; !ok {
			return nil, apperrors.WrapInvalidInput(AllowedReservationRoutesMessage)
		}
	}
	current, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find reservation")
	}
	if current != nil && current.ReservationRoute != nil && *current.ReservationRoute == nextVisitIntentRoute {
		return nil, apperrors.WrapInvalidInput("次回予定由来の予約経路（record_intent）は変更できません")
	}
	var routeValue any
	if input.Route == "" {
		routeValue = nil
	} else {
		routeValue = input.Route
	}
	reservation, err := s.repo.update(ctx, clinicID, id, map[string]any{colReservationRoute: routeValue})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update reservation_route")
	}
	slog.InfoContext(ctx, "reservation_route updated",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.String("route", input.Route))
	return reservation, nil
}

func (s *reservationService) Delete(ctx context.Context, clinicID, id uint64) error {
	// EMR-308: 削除後は owner を引けないため、post-commit の next_visit タグ除去
	// 同期のために削除前の予約（owner_id）を保持しておく。
	existing, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to find reservation")
	}
	count, err := s.repo.CountMedicalRecordsByReservationID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check reservation dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この予約にはカルテが紐付いているため削除できません")
	}
	// EMR-308: tentative の直接削除も終了 semantics を揃える — cancelled を刻んでから
	// 論理削除（ClearNextVisitIntent/汎用 cancel と同一の終端）。status CAS で
	// 並行昇格（confirm-tentative）を fail-closed に弾く。
	if existing != nil && existing.Status == model.ReservationStatusTentative {
		if s.tx == nil {
			return apperrors.WrapInternalServerError("reservation transaction dependency is required")
		}
		if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
			if _, err := s.repo.updateTentativeStatus(txCtx, clinicID, id, map[string]any{
				"status": model.ReservationStatusCancelled,
			}); err != nil {
				return err
			}
			return s.repo.Delete(txCtx, clinicID, id)
		}); err != nil {
			return apperrors.Wrap(err, "failed to delete tentative reservation")
		}
		syncNextVisitTagBestEffort(ctx, s.nextVisitTagSync, clinicID, existing.OwnerID)
		slog.InfoContext(ctx, "tentative reservation deleted",
			slog.Uint64("reservation_id", id),
			slog.Uint64("clinic_id", clinicID))
		return nil
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete reservation")
	}
	if existing != nil {
		syncNextVisitTagBestEffort(ctx, s.nextVisitTagSync, clinicID, existing.OwnerID)
	}
	slog.InfoContext(ctx, "reservation deleted",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
