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
	findAllFn           func(ctx context.Context, clinicID uint64) ([]model.TreatmentPackage, error)
	findByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.TreatmentPackage, error)
	liveItemsFn         func(ctx context.Context, clinicID uint64, ids []uint64) ([]model.TreatmentPackageItem, error)
	allItemsFn          func(ctx context.Context, clinicID, packageID uint64) ([]model.TreatmentPackageItem, error)
	lockClinicFn        func(ctx context.Context, clinicID uint64) error
	lockCatalogFn       func(ctx context.Context, clinicID uint64) (int64, error)
	catalogRevFn        func(ctx context.Context, clinicID uint64) (int64, error)
	incCatalogFn        func(ctx context.Context, clinicID uint64, expected int64) error
	lockPkgsFn          func(ctx context.Context, clinicID uint64, ids []uint64) ([]model.TreatmentPackage, error)
	lockAllForReorderFn func(ctx context.Context, clinicID uint64) ([]model.TreatmentPackage, error)
	maxSortFn           func(ctx context.Context, clinicID uint64) (int, error)
	createFn            func(ctx context.Context, pkg *model.TreatmentPackage, items []model.TreatmentPackageItem) error
	updateFn            func(ctx context.Context, clinicID, id uint64, expected int64, fields map[string]any) (*model.TreatmentPackage, error)
	replaceItemsFn      func(ctx context.Context, clinicID, packageID uint64, keep []model.TreatmentPackageItem, removeIDs []uint64, now time.Time) error
	softDeleteFn        func(ctx context.Context, clinicID, id uint64, expected int64, actorID uint64, now time.Time) error
	reorderFn           func(ctx context.Context, clinicID, id uint64, sortOrder int, actorID uint64) error
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
func (m *mockTreatmentPackageRepo) LockAllPackagesForReorder(ctx context.Context, c uint64) ([]model.TreatmentPackage, error) {
	return m.lockAllForReorderFn(ctx, c)
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
func (m *mockTreatmentPackageRepo) UpdateSortOrderBumpVersion(ctx context.Context, c, id uint64, s int, a uint64) error {
	return m.reorderFn(ctx, c, id, s, a)
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
		liveItemsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackageItem, error) {
			return nil, nil
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
		lockAllForReorderFn: func(context.Context, uint64) ([]model.TreatmentPackage, error) {
			return []model.TreatmentPackage{{ID: 201, SortOrder: 0}, {ID: 202, SortOrder: 1}}, nil
		},
		liveItemsFn:  func(context.Context, uint64, []uint64) ([]model.TreatmentPackageItem, error) { return nil, nil },
		incCatalogFn: func(context.Context, uint64, int64) error { catalogBumped = true; return nil },
		reorderFn:    func(context.Context, uint64, uint64, int, uint64) error { moved = true; return nil },
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
		lockAllForReorderFn: func(context.Context, uint64) ([]model.TreatmentPackage, error) {
			return []model.TreatmentPackage{{ID: 201, SortOrder: 0}, {ID: 202, SortOrder: 1, IsActive: false}}, nil
		},
		liveItemsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackageItem, error) { return nil, nil },
	}
	svc, _ := newPackageService(repo, nil)
	assert.ErrorIs(t, svc.Reorder(context.Background(), 7, 7, []uint64{201}, 9), apperrors.ErrInvalidInput)      // incomplete set
	assert.ErrorIs(t, svc.Reorder(context.Background(), 7, 7, []uint64{201, 999}, 9), apperrors.ErrInvalidInput) // foreign id; inactive 202 still required
}

// S06: 一覧表示の FindAll cap（MaxMasterListRows=10000）を超える医院でも、
// reorder は cap 無しの全件ロックで照合する。10,001件（inactive 混在・非連番
// sort_order）の完全 permutation を受理し、catalog revision は1回だけ上がる。
func reorderFixturePackages(total int) []model.TreatmentPackage {
	pkgs := make([]model.TreatmentPackage, 0, total)
	for i := 0; i < total; i++ {
		pkgs = append(pkgs, model.TreatmentPackage{
			ID:        uint64(i + 1),
			ClinicID:  7,
			IsActive:  i%2 == 0,
			SortOrder: (i*7)%total + 1, // 非連番の一意 sort 値。
		})
	}
	return pkgs
}

func reorderAllIDs(total int) []uint64 {
	ids := make([]uint64, 0, total)
	for i := 1; i <= total; i++ {
		ids = append(ids, uint64(i))
	}
	return ids
}

func TestTreatmentPackageService_Reorder_FullSetBeyondListCapAccepted(t *testing.T) {
	const total = 10001
	pkgs := reorderFixturePackages(total)
	// 全件の完全 permutation（反転）。部分集合ではなく真の全件列。
	ids := reorderAllIDs(total)
	for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}
	incCalls := 0
	repo := &mockTreatmentPackageRepo{
		lockClinicFn:  func(context.Context, uint64) error { return nil },
		lockCatalogFn: func(context.Context, uint64) (int64, error) { return 9, nil },
		lockAllForReorderFn: func(context.Context, uint64) ([]model.TreatmentPackage, error) {
			return pkgs, nil
		},
		liveItemsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackageItem, error) { return nil, nil },
		reorderFn:   func(context.Context, uint64, uint64, int, uint64) error { return nil },
		incCatalogFn: func(_ context.Context, _ uint64, expected int64) error {
			incCalls++
			assert.Equal(t, int64(9), expected)
			return nil
		},
	}
	svc, audit := newPackageService(repo, nil)
	require.NoError(t, svc.Reorder(context.Background(), 7, 9, ids, 5))
	assert.Equal(t, 1, incCalls) // 全件受理で revision は1回だけ。
	assert.Len(t, audit.entries, 1)
}

// S06: cap 境界の部分集合・重複・foreign id・stale order_version はすべて
// zero-write で拒否（UpdateSortOrderBumpVersion / IncrementCatalogRevision / audit なし）。
func TestTreatmentPackageService_Reorder_RejectedBeyondListCapZeroWrite(t *testing.T) {
	const total = 10001
	mkRepo := func(reorderCalled, incCalled *bool) *mockTreatmentPackageRepo {
		return &mockTreatmentPackageRepo{
			lockClinicFn:  func(context.Context, uint64) error { return nil },
			lockCatalogFn: func(context.Context, uint64) (int64, error) { return 9, nil },
			lockAllForReorderFn: func(context.Context, uint64) ([]model.TreatmentPackage, error) {
				return reorderFixturePackages(total), nil
			},
			liveItemsFn:  func(context.Context, uint64, []uint64) ([]model.TreatmentPackageItem, error) { return nil, nil },
			reorderFn:    func(context.Context, uint64, uint64, int, uint64) error { *reorderCalled = true; return nil },
			incCatalogFn: func(context.Context, uint64, int64) error { *incCalled = true; return nil },
		}
	}
	run := func(t *testing.T, orderVersion int64, ids []uint64) (error, *mockPackageAudit, bool, bool) {
		reorderCalled := false
		incCalled := false
		repo := mkRepo(&reorderCalled, &incCalled)
		svc, audit := newPackageService(repo, nil)
		err := svc.Reorder(context.Background(), 7, orderVersion, ids, 5)
		return err, audit, reorderCalled, incCalled
	}
	assertZeroWrite := func(t *testing.T, audit *mockPackageAudit, reorderCalled, incCalled bool) {
		assert.False(t, reorderCalled, "UpdateSortOrderBumpVersion must not run")
		assert.False(t, incCalled, "IncrementCatalogRevision must not run")
		assert.Empty(t, audit.entries, "audit must not run")
	}

	t.Run("10,000-id subset", func(t *testing.T) {
		err, audit, reorderCalled, incCalled := run(t, 9, reorderAllIDs(total)[:total-1])
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
		assertZeroWrite(t, audit, reorderCalled, incCalled)
	})
	t.Run("duplicated id", func(t *testing.T) {
		ids := reorderAllIDs(total)
		ids[total-1] = ids[0] // 末尾を先頭と同じ id にして重複させる（1件欠落）。
		err, audit, reorderCalled, incCalled := run(t, 9, ids)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
		assertZeroWrite(t, audit, reorderCalled, incCalled)
	})
	t.Run("foreign id", func(t *testing.T) {
		ids := reorderAllIDs(total)
		ids[total-1] = uint64(total) + 1 // 他院・存在しない id。
		err, audit, reorderCalled, incCalled := run(t, 9, ids)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
		assertZeroWrite(t, audit, reorderCalled, incCalled)
	})
	t.Run("stale order_version", func(t *testing.T) {
		err, audit, reorderCalled, incCalled := run(t, 8, reorderAllIDs(total))
		assert.ErrorIs(t, err, apperrors.ErrConflict)
		assertZeroWrite(t, audit, reorderCalled, incCalled)
	})
}

// F1: components 未指定の update でも lock 済み package の live items を書込み前に検査する。
// 権限不足の型を含む場合は 403 で全体非開示（zero-write）。
func TestTreatmentPackageService_Update_MetadataOnlyDeniedLiveKindForbidden(t *testing.T) {
	updateCalled := false
	repo := &mockTreatmentPackageRepo{
		lockClinicFn: func(context.Context, uint64) error { return nil },
		lockPkgsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackage, error) {
			return []model.TreatmentPackage{{ID: 201, ClinicID: 7, Version: 3, Name: "old"}}, nil
		},
		liveItemsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackageItem, error) {
			return []model.TreatmentPackageItem{{PackageID: 201, ItemKind: "checkup_type"}}, nil
		},
		updateFn: func(context.Context, uint64, uint64, int64, map[string]any) (*model.TreatmentPackage, error) {
			updateCalled = true
			return nil, nil
		},
	}
	svc, _ := newPackageService(repo, map[model.Resource]bool{model.ResourceCheckups: true})
	actor := uint64(9)
	name := "new"
	_, err := svc.Update(context.Background(), 7, 201, &UpdateTreatmentPackageInput{
		Version: 3, ActorID: &actor, Name: &name,
	})
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
	assert.False(t, updateCalled) // zero-write
}

// F1: delete も lock 済み package の live items を書込み前に検査し、権限不足の型を
// 含む場合は 403 zero-write。
func TestTreatmentPackageService_Delete_DeniedLiveKindForbiddenZeroWrite(t *testing.T) {
	deleteCalled := false
	repo := &mockTreatmentPackageRepo{
		lockClinicFn:  func(context.Context, uint64) error { return nil },
		lockCatalogFn: func(context.Context, uint64) (int64, error) { return 3, nil },
		lockPkgsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackage, error) {
			return []model.TreatmentPackage{{ID: 201, ClinicID: 7, Version: 3}}, nil
		},
		liveItemsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackageItem, error) {
			return []model.TreatmentPackageItem{{PackageID: 201, ItemKind: "checkup_type"}}, nil
		},
		softDeleteFn: func(context.Context, uint64, uint64, int64, uint64, time.Time) error {
			deleteCalled = true
			return nil
		},
	}
	svc, _ := newPackageService(repo, map[model.Resource]bool{model.ResourceCheckups: true})
	err := svc.Delete(context.Background(), 7, 201, 3, 9)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
	assert.False(t, deleteCalled) // zero-write
}

// F1: update の応答 items にも同じ検査を適用する。lock 後の live items は権限内だが、
// 書込後に読み直した応答に権限不足の型が混ざった場合も 403 で全体非開示にする。
func TestTreatmentPackageService_Update_ResponseDeniedKindForbidden(t *testing.T) {
	liveCalls := 0
	repo := &mockTreatmentPackageRepo{
		lockClinicFn: func(context.Context, uint64) error { return nil },
		lockPkgsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackage, error) {
			return []model.TreatmentPackage{{ID: 201, ClinicID: 7, Version: 3, Name: "old"}}, nil
		},
		liveItemsFn: func(context.Context, uint64, []uint64) ([]model.TreatmentPackageItem, error) {
			liveCalls++
			if liveCalls == 1 {
				return []model.TreatmentPackageItem{{PackageID: 201, ItemKind: "consultation"}}, nil
			}
			return []model.TreatmentPackageItem{{PackageID: 201, ItemKind: "checkup_type"}}, nil
		},
		updateFn: func(_ context.Context, _, id uint64, _ int64, fields map[string]any) (*model.TreatmentPackage, error) {
			return &model.TreatmentPackage{ID: id, ClinicID: 7, Version: 4, Name: fields["name"].(string)}, nil
		},
	}
	svc, _ := newPackageService(repo, map[model.Resource]bool{model.ResourceCheckups: true})
	actor := uint64(9)
	name := "new"
	_, err := svc.Update(context.Background(), 7, 201, &UpdateTreatmentPackageInput{
		Version: 3, ActorID: &actor, Name: &name,
	})
	assert.ErrorIs(t, err, apperrors.ErrForbidden) // 応答 items の権限不足も 403（tx rollback）。
}
