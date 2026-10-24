package lstep

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// lineBindingStateUnlinkPending は owner_line_binding_states.state の解除保留値。
// unique index が (clinic_id, line_customer_id) を保持し続けるため、「この LINE User ID を誰かがまだ
// 保持しているか」の判定（recipient 所有確認）だけが linked と併せて参照する。
const lineBindingStateUnlinkPending = "unlink_pending"

// lineRecipientReader は owner の現行 LINE 紐付けから配信先 LINE User ID を読む consumer-side 最小 view
// （EMR-504 Q1(a)）。owners.line_user_id は DROP 済みで、配信先の正本は owner_line_binding_states
// （state='linked'）が指す line_customers.line_user_id。具象は persistence.LineBindingReader。
// 未連携は (nil, nil) / map に含まれない。読み取り失敗は error で、未連携とは区別する。
type lineRecipientReader interface {
	LinkedLineUserID(ctx context.Context, clinicID, ownerID uint64) (*string, error)
	LinkedLineUserIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64]string, error)
}

// linkedLineUserID は owner の配信先を返す。未連携は ("", nil)。reader 未注入と読み取り失敗は error で、
// 「未連携」に畳まない（送信・タグ解除の誤スキップを防ぐ）。
func linkedLineUserID(ctx context.Context, reader lineRecipientReader, clinicID, ownerID uint64) (string, error) {
	if reader == nil {
		return "", apperrors.WrapInternalServerError("line recipient reader is not configured")
	}
	lineUserID, err := reader.LinkedLineUserID(ctx, clinicID, ownerID)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to read linked LINE recipient")
	}
	if lineUserID == nil {
		return "", nil
	}
	return *lineUserID, nil
}

// linkedLineUserIDs は ownerIDs の配信先を一括で返す（未連携の owner は含まれない）。
// ownerIDs が空なら reader は使わず空 map を返す。
func linkedLineUserIDs(ctx context.Context, reader lineRecipientReader, clinicID uint64, ownerIDs []uint64) (map[uint64]string, error) {
	if len(ownerIDs) == 0 {
		return map[uint64]string{}, nil
	}
	if reader == nil {
		return nil, apperrors.WrapInternalServerError("line recipient reader is not configured")
	}
	recipients, err := reader.LinkedLineUserIDs(ctx, clinicID, ownerIDs)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to read linked LINE recipients")
	}
	return recipients, nil
}

// lockOwnerLinkedRecipient は owner 行ロック取得後の同一 tx で、現行 binding の配信先に対する
// recipient advisory lock を取る（owner advisory -> owner 行 UPDATE -> recipient advisory の順序を維持）。
// recipient lock は LINE User ID が分かってから取るため、lock 待ち中に binding が付け替わっていないことを
// 別 statement（READ COMMITTED）で再確認し、変わっていれば conflict で中断する。
// 未連携なら recipient lock は不要で nil を返す。
func lockOwnerLinkedRecipient(ctx context.Context, tx *gorm.DB, clinicID, ownerID uint64) error {
	recipient, err := persistence.LinkedLineUserIDOn(tx, clinicID, ownerID)
	if err != nil {
		return err
	}
	if recipient == nil {
		return nil
	}
	if err := persistence.AcquireLineRecipientLock(persistence.WithTxValue(ctx, tx), clinicID, *recipient); err != nil {
		return err
	}
	current, err := persistence.LinkedLineUserIDOn(tx, clinicID, ownerID)
	if err != nil {
		return err
	}
	if current == nil || *current != *recipient {
		return apperrors.WrapConflict("LINE recipient changed while acquiring the recipient lock")
	}
	return nil
}
