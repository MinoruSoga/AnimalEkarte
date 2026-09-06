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
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	needsLinkValidation := input.OwnerID != nil || input.PetID != nil
	needsCapabilityCheck := reservationStaffCapabilityRequiresRevalidation(current, input)
	needsKarteCheck := isTransitioningToInConsultation(current, input)
	isCancel := input.Status != nil && *input.Status == model.ReservationStatusCancelled

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
		slog.InfoContext(ctx, "reservation updated",
			slog.Uint64("reservation_id", id),
			slog.Uint64("clinic_id", clinicID))
		return updated, nil
	}

	updated, err := s.applyReservationUpdate(ctx, clinicID, id, fields, input, needsConflictCheck, needsLinkValidation, needsCapabilityCheck, needsKarteCheck)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update reservation")
	}

	slog.InfoContext(ctx, "reservation updated",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
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
	if input.Route != "" {
		if _, ok := AllowedReservationRoutes[input.Route]; !ok {
			return nil, apperrors.WrapInvalidInput(AllowedReservationRoutesMessage)
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
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find reservation")
	}
	count, err := s.repo.CountMedicalRecordsByReservationID(ctx, clinicID, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check reservation dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この予約にはカルテが紐付いているため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete reservation")
	}
	slog.InfoContext(ctx, "reservation deleted",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
