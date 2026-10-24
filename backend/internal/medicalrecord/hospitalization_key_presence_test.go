package medicalrecord

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// 退役済み・滞在由来の入院フィールドは値ではなく JSON キーの存在で拒否する（PR #851 review A2）。
// `*T` フィールドへのバインドでは null が未送信と区別できず、`"planned_exit_at": null` などが
// 黙って受理されていた。ここでは実際の JSON を decode する境界で検証する。

func serveCreateHospitalizationJSON(t *testing.T, svc *mockHospitalizationService, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	newHandlerWithHospitalizationSvc(svc).CreateHospitalization(c)
	return w
}

func serveUpdateHospitalizationJSON(t *testing.T, svc *mockHospitalizationService, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	setClinicID(c)
	newHandlerWithHospitalizationSvc(svc).UpdateHospitalization(c)
	return w
}

// unreachableHospitalizationSvc は拒否されるはずのリクエストがサービスへ到達したらテストを失敗させる。
func unreachableHospitalizationSvc(t *testing.T) *mockHospitalizationService {
	t.Helper()
	return &mockHospitalizationService{
		createFn: func(context.Context, uint64, *CreateHospitalizationInput) (*model.Hospitalization, error) {
			t.Error("a refused create must not reach the service")
			return &model.Hospitalization{ID: 1}, nil
		},
		updateFn: func(context.Context, uint64, uint64, *UpdateHospitalizationInput) (*model.Hospitalization, error) {
			t.Error("a refused update must not reach the service")
			return &model.Hospitalization{ID: 1}, nil
		},
	}
}

// 値の種類（null / 文字列）に関わらず、キーが存在すれば 422 の固定エラーになる。
var refusedKeyValues = []struct{ name, raw string }{{"null", `null`}, {"string", `"admitted"`}}

func TestCreateHospitalization_RetiredKeyIsRefusedByPresenceEvenWhenNull(t *testing.T) {
	for _, field := range []string{"hospitalization_type", "start_date", "end_date", "status"} {
		for _, value := range refusedKeyValues {
			t.Run(field+"="+value.name, func(t *testing.T) {
				body := `{"owner_id":1,"pet_id":2,"stay_kind":"inpatient","` + field + `":` + value.raw + `}`

				w := serveCreateHospitalizationJSON(t, unreachableHospitalizationSvc(t), body)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
				assert.Contains(t, w.Body.String(), codeStayLifecycleWriteUnavailable)
				assert.Contains(t, w.Body.String(), field+" は")
			})
		}
	}
}

func TestCreateHospitalization_NullOnWritableKeysIsStillAccepted(t *testing.T) {
	var got *CreateHospitalizationInput
	svc := &mockHospitalizationService{
		createFn: func(_ context.Context, _ uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error) {
			got = input
			return &model.Hospitalization{ID: 1}, nil
		},
	}

	w := serveCreateHospitalizationJSON(t, svc, `{"owner_id":1,"pet_id":2,"stay_kind":"hotel","planned_entry_at":null,"planned_exit_at":null,"cage_id":null}`)

	require.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, got)
	assert.Nil(t, got.PlannedEntryAt)
	assert.Nil(t, got.PlannedExitAt)
	assert.Nil(t, got.CageID)
}

func TestUpdateHospitalization_StayOwnedKeyIsRefusedByPresenceEvenWhenNull(t *testing.T) {
	fields := []string{"stay_kind", "planned_entry_at", "planned_exit_at", "hospitalization_type", "start_date", "end_date", "status"}
	for _, field := range fields {
		for _, value := range refusedKeyValues {
			t.Run(field+"="+value.name, func(t *testing.T) {
				w := serveUpdateHospitalizationJSON(t, unreachableHospitalizationSvc(t), `{"`+field+`":`+value.raw+`}`)

				assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
				assert.Contains(t, w.Body.String(), codeStayLifecycleWriteUnavailable)
				assert.Contains(t, w.Body.String(), field+" は")
			})
		}
	}
}

func TestUpdateHospitalization_KeyPresenceMatchIsCaseInsensitiveLikeTheDecoder(t *testing.T) {
	w := serveUpdateHospitalizationJSON(t, unreachableHospitalizationSvc(t), `{"Planned_Exit_At":null}`)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "planned_exit_at は")
}

func TestUpdateHospitalization_NullOnWritableKeyIsStillAccepted(t *testing.T) {
	var got *UpdateHospitalizationInput
	svc := &mockHospitalizationService{
		updateFn: func(_ context.Context, _, _ uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error) {
			got = input
			return &model.Hospitalization{ID: 1}, nil
		},
	}

	w := serveUpdateHospitalizationJSON(t, svc, `{"memo":null}`)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, got)
	assert.Nil(t, got.Memo)
}
