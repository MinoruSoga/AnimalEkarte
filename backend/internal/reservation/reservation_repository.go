package reservation

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// staffAssignedToClinicsCond は Preload した staff（Doctor / CreatedByStaff）を、指定クリニック集合の
// いずれかに現在所属している場合のみ表示する条件。staff は staff_clinic_assignments による多医院所属のため
// staffs.clinic_id（主所属）単純スコープでは共有スタッフを誤って隠す。assignment-EXISTS で多医院所属を
// 尊重しつつ、別テナント単独所属スタッフ名の漏洩を防ぐ。予約は現在/未来データのため履歴表示の回帰はない。
const staffAssignedToClinicsCond = "deleted_at IS NULL AND EXISTS (SELECT 1 FROM staff_clinic_assignments sca WHERE sca.staff_id = staffs.id AND sca.clinic_id IN ? AND sca.deleted_at IS NULL)"

// reservationRelationsMatchParentClinic は、各 appointment 自身の clinic_id と関連行の
// clinic_id を相関させる。soft-delete 済みの同一 clinic 関連と過去の staff assignment は予約履歴の
// 親行/countを維持するため許容し、現在の関連を応答へ表示するかは Preload 側の条件に委ねる。
// cross-clinic FK と Owner/Pet 不一致だけは、一覧/単件のどちらでも親行ごと fail-closed にする。
func reservationRelationsMatchParentClinic(q *gorm.DB) *gorm.DB {
	return q.Where(`
		EXISTS (
			SELECT 1
			FROM reservation_types scoped_reservation_type
			WHERE scoped_reservation_type.id = appointments.reservation_type_id
			  AND scoped_reservation_type.clinic_id = appointments.clinic_id
			  AND (
				scoped_reservation_type.group_id IS NULL
				OR EXISTS (
					SELECT 1
					FROM reservation_type_groups scoped_reservation_type_group
					WHERE scoped_reservation_type_group.id = scoped_reservation_type.group_id
					  AND scoped_reservation_type_group.clinic_id = appointments.clinic_id
				)
			  )
		)
		AND (
			appointments.owner_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM owners scoped_owner
				WHERE scoped_owner.id = appointments.owner_id
				  AND scoped_owner.clinic_id = appointments.clinic_id
			)
		)
		AND (
			appointments.pet_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM pets scoped_pet
				WHERE scoped_pet.id = appointments.pet_id
				  AND scoped_pet.clinic_id = appointments.clinic_id
				  AND (
					appointments.owner_id IS NULL
					OR scoped_pet.owner_id = appointments.owner_id
				  )
				  AND EXISTS (
					SELECT 1
					FROM owners scoped_pet_owner
					WHERE scoped_pet_owner.id = scoped_pet.owner_id
					  AND scoped_pet_owner.clinic_id = appointments.clinic_id
				  )
			)
		)
		AND (
			appointments.line_customer_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM line_customers scoped_line_customer
				WHERE scoped_line_customer.id = appointments.line_customer_id
				  AND scoped_line_customer.clinic_id = appointments.clinic_id
			)
		)
		AND (
			appointments.doctor_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM staffs scoped_doctor
				JOIN staff_clinic_assignments scoped_doctor_assignment
				  ON scoped_doctor_assignment.staff_id = scoped_doctor.id
				 AND scoped_doctor_assignment.clinic_id = appointments.clinic_id
				WHERE scoped_doctor.id = appointments.doctor_id
			)
		)
		AND (
			appointments.created_by IS NULL
			OR EXISTS (
				SELECT 1
				FROM staffs scoped_creator
				JOIN staff_clinic_assignments scoped_creator_assignment
				  ON scoped_creator_assignment.staff_id = scoped_creator.id
				 AND scoped_creator_assignment.clinic_id = appointments.clinic_id
				WHERE scoped_creator.id = appointments.created_by
			)
		)
	`)
}

// ReservationCRUDRepository は owner package 内のコア persistence 操作。
// package 外の consumer はこの interface ではなく、必要な read operation と
// ReservationIntentRepository の一部だけを consumer-side interface として宣言する。
type ReservationCRUDRepository interface {
	// FindAll は指定した複数医院 (#86 拠点横断) の予約を検索する。clinicIDs はハンドラ層で所属検証済みであること。
	FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	// FindByIDForClinics は複数医院スコープで予約を1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error)
	Create(ctx context.Context, reservation *model.Reservation) error
	// update は owner package 内だけで使う汎用 persistence primitive。
	// package 外には ReservationIntentRepository の intent-specific operation だけを公開する。
	update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

// ReservationIntentRepository は reservation 外の consumer が使う appointment write operation。
// 各 consumer はこのinterface全体ではなく、必要なメソッドだけをローカルinterfaceへ宣言する。
type ReservationIntentRepository interface {
	CompleteForAccounting(ctx context.Context, clinicID uint64, medicalRecordID, ownerID, petID *uint64, scheduledDate time.Time) (int64, error)
	AssertMedicalRecordDoctorInClinic(ctx context.Context, clinicID, doctorID uint64) error
	BackfillForMedicalRecord(ctx context.Context, clinicID, id uint64, ownerID, petID, doctorID *uint64) error
	PrepareForMedicalRecordFinalization(ctx context.Context, clinicID, id uint64) error
	MarkNoShow(ctx context.Context, clinicID, id uint64) (NoShowTransition, error)
	FindTrimmingByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	LockTrimmingByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	CreateForTrimming(ctx context.Context, clinicID uint64, input CreateTrimmingReservationInput) (*model.Reservation, error)
	UpdateForTrimming(ctx context.Context, clinicID, id uint64, input UpdateTrimmingReservationInput) (*model.Reservation, error)
	DeleteForTrimming(ctx context.Context, clinicID, id uint64) error
}

// NoShowTransition reports whether MarkNoShow performed the compare-and-set and, when it did,
// the exact previous status needed for a transaction-local audit record.
type NoShowTransition struct {
	Changed        bool
	PreviousStatus model.ReservationStatus
}

// ReservationNoShowAtRepository is the narrow durable-scheduler capability.
// It is intentionally not embedded in the broad legacy interfaces so existing
// consumers do not acquire methods they do not use.
type ReservationNoShowAtRepository interface {
	FindNoShowCandidatesAt(
		ctx context.Context,
		clinicID uint64,
		evaluatedAt time.Time,
	) ([]model.Reservation, error)
	MarkNoShowAt(
		ctx context.Context,
		clinicID uint64,
		id uint64,
		evaluatedAt time.Time,
	) (NoShowTransition, error)
}

// ReservationSlotRepository はトランザクション内の競合チェック操作（5 メソッド）。
// dbOrTx でコンテキストの tx を自動使用。reservation_service で使用。
type ReservationSlotRepository interface {
	// AcquireBookingLock は clinic 単位の pg_advisory_xact_lock を取得する（BE-refactor.md X-9）。
	// CountConflicts/CountByTypeAndStartTime の SELECT FOR UPDATE は条件に合致する既存行が
	// 0 件（空枠）の場合は何もロックしないため、空き枠への同時予約がファントムで両方成功しうる。
	// WithTx トランザクションの先頭でこの advisory lock を取得することで、同一 clinic の
	// 予約競合チェック→INSERT を直列化する。トランザクション終了時に自動解放される
	// （pg_advisory_xact_lock はセッションではなくトランザクションスコープ）。
	// 【不変条件・デッドロック防止】appointments 行に対する行ロック（LockAndFindByID/
	// HasDoctorConflict/CountConflicts 等の SELECT FOR UPDATE）を取得する前に、同一
	// トランザクション内で必ず本メソッドを先頭で呼ぶこと。逆順（行ロック取得後に advisory
	// lock を取得）を許すと、2つのトランザクションが互いの advisory lock/行ロックを待ち合う
	// AB-BA デッドロックが理論上成立しうる（呼び出し元: reservation_service Create/
	// updateWithConflictCheck、reservation_validators ValidateAndCreate、
	// appointment_admin_service Create — いずれもこの順序を守ること）。
	AcquireBookingLock(ctx context.Context, clinicID uint64) error
	// LockAndFindByID は FOR UPDATE で予約を行ロック取得する（updateWithConflictCheck 用）。
	LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	// HasDoctorConflict は指定医師の時間枠重複を SELECT FOR UPDATE でチェックする。
	HasDoctorConflict(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error)
	// CountOnDutyDoctors は当日の出勤医師数を返す。
	CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error)
	// CountConflicts は時間枠の競合予約数を SELECT FOR UPDATE で返す。
	CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error)
	// CountByTypeAndStartTime は同一予約区分・同一開始時刻の予約件数を返す。
	CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error)
}

// ReservationQueryRepository はクロスフィーチャーのクエリ・依存チェック（10 メソッド）。
// reservation_type_service / staff_service / liff_service / reservation_validators で使用。
type ReservationQueryRepository interface {
	ExistsByReservationTypeID(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error)
	ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error)
	// CountMedicalRecordsByReservationID は親 appointments.clinic_id でスコープした有効カルテ件数を返す
	// （medical_records.clinic_id はフィルタしない — 参照が存在する限り削除/identity 変更ガードを fail-closed に保つ / SEC-SWEEP-02-RES-B1）。
	// 呼び出し元は UpdateForTrimming・DeleteForTrimming・reservationService.Delete の3箇所。
	// 依存チェックと同じ ambient transaction へ参加する。
	CountMedicalRecordsByReservationID(ctx context.Context, clinicID, reservationID uint64) (int64, error)
	// CountByCustomerAndDateRange は顧客・期間での予約件数を返す（日次・月次制限チェック用）。
	CountByCustomerAndDateRange(ctx context.Context, clinicID, customerID uint64, start, end time.Time) (int64, error)
	// CountByDateAndSource は日付・ソースの予約件数を返す（確認番号生成用）。
	CountByDateAndSource(ctx context.Context, clinicID uint64, date time.Time, source model.ReservationSource) (int64, error)
	// FindAllByCategory はカテゴリ（'general'/'trimming'）でフィルタした予約一覧を返す。
	// トリミング管理APIが appointments ベースで動作するために使用（BE-119）。
	FindAllByCategory(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error)
	// FindNoShowCandidates は終了から4時間以上経過した confirmed/pending 予約のうち、
	// 確定済みカルテが存在しないものを返す（BE-014 ノーショウ検知用）。
	FindNoShowCandidates(ctx context.Context, clinicID uint64) ([]model.Reservation, error)
	// AssertOwnerInClinic は owner が clinic に属することを検証する（AUD-001、dbOrTx 参加）。
	AssertOwnerInClinic(ctx context.Context, clinicID, ownerID uint64) error
	// FindPetOwnerInClinic は pet が clinic に属する場合にその OwnerID を返す（AUD-001、dbOrTx 参加）。
	FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error)
	// FindPetByIDInClinic は clinic スコープでペットを読む（SD-10 死亡 write ガード用、dbOrTx 参加）。
	// 部分列（id / owner_id / deceased_at / status）のみを返す。
	FindPetByIDInClinic(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
	// AssertLineCustomerInClinic は line_customer が clinic に属することを検証する（AUD-001、dbOrTx 参加）。
	AssertLineCustomerInClinic(ctx context.Context, clinicID, lineCustomerID uint64) error
}

// StaffAssignmentUsageRepository exposes the batch dependency lookup needed
// when the staff owner removes clinic assignments.
type StaffAssignmentUsageRepository interface {
	FindClinicIDsByStaffID(ctx context.Context, clinicIDs []uint64, staffID uint64) ([]uint64, error)
}

// ReservationRepository は owner package 内の3つのrepository capabilityを合成する。
// 汎用 update は非公開のため、package外consumerは実装も呼び出しもできない。
type ReservationRepository interface {
	ReservationCRUDRepository
	ReservationSlotRepository
	ReservationQueryRepository
}

// ReservationStore は composition root が保持する具象実装の公開method set。
// consumer はこの広い型を引数にせず、必要なreadとReservationIntentRepositoryの一部だけを
// ローカルinterfaceとして宣言する。
type ReservationStore interface {
	ReservationRepository
	ReservationIntentRepository
	StaffAssignmentUsageRepository
}

type reservationRepository struct {
	db *gorm.DB
}

// NewReservationRepository は owner内repositoryとcross-domain intentを実装するstoreを返す。
func NewReservationRepository(db *gorm.DB) ReservationStore {
	return &reservationRepository{db: db}
}

func (r *reservationRepository) FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	reservations := make([]model.Reservation, 0)
	var total int64

	// フェイルセーフ: 検証バグ等で空スライスが渡っても全件露出させない
	if len(clinicIDs) == 0 {
		return reservations, 0, nil
	}

	q := persistence.DBOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Scopes(persistence.ClinicScopeIn(clinicIDs), reservationRelationsMatchParentClinic)
	switch {
	case date != nil:
		// 単日フィルタ（当日受付など）
		start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		end := start.Add(24 * time.Hour)
		q = q.Where("start_time >= ? AND start_time < ?", start, end)
	case startDate != nil && endDate != nil:
		// 期間レンジフィルタ（予約管理カレンダーの表示中の週/月）。endDate は排他的上限
		q = q.Where("start_time >= ? AND start_time < ?", *startDate, *endDate)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if source != nil {
		q = q.Where("source = ?", *source)
	}
	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "reservation", "")
	}
	if err := reservationListPreloads(q, clinicIDs, false).
		Scopes(persistence.Paginate(page, limit)).Order("start_time ASC").Find(&reservations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "reservation", "")
	}
	return reservations, total, nil
}

func (r *reservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	return r.findReservationByID(ctx, []uint64{clinicID}, id)
}

func (r *reservationRepository) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error) {
	return r.findReservationByID(ctx, clinicIDs, id)
}

// findReservationByID は認可済みクリニック集合を受け取り予約を1件取得する共通実装。
// Preload する診療区分マスタも同じ集合で clinic 隔離する（別クリニックの診療区分混入防止）。
func (r *reservationRepository) findReservationByID(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error) {
	if len(clinicIDs) == 0 {
		return nil, apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	var reservation model.Reservation
	err := reservationListPreloads(persistence.DBOrTx(ctx, r.db), clinicIDs, true).
		Scopes(persistence.ClinicScopeIn(clinicIDs), reservationRelationsMatchParentClinic).
		Where("id = ?", id).
		First(&reservation).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
	}
	return &reservation, nil
}

// reservationListPreloads は予約一覧/単件取得（multi-clinic 版）で共有する preload チェーンを
// 構築する（BE-refactor.md E-15）。withCreatedByStaff は既存2箇所間の差分（意図不明のため
// ヘルパー化時に潰さず維持する） — 一覧版は false（CreatedByStaff を含まない）、単件版は
// true（CreatedByStaff を含む）を渡す。
func reservationListPreloads(q *gorm.DB, clinicIDs []uint64, withCreatedByStaff bool) *gorm.DB {
	q = q.Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pet", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pet.Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pet.AnimalSpecies").
		Preload("ReservationType", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("ReservationType.Group", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Doctor", staffAssignedToClinicsCond, clinicIDs)
	if withCreatedByStaff {
		q = q.Preload("CreatedByStaff", staffAssignedToClinicsCond, clinicIDs)
	}
	return q
}

func (r *reservationRepository) Create(ctx context.Context, reservation *model.Reservation) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(reservation).Error; err != nil {
		if persistence.IsUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("reservation", reservation.StartTime.String())
		}
		return apperrors.FromGORM(err, "reservation", "")
	}
	return nil
}

func (r *reservationRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
	// RSV-03: write + reload in one transaction so a post-commit Find cannot invert success.
	var loaded *model.Reservation
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := persistence.UpdateScopedByID(txCtx, tx, &model.Reservation{}, "reservation", clinicID, id, fields); err != nil {
			return err
		}
		var findErr error
		loaded, findErr = r.FindByID(txCtx, clinicID, id)
		return findErr
	})
	if err != nil {
		return nil, err
	}
	return loaded, nil
}

func (r *reservationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).Delete(&model.Reservation{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationRepository) ExistsByReservationTypeID(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("reservation_type_id = ? AND deleted_at IS NULL", reservationTypeID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "reservation", "")
	}
	return count > 0, nil
}

func (r *reservationRepository) ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("doctor_id = ? AND deleted_at IS NULL", staffID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "reservation", "")
	}
	return count > 0, nil
}

func (r *reservationRepository) FindClinicIDsByStaffID(
	ctx context.Context,
	clinicIDs []uint64,
	staffID uint64,
) ([]uint64, error) {
	result := make([]uint64, 0)
	if len(clinicIDs) == 0 {
		return result, nil
	}

	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Scopes(persistence.ClinicScopeIn(clinicIDs)).
		Where("doctor_id = ? AND deleted_at IS NULL", staffID).
		Distinct().
		Order("clinic_id ASC").
		Pluck("clinic_id", &result).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", "")
	}
	return result, nil
}

// CountMedicalRecordsByReservationID は予約を参照している有効カルテの件数を返す（BUG-201 / SEC-SWEEP-02-RES-B1）。
// 親 appointments の clinic 相関で cross-tenant 親を除外する一方、medical_records.clinic_id は
// フィルタしない（参照が存在する限り削除・identity 変更ガードを fail-closed に保つ — BILL-B1b と同型）。
// 親 appointments.deleted_at は入れない（MR-B1 / TRIM-B1 と同じく clinic 相関のみ）。
// Delete / UpdateForTrimming / DeleteForTrimming の依存チェックと同じ ambient transaction へ参加する。
func (r *reservationRepository) CountMedicalRecordsByReservationID(ctx context.Context, clinicID, reservationID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.MedicalRecord{}).
		Joins("JOIN appointments ON appointments.id = medical_records.appointment_id AND appointments.clinic_id = ?", clinicID).
		Where("medical_records.appointment_id = ? AND medical_records.deleted_at IS NULL", reservationID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return count, nil
}

// AcquireBookingLock は clinic 単位の pg_advisory_xact_lock を取得する（BE-refactor.md X-9）。
// hashtextextended で "appointments:{clinicID}" をハッシュ化した bigint をロックキーに使う。
// pg_advisory_xact_lock はトランザクションスコープのため、呼び出し元の WithTx がコミット/
// ロールバックした時点で自動解放される（明示的な unlock 不要）。dbOrTx でトランザクション
// 内の ambient tx に参加する。
func (r *reservationRepository) AcquireBookingLock(ctx context.Context, clinicID uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("booking lock requires an ambient transaction")
	}
	lockKey := fmt.Sprintf("appointments:%d", clinicID)
	if err := persistence.DBOrTx(ctx, r.db).Exec(
		"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
		lockKey,
	).Error; err != nil {
		return apperrors.Wrap(err, "failed to acquire booking lock")
	}
	return nil
}

// LockAndFindByID は FOR UPDATE で予約を行ロック取得する。
// updateWithConflictCheck のトランザクション内で使用する。
func (r *reservationRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if err := requireReservationRowLockTransaction(ctx); err != nil {
		return nil, err
	}
	var appt model.Reservation
	err := persistence.DBOrTx(ctx, r.db).Raw(
		`SELECT * FROM appointments WHERE clinic_id = ? AND id = ? AND deleted_at IS NULL FOR UPDATE`,
		clinicID, id,
	).Scan(&appt).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
	}
	if appt.ID == 0 {
		return nil, apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return &appt, nil
}

// HasDoctorConflict は指定医師の時間枠重複を SELECT FOR UPDATE でチェックする。
func (r *reservationRepository) HasDoctorConflict(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error) {
	if err := requireReservationRowLockTransaction(ctx); err != nil {
		return false, err
	}
	var existing []struct{ ID uint64 }
	excl := uint64(0)
	if excludeID != nil {
		excl = *excludeID
	}
	err := persistence.DBOrTx(ctx, r.db).Raw(`
		SELECT id FROM appointments
		WHERE clinic_id = ?
		  AND deleted_at IS NULL
		  AND status NOT IN ('cancelled', 'no_show')
		  AND start_time < ?
		  AND end_time > ?
		  AND doctor_id = ?
		  AND (? = 0 OR id != ?)
		FOR UPDATE`,
		clinicID, end, start, doctorID, excl, excl,
	).Scan(&existing).Error
	if err != nil {
		return false, apperrors.Wrap(err, "lock reservations for doctor conflict check")
	}
	return len(existing) > 0, nil
}

// CountOnDutyDoctors は当日の出勤医師数を返す。
func (r *reservationRepository) CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Raw(`
		SELECT COUNT(DISTINCT se.staff_id)
		FROM shift_entries se
		JOIN staffs s ON s.id = se.staff_id
		WHERE se.clinic_id = ?
		  AND se.date = DATE(? AT TIME ZONE 'Asia/Tokyo')
		  AND se.shift_type NOT IN ('off', 'paid_leave')
		  AND s.staff_type = 'doctor'
		  AND s.is_active = true
		  AND s.deleted_at IS NULL`,
		clinicID, date,
	).Scan(&count).Error
	if err != nil {
		return 0, apperrors.Wrap(err, "count on-duty doctors")
	}
	return count, nil
}

// CountConflicts は時間枠の競合予約数を SELECT FOR UPDATE で返す。
func (r *reservationRepository) CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error) {
	if err := requireReservationRowLockTransaction(ctx); err != nil {
		return 0, err
	}
	var existing []struct{ ID uint64 }
	excl := uint64(0)
	if excludeID != nil {
		excl = *excludeID
	}
	err := persistence.DBOrTx(ctx, r.db).Raw(`
		SELECT id FROM appointments
		WHERE clinic_id = ?
		  AND deleted_at IS NULL
		  AND status NOT IN ('cancelled')
		  AND start_time < ?
		  AND end_time > ?
		  AND (? = 0 OR id != ?)
		FOR UPDATE`,
		clinicID, end, start, excl, excl,
	).Scan(&existing).Error
	if err != nil {
		return 0, apperrors.Wrap(err, "lock reservations for capacity check")
	}
	return int64(len(existing)), nil
}

func requireReservationRowLockTransaction(ctx context.Context) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("reservation row lock requires an ambient transaction")
	}
	return nil
}

func (r *reservationRepository) CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
	var count int64
	q := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Where("clinic_id = ? AND reservation_type_id = ? AND start_time = ? AND status NOT IN ('cancelled') AND deleted_at IS NULL",
			clinicID, reservationTypeID, startTime)
	if excludeID != nil {
		q = q.Where("id != ?", *excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation", "")
	}
	return count, nil
}

// countByTypeAndStartTimeRow is the GROUP BY scan target for CountByTypeAndStartTimes.
type countByTypeAndStartTimeRow struct {
	StartTime time.Time
	Count     int64
}

// CountByTypeAndStartTimes は複数の開始時刻の予約件数を一括取得する（GROUP BY start_time）。
// BE-refactor.md R2-4 (D8): liff_service.FilterSlotsByCapacity の N+1（日付ごとの各スロットで
// CountByTypeAndStartTime を個別発行）を解消するためのバッチ経路。CountByTypeAndStartTime を
// 置き換えるものではなく（reservation_service の単発チェックは従来どおり）、追加の一括経路として
// 提供する。戻り値は startTime.Unix() 秒 → count のマップ（time.Time を map key にすると
// Location/monotonic 差異で等価判定が壊れるため Unix 秒で正規化する）。
// startTimes が空の場合は空マップを返す（クエリを発行しない）。
func (r *reservationRepository) CountByTypeAndStartTimes(ctx context.Context, clinicID, reservationTypeID uint64, startTimes []time.Time, excludeID *uint64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(startTimes))
	if len(startTimes) == 0 {
		return result, nil
	}
	var rows []countByTypeAndStartTimeRow
	q := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Select("start_time, COUNT(*) AS count").
		Where("clinic_id = ? AND reservation_type_id = ? AND start_time IN ? AND status NOT IN ('cancelled') AND deleted_at IS NULL",
			clinicID, reservationTypeID, startTimes).
		Group("start_time")
	if excludeID != nil {
		q = q.Where("id != ?", *excludeID)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation", "")
	}
	for _, row := range rows {
		result[row.StartTime.Unix()] = row.Count
	}
	return result, nil
}

// CountByCustomerAndDateRange は顧客・期間での予約件数を返す。
// 日次・月次制限チェックで使用する。
func (r *reservationRepository) CountByCustomerAndDateRange(ctx context.Context, clinicID, customerID uint64, start, end time.Time) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("line_customer_id = ? AND status NOT IN ('cancelled') AND start_time >= ? AND start_time < ? AND deleted_at IS NULL",
			customerID, start, end,
		).Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "reservation", "")
	}
	return count, nil
}

// FindAllByCategory はカテゴリでフィルタした予約一覧を返す（BE-119 トリミング管理 API）。
func (r *reservationRepository) FindAllByCategory(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error) {
	reservations := make([]model.Reservation, 0)
	var total int64

	q := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Where("appointments.clinic_id = ?", clinicID).
		Joins("JOIN reservation_types ON reservation_types.id = appointments.reservation_type_id AND reservation_types.clinic_id = appointments.clinic_id AND reservation_types.deleted_at IS NULL").
		Where("reservation_types.category = ?", category)

	if petID != nil || ownerID != nil {
		q = q.Joins("JOIN pets filter_pets ON filter_pets.id = appointments.pet_id AND filter_pets.clinic_id = appointments.clinic_id AND filter_pets.deleted_at IS NULL")
	}
	if petID != nil {
		q = q.Where("filter_pets.id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Joins("JOIN owners filter_owners ON filter_owners.id = filter_pets.owner_id AND filter_owners.clinic_id = appointments.clinic_id AND filter_owners.deleted_at IS NULL").
			Where("filter_owners.id = ?", *ownerID)
	}
	if startDate != nil {
		start, err := ParseJSTDate(*startDate)
		if err != nil {
			return nil, 0, apperrors.WrapInvalidInput(err.Error())
		}
		q = q.Where("appointments.start_time >= ?", start)
	}
	if endDate != nil {
		end, err := ParseJSTDate(*endDate)
		if err != nil {
			return nil, 0, apperrors.WrapInvalidInput(err.Error())
		}
		q = q.Where("appointments.start_time < ?", end.AddDate(0, 0, 1))
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "appointment", "")
	}
	if err := q.
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet.Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet.AnimalSpecies").
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", staffAssignedToClinicsCond, []uint64{clinicID}).
		Preload("TrimmingDetail", "clinic_id = ?", clinicID).
		Preload("TrimmingDetail.Course", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("TrimmingDetail.Options", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.Paginate(page, limit)).
		Order("appointments.start_time DESC").
		Find(&reservations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "appointment", "")
	}
	return reservations, total, nil
}

// CountByDateAndSource は日付・ソースの予約件数を返す。
// 確認番号生成で使用する。
func (r *reservationRepository) CountByDateAndSource(ctx context.Context, clinicID uint64, date time.Time, source model.ReservationSource) (int64, error) {
	var count int64
	start, end := AppointmentDayRange(date)
	err := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("start_time >= ? AND start_time < ? AND source = ? AND deleted_at IS NULL", start, end, source).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "reservation", "")
	}
	return count, nil
}

// FindNoShowCandidates は終了から4時間以上経過した confirmed/pending 予約のうち、
// 確定済みカルテが存在しないものを返す（BE-014）。
func (r *reservationRepository) FindNoShowCandidates(ctx context.Context, clinicID uint64) ([]model.Reservation, error) {
	return r.FindNoShowCandidatesAt(ctx, clinicID, time.Now().UTC())
}

// noShowCandidateMax is a safety cap for no-show batch candidate loads (G2F-08).
// Oldest-first (id ASC) keeps processing deterministic; next batch cycle drains the rest.
const noShowCandidateMax = 500

// FindNoShowCandidatesAt evaluates the complete candidate predicate against
// the durable scheduler timestamp instead of database wall-clock time.
func (r *reservationRepository) FindNoShowCandidatesAt(
	ctx context.Context,
	clinicID uint64,
	evaluatedAt time.Time,
) ([]model.Reservation, error) {
	var reservations []model.Reservation
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND deleted_at IS NULL AND status IN ? AND end_time <= CAST(? AS timestamptz) - interval '4 hours'",
			clinicID,
			[]string{string(model.ReservationStatusConfirmed), string(model.ReservationStatusPending)},
			evaluatedAt).
		Where(`NOT EXISTS (
			SELECT 1 FROM medical_records mr
			WHERE mr.clinic_id = appointments.clinic_id
			  AND mr.appointment_id = appointments.id
			  AND mr.status = ?
			  AND mr.deleted_at IS NULL
		)`, model.MedicalRecordStatusFinalized).
		Order("id ASC").
		Limit(noShowCandidateMax).
		Find(&reservations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", "")
	}
	return reservations, nil
}

func AppointmentDayRange(date time.Time) (start, end time.Time) {
	dateJST := date.In(config.JST)
	start = time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, dateJST.Location())
	end = start.AddDate(0, 0, 1)
	return start, end
}

func ParseJSTDate(value string) (time.Time, error) {
	t, err := time.ParseInLocation(time.DateOnly, value, config.JST)
	if err != nil {
		return time.Time{}, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format")
	}
	return t, nil
}

// AssertOwnerInClinic は owners を clinic スコープで存在確認する（AUD-001）。
// 別 clinic / 未存在を区別せず NotFound を返す。dbOrTx で ambient tx に参加する。
func (r *reservationRepository) AssertOwnerInClinic(ctx context.Context, clinicID, ownerID uint64) error {
	var id uint64
	db := persistence.DBOrTx(ctx, r.db).Model(&model.Owner{})
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Scopes(persistence.ClinicScope(clinicID)).
		Select("id").
		Where("id = ?", ownerID).
		Take(&id).Error
	if err != nil {
		return apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", ownerID))
	}
	return nil
}

// FindPetOwnerInClinic は pets とその owner の双方が同一 clinic に属する場合だけ OwnerID を返す。
// transaction 内では両行を共有ロックし、検証後から予約writeまでの clinic/owner 関係変更を防ぐ。
func (r *reservationRepository) FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error) {
	var pet model.Pet
	db := persistence.DBOrTx(ctx, r.db).Model(&model.Pet{})
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Select("pets.id", "pets.owner_id").
		Joins("JOIN owners ON owners.id = pets.owner_id AND owners.clinic_id = pets.clinic_id AND owners.deleted_at IS NULL").
		Where("pets.id = ? AND pets.clinic_id = ? AND pets.deleted_at IS NULL", petID, clinicID).
		First(&pet).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", petID))
	}
	return pet.OwnerID, nil
}

// FindPetByIDInClinic は clinic スコープでペットを読み、死亡 write ガード（SD-10）に必要な列を返す。
// transaction 内では行を共有ロックし、検証後から write までの deceased_at 変更を防ぐ。
func (r *reservationRepository) FindPetByIDInClinic(ctx context.Context, clinicID, petID uint64) (*model.Pet, error) {
	var pet model.Pet
	db := persistence.DBOrTx(ctx, r.db).Model(&model.Pet{})
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Select("pets.id", "pets.owner_id", "pets.deceased_at", "pets.status").
		Where("pets.id = ? AND pets.clinic_id = ? AND pets.deleted_at IS NULL", petID, clinicID).
		First(&pet).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", petID))
	}
	return &pet, nil
}

// AssertLineCustomerInClinic は line_customers を clinic スコープで存在確認する（AUD-001）。
func (r *reservationRepository) AssertLineCustomerInClinic(ctx context.Context, clinicID, lineCustomerID uint64) error {
	var id uint64
	db := persistence.DBOrTx(ctx, r.db).Model(&model.LineCustomer{})
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Scopes(persistence.ClinicScope(clinicID)).
		Select("id").
		Where("id = ?", lineCustomerID).
		Take(&id).Error
	if err != nil {
		return apperrors.FromGORM(err, "line_customer", fmt.Sprintf("%d", lineCustomerID))
	}
	return nil
}
