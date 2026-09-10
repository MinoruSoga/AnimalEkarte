package testdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicGrantFixture is the shared A/B clinic + dual-membership staff seed used
// by D3 selected-clinic grant isolation tests.
type ClinicGrantFixture struct {
	ClinicA     uint64
	ClinicB     uint64
	StaffID     uint64
	ClinicAName string
	ClinicBName string
	StaffName   string
}

// SeedDualClinicGrantFixture creates clinics A/B under one company and a
// non-admin staff with main membership on A plus assignment on B.
func SeedDualClinicGrantFixture(t *testing.T, db *gorm.DB, namePrefix string) ClinicGrantFixture {
	t.Helper()
	require.NoError(t, EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
	))

	ctx := context.Background()
	clinicAName := namePrefix + " clinic A"
	clinicBName := namePrefix + " clinic B"
	staffName := namePrefix + " dual-member non-admin"

	company := &model.Company{Name: namePrefix + " 法人"}
	require.NoError(t, db.WithContext(ctx).Create(company).Error)

	clinicA := &model.Clinic{CompanyID: company.ID, Name: clinicAName, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(clinicA).Error)
	clinicB := &model.Clinic{CompanyID: company.ID, Name: clinicBName, IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(clinicB).Error)

	staff := &model.Staff{
		ClinicID:  clinicA.ID,
		Name:      staffName,
		StaffType: model.StaffTypeDoctor,
		IsActive:  true,
	}
	require.NoError(t, db.WithContext(ctx).Create(staff).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicA.ID,
		IsMain:   true,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicB.ID,
		IsMain:   false,
	}).Error)

	return ClinicGrantFixture{
		ClinicA:     clinicA.ID,
		ClinicB:     clinicB.ID,
		StaffID:     staff.ID,
		ClinicAName: clinicAName,
		ClinicBName: clinicBName,
		StaffName:   staffName,
	}
}

// ConfigureSelectedClinicBGrant sets selected clinic B, membership A+B, and a
// ClinicPermissionChecker that grants resource/view only for grantClinicIDs.
func ConfigureSelectedClinicBGrant(
	fx ClinicGrantFixture,
	resource string,
	grantClinicIDs ...uint64,
) func(*gin.Context) {
	granted := make(map[uint64]struct{}, len(grantClinicIDs))
	for _, id := range grantClinicIDs {
		granted[id] = struct{}{}
	}
	return func(c *gin.Context) {
		c.Set("clinic_id", fmt.Sprintf("%d", fx.ClinicB))
		c.Set("clinic_ids", []uint64{fx.ClinicA, fx.ClinicB})
		c.Set("user_id", fmt.Sprintf("%d", fx.StaffID))
		c.Set("is_system_admin", false)
		httpapi.SetClinicPermissionChecker(c, func(
			_ *gin.Context,
			clinicID uint64,
			res, action string,
		) bool {
			if res != resource || action != "view" {
				return false
			}
			_, ok := granted[clinicID]
			return ok
		})
	}
}

// NewHTTPTestContext builds a gin test context with optional configure hook.
func NewHTTPTestContext(
	t *testing.T,
	method, path string,
	configure func(*gin.Context),
) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(method, path, bytes.NewReader(nil))
	if configure != nil {
		configure(c)
	}
	return c, response
}

// AssertBodyOmitsClinicArtifacts fails when the body embeds a forbidden clinic
// id (JSON number) or any forbidden name substring.
func AssertBodyOmitsClinicArtifacts(
	t *testing.T,
	body []byte,
	forbiddenClinicIDs []uint64,
	forbiddenNames ...string,
) {
	t.Helper()
	text := string(body)
	for _, id := range forbiddenClinicIDs {
		needle := fmt.Sprintf(`"clinic_id":%d`, id)
		if strings.Contains(text, needle) {
			t.Fatalf("response body must omit clinic_id=%d: %s", id, text)
		}
	}
	for _, name := range forbiddenNames {
		if name == "" {
			continue
		}
		if strings.Contains(text, name) {
			t.Fatalf("response body must omit %q: %s", name, text)
		}
	}
	var generic map[string]json.RawMessage
	if err := json.Unmarshal(body, &generic); err == nil {
		if raw, ok := generic["clinic_id"]; ok {
			var clinicID uint64
			if json.Unmarshal(raw, &clinicID) == nil {
				for _, forbidden := range forbiddenClinicIDs {
					if clinicID == forbidden {
						t.Fatalf("response body clinic_id=%d is forbidden", clinicID)
					}
				}
			}
		}
	}
}
