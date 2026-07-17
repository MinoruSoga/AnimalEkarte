package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
)

// --- mock: OwnerRepository (line_link 用の最小実装) ---

type mockLineLinkOwnerRepo struct {
	findByIDFn             func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	findAllByLineUserIDFn  func(ctx context.Context, lineUserID string) ([]model.Owner, error)
	updateLineUserIDFn     func(ctx context.Context, clinicID, id uint64, lineUserID *string) error
	updateLineFollowedAtFn func(ctx context.Context, clinicID, id uint64, t time.Time) error
	updateLineBlockedAtFn  func(ctx context.Context, clinicID, id uint64, t time.Time) error
}

func (m *mockLineLinkOwnerRepo) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Owner{ID: id, ClinicID: clinicID}, nil
}
func (m *mockLineLinkOwnerRepo) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLineLinkOwnerRepo) FindAll(_ context.Context, _ []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
	return nil, 0, nil
}
func (m *mockLineLinkOwnerRepo) FindByEmail(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLineLinkOwnerRepo) FindByPhone(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLineLinkOwnerRepo) FindByNameAndPhone(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLineLinkOwnerRepo) FindByLineUserID(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLineLinkOwnerRepo) FindAllWithLineUserID(_ context.Context, _ uint64) ([]model.Owner, error) {
	return nil, nil
}
func (m *mockLineLinkOwnerRepo) FindAllWithLineUserIDCursor(_ context.Context, _, _ uint64, _ int) ([]model.Owner, error) {
	return nil, nil
}
func (m *mockLineLinkOwnerRepo) CreateWithPets(_ context.Context, _ *model.Owner, _ []model.Pet) error {
	return nil
}
func (m *mockLineLinkOwnerRepo) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mockLineLinkOwnerRepo) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockLineLinkOwnerRepo) UpdateLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string) error {
	if m.updateLineUserIDFn != nil {
		return m.updateLineUserIDFn(ctx, clinicID, id, lineUserID)
	}
	return nil
}
func (m *mockLineLinkOwnerRepo) FindAllByLineUserID(ctx context.Context, lineUserID string) ([]model.Owner, error) {
	if m.findAllByLineUserIDFn != nil {
		return m.findAllByLineUserIDFn(ctx, lineUserID)
	}
	return nil, nil
}
func (m *mockLineLinkOwnerRepo) UpdateLineFollowedAt(ctx context.Context, clinicID, id uint64, t time.Time) error {
	if m.updateLineFollowedAtFn != nil {
		return m.updateLineFollowedAtFn(ctx, clinicID, id, t)
	}
	return nil
}
func (m *mockLineLinkOwnerRepo) UpdateLineBlockedAt(ctx context.Context, clinicID, id uint64, t time.Time) error {
	if m.updateLineBlockedAtFn != nil {
		return m.updateLineBlockedAtFn(ctx, clinicID, id, t)
	}
	return nil
}
func (m *mockLineLinkOwnerRepo) CountPetsByOwnerID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockLineLinkOwnerRepo) FindByIDs(_ context.Context, _ uint64, _ []uint64) ([]*model.Owner, error) {
	return nil, nil
}

// --- mock: LineLinkTokenRepository ---

type mockLineLinkTokenRepo struct {
	createFn      func(ctx context.Context, t *model.LineLinkToken) error
	findByTokenFn func(ctx context.Context, token string) (*model.LineLinkToken, error)
	markUsedFn    func(ctx context.Context, id uint64, usedAt time.Time) error
}

func (m *mockLineLinkTokenRepo) Create(ctx context.Context, t *model.LineLinkToken) error {
	if m.createFn != nil {
		return m.createFn(ctx, t)
	}
	return nil
}
func (m *mockLineLinkTokenRepo) FindByToken(ctx context.Context, token string) (*model.LineLinkToken, error) {
	if m.findByTokenFn != nil {
		return m.findByTokenFn(ctx, token)
	}
	return nil, apperrors.WrapNotFound("link_token", token)
}
func (m *mockLineLinkTokenRepo) MarkUsed(ctx context.Context, id uint64, usedAt time.Time) error {
	if m.markUsedFn != nil {
		return m.markUsedFn(ctx, id, usedAt)
	}
	return nil
}

// --- mock: LineReservationSettingRepository ---

type mockLineLinkSettingRepo struct {
	findByClinicIDFn func(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	findAllFn        func(ctx context.Context) ([]model.LineReservationSetting, error)
}

func (m *mockLineLinkSettingRepo) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.LineReservationSetting{ClinicID: clinicID, LiffID: "test-liff-id", LineChannelSecret: "secret", LineChannelID: "channel-id"}, nil
}
func (m *mockLineLinkSettingRepo) FindAll(ctx context.Context) ([]model.LineReservationSetting, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return []model.LineReservationSetting{{ClinicID: 1, LineChannelSecret: "secret"}}, nil
}
func (m *mockLineLinkSettingRepo) Save(_ context.Context, _ uint64, _ *model.LineReservationSetting) error {
	return nil
}

func newTestLineLinkService(
	ownerRepo *mockLineLinkOwnerRepo,
	tokenRepo *mockLineLinkTokenRepo,
	settingRepo *mockLineLinkSettingRepo,
) LineLinkService {
	return NewLineLinkService(ownerRepo, tokenRepo, settingRepo, &mockAuditService{}, nil)
}

// --- GenerateLinkToken tests ---

func TestLineLinkService_GenerateLinkToken_Success(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)
	result, err := svc.GenerateLinkToken(context.Background(), 1, 42)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.True(t, result.ExpiresAt.After(time.Now()))
	assert.Contains(t, result.LiffURL, "liff.line.me")
	assert.Contains(t, result.LiffURL, result.Token)
	assert.Contains(t, result.LiffURL, "clinic_id=1", "SD-14: LiffLinkPage requires clinic_id query param or it fails immediately")
}

func TestLineLinkService_GenerateLinkToken_OwnerNotFound(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "42")
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)
	_, err := svc.GenerateLinkToken(context.Background(), 1, 42)
	assert.Error(t, err)
}

func TestLineLinkService_GenerateLinkToken_NoLiffID(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				return &model.LineReservationSetting{ClinicID: clinicID, LiffID: ""}, nil
			},
		},
	)
	result, err := svc.GenerateLinkToken(context.Background(), 1, 42)
	require.NoError(t, err)
	assert.Empty(t, result.LiffURL)
}

// --- HandleWebhook tests ---

func makeLineSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestLineLinkService_HandleWebhook_InvalidSignature(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)
	body := []byte(`{"destination":"U123","events":[]}`)
	err := svc.HandleWebhook(context.Background(), body, "invalidsig")
	assert.Error(t, err)
}

func TestLineLinkService_HandleWebhook_FollowEvent(t *testing.T) {
	lineUserID := "Uabc123"
	followedAt := false

	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{
			findAllByLineUserIDFn: func(_ context.Context, uid string) ([]model.Owner, error) {
				return []model.Owner{{ID: 10, ClinicID: 1, LineUserID: &uid}}, nil
			},
			updateLineFollowedAtFn: func(_ context.Context, _, _ uint64, _ time.Time) error {
				followedAt = true
				return nil
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "follow", Source: struct {
				UserID string `json:"userId"`
			}{UserID: lineUserID}},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeLineSignature(body, "secret")

	err := svc.HandleWebhook(context.Background(), body, sig)
	require.NoError(t, err)
	assert.True(t, followedAt)
}

// TestLineLinkService_HandleWebhook_EncryptedSecret は DB 上の line_channel_secret が
// 暗号化されていても、復号後の平文で署名検証が成功することを確認する（H-4）。
func TestLineLinkService_HandleWebhook_EncryptedSecret(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	encSecret, err := cipher.Encrypt("secret")
	require.NoError(t, err)

	lineUserID := "Uabc123"
	followedAt := false
	ownerRepo := &mockLineLinkOwnerRepo{
		findAllByLineUserIDFn: func(_ context.Context, uid string) ([]model.Owner, error) {
			return []model.Owner{{ID: 10, ClinicID: 1, LineUserID: &uid}}, nil
		},
		updateLineFollowedAtFn: func(_ context.Context, _, _ uint64, _ time.Time) error {
			followedAt = true
			return nil
		},
	}
	settingRepo := &mockLineLinkSettingRepo{
		findAllFn: func(_ context.Context) ([]model.LineReservationSetting, error) {
			return []model.LineReservationSetting{{ClinicID: 1, LineChannelSecret: encSecret}}, nil
		},
	}
	svc := NewLineLinkService(ownerRepo, &mockLineLinkTokenRepo{}, settingRepo, &mockAuditService{}, cipher)

	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "follow", Source: struct {
				UserID string `json:"userId"`
			}{UserID: lineUserID}},
		},
	}
	body, _ := json.Marshal(payload)
	// 署名は平文シークレットで計算する（LINE 側の実挙動）。
	sig := makeLineSignature(body, "secret")

	err = svc.HandleWebhook(context.Background(), body, sig)
	require.NoError(t, err)
	assert.True(t, followedAt)
}

func TestLineLinkService_HandleWebhook_UnfollowEvent(t *testing.T) {
	lineUserID := "Uabc123"
	blockedAt := false

	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{
			findAllByLineUserIDFn: func(_ context.Context, uid string) ([]model.Owner, error) {
				return []model.Owner{{ID: 10, ClinicID: 1, LineUserID: &uid}}, nil
			},
			updateLineBlockedAtFn: func(_ context.Context, _, _ uint64, _ time.Time) error {
				blockedAt = true
				return nil
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "unfollow", Source: struct {
				UserID string `json:"userId"`
			}{UserID: lineUserID}},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeLineSignature(body, "secret")

	err := svc.HandleWebhook(context.Background(), body, sig)
	require.NoError(t, err)
	assert.True(t, blockedAt)
}

// --- LinkAccount tests ---
// verifyLineIDToken は https://api.line.me/... を直接呼び出す設計のため、
// ユニットテストでは LINE IDトークン検証失敗ケース（= Unauthorized）のみ確認する。

func TestLineLinkService_LinkAccount_InvalidLineToken(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)
	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{
		LinkToken:   "any-token",
		LineIDToken: "invalid-line-token",
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
}

// --- verifyLineSignature tests ---

func TestVerifyLineSignature_Valid(t *testing.T) {
	body := []byte(`{"events":[]}`)
	secret := "mysecret"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	assert.True(t, verifyLineSignature(body, sig, secret))
}

func TestVerifyLineSignature_Invalid(t *testing.T) {
	assert.False(t, verifyLineSignature([]byte("body"), "badsig", "secret"))
}

// --- HTTP round-trip test helpers (verifyLineIDToken seam) ---

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonRespClient(statusCode int, body string) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}
}

type errReadCloser struct{}

func (errReadCloser) Read(_ []byte) (int, error) { return 0, errors.New("read error") }
func (errReadCloser) Close() error               { return nil }

func newTestLineLinkServiceFull(
	ownerRepo *mockLineLinkOwnerRepo,
	tokenRepo *mockLineLinkTokenRepo,
	settingRepo *mockLineLinkSettingRepo,
	auditSvc *mockAuditService,
	client *http.Client,
) *lineLinkService {
	return &lineLinkService{
		ownerRepo:         ownerRepo,
		lineLinkTokenRepo: tokenRepo,
		lineSettingRepo:   settingRepo,
		auditSvc:          auditSvc,
		httpClient:        client,
	}
}

// --- verifyLineIDToken tests ---

func TestVerifyLineIDToken(t *testing.T) {
	tests := []struct {
		name        string
		settingRepo *mockLineLinkSettingRepo
		client      *http.Client
		wantErr     bool
		wantSub     string
	}{
		{
			name:    "returns line user id on success",
			client:  jsonRespClient(http.StatusOK, `{"sub":"Uabc123"}`),
			wantSub: "Uabc123",
		},
		{
			name: "returns error when setting lookup fails",
			settingRepo: &mockLineLinkSettingRepo{
				findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
		{
			name: "returns error when http request fails",
			client: &http.Client{
				Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
					return nil, errors.New("network error")
				}),
			},
			wantErr: true,
		},
		{
			name: "returns error when response body cannot be read",
			client: &http.Client{
				Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Body: errReadCloser{}, Header: make(http.Header)}, nil
				}),
			},
			wantErr: true,
		},
		{
			name:    "returns invalid input when status is not 200",
			client:  jsonRespClient(http.StatusUnauthorized, `{"error":"invalid_request"}`),
			wantErr: true,
		},
		{
			name:    "returns error when response body is invalid json",
			client:  jsonRespClient(http.StatusOK, `not-json`),
			wantErr: true,
		},
		{
			name:    "returns invalid input when sub is empty",
			client:  jsonRespClient(http.StatusOK, `{"sub":""}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settingRepo := tt.settingRepo
			if settingRepo == nil {
				settingRepo = &mockLineLinkSettingRepo{}
			}
			sub, err := verifyLineIDToken(context.Background(), "test-token", 1, settingRepo, tt.client)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantSub, sub)
			}
		})
	}
}

// --- LinkAccount additional branch tests ---

func TestLineLinkService_LinkAccount_TokenNotFound(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{
			findByTokenFn: func(_ context.Context, _ string) (*model.LineLinkToken, error) {
				return nil, errors.New("not found")
			},
		},
		&mockLineLinkSettingRepo{},
		&mockAuditService{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "bad-token", LineIDToken: "valid"})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestLineLinkService_LinkAccount_ClinicMismatch(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{
			findByTokenFn: func(_ context.Context, _ string) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 2, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		&mockAuditService{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestLineLinkService_LinkAccount_OwnerNotFound(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLineLinkOwnerRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, errors.New("not found")
			},
		},
		&mockLineLinkTokenRepo{
			findByTokenFn: func(_ context.Context, _ string) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		&mockAuditService{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
}

func TestLineLinkService_LinkAccount_AlreadyLinkedConflict(t *testing.T) {
	existing := "Uold"
	svc := newTestLineLinkServiceFull(
		&mockLineLinkOwnerRepo{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return &model.Owner{ID: id, LineUserID: &existing}, nil
			},
		},
		&mockLineLinkTokenRepo{
			findByTokenFn: func(_ context.Context, _ string) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		&mockAuditService{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid", Force: false})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
}

func TestLineLinkService_LinkAccount_ForceOverridesExisting(t *testing.T) {
	existing := "Uold"
	updateCalled := false
	svc := newTestLineLinkServiceFull(
		&mockLineLinkOwnerRepo{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return &model.Owner{ID: id, LineUserID: &existing}, nil
			},
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, lineUserID *string) error {
				updateCalled = true
				require.NotNil(t, lineUserID)
				assert.Equal(t, "Uverified123", *lineUserID)
				return nil
			},
		},
		&mockLineLinkTokenRepo{
			findByTokenFn: func(_ context.Context, _ string) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		&mockAuditService{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	owner, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid", Force: true})

	require.NoError(t, err)
	require.NotNil(t, owner)
	assert.True(t, updateCalled)
	require.NotNil(t, owner.LineUserID)
	assert.Equal(t, "Uverified123", *owner.LineUserID)
}

func TestLineLinkService_LinkAccount_UpdateLineUserIDError(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLineLinkOwnerRepo{
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
				return errors.New("db error")
			},
		},
		&mockLineLinkTokenRepo{
			findByTokenFn: func(_ context.Context, _ string) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		&mockAuditService{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
}

func TestLineLinkService_LinkAccount_MarkUsedError(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{
			findByTokenFn: func(_ context.Context, _ string) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
			markUsedFn: func(_ context.Context, _ uint64, _ time.Time) error {
				return errors.New("db error")
			},
		},
		&mockLineLinkSettingRepo{},
		&mockAuditService{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
}

func TestLineLinkService_LinkAccount_AuditLogFailureStillSucceeds(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{
			findByTokenFn: func(_ context.Context, _ string) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		&mockAuditService{
			logLstepOperationFn: func(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
				return errors.New("audit failure")
			},
		},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	owner, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.NoError(t, err, "audit log failure must not fail the overall link operation")
	require.NotNil(t, owner)
}

func TestLineLinkService_LinkAccount_Success(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{
			findByTokenFn: func(_ context.Context, _ string) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		&mockAuditService{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	owner, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.NoError(t, err)
	require.NotNil(t, owner)
	require.NotNil(t, owner.LineUserID)
	assert.Equal(t, "Uverified123", *owner.LineUserID)
}

// --- GenerateLinkToken additional branch tests ---

func TestLineLinkService_GenerateLinkToken_CreateError(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{
			createFn: func(_ context.Context, _ *model.LineLinkToken) error {
				return errors.New("db error")
			},
		},
		&mockLineLinkSettingRepo{},
	)
	_, err := svc.GenerateLinkToken(context.Background(), 1, 42)
	assert.Error(t, err)
}

func TestLineLinkService_GenerateLinkToken_SettingRepoError(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("db error")
			},
		},
	)
	_, err := svc.GenerateLinkToken(context.Background(), 1, 42)
	assert.Error(t, err)
}

// --- HandleWebhook additional branch tests ---

func TestLineLinkService_HandleWebhook_InvalidBody(t *testing.T) {
	svc := newTestLineLinkService(&mockLineLinkOwnerRepo{}, &mockLineLinkTokenRepo{}, &mockLineLinkSettingRepo{})
	body := []byte(`not-json`)
	sig := makeLineSignature(body, "secret")

	err := svc.HandleWebhook(context.Background(), body, sig)

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestLineLinkService_HandleWebhook_EmptyUserIDSkipped(t *testing.T) {
	handlerCalled := false
	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{
			findAllByLineUserIDFn: func(_ context.Context, _ string) ([]model.Owner, error) {
				handlerCalled = true
				return nil, nil
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Events: []WebhookEvent{
			{Type: "follow", Source: struct {
				UserID string `json:"userId"`
			}{UserID: ""}},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeLineSignature(body, "secret")

	err := svc.HandleWebhook(context.Background(), body, sig)

	require.NoError(t, err)
	assert.False(t, handlerCalled, "empty line_user_id events must be skipped")
}

func TestLineLinkService_HandleWebhook_FollowEvent_FindOwnersErrorIsLoggedNotReturned(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{
			findAllByLineUserIDFn: func(_ context.Context, _ string) ([]model.Owner, error) {
				return nil, errors.New("db error")
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Events: []WebhookEvent{
			{Type: "follow", Source: struct {
				UserID string `json:"userId"`
			}{UserID: "Uabc"}},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeLineSignature(body, "secret")

	err := svc.HandleWebhook(context.Background(), body, sig)

	assert.NoError(t, err, "handleFollowEvent errors must be logged, not propagated")
}

func TestLineLinkService_HandleWebhook_UnfollowEvent_FindOwnersErrorIsLoggedNotReturned(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLineLinkOwnerRepo{
			findAllByLineUserIDFn: func(_ context.Context, _ string) ([]model.Owner, error) {
				return nil, errors.New("db error")
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Events: []WebhookEvent{
			{Type: "unfollow", Source: struct {
				UserID string `json:"userId"`
			}{UserID: "Uabc"}},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeLineSignature(body, "secret")

	err := svc.HandleWebhook(context.Background(), body, sig)

	assert.NoError(t, err, "handleUnfollowEvent errors must be logged, not propagated")
}

// --- handleFollowEvent / handleUnfollowEvent direct tests ---

func TestHandleFollowEvent_UpdateErrorIsLoggedNotReturned(t *testing.T) {
	svc := &lineLinkService{
		ownerRepo: &mockLineLinkOwnerRepo{
			findAllByLineUserIDFn: func(_ context.Context, _ string) ([]model.Owner, error) {
				return []model.Owner{{ID: 1, ClinicID: 1}}, nil
			},
			updateLineFollowedAtFn: func(_ context.Context, _, _ uint64, _ time.Time) error {
				return errors.New("db error")
			},
		},
	}

	err := svc.handleFollowEvent(context.Background(), "Uabc", time.Now())

	assert.NoError(t, err)
}

func TestHandleUnfollowEvent_UpdateErrorIsLoggedNotReturned(t *testing.T) {
	svc := &lineLinkService{
		ownerRepo: &mockLineLinkOwnerRepo{
			findAllByLineUserIDFn: func(_ context.Context, _ string) ([]model.Owner, error) {
				return []model.Owner{{ID: 1, ClinicID: 1}}, nil
			},
			updateLineBlockedAtFn: func(_ context.Context, _, _ uint64, _ time.Time) error {
				return errors.New("db error")
			},
		},
	}

	err := svc.handleUnfollowEvent(context.Background(), "Uabc", time.Now())

	assert.NoError(t, err)
}

// --- verifySignatureAnyClinic additional branch tests ---

func TestVerifySignatureAnyClinic_FindAllError(t *testing.T) {
	svc := &lineLinkService{
		lineSettingRepo: &mockLineLinkSettingRepo{
			findAllFn: func(_ context.Context) ([]model.LineReservationSetting, error) {
				return nil, errors.New("db error")
			},
		},
	}

	got := svc.verifySignatureAnyClinic(context.Background(), []byte("body"), "sig")

	assert.False(t, got)
}

func TestVerifySignatureAnyClinic_EmptySecretSkipped(t *testing.T) {
	svc := &lineLinkService{
		lineSettingRepo: &mockLineLinkSettingRepo{
			findAllFn: func(_ context.Context) ([]model.LineReservationSetting, error) {
				return []model.LineReservationSetting{
					{ClinicID: 1, LineChannelSecret: ""},
					{ClinicID: 2, LineChannelSecret: "secret"},
				}, nil
			},
		},
	}
	body := []byte("body")
	sig := makeLineSignature(body, "secret")

	got := svc.verifySignatureAnyClinic(context.Background(), body, sig)

	assert.True(t, got, "should skip the empty-secret clinic and match the second")
}

func TestVerifySignatureAnyClinic_NoMatch(t *testing.T) {
	svc := &lineLinkService{
		lineSettingRepo: &mockLineLinkSettingRepo{
			findAllFn: func(_ context.Context) ([]model.LineReservationSetting, error) {
				return []model.LineReservationSetting{{ClinicID: 1, LineChannelSecret: "secret"}}, nil
			},
		},
	}
	got := svc.verifySignatureAnyClinic(context.Background(), []byte("body"), "wrong-signature")
	assert.False(t, got)
}
