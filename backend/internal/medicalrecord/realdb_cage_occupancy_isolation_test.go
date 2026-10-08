package medicalrecord

// realdb_cage_occupancy_isolation_test.go — EMR-306
//
// D3 evidence for GET /api/v1/masters/cages/occupancy (class: clinic-fixed).
// OccupiedCageIDs は選択医院スコープの占有ケージのみ返すことを実DBで示す:
// 医院Aの占有ケージが医院B選択の応答へ混入しないこと、および handler が
// service 層へ選択医院以外の clinicID を渡さないことを guard で証明する。
// middleware 側の deny 経路（grant なし選択医院）は executed-middleware
// テスト（get_head_permission_execution_test.go）が担保する。

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// cageOccupancyQueryGuard は service 層が forbidden 医院IDで呼ばれた場合に
// テストを失敗させる OccupiedCageIDs ラッパー（hospQueryGuard と同型）。
type cageOccupancyQueryGuard struct {
	CageService
	g clinicIDGuard
}

func (s *cageOccupancyQueryGuard) OccupiedCageIDs(ctx context.Context, clinicID uint64) ([]uint64, error) {
	s.g.check(clinicID, "cage OccupiedCageIDs")
	reader, ok := s.CageService.(CageOccupancyReader)
	if !ok {
		return nil, apperrors.WrapInternalServerError("cage occupancy reader is not configured")
	}
	return reader.OccupiedCageIDs(ctx, clinicID)
}

func setupRealDBCageOccupancyDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.AnimalSpecies{}, &model.Owner{}, &model.Pet{},
		&model.Cage{}, &model.Hospitalization{},
	))
	testdb.Truncate(t, db,
		"hospitalizations", "cages", "pets", "owners",
		"staff_clinic_assignments", "staffs",
	)
	return db
}

func TestRealDB_CageOccupancySelectedClinicBIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	res := string(model.ResourceMasterHospitalization)

	db := setupRealDBCageOccupancyDB(t)
	fx := testdb.SeedDualClinicGrantFixture(t, db, "D3mr cage occupancy")

	ownerA := testdb.MakeTestOwner(t, db, fx.ClinicA, realDBOwnerNameA)
	ownerB := testdb.MakeTestOwner(t, db, fx.ClinicB, realDBOwnerNameB)
	petA := testdb.MakeSpeciesAndPet(t, db, fx.ClinicA, ownerA.ID, realDBPetNameA)
	petB := testdb.MakeSpeciesAndPet(t, db, fx.ClinicB, ownerB.ID, realDBPetNameB)

	cageA := makeCageMaster(t, db, fx.ClinicA, "占有ケージA")
	cageB := makeCageMaster(t, db, fx.ClinicB, "占有ケージB")
	// 空ケージ（占有なし）は応答に含まれないことも ElementsMatch で同時に検証する。
	makeCageMaster(t, db, fx.ClinicB, "空ケージB")

	makeHospitalizationFixture(t, db, fx.ClinicA, ownerA.ID, petA.ID, func(h *model.Hospitalization) { h.CageID = &cageA.ID })
	makeHospitalizationFixture(t, db, fx.ClinicB, ownerB.ID, petB.ID, func(h *model.Hospitalization) { h.CageID = &cageB.ID })

	newHandler := func(forbidden uint64) *CageHandler {
		var svc CageService = NewCageService(NewCageRepository(db), persistenceTx(db))
		if forbidden != 0 {
			svc = &cageOccupancyQueryGuard{CageService: svc, g: clinicIDGuard{t: t, forbiddenClinicID: forbidden}}
		}
		return NewCageHandler(svc)
	}

	readIDs := func(w *httptest.ResponseRecorder) []uint64 {
		t.Helper()
		var resp cageOccupancyResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		return resp.OccupiedCageIDs
	}

	t.Run("grantB_selectedB_returns_only_clinicB_cages", func(t *testing.T) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/cages/occupancy", configureGrant(fx, res, fx.ClinicB))
		newHandler(0).OccupiedCageIDs(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.ElementsMatch(t, []uint64{cageB.ID}, readIDs(w))
	})

	// 選択医院B・占有 grant はAのみ。handler 自体は grant を見ない（deny は
	// route middleware の責務）ため 200 になるが、service には選択医院Bだけが
	// 渡り、応答に医院Aの占有ケージは混入しない。
	t.Run("grantA_selectedB_service_scoped_to_B_no_leak", func(t *testing.T) {
		c, w := testdb.NewHTTPTestContext(t, http.MethodGet, "/api/v1/masters/cages/occupancy", configureGrant(fx, res, fx.ClinicA))
		newHandler(fx.ClinicA).OccupiedCageIDs(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.ElementsMatch(t, []uint64{cageB.ID}, readIDs(w))
	})
}
