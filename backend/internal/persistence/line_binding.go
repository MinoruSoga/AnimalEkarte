package persistence

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// LineBindingStateLinked は owner_line_binding_states.state の「紐付け済み」値。
// 非 LIFF の読み手（タグ同期・配信・クリーンアップ等）が「この LINE User ID へ作用してよい」と
// 判断する唯一の状態（EMR-504 Q1(a)）。unlink_pending / unlinked / legacy_unknown は未連携として扱う。
const LineBindingStateLinked = "linked"

// LineBindingReader は owner の現行 LINE 紐付けから配信先 LINE User ID を読む read-only reader（EMR-504）。
// 旧 owners.line_user_id は DROP 済みで、配信先の正本は owner_line_binding_states（state='linked'）が指す
// line_customers.line_user_id。writer は EMR-469 が持ち、ここには置かない。旧列・キャッシュ・name 照合へは
// フォールバックしない。ambient transaction があればそれを使う。
type LineBindingReader struct {
	db *gorm.DB
}

// NewLineBindingReader は reader を返す。
func NewLineBindingReader(db *gorm.DB) *LineBindingReader {
	return &LineBindingReader{db: db}
}

type lineBindingRecipientRow struct {
	OwnerID    uint64 `gorm:"column:owner_id"`
	LineUserID string `gorm:"column:line_user_id"`
}

// scopedDB は ambient transaction があればそれを、無ければ reader の DB を返す。
func (r *LineBindingReader) scopedDB(ctx context.Context) (*gorm.DB, error) {
	if r == nil || r.db == nil {
		return nil, apperrors.WrapInternalServerError("line binding reader is not configured")
	}
	if tx := TxFromContext(ctx); tx != nil {
		return tx.WithContext(ctx), nil
	}
	return r.db.WithContext(ctx), nil
}

// LinkedLineUserID は owner の配信先 LINE User ID を返す。紐付け済みでなければ (nil, nil)。
// 読み取り失敗は error で、未連携とは区別する。
func (r *LineBindingReader) LinkedLineUserID(ctx context.Context, clinicID, ownerID uint64) (*string, error) {
	db, err := r.scopedDB(ctx)
	if err != nil {
		return nil, err
	}
	return LinkedLineUserIDOn(db, clinicID, ownerID)
}

// LinkedLineUserIDs は ownerIDs のうち紐付け済みの owner の配信先を返す（未連携の owner は含まれない）。
func (r *LineBindingReader) LinkedLineUserIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64]string, error) {
	db, err := r.scopedDB(ctx)
	if err != nil {
		return nil, err
	}
	return LinkedLineUserIDsOn(db, clinicID, ownerIDs)
}

// LinkedLineUserIDOn は db（tx でもよい）上で owner 1 件の配信先を読む。
func LinkedLineUserIDOn(db *gorm.DB, clinicID, ownerID uint64) (*string, error) {
	recipients, err := LinkedLineUserIDsOn(db, clinicID, []uint64{ownerID})
	if err != nil {
		return nil, err
	}
	lineUserID, ok := recipients[ownerID]
	if !ok {
		return nil, nil
	}
	return &lineUserID, nil
}

// LinkedLineUserIDsOn は db（tx でもよい）上で ownerIDs の配信先を一括で読む。
// state='linked' の binding が指す line_customers だけを対象にし、同一医院の行に限定する。
func LinkedLineUserIDsOn(db *gorm.DB, clinicID uint64, ownerIDs []uint64) (map[uint64]string, error) {
	recipients := make(map[uint64]string, len(ownerIDs))
	if len(ownerIDs) == 0 {
		return recipients, nil
	}
	if clinicID == 0 {
		return nil, apperrors.WrapInternalServerError("linked LINE recipient read requires a clinic scope")
	}
	var rows []lineBindingRecipientRow
	if err := db.
		Table("owner_line_binding_states AS lb").
		Select("lb.owner_id AS owner_id", "lc.line_user_id AS line_user_id").
		Joins("JOIN line_customers lc ON lc.id = lb.line_customer_id AND lc.clinic_id = lb.clinic_id").
		Where("lb.clinic_id = ? AND lb.owner_id IN ? AND lb.state = ?", clinicID, ownerIDs, LineBindingStateLinked).
		Scan(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to read linked LINE recipients")
	}
	for i := range rows {
		if strings.TrimSpace(rows[i].LineUserID) == "" {
			continue
		}
		recipients[rows[i].OwnerID] = rows[i].LineUserID
	}
	return recipients, nil
}
