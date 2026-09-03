package clinic

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

type clinicRepository struct {
	db *gorm.DB
}

func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return &clinicRepository{db: db}
}

func (r *clinicRepository) FindAll(ctx context.Context) ([]model.Clinic, error) {
	clinics := make([]model.Clinic, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Order("name ASC").
		Find(&clinics).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", "")
	}
	return clinics, nil
}

func (r *clinicRepository) FindByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error) {
	clinics := make([]model.Clinic, 0)
	err := r.db.WithContext(ctx).
		Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.clinic_id = clinics.id"+
			" AND staff_clinic_assignments.deleted_at IS NULL").
		Where("staff_clinic_assignments.staff_id = ?", staffID).
		Order("clinics.name ASC").
		Find(&clinics).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", fmt.Sprintf("staff_id=%d", staffID))
	}
	return clinics, nil
}

// LockActiveByID holds a SHARE lock on an active clinic until the caller's
// transaction ends. It fails closed without an ambient transaction.
func (r *clinicRepository) LockActiveByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("clinic lock requires an active transaction")
	}
	var clinic model.Clinic
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id = ? AND is_active = ?", id, true).
		First(&clinic).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", fmt.Sprintf("%d", id))
	}
	return &clinic, nil
}

// LockByIDForUpdate holds an UPDATE lock on a clinic until the caller's
// transaction ends. It intentionally includes inactive clinics because
// deactivation followed by physical deletion is an existing supported flow.
func (r *clinicRepository) LockByIDForUpdate(ctx context.Context, id uint64) (*model.Clinic, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("clinic update lock requires an active transaction")
	}
	var clinic model.Clinic
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&clinic).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", fmt.Sprintf("%d", id))
	}
	return &clinic, nil
}

func (r *clinicRepository) FindByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	var clinic model.Clinic
	err := persistence.DBOrTx(ctx, r.db).First(&clinic, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", fmt.Sprintf("%d", id))
	}
	return &clinic, nil
}

func (r *clinicRepository) FindCompany(ctx context.Context) (*model.Company, error) {
	var company model.Company
	err := persistence.DBOrTx(ctx, r.db).First(&company).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "company", "singleton")
	}
	return &company, nil
}

// BE-refactor.md X-7: dbOrTx で ambient tx に参加する。CreateClinic は clinic 作成 +
// デフォルト権限グループ2件の作成を transactor.WithTx で包むが、Create が r.db.WithContext(ctx)
// のまま tx 非参加だと、途中で失敗しても既にオートコミット済みの行は WithTx のロールバックで
// 巻き戻らず、デフォルト権限グループなしの孤児クリニックが生成しうるバグがあった。
func (r *clinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {
	err := persistence.DBOrTx(ctx, r.db).Create(clinic).Error
	if err != nil {
		if persistence.IsUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("clinic", clinic.Name)
		}
		return apperrors.FromGORM(err, "clinic", "")
	}
	return nil
}

func (r *clinicRepository) UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) error {
	if input == nil {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	fields, err := BuildClinicUpdate(input)
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return nil
	}
	return r.Update(ctx, id, fields)
}

// Update applies a BuildClinicUpdate map inside the caller's ambient transaction.
// Consumers must call UpdateClinic; this method stays for the DBOrTx inventory key.
func (r *clinicRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := persistence.DBOrTx(ctx, r.db).Model(&model.Clinic{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinic", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", id))
	}
	return nil
}

// Delete participates in the caller's ambient transaction. Permission-group cleanup
// is owned by PermissionGroupRepository and must be orchestrated by the service in
// the same Transactor.WithTx callback before this delete.
func (r *clinicRepository) Delete(ctx context.Context, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).Delete(&model.Clinic{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinic", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *clinicRepository) CountOwnersByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Owner{}).
		Scopes(persistence.ClinicScope(clinicID)).Where("deleted_at IS NULL").
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "owner", fmt.Sprintf("clinic_id=%d", clinicID))
	}
	return count, nil
}

func (r *clinicRepository) CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Staff{}).
		Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ? AND staff_clinic_assignments.deleted_at IS NULL", clinicID).
		Where("staffs.deleted_at IS NULL").
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "staff", fmt.Sprintf("clinic_id=%d", clinicID))
	}
	return count, nil
}

func (r *clinicRepository) CountBlockingReferencesByClinicID(ctx context.Context, clinicID uint64) ([]ClinicDependencyCount, error) {
	checks := []struct {
		table   string
		label   string
		softDel bool
	}{
		{table: "appointments", label: "予約", softDel: true},
		{table: "medical_records", label: "カルテ", softDel: true},
		{table: "hospitalizations", label: "入院記録", softDel: true},
		{table: "exams", label: "検査", softDel: true},
		{table: "vaccinations", label: "ワクチン接種記録", softDel: true},
		{table: "checkups", label: "健診記録", softDel: true},
		{table: "billings", label: "会計", softDel: true},
		{table: "clinic_settings", label: "医院設定", softDel: false},
		{table: "clinic_integrations", label: "連携設定", softDel: false},
		{table: "lstep_settings", label: "Lステップ設定", softDel: false},
		{table: "permission_groups", label: "権限グループ", softDel: true},
	}

	dependencies := make([]ClinicDependencyCount, 0, len(checks))
	for _, check := range checks {
		query := "clinic_id = ?"
		if check.softDel {
			query += " AND deleted_at IS NULL"
		}

		var count int64
		if err := persistence.DBOrTx(ctx, r.db).
			Table(check.table).
			Where(query, clinicID).
			Count(&count).Error; err != nil {
			return nil, apperrors.FromGORM(err, check.table, fmt.Sprintf("clinic_id=%d", clinicID))
		}
		if count > 0 {
			dependencies = append(dependencies, ClinicDependencyCount{Label: check.label, Count: count})
		}
	}
	return dependencies, nil
}
