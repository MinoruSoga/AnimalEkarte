package lstep

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
)

// --- mock: OwnerRepository (line_link 用の最小実装) ---

// mockLstepOwnerRepo は lstepOwnerRepo view（5メソッド）の mock。
// line_link/line_send/line_customer の 3 service が同一 view を共有する。
type mockLstepOwnerRepo struct {
	findByIDFn             func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	lockLineLinkOwnerFn    func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	findByLineUserIDFn     func(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error)
	updateLineUserIDFn     func(ctx context.Context, clinicID, id uint64, lineUserID *string) error
	updateLineFollowedAtFn func(ctx context.Context, clinicID, id uint64, expectedLineUserID string, t time.Time) (bool, error)
	updateLineBlockedAtFn  func(ctx context.Context, clinicID, id uint64, expectedLineUserID string, t time.Time) (bool, error)
}

func (m *mockLstepOwnerRepo) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Owner{ID: id, ClinicID: clinicID}, nil
}
func (m *mockLstepOwnerRepo) LockLineLinkOwner(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	if m.lockLineLinkOwnerFn != nil {
		return m.lockLineLinkOwnerFn(ctx, clinicID, id)
	}
	return &model.Owner{ID: id, ClinicID: clinicID}, nil
}
func (m *mockLstepOwnerRepo) FindByLineUserID(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error) {
	if m.findByLineUserIDFn != nil {
		return m.findByLineUserIDFn(ctx, clinicID, lineUserID)
	}
	return nil, apperrors.WrapNotFound("owner", "")
}
func (m *mockLstepOwnerRepo) UpdateLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string) error {
	if m.updateLineUserIDFn != nil {
		return m.updateLineUserIDFn(ctx, clinicID, id, lineUserID)
	}
	return nil
}
func (m *mockLstepOwnerRepo) UpdateLineFollowedAt(
	ctx context.Context,
	clinicID, id uint64,
	expectedLineUserID string,
	t time.Time,
) (bool, error) {
	if m.updateLineFollowedAtFn != nil {
		return m.updateLineFollowedAtFn(ctx, clinicID, id, expectedLineUserID, t)
	}
	return true, nil
}
func (m *mockLstepOwnerRepo) UpdateLineBlockedAt(
	ctx context.Context,
	clinicID, id uint64,
	expectedLineUserID string,
	t time.Time,
) (bool, error) {
	if m.updateLineBlockedAtFn != nil {
		return m.updateLineBlockedAtFn(ctx, clinicID, id, expectedLineUserID, t)
	}
	return true, nil
}

// --- mock: LineLinkTokenRepository ---

type mockLineLinkTokenRepo struct {
	createFn          func(ctx context.Context, t *model.LineLinkToken) error
	lockUsableTokenFn func(ctx context.Context, rawToken string, now time.Time) (*model.LineLinkToken, error)
	consumeFn         func(ctx context.Context, id uint64, usedAt time.Time) error
}

func (m *mockLineLinkTokenRepo) Create(ctx context.Context, t *model.LineLinkToken) error {
	if m.createFn != nil {
		return m.createFn(ctx, t)
	}
	return nil
}
func (m *mockLineLinkTokenRepo) LockUsableByRawToken(ctx context.Context, rawToken string, now time.Time) (*model.LineLinkToken, error) {
	if m.lockUsableTokenFn != nil {
		return m.lockUsableTokenFn(ctx, rawToken, now)
	}
	return nil, apperrors.WrapNotFound("link_token", "")
}
func (m *mockLineLinkTokenRepo) Consume(ctx context.Context, id uint64, usedAt time.Time) error {
	if m.consumeFn != nil {
		return m.consumeFn(ctx, id, usedAt)
	}
	return nil
}

type immediateLineLinkTransactor struct{}

func (immediateLineLinkTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(context.WithValue(ctx, lineLinkTxTestContextKey{}, true))
}

type lineLinkTxTestContextKey struct{}

type rollbackLineLinkTransactor struct {
	snapshot func() any
	restore  func(any)
}

func (t rollbackLineLinkTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	before := t.snapshot()
	err := fn(context.WithValue(ctx, lineLinkTxTestContextKey{}, true))
	if err != nil {
		t.restore(before)
	}
	return err
}

type mockLineLinkAuditTxLogger struct {
	logFn func(ctx context.Context, clinicID, ownerID uint64) error
}

func (m *mockLineLinkAuditTxLogger) LogOwnerLineLinkTx(ctx context.Context, clinicID, ownerID uint64) error {
	if m.logFn != nil {
		return m.logFn(ctx, clinicID, ownerID)
	}
	return nil
}

// --- mock: LineReservationSettingRepository ---

type mockLineLinkSettingRepo struct {
	findByClinicIDFn      func(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	findByLineBotUserIDFn func(ctx context.Context, lineBotUserID string) (*model.LineReservationSetting, error)
	findWebhookRouteFn    func(ctx context.Context, lineBotUserID string) (uint64, bool, error)
	findAllFn             func(ctx context.Context) ([]model.LineReservationSetting, error)

	findAllCalls          int
	findWebhookRouteCalls int
}

func (m *mockLineLinkSettingRepo) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return &model.LineReservationSetting{ClinicID: clinicID, LiffID: "test-liff-id", LineChannelSecret: "secret", LineChannelID: "channel-id"}, nil
}
func (m *mockLineLinkSettingRepo) FindByLineBotUserID(ctx context.Context, lineBotUserID string) (*model.LineReservationSetting, error) {
	if m.findByLineBotUserIDFn != nil {
		return m.findByLineBotUserIDFn(ctx, lineBotUserID)
	}
	// Default: map any non-empty destination to clinic 1 with plaintext secret "secret".
	return &model.LineReservationSetting{
		ID:                1,
		ClinicID:          1,
		LineBotUserID:     lineBotUserID,
		LineChannelSecret: "secret",
	}, nil
}
func (m *mockLineLinkSettingRepo) FindWebhookRouteByLineBotUserID(ctx context.Context, lineBotUserID string) (uint64, bool, error) {
	m.findWebhookRouteCalls++
	if m.findWebhookRouteFn != nil {
		return m.findWebhookRouteFn(ctx, lineBotUserID)
	}
	setting, err := m.FindByLineBotUserID(ctx, lineBotUserID)
	if err != nil || setting == nil {
		return 0, false, err
	}
	return setting.ClinicID, false, nil
}
func (m *mockLineLinkSettingRepo) FindAll(ctx context.Context) ([]model.LineReservationSetting, error) {
	m.findAllCalls++
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return []model.LineReservationSetting{{ClinicID: 1, LineChannelSecret: "secret"}}, nil
}
func (m *mockLineLinkSettingRepo) Save(_ context.Context, _ uint64, _ *model.LineReservationSetting) error {
	return nil
}

type mockLineChannelCredentialRepo struct {
	findByClinicServiceKeyFn func(ctx context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error)
	findCalls                int
}

func (m *mockLineChannelCredentialRepo) FindCredentialByClinicServiceKey(
	ctx context.Context,
	clinicID uint64,
	service, keyName string,
) (*model.ClinicIntegration, error) {
	m.findCalls++
	if m.findByClinicServiceKeyFn != nil {
		return m.findByClinicServiceKeyFn(ctx, clinicID, service, keyName)
	}
	// Default secret for webhook HMAC tests. Liff ID is absent unless a test
	// supplies it — avoids accidental liff_url built from the channel secret.
	if keyName == model.IntegrationKeyLiffID {
		return nil, apperrors.WrapNotFound("clinic_integration", keyName)
	}
	return &model.ClinicIntegration{
		ID:       clinicID,
		ClinicID: clinicID,
		Service:  service,
		KeyName:  keyName,
		KeyValue: "secret",
	}, nil
}

func newTestLineLinkService(
	ownerRepo *mockLstepOwnerRepo,
	tokenRepo *mockLineLinkTokenRepo,
	settingRepo *mockLineLinkSettingRepo,
	credentialRepos ...*mockLineChannelCredentialRepo,
) LineLinkService {
	plaintextCredentialRepo := &mockLineChannelCredentialRepo{}
	if len(credentialRepos) > 0 && credentialRepos[0] != nil {
		plaintextCredentialRepo = credentialRepos[0]
	}
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	if err != nil {
		panic(err)
	}
	encryptedCredentialRepo := &mockLineChannelCredentialRepo{
		findByClinicServiceKeyFn: func(
			ctx context.Context,
			clinicID uint64,
			service, keyName string,
		) (*model.ClinicIntegration, error) {
			credential, findErr := plaintextCredentialRepo.FindCredentialByClinicServiceKey(
				ctx,
				clinicID,
				service,
				keyName,
			)
			if findErr != nil || credential == nil || credential.KeyValue == "" {
				return credential, findErr
			}
			// Match production storage: only secret-shaped keys are ciphertext at rest.
			// liff_id stays plaintext so GenerateLinkToken can read it directly.
			if !model.IsEncryptedKey(keyName) {
				return credential, nil
			}
			ciphertext, encryptErr := cipher.Encrypt(credential.KeyValue)
			if encryptErr != nil {
				return nil, encryptErr
			}
			encryptedCredential := *credential
			encryptedCredential.KeyValue = ciphertext
			return &encryptedCredential, nil
		},
	}
	return NewLineLinkService(
		ownerRepo,
		tokenRepo,
		settingRepo,
		encryptedCredentialRepo,
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{},
		cipher,
	)
}

// --- GenerateLinkToken tests ---

func TestLineLinkService_GenerateLinkToken_Success(t *testing.T) {
	var storedToken string
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{
			createFn: func(_ context.Context, token *model.LineLinkToken) error {
				storedToken = token.Token
				return nil
			},
		},
		&mockLineLinkSettingRepo{},
	)
	result, err := svc.GenerateLinkToken(context.Background(), 1, 42)
	require.NoError(t, err)
	assert.Len(t, result.Token, 43, "raw token must use unpadded base64url encoding")
	assert.Len(t, storedToken, 43, "persisted SHA-256 base64url digest must fit varchar(64)")
	assert.Equal(t, digestLineLinkToken(result.Token), storedToken)
	assert.NotEqual(t, result.Token, storedToken, "raw bearer token must never be persisted")
	assert.True(t, result.ExpiresAt.After(time.Now()))
	assert.Contains(t, result.LiffURL, "liff.line.me")
	assert.Contains(t, result.LiffURL, result.Token)
	assert.Contains(t, result.LiffURL, "clinic_id=1", "SD-14: LiffLinkPage requires clinic_id query param or it fails immediately")
}

func TestLineLinkService_GenerateLinkToken_OwnerNotFound(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{
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
		&mockLstepOwnerRepo{},
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

// BUG-504: missing line_reservation_settings must not surface as opaque owner 404.
// Staff UI already shows the issue button for an existing unlinked owner.
func TestLineLinkService_GenerateLinkToken_ReservationSettingNotFoundStillIssuesToken(t *testing.T) {
	var created bool
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{
			createFn: func(_ context.Context, _ *model.LineLinkToken) error {
				created = true
				return nil
			},
		},
		&mockLineLinkSettingRepo{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("line_reservation_setting", "clinic")
			},
		},
	)
	result, err := svc.GenerateLinkToken(context.Background(), 1, 42)
	require.NoError(t, err)
	assert.True(t, created)
	assert.NotEmpty(t, result.Token)
	assert.Empty(t, result.LiffURL)
}

// BUG-504: LIFF ID configured under L-step clinic_integrations must build liff_url
// when line_reservation_settings row is absent or has empty liff_id.
func TestLineLinkService_GenerateLinkToken_FallsBackToLstepLiffID(t *testing.T) {
	const lstepLiffID = "1234567890-lstepLiff"
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, apperrors.WrapNotFound("line_reservation_setting", "clinic")
			},
		},
		&mockLineChannelCredentialRepo{
			findByClinicServiceKeyFn: func(
				_ context.Context,
				clinicID uint64,
				service, keyName string,
			) (*model.ClinicIntegration, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, model.IntegrationServiceLstep, service)
				assert.Equal(t, model.IntegrationKeyLiffID, keyName)
				return &model.ClinicIntegration{
					ClinicID: clinicID,
					Service:  service,
					KeyName:  keyName,
					KeyValue: lstepLiffID,
				}, nil
			},
		},
	)
	result, err := svc.GenerateLinkToken(context.Background(), 1, 42)
	require.NoError(t, err)
	assert.Contains(t, result.LiffURL, "liff.line.me/"+lstepLiffID)
	assert.Contains(t, result.LiffURL, result.Token)
	assert.Contains(t, result.LiffURL, "clinic_id=1")
}

func TestLineLinkService_GenerateLinkToken_PrefersReservationLiffIDOverLstep(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{
			findByClinicIDFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				return &model.LineReservationSetting{ClinicID: clinicID, LiffID: "reserve-liff"}, nil
			},
		},
		&mockLineChannelCredentialRepo{
			findByClinicServiceKeyFn: func(
				_ context.Context,
				_ uint64,
				_, _ string,
			) (*model.ClinicIntegration, error) {
				t.Fatal("must not consult L-step LiffID when reservation LiffID is present")
				return nil, nil
			},
		},
	)
	result, err := svc.GenerateLinkToken(context.Background(), 1, 42)
	require.NoError(t, err)
	assert.Contains(t, result.LiffURL, "liff.line.me/reserve-liff")
}

// --- HandleWebhook tests ---

func makeLineSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

const testLineWebhookTimestamp int64 = 1_700_000_000_000

func TestLineLinkService_HandleWebhook_InvalidSignature(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{},
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
		&mockLstepOwnerRepo{
			findByLineUserIDFn: func(_ context.Context, clinicID uint64, uid string) (*model.Owner, error) {
				return &model.Owner{ID: 10, ClinicID: clinicID, LineUserID: &uid}, nil
			},
			updateLineFollowedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				followedAt = true
				return true, nil
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "follow", Timestamp: testLineWebhookTimestamp, Source: struct {
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

// TestLineLinkService_HandleWebhook_UnsupportedEventType_NoSideEffects は follow/unfollow 以外の
// event type を business side effect なしで skip することを固定する（LINE residual FINAL R-01）。
func TestLineLinkService_HandleWebhook_UnsupportedEventType_NoSideEffects(t *testing.T) {
	ownerMutated := false
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{
			findByLineUserIDFn: func(_ context.Context, clinicID uint64, uid string) (*model.Owner, error) {
				ownerMutated = true
				return &model.Owner{ID: 10, ClinicID: clinicID, LineUserID: &uid}, nil
			},
			updateLineFollowedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				ownerMutated = true
				return true, nil
			},
			updateLineBlockedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				ownerMutated = true
				return true, nil
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "message", Timestamp: testLineWebhookTimestamp, Source: struct {
				UserID string `json:"userId"`
			}{UserID: "Uabc123"}},
			{Type: "postback", Timestamp: testLineWebhookTimestamp, Source: struct {
				UserID string `json:"userId"`
			}{UserID: "Uabc123"}},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeLineSignature(body, "secret")

	err := svc.HandleWebhook(context.Background(), body, sig)
	require.NoError(t, err)
	assert.False(t, ownerMutated, "unsupported event types must not touch owner state")
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
	ownerRepo := &mockLstepOwnerRepo{
		findByLineUserIDFn: func(_ context.Context, clinicID uint64, uid string) (*model.Owner, error) {
			return &model.Owner{ID: 10, ClinicID: clinicID, LineUserID: &uid}, nil
		},
		updateLineFollowedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
			followedAt = true
			return true, nil
		},
	}
	settingRepo := &mockLineLinkSettingRepo{
		findByLineBotUserIDFn: func(_ context.Context, lineBotUserID string) (*model.LineReservationSetting, error) {
			return &model.LineReservationSetting{
				ClinicID:          1,
				LineBotUserID:     lineBotUserID,
				LineChannelSecret: encSecret,
			}, nil
		},
	}
	credentialRepo := &mockLineChannelCredentialRepo{
		findByClinicServiceKeyFn: func(_ context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error) {
			return &model.ClinicIntegration{
				ID: clinicID, ClinicID: clinicID, Service: service, KeyName: keyName, KeyValue: encSecret,
			}, nil
		},
	}
	svc := NewLineLinkService(
		ownerRepo,
		&mockLineLinkTokenRepo{},
		settingRepo,
		credentialRepo,
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{},
		cipher,
	)

	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "follow", Timestamp: testLineWebhookTimestamp, Source: struct {
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
		&mockLstepOwnerRepo{
			findByLineUserIDFn: func(_ context.Context, clinicID uint64, uid string) (*model.Owner, error) {
				return &model.Owner{ID: 10, ClinicID: clinicID, LineUserID: &uid}, nil
			},
			updateLineBlockedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				blockedAt = true
				return true, nil
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "unfollow", Timestamp: testLineWebhookTimestamp, Source: struct {
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

func TestLineLinkService_HandleWebhook_ScopesOwnerMutationToSigningClinic(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{name: "follow only updates the signing clinic owner", eventType: "follow"},
		{name: "unfollow only updates the signing clinic owner", eventType: "unfollow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const signingClinicID uint64 = 2
			lineUserID := "U-shared-across-clinics"
			var scopedLookupClinicIDs []uint64
			var updatedClinicIDs []uint64
			var expectedLineUserIDs []string
			var updateTimes []time.Time
			ownerRepo := &mockLstepOwnerRepo{
				findByLineUserIDFn: func(_ context.Context, clinicID uint64, uid string) (*model.Owner, error) {
					scopedLookupClinicIDs = append(scopedLookupClinicIDs, clinicID)
					return &model.Owner{ID: 20, ClinicID: clinicID, LineUserID: &uid}, nil
				},
				updateLineFollowedAtFn: func(_ context.Context, clinicID, _ uint64, expectedLineUserID string, eventAt time.Time) (bool, error) {
					updatedClinicIDs = append(updatedClinicIDs, clinicID)
					expectedLineUserIDs = append(expectedLineUserIDs, expectedLineUserID)
					updateTimes = append(updateTimes, eventAt)
					return true, nil
				},
				updateLineBlockedAtFn: func(_ context.Context, clinicID, _ uint64, expectedLineUserID string, eventAt time.Time) (bool, error) {
					updatedClinicIDs = append(updatedClinicIDs, clinicID)
					expectedLineUserIDs = append(expectedLineUserIDs, expectedLineUserID)
					updateTimes = append(updateTimes, eventAt)
					return true, nil
				},
			}
			settingRepo := &mockLineLinkSettingRepo{
				findByLineBotUserIDFn: func(_ context.Context, lineBotUserID string) (*model.LineReservationSetting, error) {
					switch lineBotUserID {
					case "bot-clinic-1":
						return &model.LineReservationSetting{
							ClinicID: 1, LineBotUserID: lineBotUserID, LineChannelSecret: "clinic-one-secret",
						}, nil
					case "bot-clinic-2":
						return &model.LineReservationSetting{
							ClinicID: signingClinicID, LineBotUserID: lineBotUserID, LineChannelSecret: "clinic-two-secret",
						}, nil
					default:
						return nil, apperrors.WrapNotFound("line_reservation_setting", lineBotUserID)
					}
				},
			}
			credentialRepo := &mockLineChannelCredentialRepo{
				findByClinicServiceKeyFn: func(_ context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error) {
					return &model.ClinicIntegration{
						ID: clinicID, ClinicID: clinicID, Service: service, KeyName: keyName, KeyValue: "clinic-two-secret",
					}, nil
				},
			}
			svc := newTestLineLinkService(ownerRepo, &mockLineLinkTokenRepo{}, settingRepo, credentialRepo)
			payload := WebhookPayload{
				Destination: "bot-clinic-2",
				Events: []WebhookEvent{
					{Type: tt.eventType, Timestamp: testLineWebhookTimestamp, Source: struct {
						UserID string `json:"userId"`
					}{UserID: lineUserID}},
				},
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			err = svc.HandleWebhook(
				context.Background(),
				body,
				makeLineSignature(body, "clinic-two-secret"),
			)

			require.NoError(t, err)
			assert.Equal(t, []uint64{signingClinicID}, scopedLookupClinicIDs)
			assert.Equal(t, []uint64{signingClinicID}, updatedClinicIDs)
			assert.Equal(t, []string{lineUserID}, expectedLineUserIDs)
			assert.Equal(t, []time.Time{time.UnixMilli(testLineWebhookTimestamp)}, updateTimes)
		})
	}
}

// TestLineLinkService_HandleWebhook_RejectsForeignDestinationSecret は
// destination が指す clinic 以外の secret で署名しても通さないことを保証する。
// （旧: 共有 secret の ambiguous 全件スキャン拒否。R1 では destination 固定 routing。）
func TestLineLinkService_HandleWebhook_RejectsForeignDestinationSecret(t *testing.T) {
	lineUserID := "U-foreign-destination"
	lookupCalled := false
	updateCalled := false
	ownerRepo := &mockLstepOwnerRepo{
		findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
			lookupCalled = true
			return &model.Owner{ID: 1, ClinicID: 1}, nil
		},
		updateLineFollowedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
			updateCalled = true
			return true, nil
		},
	}
	settingRepo := &mockLineLinkSettingRepo{
		findByLineBotUserIDFn: func(_ context.Context, lineBotUserID string) (*model.LineReservationSetting, error) {
			if lineBotUserID == "bot-clinic-1" {
				return &model.LineReservationSetting{
					ClinicID: 1, LineBotUserID: lineBotUserID, LineChannelSecret: "clinic-one-secret",
				}, nil
			}
			return nil, apperrors.WrapNotFound("line_reservation_setting", lineBotUserID)
		},
	}
	svc := newTestLineLinkService(ownerRepo, &mockLineLinkTokenRepo{}, settingRepo)
	payload := WebhookPayload{
		Destination: "bot-clinic-1",
		Events: []WebhookEvent{
			{Type: "follow", Timestamp: testLineWebhookTimestamp, Source: struct {
				UserID string `json:"userId"`
			}{UserID: lineUserID}},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	// Sign with a different clinic's secret — must not authenticate as clinic 1.
	err = svc.HandleWebhook(context.Background(), body, makeLineSignature(body, "clinic-two-secret"))

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.False(t, lookupCalled)
	assert.False(t, updateCalled)
}

func TestLineLinkService_HandleWebhook_DoesNotLogRawLineUserID(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{name: "follow failure", eventType: "follow"},
		{name: "unfollow failure", eventType: "unfollow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const lineUserID = "U-sensitive-personal-identifier"
			var logBuffer bytes.Buffer
			previousLogger := slog.Default()
			slog.SetDefault(slog.New(slog.NewJSONHandler(&logBuffer, nil)))
			t.Cleanup(func() {
				slog.SetDefault(previousLogger)
			})

			ownerRepo := &mockLstepOwnerRepo{
				findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
					return nil, errors.New("db error")
				},
			}
			svc := newTestLineLinkService(
				ownerRepo,
				&mockLineLinkTokenRepo{},
				&mockLineLinkSettingRepo{},
			)
			payload := WebhookPayload{
				Destination: "dest",
				Events: []WebhookEvent{
					{Type: tt.eventType, Timestamp: testLineWebhookTimestamp, Source: struct {
						UserID string `json:"userId"`
					}{UserID: lineUserID}},
				},
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			err = svc.HandleWebhook(context.Background(), body, makeLineSignature(body, "secret"))

			require.Error(t, err)
			assert.NotContains(t, logBuffer.String(), lineUserID)
		})
	}
}

// --- LinkAccount tests ---
// verifyLineIDToken は https://api.line.me/... を直接呼び出す設計のため、
// ユニットテストでは LINE IDトークン検証失敗ケース（= Unauthorized）のみ確認する。

func TestLineLinkService_LinkAccount_InvalidLineToken(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{},
		jsonRespClient(http.StatusBadRequest, `{"error":"invalid token"}`),
	)
	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{
		LinkToken:   "any-token",
		LineIDToken: "invalid-line-token",
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
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

type closeTrackingBody struct {
	io.Reader
	closed bool
}

func (b *closeTrackingBody) Close() error {
	b.closed = true
	return nil
}

func newTestLineLinkServiceFull(
	ownerRepo *mockLstepOwnerRepo,
	tokenRepo *mockLineLinkTokenRepo,
	settingRepo *mockLineLinkSettingRepo,
	transactor Transactor,
	auditTx LineLinkAuditTxLogger,
	client *http.Client,
) *lineLinkService {
	return &lineLinkService{
		ownerRepo:          ownerRepo,
		lineLinkTokenRepo:  tokenRepo,
		lineSettingRepo:    settingRepo,
		lineCredentialRepo: &mockLineChannelCredentialRepo{},
		transactor:         transactor,
		auditTx:            auditTx,
		httpClient:         client,
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
			name:    "returns error when sub is empty",
			client:  jsonRespClient(http.StatusOK, `{"sub":""}`),
			wantErr: true,
		},
		{
			name:    "returns error when sub exceeds its bound",
			client:  jsonRespClient(http.StatusOK, `{"sub":"`+strings.Repeat("U", maxLineUserIDChars+1)+`"}`),
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

func TestVerifyLineIDToken_PropagatesContextAndClosesBoundedResponse(t *testing.T) {
	type contextKey string
	const requestKey contextKey = "request"
	body := &closeTrackingBody{Reader: strings.NewReader(`{"sub":"Uabc123"}`)}
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "context-value", req.Context().Value(requestKey))
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       body,
			Header:     make(http.Header),
		}, nil
	})}
	ctx := context.WithValue(context.Background(), requestKey, "context-value")

	got, err := verifyLineIDToken(ctx, "test-token", 1, &mockLineLinkSettingRepo{}, client)

	require.NoError(t, err)
	assert.Equal(t, "Uabc123", got)
	assert.True(t, body.closed)
}

func TestVerifyLineIDToken_RejectsOversizedResponse(t *testing.T) {
	client := jsonRespClient(
		http.StatusOK,
		strings.Repeat("x", maxLineVerifyResponseBytes+1),
	)

	_, err := verifyLineIDToken(context.Background(), "test-token", 1, &mockLineLinkSettingRepo{}, client)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrBadGateway)
}

func TestVerifyLineIDToken_NonOKResponseDoesNotLeakBody(t *testing.T) {
	client := jsonRespClient(http.StatusUnauthorized, `{"error":"private LINE response"}`)

	_, err := verifyLineIDToken(context.Background(), "test-token", 1, &mockLineLinkSettingRepo{}, client)

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "private LINE response")
}

func TestVerifyLineIDToken_PropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, req.Context().Err()
	})}

	_, err := verifyLineIDToken(ctx, "test-token", 1, &mockLineLinkSettingRepo{}, client)

	require.ErrorIs(t, err, context.Canceled)
}

func TestVerifyLineIDToken_DoesNotFollowRedirects(t *testing.T) {
	requestCount := 0
	callerRedirectCount := 0
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		if requestCount == 1 {
			return &http.Response{
				StatusCode: http.StatusTemporaryRedirect,
				Body:       io.NopCloser(strings.NewReader("redirect")),
				Header: http.Header{
					"Location": []string{"https://attacker.invalid/collect"},
				},
				Request: req,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"sub":"U-exfiltrated"}`)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	client := &http.Client{
		Timeout:   3 * time.Second,
		Transport: transport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			callerRedirectCount++
			return nil
		},
	}

	_, err := verifyLineIDToken(
		context.Background(),
		"sensitive-id-token",
		1,
		&mockLineLinkSettingRepo{},
		client,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrBadGateway)
	assert.Equal(t, 1, requestCount, "LINE ID token must never be resent to a redirect target")
	assert.Zero(t, callerRedirectCount, "caller redirect policy must be overridden for credential safety")
	assert.Equal(t, 3*time.Second, client.Timeout)
}

func TestNewLineLinkService_DefaultHTTPClientHasTimeout(t *testing.T) {
	svc := NewLineLinkService(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
		&mockLineChannelCredentialRepo{},
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{},
		nil,
	)

	impl, ok := svc.(*lineLinkService)
	require.True(t, ok)
	require.NotNil(t, impl.httpClient)
	assert.Positive(t, impl.httpClient.Timeout)
}

// --- LinkAccount additional branch tests ---

func TestLineLinkService_LinkAccount_TokenNotFound(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{
			lockUsableTokenFn: func(_ context.Context, _ string, _ time.Time) (*model.LineLinkToken, error) {
				return nil, apperrors.WrapNotFound("link_token", "")
			},
		},
		&mockLineLinkSettingRepo{},
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "bad-token", LineIDToken: "valid"})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestLineLinkService_LinkAccount_ClinicMismatch(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{
			lockUsableTokenFn: func(_ context.Context, _ string, _ time.Time) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 2, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestLineLinkService_LinkAccount_OwnerNotFound(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLstepOwnerRepo{
			lockLineLinkOwnerFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "10")
			},
		},
		&mockLineLinkTokenRepo{
			lockUsableTokenFn: func(_ context.Context, _ string, _ time.Time) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestLineLinkService_LinkAccount_AlreadyLinkedConflict(t *testing.T) {
	existing := "Uold"
	updateCalled := false
	svc := newTestLineLinkServiceFull(
		&mockLstepOwnerRepo{
			lockLineLinkOwnerFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
				return &model.Owner{ID: id, LineUserID: &existing}, nil
			},
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
				updateCalled = true
				return nil
			},
		},
		&mockLineLinkTokenRepo{
			lockUsableTokenFn: func(_ context.Context, _ string, _ time.Time) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.False(t, updateCalled, "public link flow must never overwrite an existing LINE user ID")
}

func TestLineLinkService_LinkAccount_UpdateLineUserIDError(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLstepOwnerRepo{
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
				return errors.New("db error")
			},
		},
		&mockLineLinkTokenRepo{
			lockUsableTokenFn: func(_ context.Context, _ string, _ time.Time) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
}

func TestLineLinkService_LinkAccount_ConsumeError(t *testing.T) {
	svc := newTestLineLinkServiceFull(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{
			lockUsableTokenFn: func(_ context.Context, _ string, _ time.Time) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
			consumeFn: func(_ context.Context, _ uint64, _ time.Time) error {
				return errors.New("db error")
			},
		},
		&mockLineLinkSettingRepo{},
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
}

func TestLineLinkService_LinkAccount_AuditFailureFailsClosed(t *testing.T) {
	auditCalled := false
	svc := newTestLineLinkServiceFull(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{
			lockUsableTokenFn: func(_ context.Context, _ string, _ time.Time) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
		},
		&mockLineLinkSettingRepo{},
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{
			logFn: func(_ context.Context, clinicID, ownerID uint64) error {
				auditCalled = true
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(10), ownerID)
				return errors.New("audit failure")
			},
		},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
	assert.True(t, auditCalled)
}

func TestLineLinkService_LinkAccount_AuditFailureRollsBackOwnerAndTokenState(t *testing.T) {
	type state struct {
		lineUserID string
		tokenUsed  bool
	}
	current := state{}
	transactor := rollbackLineLinkTransactor{
		snapshot: func() any { return current },
		restore:  func(before any) { current = before.(state) },
	}
	svc := newTestLineLinkServiceFull(
		&mockLstepOwnerRepo{
			lockLineLinkOwnerFn: func(_ context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
				return &model.Owner{ID: ownerID, ClinicID: clinicID}, nil
			},
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, lineUserID *string) error {
				current.lineUserID = *lineUserID
				return nil
			},
		},
		&mockLineLinkTokenRepo{
			lockUsableTokenFn: func(_ context.Context, _ string, _ time.Time) (*model.LineLinkToken, error) {
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
			consumeFn: func(_ context.Context, _ uint64, _ time.Time) error {
				current.tokenUsed = true
				return nil
			},
		},
		&mockLineLinkSettingRepo{},
		transactor,
		&mockLineLinkAuditTxLogger{
			logFn: func(_ context.Context, _, _ uint64) error {
				return errors.New("audit failure")
			},
		},
		jsonRespClient(http.StatusOK, `{"sub":"Uverified123"}`),
	)

	_, err := svc.LinkAccount(context.Background(), 1, LinkAccountInput{LinkToken: "tok", LineIDToken: "valid"})

	require.Error(t, err)
	assert.Empty(t, current.lineUserID)
	assert.False(t, current.tokenUsed)
}

func TestLineLinkService_LinkAccount_SuccessUsesOneTransactionContext(t *testing.T) {
	var txContext context.Context
	assertTxContext := func(t *testing.T, ctx context.Context) {
		t.Helper()
		require.True(t, ctx.Value(lineLinkTxTestContextKey{}).(bool))
		if txContext == nil {
			txContext = ctx
			return
		}
		assert.Same(t, txContext, ctx)
	}
	svc := newTestLineLinkServiceFull(
		&mockLstepOwnerRepo{
			findByIDFn: func(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
				assertTxContext(t, ctx)
				return &model.Owner{ID: ownerID, ClinicID: clinicID}, nil
			},
			lockLineLinkOwnerFn: func(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
				assertTxContext(t, ctx)
				return &model.Owner{ID: ownerID, ClinicID: clinicID}, nil
			},
			updateLineUserIDFn: func(ctx context.Context, _, _ uint64, _ *string) error {
				assertTxContext(t, ctx)
				return nil
			},
		},
		&mockLineLinkTokenRepo{
			lockUsableTokenFn: func(ctx context.Context, _ string, _ time.Time) (*model.LineLinkToken, error) {
				assertTxContext(t, ctx)
				return &model.LineLinkToken{ID: 1, ClinicID: 1, OwnerID: 10}, nil
			},
			consumeFn: func(ctx context.Context, _ uint64, _ time.Time) error {
				assertTxContext(t, ctx)
				return nil
			},
		},
		&mockLineLinkSettingRepo{},
		immediateLineLinkTransactor{},
		&mockLineLinkAuditTxLogger{
			logFn: func(ctx context.Context, _, _ uint64) error {
				assertTxContext(t, ctx)
				return nil
			},
		},
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
		&mockLstepOwnerRepo{},
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
	createCalled := false
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{},
		&mockLineLinkTokenRepo{
			createFn: func(_ context.Context, _ *model.LineLinkToken) error {
				createCalled = true
				return nil
			},
		},
		&mockLineLinkSettingRepo{
			findByClinicIDFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("db error")
			},
		},
	)
	_, err := svc.GenerateLinkToken(context.Background(), 1, 42)
	assert.Error(t, err)
	assert.False(t, createCalled, "setting lookup failure must not leave an unusable bearer token row")
}

// --- HandleWebhook additional branch tests ---

func TestLineLinkService_HandleWebhook_InvalidBody(t *testing.T) {
	svc := newTestLineLinkService(&mockLstepOwnerRepo{}, &mockLineLinkTokenRepo{}, &mockLineLinkSettingRepo{})
	body := []byte(`not-json`)
	sig := makeLineSignature(body, "secret")

	err := svc.HandleWebhook(context.Background(), body, sig)

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestLineLinkService_HandleWebhook_EmptyUserIDSkipped(t *testing.T) {
	handlerCalled := false
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{
			findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				handlerCalled = true
				return nil, nil
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "follow", Timestamp: testLineWebhookTimestamp, Source: struct {
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

func TestLineLinkService_HandleWebhook_FollowEvent_FindOwnerErrorIsPropagated(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{
			findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "follow", Timestamp: testLineWebhookTimestamp, Source: struct {
				UserID string `json:"userId"`
			}{UserID: "Uabc"}},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeLineSignature(body, "secret")

	err := svc.HandleWebhook(context.Background(), body, sig)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find owners by line user id")
}

func TestLineLinkService_HandleWebhook_UnfollowEvent_FindOwnerErrorIsPropagated(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{
			findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return nil, errors.New("db error")
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)

	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "unfollow", Timestamp: testLineWebhookTimestamp, Source: struct {
				UserID string `json:"userId"`
			}{UserID: "Uabc"}},
		},
	}
	body, _ := json.Marshal(payload)
	sig := makeLineSignature(body, "secret")

	err := svc.HandleWebhook(context.Background(), body, sig)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find owners by line user id")
}

// --- handleFollowEvent / handleUnfollowEvent direct tests ---

func TestHandleFollowEvent_UpdateErrorIsPropagated(t *testing.T) {
	svc := &lineLinkService{
		ownerRepo: &mockLstepOwnerRepo{
			findByLineUserIDFn: func(_ context.Context, clinicID uint64, _ string) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: clinicID}, nil
			},
			updateLineFollowedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				return false, errors.New("db error")
			},
		},
	}

	err := svc.handleFollowEvent(context.Background(), 1, "Uabc", time.Now())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update line_followed_at")
}

func TestHandleUnfollowEvent_UpdateErrorIsPropagated(t *testing.T) {
	svc := &lineLinkService{
		ownerRepo: &mockLstepOwnerRepo{
			findByLineUserIDFn: func(_ context.Context, clinicID uint64, _ string) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: clinicID}, nil
			},
			updateLineBlockedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
				return false, errors.New("db error")
			},
		},
	}

	err := svc.handleUnfollowEvent(context.Background(), 1, "Uabc", time.Now())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update line_blocked_at")
}

func TestLineLinkService_HandleWebhook_StaleOrMappingChangedEventIsNoOp(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{name: "follow CAS no-op", eventType: "follow"},
		{name: "unfollow CAS no-op", eventType: "unfollow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lineUserID := "U-stale-or-relinked"
			svc := newTestLineLinkService(
				&mockLstepOwnerRepo{
					findByLineUserIDFn: func(_ context.Context, clinicID uint64, _ string) (*model.Owner, error) {
						return &model.Owner{ID: 1, ClinicID: clinicID, LineUserID: &lineUserID}, nil
					},
					updateLineFollowedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
						return false, nil
					},
					updateLineBlockedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
						return false, nil
					},
				},
				&mockLineLinkTokenRepo{},
				&mockLineLinkSettingRepo{},
			)
			payload := WebhookPayload{
				Destination: "dest",
				Events: []WebhookEvent{
					{Type: tt.eventType, Timestamp: testLineWebhookTimestamp, Source: struct {
						UserID string `json:"userId"`
					}{UserID: lineUserID}},
				},
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			err = svc.HandleWebhook(context.Background(), body, makeLineSignature(body, "secret"))

			require.NoError(t, err)
		})
	}
}

func TestLineLinkService_HandleWebhook_NotFoundOwnerIsNoOp(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{name: "follow", eventType: "follow"},
		{name: "unfollow", eventType: "unfollow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			svc := newTestLineLinkService(
				&mockLstepOwnerRepo{
					findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
						return nil, apperrors.WrapNotFound("owner", "")
					},
					updateLineFollowedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
						updateCalled = true
						return true, nil
					},
					updateLineBlockedAtFn: func(_ context.Context, _, _ uint64, _ string, _ time.Time) (bool, error) {
						updateCalled = true
						return true, nil
					},
				},
				&mockLineLinkTokenRepo{},
				&mockLineLinkSettingRepo{},
			)
			payload := WebhookPayload{
				Destination: "dest",
				Events: []WebhookEvent{
					{Type: tt.eventType, Timestamp: testLineWebhookTimestamp, Source: struct {
						UserID string `json:"userId"`
					}{UserID: "U-not-linked"}},
				},
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			err = svc.HandleWebhook(context.Background(), body, makeLineSignature(body, "secret"))

			require.NoError(t, err)
			assert.False(t, updateCalled)
		})
	}
}

func TestLineLinkService_HandleWebhook_OwnerScopeMismatchIsPropagated(t *testing.T) {
	svc := newTestLineLinkService(
		&mockLstepOwnerRepo{
			findByLineUserIDFn: func(_ context.Context, _ uint64, uid string) (*model.Owner, error) {
				return &model.Owner{ID: 1, ClinicID: 2, LineUserID: &uid}, nil
			},
		},
		&mockLineLinkTokenRepo{},
		&mockLineLinkSettingRepo{},
	)
	payload := WebhookPayload{
		Destination: "dest",
		Events: []WebhookEvent{
			{Type: "follow", Timestamp: testLineWebhookTimestamp, Source: struct {
				UserID string `json:"userId"`
			}{UserID: "U-scope-mismatch"}},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	err = svc.HandleWebhook(context.Background(), body, makeLineSignature(body, "secret"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "scope mismatch")
}

func TestLineLinkService_HandleWebhook_RejectsInvalidEventTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
	}{
		{name: "missing timestamp", timestamp: 0},
		{name: "negative timestamp", timestamp: -1},
		{name: "timestamp too far in the future", timestamp: time.Now().Add(10 * time.Minute).UnixMilli()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookupCalled := false
			svc := newTestLineLinkService(
				&mockLstepOwnerRepo{
					findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
						lookupCalled = true
						return &model.Owner{ID: 1, ClinicID: 1}, nil
					},
				},
				&mockLineLinkTokenRepo{},
				&mockLineLinkSettingRepo{},
			)
			payload := WebhookPayload{
				Destination: "dest",
				Events: []WebhookEvent{
					{Type: "follow", Timestamp: tt.timestamp, Source: struct {
						UserID string `json:"userId"`
					}{UserID: "U-invalid-timestamp"}},
				},
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			err = svc.HandleWebhook(context.Background(), body, makeLineSignature(body, "secret"))

			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
			assert.False(t, lookupCalled)
		})
	}
}

// --- verifySignatureAnyClinic additional branch tests ---

func TestVerifySignatureAnyClinic_FindWebhookRouteError(t *testing.T) {
	routeRepo := &mockLineLinkSettingRepo{
		findWebhookRouteFn: func(_ context.Context, _ string) (uint64, bool, error) {
			return 0, false, errors.New("db error")
		},
	}
	credentialRepo := &mockLineChannelCredentialRepo{}
	svc := &lineLinkService{
		lineSettingRepo:    routeRepo,
		lineCredentialRepo: credentialRepo,
	}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-A","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "sig")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 1, routeRepo.findWebhookRouteCalls)
	assert.Equal(t, 0, credentialRepo.findCalls)
	assert.Equal(t, 0, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls)
}

func TestVerifySignatureAnyClinic_EmptySecretRejected(t *testing.T) {
	svc := &lineLinkService{
		lineSettingRepo: &mockLineLinkSettingRepo{
			findByLineBotUserIDFn: func(_ context.Context, lineBotUserID string) (*model.LineReservationSetting, error) {
				return &model.LineReservationSetting{
					ClinicID: 1, LineBotUserID: lineBotUserID, LineChannelSecret: "",
				}, nil
			},
		},
		lineCredentialRepo: &mockLineChannelCredentialRepo{
			findByClinicServiceKeyFn: func(_ context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error) {
				return &model.ClinicIntegration{
					ID: clinicID, ClinicID: clinicID, Service: service, KeyName: keyName, KeyValue: "",
				}, nil
			},
		},
	}
	body := []byte(`{"destination":"bot-A","events":[]}`)
	sig := makeLineSignature(body, "secret")

	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, sig)

	assert.False(t, ok, "empty channel secret must not authenticate")
	assert.Zero(t, clinicID)
}

func TestVerifySignatureAnyClinic_NoMatch(t *testing.T) {
	svc := &lineLinkService{
		lineSettingRepo: &mockLineLinkSettingRepo{
			findByLineBotUserIDFn: func(_ context.Context, lineBotUserID string) (*model.LineReservationSetting, error) {
				return &model.LineReservationSetting{
					ClinicID: 1, LineBotUserID: lineBotUserID, LineChannelSecret: "secret",
				}, nil
			},
		},
		lineCredentialRepo: &mockLineChannelCredentialRepo{},
	}
	body := []byte(`{"destination":"bot-A","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "wrong-signature")
	assert.False(t, ok)
	assert.Zero(t, clinicID)
}

// TestVerifySignatureAnyClinic_InvalidSignatureDoesNotReDecryptOnCacheHit
// は SEC-CS-F05: 無効署名パスでも setting ID キャッシュが効き、2 回目以降の
// AES 復号をスキップすることを保証する（destination 固定後も 1 setting 単位）。
func TestVerifySignatureAnyClinic_InvalidSignatureDoesNotReDecryptOnCacheHit(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)
	encSecret, err := cipher.Encrypt("clinic-secret")
	require.NoError(t, err)

	svc := &lineLinkService{
		lineSettingRepo: &mockLineLinkSettingRepo{
			findByLineBotUserIDFn: func(_ context.Context, lineBotUserID string) (*model.LineReservationSetting, error) {
				return &model.LineReservationSetting{
					ID: 11, ClinicID: 1, LineBotUserID: lineBotUserID, LineChannelSecret: encSecret,
				}, nil
			},
		},
		lineCredentialRepo: &mockLineChannelCredentialRepo{
			findByClinicServiceKeyFn: func(_ context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error) {
				return &model.ClinicIntegration{
					ID: 11, ClinicID: clinicID, Service: service, KeyName: keyName, KeyValue: encSecret,
				}, nil
			},
		},
		cipher: cipher,
	}

	var decryptCalls int
	prev := lineCredentialDecrypt
	lineCredentialDecrypt = func(ctx context.Context, c *crypto.AESGCMCipher, value string) string {
		decryptCalls++
		return prev(ctx, c, value)
	}
	t.Cleanup(func() { lineCredentialDecrypt = prev })

	body := []byte(`{"destination":"bot-A","events":[]}`)
	// 1st pass: cold cache → decrypt once for the single looked-up setting
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "invalid-signature")
	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 1, decryptCalls, "cold cache must decrypt the destination setting once")

	// 2nd pass: warm cache → no additional decrypts on invalid signature path
	clinicID, ok = svc.verifySignatureAnyClinic(context.Background(), body, "still-invalid")
	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 1, decryptCalls, "cache hit must not re-decrypt secrets")
}

// TestVerifySignatureAnyClinic_CacheInvalidatesWhenCiphertextRotates は
// channel secret ローテ後に古い平文を使い続けないことを確認する。
func TestVerifySignatureAnyClinic_CacheInvalidatesWhenCiphertextRotates(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)
	encOld, err := cipher.Encrypt("old-secret")
	require.NoError(t, err)
	encNew, err := cipher.Encrypt("new-secret")
	require.NoError(t, err)

	currentCiphertext := encOld
	svc := &lineLinkService{
		lineSettingRepo: &mockLineLinkSettingRepo{
			findByLineBotUserIDFn: func(_ context.Context, lineBotUserID string) (*model.LineReservationSetting, error) {
				return &model.LineReservationSetting{
					ID: 21, ClinicID: 1, LineBotUserID: lineBotUserID, LineChannelSecret: currentCiphertext,
				}, nil
			},
		},
		lineCredentialRepo: &mockLineChannelCredentialRepo{
			findByClinicServiceKeyFn: func(_ context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error) {
				return &model.ClinicIntegration{
					ID: 21, ClinicID: clinicID, Service: service, KeyName: keyName, KeyValue: currentCiphertext,
				}, nil
			},
		},
		cipher: cipher,
	}

	var decryptCalls int
	prev := lineCredentialDecrypt
	lineCredentialDecrypt = func(ctx context.Context, c *crypto.AESGCMCipher, value string) string {
		decryptCalls++
		return prev(ctx, c, value)
	}
	t.Cleanup(func() { lineCredentialDecrypt = prev })

	body := []byte(`{"destination":"bot-A","events":[]}`)
	_, _ = svc.verifySignatureAnyClinic(context.Background(), body, "nope")
	assert.Equal(t, 1, decryptCalls)

	currentCiphertext = encNew
	_, _ = svc.verifySignatureAnyClinic(context.Background(), body, "nope")
	assert.Equal(t, 2, decryptCalls, "rotated ciphertext must force re-decrypt")
}

func TestCachedDecryptChannelSecret_EvictsStaleEntryBeforeDecrypt(t *testing.T) {
	tests := []struct {
		name                 string
		cachedCiphertext     string
		credentialCiphertext string
		cachedExpiry         time.Time
	}{
		{
			name:                 "expired entry",
			cachedCiphertext:     "same-ciphertext-placeholder",
			credentialCiphertext: "same-ciphertext-placeholder",
			cachedExpiry:         time.Now().Add(-time.Second),
		},
		{
			name:                 "ciphertext mismatch",
			cachedCiphertext:     "old-ciphertext-placeholder",
			credentialCiphertext: "new-ciphertext-placeholder",
			cachedExpiry:         time.Now().Add(time.Minute),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const credentialID = uint64(81)
			svc := &lineLinkService{
				secretCache: map[uint64]lineChannelSecretCacheEntry{
					credentialID: {
						ciphertext: tt.cachedCiphertext,
						plaintext:  "cached-plaintext-placeholder",
						expiresAt:  tt.cachedExpiry,
					},
				},
			}
			credential := &model.ClinicIntegration{
				ID:       credentialID,
				ClinicID: 7,
				Service:  model.IntegrationServiceLstep,
				KeyName:  model.IntegrationKeyLineChannelSecret,
				KeyValue: tt.credentialCiphertext,
			}
			entryPresentDuringDecrypt := false
			previousDecrypt := lineCredentialDecrypt
			lineCredentialDecrypt = func(_ context.Context, _ *crypto.AESGCMCipher, _ string) string {
				_, entryPresentDuringDecrypt = svc.secretCache[credentialID]
				return "replacement-plaintext-placeholder"
			}
			t.Cleanup(func() { lineCredentialDecrypt = previousDecrypt })

			got := svc.cachedDecryptChannelSecret(context.Background(), credential)

			assert.False(t, entryPresentDuringDecrypt, "stale plaintext must be evicted before decrypt")
			assert.NotEmpty(t, got)
			replacement, ok := svc.secretCache[credentialID]
			require.True(t, ok)
			assert.Equal(t, tt.credentialCiphertext, replacement.ciphertext)
			assert.True(t, replacement.expiresAt.After(time.Now()))
		})
	}
}

// TestVerifySignatureAnyClinic_ConcurrencyLimitConstant documents the SEC-CS-F05
// global verification semaphore capacity.
func TestVerifySignatureAnyClinic_ConcurrencyLimitConstant(t *testing.T) {
	assert.Equal(t, int64(32), maxConcurrentLineWebhookVerifications)
	assert.Equal(t, 30*time.Second, lineChannelSecretCacheTTL)
}
