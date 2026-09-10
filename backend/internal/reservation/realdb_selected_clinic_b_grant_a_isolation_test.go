package reservation

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 reservation
// clinic-fixed 12 + cross-clinic 3 through real repo/service/HTTP.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBResTypeSeedNameA     = "D3res-realdb-type-A"
	realDBResTypeSeedNameB     = "D3res-realdb-type-B"
	realDBResGroupSeedNameA    = "D3res-realdb-group-A"
	realDBResGroupSeedNameB    = "D3res-realdb-group-B"
	realDBResStaffSeedNameA    = "D3res-realdb-staff-A"
	realDBResStaffSeedNameB    = "D3res-realdb-staff-B"
	realDBResOccSeedNameB      = "D3res-realdb-occupation-B"
	realDBResNotesA            = "D3res-realdb-notes-A"
	realDBResNotesB            = "D3res-realdb-notes-B"
	realDBResLineHeaderA       = "D3res-realdb-line-header-A"
	realDBResLineHeaderB       = "D3res-realdb-line-header-B"
	realDBResSlotStartB        = "1030"
	realDBResUnavailableStartB = "1200"
	realDBResAvailableStartB   = "1400"
	realDBResScheduleMonth     = "2026-09"
	realDBResScheduleDay       = "2026-09-15"
	realDBResAdminDay          = "2026-09-15"
)

type realDBResNameDTO struct {
	ID       uint64 `json:"id"`
	ClinicID uint64 `json:"clinic_id"`
	Name     string `json:"name"`
}
type realDBResStaffDTO struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}
type realDBResNotesDTO struct {
	ID       uint64 `json:"id"`
	ClinicID uint64 `json:"clinic_id"`
	Notes    string `json:"notes"`
}
type realDBResSlotDTO struct {
	StartTime string `json:"start_time"`
}
type realDBResUnavailableDTO struct {
	ID        uint64 `json:"id"`
	StartTime string `json:"start_time"`
}
type realDBResAvailableDTO struct {
	ID        uint64 `json:"id"`
	StartTime string `json:"start_time"`
}
type realDBResOccupationDTO struct {
	OccupationID uint64 `json:"occupation_id"`
}
type realDBResLineSettingDTO struct {
	ClinicID   uint64 `json:"clinic_id"`
	HeaderText string `json:"header_text"`
}
type realDBResScheduleDTO struct {
	StaffID uint64 `json:"staff_id"`
}

type occupationFinderStub struct{ db *gorm.DB }

func (s occupationFinderStub) FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	var occ model.Occupation
	err := s.db.WithContext(ctx).Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, id).First(&occ).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "occupation", fmt.Sprintf("%d", id))
	}
	return &occ, nil
}

type availableTimesStub struct {
	t                 *testing.T
	forbiddenClinicID uint64
	clinicB           uint64
	typeAID           uint64
	typeBID           uint64
}

func (s *availableTimesStub) GetAvailableTimes(_ context.Context, clinicID, typeID, _ uint64, _ time.Time) ([]TimeSlot, error) {
	if s.forbiddenClinicID != 0 && clinicID == s.forbiddenClinicID {
		s.t.Fatalf("GetAvailableTimes must not query clinic %d without selected-clinic grant", clinicID)
	}
	if clinicID == s.clinicB && typeID == s.typeBID {
		return []TimeSlot{{StartTime: realDBResSlotStartB, EndTime: "1100"}}, nil
	}
	if typeID == s.typeAID {
		return nil, apperrors.WrapNotFound("reservation_type", fmt.Sprintf("%d", typeID))
	}
	return []TimeSlot{}, nil
}

type clinicScopedGuard struct {
	t                 *testing.T
	forbiddenClinicID uint64
}

func (g clinicScopedGuard) reject(clinicID uint64, op string) {
	if g.forbiddenClinicID != 0 && clinicID == g.forbiddenClinicID {
		g.t.Fatalf("%s must not query clinic %d without selected-clinic grant", op, clinicID)
	}
}

type reservationQueryGuardService struct {
	ReservationService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *reservationQueryGuardService) List(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	for _, id := range clinicIDs {
		if id == s.forbiddenClinicID {
			s.t.Fatalf("List must not query clinic %d without selected-clinic grant", id)
		}
	}
	return s.ReservationService.List(ctx, clinicIDs, page, limit, date, startDate, endDate, status, source, petID, ownerID)
}
func (s *reservationQueryGuardService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error) {
	for _, clinicID := range clinicIDs {
		if clinicID == s.forbiddenClinicID {
			s.t.Fatalf("GetByIDForClinics must not query clinic %d without selected-clinic grant", clinicID)
		}
	}
	return s.ReservationService.GetByIDForClinics(ctx, clinicIDs, id)
}

type reservationTypeQueryGuardService struct {
	ReservationTypeService
	clinicScopedGuard
}

func (s *reservationTypeQueryGuardService) List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	s.reject(clinicID, "reservation-type List")
	return s.ReservationTypeService.List(ctx, clinicID)
}
func (s *reservationTypeQueryGuardService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	s.reject(clinicID, "reservation-type GetByID")
	return s.ReservationTypeService.GetByID(ctx, clinicID, id)
}
func (s *reservationTypeQueryGuardService) ListUnavailableTimes(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error) {
	s.reject(clinicID, "ListUnavailableTimes")
	return s.ReservationTypeService.ListUnavailableTimes(ctx, clinicID, reservationTypeID)
}
func (s *reservationTypeQueryGuardService) ListAvailableSlots(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error) {
	s.reject(clinicID, "ListAvailableSlots")
	return s.ReservationTypeService.ListAvailableSlots(ctx, clinicID, reservationTypeID)
}
func (s *reservationTypeQueryGuardService) ListOccupations(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error) {
	s.reject(clinicID, "ListOccupations")
	return s.ReservationTypeService.ListOccupations(ctx, clinicID, reservationTypeID)
}

type reservationTypeGroupQueryGuardService struct {
	ReservationTypeGroupService
	clinicScopedGuard
}

func (s *reservationTypeGroupQueryGuardService) List(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error) {
	s.reject(clinicID, "group List")
	return s.ReservationTypeGroupService.List(ctx, clinicID)
}
func (s *reservationTypeGroupQueryGuardService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error) {
	s.reject(clinicID, "group GetByID")
	return s.ReservationTypeGroupService.GetByID(ctx, clinicID, id)
}

type reservationTypeLiffQueryGuardService struct {
	ReservationTypeLiffService
	clinicScopedGuard
}

func (s *reservationTypeLiffQueryGuardService) List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	s.reject(clinicID, "liff-type List")
	return s.ReservationTypeLiffService.List(ctx, clinicID)
}

type reservationStaffQueryGuardService struct {
	ReservationStaffService
	clinicScopedGuard
}

func (s *reservationStaffQueryGuardService) List(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	s.reject(clinicID, "staff List")
	return s.ReservationStaffService.List(ctx, clinicID)
}
func (s *reservationStaffQueryGuardService) ListExcludedByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) (map[uint64][]model.StaffReservationExclusion, error) {
	s.reject(clinicID, "ListExcludedByStaffIDs")
	return s.ReservationStaffService.ListExcludedByStaffIDs(ctx, clinicID, staffIDs)
}
func (s *reservationStaffQueryGuardService) ListCapableByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) (map[uint64][]model.StaffReservationCapability, error) {
	s.reject(clinicID, "ListCapableByStaffIDs")
	return s.ReservationStaffService.ListCapableByStaffIDs(ctx, clinicID, staffIDs)
}

type reservationScheduleQueryGuardService struct {
	ReservationScheduleService
	clinicScopedGuard
}

func (s *reservationScheduleQueryGuardService) ListByMonth(ctx context.Context, clinicID, staffID uint64, month string) ([]ScheduleEntry, error) {
	s.reject(clinicID, "schedule ListByMonth")
	return s.ReservationScheduleService.ListByMonth(ctx, clinicID, staffID, month)
}

type reservationAdminQueryGuardService struct {
	ReservationAdminService
	clinicScopedGuard
}

func (s *reservationAdminQueryGuardService) ListByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error) {
	s.reject(clinicID, "admin ListByDay")
	return s.ReservationAdminService.ListByDay(ctx, clinicID, date)
}
func (s *reservationAdminQueryGuardService) ListByMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.Reservation, error) {
	s.reject(clinicID, "admin ListByMonth")
	return s.ReservationAdminService.ListByMonth(ctx, clinicID, yearMonth)
}

type lineReservationSettingQueryGuardService struct {
	LineReservationSettingService
	t     *testing.T
	armed bool
}

func (s *lineReservationSettingQueryGuardService) Get(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	if s.armed {
		s.t.Fatalf("line reservation setting Get must not run without hospital-settings grant (clinicID=%d)", clinicID)
	}
	return s.LineReservationSettingService.Get(ctx, clinicID)
}

type realDBReservationHandlers struct {
	typeHandler        *ReservationTypeHandler
	groupHandler       *ReservationTypeGroupHandler
	liffTypeHandler    *ReservationTypeLiffHandler
	staffHandler       *ReservationStaffHandler
	scheduleHandler    *ReservationScheduleHandler
	adminHandler       *ReservationAdminHandler
	crudHandler        *CRUDHandler
	lineSettingHandler *LineReservationSettingHandler
}

type realDBReservationFixture struct {
	fx             testdb.ClinicGrantFixture
	typeA, typeB   *model.ReservationType
	groupA, groupB *model.ReservationTypeGroup
	staffA, staffB *model.Staff
	occB           *model.Occupation
	resA, resB     *model.Reservation
	unavailableB   *model.ReservationTypeUnavailableTime
	availableSlotB *model.ReservationTypeAvailableSlot
	lineA, lineB   *model.LineReservationSetting
	handlers       realDBReservationHandlers
}

func setupRealDBReservationIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.Occupation{}, &model.ReservationTypeGroup{}, &model.ReservationType{},
		&model.ReservationTypeUnavailableTime{}, &model.ReservationTypeAvailableSlot{},
		&model.ReservationTypeOccupation{}, &model.Reservation{}, &model.ShiftEntry{},
		&model.ShiftEntryBreak{}, &model.LineReservationSetting{},
		&model.StaffReservationCapability{}, &model.StaffReservationExclusion{},
	))
	testdb.Truncate(t, db,
		"staff_reservation_capabilities", "staff_reservation_exclusions",
		"reservation_type_occupations", "reservation_type_available_slots",
		"reservation_type_unavailable_times", "appointments", "shift_entry_breaks",
		"shift_entries", "line_reservation_settings", "reservation_types",
		"reservation_type_groups", "staff_clinic_assignments", "staffs", "occupations",
	)
	return db
}

func seedClinicStaffWithAssignment(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	staff := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor, IsActive: true, ReservationVisible: true}
	require.NoError(t, db.WithContext(context.Background()).Create(staff).Error)
	require.NoError(t, db.WithContext(context.Background()).Create(&model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID, IsMain: true}).Error)
	return staff
}

func newRealDBReservationHandlers(t *testing.T, db *gorm.DB, fx realDBReservationFixture, forbiddenClinicID uint64) realDBReservationHandlers {
	t.Helper()
	typeRepo := NewReservationTypeRepository(db)
	unavailableRepo := NewReservationTypeUnavailableTimeRepository(db)
	slotRepo := NewReservationTypeAvailableSlotRepository(db)
	occRepo := NewReservationTypeOccupationRepository(db)
	groupRepo := NewReservationTypeGroupRepository(db)
	resStore := NewReservationRepository(db)
	adminRepo := NewReservationAdminRepository(db)
	staffRepo := NewReservationStaffRepository(db, nil)
	scheduleRepo := NewReservationScheduleRepository(db, nil)
	lineRepo := NewLineReservationSettingRepository(db)

	typeSvc := ReservationTypeService(NewReservationTypeService(typeRepo, unavailableRepo, occRepo, occupationFinderStub{db: db}, groupRepo, slotRepo))
	groupSvc := ReservationTypeGroupService(NewReservationTypeGroupService(groupRepo))
	liffTypeSvc := ReservationTypeLiffService(NewReservationTypeLiffService(NewReservationTypeLiffRepository(db), resStore))
	staffSvc := ReservationStaffService(NewReservationStaffService(staffRepo, &mockTransactor{}, nil))
	scheduleSvc := ReservationScheduleService(NewReservationScheduleService(scheduleRepo))
	adminSvc := ReservationAdminService(NewReservationAdminServiceWithAvailabilityAndType(adminRepo, resStore, typeRepo, &mockTransactor{}, nil, nil))
	crudSvc := ReservationService(NewReservationServiceWithAvailabilityAndType(resStore, typeRepo, &mockTransactor{}, nil, nil))
	lineSvc := LineReservationSettingService(NewLineReservationSettingService(lineRepo, nil, nil))
	avail := &availableTimesStub{t: t, clinicB: fx.fx.ClinicB, typeAID: fx.typeA.ID, typeBID: fx.typeB.ID}
	if forbiddenClinicID != 0 {
		guard := clinicScopedGuard{t: t, forbiddenClinicID: forbiddenClinicID}
		typeSvc = &reservationTypeQueryGuardService{ReservationTypeService: typeSvc, clinicScopedGuard: guard}
		groupSvc = &reservationTypeGroupQueryGuardService{ReservationTypeGroupService: groupSvc, clinicScopedGuard: guard}
		liffTypeSvc = &reservationTypeLiffQueryGuardService{ReservationTypeLiffService: liffTypeSvc, clinicScopedGuard: guard}
		staffSvc = &reservationStaffQueryGuardService{ReservationStaffService: staffSvc, clinicScopedGuard: guard}
		scheduleSvc = &reservationScheduleQueryGuardService{ReservationScheduleService: scheduleSvc, clinicScopedGuard: guard}
		adminSvc = &reservationAdminQueryGuardService{ReservationAdminService: adminSvc, clinicScopedGuard: guard}
		crudSvc = &reservationQueryGuardService{ReservationService: crudSvc, t: t, forbiddenClinicID: forbiddenClinicID}
		lineSvc = &lineReservationSettingQueryGuardService{LineReservationSettingService: lineSvc, t: t, armed: true}
		avail.forbiddenClinicID = forbiddenClinicID
	}
	return realDBReservationHandlers{
		typeHandler:        NewReservationTypeHandler(typeSvc, typeSvc, typeSvc, typeSvc),
		groupHandler:       NewReservationTypeGroupHandler(groupSvc),
		liffTypeHandler:    NewReservationTypeLiffHandler(liffTypeSvc),
		staffHandler:       NewReservationStaffHandler(staffSvc),
		scheduleHandler:    NewReservationScheduleHandler(scheduleSvc),
		adminHandler:       NewReservationAdminHandler(adminSvc, nil),
		crudHandler:        NewCRUDHandler(crudSvc, nil, avail, nil),
		lineSettingHandler: NewLineReservationSettingHandler(lineSvc),
	}
}

func seedRealDBReservationFixture(t *testing.T, db *gorm.DB, forbiddenClinicID uint64) realDBReservationFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3res realDB")
	ctx := context.Background()
	groupA := &model.ReservationTypeGroup{ClinicID: fx.ClinicA, Name: realDBResGroupSeedNameA, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(groupA).Error)
	groupB := &model.ReservationTypeGroup{ClinicID: fx.ClinicB, Name: realDBResGroupSeedNameB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(groupB).Error)
	typeA := &model.ReservationType{ClinicID: fx.ClinicA, Name: realDBResTypeSeedNameA, Category: model.ReservationTypeCategoryGeneral, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(typeA).Error)
	typeB := &model.ReservationType{ClinicID: fx.ClinicB, Name: realDBResTypeSeedNameB, Category: model.ReservationTypeCategoryGeneral, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(typeB).Error)
	staffA := seedClinicStaffWithAssignment(t, db, fx.ClinicA, realDBResStaffSeedNameA)
	staffB := seedClinicStaffWithAssignment(t, db, fx.ClinicB, realDBResStaffSeedNameB)
	occB := &model.Occupation{ClinicID: fx.ClinicB, Name: realDBResOccSeedNameB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(occB).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ReservationTypeOccupation{ClinicID: fx.ClinicB, ReservationTypeID: typeB.ID, OccupationID: occB.ID}).Error)
	dow := int8(1)
	unavailableB := &model.ReservationTypeUnavailableTime{ClinicID: fx.ClinicB, ReservationTypeID: typeB.ID, UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: &dow, StartTime: realDBResUnavailableStartB, EndTime: "1300"}
	require.NoError(t, db.WithContext(ctx).Create(unavailableB).Error)
	availableSlotB := &model.ReservationTypeAvailableSlot{ClinicID: fx.ClinicB, ReservationTypeID: typeB.ID, AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &dow, StartTime: realDBResAvailableStartB, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(availableSlotB).Error)
	day, err := time.ParseInLocation(time.DateOnly, realDBResScheduleDay, time.Local)
	require.NoError(t, err)
	require.NoError(t, db.WithContext(ctx).Create(&model.ShiftEntry{ClinicID: fx.ClinicA, StaffID: staffA.ID, Date: day, ShiftType: model.ShiftTypeFull}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ShiftEntry{ClinicID: fx.ClinicB, StaffID: staffB.ID, Date: day, ShiftType: model.ShiftTypeFull}).Error)
	startA := time.Date(2026, 9, 15, 10, 0, 0, 0, time.Local)
	resA := &model.Reservation{ClinicID: fx.ClinicA, StartTime: startA, EndTime: startA.Add(30 * time.Minute), ReservationTypeID: typeA.ID, Status: model.ReservationStatusConfirmed, Source: model.ReservationSourceManual, Notes: realDBResNotesA, CustomerFields: json.RawMessage(`{}`)}
	require.NoError(t, db.WithContext(ctx).Create(resA).Error)
	startB := time.Date(2026, 9, 15, 11, 0, 0, 0, time.Local)
	resB := &model.Reservation{ClinicID: fx.ClinicB, StartTime: startB, EndTime: startB.Add(30 * time.Minute), ReservationTypeID: typeB.ID, Status: model.ReservationStatusConfirmed, Source: model.ReservationSourceManual, Notes: realDBResNotesB, CustomerFields: json.RawMessage(`{}`)}
	require.NoError(t, db.WithContext(ctx).Create(resB).Error)
	lineA := &model.LineReservationSetting{ClinicID: fx.ClinicA, Status: "running", HeaderText: realDBResLineHeaderA, ClosedWeekdays: json.RawMessage(`[]`), ClosedDates: json.RawMessage(`[]`), BusinessHours: json.RawMessage(`{"start":"0900","end":"1900"}`), BreakHours: json.RawMessage(`[]`), AdditionalFields: json.RawMessage(`[]`)}
	require.NoError(t, db.WithContext(ctx).Create(lineA).Error)
	lineB := &model.LineReservationSetting{ClinicID: fx.ClinicB, Status: "running", HeaderText: realDBResLineHeaderB, ClosedWeekdays: json.RawMessage(`[]`), ClosedDates: json.RawMessage(`[]`), BusinessHours: json.RawMessage(`{"start":"0900","end":"1900"}`), BreakHours: json.RawMessage(`[]`), AdditionalFields: json.RawMessage(`[]`)}
	require.NoError(t, db.WithContext(ctx).Create(lineB).Error)
	out := realDBReservationFixture{fx: fx, typeA: typeA, typeB: typeB, groupA: groupA, groupB: groupB, staffA: staffA, staffB: staffB, occB: occB, resA: resA, resB: resB, unavailableB: unavailableB, availableSlotB: availableSlotB, lineA: lineA, lineB: lineB}
	out.handlers = newRealDBReservationHandlers(t, db, out, forbiddenClinicID)
	return out
}

func configureResGrant(fx realDBReservationFixture, resource string, grantClinicID uint64) func(*gin.Context) {
	return testdb.ConfigureSelectedClinicBGrant(fx.fx, resource, grantClinicID)
}
func withParams(configure func(*gin.Context), params gin.Params) func(*gin.Context) {
	return func(c *gin.Context) { configure(c); c.Params = params }
}

func TestRealDB_SelectedClinicBGrantAIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("reservation_types_list_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/reservation-types", configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicA))
		fx.handlers.typeHandler.ListReservationTypes(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA, fx.fx.ClinicB}, realDBResTypeSeedNameA, realDBResTypeSeedNameB)
	})
	t.Run("reservation_types_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/reservation-types", configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB))
		fx.handlers.typeHandler.ListReservationTypes(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []realDBResNameDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBResTypeSeedNameA, item.Name)
			if item.ID == fx.typeB.ID {
				foundB = true
				assert.Equal(t, realDBResTypeSeedNameB, item.Name)
			}
		}
		require.True(t, foundB)
	})
	t.Run("reservation_types_get_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d", fx.typeB.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicA), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeB.ID)}}))
		fx.handlers.typeHandler.GetReservationType(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("reservation_types_get_grantB_ok", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d", fx.typeB.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeB.ID)}}))
		fx.handlers.typeHandler.GetReservationType(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBResTypeSeedNameB)
		assert.NotContains(t, w.Body.String(), realDBResTypeSeedNameA)
	})
	t.Run("reservation_types_get_A_id_404", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d", fx.typeA.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeA.ID)}}))
		fx.handlers.typeHandler.GetReservationType(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBResTypeSeedNameA)
	})

	t.Run("reservation_type_groups_list_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/reservation-type-groups", configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicA))
		fx.handlers.groupHandler.ListReservationTypeGroups(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("reservation_type_groups_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/reservation-type-groups", configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB))
		fx.handlers.groupHandler.ListReservationTypeGroups(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []realDBResNameDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBResGroupSeedNameA, item.Name)
			if item.ID == fx.groupB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
	})
	t.Run("reservation_type_groups_get_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-type-groups/%d", fx.groupB.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicA), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.groupB.ID)}}))
		fx.handlers.groupHandler.GetReservationTypeGroup(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("reservation_type_groups_get_grantB_ok", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-type-groups/%d", fx.groupB.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.groupB.ID)}}))
		fx.handlers.groupHandler.GetReservationTypeGroup(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBResGroupSeedNameB)
		assert.NotContains(t, w.Body.String(), realDBResGroupSeedNameA)
	})
	t.Run("reservation_type_groups_get_A_id_404", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-type-groups/%d", fx.groupA.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.groupA.ID)}}))
		fx.handlers.groupHandler.GetReservationTypeGroup(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBResGroupSeedNameA)
	})

	t.Run("unavailable_times_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d/unavailable-times", fx.typeB.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicA), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeB.ID)}}))
		fx.handlers.typeHandler.ListUnavailableTimes(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("unavailable_times_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d/unavailable-times", fx.typeB.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeB.ID)}}))
		fx.handlers.typeHandler.ListUnavailableTimes(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []realDBResUnavailableDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		found := false
		for _, item := range listed {
			if item.ID == fx.unavailableB.ID {
				found = true
				assert.Equal(t, realDBResUnavailableStartB, item.StartTime)
			}
		}
		require.True(t, found)
	})
	t.Run("unavailable_times_A_type_404", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d/unavailable-times", fx.typeA.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeA.ID)}}))
		fx.handlers.typeHandler.ListUnavailableTimes(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("available_slots_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d/available-slots", fx.typeB.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicA), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeB.ID)}}))
		fx.handlers.typeHandler.ListAvailableSlots(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("available_slots_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d/available-slots", fx.typeB.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeB.ID)}}))
		fx.handlers.typeHandler.ListAvailableSlots(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []realDBResAvailableDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		found := false
		for _, item := range listed {
			if item.ID == fx.availableSlotB.ID {
				found = true
				assert.Equal(t, realDBResAvailableStartB, item.StartTime)
			}
		}
		require.True(t, found)
	})
	t.Run("available_slots_A_type_404", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d/available-slots", fx.typeA.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeA.ID)}}))
		fx.handlers.typeHandler.ListAvailableSlots(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("occupations_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d/occupations", fx.typeB.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicA), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeB.ID)}}))
		fx.handlers.typeHandler.ListReservationTypeOccupations(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("occupations_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d/occupations", fx.typeB.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeB.ID)}}))
		fx.handlers.typeHandler.ListReservationTypeOccupations(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []realDBResOccupationDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		found := false
		for _, item := range listed {
			if item.OccupationID == fx.occB.ID {
				found = true
			}
		}
		require.True(t, found)
		assert.Contains(t, w.Body.String(), realDBResOccSeedNameB)
	})
	t.Run("occupations_A_type_404", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/masters/reservation-types/%d/occupations", fx.typeA.ID), withParams(configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.typeA.ID)}}))
		fx.handlers.typeHandler.ListReservationTypeOccupations(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("liff_types_list_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/reservation-types", fx.fx.ClinicB), configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicA))
		fx.handlers.liffTypeHandler.ListReservationTypeLiffs(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("liff_types_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/reservation-types", fx.fx.ClinicB), configureResGrant(fx, string(model.ResourceMasterReservationType), fx.fx.ClinicB))
		fx.handlers.liffTypeHandler.ListReservationTypeLiffs(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []realDBResNameDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.NotEqual(t, realDBResTypeSeedNameA, item.Name)
			if item.ID == fx.typeB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
	})

	t.Run("reservation_staffs_list_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/reservation-staffs", fx.fx.ClinicB), configureResGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicA))
		fx.handlers.staffHandler.ListReservationStaffs(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBResStaffSeedNameB)
	})
	t.Run("reservation_staffs_list_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/reservation-staffs", fx.fx.ClinicB), configureResGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB))
		fx.handlers.staffHandler.ListReservationStaffs(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []realDBResStaffDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.NotEqual(t, realDBResStaffSeedNameA, item.Name)
			if item.ID == fx.staffB.ID {
				foundB = true
				assert.Equal(t, realDBResStaffSeedNameB, item.Name)
			}
		}
		require.True(t, foundB)
	})

	t.Run("schedules_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/reservation-staffs/%d/schedules?month=%s", fx.fx.ClinicB, fx.staffB.ID, realDBResScheduleMonth), withParams(configureResGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicA), gin.Params{{Key: "staffId", Value: fmt.Sprintf("%d", fx.staffB.ID)}}))
		fx.handlers.scheduleHandler.ListReservationSchedules(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("schedules_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/reservation-staffs/%d/schedules?month=%s", fx.fx.ClinicB, fx.staffB.ID, realDBResScheduleMonth), withParams(configureResGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), gin.Params{{Key: "staffId", Value: fmt.Sprintf("%d", fx.staffB.ID)}}))
		fx.handlers.scheduleHandler.ListReservationSchedules(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []realDBResScheduleDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		foundB := false
		for _, item := range listed {
			assert.NotEqual(t, fx.staffA.ID, item.StaffID)
			if item.StaffID == fx.staffB.ID {
				foundB = true
			}
		}
		require.True(t, foundB)
	})
	t.Run("schedules_A_staff_empty_no_leak", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/reservation-staffs/%d/schedules?month=%s", fx.fx.ClinicB, fx.staffA.ID, realDBResScheduleMonth), withParams(configureResGrant(fx, string(model.ResourceMasterStaff), fx.fx.ClinicB), gin.Params{{Key: "staffId", Value: fmt.Sprintf("%d", fx.staffA.ID)}}))
		fx.handlers.scheduleHandler.ListReservationSchedules(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "[]", w.Body.String())
	})

	t.Run("admin_reservations_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/reservations?view=day&date=%s", fx.fx.ClinicB, realDBResAdminDay), configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicA))
		fx.handlers.adminHandler.ListReservationsAdmin(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBResNotesB)
	})
	t.Run("admin_reservations_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/reservations?view=day&date=%s", fx.fx.ClinicB, realDBResAdminDay), configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicB))
		fx.handlers.adminHandler.ListReservationsAdmin(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBResNotesB)
		assert.NotContains(t, w.Body.String(), realDBResNotesA)
	})

	t.Run("available_times_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		path := fmt.Sprintf("/api/v1/reservations/available-times?reservation_type_id=%d&staff_id=%d&date=%s", fx.typeB.ID, fx.staffB.ID, realDBResAdminDay)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, path, configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicA))
		fx.handlers.crudHandler.GetReservationAvailableTimes(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBResSlotStartB)
	})
	t.Run("available_times_grantB_nonempty", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		path := fmt.Sprintf("/api/v1/reservations/available-times?reservation_type_id=%d&staff_id=%d&date=%s", fx.typeB.ID, fx.staffB.ID, realDBResAdminDay)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, path, configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicB))
		fx.handlers.crudHandler.GetReservationAvailableTimes(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed []realDBResSlotDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed)
		require.Equal(t, realDBResSlotStartB, listed[0].StartTime)
	})
	t.Run("available_times_A_type_404", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		path := fmt.Sprintf("/api/v1/reservations/available-times?reservation_type_id=%d&staff_id=%d&date=%s", fx.typeA.ID, fx.staffB.ID, realDBResAdminDay)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, path, configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicB))
		fx.handlers.crudHandler.GetReservationAvailableTimes(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		assert.NotContains(t, w.Body.String(), realDBResSlotStartB)
	})
}

func TestRealDB_CrossClinic_Reservations_ConditionSeparated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list_grantA_selectedB_no_clinic_ids_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/reservations?page=1&limit=50", configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicA))
		fx.handlers.crudHandler.ListReservations(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBResNotesA)
		assert.NotContains(t, w.Body.String(), realDBResNotesB)
	})
	t.Run("list_grantA_selectedB_clinic_ids_AB_returns_A_only", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		path := fmt.Sprintf("/api/v1/reservations?page=1&limit=50&clinic_ids=%d,%d", fx.fx.ClinicA, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, path, configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicA))
		fx.handlers.crudHandler.ListReservations(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed httpapi.PaginatedResponse[[]realDBResNotesDTO]
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed.Data)
		foundA := false
		for _, item := range listed.Data {
			assert.Equal(t, fx.fx.ClinicA, item.ClinicID)
			assert.NotEqual(t, realDBResNotesB, item.Notes)
			if item.ID == fx.resA.ID {
				foundA = true
				assert.Equal(t, realDBResNotesA, item.Notes)
			}
		}
		require.True(t, foundA)
	})
	t.Run("list_grantB_selectedB_returns_B_only", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/reservations?page=1&limit=50", configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicB))
		fx.handlers.crudHandler.ListReservations(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var listed httpapi.PaginatedResponse[[]realDBResNotesDTO]
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
		require.NotEmpty(t, listed.Data)
		foundB := false
		for _, item := range listed.Data {
			assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
			assert.NotEqual(t, realDBResNotesA, item.Notes)
			if item.ID == fx.resB.ID {
				foundB = true
				assert.Equal(t, realDBResNotesB, item.Notes)
			}
		}
		require.True(t, foundB)
	})
	t.Run("get_grantA_selectedB_returns_A", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/reservations/%d", fx.resA.ID), withParams(configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicA), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.resA.ID)}}))
		fx.handlers.crudHandler.GetReservation(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBResNotesA)
		assert.NotContains(t, w.Body.String(), realDBResNotesB)
	})
	t.Run("get_grantB_selectedB_A_id_404", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/reservations/%d", fx.resA.ID), withParams(configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.resA.ID)}}))
		fx.handlers.crudHandler.GetReservation(c)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBResNotesA)
	})
	t.Run("get_grantB_selectedB_returns_B", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/reservations/%d", fx.resB.ID), withParams(configureResGrant(fx, string(model.ResourceReservations), fx.fx.ClinicB), gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.resB.ID)}}))
		fx.handlers.crudHandler.GetReservation(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), realDBResNotesB)
		assert.NotContains(t, w.Body.String(), realDBResNotesA)
	})
}

func TestRealDB_CrossClinic_LineReservationSettings_ConditionSeparated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("path_B_grantA_403", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		fx.handlers = newRealDBReservationHandlers(t, db, fx, fx.fx.ClinicB)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/line-reservation-settings", fx.fx.ClinicB), withParams(configureResGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicA), gin.Params{{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicB)}}))
		fx.handlers.lineSettingHandler.GetLineReservationSetting(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.NotContains(t, w.Body.String(), realDBResLineHeaderA)
		assert.NotContains(t, w.Body.String(), realDBResLineHeaderB)
	})
	t.Run("path_A_grantA_returns_A", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/line-reservation-settings", fx.fx.ClinicA), withParams(configureResGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicA), gin.Params{{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicA)}}))
		fx.handlers.lineSettingHandler.GetLineReservationSetting(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got realDBResLineSettingDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, fx.fx.ClinicA, got.ClinicID)
		assert.Equal(t, realDBResLineHeaderA, got.HeaderText)
		assert.NotEqual(t, realDBResLineHeaderB, got.HeaderText)
	})
	t.Run("path_B_grantB_returns_B", func(t *testing.T) {
		db := setupRealDBReservationIsolationTestDB(t)
		fx := seedRealDBReservationFixture(t, db, 0)
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, fmt.Sprintf("/api/v1/clinics/%d/line-reservation-settings", fx.fx.ClinicB), withParams(configureResGrant(fx, string(model.ResourceHospitalSettings), fx.fx.ClinicB), gin.Params{{Key: "clinic_id", Value: fmt.Sprintf("%d", fx.fx.ClinicB)}}))
		fx.handlers.lineSettingHandler.GetLineReservationSetting(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got realDBResLineSettingDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, fx.fx.ClinicB, got.ClinicID)
		assert.Equal(t, realDBResLineHeaderB, got.HeaderText)
		assert.NotEqual(t, realDBResLineHeaderA, got.HeaderText)
	})
}
