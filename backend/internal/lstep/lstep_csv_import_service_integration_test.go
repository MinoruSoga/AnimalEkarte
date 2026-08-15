package lstep

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

type csvImportTestStaffRepository struct{ accountID uint64 }

func (r *csvImportTestStaffRepository) FindByID(_ context.Context, clinicID, staffID uint64) (*model.Staff, error) {
	accountID := r.accountID
	return &model.Staff{ID: staffID, ClinicID: clinicID, AccountID: &accountID}, nil
}

func TestImportFriendAttributesCSV_Integration_HappyPath(t *testing.T) {
	db := setupLstepCsvImportServiceTestDB(t)
	ctx := context.Background()
	clinicID := seedLstepCsvImportClinic(t, db)

	lineUserID := "U_line_linked_001"
	if err := db.Create(&model.Owner{
		ClinicID:   clinicID,
		Name:       "テスト飼主",
		NameKana:   "テストカイヌシ",
		LineUserID: &lineUserID,
	}).Error; err != nil {
		t.Fatalf("failed to seed owner: %v", err)
	}

	svc := NewLstepCsvImportService(
		db,
		NewLstepCsvImportRepository(db),
		NewLstepCSVImportOwnerLookup(),
		&csvImportTestStaffRepository{accountID: 101},
	)

	csv := strings.Join([]string{
		"LINE ID,表示名,登録日,タグ,シナリオ,流入元,ブロック,最終メッセージ",
		"U_line_linked_001,山田太郎,2026-05-01,tagA;tagB,scenario1;scenario2,Instagram,active,2026-05-10 09:30:00",
		"U_unknown_999,未紐付け,2026-05-02,tagX,scenarioX,Web,blocked,2026-05-11 10:00:00",
	}, "\n")

	imp, err := svc.ImportFriendAttributesCSV(ctx, clinicID, "friend-attributes.csv", strings.NewReader(csv), 1)
	if err != nil {
		t.Fatalf("ImportFriendAttributesCSV returned error: %v", err)
	}

	if imp.Status != csvImportStatusCompleted {
		t.Fatalf("status = %s, want %s", imp.Status, csvImportStatusCompleted)
	}
	if imp.RowCount != 2 || imp.SuccessCount != 1 || imp.ErrorCount != 1 {
		t.Fatalf("unexpected counts: row=%d success=%d error=%d", imp.RowCount, imp.SuccessCount, imp.ErrorCount)
	}
	if imp.ImportedAt == nil {
		t.Fatal("ImportedAt is nil")
	}

	var persistedImport model.LstepCsvImport
	if err := db.Where("clinic_id = ? AND id = ?", clinicID, imp.ID).First(&persistedImport).Error; err != nil {
		t.Fatalf("failed to load csv import: %v", err)
	}
	if persistedImport.Status != csvImportStatusCompleted {
		t.Fatalf("persisted import status = %s, want %s", persistedImport.Status, csvImportStatusCompleted)
	}

	var snapshots []model.LstepFriendAttributeSnapshot
	if err := db.Where("clinic_id = ?", clinicID).Find(&snapshots).Error; err != nil {
		t.Fatalf("failed to load snapshots: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("snapshot count = %d, want 1", len(snapshots))
	}

	snapshot := snapshots[0]
	if snapshot.LineUserID != lineUserID {
		t.Fatalf("snapshot line_user_id = %s, want %s", snapshot.LineUserID, lineUserID)
	}
	if snapshot.CsvImportID == nil || *snapshot.CsvImportID != imp.ID {
		t.Fatalf("snapshot csv_import_id = %v, want %s", snapshot.CsvImportID, imp.ID)
	}

	var tags []string
	if err := json.Unmarshal(snapshot.Tags, &tags); err != nil {
		t.Fatalf("failed to unmarshal tags: %v", err)
	}
	if len(tags) != 2 || tags[0] != "tagA" || tags[1] != "tagB" {
		t.Fatalf("tags = %#v, want [tagA tagB]", tags)
	}

	var scenarios []string
	if err := json.Unmarshal(snapshot.Scenarios, &scenarios); err != nil {
		t.Fatalf("failed to unmarshal scenarios: %v", err)
	}
	if len(scenarios) != 2 || scenarios[0] != "scenario1" || scenarios[1] != "scenario2" {
		t.Fatalf("scenarios = %#v, want [scenario1 scenario2]", scenarios)
	}

	var errorLog []csvImportErrorEntry
	if err := json.Unmarshal(persistedImport.ErrorLog, &errorLog); err != nil {
		t.Fatalf("failed to unmarshal error log: %v", err)
	}
	if len(errorLog) != 1 {
		t.Fatalf("error log length = %d, want 1", len(errorLog))
	}
	if errorLog[0].Reason != csvErrorReasonUnknownLineUserID {
		t.Fatalf("unexpected error log entry: %#v", errorLog[0])
	}
	if strings.Contains(string(persistedImport.ErrorLog), "U_unknown_999") {
		t.Fatal("error log must not persist the unmatched LINE user ID")
	}
}

func TestImportFriendAttributesCSV_Integration_BoundsOwnerLookupToCSVBatch(t *testing.T) {
	db := setupLstepCsvImportServiceTestDB(t)
	ctx := context.Background()
	clinicID := seedLstepCsvImportClinic(t, db)
	lookupBatchSizes := make([]int, 0, 2)
	transactionIsolation := ""
	lookup := &mockLstepImportOwnerLookup{
		findExistingLineUserIDsFn: func(ctx context.Context, tx *gorm.DB, gotClinicID uint64, lineUserIDs []string) (map[string]struct{}, error) {
			if gotClinicID != clinicID {
				t.Fatalf("lookup clinic_id = %d, want %d", gotClinicID, clinicID)
			}
			if transactionIsolation == "" {
				if err := tx.WithContext(ctx).Raw("SHOW transaction_isolation").Scan(&transactionIsolation).Error; err != nil {
					t.Fatalf("failed to inspect transaction isolation: %v", err)
				}
			}
			lookupBatchSizes = append(lookupBatchSizes, len(lineUserIDs))
			matched := make(map[string]struct{}, len(lineUserIDs))
			for _, lineUserID := range lineUserIDs {
				matched[lineUserID] = struct{}{}
			}
			return matched, nil
		},
	}
	svc := NewLstepCsvImportService(
		db,
		NewLstepCsvImportRepository(db),
		lookup,
		&csvImportTestStaffRepository{accountID: 101},
	)
	rows := []string{"line_user_id"}
	for i := 0; i < csvSnapshotBatchSize+1; i++ {
		rows = append(rows, fmt.Sprintf("U-batch-%03d", i))
	}

	imp, err := svc.ImportFriendAttributesCSV(
		ctx,
		clinicID,
		"bounded-owner-lookup.csv",
		strings.NewReader(strings.Join(rows, "\n")),
		1,
	)

	if err != nil {
		t.Fatalf("ImportFriendAttributesCSV returned error: %v", err)
	}
	if imp.SuccessCount != csvSnapshotBatchSize+1 {
		t.Fatalf("success_count = %d, want %d", imp.SuccessCount, csvSnapshotBatchSize+1)
	}
	if fmt.Sprint(lookupBatchSizes) != fmt.Sprint([]int{csvSnapshotBatchSize, 1}) {
		t.Fatalf("lookup batch sizes = %v, want [%d 1]", lookupBatchSizes, csvSnapshotBatchSize)
	}
	if transactionIsolation != "repeatable read" {
		t.Fatalf("transaction isolation = %q, want repeatable read", transactionIsolation)
	}
}

func TestImportFriendAttributesCSV_Integration_InvalidHeaderCreatesFailedImport(t *testing.T) {
	db := setupLstepCsvImportServiceTestDB(t)
	ctx := context.Background()
	clinicID := seedLstepCsvImportClinic(t, db)

	svc := NewLstepCsvImportService(
		db,
		NewLstepCsvImportRepository(db),
		NewLstepCSVImportOwnerLookup(),
		&csvImportTestStaffRepository{accountID: 101},
	)

	csv := "表示名,タグ\n山田太郎,tagA\n"
	imp, err := svc.ImportFriendAttributesCSV(ctx, clinicID, "invalid-header.csv", strings.NewReader(csv), 1)
	if err == nil {
		t.Fatal("expected invalid header error, got nil")
	}
	if imp != nil {
		t.Fatalf("expected nil import on invalid header, got %+v", imp)
	}

	var imports []model.LstepCsvImport
	if err := db.Where("clinic_id = ?", clinicID).Order("created_at ASC").Find(&imports).Error; err != nil {
		t.Fatalf("failed to load imports: %v", err)
	}
	if len(imports) != 1 {
		t.Fatalf("import count = %d, want 1", len(imports))
	}
	if imports[0].Status != csvImportStatusFailed {
		t.Fatalf("failed import status = %s, want %s", imports[0].Status, csvImportStatusFailed)
	}
	if imports[0].FileName != "invalid-header.csv" {
		t.Fatalf("failed import file_name = %s", imports[0].FileName)
	}
}

// TestUpdateCsvImportRecordTx_ClinicIsolation は updateCsvImportRecordTx の clinic_id 境界を検証する
// (HIGH follow-up: 独立security-reviewが発見した tx.Where("clinic_id=?").Save(imp) パターン修正の回帰テスト)。
// GORM の Save() は imp.ID (主キー) がセット済みだと事前に連結した Where(clinic_id) を無視して主キーの
// みで UPDATE するため、誤って別クリニックの clinicID を渡しても他クリニックの行を上書きできてしまう
// 既知の落とし穴がある。本テストは2クリニックの実DBフィクスチャで、同一クリニックの更新は成功し、
// clinic_id が一致しない場合は他クリニックの行が一切変化しないことを直接確認する。
func TestUpdateCsvImportRecordTx_ClinicIsolation(t *testing.T) {
	db := setupLstepCsvImportServiceTestDB(t)
	ctx := context.Background()

	clinicA := seedLstepCsvImportClinic(t, db)
	clinicB := seedLstepCsvImportClinic(t, db)

	repo := NewLstepCsvImportRepository(db)

	rowA := &model.LstepCsvImport{ClinicID: clinicA, CsvType: "friend_attribute", FileName: "a.csv", Status: csvImportStatusProcessing}
	if err := repo.Create(ctx, rowA); err != nil {
		t.Fatalf("failed to create clinicA row: %v", err)
	}
	rowB := &model.LstepCsvImport{ClinicID: clinicB, CsvType: "friend_attribute", FileName: "b.csv", Status: csvImportStatusProcessing}
	if err := repo.Create(ctx, rowB); err != nil {
		t.Fatalf("failed to create clinicB row: %v", err)
	}

	t.Run("same-clinic update succeeds", func(t *testing.T) {
		rowA.Status = csvImportStatusCompleted
		rowA.RowCount = 5

		err := db.Transaction(func(tx *gorm.DB) error {
			return updateCsvImportRecordTx(tx, clinicA, rowA)
		})
		if err != nil {
			t.Fatalf("same-clinic update returned error: %v", err)
		}

		var stored model.LstepCsvImport
		if err := db.Where("id = ?", rowA.ID).First(&stored).Error; err != nil {
			t.Fatalf("failed to load clinicA row: %v", err)
		}
		if stored.Status != csvImportStatusCompleted || stored.RowCount != 5 {
			t.Fatalf("clinicA row not updated as expected: %+v", stored)
		}
	})

	t.Run("clinic mismatch does not overwrite the other clinic's row", func(t *testing.T) {
		// rowB.ID は clinicB 所有だが、誤って clinicA を境界として渡すケース（防御的境界チェック）。
		mismatched := *rowB
		mismatched.Status = csvImportStatusFailed
		mismatched.RowCount = 999

		err := db.Transaction(func(tx *gorm.DB) error {
			return updateCsvImportRecordTx(tx, clinicA, &mismatched)
		})
		if err == nil {
			t.Fatal("expected an error when clinicID does not match the row's owning clinic, got nil")
		}

		var stored model.LstepCsvImport
		if err := db.Where("id = ?", rowB.ID).First(&stored).Error; err != nil {
			t.Fatalf("failed to load clinicB row: %v", err)
		}
		if stored.Status != csvImportStatusProcessing || stored.RowCount != 0 {
			t.Fatalf("clinicB's row must remain unchanged, got: %+v", stored)
		}
	})
}

func setupLstepCsvImportServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := getLstepCsvImportTestDatabaseConnection(t)

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		t.Fatalf("failed to enable pgcrypto: %v", err)
	}
	if err := db.Exec("DROP TYPE IF EXISTS membership_type CASCADE").Error; err != nil {
		t.Fatalf("failed to drop membership_type: %v", err)
	}
	if err := db.Exec("CREATE TYPE membership_type AS ENUM ('non_member', 'member', 'deceased', 'transferred')").Error; err != nil {
		t.Fatalf("failed to create membership_type: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Company{},
		&model.Clinic{},
		&model.Owner{},
		&model.LstepCsvImport{},
		&model.LstepFriendAttributeSnapshot{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	db.Exec("TRUNCATE TABLE lstep_friend_attribute_snapshots RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE lstep_csv_imports CASCADE")
	db.Exec("TRUNCATE TABLE owners RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE clinics RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE companies RESTART IDENTITY CASCADE")

	return db
}

func getLstepCsvImportTestDatabaseConnection(t *testing.T) *gorm.DB {
	t.Helper()

	dbHost := envOrDefault("DB_HOST", "db")
	dbPort := envOrDefault("DB_PORT", "5432")
	dbUser := envOrDefault("DB_USER", "ekarte_user")
	dbPassword := envOrDefault("DB_PASSWORD", "ekarte_password")
	dbName := envOrDefault("DB_NAME", "ekarte_db")

	sanitizedTestName := strings.NewReplacer("/", "_", " ", "_").Replace(strings.ToLower(t.Name()))
	// PostgreSQLのidentifier 63-byte切捨てでも並列process間の一意suffixが失われないよう、
	// timestampを可変長test名より前へ置く。
	testDBName := fmt.Sprintf("%s_lstep_csv_import_%d_%s", dbName, time.Now().UnixNano(), sanitizedTestName)
	mainDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	mainDB, err := gorm.Open(postgres.Open(mainDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to main db: %v", err)
	}
	if err := mainDB.Exec("CREATE DATABASE " + testDBName).Error; err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	testDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, testDBName)
	if envDSN := os.Getenv("TEST_DATABASE_URL"); envDSN != "" {
		testDSN = envDSN
	}

	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
		_ = mainDB.Exec("DROP DATABASE IF EXISTS " + testDBName).Error
	})
	return db
}

func seedLstepCsvImportClinic(t *testing.T, db *gorm.DB) uint64 {
	t.Helper()

	company := &model.Company{Name: "Test Company"}
	if err := db.Create(company).Error; err != nil {
		t.Fatalf("failed to create company: %v", err)
	}

	clinic := &model.Clinic{
		CompanyID: company.ID,
		Name:      "Test Clinic",
	}
	if err := db.Create(clinic).Error; err != nil {
		t.Fatalf("failed to create clinic: %v", err)
	}

	return clinic.ID
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
