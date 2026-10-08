package lstep

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type deliveryOwnerRepository interface {
	FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
}

type deliveryMedicalRecordRepository interface {
	FindOwnersByFirstVisitDate(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
	FindOwnersByLastVisitDays(ctx context.Context, clinicID uint64, exactDays int, asOf time.Time) ([]uint64, error)
}

// deliveryAppointmentRepository は tentative リマインダー配信が必要とする
// appointment 参照の最小view（EMR-308）。candidate 抽出と送信直前の stale
// 再検証の両方で使う。具象は reservation domain の store（composition 注入）。
type deliveryAppointmentRepository interface {
	// FindOwnerIDsWithTentativeIntendedDate は intended_date が JST 暦日で
	// targetDate に一致する active tentative を持つ飼い主IDを返す。
	FindOwnerIDsWithTentativeIntendedDate(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
	// FindTentativeByOwnerAndIntendedDate は飼い主の active tentative のうち
	// intended_date が targetDate（JST 暦日）の1件を返す。送信直前の再検証で
	// 取消・完了・削除・実予約昇格済みの stale candidate を排除するために使う。
	// 該当なしは (nil, nil)。
	FindTentativeByOwnerAndIntendedDate(ctx context.Context, clinicID, ownerID uint64, targetDate time.Time) (*model.Reservation, error)
	// WithTentativeReminderSerialization は tentative mutation と同じ
	// (clinic,pet) advisory lock を保持する tx 内で fn を実行する。
	// 送信直前再検証〜外部送信を直列化し、stale candidate への配信を防ぐ。
	WithTentativeReminderSerialization(ctx context.Context, clinicID, petID uint64, fn func(txCtx context.Context) error) error
}

type deliveryVaccinationRepository interface {
	FindOwnersByVaccineDeadline(ctx context.Context, clinicID uint64, targetDate time.Time) ([]uint64, error)
}

type deliveryPetRepository interface {
	FindOwnersByPetBirthday(ctx context.Context, clinicID uint64, month, day int) ([]uint64, error)
}

type deliveryTagCacheRepository interface {
	FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
	FindOwnerIDsByTag(ctx context.Context, clinicID uint64, tagName string) ([]uint64, error)
}
