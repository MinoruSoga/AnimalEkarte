package lstep

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/model"
)

// SEC-CS-F05-R1: fixed-work LINE webhook signature routing.
// Invalid/valid webhooks must resolve at most one clinic via destination →
// route metadata → canonical credential → one decrypt → one HMAC. Never FindAll / O(clinic_count).

func installHMACCounter(t *testing.T) *int {
	t.Helper()
	var hmacCalls int
	prev := lineSignatureVerifier
	lineSignatureVerifier = func(body []byte, signature, channelSecret string) bool {
		hmacCalls++
		return prev(body, signature, channelSecret)
	}
	t.Cleanup(func() { lineSignatureVerifier = prev })
	return &hmacCalls
}

func installDecryptCounter(t *testing.T) *int {
	t.Helper()
	var decryptCalls int
	prev := lineCredentialDecrypt
	lineCredentialDecrypt = func(ctx context.Context, c *crypto.AESGCMCipher, value string) string {
		decryptCalls++
		return prev(ctx, c, value)
	}
	t.Cleanup(func() { lineCredentialDecrypt = prev })
	return &decryptCalls
}

// nClinicSettingRepos builds N clinics with distinct encrypted canonical credentials;
// "bot-A" always maps to clinic 7 (clinic 7 is included even when n < 7 so
// fixed-routing scenarios stay valid).
func nClinicSettingRepos(
	t *testing.T,
	n int,
) (*mockLineLinkSettingRepo, *mockLineChannelCredentialRepo, *crypto.AESGCMCipher) {
	t.Helper()
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)
	if n < 1 {
		n = 1
	}
	byBot := make(map[string]uint64, n+1)
	credentialByClinic := make(map[uint64]string, n+1)
	all := make([]model.LineReservationSetting, 0, n+1)
	add := func(clinicID uint64, botID string) {
		credential, encryptErr := cipher.Encrypt(fmt.Sprintf("secret-%d", clinicID))
		require.NoError(t, encryptErr)
		byBot[botID] = clinicID
		credentialByClinic[clinicID] = credential
		all = append(all, model.LineReservationSetting{ID: clinicID, ClinicID: clinicID, LineBotUserID: botID})
	}
	for i := 1; i <= n; i++ {
		if i == 7 {
			continue // added below as bot-A
		}
		add(uint64(i), fmt.Sprintf("bot-%d", i))
	}
	// Always provision bot-A → clinic 7 (target of fixed-work assertions).
	add(7, "bot-A")
	routeRepo := &mockLineLinkSettingRepo{
		findWebhookRouteFn: func(_ context.Context, lineBotUserID string) (uint64, bool, error) {
			clinicID, ok := byBot[lineBotUserID]
			if !ok {
				return 0, false, apperrors.WrapNotFound("line_reservation_setting", lineBotUserID)
			}
			return clinicID, false, nil
		},
		findAllFn: func(_ context.Context) ([]model.LineReservationSetting, error) {
			out := make([]model.LineReservationSetting, len(all))
			copy(out, all)
			return out, nil
		},
	}
	credentialRepo := &mockLineChannelCredentialRepo{
		findByClinicServiceKeyFn: func(_ context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error) {
			return &model.ClinicIntegration{
				ID:       clinicID,
				ClinicID: clinicID,
				Service:  service,
				KeyName:  keyName,
				KeyValue: credentialByClinic[clinicID],
			}, nil
		},
	}
	return routeRepo, credentialRepo, cipher
}

func TestVerifySignatureAnyClinic_InvalidSignature_FixedWork_N20(t *testing.T) {
	repo, credentialRepo, cipher := nClinicSettingRepos(t, 20)
	svc := &lineLinkService{lineSettingRepo: repo, lineCredentialRepo: credentialRepo, cipher: cipher}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-A","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "invalid-signature")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 0, repo.findAllCalls, "must not call FindAll")
	assert.Equal(t, 1, repo.findWebhookRouteCalls, "must lookup exactly one route by destination")
	assert.Equal(t, 1, credentialRepo.findCalls, "must lookup exactly one canonical credential")
	assert.LessOrEqual(t, *decryptCalls, 1, "at most one secret decrypt")
	assert.LessOrEqual(t, *hmacCalls, 1, "at most one HMAC")
}

func TestVerifySignatureAnyClinic_InvalidSignature_HMACIndependentOfClinicCount(t *testing.T) {
	run := func(n int) int {
		repo, credentialRepo, cipher := nClinicSettingRepos(t, n)
		svc := &lineLinkService{lineSettingRepo: repo, lineCredentialRepo: credentialRepo, cipher: cipher}
		hmacCalls := installHMACCounter(t)
		body := []byte(`{"destination":"bot-A","events":[]}`)
		_, ok := svc.verifySignatureAnyClinic(context.Background(), body, "invalid-signature")
		require.False(t, ok)
		assert.Equal(t, 0, repo.findAllCalls)
		assert.Equal(t, 1, credentialRepo.findCalls)
		return *hmacCalls
	}

	hmac5 := run(5)
	hmac50 := run(50)
	assert.Equal(t, hmac5, hmac50, "HMAC count must be independent of clinic count")
	assert.LessOrEqual(t, hmac5, 1)
	assert.LessOrEqual(t, hmac50, 1)
}

func TestVerifySignatureAnyClinic_UnknownDestination_ZeroCrypto(t *testing.T) {
	repo, credentialRepo, cipher := nClinicSettingRepos(t, 10)
	svc := &lineLinkService{lineSettingRepo: repo, lineCredentialRepo: credentialRepo, cipher: cipher}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-unknown","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "anything")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 0, repo.findAllCalls)
	assert.Equal(t, 1, repo.findWebhookRouteCalls)
	assert.Equal(t, 0, credentialRepo.findCalls)
	assert.Equal(t, 0, *decryptCalls, "unknown destination must not decrypt secrets")
	assert.Equal(t, 0, *hmacCalls, "unknown destination must not HMAC")
}

func TestVerifySignatureAnyClinic_ValidDestination_OneHMAC_ReturnsClinic(t *testing.T) {
	repo, credentialRepo, cipher := nClinicSettingRepos(t, 20)
	svc := &lineLinkService{lineSettingRepo: repo, lineCredentialRepo: credentialRepo, cipher: cipher}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-A","events":[]}`)
	sig := makeLineSignature(body, "secret-7")
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, sig)

	require.True(t, ok)
	assert.Equal(t, uint64(7), clinicID)
	assert.Equal(t, 0, repo.findAllCalls)
	assert.Equal(t, 1, repo.findWebhookRouteCalls)
	assert.Equal(t, 1, credentialRepo.findCalls)
	assert.LessOrEqual(t, *decryptCalls, 1)
	assert.Equal(t, 1, *hmacCalls)
}

func TestVerifySignatureAnyClinic_LegacyCredentialPresent_HoldsBeforeCanonicalLookup(t *testing.T) {
	routeRepo := &mockLineLinkSettingRepo{
		findWebhookRouteFn: func(_ context.Context, _ string) (uint64, bool, error) {
			return 7, true, nil
		},
	}
	credentialRepo := &mockLineChannelCredentialRepo{}
	svc := &lineLinkService{lineSettingRepo: routeRepo, lineCredentialRepo: credentialRepo}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-A","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "signature-placeholder")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 1, routeRepo.findWebhookRouteCalls)
	assert.Equal(t, 0, credentialRepo.findCalls, "legacy presence must HOLD before canonical lookup")
	assert.Equal(t, 0, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls)
}

func TestVerifySignatureAnyClinic_CanonicalCredentialIdentityMismatch_FailsClosed(t *testing.T) {
	routeRepo := &mockLineLinkSettingRepo{
		findWebhookRouteFn: func(_ context.Context, _ string) (uint64, bool, error) {
			return 7, false, nil
		},
	}
	credentialRepo := &mockLineChannelCredentialRepo{
		findByClinicServiceKeyFn: func(_ context.Context, _ uint64, service, keyName string) (*model.ClinicIntegration, error) {
			return &model.ClinicIntegration{
				ID: 8, ClinicID: 8, Service: service, KeyName: keyName, KeyValue: "canonical-placeholder",
			}, nil
		},
	}
	svc := &lineLinkService{lineSettingRepo: routeRepo, lineCredentialRepo: credentialRepo}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-A","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "signature-placeholder")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 1, credentialRepo.findCalls)
	assert.Equal(t, 0, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls)
}

func TestVerifySignatureAnyClinic_MissingCanonicalCredential_FailsClosed(t *testing.T) {
	routeRepo := &mockLineLinkSettingRepo{
		findWebhookRouteFn: func(_ context.Context, _ string) (uint64, bool, error) {
			return 7, false, nil
		},
	}
	credentialRepo := &mockLineChannelCredentialRepo{
		findByClinicServiceKeyFn: func(_ context.Context, _ uint64, _, _ string) (*model.ClinicIntegration, error) {
			return nil, apperrors.WrapNotFound("clinic_integration", "canonical credential")
		},
	}
	svc := &lineLinkService{lineSettingRepo: routeRepo, lineCredentialRepo: credentialRepo}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-A","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "signature-placeholder")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 1, credentialRepo.findCalls)
	assert.Equal(t, 0, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls)
}

func TestVerifySignatureAnyClinic_NilCipherFailsBeforeHMAC(t *testing.T) {
	routeRepo, credentialRepo, _ := nClinicSettingRepos(t, 8)
	svc := &lineLinkService{
		lineSettingRepo:    routeRepo,
		lineCredentialRepo: credentialRepo,
	}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-A","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "signature-placeholder")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 1, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls, "unavailable canonical decrypt capability must fail before HMAC")
}

func TestVerifySignatureAnyClinic_MalformedCanonicalCiphertext_FailsBeforeHMAC(t *testing.T) {
	cipher, err := crypto.NewAESGCMCipher(testIntegrationKeyHex)
	require.NoError(t, err)
	routeRepo := &mockLineLinkSettingRepo{
		findWebhookRouteFn: func(_ context.Context, _ string) (uint64, bool, error) {
			return 7, false, nil
		},
	}
	credentialRepo := &mockLineChannelCredentialRepo{
		findByClinicServiceKeyFn: func(_ context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error) {
			return &model.ClinicIntegration{
				ID:       77,
				ClinicID: clinicID,
				Service:  service,
				KeyName:  keyName,
				KeyValue: "malformed-ciphertext-placeholder",
			}, nil
		},
	}
	svc := &lineLinkService{
		lineSettingRepo:    routeRepo,
		lineCredentialRepo: credentialRepo,
		cipher:             cipher,
	}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-A","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "signature-placeholder")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 1, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls, "malformed canonical ciphertext must fail before HMAC")
}

func TestVerifySignatureAnyClinic_MissingDestination_NoFindAll(t *testing.T) {
	repo, credentialRepo, cipher := nClinicSettingRepos(t, 8)
	svc := &lineLinkService{lineSettingRepo: repo, lineCredentialRepo: credentialRepo, cipher: cipher}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	// No destination field at all.
	body := []byte(`{"events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "sig")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 0, repo.findAllCalls)
	assert.Equal(t, 0, repo.findWebhookRouteCalls, "missing destination must not query settings")
	assert.Equal(t, 0, credentialRepo.findCalls)
	assert.Equal(t, 0, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls)
}

func TestVerifySignatureAnyClinic_EmptyDestination_NoFindAll(t *testing.T) {
	repo, credentialRepo, cipher := nClinicSettingRepos(t, 8)
	svc := &lineLinkService{lineSettingRepo: repo, lineCredentialRepo: credentialRepo, cipher: cipher}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "sig")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 0, repo.findAllCalls)
	assert.Equal(t, 0, repo.findWebhookRouteCalls)
	assert.Equal(t, 0, credentialRepo.findCalls)
	assert.Equal(t, 0, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls)
}

func TestVerifySignatureAnyClinic_OversizedDestination_NoFindAll(t *testing.T) {
	repo, credentialRepo, cipher := nClinicSettingRepos(t, 8)
	svc := &lineLinkService{lineSettingRepo: repo, lineCredentialRepo: credentialRepo, cipher: cipher}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	// Bound is 128 chars; 200 exceeds it.
	oversized := strings.Repeat("U", 200)
	body := []byte(fmt.Sprintf(`{"destination":%q,"events":[]}`, oversized))
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "sig")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 0, repo.findAllCalls)
	assert.Equal(t, 0, repo.findWebhookRouteCalls)
	assert.Equal(t, 0, credentialRepo.findCalls)
	assert.Equal(t, 0, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls)
}
