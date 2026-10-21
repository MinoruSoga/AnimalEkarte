package medicalrecord

// EMR-521 service tests (NOT_RUN — fn-field mocks, execution deferred).

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockTreatmentPackageRepo struct {
	findAllFn      func(ctx context.Context, clinicID uint64) ([]model.TreatmentPackage, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.TreatmentPackage, error)
	liveItemsFn    func(ctx context.Context, clinicID uint64, ids []uint64) ([]model.TreatmentPackageItem, error)
	allItemsFn     func(ctx context.Context, clinicID, packageID uint64) ([]model.TreatmentPackageItem, error)
	lockClinicFn   func(ctx context.Context, clinicID uint64) error
	lockCatalogFn  func(ctx context.Context, clinicID uint64) (int64, error)
	catalogRevFn   func(ctx context.Context, clinicID uint64) (int64, error)
	incCatalogFn   func(ctx context.Context, clinicID uint64, expected int64) error
	lockPkgsFn     func(ctx context.Context, clinicID uint64, ids []uint64) ([]model.TreatmentPackage, error)
	maxSortFn      func(ctx context.Context, clinicID uint64) (int, error)
	createFn       func(ctx context.Context, pkg *model.TreatmentPackage, items []model.TreatmentPackageItem) error
	updateFn       func(ctx context.Context, clinicID, id uint64, expected int64, fields map[string]any) (*model.TreatmentPackage, error)
	replaceItemsFn func(ctx context.Context, clinicID, packageID uint64, keep []model.TreatmentPackageItem, removeIDs []uint64, now time.Time) error
	softDeleteFn   func(ctx context.Context, clinicID, id uint64, expected int64, actorID uint64, now time.Time) error
	reorderFn      func(ctx context.Context, clinicID, id uint64, sortOrder int) error
}

func (m *mockTreatmentPackageRepo) FindAll(ctx context.Context, c uint64) ([]model.TreatmentPackage, error) {
	return m.findAllFn(ctx, c)
}
func (m *mockTreatmentPackageRepo) FindByID(ctx context.Context, c, id uint64) (*model.TreatmentPackage, error) {
	return m.findByIDFn(ctx, c, id)
}
func (m *mockTreatmentPackageRepo) FindLiveItemsByPackageIDs(ctx context.Context, c uint64, ids []uint64) ([]model.TreatmentPackageItem, error) {
	return m.liveItemsFn(ctx, c, ids)
}
func (m *mockTreatmentPackageRepo) FindAllItemsForPackage(ctx context.Context, c, p uint64) ([]model.TreatmentPackageItem, error) {
	return m.allItemsFn(ctx, c, p)
}
func (m *mockTreatmentPackageRepo) LockClinicRow(ctx context.Context, c uint64) error {
	return m.lockClinicFn(ctx, c)
}
func (m *mockTreatmentPackageRepo) LockCatalogRow(ctx context.Context, c uint64) (int64, error) {
	return m.lockCatalogFn(ctx, c)
}
func (m *mockTreatmentPackageRepo) GetCatalogRevision(ctx context.Context, c uint64) (int64, error) {
	return m.catalogRevFn(ctx, c)
}
func (m *mockTreatmentPackageRepo) IncrementCatalogRevision(ctx context.Context, c uint64, e int64) error {
	return m.incCatalogFn(ctx, c, e)
}
func (m *mockTreatmentPackageRepo) LockPackagesByIDAsc(ctx context.Context, c uint64, ids []uint64) ([]model.TreatmentPackage, error) {
	return m.lockPkgsFn(ctx, c, ids)
}
func (m *mockTreatmentPackageRepo) MaxSortOrder(ctx context.Context, c uint64) (int, error) {
	return m.maxSortFn(ctx, c)
}
func (m *mockTreatmentPackageRepo) Create(ctx context.Context, p *model.TreatmentPackage, items []model.TreatmentPackageItem) error {
	return m.createFn(ctx, p, items)
}
func (m *mockTreatmentPackageRepo) UpdateFieldsCAS(ctx context.Context, c, id uint64, e int64, f map[string]any) (*model.TreatmentPackage, error) {
	return m.updateFn(ctx, c, id, e, f)
}
func (m *mockTreatmentPackageRepo) ReplaceItems(ctx context.Context, c, p uint64, k []model.TreatmentPackageItem, rm []uint64, now time.Time) error {
	return m.replaceItemsFn(ctx, c, p, k, rm, now)
}
func (m *mockTreatmentPackageRepo) SoftDeleteCAS(ctx context.Context, c, id uint64, e int64, a uint64, now time.Time) error {
	return m.softDeleteFn(ctx, c, id, e, a, now)
}
func (m *mockTreatmentPackageRepo) UpdateSortOrderBumpVersion(ctx context.Context, c, id uint64, s int) error {
	return m.reorderFn(ctx, c, id, s)
}

type mockPackageTx struct{}

func (mockPackageTx) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type mockPackageAudit struct{ entries []*AuditEntry }

func (m *mockPackageAudit) LogEntryTx(ctx context.Context, e *AuditEntry) error {
	m.entries = append(m.entries, e)
	return nil
}

type mockPackageAuthorizer struct{ denied map[model.Resource]bool }

func (m mockPackageAuthorizer) Granted(_ context.Context, _ uint64, r model.Resource, _ string) bool {
	return !m.denied[r]
}

type fnPackageFinder[T any] struct {
	find func(ctx context.Context, clinicID, id uint64) (*T, error)
}

func (f fnPackageFinder[T]) FindByID(ctx context.Context, c, id uint64) (*T, error) {
	return f.find(ctx, c, id)
}

func activeFinder[T any](mk func() *T) func(ctx context.Context, c, id uint64) (*T, error) {
	return func(context.Context, uint64, uint64) (*T, error) { return mk(), nil }
}

func newPackageService(repo *mockTreatmentPackageRepo, denied map[model.Resource]bool) (*treatmentPackageService, *mockPackageAudit) {
	audit := &mockPackageAudit{}
	deps := TreatmentPackageDependencies{
		Transactor: mockPackageTx{}, Audit: audit,
		Authorizer:       mockPackageAuthorizer{denied: denied},
		Consultations:    fnPackageFinder[model.Consultation]{find: activeFinder(func() *model.Consultation { return &model.Consultation{IsActive: true} })},
		Procedures:       fnPackageFinder[model.Procedure]{find: activeFinder(func() *model.Procedure { return &model.Procedure{IsActive: true} })},
		Medicines:        fnPackageFinder[model.Medicine]{find: activeFinder(func() *model.Medicine { return &model.Medicine{IsActive: true} })},
		ExamTypes:        fnPackageFinder[model.ExaminationType]{find: activeFinder(func() *model.ExaminationType { return &model.ExaminationType{IsActive: true} })},
		Vaccines:         fnPackageFinder[model.Vaccine]{find: activeFinder(func() *model.Vaccine { return &model.Vaccine{IsActive: true} })},
		CheckupTypes:     fnPackageFinder[model.CheckupType]{find: activeFinder(func() *model.CheckupType { return &model.CheckupType{IsActive: true} })},
		MerchandiseItems: fnPackageFinder[model.MerchandiseItem]{find: activeFinder(func() *model.MerchandiseItem { return &model.MerchandiseItem{IsActive: true} })},
	}
	return &treatmentPackageService{repo: repo, deps: deps}, audit
}

func pkgComponent(kind string, id uint64) TreatmentPackageComponentInput {
	var q model.ExactNumeric
	_ = q.Scan("1")
	return TreatmentPackageComponentInput{ComponentKey: uuid.New(), Kind: kind, TypedID: id, Quantity: q, QuantityText: "1"}
}

func TestTreatmentPackageService_List_MissingCatalogIsInvariantError(t *testing.T) {
	repo := &mockTreatmentPackageRepo{
		catalogRevFn: func(context.Context, uint64) (int64, error) { return 0, errPackageCatalogNotInitialized },
	}
	svc, _ := newPackageService(repo, nil)
	_, _, err := svc.List(context.Background(), 7)
	assert.Error(t, err) // GET never INSERTs or returns a fake version 1.
}

func TestTreatmentPackageService_List_UnauthorizedKindIsForbidden(t *testing.T) {
	repo := &mockTreatmentPackageRepo{
		catalogRevFn: func(context.Context, uint64) (int64, error) { return 3, nil },
		findAllFn:    func(context.Context, uint64) ([]model.TreatmentPackage, error) {
			return []model.TreatmentPackage{{ID: 201, ClinicID: 7}}, nil
		},
		liveItemsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackageItem, error) {
			return []model.TreatmentPackageItem{{PackageID: 201, ItemKind: "checkup_type"}}, nil
		},
	}
	svc, _ := newPackageService(repo, map[model.Resource]bool{model.ResourceCheckups: true})
	_, _, err := svc.List(context.Background(), 7)
	assert.ErrorIs(t, err, apperrors.ErrForbidden) // 権限不足の構成を含む read は全体非開示。
}

func TestTreatmentPackageService_Create_AppendsAndBumpsCatalog(t *testing.T) {
	var createdPkg *model.TreatmentPackage
	var bumped int64 = -1
	repo := &mockTreatmentPackageRepo{
		lockClinicFn:  func(context.Context, uint64) error { return nil },
		lockCatalogFn: func(context.Context, uint64) (int64, error) { return 7, nil },
		maxSortFn:     func(context.Context, uint64) (int, error) { return 4, nil },
		createFn: func(_ context.Context, p *model.TreatmentPackage, _ []model.TreatmentPackageItem) error {
			p.ID = 201
			createdPkg = p
			return nil
		},
		incCatalogFn: func(_ context.Context, _ uint64, expected int64) error { bumped = expected; return nil },
	}
	svc, audit := newPackageService(repo, nil)
	actor := uint64(9)
	view, err := svc.Create(context.Background(), 7, &CreateTreatmentPackageInput{
		ActorID: &actor, Name: "pkg", IsActive: true,
		Components: []TreatmentPackageComponentInput{pkgComponent("consultation", 101)},
	})
	require.NoError(t, err)
	assert.Equal(t, 5, createdPkg.SortOrder) // 末尾追加。
	assert.Equal(t, int64(7), bumped)
	assert.Len(t, view.Items, 1)
	assert.Len(t, audit.entries, 1)
}

func TestTreatmentPackageService_Create_InactiveSourceRejectedZeroWrite(t *testing.T) {
	createCalled := false
	repo := &mockTreatmentPackageRepo{
		lockClinicFn:  func(context.Context, uint64) error { return nil },
		lockCatalogFn: func(context.Context, uint64) (int64, error) { return 1, nil },
		createFn:      func(context.Context, *model.TreatmentPackage, []model.TreatmentPackageItem) error {
			createCalled = true
			return nil
		},
	}
	svc, _ := newPackageService(repo, nil)
	svc.deps.Consultations = fnPackageFinder[model.Consultation]{
		find: activeFinder(func() *model.Consultation { return &model.Consultation{IsActive: false} }),
	}
	actor := uint64(9)
	_, err := svc.Create(context.Background(), 7, &CreateTreatmentPackageInput{
		ActorID: &actor, Name: "pkg", IsActive: true,
		Components: []TreatmentPackageComponentInput{pkgComponent("consultation", 101)},
	})
	assert.ErrorIs(t, err, apperrors.ErrUnprocessable)
	assert.False(t, createCalled) // zero-write
}

func TestTreatmentPackageService_Delete_CatalogLockedBeforePackageAndCAS(t *testing.T) {
	var order []string
	repo := &mockTreatmentPackageRepo{
		lockClinicFn:  func(context.Context, uint64) error { order = append(order, "clinic"); return nil },
		lockCatalogFn: func(context.Context, uint64) (int64, error) { order = append(order, "catalog"); return 3, nil },
		lockPkgsFn:    func(context.Context, uint64, []uint64) ([]model.TreatmentPackage, error) {
			order = append(order, "packages")
			return []model.TreatmentPackage{{ID: 201, ClinicID: 7, Version: 3}}, nil
		},
		softDeleteFn: func(context.Context, uint64, uint64, int64, uint64, time.Time) error {
			return apperrors.WrapConflict("stale")
		},
	}
	svc, _ := newPackageService(repo, nil)
	err := svc.Delete(context.Background(), 7, 201, 2, 9)
	assert.ErrorIs(t, err, apperrors.ErrConflict) // CAS conflict -> zero-write.
	assert.Equal(t, []string{"clinic", "catalog", "packages"}, order)
}

func TestTreatmentPackageService_Reorder_StaleRevisionZeroWrite(t *testing.T) {
	repo := &mockTreatmentPackageRepo{
		lockClinicFn:  func(context.Context, uint64) error { return nil },
		lockCatalogFn: func(context.Context, uint64) (int64, error) { return 8, nil },
	}
	svc, _ := newPackageService(repo, nil)
	err := svc.Reorder(context.Background(), 7, 7, []uint64{201}, 9)
	assert.ErrorIs(t, err, apperrors.ErrConflict)
}

func TestTreatmentPackageService_Reorder_NoOpKeepsVersions(t *testing.T) {
	catalogBumped := false
	moved := false
	repo := &mockTreatmentPackageRepo{
		lockClinicFn:  func(context.Context, uint64) error { return nil },
		lockCatalogFn: func(context.Context, uint64) (int64, error) { return 7, nil },
		findAllFn: func(context.Context, uint64) ([]model.TreatmentPackage, error) {
			return []model.TreatmentPackage{{ID: 201, SortOrder: 0}, {ID: 202, SortOrder: 1}}, nil
		},
		lockPkgsFn: func(_ context.Context, _ uint64, ids []uint64) ([]model.TreatmentPackage, error) {
			return []model.TreatmentPackage{{ID: 201, SortOrder: 0}, {ID: 202, SortOrder: 1}}, nil
		},
		incCatalogFn: func(context.Context, uint64, int64) error { catalogBumped = true; return nil },
		reorderFn:    func(context.Context, uint64, uint64, int) error { moved = true; return nil },
	}
	svc, _ := newPackageService(repo, nil)
	require.NoError(t, svc.Reorder(context.Background(), 7, 7, []uint64{201, 202}, 9))
	assert.False(t, moved)
	assert.False(t, catalogBumped) // no-op は版不変。
}

func TestTreatmentPackageService_Reorder_MembershipMismatchRejected(t *testing.T) {
	repo := &mockTreatmentPackageRepo{
		lockClinicFn:  func(context.Context, uint64) error { return nil },
		lockCatalogFn: func(context.Context, uint64) (int64, error) { return 7, nil },
		findAllFn: func(context.Context, uint64) ([]model.TreatmentPackage, error) {
			return []model.TreatmentPackage{{ID: 201, SortOrder: 0}, {ID: 202, SortOrder: 1, IsActive: false}}, nil
		},
		lockPkgsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackage, error) {
			return []model.TreatmentPackage{{ID: 201, SortOrder: 0}, {ID: 202, SortOrder: 1}}, nil
		},
	}
	svc, _ := newPackageService(repo, nil)
	assert.ErrorIs(t, svc.Reorder(context.Background(), 7, 7, []uint64{201}, 9), apperrors.ErrInvalidInput)      // incomplete set
	assert.ErrorIs(t, svc.Reorder(context.Background(), 7, 7, []uint64{201, 999}, 9), apperrors.ErrInvalidInput) // foreign id; inactive 202 still required
}
