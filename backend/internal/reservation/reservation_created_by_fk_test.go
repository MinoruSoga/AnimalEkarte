package reservation

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
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
)

const reservationCreatedByMigration = "../../migrations/003_appointments_created_by_staff_fk.sql"

// Explicit opt-in is required: this suite applies real DDL only in an approved,
// disposable loopback database. It never uses the shared testdb helper.
func reservationCreatedByTestConfig(dsn string) (*pgx.ConnConfig, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("invalid reservation test DSN")
	}
	if cfg.Database != "ae_reservation_created_by_test" {
		return nil, errors.New("reservation test requires its exact disposable database name")
	}
	loopback := func(host string) bool {
		ip := net.ParseIP(host)
		return host == "localhost" || (ip != nil && ip.IsLoopback())
	}
	if !loopback(cfg.Host) {
		return nil, errors.New("reservation test requires a loopback host")
	}
	for _, fallback := range cfg.Fallbacks {
		if !loopback(fallback.Host) {
			return nil, errors.New("reservation test requires loopback fallback hosts")
		}
	}
	cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	return cfg, nil
}

func TestReservationCreatedByTestConfig_RejectsOtherDatabases(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		dsn  string
		ok   bool
	}{
		{"dedicated", "host=127.0.0.1 dbname=ae_reservation_created_by_test", true},
		{"ipv6", "host=::1 dbname=ae_reservation_created_by_test", true},
		{"shared", "host=127.0.0.1 dbname=animal_ekarte_test", false},
		{"remote", "host=192.0.2.1 dbname=ae_reservation_created_by_test", false},
		{"remote fallback", "host=127.0.0.1,192.0.2.1 dbname=ae_reservation_created_by_test", false},
		{"unix socket", "host=/tmp dbname=ae_reservation_created_by_test", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := reservationCreatedByTestConfig(tc.dsn)
			if tc.ok {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

type reservationCreatedByFixture struct {
	db     *gorm.DB
	lastPG *pgconn.PgError
}

func setupReservationCreatedByFK(t *testing.T) *reservationCreatedByFixture {
	t.Helper()
	if testing.Short() {
		t.Skip("disposable PostgreSQL migration test")
	}
	dsn := os.Getenv("RESERVATION_CREATED_BY_TEST_DSN")
	if dsn == "" {
		t.Skip("requires approved disposable database in RESERVATION_CREATED_BY_TEST_DSN")
	}
	cfg, err := reservationCreatedByTestConfig(dsn)
	require.NoError(t, err)
	var nonce [10]byte
	_, err = rand.Read(nonce[:])
	require.NoError(t, err)
	schema := "rsv_actor_" + hex.EncodeToString(nonce[:])
	privateSchema := schema + "_private"
	// The dedicated database may already contain the real migration smoke gate.
	// Its public extensions supply operator classes; all fixture tables are
	// created in our first, exclusively owned schema.
	cfg.RuntimeParams["search_path"] = schema
	sqlDB := stdlib.OpenDB(*cfg)
	sqlDB.SetMaxOpenConns(6)
	t.Cleanup(func() { assert.NoError(t, sqlDB.Close()) })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() {
		// Identifiers are exclusively generated above, never taken from the DSN.
		assert.NoError(t, db.Exec("DROP SCHEMA IF EXISTS "+privateSchema+" CASCADE").Error)
		assert.NoError(t, db.Exec("DROP SCHEMA "+schema+" CASCADE").Error)
	})
	raw, err := os.ReadFile("../../migrations/001_init.sql")
	require.NoError(t, err)
	// Namespace-only adaptation prevents the full migration's private functions
	// and public-schema RLS scan from touching any other fixture's objects.
	ddl := strings.ReplaceAll(string(raw), "app_private", privateSchema)
	ddl = strings.ReplaceAll(ddl, "'public'", "'"+schema+"'")
	require.NoError(t, db.Exec(ddl).Error)
	applyReservationCreatedByMigration(t, db, "../../migrations/002_medical_records_entered_by_staff_fk.sql")
	fixture := &reservationCreatedByFixture{db: db}
	require.NoError(t, db.Callback().Create().After("gorm:create").Register("test:capture_reservation_fk", func(tx *gorm.DB) {
		var pgErr *pgconn.PgError
		if errors.As(tx.Error, &pgErr) {
			fixture.lastPG = pgErr
		}
	}))
	return fixture
}

func applyReservationCreatedByMigration(t *testing.T, db *gorm.DB, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(raw)).Error)
}

type reservationCreatedByActors struct {
	clinicA         model.Clinic
	clinicB         model.Clinic
	actor           model.Staff
	account         model.Account
	owner           model.Owner
	pets            [2]model.Pet
	reservationType model.ReservationType
}

func seedReservationCreatedByActors(t *testing.T, db *gorm.DB, assigned, admin bool) *reservationCreatedByActors {
	t.Helper()
	f := &reservationCreatedByActors{}
	company := model.Company{Name: "Reservation FK fixture"}
	require.NoError(t, db.Create(&company).Error)
	f.clinicA = model.Clinic{CompanyID: company.ID, Name: "Synthetic home clinic"}
	f.clinicB = model.Clinic{CompanyID: company.ID, Name: "Synthetic selected clinic"}
	require.NoError(t, db.Create(&f.clinicA).Error)
	require.NoError(t, db.Create(&f.clinicB).Error)
	f.account = model.Account{Email: fmt.Sprintf("reservation-actor-%d@example.test", company.ID), PasswordHash: "synthetic-unused-hash", IsActive: true, IsSystemAdmin: admin}
	require.NoError(t, db.Create(&f.account).Error)
	f.actor = model.Staff{ClinicID: f.clinicA.ID, AccountID: &f.account.ID, Name: "Synthetic reservation creator", StaffType: model.StaffTypeNurse, IsActive: true, LicenseNumber: "must-not-be-preloaded"}
	require.NoError(t, db.Create(&f.actor).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: f.actor.ID, ClinicID: f.clinicA.ID, IsMain: true}).Error)
	if assigned {
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: f.actor.ID, ClinicID: f.clinicB.ID}).Error)
	}
	f.owner = model.Owner{ClinicID: f.clinicB.ID, Name: "Synthetic reservation owner"}
	require.NoError(t, db.Create(&f.owner).Error)
	species := model.AnimalSpecies{Name: fmt.Sprintf("Synthetic species %d", company.ID)}
	require.NoError(t, db.Create(&species).Error)
	for i := range f.pets {
		f.pets[i] = model.Pet{ClinicID: f.clinicB.ID, OwnerID: f.owner.ID, AnimalSpeciesID: species.ID, Name: fmt.Sprintf("Synthetic pet %d", i), Gender: model.PetGenderUnknown, Status: model.PetStatusAlive}
		require.NoError(t, db.Create(&f.pets[i]).Error)
	}
	f.reservationType = model.ReservationType{ClinicID: f.clinicB.ID, Name: "Synthetic reservation type"}
	require.NoError(t, db.Create(&f.reservationType).Error)
	// Capacity is independent of the recording actor: two on-duty doctors allow
	// the two-pet batch while doctor_id remains unset on the appointment.
	for i := 0; i < 2; i++ {
		doctor := model.Staff{ClinicID: f.clinicB.ID, Name: fmt.Sprintf("Synthetic duty doctor %d", i), StaffType: model.StaffTypeDoctor, IsActive: true}
		require.NoError(t, db.Create(&doctor).Error)
		require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: f.clinicB.ID}).Error)
		require.NoError(t, db.Create(&model.ShiftEntry{ClinicID: f.clinicB.ID, StaffID: doctor.ID, Date: f.input().StartTime, ShiftType: model.ShiftTypeFull}).Error)
	}
	return f
}

func reservationCreatedByService(db *gorm.DB) ReservationService {
	return NewReservationServiceWithClinicHolidays(NewReservationRepository(db), NewReservationTypeRepository(db), persistence.NewTransactor(db), nil, nil, nil,
		&mockLineReservationSettingFinder{findByClinicIDFn: func(context.Context, uint64) (*model.LineReservationSetting, error) {
			return nil, apperrors.WrapNotFound("settings", "")
		}},
		openDayHolidayFinder())
}

func (f *reservationCreatedByActors) input() *CreateManualReservationInput {
	start := time.Date(2027, 6, 1, 9, 0, 0, 0, time.UTC)
	return &CreateManualReservationInput{ClinicID: f.clinicB.ID, StartTime: start, EndTime: start.Add(time.Hour), OwnerID: &f.owner.ID, PetID: &f.pets[0].ID, ReservationTypeID: f.reservationType.ID, VisitType: model.VisitTypeRevisit, Status: model.ReservationStatusPending, Source: model.ReservationSourceManual, CreatedBy: &f.actor.ID}
}

func reservationCreatedByRequest(t *testing.T, db *gorm.DB, f *reservationCreatedByActors, route, doctorJSON string) *httptest.ResponseRecorder {
	t.Helper()
	input := f.input()
	body := fmt.Sprintf(`{"start_time":%q,"end_time":%q,"reservation_type_id":%d,"visit_type":"revisit","status":"pending","owner_id":%d,"pet_id":%d,"created_by":999999,"clinic_id":999999%s`, input.StartTime.Format(time.RFC3339), input.EndTime.Format(time.RFC3339), f.reservationType.ID, f.owner.ID, f.pets[0].ID, doctorJSON)
	if route == "batch" {
		body += fmt.Sprintf(`,"pets":[{"owner_id":%d,"pet_id":%d},{"owner_id":%d,"pet_id":%d}]`, f.owner.ID, f.pets[0].ID, f.owner.ID, f.pets[1].ID)
	}
	body += "}"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/reservations", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("clinic_id", strconv.FormatUint(f.clinicB.ID, 10))
	c.Set("user_id", strconv.FormatUint(f.actor.ID, 10))
	switch route {
	case "create":
		NewCRUDHandler(reservationCreatedByService(db), nil, nil, nil).CreateReservation(c)
	case "batch":
		NewCRUDHandler(reservationCreatedByService(db), nil, nil, nil).CreateReservationBatch(c)
	case "admin":
		svc := NewReservationAdminServiceWithClinicHolidays(NewReservationAdminRepository(db), NewReservationRepository(db), NewReservationTypeRepository(db), persistence.NewTransactor(db), nil, nil, nil, openDayHolidayFinder())
		NewReservationAdminHandler(svc, nil).CreateReservationAdmin(c)
	default:
		t.Fatalf("unknown route %q", route)
	}
	return w
}

func assertReservationCreatedByReadback(t *testing.T, db *gorm.DB, f *reservationCreatedByActors, wantCount int) []model.Reservation {
	t.Helper()
	var rows []model.Reservation
	require.NoError(t, db.Where("clinic_id = ?", f.clinicB.ID).Find(&rows).Error)
	require.Len(t, rows, wantCount)
	repo := NewReservationRepository(db)
	for _, row := range rows {
		assert.Nil(t, row.DoctorID)
		require.NotNil(t, row.CreatedBy)
		assert.Equal(t, f.actor.ID, *row.CreatedBy)
		got, err := repo.FindByID(context.Background(), f.clinicB.ID, row.ID)
		require.NoError(t, err)
		require.NotNil(t, got.CreatedByStaff)
		assert.Equal(t, f.actor.ID, got.CreatedByStaff.ID)
		assert.Equal(t, f.actor.Name, got.CreatedByStaff.Name)
		assert.Zero(t, got.CreatedByStaff.ClinicID, "historical actor projection must expose only id/name")
		assert.Empty(t, got.CreatedByStaff.LicenseNumber)
		assert.Nil(t, got.CreatedByStaff.AccountID)
		_, err = repo.FindByID(context.Background(), f.clinicA.ID, row.ID)
		assert.True(t, apperrors.IsNotFound(err), "a different clinic must not read the reservation")
	}
	return rows
}

func TestReservationCreatedByFK(t *testing.T) {
	fixture := setupReservationCreatedByFK(t)
	db := fixture.db
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(previousMode) })
	t.Run("legacy composite rejects assigned creator with no doctor", func(t *testing.T) {
		f := seedReservationCreatedByActors(t, db, true, false)
		_, err := reservationCreatedByService(db).Create(context.Background(), f.input())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "参照先が存在しません")
		require.NotNil(t, fixture.lastPG)
		assert.Equal(t, "23503", fixture.lastPG.Code)
		assert.Equal(t, "fk_appointments_created_by_clinic", fixture.lastPG.ConstraintName)
		var count int64
		require.NoError(t, db.Model(&model.Reservation{}).Where("clinic_id = ?", f.clinicB.ID).Count(&count).Error)
		assert.Zero(t, count)
	})
	applyReservationCreatedByMigration(t, db, reservationCreatedByMigration)
	t.Run("other clinic foreign keys remain enforced", func(t *testing.T) {
		f := seedReservationCreatedByActors(t, db, true, false)
		foreign := seedReservationCreatedByActors(t, db, true, false)
		for _, tc := range []struct {
			name       string
			constraint string
			change     func(*model.Reservation)
		}{
			{"owner", "fk_appointments_owner_clinic", func(r *model.Reservation) { r.OwnerID = &foreign.owner.ID }},
			{"pet", "fk_appointments_pet_clinic", func(r *model.Reservation) { r.PetID = &foreign.pets[0].ID }},
			{"type", "fk_appointments_reservation_type_clinic", func(r *model.Reservation) { r.ReservationTypeID = foreign.reservationType.ID }},
			{"doctor", "fk_appointments_doctor_clinic", func(r *model.Reservation) { r.DoctorID = &foreign.actor.ID }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				input := f.input()
				row := &model.Reservation{ClinicID: input.ClinicID, StartTime: input.StartTime, EndTime: input.EndTime, ReservationTypeID: input.ReservationTypeID, CreatedBy: input.CreatedBy}
				tc.change(row)
				err := db.Create(row).Error // Bypass application validation to exercise the remaining DB defense.
				var pgErr *pgconn.PgError
				require.ErrorAs(t, err, &pgErr)
				assert.Equal(t, "23503", pgErr.Code)
				assert.Equal(t, tc.constraint, pgErr.ConstraintName)
			})
		}
		var count int64
		require.NoError(t, db.Model(&model.Reservation{}).Where("clinic_id = ?", f.clinicB.ID).Count(&count).Error)
		assert.Zero(t, count)
	})

	for _, route := range []string{"create", "batch", "admin"} {
		for _, doctor := range []struct{ name, json string }{{"omitted", ""}, {"null", `,"doctor_id":null`}, {"zero", `,"doctor_id":0`}} {
			t.Run(route+"/"+doctor.name, func(t *testing.T) {
				f := seedReservationCreatedByActors(t, db, true, false)
				w := reservationCreatedByRequest(t, db, f, route, doctor.json)
				require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
				count := 1
				if route == "batch" {
					count = 2
				}
				assertReservationCreatedByReadback(t, db, f, count)
			})
		}
	}

	for _, tc := range []struct {
		name     string
		assigned bool
		admin    bool
		change   func(*testing.T, *gorm.DB, *reservationCreatedByActors)
	}{
		{name: "unassigned"},
		{name: "inactive staff", assigned: true, change: func(t *testing.T, db *gorm.DB, f *reservationCreatedByActors) {
			require.NoError(t, db.Model(&f.actor).UpdateColumn("is_active", false).Error)
		}},
		{name: "deleted staff", assigned: true, change: func(t *testing.T, db *gorm.DB, f *reservationCreatedByActors) {
			require.NoError(t, db.Delete(&f.actor).Error)
		}},
		{name: "deleted assignment", assigned: true, change: func(t *testing.T, db *gorm.DB, f *reservationCreatedByActors) {
			require.NoError(t, db.Where("staff_id = ? AND clinic_id = ?", f.actor.ID, f.clinicB.ID).Delete(&model.StaffClinicAssignment{}).Error)
		}},
		{name: "inactive admin account", admin: true, change: func(t *testing.T, db *gorm.DB, f *reservationCreatedByActors) {
			require.NoError(t, db.Model(&f.account).UpdateColumn("is_active", false).Error)
		}},
		{name: "deleted admin account", admin: true, change: func(t *testing.T, db *gorm.DB, f *reservationCreatedByActors) {
			require.NoError(t, db.Delete(&f.account).Error)
		}},
		{name: "revoked admin account", admin: true, change: func(t *testing.T, db *gorm.DB, f *reservationCreatedByActors) {
			require.NoError(t, db.Model(&f.account).UpdateColumn("is_system_admin", false).Error)
		}},
	} {
		t.Run("rejected/"+tc.name, func(t *testing.T) {
			f := seedReservationCreatedByActors(t, db, tc.assigned, tc.admin)
			if tc.change != nil {
				tc.change(t, db, f)
			}
			for _, route := range []string{"create", "batch", "admin"} {
				w := reservationCreatedByRequest(t, db, f, route, "")
				assert.Equal(t, http.StatusForbidden, w.Code, route+": "+w.Body.String())
				assert.NotContains(t, w.Body.String(), "fk_appointments")
			}
			var count int64
			require.NoError(t, db.Model(&model.Reservation{}).Where("clinic_id = ?", f.clinicB.ID).Count(&count).Error)
			assert.Zero(t, count, "all rejected entrypoints must leave no partial appointments")
		})
	}

	t.Run("history survives revoked assignment and staff soft deletion", func(t *testing.T) {
		f := seedReservationCreatedByActors(t, db, true, false)
		w := reservationCreatedByRequest(t, db, f, "create", "")
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		require.NoError(t, db.Where("staff_id = ? AND clinic_id = ?", f.actor.ID, f.clinicB.ID).Delete(&model.StaffClinicAssignment{}).Error)
		rows := assertReservationCreatedByReadback(t, db, f, 1)
		require.NoError(t, db.Delete(&f.actor).Error)
		loaded, err := NewReservationRepository(db).FindByID(context.Background(), f.clinicB.ID, rows[0].ID)
		require.NoError(t, err)
		require.NotNil(t, loaded.CreatedBy)
		assert.Equal(t, f.actor.ID, *loaded.CreatedBy)
		assert.Nil(t, loaded.CreatedByStaff, "deleted staff retain attribution ID without exposing profile")
	})

	t.Run("system admin without selected clinic assignment and history after revocation", func(t *testing.T) {
		f := seedReservationCreatedByActors(t, db, false, true)
		w := reservationCreatedByRequest(t, db, f, "admin", "")
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		assertReservationCreatedByReadback(t, db, f, 1)
		require.NoError(t, db.Model(&f.account).UpdateColumn("is_system_admin", false).Error)
		assertReservationCreatedByReadback(t, db, f, 1)
		items, err := NewReservationAdminRepository(db).FindAllByDay(context.Background(), f.clinicB.ID, f.input().StartTime)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.NotNil(t, items[0].CreatedByStaff)
		assert.Equal(t, f.actor.Name, items[0].CreatedByStaff.Name)
	})

	t.Run("same home clinic creator remains supported", func(t *testing.T) {
		f := seedReservationCreatedByActors(t, db, true, false)
		require.NoError(t, db.Model(&f.actor).UpdateColumn("clinic_id", f.clinicB.ID).Error)
		w := reservationCreatedByRequest(t, db, f, "create", "")
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		assertReservationCreatedByReadback(t, db, f, 1)
	})

	t.Run("internal nil actor remains supported", func(t *testing.T) {
		f := seedReservationCreatedByActors(t, db, false, false)
		input := f.input()
		input.CreatedBy = nil
		got, err := reservationCreatedByService(db).Create(context.Background(), input)
		require.NoError(t, err)
		assert.Nil(t, got.CreatedBy)
		loaded, err := NewReservationRepository(db).FindByID(context.Background(), f.clinicB.ID, got.ID)
		require.NoError(t, err)
		assert.Nil(t, loaded.CreatedBy)
	})

	t.Run("both repositories require ambient transaction for actor", func(t *testing.T) {
		f := seedReservationCreatedByActors(t, db, true, false)
		for _, create := range []func(context.Context, *model.Reservation) error{NewReservationRepository(db).Create, NewReservationAdminRepository(db).Create} {
			input := f.input()
			row := &model.Reservation{ClinicID: input.ClinicID, StartTime: input.StartTime, EndTime: input.EndTime, ReservationTypeID: input.ReservationTypeID, CreatedBy: input.CreatedBy}
			require.Error(t, create(context.Background(), row))
			assert.Zero(t, row.ID)
		}
	})

	for _, target := range []string{"staff", "assignment", "admin account"} {
		t.Run("authorization lock/"+target, func(t *testing.T) {
			f := seedReservationCreatedByActors(t, db, target != "admin account", target == "admin account")
			tx := db.Begin()
			require.NoError(t, tx.Error)
			t.Cleanup(func() {
				if err := tx.Rollback().Error; err != nil && !errors.Is(err, sql.ErrTxDone) {
					t.Error(err)
				}
			})
			input := f.input()
			row := &model.Reservation{ClinicID: input.ClinicID, StartTime: input.StartTime, EndTime: input.EndTime, ReservationTypeID: input.ReservationTypeID, CreatedBy: input.CreatedBy}
			ctx := persistence.WithTxValue(context.Background(), tx)
			require.NoError(t, NewReservationRepository(db).Create(ctx, row))
			contender := db.Begin()
			require.NoError(t, contender.Error)
			t.Cleanup(func() {
				if err := contender.Rollback().Error; err != nil && !errors.Is(err, sql.ErrTxDone) {
					t.Error(err)
				}
			})
			require.NoError(t, contender.Exec("SET LOCAL lock_timeout = '100ms'").Error)
			var err error
			switch target {
			case "staff":
				err = contender.Model(&f.actor).UpdateColumn("is_active", false).Error
			case "assignment":
				err = contender.Model(&model.StaffClinicAssignment{}).Where("staff_id = ? AND clinic_id = ?", f.actor.ID, f.clinicB.ID).UpdateColumn("deleted_at", time.Now()).Error
			case "admin account":
				err = contender.Model(&f.account).UpdateColumn("is_system_admin", false).Error
			}
			var pgErr *pgconn.PgError
			require.ErrorAs(t, err, &pgErr, "revocation must wait for the booking transaction")
			assert.Equal(t, "55P03", pgErr.Code)
			require.NoError(t, contender.Rollback().Error)
			require.NoError(t, tx.Commit().Error)
			// Committing the booking releases its authorization locks.
			switch target {
			case "staff":
				require.NoError(t, db.Model(&f.actor).UpdateColumn("is_active", false).Error)
			case "assignment":
				require.NoError(t, db.Where("staff_id = ? AND clinic_id = ?", f.actor.ID, f.clinicB.ID).Delete(&model.StaffClinicAssignment{}).Error)
			case "admin account":
				require.NoError(t, db.Model(&f.account).UpdateColumn("is_system_admin", false).Error)
			}
		})
	}

	t.Run("zero actor cannot masquerade as internal creation", func(t *testing.T) {
		f := seedReservationCreatedByActors(t, db, true, false)
		input := f.input()
		zero := uint64(0)
		input.CreatedBy = &zero
		_, err := reservationCreatedByService(db).Create(context.Background(), input)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}
