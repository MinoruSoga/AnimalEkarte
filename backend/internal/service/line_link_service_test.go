package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
func (m *mockLineLinkOwnerRepo) FindAll(_ context.Context, _ uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
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

// --- mock: AuditService ---

type mockLineLinkAuditService struct{}

func (m *mockLineLinkAuditService) Log(_ context.Context, _ *model.AuditLog) error { return nil }
func (m *mockLineLinkAuditService) LogEntry(_ context.Context, _ *AuditLogInput) error {
	return nil
}
func (m *mockLineLinkAuditService) LogAuthLogin(_ context.Context, _, _ *uint64, _, _, _ string) error {
	return nil
}
func (m *mockLineLinkAuditService) LogLstepOperation(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
	return nil
}

func (m *mockLineLinkAuditService) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64, _ any) error {
	return nil
}
func (m *mockLineLinkAuditService) LogMedicalRecordChange(_ context.Context, _ uint64, _ *uint64, _ string, _ uint64, _, _ map[string]any) error {
	return nil
}
func (m *mockLineLinkAuditService) LogVitalChange(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
	return nil
}
func (m *mockLineLinkAuditService) LogAddendumCreate(_ context.Context, _ uint64, _ *uint64, _, _ uint64, _ *model.MedicalRecordAddendum) error {
	return nil
}
func (m *mockLineLinkAuditService) LogClinicSwitch(_ context.Context, _ *uint64, _, _ uint64, _, _ string) error {
	return nil
}

func newTestLineLinkService(
	ownerRepo *mockLineLinkOwnerRepo,
	tokenRepo *mockLineLinkTokenRepo,
	settingRepo *mockLineLinkSettingRepo,
) LineLinkService {
	return NewLineLinkService(ownerRepo, tokenRepo, settingRepo, &mockLineLinkAuditService{})
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
