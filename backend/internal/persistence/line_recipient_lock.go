package persistence

import (
	"context"
	"fmt"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func AcquireLineRecipientLock(ctx context.Context, clinicID uint64, lineUserID string) error {
	tx := TxFromContext(ctx)
	if tx == nil || clinicID == 0 || strings.TrimSpace(lineUserID) == "" {
		return apperrors.WrapInternalServerError("line recipient lock requires an ambient transaction and non-zero scope")
	}
	key := fmt.Sprintf("lstep-line-recipient:%d:%s", clinicID, lineUserID)
	if err := tx.WithContext(ctx).Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", key).Error; err != nil {
		return apperrors.Wrap(err, "failed to lock line recipient")
	}
	return nil
}
