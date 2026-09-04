package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// AccountRepository provides global login-account persistence.
// Accounts intentionally have no clinic_id; clinic access is authorized through staff assignments.
type AccountRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.Account, error)
	FindByIDForUpdate(ctx context.Context, id uint64) (*model.Account, error)
	FindByEmail(ctx context.Context, email string) (*model.Account, error)
	Create(ctx context.Context, account *model.Account) error
	CompareAndSwapPasswordHash(
		ctx context.Context,
		id uint64,
		expectedHash, newHash string,
		updatedAt time.Time,
	) (bool, error)
	UpdatePasswordHash(
		ctx context.Context,
		accountID uint64,
		newHash string,
		updatedAt time.Time,
	) error
}

type accountRepository struct {
	db *gorm.DB
}

// FindByIDForUpdate locks a non-deleted account row for a transaction-scoped mutation.
func (r *accountRepository) FindByIDForUpdate(
	ctx context.Context,
	id uint64,
) (*model.Account, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return nil, apperrors.WrapInternalServerError("account lock requires an ambient transaction")
	}
	var account model.Account
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&account, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
		return nil, apperrors.FromGORM(err, "account", fmt.Sprintf("%d", id))
	}
	return &account, nil
}

// NewAccountRepository constructs global login-account persistence.
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

// FindByID はIDでアカウントを取得する
func (r *accountRepository) FindByID(ctx context.Context, id uint64) (*model.Account, error) {
	var account model.Account
	if err := persistence.DBOrTx(ctx, r.db).First(&account, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("account", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.FromGORM(err, "account", fmt.Sprintf("%d", id))
	}
	return &account, nil
}

// FindByEmail はメールアドレスでアカウントを検索する
func (r *accountRepository) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	var account model.Account
	db := persistence.DBOrTx(ctx, r.db)
	// Email is authentication PII. Suppress SQL value logging for this lookup
	// and surface only the sanitized repository error below.
	result := db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Silent),
	}).Where("email = ? AND deleted_at IS NULL", email).Limit(1).Find(&account)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "account", "email_lookup")
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("account", "email_lookup")
	}
	return &account, nil
}

// Create はアカウントを新規作成する
func (r *accountRepository) Create(ctx context.Context, account *model.Account) error {
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Silent),
	}).Create(account).Error; err != nil {
		return apperrors.FromGORM(err, "account", "create")
	}
	return nil
}

// CompareAndSwapPasswordHash replaces the credential only while the persisted
// hash still matches the hash that the caller verified.
func (r *accountRepository) CompareAndSwapPasswordHash(
	ctx context.Context,
	id uint64,
	expectedHash, newHash string,
	updatedAt time.Time,
) (bool, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return false, apperrors.WrapInternalServerError(
			"password credential compare-and-swap requires an ambient transaction",
		)
	}
	// Password hashes are credentials. Suppress SQL value logging even when
	// the database rejects the update.
	result := tx.Session(&gorm.Session{
		Logger: tx.Logger.LogMode(logger.Silent),
	}).
		Model(&model.Account{}).
		Where(
			"id = ? AND deleted_at IS NULL AND password_hash = ?",
			id,
			expectedHash,
		).
		Updates(map[string]any{
			"password_hash": newHash,
			"updated_at":    updatedAt,
		})
	if result.Error != nil {
		return false, apperrors.FromGORM(
			result.Error,
			"account",
			fmt.Sprintf("%d", id),
		)
	}
	if result.RowsAffected > 1 {
		return false, apperrors.WrapInternalServerError(
			"password hash compare-and-swap affected multiple accounts",
		)
	}
	return result.RowsAffected == 1, nil
}

// UpdatePasswordHash updates a credential only inside an ambient transaction.
// Password hashes never pass through the generic Update path because GORM's
// default error and slow-query logger can interpolate SQL parameters.
func (r *accountRepository) UpdatePasswordHash(
	ctx context.Context,
	id uint64,
	newHash string,
	updatedAt time.Time,
) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError(
			"password credential update requires an ambient transaction",
		)
	}
	result := tx.Session(&gorm.Session{
		Logger: tx.Logger.LogMode(logger.Silent),
	}).
		Model(&model.Account{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"password_hash": newHash,
			"updated_at": gorm.Expr(
				"GREATEST(updated_at + INTERVAL '1 microsecond', ?)",
				updatedAt,
			),
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "account", fmt.Sprintf("%d", id))
	}
	switch result.RowsAffected {
	case 1:
		return nil
	case 0:
		return apperrors.WrapNotFound("account", fmt.Sprintf("%d", id))
	default:
		return apperrors.WrapInternalServerError(
			"password credential update affected multiple accounts",
		)
	}
}
