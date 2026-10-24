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
	ImageKeyInUse(ctx context.Context, clinicID uint64, key string, exceptAppointmentID uint64) (bool, error)
}

type appointmentTrimmingDetailRepository struct {
	db *gorm.DB
}

func NewAppointmentTrimmingDetailRepository(db *gorm.DB) AppointmentTrimmingDetailRepository {
	return &appointmentTrimmingDetailRepository{db: db}
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
		Where("appointment_trimming_details.clinic_id = ? AND appointment_trimming_details.appointment_id = ?", clinicID, appointmentID).
		First(&detail).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment_trimming_detail", fmt.Sprintf("appointment_id=%d", appointmentID))
	}
	return &detail, nil
}

func (r *appointmentTrimmingDetailRepository) Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(detail).Error; err != nil {
		return apperrors.FromGORM(err, "appointment_trimming_detail", fmt.Sprintf("appointment_id=%d", detail.AppointmentID))
	}
	return nil
}

// Update は trimming 詳細の可変フィールドをすべて明示的に map で更新する。
// Updates(struct) はゼロ値フィールド（空文字等）をスキップするため、map を使用する。
func (r *appointmentTrimmingDetailRepository) Update(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
	fields := map[string]any{
		"style_request":    detail.StyleRequest,
		"body_weight":      detail.BodyWeight,
		"bw_unit":          bwUnitUpdateValue(detail.BWUnit),
		"body_temperature": detail.BodyTemperature,
		"used_shampoo":     detail.UsedShampoo,
		"used_ribbon":      detail.UsedRibbon,
		"remarks":          detail.Remarks,
		"style_image":      detail.StyleImage,
		"completed_image":  detail.CompletedImage,
	}
	result := persistence.DBOrTx(ctx, r.db).
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

// bwUnitUpdateValue は bw_unit の更新値を返す。NULL（未記録）は NULL のまま書き戻し、"Kg" で補完しない。
func bwUnitUpdateValue(unit *model.BodyWeightUnit) any {
	if unit == nil {
		return nil
	}
	return string(*unit)
}

// ImageKeyInUse はいずれかの trimming 詳細レコードが key を画像として参照して
// いるかを返す（DELETE /trimmings/images と superseded 削除の fail-closed 参照
// チェック用）。exceptAppointmentID != 0 のとき、その appointment が同じ医院で
// 論理削除済みの場合だけ detail 行を除外する（生きている参照は除外しない）。
//
// 保存値は canonical key に正規化されるが、旧世代の `/uploads/` ラップ形式など
// resolveTrimmingImageStorageKey が同一 key と解決する値も参照として数えるため、
// 候補を SQL で絞り込み（trimming-images/ 部分文字列）→ Go 側で厳密に解決して
// 比較する。論理削除済み appointment の detail 行もカウント対象 — 参照残りを
// 消さない方向への過剰保持は許容する。
func (r *appointmentTrimmingDetailRepository) ImageKeyInUse(ctx context.Context, clinicID uint64, key string, exceptAppointmentID uint64) (bool, error) {
	var rows []model.AppointmentTrimmingDetail
	query := persistence.DBOrTx(ctx, r.db).
		Model(&model.AppointmentTrimmingDetail{}).
		Select("style_image", "completed_image").
		Where("clinic_id = ? AND (style_image LIKE ? OR completed_image LIKE ?)",
			clinicID, "%"+trimmingImageStorageKeyPrefix+"%", "%"+trimmingImageStorageKeyPrefix+"%")
	if exceptAppointmentID != 0 {
		query = query.Where(`(appointment_id <> ? OR NOT EXISTS (
			SELECT 1 FROM appointments deleted_image_owner
			WHERE deleted_image_owner.id = appointment_trimming_details.appointment_id
			  AND deleted_image_owner.clinic_id = appointment_trimming_details.clinic_id
			  AND deleted_image_owner.deleted_at IS NOT NULL
		))`, exceptAppointmentID)
	}
	if err := query.Find(&rows).Error; err != nil {
		return false, apperrors.Wrap(err, "failed to check trimming image reference")
	}
	for i := range rows {
		for _, stored := range []string{rows[i].StyleImage, rows[i].CompletedImage} {
			if trimmingImageStoredReferenceMatchesKey(stored, key) {
				return true, nil
			}
		}
	}
	return trimmingCourseImageKeyInUse(ctx, r.db, clinicID, key)
}

// trimmingCourseImageKeyInUse は施術コースマスタの画像参照（EMR-457）を数える。
// Current canonical keys are checked exactly, including soft-deleted courses.
// Replaced/cleared originals remain protected through the durable audit history.
func trimmingCourseImageKeyInUse(ctx context.Context, db *gorm.DB, clinicID uint64, key string) (bool, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, db).
		Unscoped().
		Model(&model.TrimmingCourse{}).
		Where("clinic_id = ? AND image_object_key = ?", clinicID, key).
		Count(&count).Error; err != nil {
		return false, apperrors.Wrap(err, "failed to check trimming course image reference")
	}
	if count > 0 {
		return true, nil
	}
	return trimmingCourseImageHistoryInUse(ctx, db, clinicID, key)
}
