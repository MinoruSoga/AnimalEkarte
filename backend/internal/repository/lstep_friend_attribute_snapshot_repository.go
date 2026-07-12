package repository

import (
	"context"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LstepFriendAttributeSnapshotRepository は lstep_friend_attribute_snapshots テーブルの永続化インターフェース。
type LstepFriendAttributeSnapshotRepository interface {
	// Create は新規友だち属性スナップショットを作成する。
	Create(ctx context.Context, snapshot *model.LstepFriendAttributeSnapshot) error
	// FindLatestByOwner はクリニックスコープで指定 LINE ユーザーの最新スナップショットを返す。
	FindLatestByOwner(ctx context.Context, clinicID uint64, lineUserID string) (*model.LstepFriendAttributeSnapshot, error)
}

type lstepFriendAttributeSnapshotRepository struct{ db *gorm.DB }

// NewLstepFriendAttributeSnapshotRepository は LstepFriendAttributeSnapshotRepository を初期化して返す。
func NewLstepFriendAttributeSnapshotRepository(db *gorm.DB) LstepFriendAttributeSnapshotRepository {
	return &lstepFriendAttributeSnapshotRepository{db: db}
}

func (r *lstepFriendAttributeSnapshotRepository) Create(ctx context.Context, snapshot *model.LstepFriendAttributeSnapshot) error {
	if err := r.db.WithContext(ctx).Create(snapshot).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_friend_attribute_snapshot", "")
	}
	return nil
}

func (r *lstepFriendAttributeSnapshotRepository) FindLatestByOwner(ctx context.Context, clinicID uint64, lineUserID string) (*model.LstepFriendAttributeSnapshot, error) {
	var snapshot model.LstepFriendAttributeSnapshot
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND line_user_id = ?", clinicID, lineUserID).
		Order("snapshot_taken_at DESC").
		First(&snapshot).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_friend_attribute_snapshot", lineUserID)
	}
	return &snapshot, nil
}
