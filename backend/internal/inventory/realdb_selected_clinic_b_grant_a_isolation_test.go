package inventory

// realdb_selected_clinic_b_grant_a_isolation_test.go — D3 inventory package (4 routes)
//
// Proves clinic-fixed inventory + merchandise list/detail isolation through the
// real repository + service + HTTP handler path with a deterministic
// ClinicPermissionChecker. Offline `go test -short` SKIPs via testdb.SetupTestDB.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const (
	realDBInventorySeedNameA   = "D3inv-realdb-inventory-A"
	realDBInventorySeedNameB   = "D3inv-realdb-inventory-B"
	realDBMerchandiseSeedNameA = "D3inv-realdb-merchandise-A"
	realDBMerchandiseSeedNameB = "D3inv-realdb-merchandise-B"
)

func setupRealDBSelectedClinicBGrantAIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.InventoryItem{},
		&model.MerchandiseItem{},
	))
	testdb.Truncate(t, db,
		"inventory_items",
		"merchandise_items",
		"staff_clinic_assignments",
		"staffs",
	)
	return db
}

type inventoryQueryGuardService struct {
	InventoryService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *inventoryQueryGuardService) List(
	ctx context.Context,
	clinicID uint64,
	category, status *string,
	page, limit int,
) ([]model.InventoryItem, int64, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.InventoryService.List(ctx, clinicID, category, status, page, limit)
}

func (s *inventoryQueryGuardService) GetByID(
	ctx context.Context,
	clinicID, id uint64,
) (*model.InventoryItem, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf(
			"GetByID must not query clinic %d (id=%d) without selected-clinic grant",
			clinicID,
			id,
		)
	}
	return s.InventoryService.GetByID(ctx, clinicID, id)
}

type merchandiseQueryGuardService struct {
	MerchandiseItemService
	t                 *testing.T
	forbiddenClinicID uint64
}

func (s *merchandiseQueryGuardService) List(
	ctx context.Context,
	clinicID uint64,
	category string,
) ([]model.MerchandiseItem, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf("merchandise List must not query clinic %d without selected-clinic grant", clinicID)
	}
	return s.MerchandiseItemService.List(ctx, clinicID, category)
}

func (s *merchandiseQueryGuardService) GetByID(
	ctx context.Context,
	clinicID, id uint64,
) (*model.MerchandiseItem, error) {
	if clinicID == s.forbiddenClinicID {
		s.t.Fatalf(
			"merchandise GetByID must not query clinic %d (id=%d) without selected-clinic grant",
			clinicID,
			id,
		)
	}
	return s.MerchandiseItemService.GetByID(ctx, clinicID, id)
}

type realDBInventoryFixture struct {
	fx           testdb.ClinicGrantFixture
	inventoryA   *model.InventoryItem
	inventoryB   *model.InventoryItem
	merchandiseA *model.MerchandiseItem
	merchandiseB *model.MerchandiseItem
	handler      *Handler
}

func newRealDBInventoryHandler(
	t *testing.T,
	db *gorm.DB,
	forbiddenClinicID uint64,
) *Handler {
	t.Helper()
	invSvc := InventoryService(NewInventoryService(New(db)))
	merchSvc := MerchandiseItemService(NewMerchandiseItemService(
		NewMerchandiseItemRepository(db),
		nil,
	))
	if forbiddenClinicID != 0 {
		invSvc = &inventoryQueryGuardService{
			InventoryService:  invSvc,
			t:                 t,
			forbiddenClinicID: forbiddenClinicID,
		}
		merchSvc = &merchandiseQueryGuardService{
			MerchandiseItemService: merchSvc,
			t:                      t,
			forbiddenClinicID:      forbiddenClinicID,
		}
	}
	return NewHandler(invSvc, merchSvc, nil)
}

func seedRealDBInventoryFixture(
	t *testing.T,
	db *gorm.DB,
	forbiddenClinicID uint64,
) realDBInventoryFixture {
	t.Helper()
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3inv realDB")
	ctx := context.Background()

	inventoryA := &model.InventoryItem{
		ClinicID: fx.ClinicA,
		Name:     realDBInventorySeedNameA,
		Category: model.InventoryCategoryConsumable,
		Quantity: 5,
		Unit:     "本",
		Status:   model.InventoryStatusSufficient,
	}
	require.NoError(t, db.WithContext(ctx).Create(inventoryA).Error)
	inventoryB := &model.InventoryItem{
		ClinicID: fx.ClinicB,
		Name:     realDBInventorySeedNameB,
		Category: model.InventoryCategoryConsumable,
		Quantity: 8,
		Unit:     "本",
		Status:   model.InventoryStatusSufficient,
	}
	require.NoError(t, db.WithContext(ctx).Create(inventoryB).Error)

	merchandiseA := &model.MerchandiseItem{
		ClinicID:  fx.ClinicA,
		Name:      realDBMerchandiseSeedNameA,
		Category:  model.ItemCategoryGoods,
		UnitPrice: 1000,
		TaxType:   model.TaxTypeExcluded,
		TaxRate:   0.10,
		IsActive:  true,
	}
	require.NoError(t, db.WithContext(ctx).Create(merchandiseA).Error)
	merchandiseB := &model.MerchandiseItem{
		ClinicID:  fx.ClinicB,
		Name:      realDBMerchandiseSeedNameB,
		Category:  model.ItemCategoryGoods,
		UnitPrice: 2000,
		TaxType:   model.TaxTypeExcluded,
		TaxRate:   0.10,
		IsActive:  true,
	}
	require.NoError(t, db.WithContext(ctx).Create(merchandiseB).Error)

	return realDBInventoryFixture{
		fx:           fx,
		inventoryA:   inventoryA,
		inventoryB:   inventoryB,
		merchandiseA: merchandiseA,
		merchandiseB: merchandiseB,
		handler:      newRealDBInventoryHandler(t, db, forbiddenClinicID),
	}
}

func TestRealDB_Inventory_List_MembershipABGrantASelectedB_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBInventoryFixture(t, db, 0)
	fx.handler = newRealDBInventoryHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/inventory?page=1&limit=50",
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceInventory), fx.fx.ClinicA),
	)
	fx.handler.ListInventory(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBInventorySeedNameA,
		realDBInventorySeedNameB,
	)
}

func TestRealDB_Inventory_Get_MembershipABGrantASelectedB_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBInventoryFixture(t, db, 0)
	fx.handler = newRealDBInventoryHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/inventory/%d", fx.inventoryB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceInventory), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.inventoryB.ID)}}
		},
	)
	fx.handler.GetInventory(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBInventorySeedNameA,
		realDBInventorySeedNameB,
	)
}

func TestRealDB_Inventory_List_MembershipABGrantBSelectedB_ReturnsBOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBInventoryFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/inventory?page=1&limit=50",
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceInventory), fx.fx.ClinicB),
	)
	fx.handler.ListInventory(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed httpapi.PaginatedResponse[[]inventoryResponse]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed.Data, "B seed list must be nonempty; empty list is not observation")
	foundB := false
	for _, item := range listed.Data {
		assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
		assert.NotEqual(t, fx.inventoryA.ID, item.ID)
		assert.NotEqual(t, realDBInventorySeedNameA, item.Name)
		assert.NotEqual(t, fx.fx.ClinicA, item.ClinicID)
		if item.ID == fx.inventoryB.ID {
			foundB = true
			assert.Equal(t, realDBInventorySeedNameB, item.Name)
		}
	}
	require.True(t, foundB, "B seed id/name/clinic_id must appear in list")
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBInventorySeedNameA)
}

func TestRealDB_Inventory_Get_SelectedB_ASeedID_Returns404NoABody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBInventoryFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/inventory/%d", fx.inventoryA.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceInventory), fx.fx.ClinicB)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.inventoryA.ID)}}
		},
	)
	fx.handler.GetInventory(c)

	require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBInventorySeedNameA,
		realDBInventorySeedNameB,
	)
}

func TestRealDB_Merchandise_List_MembershipABGrantASelectedB_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBInventoryFixture(t, db, 0)
	fx.handler = newRealDBInventoryHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/masters/merchandise-items",
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceMasterMerchandise), fx.fx.ClinicA),
	)
	fx.handler.ListMerchandiseItems(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBMerchandiseSeedNameA,
		realDBMerchandiseSeedNameB,
	)
}

func TestRealDB_Merchandise_Get_MembershipABGrantASelectedB_Returns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBInventoryFixture(t, db, 0)
	fx.handler = newRealDBInventoryHandler(t, db, fx.fx.ClinicB)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/masters/merchandise-items/%d", fx.merchandiseB.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceMasterMerchandise), fx.fx.ClinicA)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.merchandiseB.ID)}}
		},
	)
	fx.handler.GetMerchandiseItem(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBMerchandiseSeedNameA,
		realDBMerchandiseSeedNameB,
	)
}

func TestRealDB_Merchandise_List_MembershipABGrantBSelectedB_ReturnsBOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBInventoryFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		"/api/v1/masters/merchandise-items",
		testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceMasterMerchandise), fx.fx.ClinicB),
	)
	fx.handler.ListMerchandiseItems(c)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var listed []merchandiseItemResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listed))
	require.NotEmpty(t, listed, "B seed list must be nonempty; empty list is not observation")
	foundB := false
	for _, item := range listed {
		assert.Equal(t, fx.fx.ClinicB, item.ClinicID)
		assert.NotEqual(t, fx.merchandiseA.ID, item.ID)
		assert.NotEqual(t, realDBMerchandiseSeedNameA, item.Name)
		if item.ID == fx.merchandiseB.ID {
			foundB = true
			assert.Equal(t, realDBMerchandiseSeedNameB, item.Name)
		}
	}
	require.True(t, foundB, "B seed id/name/clinic_id must appear in merchandise list")
	testdb.AssertBodyOmitsClinicArtifacts(t, w.Body.Bytes(), []uint64{fx.fx.ClinicA}, realDBMerchandiseSeedNameA)
}

func TestRealDB_Merchandise_Get_SelectedB_ASeedID_Returns404NoABody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupRealDBSelectedClinicBGrantAIsolationTestDB(t)
	fx := seedRealDBInventoryFixture(t, db, 0)

	c, w := testdb.NewHTTPTestContext(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/masters/merchandise-items/%d", fx.merchandiseA.ID),
		func(c *gin.Context) {
			testdb.ConfigureSelectedClinicBGrant(fx.fx, string(model.ResourceMasterMerchandise), fx.fx.ClinicB)(c)
			c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", fx.merchandiseA.ID)}}
		},
	)
	fx.handler.GetMerchandiseItem(c)

	require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	testdb.AssertBodyOmitsClinicArtifacts(
		t,
		w.Body.Bytes(),
		[]uint64{fx.fx.ClinicA, fx.fx.ClinicB},
		realDBMerchandiseSeedNameA,
		realDBMerchandiseSeedNameB,
	)
}
