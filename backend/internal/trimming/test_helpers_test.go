package trimming

import (
	"context"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testdb.CloseSharedTestDB()
	os.Exit(code)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.SetupTestDB(t)
}

func setupIsolatedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.SetupIsolatedTestDB(t)
}

func ensureAutoMigrated(db *gorm.DB, models ...any) error {
	return testdb.EnsureAutoMigrated(db, models...)
}

func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}

func strPtr(value string) *string {
	return &value
}

func ptrString(value string) *string {
	return &value
}

func ptrUint64(value uint64) *uint64 {
	return &value
}

type testTransactor struct {
	db *gorm.DB
}

type noopTrimmingAuditTxLogger struct{}

func (noopTrimmingAuditTxLogger) LogEntryTx(context.Context, *AuditEntry) error {
	return nil
}

type trimmingServiceWithTestActor struct {
	Service
}

func withTrimmingTestActor(svc Service) Service {
	return &trimmingServiceWithTestActor{Service: svc}
}

func (s *trimmingServiceWithTestActor) Create(
	ctx context.Context,
	clinicID uint64,
	input *CreateTrimmingInput,
) (*model.Reservation, error) {
	if input == nil || input.ActorID != nil {
		return s.Service.Create(ctx, clinicID, input)
	}
	cloned := *input
	cloned.ActorID = ptrUint64(42)
	return s.Service.Create(ctx, clinicID, &cloned)
}

func (s *trimmingServiceWithTestActor) Update(
	ctx context.Context,
	clinicID, id uint64,
	input *UpdateTrimmingInput,
) (*model.Reservation, error) {
	if input == nil || input.ActorID != nil {
		return s.Service.Update(ctx, clinicID, id, input)
	}
	cloned := *input
	cloned.ActorID = ptrUint64(42)
	return s.Service.Update(ctx, clinicID, id, &cloned)
}

func (s *trimmingServiceWithTestActor) Delete(
	ctx context.Context,
	clinicID, id uint64,
	actorID *uint64,
) error {
	if actorID == nil {
		actorID = ptrUint64(42)
	}
	return s.Service.Delete(ctx, clinicID, id, actorID)
}

func newTestTransactor(db *gorm.DB) Transactor {
	return &testTransactor{db: db}
}

func (t *testTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	})
}

func okTrimmingCourseRepo() TrimmingCourseRepository {
	return &mockTrimmingCourseRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		return &model.TrimmingCourse{ID: id, IsActive: true}, nil
	}}
}

func okTrimmingOptionRepo() TrimmingOptionRepository {
	return &mockTrimmingOptionRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		return &model.TrimmingOption{ID: id, IsActive: true}, nil
	}}
}

// mockReservationStaffRepository carries only the capability consumed by trimming.
type mockReservationStaffRepository struct {
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	supportsReservationTypeFn func(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
}

func (m *mockReservationStaffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockReservationStaffRepository) SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
	if m.supportsReservationTypeFn != nil {
		return m.supportsReservationTypeFn(ctx, clinicID, staffID, reservationTypeID)
	}
	return true, nil
}
