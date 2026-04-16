package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AppointmentTrimmingDetailRepository はトリミング詳細の CRUD を提供する。
type AppointmentTrimmingDetailRepository interface {
	FindByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error)
	Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	Update(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	SetOptions(ctx context.Context, appointmentID uint64, optionIDs []uint64) error
}

type appointmentTrimmingDetailRepository struct {
	db *gorm.DB
}

func NewAppointmentTrimmingDetailRepository(db *gorm.DB) AppointmentTrimmingDetailRepository {
	return &appointmentTrimmingDetailRepository{db: db}
}

func (r *appointmentTrimmingDetailRepository) FindByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error) {
	var detail model.AppointmentTrimmingDetail
	err := dbOrTx(ctx, r.db).
		Preload("Course").
		Preload("Options").
		Where("clinic_id = ? AND appointment_id = ?", clinicID, appointmentID).
		First(&detail).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment_trimming_detail", fmt.Sprintf("appointment_id=%d", appointmentID))
	}
	return &detail, nil
}

func (r *appointmentTrimmingDetailRepository) Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
	if err := dbOrTx(ctx, r.db).Create(detail).Error; err != nil {
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
	result := dbOrTx(ctx, r.db).
		Model(&model.AppointmentTrimmingDetail{}).
		Where("clinic_id = ? AND appointment_id = ?", detail.ClinicID, detail.AppointmentID).
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
func (r *appointmentTrimmingDetailRepository) SetOptions(ctx context.Context, appointmentID uint64, optionIDs []uint64) error {
	// many2many: joinForeignKey:AppointmentID を使用するため AppointmentID のみで関連を操作できる
	detail := &model.AppointmentTrimmingDetail{AppointmentID: appointmentID}
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(detail).Association("Options").Unscoped().Clear(); err != nil {
			return apperrors.FromGORM(err, "appointment_trimming_option", fmt.Sprintf("appointment_id=%d", appointmentID))
		}
		if len(optionIDs) == 0 {
			return nil
		}
		options := make([]model.TrimmingOption, 0, len(optionIDs))
		for _, id := range optionIDs {
			options = append(options, model.TrimmingOption{ID: id})
		}
		if err := tx.Model(detail).Association("Options").Replace(options); err != nil {
			return apperrors.FromGORM(err, "appointment_trimming_option", fmt.Sprintf("appointment_id=%d", appointmentID))
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to set trimming options")
	}
	return nil
}
