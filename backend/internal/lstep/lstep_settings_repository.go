package lstep

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LstepSettingsRepository は clinic_integrations テーブルへのアクセスインターフェース。
type LstepSettingsRepository interface {
	// FindByClinicAndService は指定クリニック・サービスの全連携設定を返す。
	FindByClinicAndService(ctx context.Context, clinicID uint64, service string) ([]*model.ClinicIntegration, error)
	// FindCredentialByClinicServiceKey は clinic/service/key で canonical credential を1件だけ返す。
	FindCredentialByClinicServiceKey(ctx context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error)
	// Upsert は key_name ごとに UPSERT する。
	Upsert(ctx context.Context, integration *model.ClinicIntegration) error
	// DeleteByClinicAndService は指定クリニック・サービスの全設定を削除する。
	DeleteByClinicAndService(ctx context.Context, clinicID uint64, service string) error
	// DeleteByClinicServiceAndKey は指定 key 1件だけ削除する（LIFF ID クリア用）。
	DeleteByClinicServiceAndKey(ctx context.Context, clinicID uint64, service, keyName string) error
}

type lstepSettingsRepository struct{ db *gorm.DB }

func (r *lstepSettingsRepository) FindCredentialByClinicServiceKey(
	ctx context.Context,
	clinicID uint64,
	service, keyName string,
) (*model.ClinicIntegration, error) {
	var record model.ClinicIntegration
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("service = ? AND key_name = ?", service, keyName).
		Take(&record).Error
	if err != nil {
		return nil, apperrors.FromGORM(
			err,
			"clinic_integrations",
			fmt.Sprintf("clinic=%d service=%s key=%s", clinicID, service, keyName),
		)
	}
	return &record, nil
}

// NewLstepSettingsRepository は LstepSettingsRepository を初期化して返す。
func NewLstepSettingsRepository(db *gorm.DB) LstepSettingsRepository {
	return &lstepSettingsRepository{db: db}
}

func (r *lstepSettingsRepository) FindByClinicAndService(ctx context.Context, clinicID uint64, service string) ([]*model.ClinicIntegration, error) {
	var records []*model.ClinicIntegration
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("service = ?", service).
		Find(&records).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic_integrations", fmt.Sprintf("clinic=%d service=%s", clinicID, service))
	}
	return records, nil
}

func (r *lstepSettingsRepository) Upsert(ctx context.Context, integration *model.ClinicIntegration) error {
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(integration.ClinicID)).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "service"}, {Name: "key_name"}},
			DoUpdates: clause.AssignmentColumns([]string{"key_value", "updated_at"}),
		}).
		Create(integration).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_integrations", fmt.Sprintf("clinic=%d key=%s", integration.ClinicID, integration.KeyName))
	}
	return nil
}

func (r *lstepSettingsRepository) DeleteByClinicAndService(ctx context.Context, clinicID uint64, service string) error {
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("service = ?", service).
		Delete(&model.ClinicIntegration{}).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_integrations", fmt.Sprintf("clinic=%d service=%s", clinicID, service))
	}
	return nil
}

func (r *lstepSettingsRepository) DeleteByClinicServiceAndKey(ctx context.Context, clinicID uint64, service, keyName string) error {
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("service = ? AND key_name = ?", service, keyName).
		Delete(&model.ClinicIntegration{}).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic_integrations", fmt.Sprintf("clinic=%d service=%s key=%s", clinicID, service, keyName))
	}
	return nil
}
