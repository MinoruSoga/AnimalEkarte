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

// trimmingCourseFinder / trimmingOptionFinder は trimming マスタ（未移行 domain）の
// 所有権確認に使う最小 view（LINE予約 validator 用）。
type trimmingCourseFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
}

type trimmingOptionFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
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
