package lstep

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type LineSendLogRepository interface {
	Create(ctx context.Context, log *model.LineSendLog) error
	FindByOwner(ctx context.Context, clinicID, ownerID uint64, limit int) ([]*model.LineSendLog, error)
}

type lineSendLogRepository struct {
	db *gorm.DB
}

func NewLineSendLogRepository(db *gorm.DB) LineSendLogRepository {
	return &lineSendLogRepository{db: db}
}

func (r *lineSendLogRepository) Create(ctx context.Context, log *model.LineSendLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return apperrors.FromGORM(err, "line_send_log", "")
	}
	return nil
}

func (r *lineSendLogRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64, limit int) ([]*model.LineSendLog, error) {
	var logs []*model.LineSendLog
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND owner_id = ?", clinicID, ownerID).
		Order("sent_at DESC").
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "line_send_log", "")
	}
	return logs, nil
}
