package main

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
)

// staffAccountStoreAdapter combines auth-owned account persistence with the
// reset-token invalidation required by staff-admin credential changes.
type staffAccountStoreAdapter struct {
	accounts    auth.AccountRepository
	resetTokens auth.PasswordResetTokenInvalidator
}

func (a staffAccountStoreAdapter) FindByEmail(
	ctx context.Context,
	email string,
) (*model.Account, error) {
	return a.accounts.FindByEmail(ctx, email)
}

func (a staffAccountStoreAdapter) Create(
	ctx context.Context,
	account *model.Account,
) error {
	return a.accounts.Create(ctx, account)
}

func (a staffAccountStoreAdapter) UpdatePasswordHash(
	ctx context.Context,
	accountID uint64,
	newHash string,
	updatedAt time.Time,
) error {
	return a.accounts.UpdatePasswordHash(
		ctx,
		accountID,
		newHash,
		updatedAt,
	)
}

func (a staffAccountStoreAdapter) DeletePasswordResetTokens(
	ctx context.Context,
	accountID uint64,
) error {
	return a.resetTokens.DeleteByAccountID(ctx, accountID)
}
