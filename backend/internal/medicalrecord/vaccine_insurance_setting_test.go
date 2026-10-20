package medicalrecord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-455: 予防接種マスタの品目別保険設定（is_non_insurance）。

func vaccineBoolPtr(v bool) *bool { return &v }

func TestCreateVaccineRequest_IsNonInsurance_Decode(t *testing.T) {
	tests := []struct {
		name string
		body string
		want *bool
	}{
		{name: "未送信はnil（service が新規初期値を適用）", body: `{"version":1,"name":"v"}`, want: nil},
		{name: "trueは明示値", body: `{"version":1,"name":"v","is_non_insurance":true}`, want: vaccineBoolPtr(true)},
		{name: "falseは明示値", body: `{"version":1,"name":"v","is_non_insurance":false}`, want: vaccineBoolPtr(false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req createVaccineRequest
			require.NoError(t, json.Unmarshal([]byte(tt.body), &req))
			assert.Equal(t, tt.want, req.toServiceInput().IsNonInsurance)
		})
	}
}

func TestCreateVaccineRequest_IsNonInsurance_RejectsNullAndNonBoolean(t *testing.T) {
	for _, body := range []string{
		`{"version":1,"name":"v","is_non_insurance":null}`,
		`{"version":1,"name":"v","is_non_insurance":"true"}`,
		`{"version":1,"name":"v","is_non_insurance":1}`,
	} {
		t.Run(body, func(t *testing.T) {
			var req createVaccineRequest
			err := json.Unmarshal([]byte(body), &req)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "want invalid input, got %v", err)
		})
	}
}

func TestUpdateVaccineRequest_IsNonInsurance_Decode(t *testing.T) {
	tests := []struct {
		name string
		body string
		want *bool
	}{
		{name: "未送信は変更なし", body: `{"version":1,"name":"v"}`, want: nil},
		{name: "trueは明示保存", body: `{"version":1,"is_non_insurance":true}`, want: vaccineBoolPtr(true)},
		{name: "falseは明示保存", body: `{"version":1,"is_non_insurance":false}`, want: vaccineBoolPtr(false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req updateVaccineRequest
			require.NoError(t, json.Unmarshal([]byte(tt.body), &req))
			assert.Equal(t, tt.want, req.toServiceInput().IsNonInsurance)
		})
	}
}

func TestUpdateVaccineRequest_IsNonInsurance_RejectsNullAndNonBoolean(t *testing.T) {
	for _, body := range []string{
		`{"version":1,"is_non_insurance":null}`,
		`{"version":1,"is_non_insurance":"true"}`,
		`{"version":1,"is_non_insurance":1}`,
	} {
		t.Run(body, func(t *testing.T) {
			var req updateVaccineRequest
			err := json.Unmarshal([]byte(body), &req)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "want invalid input, got %v", err)
		})
	}
}

func TestVaccineService_Create_IsNonInsurance(t *testing.T) {
	tests := []struct {
		name  string
		input *bool
		want  bool
	}{
		{name: "未指定は新規初期値の対象外(true)を保存", input: nil, want: true},
		{name: "明示falseは対象として保存", input: vaccineBoolPtr(false), want: false},
		{name: "明示trueは対象外として保存", input: vaccineBoolPtr(true), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var saved *model.Vaccine
			repo := &mockVaccineRepository{
				createFn: func(_ context.Context, v *model.Vaccine) error {
					saved = v
					return nil
				},
			}
			svc := newVaccineServiceForTest(repo)

			_, err := svc.Create(context.Background(), 1, &CreateVaccineInput{ActorID: vaccineTestActorID(), Name: "v", IsNonInsurance: tt.input})

			require.NoError(t, err)
			require.NotNil(t, saved)
			require.NotNil(t, saved.IsNonInsurance)
			assert.Equal(t, tt.want, *saved.IsNonInsurance)
		})
	}
}

func TestVaccineService_Update_IsNonInsuranceOnly(t *testing.T) {
	var gotCmd UpdateVaccineInput
	repo := &mockVaccineRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccine, error) {
			return &model.Vaccine{ID: id, ClinicID: 1, Name: "v"}, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, cmd UpdateVaccineInput) (*model.Vaccine, error) {
			gotCmd = cmd
			return &model.Vaccine{ID: id, ClinicID: 1, Name: "v", IsNonInsurance: cmd.IsNonInsurance}, nil
		},
	}
	svc := newVaccineServiceForTest(repo)

	got, err := svc.Update(context.Background(), 1, 9, &UpdateVaccineInput{Version: 1, ActorID: vaccineTestActorID(), IsNonInsurance: vaccineBoolPtr(false)})

	require.NoError(t, err, "is_non_insurance 単独更新は at-least-one-field ガードを通過する")
	require.NotNil(t, gotCmd.IsNonInsurance)
	assert.False(t, *gotCmd.IsNonInsurance)
	require.NotNil(t, got.IsNonInsurance)
	assert.False(t, *got.IsNonInsurance)
}

func TestBuildVaccineUpdate_IsNonInsurance(t *testing.T) {
	name := "renamed"

	assert.Equal(t, map[string]any{colVaccineIsNonInsurance: true},
		buildVaccineUpdate(&UpdateVaccineInput{Version: 1, ActorID: vaccineTestActorID(), IsNonInsurance: vaccineBoolPtr(true)}))
	assert.Equal(t, map[string]any{colVaccineIsNonInsurance: false},
		buildVaccineUpdate(&UpdateVaccineInput{Version: 1, ActorID: vaccineTestActorID(), IsNonInsurance: vaccineBoolPtr(false)}))

	fields := buildVaccineUpdate(&UpdateVaccineInput{Version: 1, ActorID: vaccineTestActorID(), Name: &name})
	_, touched := fields[colVaccineIsNonInsurance]
	assert.False(t, touched, "未指定の再編集は保存済み/旧欠落値を上書きしない")
}

func TestToVaccineResponse_IsNonInsurancePreservesMissing(t *testing.T) {
	tests := []struct {
		name string
		in   *bool
		want string
	}{
		{name: "旧欠落はnullのまま（falseへ補完しない）", in: nil, want: `null`},
		{name: "対象外", in: vaccineBoolPtr(true), want: `true`},
		{name: "対象", in: vaccineBoolPtr(false), want: `false`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := json.Marshal(toVaccineResponse(&model.Vaccine{ID: 1, Name: "v", IsNonInsurance: tt.in}))
			require.NoError(t, err)
			var body map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(raw, &body))
			value, ok := body["is_non_insurance"]
			require.True(t, ok, "is_non_insurance は常に wire に出す")
			assert.JSONEq(t, tt.want, string(value))
		})
	}
}

func serveVaccineInsuranceRequest(t *testing.T, method string, body string, svc VaccineService) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	setClinicID(c)
	setStaffID(c)
	h := NewVaccineHandler(svc)
	if method == http.MethodPost {
		h.CreateVaccine(c)
	} else {
		h.UpdateVaccine(c)
	}
	return w
}

func TestUpdateVaccineHandler_IsNonInsurance(t *testing.T) {
	t.Run("null は 400 で service を呼ばない", func(t *testing.T) {
		svc := &mockVaccineService{
			updateFn: func(context.Context, uint64, uint64, *UpdateVaccineInput) (*model.Vaccine, error) {
				t.Fatal("service must not be called for null is_non_insurance")
				return nil, nil
			},
		}
		w := serveVaccineInsuranceRequest(t, http.MethodPatch, `{"version":1,"is_non_insurance":null}`, svc)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("明示 false を service へ渡し再読込値を返す", func(t *testing.T) {
		svc := &mockVaccineService{
			updateFn: func(_ context.Context, _, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error) {
				require.NotNil(t, input.IsNonInsurance)
				assert.False(t, *input.IsNonInsurance)
				return &model.Vaccine{ID: id, Name: "v", IsNonInsurance: input.IsNonInsurance}, nil
			},
		}
		w := serveVaccineInsuranceRequest(t, http.MethodPatch, `{"version":1,"is_non_insurance":false}`, svc)
		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `false`, string(decodeVaccineField(t, w, "is_non_insurance")))
	})
}

func TestCreateVaccineHandler_IsNonInsuranceOmittedReachesServiceAsNil(t *testing.T) {
	svc := &mockVaccineService{
		createFn: func(_ context.Context, _ uint64, input *CreateVaccineInput) (*model.Vaccine, error) {
			assert.Nil(t, input.IsNonInsurance, "未送信は service の新規初期値へ委ねる")
			v := true
			return &model.Vaccine{ID: 1, Name: input.Name, IsNonInsurance: &v}, nil
		},
	}
	w := serveVaccineInsuranceRequest(t, http.MethodPost, `{"version":1,"name":"v"}`, svc)
	require.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, `true`, string(decodeVaccineField(t, w, "is_non_insurance")))
}

func TestCreateVaccineHandler_IsNonInsuranceRejectsNull(t *testing.T) {
	svc := &mockVaccineService{
		createFn: func(context.Context, uint64, *CreateVaccineInput) (*model.Vaccine, error) {
			t.Fatal("service must not be called for null is_non_insurance")
			return nil, nil
		},
	}
	w := serveVaccineInsuranceRequest(t, http.MethodPost, `{"version":1,"name":"v","is_non_insurance":null}`, svc)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func decodeVaccineField(t *testing.T, w *httptest.ResponseRecorder, key string) json.RawMessage {
	t.Helper()
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	value, ok := body[key]
	require.True(t, ok, "%s missing in response", key)
	return value
}
