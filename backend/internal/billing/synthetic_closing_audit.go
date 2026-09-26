package billing

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-211 (b) 匿名化 sentinel の識別名。いずれも s09-clinic- / s09-synthetic-
// の teardown 接頭辞に一致させてはいけない — 一致すると次回 teardown の対象に
// なり sentinel 自体が消され、匿名化済みの監査行が再び RESTRICT で詰まる。
// sentinel は login 不能（account_id NULL）・非活性で、audit_logs の RESTRICT
// 参照先として一度だけ find-or-create され、以後すべての synthetic teardown で
// 再利用される。teardown が sentinel を削除することはない。
const (
	syntheticAuditSentinelCompanyName = "emr211-audit-sentinel-company"
	syntheticAuditSentinelClinicName  = "emr211-audit-sentinel-clinic"
	syntheticAuditSentinelStaffName   = "emr211-audit-sentinel-staff"
)

// syntheticAuditSentinelLockKey は sentinel find-or-create を直列化する
// transaction-scoped advisory lock の固定キー（値はこの用途専用の任意の安定定数）。
// 並行する synthetic teardown が sentinel を二重作成しないよう作成区間を排他する。
// xact 版なので lock は teardown tx の commit/rollback で自動解放される。
const syntheticAuditSentinelLockKey int64 = 90211210

// SyntheticClosingAuditAnonymizePolicy は EMR-211 決定 (b) の既定 audit policy。
// audit_logs 行は一切削除せず、削除対象 clinic の staff が actor の行は sentinel
// staff へ、削除対象 clinic 上の行は sentinel clinic へ付け替える。
//   - actor_id: staffs.clinic_id = clinicID の staff が actor の行だけ
//     （switch_clinic で他 clinic 属性になった行を含む）。actor_type は 'staff' のまま。
//   - clinic_id: 削除対象 clinic 上の行だけ（system・実 staff actor を含む）。
//   - action / resource / created_at 等は不変。無関係な行は触らない。
//
// 付け替え対象が無い場合は sentinel を作らず no-op で返る。すべての処理は
// 呼び出し側の teardown tx 内で走り、失敗は全体 rollback となる。
func SyntheticClosingAuditAnonymizePolicy(ctx context.Context, tx *gorm.DB, clinicID uint64) error {
	if tx == nil {
		return apperrors.WrapInvalidInput("audit policy tx is required")
	}
	tx = tx.WithContext(ctx)

	// testdb の AutoMigrate double には audit_logs が無い場合がある。
	// 削除系列と同じく、存在しない対象は no-op として扱う。
	var auditTableExists bool
	if err := tx.Raw("SELECT to_regclass('audit_logs') IS NOT NULL").Scan(&auditTableExists).Error; err != nil {
		return apperrors.Wrap(err, "check audit_logs table")
	}
	if !auditTableExists {
		return nil
	}

	// 再割当が必要な行が無ければ sentinel は作らない。対象は削除対象 clinic 上の
	// 行と、削除対象 clinic の staff が他 clinic に残した行の両方。
	var pending int64
	if err := tx.Raw(`SELECT count(*) FROM audit_logs
		WHERE clinic_id = ? OR actor_id IN (SELECT id FROM staffs WHERE clinic_id = ?)`,
		clinicID, clinicID).Scan(&pending).Error; err != nil {
		return apperrors.Wrap(err, "count synthetic audit_logs")
	}
	if pending == 0 {
		return nil
	}

	if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", syntheticAuditSentinelLockKey).Error; err != nil {
		return apperrors.Wrap(err, "serialize audit sentinel creation")
	}
	sentinel, err := findOrCreateSyntheticAuditSentinel(tx)
	if err != nil {
		return err
	}

	// actor_id: 削除対象 clinic の staff が actor の行だけ（他 clinic 属性の行を含む）。
	if err := tx.Exec(`UPDATE audit_logs SET actor_id = ?
		WHERE actor_id IN (SELECT id FROM staffs WHERE clinic_id = ?)`,
		sentinel.staffID, clinicID).Error; err != nil {
		return apperrors.Wrap(err, "reassign synthetic audit actor")
	}
	// clinic_id: 削除対象 clinic 上の行だけ（system・実 staff actor を含む）。
	if err := tx.Exec("UPDATE audit_logs SET clinic_id = ? WHERE clinic_id = ?",
		sentinel.clinicID, clinicID).Error; err != nil {
		return apperrors.Wrap(err, "reassign synthetic audit clinic")
	}
	return nil
}

type syntheticAuditSentinel struct {
	companyID uint64
	clinicID  uint64
	staffID   uint64
}

// findOrCreateSyntheticAuditSentinel は sentinel company → clinic → staff を
// 一度だけ作り、以後は名前で再利用する。呼び出し側が advisory lock を持つ前提。
func findOrCreateSyntheticAuditSentinel(tx *gorm.DB) (*syntheticAuditSentinel, error) {
	var company model.Company
	err := tx.Where("name = ?", syntheticAuditSentinelCompanyName).Take(&company).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		company = model.Company{Name: syntheticAuditSentinelCompanyName}
		if createErr := tx.Create(&company).Error; createErr != nil {
			return nil, apperrors.Wrap(createErr, "create audit sentinel company")
		}
	default:
		return nil, apperrors.Wrap(err, "load audit sentinel company")
	}

	var clinic model.Clinic
	err = tx.Where("name = ?", syntheticAuditSentinelClinicName).Take(&clinic).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		clinic = model.Clinic{CompanyID: company.ID, Name: syntheticAuditSentinelClinicName}
		if createErr := tx.Create(&clinic).Error; createErr != nil {
			return nil, apperrors.Wrap(createErr, "create audit sentinel clinic")
		}
		// clinics.is_active は gorm:"default:true" 側にあるため Create の zero 値は
		// 明示値として書けず DB 既定（true）になる。sentinel は非活性が必須なので
		// 明示 UPDATE で立て直す。
		if updateErr := tx.Model(&model.Clinic{}).Where("id = ?", clinic.ID).
			Update("is_active", false).Error; updateErr != nil {
			return nil, apperrors.Wrap(updateErr, "disable audit sentinel clinic")
		}
		clinic.IsActive = false
	default:
		return nil, apperrors.Wrap(err, "load audit sentinel clinic")
	}

	var staffRow model.Staff
	err = tx.Where("clinic_id = ? AND name = ?", clinic.ID, syntheticAuditSentinelStaffName).Take(&staffRow).Error
	switch {
	case err == nil:
	case errors.Is(err, gorm.ErrRecordNotFound):
		staffRow = model.Staff{
			ClinicID:  clinic.ID,
			AccountID: nil, // login 不能: sentinel にアカウントは持たせない
			Name:      syntheticAuditSentinelStaffName,
			StaffType: model.StaffTypeResource,
		}
		if createErr := tx.Create(&staffRow).Error; createErr != nil {
			return nil, apperrors.Wrap(createErr, "create audit sentinel staff")
		}
		// staffs.is_active / reservation_visible も default:true の zero 値 trap が
		// あるため明示 UPDATE で false にする。
		if updateErr := tx.Model(&model.Staff{}).Where("id = ?", staffRow.ID).
			Updates(map[string]any{"is_active": false, "reservation_visible": false}).Error; updateErr != nil {
			return nil, apperrors.Wrap(updateErr, "disable audit sentinel staff")
		}
		staffRow.IsActive = false
		staffRow.ReservationVisible = false
	default:
		return nil, apperrors.Wrap(err, "load audit sentinel staff")
	}

	return &syntheticAuditSentinel{companyID: company.ID, clinicID: clinic.ID, staffID: staffRow.ID}, nil
}
