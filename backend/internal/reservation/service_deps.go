package reservation

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// consumer-side narrow views（ADR-006/BE9-2C R①）: reservation が必要とする他domain・
// 未移行repoの最小メソッド集合。実装は cmd/api/main.go（または service 集約）が注入する。

// occupationFinder は職種マスタ（staff domain）の所有権確認に使う最小view。
type occupationFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
}

// reservationTypeUsageChecker は予約区分の使用中判定に使う最小view（未移行のreservation query repo）。
type reservationTypeUsageChecker interface {
	ExistsByReservationTypeID(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error)
}

// Transactor は repository.Transactor の consumer-side view（WithTx のみ・medicalrecord 同型）。
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// ReservationStaffDeleter is the reservation-owned business-intent port for
// removing a staff member through the canonical staff lifecycle.
// isSystemAdmin mirrors the staff domain's signature. This route is addressed
// per clinic (/clinics/:clinic_id/reservation-staffs/:staffId), so it always
// passes false: deletion here stays scoped to the clinic in the path.
type ReservationStaffDeleter interface {
	Delete(ctx context.Context, clinicID, staffID uint64, isSystemAdmin bool) error
}

// trimmingCourseFinder / trimmingOptionFinder は trimming マスタ（未移行 domain）の
// 所有権確認に使う最小 view（LINE予約 validator 用）。
type trimmingCourseFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
}

type trimmingOptionFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
}

// medicalRecordAutoCreator は予約確定/キャンセルに伴うカルテ自動作成/下書き削除の最小view
// （medicalrecord domainはADR-006 topoでreservationより下流のためimport不可・main.goが具象注入）。
type medicalRecordAutoCreator interface {
	AutoCreateFromReservation(ctx context.Context, clinicID uint64, reservation *model.Reservation)
	DeleteDraftFromReservation(ctx context.Context, clinicID, reservationID uint64)
}

// liffAvailability は空き時間一覧（liff service=R⑤未移行）の最小view。
type liffAvailability interface {
	GetAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error)
}

// staffAssignmentFinder は医師のクリニック所属確認（staff domain）の最小view。
type staffAssignmentFinder interface {
	FindAllByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
}

// lineReservationSettingFinder は LINE 予約設定（R⑥未移行 repo）の最小view（通知用）。
type lineReservationSettingFinder interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
}

// liffSettingFinder は LIFF 公開設定の読取り（予約設定 + clinic_integrations の liff_id）に使う最小 view。
type liffSettingFinder interface {
	lineReservationSettingFinder
	FindIntegrationValue(ctx context.Context, clinicID uint64, keyName string) (string, error)
}

// reservationNotificationSettingFinder は予約の完了・キャンセル通知に使う最小 view。
// 通知先メールの予約設定に加え、LINE Push 専用の channel access token 暗号文リーダーを持つ
// （汎用の secret reader ではなく、復号は composition が注入する cipher closure が行う）。
type reservationNotificationSettingFinder interface {
	lineReservationSettingFinder
	FindLineChannelAccessTokenCiphertext(ctx context.Context, clinicID uint64) (string, error)
}

// clinicHolidayFinder は clinic_holidays の日付検索に使う consumer-side 最小 view。
// reservation は clinic package を import せず、composition が具象を注入する。
type clinicHolidayFinder interface {
	FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error)
}

// ---- LIFF 用 consumer-side views（未移行 domain の repo・BE9-2C R⑤）----

// LiffLineCustomerRepository is the reservation-owned view used by LIFF booking.
// It is exported only so the composition boundary can supply the LSTEP-owned implementation.
type LiffLineCustomerRepository interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error)
	UpdateOwnerLink(ctx context.Context, clinicID, id uint64, ownerID *uint64) error
	UpdateAdditionalFields(ctx context.Context, clinicID, id uint64, fields []byte) error
}

type liffOwnerRepo interface {
	FindByNameAndPhone(ctx context.Context, clinicID uint64, name, phone string) (*model.Owner, error)
}

type liffTrimmingDetailRepo interface {
	Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	SetOptions(ctx context.Context, clinicID, appointmentID uint64, optionIDs []uint64) error
}

type liffVaccinationRepo interface {
	FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error)
}

// NextVisitTagSyncer は EMR-308 の次回来院タグ（next_visit_YYYY-MM-DD）を
// appointment 状態から再同期する cross-domain view。reservation は lstep
// （topo 最下流）を import できないため、composition root が lstep の
// TagSync 実装を注入する。nil 許容・best-effort: 未注入や同期失敗は
// 予約変更容易点の成否に影響しない。
type NextVisitTagSyncer interface {
	SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error
}
