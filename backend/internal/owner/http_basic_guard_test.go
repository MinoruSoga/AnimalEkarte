package owner

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func emr491Keys(t *testing.T, body string) map[string]json.RawMessage {
	t.Helper()
	var keys map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(body), &keys))
	return keys
}

func emr491PatchOwner(t *testing.T, svc Service, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := newHandlerWithOwnerSvc(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader([]byte(body)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	setClinicID(c)
	h.UpdateOwner(c)
	return w
}

func TestRejectPresentOwnerBasicKeys(t *testing.T) {
	t.Run("rejects every basic key whether it carries a value or an explicit null", func(t *testing.T) {
		for _, name := range ownerBasicRequestKeys {
			for _, value := range []string{`"x"`, `null`} {
				err := rejectPresentOwnerBasicKeys(emr491Keys(t, `{"`+name+`":`+value+`}`))
				require.Error(t, err, name+"="+value)
				assert.True(t, apperrors.HasCode(err, apperrors.CodeOwnerBasicRevisionPending), name)
			}
		}
	})

	t.Run("matches key names case-insensitively like the JSON decoder", func(t *testing.T) {
		err := rejectPresentOwnerBasicKeys(emr491Keys(t, `{"Owner_Name":"x"}`))
		require.Error(t, err)
		assert.True(t, apperrors.HasCode(err, apperrors.CodeOwnerBasicRevisionPending))
	})

	t.Run("lets local-only bodies through, including explicit null on a local key", func(t *testing.T) {
		body := `{"remarks":"memo","is_dangerous":true,"discount_rate":5,"membership_type":"member","dm_preference":null}`
		assert.NoError(t, rejectPresentOwnerBasicKeys(emr491Keys(t, body)))
	})

	t.Run("lets an empty object through", func(t *testing.T) {
		assert.NoError(t, rejectPresentOwnerBasicKeys(emr491Keys(t, `{}`)))
	})
}

func TestUpdateOwner_BasicKeysReturnFixed422BeforeAnyServiceCall(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "basic key with a value", body: `{"owner_name":"山田太郎"}`},
		{name: "basic key with explicit null", body: `{"phone":null}`},
		{name: "basic key mixed with local fields", body: `{"remarks":"memo","email":"a@example.com"}`},
		{name: "key case differs from the wire name", body: `{"Company":"x"}`},
		{name: "basic key mixed with a discount change", body: `{"discount_rate":5,"address1":"x"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockOwnerService{
				getByIDFn: func(context.Context, uint64, uint64) (*model.Owner, error) {
					t.Error("GetByID must not be called for a rejected body")
					return nil, nil
				},
				updateFn: func(context.Context, uint64, uint64, *UpdateOwnerInput) (*model.Owner, error) {
					t.Error("Update must not be called for a rejected body")
					return nil, nil
				},
			}

			w := emr491PatchOwner(t, svc, tt.body)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), `"code":"owner_basic_revision_pending"`)
		})
	}
}

func TestUpdateOwner_LocalOnlyBodyKeepsTheUpdatePath(t *testing.T) {
	var got *UpdateOwnerInput
	svc := &mockOwnerService{
		updateFn: func(_ context.Context, _, _ uint64, input *UpdateOwnerInput) (*model.Owner, error) {
			got = input
			return emr491Owner(), nil
		},
	}

	w := emr491PatchOwner(t, svc, `{"remarks":"memo","is_dangerous":true}`)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.NotNil(t, got)
	require.NotNil(t, got.Remarks)
	assert.Equal(t, "memo", *got.Remarks)
	assert.False(t, updateInputTouchesBasic(got))
}

func TestUpdateInputTouchesBasic(t *testing.T) {
	text := "x"
	assert.False(t, updateInputTouchesBasic(nil))
	assert.False(t, updateInputTouchesBasic(&UpdateOwnerInput{Remarks: &text}))
	assert.True(t, updateInputTouchesBasic(&UpdateOwnerInput{OwnerName: &text}))
	assert.True(t, updateInputTouchesBasic(&UpdateOwnerInput{Phone: &text}))
	assert.True(t, updateInputTouchesBasic(&UpdateOwnerInput{Email: &text}))
	var birthDate *time.Time
	assert.True(t, updateInputTouchesBasic(&UpdateOwnerInput{BirthDate: &birthDate}))
}

func TestOwnerService_Update_BasicFieldsAreRejectedBeforeAnyRead(t *testing.T) {
	phone := "090-0000-0000"
	svc := NewService(nil, nil, nil, nil)

	_, err := svc.Update(context.Background(), 1, 1, &UpdateOwnerInput{Phone: &phone})

	require.Error(t, err)
	assert.True(t, apperrors.HasCode(err, apperrors.CodeOwnerBasicRevisionPending))
}

func TestOwnerRepository_CreateWithPets_FailsClosedBeforeAnyWrite(t *testing.T) {
	// db が nil のため、書き込みどころか DB に触れた時点で panic する。
	repo := NewRepository(nil, nil)

	err := repo.CreateWithPets(context.Background(), &model.Owner{ClinicID: 1}, nil)

	require.Error(t, err)
	assert.True(t, apperrors.HasCode(err, apperrors.CodeOwnerRegistrationPending))
}

func TestCreateOwner_ReturnsRegistrationPending(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithOwnerSvc(NewService(NewRepository(nil, nil), nil, nil, nil))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"owner_name":"山田太郎"}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)

	h.CreateOwner(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), `"code":"owner_registration_pending"`)
}
