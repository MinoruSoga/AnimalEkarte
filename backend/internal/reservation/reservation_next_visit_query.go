package reservation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ---------------------------------------------------------------------------
// EMR-308: next-visit appointment queries（LSTEP 連携 read 面）
//
// LSTEP の next_visit_* タグ同期と tentative リマインダー配信は appointment
// 状態を source of truth にする（旧カルテ推奨日フィールド依存を廃止）。「active」は終端ステータス（cancelled/no_show/completed）と
// 削除済みを除く集合。tentative の意味上の予定日は intended_date、それ以外は
// start_time。日付比較はすべて JST 暦日で行う。
// ---------------------------------------------------------------------------

// nextVisitActiveStatuses は次回来院とみなす非終端ステータス集合。
// checked_in/in_consultation/accounting の当日進行中の予約も「今回の来院」として
// タグ・リマインダー判定に含める（start_time/intended_date の当日以降フィルタにより
// 過去の進行中残りは除外される）。
var nextVisitActiveStatuses = []model.ReservationStatus{
	model.ReservationStatusTentative,
	model.ReservationStatusPending,
	model.ReservationStatusConfirmed,
	model.ReservationStatusCheckedIn,
	model.ReservationStatusInConsultation,
	model.ReservationStatusAccounting,
}

// nextVisitLivingPetClause は appointment のペットが死亡していないことを要求する
// NOT EXISTS 条件。死亡判定は EMR-290/BUG-407 の OR 契約（pets.status='deceased'
// または pets.deceased_at 非NULL）で、両フィールドは独立管理のため OR で評価する。
// 死亡ペットの tentative 行は自動キャンセルされず残り得るため、candidate 抽出・
// 送信直前再検証・タグ同期の3経路すべてで共通適用する。
// pet_id NULL（飼い主直結予約）は pets 行が存在しないので対象に残す一方、
// pet_id 非 NULL は「生存かつ未削除」の pets 行を必須とする — soft-delete 済みの
// 生存ペットをリマインダー・タグ同期の対象から外す。
const nextVisitLivingPetClause = `NOT EXISTS (
	SELECT 1 FROM pets
	WHERE pets.id = appointments.pet_id
	  AND pets.clinic_id = appointments.clinic_id
	  AND (pets.status = 'deceased' OR pets.deceased_at IS NOT NULL)
) AND (
	appointments.pet_id IS NULL
	OR EXISTS (
		SELECT 1 FROM pets live_pet
		WHERE live_pet.id = appointments.pet_id
		  AND live_pet.clinic_id = appointments.clinic_id
		  AND live_pet.deleted_at IS NULL
	)
)`

// A pet's current owner is authoritative for its tentative intent, including
// intents created before a transfer. Real/owner-only appointments retain their
// stored owner; this read capability does not rewrite booking history.
const nextVisitOwnerExpression = `CASE WHEN appointments.status <> 'tentative' OR appointments.pet_id IS NULL THEN appointments.owner_id ELSE (
	SELECT current_pet.owner_id FROM pets current_pet
	WHERE current_pet.id = appointments.pet_id
	  AND current_pet.clinic_id = appointments.clinic_id
	  AND current_pet.deleted_at IS NULL
) END`

// jstCalendarDate は t を含む JST 暦日の開始時刻と YYYY-MM-DD 文字列を返す。
// intended_date（date 列）との比較は暦日文字列経由で行い、セッション TZ の
// 影響を排除する。
func jstCalendarDate(t time.Time) (dayStart time.Time, dateOnly string) {
	jst := t.In(config.JST)
	dayStart = time.Date(jst.Year(), jst.Month(), jst.Day(), 0, 0, 0, 0, config.JST)
	return dayStart, dayStart.Format(time.DateOnly)
}

// FindNextAppointmentForOwner は飼い主の次回来院に相当する active appointment を返す。
// 意味上の日付が本日（JST）以降のものに限り、tentative は intended_date・それ以外は
// start_time の JST 暦日が早い順、同日は start_time・id 順で確定的に1件選ぶ。
func (r *reservationRepository) FindNextAppointmentForOwner(
	ctx context.Context,
	clinicID, ownerID uint64,
) (*model.Reservation, error) {
	todayStart, todayDate := jstCalendarDate(time.Now())
	var appointment model.Reservation
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("("+nextVisitOwnerExpression+") = ?", ownerID).
		Where("status IN ?", nextVisitActiveStatuses).
		Where(nextVisitLivingPetClause).
		Where(
			"(status = ? AND intended_date IS NOT NULL AND intended_date >= ?::date) "+
				"OR (status <> ? AND start_time >= ?)",
			model.ReservationStatusTentative, todayDate,
			model.ReservationStatusTentative, todayStart,
		).
		// tentative の JST 暦日は intended_date、それ以外は start_time の JST 暦日で比較する。
		// timestamptz の AT TIME ZONE 'Asia/Tokyo' は暦日部分の取得に使いセッション TZ に依らない。
		Order("CASE WHEN status = '" + string(model.ReservationStatusTentative) +
			"' THEN intended_date ELSE (start_time AT TIME ZONE 'Asia/Tokyo')::date END ASC").
		Order("start_time ASC").
		Order("id ASC").
		Take(&appointment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("clinic=%d owner=%d next", clinicID, ownerID))
	}
	return &appointment, nil
}

// FindOwnerIDsWithTentativeIntendedDate は intended_date が JST 暦日で targetDate に
// 一致する active tentative を持つ飼い主 ID を重複なく返す（3日前・当日リマインダーの
// candidate 抽出用）。同じ飼い主に複数 tentative が紐づく場合でも1件だけ返す。
func (r *reservationRepository) FindOwnerIDsWithTentativeIntendedDate(
	ctx context.Context,
	clinicID uint64,
	targetDate time.Time,
) ([]uint64, error) {
	_, dateOnly := jstCalendarDate(targetDate)
	ownerIDs := make([]uint64, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Select("DISTINCT "+nextVisitOwnerExpression+" AS owner_id").
		Where("clinic_id = ? AND status = ? AND deleted_at IS NULL", clinicID, model.ReservationStatusTentative).
		Where("("+nextVisitOwnerExpression+") IS NOT NULL").
		Where("intended_date = ?::date", dateOnly).
		Where(nextVisitLivingPetClause).
		Scan(&ownerIDs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("clinic=%d tentative intended=%s", clinicID, dateOnly))
	}
	return ownerIDs, nil
}

// FindTentativeByOwnerAndIntendedDate は飼い主の active tentative のうち intended_date が
// JST 暦日で targetDate に一致する1件を返す（配信直前の stale 再検証用）。
// ステータス遷移済み・削除済み・日付変更済みで一致しなければ (nil, nil)。
func (r *reservationRepository) FindTentativeByOwnerAndIntendedDate(
	ctx context.Context,
	clinicID, ownerID uint64,
	targetDate time.Time,
) (*model.Reservation, error) {
	_, dateOnly := jstCalendarDate(targetDate)
	var appointment model.Reservation
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND status = ? AND deleted_at IS NULL AND intended_date = ?::date",
			clinicID, model.ReservationStatusTentative, dateOnly).
		Where("("+nextVisitOwnerExpression+") = ?", ownerID).
		Where(nextVisitLivingPetClause).
		Order("id DESC").
		Take(&appointment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("clinic=%d owner=%d tentative intended=%s", clinicID, ownerID, dateOnly))
	}
	return &appointment, nil
}

// FindTentativesByOwnerAndIntendedDate は同条件の active tentative を全件 id 昇順で返す。
// Codex #22: 同一 owner の同日 tentative が複数ペットに跨る場合、送信直前の再検証は
// candidate 抽出時点の全ペットを直列化ロック対象として列挙する必要がある。
func (r *reservationRepository) FindTentativesByOwnerAndIntendedDate(
	ctx context.Context,
	clinicID, ownerID uint64,
	targetDate time.Time,
) ([]model.Reservation, error) {
	_, dateOnly := jstCalendarDate(targetDate)
	appointments := make([]model.Reservation, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND status = ? AND deleted_at IS NULL AND intended_date = ?::date",
			clinicID, model.ReservationStatusTentative, dateOnly).
		Where("("+nextVisitOwnerExpression+") = ?", ownerID).
		Where(nextVisitLivingPetClause).
		Order("id ASC").
		Find(&appointments).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("clinic=%d owner=%d tentatives intended=%s", clinicID, ownerID, dateOnly))
	}
	return appointments, nil
}
