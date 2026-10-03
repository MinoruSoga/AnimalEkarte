package reservation

// reservation_staff_reference_test.go — EMR-114 / SLACK-RESERVATION-REFERENCE
// 「スタッフ/シフト作成後の予約保存が『参照先が存在しません』で失敗した」報告の現行 build 再検証。
//
// 2 つのスキーマ面で write path を固定する:
//
//   (A) 共有テスト DB（AutoMigrate 由来スキーマ）:
//       handler(POST /reservations 相当) と service の実経路で
//       新規スタッフ(所属+capability+shift 済み)を参照する予約の save+reload 成功、
//       非存在スタッフ参照の拒否、他医院のみ所属スタッフ参照の拒否、
//       doctor_id=0 の nil 正規化を pin する。
//       なお AutoMigrate スキーマの doctor FK は単一カラム (doctor_id -> staffs.id) のみで、
//       migration 由来の複合 FK fk_appointments_doctor_clinic は存在しない点に注意
//       （= この層では兼務スタッフの主所属不一致を再現できない）。
//
//   (B) migration 実 DDL スクラッチスキーマ:
//       共有 *_test DB 内の専用スキーマに 001_init.sql + 002 + 003 + 017 を適用し、
//       本番と同じ制約セットで同一操作を再実行する。兼務スタッフ
//       （主所属=別医院、当該医院に assignment + capability あり）を担当医にした
//       予約保存が 017_multiclinic_staff_fk_fix.sql 適用後の実 DDL で
//       受理されることを pin する（017 適用前は fk_appointments_doctor_clinic
//       複合 FK が主所属一致を要求して 23503 → 400「参照先が存在しません」で
//       再現していた — EMR-114 の残余欠陥）。
//
// 見つかった差異を隠さず pin する。「legacy composite rejects ...」（
// reservation_created_by_fk_test.go）と同じく、現在の実 DDL が実際に返す挙動を
// 固定する。
import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	staffpkg "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	reservationRefClinicA = uint64(1)
	reservationRefClinicB = uint64(2)
)

// setupReservationStaffRefSharedDB は共有テスト DB 上に予約スタッフ参照テスト用の
// 最小スキーマを用意し、対象テーブルをクリアして clinic 1/2 を seed する。
func setupReservationStaffRefSharedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.Owner{}, &model.AnimalSpecies{}, &model.Pet{}, &model.ReservationTypeGroup{},
		&model.ReservationType{}, &model.StaffReservationCapability{}, &model.ShiftEntry{},
		&model.Reservation{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE appointments, shift_entries, staff_reservation_capabilities, "+
			"staff_clinic_assignments, reservation_types, pets, owners, animal_species, staffs CASCADE",
	).Error)
	seedClinicsForFK(t, db, reservationRefClinicA, reservationRefClinicB)
	return db
}

// reservationRefService は実 repository + 実 transactor で予約作成 service を構築する。
// settings/clinic_holiday は「未設定=営業日扱い」ファインダで no-op にし、参照チェック
// 経路（staff guard → owner/pet link → slot conflict → repo.Create）のみを実物で通す。
func reservationRefService(db *gorm.DB) ReservationService {
	return NewReservationServiceWithClinicHolidays(
		NewReservationRepository(db),
		NewReservationTypeRepository(db),
		persistence.NewTransactor(db),
		NewReservationStaffRepository(db, nil),
		nil,
		nil,
		&mockLineReservationSettingFinder{findByClinicIDFn: func(context.Context, uint64) (*model.LineReservationSetting, error) {
			return nil, apperrors.WrapNotFound("settings", "")
		}},
		openDayHolidayFinder())
}

// reservationRefHandler は実 staff_clinic_assignments 参照経路を持つ CRUD handler を構築する。
func reservationRefHandler(db *gorm.DB) *CRUDHandler {
	return NewCRUDHandler(reservationRefService(db), nil, nil,
		staffpkg.NewStaffClinicAssignmentService(staffpkg.NewStaffClinicAssignmentRepository(db)))
}

// postReservationRef は POST /reservations 相当の create を handler 経路で実行し、
// clinic_id / user_id の認証コンテキストを付与する。
func postReservationRef(t *testing.T, h *CRUDHandler, clinicID, actorID uint64, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/reservations", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("clinic_id", strconv.FormatUint(clinicID, 10))
	c.Set("user_id", strconv.FormatUint(actorID, 10))
	h.CreateReservation(c)
	return w
}

// reservationRefBody は createReservationRequest の JSON 本文を組み立てる。
func reservationRefBody(start, end time.Time, typeID uint64, doctorJSON string) string {
	return fmt.Sprintf(
		`{"start_time":%q,"end_time":%q,"reservation_type_id":%d,"visit_type":"revisit","status":"pending"%s}`,
		start.Format(time.RFC3339), end.Format(time.RFC3339), typeID, doctorJSON)
}

// createdReservationID は 201 応答の Location ヘッダから採番 ID を取り出す。
func createdReservationID(t *testing.T, w *httptest.ResponseRecorder) uint64 {
	t.Helper()
	loc := w.Header().Get("Location")
	id, err := strconv.ParseUint(strings.TrimPrefix(loc, "/api/v1/reservations/"), 10, 64)
	require.NoError(t, err, "Location header %q must carry the created reservation id", loc)
	return id
}

// makeReservationRefStaff は指定医院を主所属とするスタッフを1件作成する。
// 担当医以外（creator 等）にも使えるよう staffType を引数化する。
// 所属紐付けは既存 helper の makeStaffClinicAssignment を使う（本経路は IsMain を参照しない）。
func makeReservationRefStaff(t *testing.T, db *gorm.DB, clinicID uint64, name string, staffType model.StaffType) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name, StaffType: staffType, LicenseNumber: "L-EMR114"}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	return s
}

// grantReservationCapability は (clinic, staff, type) の対応可能行を1件作成する。
func grantReservationCapability(t *testing.T, db *gorm.DB, clinicID, staffID, typeID uint64) {
	t.Helper()
	require.NoError(t, db.WithContext(context.Background()).Create(
		&model.StaffReservationCapability{ClinicID: clinicID, StaffID: staffID, ReservationTypeID: typeID}).Error)
}

// countAppointments は指定医院の予約行数を返す（拒否時の部分書き込み無し確認用）。
func countAppointments(t *testing.T, db *gorm.DB, clinicID uint64) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&model.Reservation{}).Where("clinic_id = ?", clinicID).Count(&n).Error)
	return n
}

// TestReservationStaffReference_SharedSchema は共有テスト DB 上の handler/service 実経路で
// 予約のスタッフ参照 write/read を pin する。スタッフ作成（+assignment/capability/shift）後の
// 保存・再読込成功と、無効参照・他医院スタッフ参照の拒否をそれぞれ固定する。
func TestReservationStaffReference_SharedSchema(t *testing.T) {
	db := setupReservationStaffRefSharedDB(t)
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })

	ctx := context.Background()
	repo := NewReservationRepository(db)
	svc := reservationRefService(db)
	handler := reservationRefHandler(db)

	start := time.Date(2027, 6, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)

	rtA := makeReservationType(t, db, reservationRefClinicA)

	creator := makeReservationRefStaff(t, db, reservationRefClinicA, "EMR114 作成者", model.StaffTypeNurse)
	makeStaffClinicAssignment(t, db, creator.ID, reservationRefClinicA)

	// 担当医: 主所属も assignment も clinic A、対応可能区分+当日シフトあり（報告された操作の再現）。
	doctor := makeReservationRefStaff(t, db, reservationRefClinicA, "EMR114 担当医A", model.StaffTypeDoctor)
	makeStaffClinicAssignment(t, db, doctor.ID, reservationRefClinicA)
	grantReservationCapability(t, db, reservationRefClinicA, doctor.ID, rtA.ID)
	makeShiftEntry(t, db, reservationRefClinicA, doctor.ID, start)

	// 他医院のみ所属のスタッフ（clinic A への assignment / capability なし）。
	otherClinicStaff := makeReservationRefStaff(t, db, reservationRefClinicB, "EMR114 他医院スタッフ", model.StaffTypeDoctor)
	makeStaffClinicAssignment(t, db, otherClinicStaff.ID, reservationRefClinicB)

	// 兼務スタッフ: 主所属は clinic B だが clinic A に assignment + capability + シフトあり。
	sharedDoctor := makeReservationRefStaff(t, db, reservationRefClinicB, "EMR114 兼務医", model.StaffTypeDoctor)
	makeStaffClinicAssignment(t, db, sharedDoctor.ID, reservationRefClinicB)
	makeStaffClinicAssignment(t, db, sharedDoctor.ID, reservationRefClinicA)
	grantReservationCapability(t, db, reservationRefClinicA, sharedDoctor.ID, rtA.ID)
	makeShiftEntry(t, db, reservationRefClinicA, sharedDoctor.ID, start)

	t.Run("assigned doctor saves and reloads via handler path", func(t *testing.T) {
		w := postReservationRef(t, handler, reservationRefClinicA, creator.ID,
			reservationRefBody(start, end, rtA.ID, fmt.Sprintf(`,"doctor_id":%d`, doctor.ID)))
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), fmt.Sprintf(`"doctor_id":%d`, doctor.ID))
		id := createdReservationID(t, w)

		loaded, err := repo.FindByID(ctx, reservationRefClinicA, id)
		require.NoError(t, err)
		require.NotNil(t, loaded.DoctorID)
		assert.Equal(t, doctor.ID, *loaded.DoctorID)
		require.NotNil(t, loaded.Doctor, "doctor preload must resolve the assigned staff")
		assert.Equal(t, doctor.ID, loaded.Doctor.ID)
		assert.Equal(t, doctor.Name, loaded.Doctor.Name)
	})

	t.Run("nonexistent doctor is rejected before any write", func(t *testing.T) {
		var maxID uint64
		require.NoError(t, db.Model(&model.Staff{}).Select("COALESCE(MAX(id),0)").Scan(&maxID).Error)
		missingID := maxID + 900001

		// handler 層（assignment 事前チェック）: 400 invalid-input。
		w := postReservationRef(t, handler, reservationRefClinicA, creator.ID,
			reservationRefBody(start, end, rtA.ID, fmt.Sprintf(`,"doctor_id":%d`, missingID)))
		assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "指定されたスタッフはこの医院に所属していません")
		assert.NotContains(t, w.Body.String(), "fk_")

		// service 層（トランザクション内 staff guard）: not-found。
		_, err := svc.Create(ctx, &CreateManualReservationInput{
			ClinicID: reservationRefClinicA, StartTime: start, EndTime: end,
			ReservationTypeID: rtA.ID, DoctorID: &missingID,
			VisitType: model.VisitTypeRevisit, Status: model.ReservationStatusPending,
			Source: model.ReservationSourceManual, CreatedBy: &creator.ID,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "expected not-found, got: %v", err)
	})

	t.Run("other-clinic-only staff reference is rejected at both layers", func(t *testing.T) {
		w := postReservationRef(t, handler, reservationRefClinicA, creator.ID,
			reservationRefBody(start, end, rtA.ID, fmt.Sprintf(`,"doctor_id":%d`, otherClinicStaff.ID)))
		assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "指定されたスタッフはこの医院に所属していません")

		_, err := svc.Create(ctx, &CreateManualReservationInput{
			ClinicID: reservationRefClinicA, StartTime: start, EndTime: end,
			ReservationTypeID: rtA.ID, DoctorID: &otherClinicStaff.ID,
			VisitType: model.VisitTypeRevisit, Status: model.ReservationStatusPending,
			Source: model.ReservationSourceManual, CreatedBy: &creator.ID,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "expected not-found, got: %v", err)
	})

	t.Run("assigned multi-clinic doctor saves on automigrate schema", func(t *testing.T) {
		// 共有テスト DB には複合 FK が無いため、この操作は成功する。
		// 実 DDL（migration 適用スキーマ）での同一操作は TestReservationStaffReference_RealDDL が pin する。
		w := postReservationRef(t, handler, reservationRefClinicA, creator.ID,
			reservationRefBody(start.Add(2*time.Hour), end.Add(2*time.Hour), rtA.ID,
				fmt.Sprintf(`,"doctor_id":%d`, sharedDoctor.ID)))
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		id := createdReservationID(t, w)
		loaded, err := repo.FindByID(ctx, reservationRefClinicA, id)
		require.NoError(t, err)
		require.NotNil(t, loaded.DoctorID)
		assert.Equal(t, sharedDoctor.ID, *loaded.DoctorID)
	})

	t.Run("doctor_id zero normalizes to unset", func(t *testing.T) {
		w := postReservationRef(t, handler, reservationRefClinicA, creator.ID,
			reservationRefBody(start.Add(3*time.Hour), end.Add(3*time.Hour), rtA.ID, `,"doctor_id":0`))
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		id := createdReservationID(t, w)
		loaded, err := repo.FindByID(ctx, reservationRefClinicA, id)
		require.NoError(t, err)
		assert.Nil(t, loaded.DoctorID)
	})

	assert.Zero(t, countAppointments(t, db, reservationRefClinicB), "clinic B must remain untouched by clinic A writes")
}

// reservationRefScratchConfig は共有テスト DB と同じ env 解決（DB_HOST/DB_PORT/DB_USER/
// DB_PASSWORD/DB_NAME）で <DB_NAME>_test への接続設定を返す。作成するのはその DB 内部の
// 専用スキーマのみであり、DB 自体は SetupTestDB が保証する *_test データベースに限定する
// （TEST_DATABASE_URL は読まず、常に派生名に固定して誤接続を防ぐ）。
func reservationRefScratchConfig(t *testing.T) *pgx.ConnConfig {
	t.Helper()
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "db"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "ekarte_user"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "ekarte_password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "ekarte_db"
	}
	testDBName := dbName + "_test"
	require.True(t, strings.HasSuffix(testDBName, "_test"),
		"scratch schema target must stay inside a dedicated *_test database, got %q", testDBName)
	cfg, err := pgx.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, testDBName))
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(cfg.Database, "_test"))
	cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	return cfg
}

// reservationRefScratchFixture は共有 *_test DB 内部の専用スキーマに migration 実 DDL
// （001_init.sql + 002 + 003）を適用した gorm.DB を返す。search_path をスキーマに固定し、
// cleanup でスキーマごと CASCADE 削除する。発生した pg エラーは lastPG に保持される
// （apperrors 変換前の制約名を検証するため reservation_created_by_fk_test.go と同型）。
type reservationRefScratchFixture struct {
	db     *gorm.DB
	lastPG *pgconn.PgError
}

func setupReservationRefScratchSchema(t *testing.T) *reservationRefScratchFixture {
	t.Helper()
	if testing.Short() {
		t.Skip("scratch-schema migration test")
	}
	testdb.SetupTestDB(t) // <DB_NAME>_test の存在を保証する（作成責務は testdb 側）。
	cfg := reservationRefScratchConfig(t)
	var nonce [10]byte
	_, err := rand.Read(nonce[:])
	require.NoError(t, err)
	schema := "rsv_ref_" + hex.EncodeToString(nonce[:])
	privateSchema := schema + "_private"
	// search_path の先頭は本スキーマ（新規オブジェクトは全てここに作成される）。
	// ekarte_db_test には pg_trgm 等の extension が public に既にインストール済みで、
	// 001_init.sql の `CREATE EXTENSION IF NOT EXISTS` は既存を skip するため、
	// public も検索対象に含めないと gin_trgm_ops 等の演算子クラスが解決できない。
	cfg.RuntimeParams["search_path"] = schema + ",public"
	sqlDB := stdlib.OpenDB(*cfg)
	sqlDB.SetMaxOpenConns(4)
	t.Cleanup(func() { assert.NoError(t, sqlDB.Close()) })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() {
		assert.NoError(t, db.Exec("DROP SCHEMA IF EXISTS "+privateSchema+" CASCADE").Error)
		assert.NoError(t, db.Exec("DROP SCHEMA "+schema+" CASCADE").Error)
	})
	raw, err := os.ReadFile("../../migrations/001_init.sql")
	require.NoError(t, err)
	// reservation_created_by_fk_test.go と同じ namespace 置換でプライベート関数と
	// public スキーマ RLS 走査をこのスキーマ内に閉じ込める。
	ddl := strings.ReplaceAll(string(raw), "app_private", privateSchema)
	ddl = strings.ReplaceAll(ddl, "'public'", "'"+schema+"'")
	require.NoError(t, db.Exec(ddl).Error)
	applyReservationCreatedByMigration(t, db, "../../migrations/002_medical_records_entered_by_staff_fk.sql")
	applyReservationCreatedByMigration(t, db, reservationCreatedByMigration)                       // 003: created_by 単一カラム FK 化
	applyReservationCreatedByMigration(t, db, "../../migrations/017_multiclinic_staff_fk_fix.sql") // 017: 兼務スタッフ参照の複合 FK 除去
	fixture := &reservationRefScratchFixture{db: db}
	require.NoError(t, db.Callback().Create().After("gorm:create").Register("test:capture_reservation_ref_fk", func(tx *gorm.DB) {
		var pgErr *pgconn.PgError
		if errors.As(tx.Error, &pgErr) {
			fixture.lastPG = pgErr
		}
	}))
	return fixture
}

// reservationRefScratchActors は実 DDL スキーマ内に最小限の参照グラフを seed する。
type reservationRefScratchActors struct {
	clinicA         model.Clinic
	clinicB         model.Clinic
	creator         model.Staff
	doctorHome      model.Staff // 主所属=clinic A、clinic A に assignment+capability
	doctorShared    model.Staff // 主所属=clinic B、clinic A/B 両方に assignment、clinic A に capability
	reservationType model.ReservationType
	start           time.Time
	end             time.Time
}

func seedReservationRefScratchActors(t *testing.T, db *gorm.DB) *reservationRefScratchActors {
	t.Helper()
	f := &reservationRefScratchActors{}
	company := model.Company{Name: "EMR114 ref fixture"}
	require.NoError(t, db.Create(&company).Error)
	f.clinicA = model.Clinic{CompanyID: company.ID, Name: "EMR114 clinic A"}
	f.clinicB = model.Clinic{CompanyID: company.ID, Name: "EMR114 clinic B"}
	require.NoError(t, db.Create(&f.clinicA).Error)
	require.NoError(t, db.Create(&f.clinicB).Error)

	f.creator = model.Staff{ClinicID: f.clinicA.ID, Name: "EMR114 creator", StaffType: model.StaffTypeNurse, IsActive: true, LicenseNumber: "L-C"}
	require.NoError(t, db.Create(&f.creator).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: f.creator.ID, ClinicID: f.clinicA.ID, IsMain: true}).Error)

	f.doctorHome = model.Staff{ClinicID: f.clinicA.ID, Name: "EMR114 doctor home A", StaffType: model.StaffTypeDoctor, IsActive: true, LicenseNumber: "L-A"}
	require.NoError(t, db.Create(&f.doctorHome).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: f.doctorHome.ID, ClinicID: f.clinicA.ID, IsMain: true}).Error)

	f.doctorShared = model.Staff{ClinicID: f.clinicB.ID, Name: "EMR114 shared doctor", StaffType: model.StaffTypeDoctor, IsActive: true, LicenseNumber: "L-S"}
	require.NoError(t, db.Create(&f.doctorShared).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: f.doctorShared.ID, ClinicID: f.clinicB.ID, IsMain: true}).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: f.doctorShared.ID, ClinicID: f.clinicA.ID, IsMain: false}).Error)

	f.reservationType = model.ReservationType{ClinicID: f.clinicA.ID, Name: "EMR114 type A", Category: model.ReservationTypeCategoryGeneral}
	require.NoError(t, db.Create(&f.reservationType).Error)

	f.start = time.Date(2027, 6, 1, 10, 0, 0, 0, time.UTC)
	f.end = f.start.Add(30 * time.Minute)

	// 担当医の対応可能区分と当日シフト（報告操作の再現条件）。
	require.NoError(t, db.Create(&model.StaffReservationCapability{ClinicID: f.clinicA.ID, StaffID: f.doctorHome.ID, ReservationTypeID: f.reservationType.ID}).Error)
	require.NoError(t, db.Create(&model.StaffReservationCapability{ClinicID: f.clinicA.ID, StaffID: f.doctorShared.ID, ReservationTypeID: f.reservationType.ID}).Error)
	require.NoError(t, db.Create(&model.ShiftEntry{ClinicID: f.clinicA.ID, StaffID: f.doctorHome.ID, Date: f.start, ShiftType: model.ShiftTypeFull}).Error)
	require.NoError(t, db.Create(&model.ShiftEntry{ClinicID: f.clinicA.ID, StaffID: f.doctorShared.ID, Date: f.start, ShiftType: model.ShiftTypeFull}).Error)
	return f
}

func (f *reservationRefScratchActors) input(doctorID *uint64) *CreateManualReservationInput {
	return &CreateManualReservationInput{
		ClinicID: f.clinicA.ID, StartTime: f.start, EndTime: f.end,
		ReservationTypeID: f.reservationType.ID, DoctorID: doctorID,
		VisitType: model.VisitTypeRevisit, Status: model.ReservationStatusPending,
		Source: model.ReservationSourceManual, CreatedBy: &f.creator.ID,
	}
}

// TestReservationStaffReference_RealDDL は migration 実 DDL（001+002+003 適用）上で
// 報告操作と同一の予約作成を実行し、doctor 参照の現行挙動を制約名まで pin する。
func TestReservationStaffReference_RealDDL(t *testing.T) {
	fixture := setupReservationRefScratchSchema(t)
	db := fixture.db
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })
	ctx := context.Background()
	repo := NewReservationRepository(db)
	svc := reservationRefService(db)
	handler := reservationRefHandler(db)

	f := seedReservationRefScratchActors(t, db)

	t.Run("same-primary-clinic doctor saves and reloads on real DDL", func(t *testing.T) {
		// 報告された操作（スタッフ+シフト作成後に担当医を指定して予約保存）を実 DDL で再現する。
		got, err := svc.Create(ctx, f.input(&f.doctorHome.ID))
		require.NoError(t, err)
		require.NotNil(t, got.DoctorID)
		assert.Equal(t, f.doctorHome.ID, *got.DoctorID)

		loaded, err := repo.FindByID(ctx, f.clinicA.ID, got.ID)
		require.NoError(t, err)
		require.NotNil(t, loaded.DoctorID)
		assert.Equal(t, f.doctorHome.ID, *loaded.DoctorID)
		require.NotNil(t, loaded.Doctor)
		assert.Equal(t, f.doctorHome.Name, loaded.Doctor.Name)
	})

	t.Run("handler create with assigned doctor returns 201", func(t *testing.T) {
		w := postReservationRef(t, handler, f.clinicA.ID, f.creator.ID,
			reservationRefBody(f.start.Add(time.Hour), f.end.Add(time.Hour), f.reservationType.ID,
				fmt.Sprintf(`,"doctor_id":%d`, f.doctorHome.ID)))
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		id := createdReservationID(t, w)
		loaded, err := repo.FindByID(ctx, f.clinicA.ID, id)
		require.NoError(t, err)
		require.NotNil(t, loaded.DoctorID)
		assert.Equal(t, f.doctorHome.ID, *loaded.DoctorID)
	})

	t.Run("multi-clinic creator saves on post-003 created_by FK", func(t *testing.T) {
		// 兼務作成者（主所属 clinic B、clinic A に assignment あり）: migration 003 で
		// created_by FK は単一カラム化済みのため、現行 DDL では保存に成功するはず。
		sharedCreator := model.Staff{ClinicID: f.clinicB.ID, Name: "EMR114 shared creator", StaffType: model.StaffTypeNurse, IsActive: true, LicenseNumber: "L-SC"}
		require.NoError(t, db.Create(&sharedCreator).Error)
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: sharedCreator.ID, ClinicID: f.clinicB.ID, IsMain: true}).Error)
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: sharedCreator.ID, ClinicID: f.clinicA.ID, IsMain: false}).Error)

		// created_by のみ検証するため doctor は未指定＋別時間枠にする
		// （doctorHome は先の subtest で 10:00 UTC の枠を確保済みのため）。
		// doctor_id=nil の経路は当日（Asia/Tokyo 日付）の出勤医師数を上限とする
		// 容量チェックを行うため、シフトがある 2027-06-01 JST 内の枠を使う。
		input := f.input(nil)
		input.StartTime = f.start.Add(3 * time.Hour)
		input.EndTime = f.end.Add(3 * time.Hour)
		input.CreatedBy = &sharedCreator.ID
		got, err := svc.Create(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, got.CreatedBy)
		assert.Equal(t, sharedCreator.ID, *got.CreatedBy)
	})

	t.Run("assigned multi-clinic doctor saves on real DDL post-017", func(t *testing.T) {
		// 兼務担当医: staffs.clinic_id=B だが clinic A に assignment+capability+shift あり。
		// 017 適用前は fk_appointments_doctor_clinic の主所属一致要求で 23503 →
		// 400「参照先が存在しません」（報告された障害そのもの）だった。
		// 017 で複合 FK は単一カラム化済みのため、現行 DDL では受理される。
		fixture.lastPG = nil
		w := postReservationRef(t, handler, f.clinicA.ID, f.creator.ID,
			reservationRefBody(f.start.Add(4*time.Hour), f.end.Add(4*time.Hour), f.reservationType.ID,
				fmt.Sprintf(`,"doctor_id":%d`, f.doctorShared.ID)))

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		assert.Nil(t, fixture.lastPG, "no postgres error expected post-017")
		id := createdReservationID(t, w)
		loaded, err := repo.FindByID(ctx, f.clinicA.ID, id)
		require.NoError(t, err)
		require.NotNil(t, loaded.DoctorID)
		assert.Equal(t, f.doctorShared.ID, *loaded.DoctorID)
	})

	t.Run("post-017 constraint shape on real DDL", func(t *testing.T) {
		// 017 が複合 FK (col, clinic_id) -> staffs(id, clinic_id) を除去し、
		// 単一カラム FK -> staffs(id) に置き換えたことを構造的に pin する。
		// ON DELETE 意味論（doctor=SET NULL / actor・staff=RESTRICT）は維持。
		type fkRow struct {
			Table      string `gorm:"column:table_name"`
			Constraint string `gorm:"column:constraint_name"`
			Columns    int    `gorm:"column:columns"`
			DeleteRule string `gorm:"column:delete_rule"`
		}
		var rows []fkRow
		require.NoError(t, db.Raw(`
			SELECT tc.table_name, tc.constraint_name,
			       COUNT(kcu.column_name) AS columns,
			       rc.delete_rule
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
			  ON kcu.constraint_name = tc.constraint_name
			 AND kcu.table_name = tc.table_name
			 AND kcu.table_schema = tc.table_schema
			JOIN information_schema.referential_constraints rc
			  ON rc.constraint_name = tc.constraint_name
			 AND rc.constraint_schema = tc.constraint_schema
			WHERE tc.table_schema = current_schema()
			  AND tc.constraint_type = 'FOREIGN KEY'
			  AND tc.constraint_name IN (
			    'fk_appointments_doctor',
			    'fk_hospitalizations_doctor',
			    'fk_medical_records_doctor',
			    'fk_cash_register_close_adjustments_actor',
			    'fk_medical_record_image_upload_quota_staff',
			    'fk_lab_device_waits_staff'
			  )
			GROUP BY tc.table_name, tc.constraint_name, rc.delete_rule
			ORDER BY tc.constraint_name`).Scan(&rows).Error)
		require.Len(t, rows, 6, "all six single-column staff FKs must exist post-017")
		for _, r := range rows {
			assert.Equal(t, 1, r.Columns, "%s must be a single-column FK", r.Constraint)
		}
		// ON DELETE 意味論の維持: doctor 系は SET NULL、actor/staff 系は RESTRICT。
		deleteRule := map[string]string{}
		for _, r := range rows {
			deleteRule[r.Constraint] = r.DeleteRule
		}
		assert.Equal(t, "SET NULL", deleteRule["fk_appointments_doctor"])
		assert.Equal(t, "SET NULL", deleteRule["fk_hospitalizations_doctor"])
		assert.Equal(t, "SET NULL", deleteRule["fk_medical_records_doctor"])
		assert.Equal(t, "RESTRICT", deleteRule["fk_cash_register_close_adjustments_actor"])
		assert.Equal(t, "RESTRICT", deleteRule["fk_medical_record_image_upload_quota_staff"])
		assert.Equal(t, "RESTRICT", deleteRule["fk_lab_device_waits_staff"])
		// 旧複合 FK（同一欠陥クラス6件）が残っていないこと。owners/reservation_types
		// 等の per-clinic エンティティへの *_clinic 複合 FK は正規のため存続する。
		var stale int64
		require.NoError(t, db.Raw(`
			SELECT COUNT(*) FROM information_schema.table_constraints
			WHERE table_schema = current_schema() AND constraint_type = 'FOREIGN KEY'
			  AND constraint_name IN (
			    'fk_appointments_doctor_clinic',
			    'fk_hospitalizations_doctor_clinic',
			    'fk_medical_records_doctor_clinic',
			    'fk_cash_register_close_adjustments_actor_clinic',
			    'fk_medical_record_image_upload_quota_staff_clinic',
			    'fk_lab_device_waits_staff_clinic'
			  )`).Scan(&stale).Error)
		assert.Zero(t, stale, "the six composite staff FKs must be dropped post-017")
	})

	t.Run("nonexistent doctor stays rejected on real DDL", func(t *testing.T) {
		missingID := uint64(90000042)
		_, err := svc.Create(ctx, f.input(&missingID))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "staff guard must report not-found, got: %v", err)
	})

	t.Run("other-clinic-only staff stays rejected on real DDL", func(t *testing.T) {
		foreignStaff := model.Staff{ClinicID: f.clinicB.ID, Name: "EMR114 B-only", StaffType: model.StaffTypeDoctor, IsActive: true, LicenseNumber: "L-B"}
		require.NoError(t, db.Create(&foreignStaff).Error)
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: foreignStaff.ID, ClinicID: f.clinicB.ID, IsMain: true}).Error)
		_, err := svc.Create(ctx, f.input(&foreignStaff.ID))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "missing clinic A assignment must be not-found, got: %v", err)
	})
}
