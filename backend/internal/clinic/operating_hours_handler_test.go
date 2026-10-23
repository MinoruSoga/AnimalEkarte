package clinic

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

const operatingHoursFixtureDir = "../../docs/fixtures/emr445-a03"

// operatingHoursFixture reads one merged api fixture and returns its compact form, which is the exact byte
// sequence gin's c.JSON writes for the same document (key order follows the schema's declared order).
func operatingHoursFixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(operatingHoursFixtureDir, name))
	require.NoError(t, err)
	var compact bytes.Buffer
	require.NoError(t, json.Compact(&compact, raw))
	return compact.Bytes()
}

type fixtureStoredSettings struct {
	WeeklyHours           json.RawMessage `json:"weekly_hours"`
	ClosedDates           json.RawMessage `json:"closed_dates"`
	NationalHolidayClosed bool            `json:"national_holiday_closed"`
}

type fixtureOperatingHours struct {
	Settings fixtureStoredSettings `json:"settings"`
}

// storedRowFromFixture builds the stored row that must render as the given get-configured fixture.
func storedRowFromFixture(t *testing.T, name string) model.ClinicOperatingHours {
	t.Helper()
	var fx fixtureOperatingHours
	require.NoError(t, json.Unmarshal(operatingHoursFixture(t, name), &fx))
	return model.ClinicOperatingHours{
		ClinicID:              7,
		SchemaVersion:         model.ClinicOperatingHoursSchemaV1,
		Version:               3,
		WeeklyHours:           fx.Settings.WeeklyHours,
		ClosedDates:           fx.Settings.ClosedDates,
		NationalHolidayClosed: fx.Settings.NationalHolidayClosed,
	}
}

// requestIdentity is the authenticated context the middleware chain would have set.
type requestIdentity struct {
	selectedClinic string
	memberClinics  []uint64
	systemAdmin    bool
	grant          func(clinicID uint64, resource, action string) bool
}

func grantAll(uint64, string, string) bool { return true }

func memberOfClinic7() requestIdentity {
	return requestIdentity{selectedClinic: "7", memberClinics: []uint64{7}, grant: grantAll}
}

// permissionCall records one route middleware: kind is "strict" (selected clinic) or "assigned"
// (GET/HEAD may fall back to an assigned clinic's grant).
type permissionCall struct{ kind, resource, action string }

type operatingHoursHarness struct {
	router      *gin.Engine
	repo        *fakeOperatingHoursRepository
	registered  []permissionCall
	denyRoute   func(permissionCall) bool
	routeChecks []permissionCall
}

func newOperatingHoursHarness(t *testing.T, repo *fakeOperatingHoursRepository, id requestIdentity) *operatingHoursHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := &operatingHoursHarness{repo: repo}
	middlewareOf := func(kind string) PermissionMiddleware {
		return func(resource, action string) gin.HandlerFunc {
			call := permissionCall{kind, resource, action}
			h.registered = append(h.registered, call)
			return func(c *gin.Context) {
				h.routeChecks = append(h.routeChecks, call)
				if h.denyRoute != nil && h.denyRoute(call) {
					httpapi.RespondError(c, apperrors.WrapForbidden("forbidden"))
					c.Abort()
					return
				}
				c.Next()
			}
		}
	}
	handler := NewOperatingHoursHandler(NewOperatingHoursReadService(repo), middlewareOf("strict"), middlewareOf("assigned"))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Mimic the global SecurityHeaders value that the route-level middleware must override.
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Set("clinic_id", id.selectedClinic)
		c.Set("clinic_ids", id.memberClinics)
		c.Set("is_system_admin", id.systemAdmin)
		c.Set("user_id", "17")
		httpapi.SetClinicPermissionChecker(c, func(_ *gin.Context, clinicID uint64, resource, action string) bool {
			return id.grant(clinicID, resource, action)
		})
		c.Next()
	})
	handler.RegisterOperatingHoursRoutes(router.Group("/api/v1"))
	h.router = router
	return h
}

func (h *operatingHoursHarness) get(target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, http.NoBody))
	return w
}

func requireNoStore(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	assert.Equal(t, []string{"no-store"}, w.Header().Values("Cache-Control"),
		"responses produced by these routes (200/400/403/503, including route-level permission denials) declare exactly Cache-Control: no-store")
}

func requireErrorBody(t *testing.T, w *httptest.ResponseRecorder, wantMessage string) {
	t.Helper()
	var body struct {
		Error string `json:"error"`
	}
	dec := json.NewDecoder(w.Body)
	dec.DisallowUnknownFields()
	require.NoError(t, dec.Decode(&body))
	if wantMessage != "" {
		assert.Equal(t, wantMessage, body.Error)
	}
	assert.NotEmpty(t, body.Error)
}

func configuredRepo(t *testing.T) *fakeOperatingHoursRepository {
	t.Helper()
	return &fakeOperatingHoursRepository{found: true, row: storedRowFromFixture(t, "get-configured.response.json")}
}

func TestOperatingHoursRoutes_PermissionWiring(t *testing.T) {
	h := newOperatingHoursHarness(t, &fakeOperatingHoursRepository{}, memberOfClinic7())

	var routes []string
	for _, r := range h.router.Routes() {
		routes = append(routes, r.Method+" "+r.Path)
	}
	assert.Equal(t, []string{
		"GET /api/v1/clinics/:clinic_id/operating-hours",
		"GET /api/v1/clinic-holidays/common",
		"GET /api/v1/closing-settings/holidays/common",
	}, routes, "exactly the three read routes; no write/preview/lookup routes in this slice")

	assert.Equal(t, []permissionCall{
		{"assigned", string(model.ResourceHospitalSettings), "view"},
		{"strict", string(model.ResourceShifts), "view"},
		{"strict", string(model.ResourceClosingSettings), "view"},
	}, h.registered, "clinic-path route falls back to assigned-clinic grants; the two holiday routes stay strict on the selected clinic")
}

func TestGetClinicOperatingHours_Configured(t *testing.T) {
	h := newOperatingHoursHarness(t, configuredRepo(t), memberOfClinic7())
	w := h.get("/api/v1/clinics/7/operating-hours")

	require.Equal(t, http.StatusOK, w.Code)
	requireNoStore(t, w)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	assert.Equal(t, string(operatingHoursFixture(t, "get-configured.response.json")), w.Body.String())
	assert.Equal(t, []uint64{7}, h.repo.gotIDs)
	assert.Equal(t, []permissionCall{{"assigned", string(model.ResourceHospitalSettings), "view"}}, h.routeChecks)
}

func TestGetClinicOperatingHours_Unconfigured(t *testing.T) {
	h := newOperatingHoursHarness(t, &fakeOperatingHoursRepository{}, memberOfClinic7())
	w := h.get("/api/v1/clinics/7/operating-hours")

	require.Equal(t, http.StatusOK, w.Code)
	requireNoStore(t, w)
	assert.Equal(t, string(operatingHoursFixture(t, "get-unconfigured.response.json")), w.Body.String())
}

func TestGetClinicOperatingHours_NegativeFixturesAreNeverProduced(t *testing.T) {
	configured := storedRowFromFixture(t, "get-configured.response.json")
	rows := map[string]*fakeOperatingHoursRepository{
		"configured":   {found: true, row: configured},
		"unconfigured": {},
	}
	negative := string(operatingHoursFixture(t, "get-configured-null-settings.response.json"))
	for name, repo := range rows {
		w := newOperatingHoursHarness(t, repo, memberOfClinic7()).get("/api/v1/clinics/7/operating-hours")
		require.Equal(t, http.StatusOK, w.Code, name)
		assert.NotEqual(t, negative, w.Body.String(), name)
		assert.NotContains(t, w.Body.String(), `"state":"configured","version":null`)
	}
}

func TestGetClinicOperatingHours_ReasonUsesCanonicalMarshalBytes(t *testing.T) {
	row := operatingHoursStoredRow(operatingHoursSevenClosedDays, `[{"date":"2026-11-03","reason":"a<b>&\"c\""}]`)
	h := newOperatingHoursHarness(t, &fakeOperatingHoursRepository{found: true, row: row}, memberOfClinic7())
	w := h.get("/api/v1/clinics/7/operating-hours")
	require.Equal(t, http.StatusOK, w.Code)

	loaded, err := LoadOperatingHoursSettings(row)
	require.NoError(t, err)
	canonical, err := loaded.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, w.Body.String(), `"settings":`+string(canonical))
	assert.Contains(t, w.Body.String(), "a\\u003cb\\u003e\\u0026\\\"c\\\"", "HTML-escaped by the canonical MarshalJSON")
}

func TestGetClinicOperatingHours_ClinicAuthorization(t *testing.T) {
	t.Run("URL clinic the actor does not belong to is 403 and never reads", func(t *testing.T) {
		id := requestIdentity{selectedClinic: "1", memberClinics: []uint64{1}, grant: grantAll}
		h := newOperatingHoursHarness(t, configuredRepo(t), id)
		w := h.get("/api/v1/clinics/7/operating-hours")
		require.Equal(t, http.StatusForbidden, w.Code)
		requireNoStore(t, w)
		requireErrorBody(t, w, "")
		assert.Zero(t, h.repo.calls)
	})

	t.Run("member without hospital-settings:view on the URL clinic is 403", func(t *testing.T) {
		id := requestIdentity{
			selectedClinic: "2", memberClinics: []uint64{1, 2},
			grant: func(clinicID uint64, resource, action string) bool {
				return clinicID == 2 && resource == string(model.ResourceHospitalSettings) && action == "view"
			},
		}
		h := newOperatingHoursHarness(t, configuredRepo(t), id)
		w := h.get("/api/v1/clinics/1/operating-hours")
		require.Equal(t, http.StatusForbidden, w.Code)
		requireNoStore(t, w)
		assert.Zero(t, h.repo.calls)
	})

	t.Run("member with the grant on the URL clinic reads that clinic, not the selected one", func(t *testing.T) {
		id := requestIdentity{
			selectedClinic: "2", memberClinics: []uint64{1, 2},
			grant: func(clinicID uint64, resource, action string) bool {
				return resource == string(model.ResourceHospitalSettings) && action == "view"
			},
		}
		row := storedRowFromFixture(t, "get-configured.response.json")
		row.ClinicID = 1
		h := newOperatingHoursHarness(t, &fakeOperatingHoursRepository{found: true, row: row}, id)
		w := h.get("/api/v1/clinics/1/operating-hours")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, []uint64{1}, h.repo.gotIDs)
		assert.Contains(t, w.Body.String(), `"clinic_id":"1"`)
	})

	t.Run("system admin passes the grant check for an active clinic in the context", func(t *testing.T) {
		id := requestIdentity{selectedClinic: "1", memberClinics: []uint64{1, 7}, systemAdmin: true, grant: func(uint64, string, string) bool { return false }}
		h := newOperatingHoursHarness(t, configuredRepo(t), id)
		w := h.get("/api/v1/clinics/7/operating-hours")
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("route-level denial keeps no-store and does not read", func(t *testing.T) {
		h := newOperatingHoursHarness(t, configuredRepo(t), memberOfClinic7())
		h.denyRoute = func(permissionCall) bool { return true }
		w := h.get("/api/v1/clinics/7/operating-hours")
		require.Equal(t, http.StatusForbidden, w.Code)
		requireNoStore(t, w)
		assert.Zero(t, h.repo.calls)
	})
}

func TestGetClinicOperatingHours_InvalidClinicIDIs400(t *testing.T) {
	for _, raw := range []string{"0", "abc", "007", "-1", "1.5", "9223372036854775808", "18446744073709551616", "%207"} {
		t.Run(raw, func(t *testing.T) {
			h := newOperatingHoursHarness(t, configuredRepo(t), memberOfClinic7())
			w := h.get("/api/v1/clinics/" + raw + "/operating-hours")
			require.Equal(t, http.StatusBadRequest, w.Code)
			requireNoStore(t, w)
			requireErrorBody(t, w, "")
			assert.Zero(t, h.repo.calls)
		})
	}
}

func TestGetClinicOperatingHours_MaxInt64ClinicIDIsAccepted(t *testing.T) {
	const maxID = "9223372036854775807"
	id := requestIdentity{selectedClinic: maxID, memberClinics: []uint64{9223372036854775807}, grant: grantAll}
	h := newOperatingHoursHarness(t, &fakeOperatingHoursRepository{}, id)
	w := h.get("/api/v1/clinics/" + maxID + "/operating-hours")
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"clinic_id":"`+maxID+`"`)
}

func TestGetClinicOperatingHours_ServerSideFailuresAre503(t *testing.T) {
	corruptRows := map[string]model.ClinicOperatingHours{
		"corrupt weekly_hours": operatingHoursStoredRow(`{"not":"an array"}`, `[]`),
		"corrupt closed_dates": operatingHoursStoredRow(operatingHoursSevenClosedDays, `[null]`),
		"unknown schema_version": func() model.ClinicOperatingHours {
			r := operatingHoursStoredRow(operatingHoursSevenClosedDays, `[]`)
			r.SchemaVersion = "x"
			return r
		}(),
		"version zero": func() model.ClinicOperatingHours {
			r := operatingHoursStoredRow(operatingHoursSevenClosedDays, `[]`)
			r.Version = 0
			return r
		}(),
		"row for a different clinic": func() model.ClinicOperatingHours {
			r := operatingHoursStoredRow(operatingHoursSevenClosedDays, `[]`)
			r.ClinicID = 99
			return r
		}(),
	}
	for name, row := range corruptRows {
		t.Run(name, func(t *testing.T) {
			h := newOperatingHoursHarness(t, &fakeOperatingHoursRepository{found: true, row: row}, memberOfClinic7())
			w := h.get("/api/v1/clinics/7/operating-hours")
			require.Equal(t, http.StatusServiceUnavailable, w.Code)
			requireNoStore(t, w)
			requireErrorBody(t, w, operatingHoursUnavailableMessage)
			assert.NotContains(t, w.Body.String(), "unconfigured")
			assert.NotContains(t, w.Body.String(), "settings")
		})
	}

	t.Run("repository error leaks nothing and is not unconfigured", func(t *testing.T) {
		repo := &fakeOperatingHoursRepository{err: errors.New("pq: password authentication failed for user secret")}
		w := newOperatingHoursHarness(t, repo, memberOfClinic7()).get("/api/v1/clinics/7/operating-hours")
		require.Equal(t, http.StatusServiceUnavailable, w.Code)
		requireNoStore(t, w)
		requireErrorBody(t, w, operatingHoursUnavailableMessage)
		assert.NotContains(t, w.Body.String(), "secret")
	})
}

func TestListCommonHolidays_Fixtures(t *testing.T) {
	storedClosed := `[{"date":"2026-10-31","reason":"前月末"},{"date":"2026-11-03","reason":"臨時休診"},{"date":"2026-12-01","reason":"翌月初"}]`
	configured := &fakeOperatingHoursRepository{found: true, row: operatingHoursStoredRow(operatingHoursSevenClosedDays, storedClosed)}

	for _, path := range []string{"/api/v1/clinic-holidays/common", "/api/v1/closing-settings/holidays/common"} {
		t.Run(path+" configured", func(t *testing.T) {
			h := newOperatingHoursHarness(t, configured, memberOfClinic7())
			w := h.get(path + "?year_month=2026-11")
			require.Equal(t, http.StatusOK, w.Code)
			requireNoStore(t, w)
			assert.Equal(t, string(operatingHoursFixture(t, "holiday-month-configured.response.json")), w.Body.String())
		})
		t.Run(path+" unconfigured", func(t *testing.T) {
			h := newOperatingHoursHarness(t, &fakeOperatingHoursRepository{}, memberOfClinic7())
			w := h.get(path + "?year_month=2026-11")
			require.Equal(t, http.StatusOK, w.Code)
			requireNoStore(t, w)
			assert.Equal(t, string(operatingHoursFixture(t, "holiday-month-unconfigured.response.json")), w.Body.String())
		})
	}

	t.Run("negative holiday fixtures are never produced", func(t *testing.T) {
		negatives := []string{
			"holiday-month-configured-null-version.response.json",
			"holiday-month-unconfigured-with-closed-dates.response.json",
			"holiday-month-unconfigured-with-version.response.json",
		}
		for _, repo := range []*fakeOperatingHoursRepository{configured, {}} {
			w := newOperatingHoursHarness(t, repo, memberOfClinic7()).get("/api/v1/clinic-holidays/common?year_month=2026-11")
			require.Equal(t, http.StatusOK, w.Code)
			for _, name := range negatives {
				assert.NotEqual(t, string(operatingHoursFixture(t, name)), w.Body.String(), name)
			}
		}
	})
}

func TestListCommonHolidays_ReturnsOnlyTheMonthAndNeverTheFullSettings(t *testing.T) {
	closed := `[{"date":"2026-10-31","reason":"a"},{"date":"2026-11-01","reason":"b"},{"date":"2026-11-30","reason":"c"},{"date":"2026-12-01","reason":"d"}]`
	repo := &fakeOperatingHoursRepository{found: true, row: operatingHoursStoredRow(operatingHoursSevenClosedDays, closed)}
	h := newOperatingHoursHarness(t, repo, memberOfClinic7())
	w := h.get("/api/v1/clinic-holidays/common?year_month=2026-11")
	require.Equal(t, http.StatusOK, w.Code)

	var body struct {
		ClinicID    string `json:"clinic_id"`
		YearMonth   string `json:"year_month"`
		State       string `json:"state"`
		Version     string `json:"version"`
		ClosedDates []struct {
			Date   string `json:"date"`
			Reason string `json:"reason"`
		} `json:"closed_dates"`
	}
	dec := json.NewDecoder(bytes.NewReader(w.Body.Bytes()))
	dec.DisallowUnknownFields()
	require.NoError(t, dec.Decode(&body), "exact key set: no weekly_hours, national_holiday_closed or settings")
	assert.Equal(t, "7", body.ClinicID)
	assert.Equal(t, "2026-11", body.YearMonth)
	assert.Equal(t, "configured", body.State)
	assert.Equal(t, "3", body.Version)
	require.Len(t, body.ClosedDates, 2)
	assert.Equal(t, "2026-11-01", body.ClosedDates[0].Date)
	assert.Equal(t, "2026-11-30", body.ClosedDates[1].Date)
	assert.NotContains(t, w.Body.String(), "weekly_hours")

	t.Run("configured month without closed dates is an empty array, not null", func(t *testing.T) {
		w := newOperatingHoursHarness(t, repo, memberOfClinic7()).get("/api/v1/clinic-holidays/common?year_month=2026-02")
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"clinic_id":"7","year_month":"2026-02","state":"configured","version":"3","closed_dates":[]}`, w.Body.String())
	})
}

func TestListCommonHolidays_UsesTheSelectedClinicNotTheQuery(t *testing.T) {
	repo := &fakeOperatingHoursRepository{}
	h := newOperatingHoursHarness(t, repo, memberOfClinic7())
	w := h.get("/api/v1/clinic-holidays/common?year_month=2026-11&clinic_id=9")
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []uint64{7}, repo.gotIDs)
	assert.Contains(t, w.Body.String(), `"clinic_id":"7"`)
}

func TestListCommonHolidays_InvalidYearMonthIs400(t *testing.T) {
	queries := []string{
		"", "?year_month=", "?year_month=2026-13", "?year_month=2026-00", "?year_month=2026-1", "?year_month=26-11",
		"?year_month=2026-11-01", "?year_month=2026-11%20", "?year_month=abcd-ef", "?year_month=2026-11&year_month=2026-12",
		"?year-month=2026-11", "?YEAR_MONTH=2026-11",
	}
	for _, path := range []string{"/api/v1/clinic-holidays/common", "/api/v1/closing-settings/holidays/common"} {
		for _, q := range queries {
			t.Run(path+q, func(t *testing.T) {
				h := newOperatingHoursHarness(t, configuredRepo(t), memberOfClinic7())
				w := h.get(path + q)
				require.Equal(t, http.StatusBadRequest, w.Code)
				requireNoStore(t, w)
				requireErrorBody(t, w, "")
				assert.Zero(t, h.repo.calls)
			})
		}
	}
}

func TestListCommonHolidays_ServerSideFailuresAre503(t *testing.T) {
	for _, path := range []string{"/api/v1/clinic-holidays/common", "/api/v1/closing-settings/holidays/common"} {
		repos := map[string]*fakeOperatingHoursRepository{
			"repository error": {err: errors.New("dial tcp: connection refused")},
			"corrupt row":      {found: true, row: operatingHoursStoredRow(operatingHoursSevenClosedDays, `[{"date":"2026-11-03"}]`)},
		}
		for name, repo := range repos {
			t.Run(path+" "+name, func(t *testing.T) {
				w := newOperatingHoursHarness(t, repo, memberOfClinic7()).get(path + "?year_month=2026-11")
				require.Equal(t, http.StatusServiceUnavailable, w.Code)
				requireNoStore(t, w)
				requireErrorBody(t, w, operatingHoursUnavailableMessage)
				assert.NotContains(t, w.Body.String(), "closed_dates")
			})
		}
	}
}

// 旧 ListClinicHolidays が共有 handler で課す「選択医院の shifts:view」を、closing 経路の新 GET も同じ形で課す。
func TestListCommonHolidays_SharedHandlerRequiresSelectedClinicShiftsView(t *testing.T) {
	onlyClosingSettings := func(_ uint64, resource, action string) bool {
		return resource == string(model.ResourceClosingSettings) && action == "view"
	}
	onlyShifts := func(_ uint64, resource, action string) bool {
		return resource == string(model.ResourceShifts) && action == "view"
	}
	for _, path := range []string{"/api/v1/closing-settings/holidays/common", "/api/v1/clinic-holidays/common"} {
		t.Run(path+" without shifts:view in the selected clinic", func(t *testing.T) {
			id := memberOfClinic7()
			id.grant = onlyClosingSettings
			h := newOperatingHoursHarness(t, configuredRepo(t), id)
			w := h.get(path + "?year_month=2026-11")
			require.Equal(t, http.StatusForbidden, w.Code)
			requireNoStore(t, w)
			assert.Zero(t, h.repo.calls)
		})
	}

	t.Run("closing route with shifts:view in the selected clinic passes the handler check", func(t *testing.T) {
		id := memberOfClinic7()
		id.grant = onlyShifts
		h := newOperatingHoursHarness(t, configuredRepo(t), id)
		w := h.get("/api/v1/closing-settings/holidays/common?year_month=2026-11")
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("shifts:view held only by another assigned clinic is not enough", func(t *testing.T) {
		id := requestIdentity{
			selectedClinic: "7", memberClinics: []uint64{3, 7},
			grant: func(clinicID uint64, resource, action string) bool {
				return clinicID == 3 && resource == string(model.ResourceShifts) && action == "view"
			},
		}
		h := newOperatingHoursHarness(t, configuredRepo(t), id)
		w := h.get("/api/v1/clinic-holidays/common?year_month=2026-11")
		require.Equal(t, http.StatusForbidden, w.Code)
		assert.Zero(t, h.repo.calls)
	})

	t.Run("route-level closing-settings:view denial keeps no-store", func(t *testing.T) {
		h := newOperatingHoursHarness(t, configuredRepo(t), memberOfClinic7())
		h.denyRoute = func(c permissionCall) bool { return c.resource == string(model.ResourceClosingSettings) }
		w := h.get("/api/v1/closing-settings/holidays/common?year_month=2026-11")
		requireNoStore(t, w)
		assert.Zero(t, h.repo.calls)
		assert.True(t, slices.Contains(h.routeChecks, permissionCall{"strict", string(model.ResourceClosingSettings), "view"}))
	})
}

func TestListCommonHolidays_InvalidSelectedClinicContext(t *testing.T) {
	cases := map[string]int{
		"abc":                 http.StatusBadRequest,
		"9223372036854775808": http.StatusBadRequest, // above the int64 wire range
		"0":                   http.StatusForbidden,  // the shared selected-clinic grant check never passes for clinic 0
	}
	for selected, want := range cases {
		t.Run(selected, func(t *testing.T) {
			id := memberOfClinic7()
			id.selectedClinic = selected
			h := newOperatingHoursHarness(t, configuredRepo(t), id)
			w := h.get("/api/v1/clinic-holidays/common?year_month=2026-11")
			require.Equal(t, want, w.Code)
			requireNoStore(t, w)
			assert.Zero(t, h.repo.calls)
		})
	}
}

func TestOperatingHoursFixtures_AreTheOnesThisSliceCompares(t *testing.T) {
	// Guard against the fixture directory being moved: a missing file must fail loudly, not skip.
	for _, name := range []string{
		"get-configured.response.json", "get-unconfigured.response.json",
		"holiday-month-configured.response.json", "holiday-month-unconfigured.response.json",
	} {
		assert.True(t, strings.HasSuffix(name, ".response.json"))
		assert.NotEmpty(t, operatingHoursFixture(t, name))
	}
}

// line-reservation-settings の MembershipABGrantASelectedB と同じ形: 所属 A,B・選択 B・grant は A のみ。
// route middleware（assigned fallback）は通り、URL 医院の認可は handler 内の AuthorizeClinicIDsForPermission が行う。
func TestGetClinicOperatingHours_MembershipABGrantASelectedB(t *testing.T) {
	identity := func() requestIdentity {
		return requestIdentity{
			selectedClinic: "2", memberClinics: []uint64{1, 2},
			grant: func(clinicID uint64, resource, action string) bool {
				return clinicID == 1 && resource == string(model.ResourceHospitalSettings) && action == "view"
			},
		}
	}
	rowFor := func(t *testing.T, clinicID uint64) *fakeOperatingHoursRepository {
		row := storedRowFromFixture(t, "get-configured.response.json")
		row.ClinicID = clinicID
		return &fakeOperatingHoursRepository{found: true, row: row}
	}

	t.Run("path clinic A with the grant on A only reads clinic A", func(t *testing.T) {
		h := newOperatingHoursHarness(t, rowFor(t, 1), identity())
		w := h.get("/api/v1/clinics/1/operating-hours")
		require.Equal(t, http.StatusOK, w.Code)
		requireNoStore(t, w)
		assert.Equal(t, []uint64{1}, h.repo.gotIDs)
		assert.Contains(t, w.Body.String(), `"clinic_id":"1"`)
		assert.Equal(t, []permissionCall{{"assigned", string(model.ResourceHospitalSettings), "view"}}, h.routeChecks)
	})

	t.Run("path clinic B (selected, no grant) is 403 and never reads", func(t *testing.T) {
		h := newOperatingHoursHarness(t, rowFor(t, 2), identity())
		w := h.get("/api/v1/clinics/2/operating-hours")
		require.Equal(t, http.StatusForbidden, w.Code)
		requireNoStore(t, w)
		requireErrorBody(t, w, "forbidden")
		assert.Zero(t, h.repo.calls)
	})

	t.Run("non-member path clinic keeps the existing 'not assigned' 403 and never reads", func(t *testing.T) {
		h := newOperatingHoursHarness(t, rowFor(t, 9), identity())
		w := h.get("/api/v1/clinics/9/operating-hours")
		require.Equal(t, http.StatusForbidden, w.Code)
		requireNoStore(t, w)
		requireErrorBody(t, w, "not assigned to this clinic")
		assert.Zero(t, h.repo.calls)
	})
}
