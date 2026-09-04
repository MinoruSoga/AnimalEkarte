package reservation

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// staffAssignedToClinicsCond は Preload した staff（Doctor / CreatedByStaff）を、指定クリニック集合の
// いずれかに現在所属している場合のみ表示する条件。staff は staff_clinic_assignments による多医院所属のため
// staffs.clinic_id（主所属）単純スコープでは共有スタッフを誤って隠す。assignment-EXISTS で多医院所属を
// 尊重しつつ、別テナント単独所属スタッフ名の漏洩を防ぐ。予約は現在/未来データのため履歴表示の回帰はない。
const staffAssignedToClinicsCond = "deleted_at IS NULL AND EXISTS (SELECT 1 FROM staff_clinic_assignments sca WHERE sca.staff_id = staffs.id AND sca.clinic_id IN ? AND sca.deleted_at IS NULL)"

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

// countByTypeAndStartTimeRow is the GROUP BY scan target for CountByTypeAndStartTimes.
type countByTypeAndStartTimeRow struct {
	StartTime time.Time
	Count     int64
}

// noShowCandidateMax is a safety cap for no-show batch candidate loads (G2F-08).
// Oldest-first (id ASC) keeps processing deterministic; next batch cycle drains the rest.
const noShowCandidateMax = 500
