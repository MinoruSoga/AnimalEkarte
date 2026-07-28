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
	if err := persistence.DBOrTx(ctx, r.db).Create(detail).Error; err != nil {
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
