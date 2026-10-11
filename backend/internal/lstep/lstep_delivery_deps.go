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
	// FindTentativesByOwnerAndIntendedDate は飼い主の active tentative のうち
	// intended_date が targetDate（JST 暦日）のものを全件返す。candidate 抽出と
	// 送信直前の再検証で 取消・完了・削除・実予約昇格済みの stale candidate を
	// 排除するために使う。Codex #22: 同一 owner の複数 pet が対象日の tentative
	// を持ち得るため単一 candidate のみのロックでは不十分 —— 全件を取得して
	// 全 pet のロックを保持する。
	FindTentativesByOwnerAndIntendedDate(ctx context.Context, clinicID, ownerID uint64, targetDate time.Time) ([]model.Reservation, error)
	// WithTentativeReminderSerialization は tentative mutation と同じ
	// (clinic,pet) advisory lock を指定された全 pet について保持する tx 内で
	// fn を実行する。送信直前再検証〜外部送信を直列化し、stale candidate への
	// 配信を防ぐ。petIDs は昇順・重複排除してからロックされる。
	WithTentativeReminderSerialization(ctx context.Context, clinicID uint64, petIDs []uint64, fn func(txCtx context.Context) error) error
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
