package trimming

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// AppointmentTrimmingDetailRepository はトリミング詳細の CRUD を提供する。
type AppointmentTrimmingDetailRepository interface {
	FindByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error)
	Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	Update(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	SetOptions(ctx context.Context, clinicID, appointmentID uint64, optionIDs []uint64) error
	ImageKeyInUse(ctx context.Context, clinicID uint64, key string) (bool, error)
}

type appointmentTrimmingDetailRepository struct {
	db *gorm.DB
	// imageObjects は writer 側 claim がオブジェクト存在を確認する durable
	// ストア。Handler.WithImageSupport から imageObjectStoreBinder 経由で
	// 注入する。nil = 存在確認不可 → キー付き書き込みは fail-closed で拒否。
	imageObjects FileUploader
}

func NewAppointmentTrimmingDetailRepository(db *gorm.DB) AppointmentTrimmingDetailRepository {
	return &appointmentTrimmingDetailRepository{db: db}
}

func (r *appointmentTrimmingDetailRepository) bindImageObjectStore(uploader FileUploader) {
	r.imageObjects = uploader
}

func (r *appointmentTrimmingDetailRepository) FindByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error) {
	var detail model.AppointmentTrimmingDetail
	// Parent appointments clinic correlation (SEC-SWEEP-02-TRIM-B1): child clinic alone
	// is insufficient when appointment_id is a corrupt cross-tenant FK.
	// Qualify appointment_trimming_details.clinic_id (not ClinicScope) so JOIN appointments
	// does not make the bare clinic_id predicate ambiguous. No appointments.deleted_at
	// filter — matches MR-B1 appointments-parent pattern (clinic correlation only).
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN appointments ON appointments.id = appointment_trimming_details.appointment_id AND appointments.clinic_id = appointment_trimming_details.clinic_id").
		Preload("Course", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Options", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("appointment_trimming_details.clinic_id = ? AND appointment_trimming_details.appointment_id = ?", clinicID, appointmentID).
		First(&detail).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment_trimming_detail", fmt.Sprintf("appointment_id=%d", appointmentID))
	}
	return &detail, nil
}

func (r *appointmentTrimmingDetailRepository) Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
	db := persistence.DBOrTx(ctx, r.db)
	if err := r.claimImageKeysForStore(ctx, db, detail.StyleImage, detail.CompletedImage); err != nil {
		return err
	}
	if err := db.Create(detail).Error; err != nil {
		return apperrors.FromGORM(err, "appointment_trimming_detail", fmt.Sprintf("appointment_id=%d", detail.AppointmentID))
	}
	return nil
}

// Update は trimming 詳細の可変フィールドをすべて明示的に map で更新する。
// Updates(struct) はゼロ値フィールド（空文字等）をスキップするため、map を使用する。
func (r *appointmentTrimmingDetailRepository) Update(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
	fields := map[string]any{
		"course_id":        detail.CourseID,
		"style_request":    detail.StyleRequest,
		"body_weight":      detail.BodyWeight,
		"bw_unit":          string(detail.BWUnit),
		"body_temperature": detail.BodyTemperature,
		"used_shampoo":     detail.UsedShampoo,
		"used_ribbon":      detail.UsedRibbon,
		"remarks":          detail.Remarks,
		"style_image":      detail.StyleImage,
		"completed_image":  detail.CompletedImage,
	}
	db := persistence.DBOrTx(ctx, r.db)
	if err := r.claimImageKeysForStore(ctx, db, detail.StyleImage, detail.CompletedImage); err != nil {
		return err
	}
	result := db.
		Scopes(persistence.ClinicScope(detail.ClinicID)).
		Model(&model.AppointmentTrimmingDetail{}).
		Where("appointment_id = ?", detail.AppointmentID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "appointment_trimming_detail", fmt.Sprintf("appointment_id=%d", detail.AppointmentID))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("appointment_trimming_detail", fmt.Sprintf("appointment_id=%d", detail.AppointmentID))
	}
	return nil
}

// claimImageKeysForStore は style_image / completed_image を永続化する
// 全 writer が共通して辿るガード（repository Create/Update の入口）。
// 書き込み予定キーごとにキースコープの advisory xact lock を取得し、
// ロック下で durable ストアの Exists を確認する。
// 削除確定済み（オブジェクト不在）のキーへの格納は Conflict で拒否 —
// 消えた画像への dangling reference を作らせない。Exists のインフラ
// エラー・uploader 未配線は fail-closed エラー。ロック未対象の値
// （署名 URL 等）やキー無しの場合は何もしない。
func (r *appointmentTrimmingDetailRepository) claimImageKeysForStore(ctx context.Context, db *gorm.DB, storedValues ...string) error {
	keys := storedTrimmingImageObjectKeys(storedValues...)
	if len(keys) == 0 {
		return nil
	}
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("trimming image key store requires an ambient transaction")
	}
	if r.imageObjects == nil {
		return apperrors.WrapInternalServerError("trimming image object store is not configured")
	}
	for _, key := range keys {
		if err := acquireTrimmingImageKeyLock(ctx, db, key); err != nil {
			return err
		}
		exists, err := r.imageObjects.Exists(ctx, key)
		if err != nil {
			return apperrors.Wrap(err, "failed to check trimming image object existence")
		}
		if !exists {
			return apperrors.WrapConflict("trimming image is no longer available; upload a new image")
		}
	}
	return nil
}

// ImageKeyInUse は live な trimming detail 行が key を画像として参照して
// いるかを返す。DELETE /trimmings/images と post-commit tombstone
// クリーンアップの fail-closed 参照チェック用。同じ予約自身の参照も
// カウント対象 — 自レコードを除外すると更新直後に同じキーへ戻した参照を
// 見落としてオブジェクトを消してしまう。
//
// Soft-delete 参照契約（primary/live-parent）: detail 参照は親
// appointment が live (deleted_at IS NULL) の場合のみ live 参照として
// 数える。appointments.deleted_at を NULL へ戻す restore/undelete の
// write path が versioned source に存在しないことは frozen 列挙で立証済み
// （campaign evidence/appointments-no-restore-r9.json: 2 独立監査が
// reservation/csvimport/seed/migration/scripts の全 write path を網羅、
// source bound = image head a44b7c9 / outside-crud head dc96）。
// この前提は静的・versioned bound であり、手動 SQL・out-of-tree writer・
// latest-main 統合後の新規 write surface は証明の範囲外 — restore path が
// 追加された場合は soft-deleted 参照をカウントする保持側契約へ戻し、
// それまでオブジェクト削除を無効化することも選択肢となる。
//
// ambient tx 内では対象キーの advisory xact lock を取得してから走査する。
// 「参照確認→tombstone→オブジェクト削除」が同時 writer のコミットに対して
// 直列化され、check-to-delete ウィンドウが閉じる（writer 側は
// claimTrimmingImageKeysForStore が同じロックを取る）。
//
// 保存値は canonical key に正規化されるが、旧世代の `/uploads/` ラップ形式など
// resolveTrimmingImageStorageKey が同一 key と解決する値も参照として数えるため、
// 候補を SQL で絞り込み（trimming-images/ 部分文字列）→ Go 側で厳密に解決して
// 比較する。
func (r *appointmentTrimmingDetailRepository) ImageKeyInUse(ctx context.Context, clinicID uint64, key string) (bool, error) {
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		if err := acquireTrimmingImageKeyLock(ctx, db, key); err != nil {
			return false, err
		}
	}
	var rows []model.AppointmentTrimmingDetail
	if err := db.
		Model(&model.AppointmentTrimmingDetail{}).
		Select("appointment_trimming_details.style_image", "appointment_trimming_details.completed_image").
		Joins("JOIN appointments ON appointments.id = appointment_trimming_details.appointment_id AND appointments.clinic_id = appointment_trimming_details.clinic_id AND appointments.deleted_at IS NULL").
		Where("appointment_trimming_details.clinic_id = ? AND (appointment_trimming_details.style_image LIKE ? OR appointment_trimming_details.completed_image LIKE ?)",
			clinicID, "%"+trimmingImageStorageKeyPrefix+"%", "%"+trimmingImageStorageKeyPrefix+"%").
		Find(&rows).Error; err != nil {
		return false, apperrors.Wrap(err, "failed to check trimming image reference")
	}
	for i := range rows {
		for _, stored := range []string{rows[i].StyleImage, rows[i].CompletedImage} {
			resolved, isObject, err := resolveTrimmingImageStorageKey(stored)
			if err == nil && isObject && resolved == key {
				return true, nil
			}
		}
	}
	return false, nil
}

// SetOptions は appointment に紐づく trimming オプションを全置換する。
// WithTx コンテキスト内から呼ばれた場合は外側のトランザクションに参加する（savepoint）。
// 単独呼び出しの場合は Clear + Replace を単一トランザクションで実行する。
func (r *appointmentTrimmingDetailRepository) SetOptions(ctx context.Context, clinicID, appointmentID uint64, optionIDs []uint64) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// Options の many2many join には foreignKey:AppointmentID を使うが、GORM の Association()
		// は付随して自モデル行（appointment_trimming_details）の touch-update も行うため、
		// primaryKey(ID) が設定された実インスタンスが必要（#212 の追加発現。AppointmentID のみの
		// ゼロ値インスタンスを渡すと "WHERE conditions required" で失敗する）。
		// clinic_id 述語は P4 の defense-in-depth（checkup_field_repository.go の
		// ReplaceForCheckup / examination_repository.go の ReplaceItemsByExamID と同型パターン）:
		// 呼び出し元（trimming_service.go / liff_service_reservations.go）は appointmentID を
		// 事前に clinic 検証済みだが、repository 層でも再検証し fail-closed にする。
		var detail model.AppointmentTrimmingDetail
		// Parent appointments clinic correlation (SEC-SWEEP-02-TRIM-B1) on target-row
		// lookup only — Clear/Replace write semantics are unchanged.
		if err := tx.Select("appointment_trimming_details.id", "appointment_trimming_details.appointment_id").
			Joins("JOIN appointments ON appointments.id = appointment_trimming_details.appointment_id AND appointments.clinic_id = appointment_trimming_details.clinic_id").
			Where("appointment_trimming_details.appointment_id = ? AND appointment_trimming_details.clinic_id = ?", appointmentID, clinicID).
			First(&detail).Error; err != nil {
			return apperrors.FromGORM(err, "appointment_trimming_detail", fmt.Sprintf("appointment_id=%d", appointmentID))
		}
		if err := tx.Model(&detail).Association("Options").Unscoped().Clear(); err != nil {
			return apperrors.FromGORM(err, "appointment_trimming_option", fmt.Sprintf("appointment_id=%d", appointmentID))
		}
		if len(optionIDs) == 0 {
			return nil
		}
		options := make([]model.TrimmingOption, 0, len(optionIDs))
		for _, id := range optionIDs {
			options = append(options, model.TrimmingOption{ID: id})
		}
		if err := tx.Model(&detail).Association("Options").Replace(options); err != nil {
			return apperrors.FromGORM(err, "appointment_trimming_option", fmt.Sprintf("appointment_id=%d", appointmentID))
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to set trimming options")
	}
	return nil
}
