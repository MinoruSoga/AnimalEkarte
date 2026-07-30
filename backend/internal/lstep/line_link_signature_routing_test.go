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
// FindByLineBotUserID → one decrypt → one HMAC. Never FindAll / O(clinic_count).

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

// nClinicSettingRepo builds N clinics with distinct secrets; bot-A maps to clinic 7.
func nClinicSettingRepo(n int) *mockLineLinkSettingRepo {
	byBot := make(map[string]model.LineReservationSetting, n)
	all := make([]model.LineReservationSetting, 0, n)
	for i := 1; i <= n; i++ {
		botID := fmt.Sprintf("bot-%d", i)
		// Clinic 7 is addressable as "bot-A" for the fixed routing scenarios.
		if i == 7 {
			botID = "bot-A"
		}
		s := model.LineReservationSetting{
			ID:                uint64(i),
			ClinicID:          uint64(i),
			LineBotUserID:     botID,
			LineChannelSecret: fmt.Sprintf("secret-%d", i),
		}
		byBot[botID] = s
		all = append(all, s)
	}
	return &mockLineLinkSettingRepo{
		findByLineBotUserIDFn: func(_ context.Context, lineBotUserID string) (*model.LineReservationSetting, error) {
			s, ok := byBot[lineBotUserID]
			if !ok {
				return nil, apperrors.WrapNotFound("line_reservation_setting", lineBotUserID)
			}
			cp := s
			return &cp, nil
		},
		findAllFn: func(_ context.Context) ([]model.LineReservationSetting, error) {
			out := make([]model.LineReservationSetting, len(all))
			copy(out, all)
			return out, nil
		},
	}
}

func TestVerifySignatureAnyClinic_InvalidSignature_FixedWork_N20(t *testing.T) {
	repo := nClinicSettingRepo(20)
	svc := &lineLinkService{lineSettingRepo: repo}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-A","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "invalid-signature")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 0, repo.findAllCalls, "must not call FindAll")
	assert.Equal(t, 1, repo.findByLineBotUserIDCalls, "must lookup exactly one setting by destination")
	assert.LessOrEqual(t, *decryptCalls, 1, "at most one secret decrypt")
	assert.LessOrEqual(t, *hmacCalls, 1, "at most one HMAC")
}

func TestVerifySignatureAnyClinic_InvalidSignature_HMACIndependentOfClinicCount(t *testing.T) {
	run := func(n int) int {
		repo := nClinicSettingRepo(n)
		svc := &lineLinkService{lineSettingRepo: repo}
		hmacCalls := installHMACCounter(t)
		body := []byte(`{"destination":"bot-A","events":[]}`)
		_, ok := svc.verifySignatureAnyClinic(context.Background(), body, "invalid-signature")
		require.False(t, ok)
		assert.Equal(t, 0, repo.findAllCalls)
		return *hmacCalls
	}

	hmac5 := run(5)
	hmac50 := run(50)
	assert.Equal(t, hmac5, hmac50, "HMAC count must be independent of clinic count")
	assert.LessOrEqual(t, hmac5, 1)
	assert.LessOrEqual(t, hmac50, 1)
}

func TestVerifySignatureAnyClinic_UnknownDestination_ZeroCrypto(t *testing.T) {
	repo := nClinicSettingRepo(10)
	svc := &lineLinkService{lineSettingRepo: repo}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-unknown","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "anything")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 0, repo.findAllCalls)
	assert.Equal(t, 1, repo.findByLineBotUserIDCalls)
	assert.Equal(t, 0, *decryptCalls, "unknown destination must not decrypt secrets")
	assert.Equal(t, 0, *hmacCalls, "unknown destination must not HMAC")
}

func TestVerifySignatureAnyClinic_ValidDestination_OneHMAC_ReturnsClinic(t *testing.T) {
	repo := nClinicSettingRepo(20)
	svc := &lineLinkService{lineSettingRepo: repo}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"bot-A","events":[]}`)
	sig := makeLineSignature(body, "secret-7")
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, sig)

	require.True(t, ok)
	assert.Equal(t, uint64(7), clinicID)
	assert.Equal(t, 0, repo.findAllCalls)
	assert.Equal(t, 1, repo.findByLineBotUserIDCalls)
	assert.LessOrEqual(t, *decryptCalls, 1)
	assert.Equal(t, 1, *hmacCalls)
}

func TestVerifySignatureAnyClinic_MissingDestination_NoFindAll(t *testing.T) {
	repo := nClinicSettingRepo(8)
	svc := &lineLinkService{lineSettingRepo: repo}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	// No destination field at all.
	body := []byte(`{"events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "sig")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 0, repo.findAllCalls)
	assert.Equal(t, 0, repo.findByLineBotUserIDCalls, "missing destination must not query settings")
	assert.Equal(t, 0, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls)
}

func TestVerifySignatureAnyClinic_EmptyDestination_NoFindAll(t *testing.T) {
	repo := nClinicSettingRepo(8)
	svc := &lineLinkService{lineSettingRepo: repo}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	body := []byte(`{"destination":"","events":[]}`)
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "sig")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 0, repo.findAllCalls)
	assert.Equal(t, 0, repo.findByLineBotUserIDCalls)
	assert.Equal(t, 0, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls)
}

func TestVerifySignatureAnyClinic_OversizedDestination_NoFindAll(t *testing.T) {
	repo := nClinicSettingRepo(8)
	svc := &lineLinkService{lineSettingRepo: repo}
	hmacCalls := installHMACCounter(t)
	decryptCalls := installDecryptCounter(t)

	// Bound is 128 chars; 200 exceeds it.
	oversized := strings.Repeat("U", 200)
	body := []byte(fmt.Sprintf(`{"destination":%q,"events":[]}`, oversized))
	clinicID, ok := svc.verifySignatureAnyClinic(context.Background(), body, "sig")

	assert.False(t, ok)
	assert.Zero(t, clinicID)
	assert.Equal(t, 0, repo.findAllCalls)
	assert.Equal(t, 0, repo.findByLineBotUserIDCalls)
	assert.Equal(t, 0, *decryptCalls)
	assert.Equal(t, 0, *hmacCalls)
}
