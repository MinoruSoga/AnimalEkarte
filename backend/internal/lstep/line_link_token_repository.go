package lstep

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LineLinkTokenRepository は LINE 紐付けトークンのデータアクセスインターフェース（BE-021）。
type LineLinkTokenRepository interface {
	Create(ctx context.Context, t *model.LineLinkToken) error
	// LockUsableByRawToken は raw token の digest を検索し、使用可能な行を
	// ambient transaction 内で FOR UPDATE ロックする。
	LockUsableByRawToken(ctx context.Context, rawToken string, now time.Time) (*model.LineLinkToken, error)
	// Consume は未使用・未失効の行だけを使用済みにする単回 CAS。
	Consume(ctx context.Context, id uint64, usedAt time.Time) error
}

type lineLinkTokenRepository struct{ db *gorm.DB }

func NewLineLinkTokenRepository(db *gorm.DB) LineLinkTokenRepository {
	return &lineLinkTokenRepository{db: db}
}

func (r *lineLinkTokenRepository) Create(ctx context.Context, t *model.LineLinkToken) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(t).Error; err != nil {
		return apperrors.FromGORM(err, "line_link_token", "")
	}
	return nil
}

func (r *lineLinkTokenRepository) LockUsableByRawToken(
	ctx context.Context,
	rawToken string,
	now time.Time,
) (*model.LineLinkToken, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("LINE link token lock requires an ambient transaction")
	}
	db := persistence.DBOrTx(ctx, r.db)
	token, err := r.lockUsableByStoredToken(ctx, db, digestLineLinkToken(rawToken), now)
	if err == nil {
		return token, nil
	}
	if !apperrors.IsNotFound(err) || !isLegacyLineLinkRawToken(rawToken) {
		return nil, err
	}
	return r.lockUsableByStoredToken(ctx, db, rawToken, now)
}

func (r *lineLinkTokenRepository) lockUsableByStoredToken(
	ctx context.Context,
	db *gorm.DB,
	storedToken string,
	now time.Time,
) (*model.LineLinkToken, error) {
	var t model.LineLinkToken
	err := db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token = ? AND used_at IS NULL AND expires_at > ?", storedToken, now).
		First(&t).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "line_link_token", "")
	}
	return &t, nil
}

func (r *lineLinkTokenRepository) Consume(ctx context.Context, id uint64, usedAt time.Time) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("LINE link token consume requires an ambient transaction")
	}
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.LineLinkToken{}).
		Where("id = ? AND used_at IS NULL AND expires_at > ?", id, usedAt).
		Update("used_at", usedAt)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "line_link_token", "")
	}
	if result.RowsAffected != 1 {
		return apperrors.WrapConflict("LINE link token is already used or expired")
	}
	return nil
}

func digestLineLinkToken(rawToken string) string {
	digest := sha256.Sum256([]byte(rawToken))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func isLegacyLineLinkRawToken(rawToken string) bool {
	if len(rawToken) != hex.EncodedLen(32) {
		return false
	}
	_, err := hex.DecodeString(rawToken)
	return err == nil
}
