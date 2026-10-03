package clinicale2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	padPetCount            = 20
	medicalRecordCount     = 21
	ownerSearchToken       = "e2e-owner"
	outsideFirstPagePrefix = "e2e-zebra-"
	v04GroupName           = "e2e-v04-view-only"
	v04RowPrefix           = "V04-e2e-"
	maxTeardownPasses      = 8
	pgForeignKeyViolation  = "23503"
	pgRestrictViolation    = "23001"
)

// v04ViewOnlyResources は V04 設定/マスタ画面で view のみ許可する resource。
// accounting/campaign 系と master-permission・master-staff は意図的に除外する。
var v04ViewOnlyResources = []model.Resource{
	model.ResourceMasterAnimalSpecies,
	model.ResourceMasterMedical,
	model.ResourceMasterReservationType,
	model.ResourceMasterHospitalization,
	model.ResourceMasterTrimming,
	model.ResourceMasterInsurance,
	model.ResourceMasterMerchandise,
	model.ResourcePaymentMethod,
	model.ResourceClosingSettings,
	model.ResourceShifts,
	model.ResourceHospitalSettings,
	model.ResourceLabImport,
}

// Request は disposable clinic を作る入力。PasswordHash はログに出さない。
type Request struct {
	AppEnv       string
	DBHost       string
	PasswordHash string
}

// V04Result は view-only account と V04 行の参照情報。clinicId は含めない
// （run-e2e.sh が JSON 全体から clinicId を greedy に抽出するため、重複キーを埋め込まない）。
type V04Result struct {
	ViewOnlyEmail     string   `json:"viewOnlyEmail"`
	ViewOnlyResources []string `json:"viewOnlyResources"`
	CageName          string   `json:"cageName"`
	LabDeviceName     string   `json:"labDeviceName"`
}

// Result は Playwright が参照する合成 ID / 氏名。秘密は含めない。
type Result struct {
	ClinicID                uint64    `json:"clinicId"`
	OwnerName               string    `json:"ownerName"`
	OwnerSearch             string    `json:"ownerSearch"`
	PetID                   uint64    `json:"petId"`
	PetName                 string    `json:"petName"`
	OutsideFirstPagePetID   uint64    `json:"outsideFirstPagePetId"`
	OutsideFirstPagePetName string    `json:"outsideFirstPagePetName"`
	EstimateTitle           string    `json:"estimateTitle"`
	MedicalRecordCount      int       `json:"medicalRecordCount"`
	V04                     V04Result `json:"v04"`
}

// Create は新規 clinic / staff / owner / pet / 確定カルテと allowlist 用の行を INSERT する。
func Create(ctx context.Context, db *gorm.DB, req Request) (*Result, error) {
	if err := Allow(req.AppEnv, req.DBHost); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}
	if strings.TrimSpace(req.PasswordHash) == "" {
		return nil, apperrors.WrapInvalidInput("password hash is required")
	}
	if db == nil {
		return nil, apperrors.WrapInvalidInput("db is required")
	}

	clinicID := clinicIDBase + uint64(time.Now().UnixNano()%8000)
	if err := RejectReservedClinicID(clinicID); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, apperrors.Wrap(err, "load Asia/Tokyo")
	}
	day := time.Now().In(jst)
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, jst)

	var result *Result
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		company := &model.Company{Name: fmt.Sprintf("%s%d", companyNamePrefix, clinicID)}
		if err := tx.Create(company).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic company")
		}
		clinic := &model.Clinic{
			ID:        clinicID,
			CompanyID: company.ID,
			Name:      fmt.Sprintf("%s%d", clinicNamePrefix, clinicID),
			IsActive:  true,
		}
		if err := tx.Create(clinic).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic clinic")
		}

		account := &model.Account{
			Email:         LoginEmail(clinicID),
			PasswordHash:  req.PasswordHash,
			IsActive:      true,
			IsSystemAdmin: true,
		}
		if err := tx.Create(account).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic account")
		}
		staff := &model.Staff{
			ClinicID:  clinicID,
			AccountID: &account.ID,
			Name:      fmt.Sprintf("e2e-staff-%d", clinicID),
			IsActive:  true,
			StaffType: model.StaffTypeDoctor,
		}
		if err := tx.Create(staff).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic staff")
		}
		assignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID, IsMain: true}
		if err := tx.Create(assignment).Error; err != nil {
			return apperrors.Wrap(err, "assign synthetic staff clinic")
		}

		owner := &model.Owner{
			ClinicID: clinicID,
			Name:     fmt.Sprintf("%s-%d", ownerSearchToken, clinicID),
			NameKana: ownerSearchToken,
		}
		if err := tx.Create(owner).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic owner")
		}
		species := &model.AnimalSpecies{Name: fmt.Sprintf("e2e-species-%d", clinicID)}
		if err := tx.Create(species).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic species")
		}

		mainPet := &model.Pet{
			ClinicID:        clinicID,
			OwnerID:         owner.ID,
			AnimalSpeciesID: species.ID,
			Name:            fmt.Sprintf("e2e-pet-%d", clinicID),
		}
		if err := tx.Create(mainPet).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic pet")
		}
		for i := 0; i < padPetCount; i++ {
			pad := &model.Pet{
				ClinicID:        clinicID,
				OwnerID:         owner.ID,
				AnimalSpeciesID: species.ID,
				Name:            fmt.Sprintf("e2e-a%02d-%d", i, clinicID),
			}
			if err := tx.Create(pad).Error; err != nil {
				return apperrors.Wrap(err, "create pad pet")
			}
		}
		outsidePet := &model.Pet{
			ClinicID:        clinicID,
			OwnerID:         owner.ID,
			AnimalSpeciesID: species.ID,
			Name:            fmt.Sprintf("%s%d", outsideFirstPagePrefix, clinicID),
		}
		if err := tx.Create(outsidePet).Error; err != nil {
			return apperrors.Wrap(err, "create outside-first-page pet")
		}

		for i := 0; i < medicalRecordCount; i++ {
			petID := mainPet.ID
			record := &model.MedicalRecord{
				ClinicID: clinicID,
				RecordNo: fmt.Sprintf("E2E-%d-%02d", clinicID, i+1),
				Date:     day,
				OwnerID:  &owner.ID,
				PetID:    &petID,
				DoctorID: &staff.ID,
				Status:   model.MedicalRecordStatusFinalized,
			}
			if err := tx.Create(record).Error; err != nil {
				return apperrors.Wrap(err, "create synthetic medical record")
			}
		}

		var firstRecord model.MedicalRecord
		if err := tx.Where("clinic_id = ?", clinicID).Order("id ASC").First(&firstRecord).Error; err != nil {
			return apperrors.Wrap(err, "load first synthetic medical record")
		}

		examType := &model.ExaminationType{ClinicID: clinicID, Name: fmt.Sprintf("e2e-exam-%d", clinicID), IsActive: true}
		if err := tx.Create(examType).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic exam type")
		}
		exam := &model.Examination{
			ClinicID:        clinicID,
			MedicalRecordID: &firstRecord.ID,
			PetID:           &mainPet.ID,
			ExamTypeID:      examType.ID,
			DoctorID:        &staff.ID,
			Date:            day,
			Status:          model.ExaminationStatusCompleted,
		}
		if err := tx.Create(exam).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic examination")
		}
		// medicalRecordID なし行は一覧の standalone 詳細遷移（/examinations/:id）を担う。
		standaloneExam := &model.Examination{
			ClinicID:   clinicID,
			PetID:      &outsidePet.ID,
			ExamTypeID: examType.ID,
			DoctorID:   &staff.ID,
			Date:       day,
			Status:     model.ExaminationStatusCompleted,
		}
		if err := tx.Create(standaloneExam).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic standalone examination")
		}

		vaccine := &model.Vaccine{ClinicID: clinicID, Name: fmt.Sprintf("e2e-vac-%d", clinicID), IsActive: true}
		if err := tx.Create(vaccine).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic vaccine")
		}
		vaccination := &model.Vaccination{
			ClinicID:        clinicID,
			MedicalRecordID: &firstRecord.ID,
			PetID:           &mainPet.ID,
			VaccineID:       vaccine.ID,
			Date:            day,
			DoctorID:        &staff.ID,
		}
		if err := tx.Create(vaccination).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic vaccination")
		}
		// medicalRecordID なし行は一覧の standalone 詳細遷移（/vaccinations/:id）を担う。
		standaloneVaccination := &model.Vaccination{
			ClinicID:  clinicID,
			PetID:     &outsidePet.ID,
			VaccineID: vaccine.ID,
			Date:      day,
			DoctorID:  &staff.ID,
		}
		if err := tx.Create(standaloneVaccination).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic standalone vaccination")
		}

		checkupType := &model.CheckupType{ClinicID: clinicID, Name: fmt.Sprintf("e2e-chk-%d", clinicID), IsActive: true}
		if err := tx.Create(checkupType).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic checkup type")
		}
		checkup := &model.Checkup{
			ClinicID:        clinicID,
			MedicalRecordID: firstRecord.ID,
			PetID:           &mainPet.ID,
			CheckupTypeID:   checkupType.ID,
			Date:            day,
			DoctorID:        &staff.ID,
		}
		if err := tx.Create(checkup).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic checkup")
		}

		hospitalization := &model.Hospitalization{
			ClinicID:            clinicID,
			OwnerID:             owner.ID,
			PetID:               mainPet.ID,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           day,
			EndDate:             day.Add(24 * time.Hour),
			Status:              model.HospitalizationStatusAdmitted,
			DoctorID:            &staff.ID,
		}
		if err := tx.Create(hospitalization).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic hospitalization")
		}

		estimateTitle := fmt.Sprintf("e2e-est-%d", clinicID)
		estimate := &model.Estimate{
			ClinicID:        clinicID,
			EstimateNo:      fmt.Sprintf("E2-%d", clinicID),
			MedicalRecordID: &firstRecord.ID,
			Title:           estimateTitle,
			OwnerID:         &owner.ID,
			PetID:           &mainPet.ID,
			Status:          model.EstimateStatusDraft,
			CreatedBy:       &staff.ID,
		}
		if err := tx.Create(estimate).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic estimate")
		}

		v04, err := createV04Fixture(tx, clinicID, req.PasswordHash)
		if err != nil {
			return err
		}

		result = &Result{
			ClinicID:                clinicID,
			OwnerName:               owner.Name,
			OwnerSearch:             ownerSearchToken,
			PetID:                   mainPet.ID,
			PetName:                 mainPet.Name,
			OutsideFirstPagePetID:   outsidePet.ID,
			OutsideFirstPagePetName: outsidePet.Name,
			EstimateTitle:           estimateTitle,
			MedicalRecordCount:      medicalRecordCount,
			V04:                     v04,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// createV04Fixture は V04 settings/master view-only 検証用の account・権限グループ・
// 一覧表示用の cage / lab device 行を作る。admin と同じ bcrypt hash を再利用する。
func createV04Fixture(tx *gorm.DB, clinicID uint64, passwordHash string) (V04Result, error) {
	account := &model.Account{
		Email:         ViewOnlyLoginEmail(clinicID),
		PasswordHash:  passwordHash,
		IsActive:      true,
		IsSystemAdmin: false,
	}
	if err := tx.Create(account).Error; err != nil {
		return V04Result{}, apperrors.Wrap(err, "create v04 view-only account")
	}
	staff := &model.Staff{
		ClinicID:  clinicID,
		AccountID: &account.ID,
		Name:      fmt.Sprintf("e2e-v04-view-%d", clinicID),
		IsActive:  true,
		StaffType: model.StaffTypeNurse,
	}
	if err := tx.Create(staff).Error; err != nil {
		return V04Result{}, apperrors.Wrap(err, "create v04 view-only staff")
	}
	assignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID, IsMain: true}
	if err := tx.Create(assignment).Error; err != nil {
		return V04Result{}, apperrors.Wrap(err, "assign v04 view-only staff clinic")
	}

	group := &model.PermissionGroup{
		ClinicID: clinicID,
		Name:     v04GroupName,
		IsActive: true,
	}
	if err := tx.Create(group).Error; err != nil {
		return V04Result{}, apperrors.Wrap(err, "create v04 view-only permission group")
	}
	resources := make([]string, 0, len(v04ViewOnlyResources))
	for _, resource := range v04ViewOnlyResources {
		rule := &model.PermissionGroupRule{
			GroupID:   group.ID,
			Resource:  string(resource),
			CanView:   true,
			CanCreate: false,
			CanEdit:   false,
			CanDelete: false,
		}
		if err := tx.Create(rule).Error; err != nil {
			return V04Result{}, apperrors.Wrap(err, "create v04 view-only permission rule")
		}
		resources = append(resources, string(resource))
	}
	link := &model.StaffPermissionGroup{StaffID: staff.ID, GroupID: group.ID}
	if err := tx.Create(link).Error; err != nil {
		return V04Result{}, apperrors.Wrap(err, "link v04 view-only staff to group")
	}

	cage := &model.Cage{
		ClinicID: clinicID,
		Name:     fmt.Sprintf("%scage-%d", v04RowPrefix, clinicID),
		IsActive: true,
		CageType: model.CageTypeGeneral,
		CageSize: model.CageSizeMedium,
	}
	if err := tx.Create(cage).Error; err != nil {
		return V04Result{}, apperrors.Wrap(err, "create v04 cage")
	}
	// exam_type_id は意図的に nil — 臨床 fixture の検査種別へリンクすると
	// examinations flow の挙動が変わりうるため。
	device := &model.LabDevice{
		ClinicID:   clinicID,
		SourceType: string(model.LabImportSourceTypeFujiNX600),
		Name:       fmt.Sprintf("%sdevice-%d", v04RowPrefix, clinicID),
		IsActive:   true,
	}
	if err := tx.Create(device).Error; err != nil {
		return V04Result{}, apperrors.Wrap(err, "create v04 lab device")
	}

	return V04Result{
		ViewOnlyEmail:     account.Email,
		ViewOnlyResources: resources,
		CageName:          cage.Name,
		LabDeviceName:     device.Name,
	}, nil
}

// Delete は合成 clinic とその子孫だけを消す。clinic 1/2 と接頭辞不一致は拒否する。
func Delete(ctx context.Context, db *gorm.DB, appEnv, dbHost string, clinicID uint64) error {
	if err := Allow(appEnv, dbHost); err != nil {
		return apperrors.WrapInvalidInput(err.Error())
	}
	if err := RejectReservedClinicID(clinicID); err != nil {
		return apperrors.WrapInvalidInput(err.Error())
	}
	if db == nil {
		return apperrors.WrapInvalidInput("db is required")
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var clinic model.Clinic
		if err := tx.First(&clinic, clinicID).Error; err != nil {
			return apperrors.Wrap(err, "load synthetic clinic")
		}
		if !strings.HasPrefix(clinic.Name, clinicNamePrefix) {
			return apperrors.WrapInvalidInput("clinic name is not a clinical e2e fixture")
		}

		var staffs []model.Staff
		if err := tx.Unscoped().Where("clinic_id = ?", clinicID).Find(&staffs).Error; err != nil {
			return apperrors.Wrap(err, "list synthetic staff")
		}
		accountIDs := make([]uint64, 0, len(staffs))
		staffIDs := make([]uint64, 0, len(staffs))
		for _, staff := range staffs {
			staffIDs = append(staffIDs, staff.ID)
			if staff.AccountID != nil {
				accountIDs = append(accountIDs, *staff.AccountID)
			}
		}

		var estimates []model.Estimate
		if err := tx.Unscoped().Where("clinic_id = ?", clinicID).Find(&estimates).Error; err != nil {
			return apperrors.Wrap(err, "list synthetic estimates")
		}
		estimateIDs := make([]uint64, 0, len(estimates))
		for _, estimate := range estimates {
			estimateIDs = append(estimateIDs, estimate.ID)
		}

		var groups []model.PermissionGroup
		if err := tx.Unscoped().Where("clinic_id = ?", clinicID).Find(&groups).Error; err != nil {
			return apperrors.Wrap(err, "list synthetic permission groups")
		}
		groupIDs := make([]uint64, 0, len(groups))
		for _, group := range groups {
			groupIDs = append(groupIDs, group.ID)
		}

		// testdb は AutoMigrate 構築で FK CASCADE を期待できないため、staff/group 削除前に
		// 中間テーブルと rule を明示削除する。
		if len(staffIDs) > 0 || len(groupIDs) > 0 {
			linkTx := tx.Unscoped()
			switch {
			case len(staffIDs) > 0 && len(groupIDs) > 0:
				linkTx = linkTx.Where("staff_id IN ? OR group_id IN ?", staffIDs, groupIDs)
			case len(staffIDs) > 0:
				linkTx = linkTx.Where("staff_id IN ?", staffIDs)
			default:
				linkTx = linkTx.Where("group_id IN ?", groupIDs)
			}
			if err := linkTx.Delete(&model.StaffPermissionGroup{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic staff permission group links")
			}
		}
		if len(groupIDs) > 0 {
			if err := tx.Unscoped().Where("group_id IN ?", groupIDs).Delete(&model.PermissionGroupRule{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic permission group rules")
			}
		}

		if len(estimateIDs) > 0 {
			if err := tx.Unscoped().Where("estimate_id IN ?", estimateIDs).Delete(&model.EstimateItem{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic estimate items")
			}
		}
		if len(staffIDs) > 0 {
			if err := tx.Where("clinic_id = ? AND actor_type = ? AND actor_id IN ?", clinicID, model.AuditActorTypeStaff, staffIDs).Delete(&model.AuditLog{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic fixture staff audit logs")
			}
		}
		// clinic_id を持つ全 public テーブルを対象に物理削除する。v04/clinical spec が
		// UI 経由で作成したマスタ行（職種・予約区分・診療項目・支払方法など）は soft-delete
		// でも行が残り、clinics への RESTRICT FK で親削除が失敗するため、固定モデル列挙では
		// 追従できない。子→親の順序は RESTRICT/FK 違反 (23001/23503) を検出してリトライすることで
		// FK グラフ (DAG) 上で収束させる。clinic_id を持たない子行（estimate_items 等）は
		// 上の明示削除に残す。
		if err := deleteClinicScopedRows(tx, clinicID); err != nil {
			return err
		}
		if len(accountIDs) > 0 {
			if err := tx.Unscoped().Where("id IN ?", accountIDs).Delete(&model.Account{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic accounts")
			}
		}
		if err := tx.Unscoped().Where("name LIKE ?", fmt.Sprintf("e2e-species-%d", clinicID)).Delete(&model.AnimalSpecies{}).Error; err != nil {
			return apperrors.Wrap(err, "delete synthetic species")
		}
		if err := tx.Delete(&clinic).Error; err != nil {
			return apperrors.Wrap(err, "delete synthetic clinic")
		}
		if clinic.CompanyID != 0 {
			if err := tx.Where("id = ? AND name LIKE ?", clinic.CompanyID, companyNamePrefix+"%").Delete(&model.Company{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic company")
			}
		}
		return nil
	})
}

// purgeTarget は teardown で物理削除するテーブルと、その行を特定する単一引数 predicate。
// scoped テーブルは `clinic_id = ?`、clinic_id を持たない子テーブルは
// `<fk> IN (SELECT <pk> FROM <parent> WHERE clinic_id = ?)` を使う。
type purgeTarget struct {
	table string
	where string
}

// deleteClinicScopedRows は対象 clinic に属する行を全て物理削除する。
//   - clinic_id 列を持つ public テーブル: clinic_id で直接削除
//   - clinic_id を持たない depth-1 子テーブル（RESTRICT / NO ACTION で clinic スコープの
//     親を参照するもの。care_plan_items / estimate_items / staff_notes 等）:
//     親の clinic_id を辿るサブクエリで削除
//
// RESTRICT (23001) / FK (23503) 違反は親が後のパスで消える前提でリトライし、
// FK グラフ (DAG) 上で収束させる。1 パスで進捗ゼロなら残余 blocker テーブル名を
// 添えて失敗させる。
func deleteClinicScopedRows(tx *gorm.DB, clinicID uint64) error {
	var scoped []string
	if err := tx.Raw(`SELECT table_name
		FROM information_schema.columns
		WHERE table_schema = 'public' AND column_name = 'clinic_id'
		ORDER BY table_name`).Scan(&scoped).Error; err != nil {
		return apperrors.Wrap(err, "list clinic-scoped tables")
	}
	scopedSet := make(map[string]bool, len(scoped))
	targets := make([]purgeTarget, 0, len(scoped))
	for _, table := range scoped {
		scopedSet[table] = true
		targets = append(targets, purgeTarget{table: table, where: "clinic_id = ?"})
	}

	// 単一列 FK のみ対象（conkey の array_length=1）。RESTRICT/NO ACTION のみ列挙する
	// のは CASCADE 子は親削除に追従するため。複数親を持つ子は OR で畳む。
	type fkLink struct {
		child, childCol, parent, parentCol string
	}
	var links []fkLink
	if err := tx.Raw(`SELECT
			child.relname, child_col.attname, parent.relname, parent_col.attname
		FROM pg_constraint con
		JOIN pg_class child ON child.oid = con.conrelid
		JOIN pg_class parent ON parent.oid = con.confrelid
		JOIN pg_namespace n ON n.oid = con.connamespace
		JOIN pg_attribute child_col ON child_col.attrelid = con.conrelid AND child_col.attnum = con.conkey[1]
		JOIN pg_attribute parent_col ON parent_col.attrelid = con.confrelid AND parent_col.attnum = con.confkey[1]
		WHERE con.contype = 'f' AND n.nspname = 'public'
			AND con.confdeltype IN ('r', 'n')
			AND array_length(con.conkey, 1) = 1`).Scan(&links).Error; err != nil {
		return apperrors.Wrap(err, "list clinic-scoped child links")
	}
	childWheres := make(map[string][]string)
	for _, link := range links {
		if !scopedSet[link.parent] || scopedSet[link.child] {
			continue
		}
		childWheres[link.child] = append(childWheres[link.child], fmt.Sprintf(
			`%q IN (SELECT %q FROM %q WHERE clinic_id = ?)`,
			link.childCol, link.parentCol, link.parent))
	}
	for _, table := range sortedKeys(childWheres) {
		targets = append(targets, purgeTarget{table: table, where: strings.Join(childWheres[table], " OR ")})
	}

	for pass := 0; len(targets) > 0 && pass < maxTeardownPasses; pass++ {
		blocked := make([]purgeTarget, 0, len(targets))
		for _, target := range targets {
			if err := tx.Exec("SAVEPOINT clinic_scoped_delete").Error; err != nil {
				return apperrors.Wrap(err, "teardown savepoint")
			}
			args := make([]any, strings.Count(target.where, "?"))
			for i := range args {
				args[i] = clinicID
			}
			err := tx.Exec(`DELETE FROM "`+target.table+`" WHERE `+target.where, args...).Error
			if err == nil {
				if relErr := tx.Exec("RELEASE SAVEPOINT clinic_scoped_delete").Error; relErr != nil {
					return apperrors.Wrap(relErr, "teardown release savepoint")
				}
				continue
			}
			if rbErr := tx.Exec("ROLLBACK TO SAVEPOINT clinic_scoped_delete").Error; rbErr != nil {
				return apperrors.Wrap(rbErr, "teardown rollback savepoint")
			}
			var pgErr *pgconn.PgError
			// ON DELETE RESTRICT は 23503 ではなく 23001 (restrict_violation) を返す。
			// 親が後続パスで消えれば再試行で通るため、両方を blocked 扱いにして収束させる。
			if errors.As(err, &pgErr) && (pgErr.Code == pgForeignKeyViolation || pgErr.Code == pgRestrictViolation) {
				blocked = append(blocked, target)
				continue
			}
			return apperrors.Wrap(err, fmt.Sprintf("delete synthetic clinic-scoped rows from %s", target.table))
		}
		if len(blocked) == len(targets) {
			names := make([]string, 0, len(blocked))
			for _, target := range blocked {
				names = append(names, target.table)
			}
			return fmt.Errorf("clinic-scoped teardown stalled on FK-restricted tables: %s", strings.Join(names, ", "))
		}
		targets = blocked
	}
	if len(targets) > 0 {
		names := make([]string, 0, len(targets))
		for _, target := range targets {
			names = append(names, target.table)
		}
		return fmt.Errorf("clinic-scoped teardown did not converge: %s", strings.Join(names, ", "))
	}
	return nil
}

// sortedKeys は map のキーを決定的順序で返す（テーブル削除順を再現可能にするため）。
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// EncodeResult は stdout 用の 1 行 JSON。秘密は載らない。
func EncodeResult(result *Result) ([]byte, error) {
	if result == nil {
		return nil, apperrors.WrapInvalidInput("result is required")
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return nil, apperrors.Wrap(err, "encode clinical fixture")
	}
	return payload, nil
}
