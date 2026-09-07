package lstep

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/infra/line"
	"github.com/animal-ekarte/backend/internal/model"
)

// LinkTokenResult は GenerateLinkToken の返り値。
type LinkTokenResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	LiffURL   string    `json:"liff_url"`
}

// LinkAccountInput は LinkAccount の入力。
type LinkAccountInput struct {
	LinkToken   string `json:"link_token"`
	LineIDToken string `json:"line_id_token"`
}

// WebhookEvent は LINE Webhook イベントを表す最小構造体。
type WebhookEvent struct {
	Type           string `json:"type"`
	Timestamp      int64  `json:"timestamp"`
	WebhookEventID string `json:"webhookEventId"`
	Source         struct {
		UserID string `json:"userId"`
	} `json:"source"`
}

// WebhookPayload は LINE Webhook リクエストボディ。
type WebhookPayload struct {
	Destination string         `json:"destination"`
	Events      []WebhookEvent `json:"events"`
}

// LineLinkService は LINE User ID 自動取得・飼い主紐付けフローを実装する（BE-021）。
type LineLinkService interface {
	GenerateLinkToken(ctx context.Context, clinicID, ownerID uint64) (*LinkTokenResult, error)
	LinkAccount(ctx context.Context, clinicID uint64, input LinkAccountInput) (*model.Owner, error)
	HandleWebhook(ctx context.Context, body []byte, signature string) error
}

type lineLinkService struct {
	ownerRepo          lineLinkOwnerRepo
	lineLinkTokenRepo  LineLinkTokenRepository
	lineSettingRepo    lstepLineSettingReader
	lineCredentialRepo lineChannelCredentialReader
	transactor         Transactor
	auditTx            LineLinkAuditTxLogger
	// cipher は Webhook 署名検証時に line_channel_secret を復号するために使う（H-4）。
	// nil の場合は復号なしで動作する（開発環境で INTEGRATION_ENCRYPTION_KEY 未設定時）。
	cipher *crypto.AESGCMCipher
	// httpClient は LINE ID Token 検証 API 呼び出しに使う。
	httpClient *http.Client

	// secretCache は canonical clinic_integration ID ごとの復号済み channel secret を短 TTL で保持する。
	// 同一 credential への再検証で AES 復号を繰り返さない（SEC-CS-F05 二次防御）。
	// 一次防御は destination → route metadata → canonical credential の固定コスト routing。
	secretCacheMu sync.Mutex
	secretCache   map[uint64]lineChannelSecretCacheEntry
}

// lineChannelSecretCacheEntry は decrypt 結果のキャッシュ 1 件。
// ciphertext が変わった場合はヒットさせず再復号する。
type lineChannelSecretCacheEntry struct {
	ciphertext string
	plaintext  string
	expiresAt  time.Time
}

const (
	lineLinkTokenBytes         = 32
	lineLinkTokenTTL           = 24 * time.Hour
	lineVerifyHTTPTimeout      = 5 * time.Second
	maxLineVerifyResponseBytes = 64 * 1024
	maxLineUserIDChars         = 64
	maxLineWebhookFutureSkew   = 5 * time.Minute

	// lineChannelSecretCacheTTL は復号済み secret キャッシュの寿命。
	// 設定ローテ後も長くてもこの時間で平文が捨てられる。
	lineChannelSecretCacheTTL = 30 * time.Second

	// maxLineWebhookDestinationChars は webhook body から抜く destination の上限。
	// 過大な値は lookup 前に reject し、秘密計算を一切走らせない。
	maxLineWebhookDestinationChars = 128

	// maxConcurrentLineWebhookVerifications は同時に走れる署名検証の上限。
	// destination ルーティング後の二次 backpressure（SEC-CS-F05）。
	maxConcurrentLineWebhookVerifications int64 = 32
)

// lineWebhookVerifySem はプロセス全体の webhook 署名検証同時実行数を制限する。
var lineWebhookVerifySem = semaphore.NewWeighted(maxConcurrentLineWebhookVerifications)

// lineCredentialDecrypt は canonical webhook credential 専用の strict 復号関数。
// テストで呼び出し回数を観測するために差し替え可能。
var lineCredentialDecrypt = decryptCanonicalLineCredential

// decryptCanonicalLineCredential は canonical credential の復号能力欠落と復号失敗を
// fail-closed にする。
// generic DecryptLineCredential の legacy plaintext fallback は reservation 移行用であり、
// canonical clinic_integrations credential の署名検証には適用しない。
func decryptCanonicalLineCredential(
	_ context.Context,
	cipher *crypto.AESGCMCipher,
	value string,
) string {
	if value == "" {
		return ""
	}
	if cipher == nil {
		return ""
	}
	plaintext, err := cipher.Decrypt(value)
	if err != nil {
		return ""
	}
	return plaintext
}

// lineSignatureVerifier は webhook 検証パスの HMAC 検証関数。
// 本番は verifyLineSignature。テストで呼び出し回数を観測するために差し替え可能。
var lineSignatureVerifier = verifyLineSignature

// NewLineLinkService は LineLinkService を初期化して返す。
// webhook 署名検証には lstep 連携と同一の non-nil cipher が必須であり、nil の場合は
// canonical credential を復号せず fail-closed にする。
func NewLineLinkService(
	ownerRepo lineLinkOwnerRepo,
	lineLinkTokenRepo LineLinkTokenRepository,
	lineSettingRepo lstepLineSettingReader,
	lineCredentialRepo lineChannelCredentialReader,
	transactor Transactor,
	auditTx LineLinkAuditTxLogger,
	cipher *crypto.AESGCMCipher,
) LineLinkService {
	return &lineLinkService{
		ownerRepo:          ownerRepo,
		lineLinkTokenRepo:  lineLinkTokenRepo,
		lineSettingRepo:    lineSettingRepo,
		lineCredentialRepo: lineCredentialRepo,
		transactor:         transactor,
		auditTx:            auditTx,
		cipher:             cipher,
		httpClient:         &http.Client{Timeout: lineVerifyHTTPTimeout},
	}
}

// GenerateLinkToken は飼い主向けの一時トークンを生成し、LIFF URL を返す（24時間有効）。
func (s *lineLinkService) GenerateLinkToken(ctx context.Context, clinicID, ownerID uint64) (*LinkTokenResult, error) {
	// 所有者の存在確認
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}

	// Resolve LIFF ID before writing so a lookup failure cannot leave an
	// unusable bearer-token row behind. Missing line_reservation_settings is not
	// an owner miss (BUG-504): fall through to L-step clinic_integrations LiffID.
	liffID, err := s.resolveLinkTokenLiffID(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	tokenBytes := make([]byte, lineLinkTokenBytes)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, apperrors.Wrap(fmt.Errorf("failed to generate token: %w", err), "internal error")
	}
	rawToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(lineLinkTokenTTL)

	t := &model.LineLinkToken{
		ClinicID:  clinicID,
		OwnerID:   owner.ID,
		Token:     digestLineLinkToken(rawToken),
		ExpiresAt: expiresAt,
	}
	if err := s.lineLinkTokenRepo.Create(ctx, t); err != nil {
		return nil, apperrors.Wrap(err, "failed to create link token")
	}

	liffURL := ""
	if liffID != "" {
		// FE の LiffLinkPage（frontend/liff/src/pages/LiffLinkPage.tsx）は
		// token と clinic_id の両方をクエリから読む（clinic_id 欠落だと useLiffLink が
		// 即座に「無効なURL」エラーで停止する・SD-14）。
		query := url.Values{
			"token":     {rawToken},
			"clinic_id": {fmt.Sprintf("%d", clinicID)},
		}
		liffURL = fmt.Sprintf("https://liff.line.me/%s?%s", liffID, query.Encode())
	}

	return &LinkTokenResult{
		Token:     rawToken,
		ExpiresAt: expiresAt,
		LiffURL:   liffURL,
	}, nil
}

// resolveLinkTokenLiffID prefers line_reservation_settings.liff_id, then falls
// back to the L-step clinic_integrations liff_id staff configure in settings.
// Absence of either store is not an error; empty LiffID yields empty liff_url.
func (s *lineLinkService) resolveLinkTokenLiffID(ctx context.Context, clinicID uint64) (string, error) {
	if s.lineSettingRepo != nil {
		setting, err := s.lineSettingRepo.FindByClinicID(ctx, clinicID)
		if err != nil && !apperrors.IsNotFound(err) {
			return "", apperrors.Wrap(err, "failed to find line reservation setting")
		}
		if setting != nil && setting.LiffID != "" {
			return setting.LiffID, nil
		}
	}

	if s.lineCredentialRepo == nil {
		return "", nil
	}
	credential, err := s.lineCredentialRepo.FindCredentialByClinicServiceKey(
		ctx,
		clinicID,
		model.IntegrationServiceLstep,
		model.IntegrationKeyLiffID,
	)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return "", nil
		}
		return "", apperrors.Wrap(err, "failed to find lstep liff id")
	}
	if credential == nil || credential.KeyValue == "" {
		return "", nil
	}
	// liff_id is stored plaintext (model.IsEncryptedKey is false).
	return credential.KeyValue, nil
}

// LinkAccount は LINE ID Token を検証してトークン対応の飼い主に LINE User ID を紐付ける。
func (s *lineLinkService) LinkAccount(ctx context.Context, clinicID uint64, input LinkAccountInput) (*model.Owner, error) {
	if input.LinkToken == "" || len(input.LinkToken) > maxLineLinkTokenChars {
		return nil, apperrors.WrapInvalidInput("invalid link token")
	}
	if input.LineIDToken == "" || len(input.LineIDToken) > maxLineIDTokenChars {
		return nil, apperrors.WrapInvalidInput("invalid LINE ID token")
	}
	if s.transactor == nil || s.auditTx == nil {
		return nil, apperrors.WrapInternalServerError("LINE link transaction dependencies are not configured")
	}

	lineUserID, err := verifyLineIDToken(ctx, input.LineIDToken, clinicID, s.lineSettingRepo, s.httpClient)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		if errors.Is(err, apperrors.ErrUnauthorized) || errors.Is(err, apperrors.ErrBadGateway) {
			return nil, err
		}
		return nil, apperrors.WrapInternalServerError("LINE ID token verification unavailable")
	}

	lookupAt := time.Now()
	var linkedOwner *model.Owner
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		linkToken, err := s.lineLinkTokenRepo.LockUsableByRawToken(txCtx, input.LinkToken, lookupAt)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return apperrors.WrapInvalidInput("invalid or expired link token")
			}
			return apperrors.Wrap(err, "failed to lock link token")
		}
		if linkToken.ClinicID != clinicID {
			return apperrors.WrapInvalidInput("link token clinic mismatch")
		}

		owner, err := s.ownerRepo.LockLineLinkOwner(txCtx, clinicID, linkToken.OwnerID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return apperrors.WrapInvalidInput("invalid or expired link token")
			}
			return apperrors.Wrap(err, "failed to lock owner")
		}
		if owner.LineUserID != nil && *owner.LineUserID != "" {
			return apperrors.WrapConflict("line user id already set")
		}
		if err := s.ownerRepo.UpdateLineUserID(txCtx, clinicID, linkToken.OwnerID, &lineUserID); err != nil {
			return apperrors.Wrap(err, "failed to update line user id")
		}
		if err := s.lineLinkTokenRepo.Consume(txCtx, linkToken.ID, time.Now()); err != nil {
			return apperrors.Wrap(err, "failed to consume link token")
		}
		if err := s.auditTx.LogOwnerLineLinkTx(txCtx, clinicID, linkToken.OwnerID); err != nil {
			return apperrors.Wrap(err, "failed to write LINE link audit")
		}

		reloaded, err := s.ownerRepo.FindByID(txCtx, clinicID, linkToken.OwnerID)
		if err != nil {
			return apperrors.Wrap(err, "failed to reload linked owner")
		}
		reloaded.LineUserID = &lineUserID
		linkedOwner = reloaded
		return nil
	}); err != nil {
		return nil, err
	}
	return linkedOwner, nil
}

// HandleWebhook は LINE Webhook を処理する。
func (s *lineLinkService) HandleWebhook(ctx context.Context, body []byte, signature string) error {
	clinicID, ok := s.verifySignatureAnyClinic(ctx, body, signature)
	if !ok {
		return apperrors.WrapInvalidInput("invalid line signature")
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return apperrors.WrapInvalidInput("invalid webhook body")
	}

	receivedAt := time.Now()
	var eventErrors []error
	for _, event := range payload.Events {
		lineUserID := event.Source.UserID
		if lineUserID == "" {
			continue
		}
		if event.Type != "follow" && event.Type != "unfollow" {
			continue
		}
		eventAt, err := lineWebhookEventTime(event.Timestamp, receivedAt)
		if err != nil {
			eventErrors = append(eventErrors, err)
			continue
		}
		switch event.Type {
		case "follow":
			if err := s.handleFollowEvent(ctx, clinicID, lineUserID, eventAt); err != nil {
				eventErrors = append(eventErrors, err)
			}
		case "unfollow":
			if err := s.handleUnfollowEvent(ctx, clinicID, lineUserID, eventAt); err != nil {
				eventErrors = append(eventErrors, err)
			}
		}
	}
	return errors.Join(eventErrors...)
}

func lineWebhookEventTime(timestamp int64, receivedAt time.Time) (time.Time, error) {
	if timestamp <= 0 {
		return time.Time{}, apperrors.WrapInvalidInput("invalid LINE webhook event timestamp")
	}
	eventAt := time.UnixMilli(timestamp)
	if eventAt.After(receivedAt.Add(maxLineWebhookFutureSkew)) {
		return time.Time{}, apperrors.WrapInvalidInput("invalid LINE webhook event timestamp")
	}
	return eventAt, nil
}

func (s *lineLinkService) handleFollowEvent(
	ctx context.Context,
	clinicID uint64,
	lineUserID string,
	eventAt time.Time,
) error {
	owner, err := s.ownerRepo.FindByLineUserID(ctx, clinicID, lineUserID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil
		}
		return apperrors.Wrap(err, "failed to find owners by line user id")
	}
	if owner == nil || owner.ClinicID != clinicID {
		return apperrors.WrapInternalServerError("LINE webhook owner scope mismatch")
	}
	if _, err := s.ownerRepo.UpdateLineFollowedAt(ctx, clinicID, owner.ID, lineUserID, eventAt); err != nil {
		return apperrors.Wrap(err, "failed to update line_followed_at")
	}
	return nil
}

func (s *lineLinkService) handleUnfollowEvent(
	ctx context.Context,
	clinicID uint64,
	lineUserID string,
	eventAt time.Time,
) error {
	owner, err := s.ownerRepo.FindByLineUserID(ctx, clinicID, lineUserID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil
		}
		return apperrors.Wrap(err, "failed to find owners by line user id")
	}
	if owner == nil || owner.ClinicID != clinicID {
		return apperrors.WrapInternalServerError("LINE webhook owner scope mismatch")
	}
	if _, err := s.ownerRepo.UpdateLineBlockedAt(ctx, clinicID, owner.ID, lineUserID, eventAt); err != nil {
		return apperrors.Wrap(err, "failed to update line_blocked_at")
	}
	return nil
}

// verifySignatureAnyClinic は webhook body の destination（LINE bot user ID）で
// 対象 clinic を1件に絞り、canonical clinic_integrations credential だけで HMAC を検証する。
//
// SEC-CS-F05-R1: FindAll + 全 clinic HMAC は禁止。固定コストは
// destination 抽出 → route metadata → canonical credential → 最大1回 decrypt → 最大1回 HMAC。
// セマフォと decrypt キャッシュは二次 backpressure として維持する。
func (s *lineLinkService) verifySignatureAnyClinic(
	ctx context.Context,
	body []byte,
	signature string,
) (uint64, bool) {
	if err := lineWebhookVerifySem.Acquire(ctx, 1); err != nil {
		slog.ErrorContext(ctx, "LINE webhook verification concurrency limit", "error", err)
		return 0, false
	}
	defer lineWebhookVerifySem.Release(1)

	destination, ok := extractLineWebhookDestination(body)
	if !ok {
		return 0, false
	}

	if s.lineSettingRepo == nil || s.lineCredentialRepo == nil {
		return 0, false
	}
	clinicID, legacyCredentialPresent, err := s.lineSettingRepo.FindWebhookRouteByLineBotUserID(ctx, destination)
	if err != nil {
		// not found / DB error とも invalid signature として fail-closed（情報漏洩防止）。
		return 0, false
	}
	if clinicID == 0 || legacyCredentialPresent {
		return 0, false
	}

	credential, err := s.lineCredentialRepo.FindCredentialByClinicServiceKey(
		ctx,
		clinicID,
		model.IntegrationServiceLstep,
		model.IntegrationKeyLineChannelSecret,
	)
	if err != nil || credential == nil {
		return 0, false
	}
	if credential.ClinicID != clinicID ||
		credential.Service != model.IntegrationServiceLstep ||
		credential.KeyName != model.IntegrationKeyLineChannelSecret {
		return 0, false
	}

	secret := s.cachedDecryptChannelSecret(ctx, credential)
	if secret == "" {
		return 0, false
	}
	if !lineSignatureVerifier(body, signature, secret) {
		return 0, false
	}
	return clinicID, true
}

// extractLineWebhookDestination は raw webhook body から destination だけを抜く。
// 業務イベントのフルパース前に呼び、欠落・空・過大は reject（秘密計算ゼロ）。
func extractLineWebhookDestination(body []byte) (string, bool) {
	var probe struct {
		Destination string `json:"destination"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return "", false
	}
	dest := strings.TrimSpace(probe.Destination)
	if dest == "" {
		return "", false
	}
	if len(dest) > maxLineWebhookDestinationChars {
		return "", false
	}
	return dest, true
}

// cachedDecryptChannelSecret は canonical integration ID をキーに復号済み secret を短 TTL キャッシュする。
// ID が 0（未採番のテスト行など）の場合はキャッシュせず都度復号する。
func (s *lineLinkService) cachedDecryptChannelSecret(
	ctx context.Context,
	credential *model.ClinicIntegration,
) string {
	if credential == nil {
		return ""
	}
	ciphertext := credential.KeyValue
	if ciphertext == "" {
		return ""
	}
	now := time.Now()
	if credential.ID != 0 {
		if plaintext, ok := s.lookupSecretCache(credential.ID, ciphertext, now); ok {
			return plaintext
		}
	}

	plaintext := lineCredentialDecrypt(ctx, s.cipher, ciphertext)
	if credential.ID == 0 || plaintext == "" {
		return plaintext
	}

	s.secretCacheMu.Lock()
	if s.secretCache == nil {
		s.secretCache = make(map[uint64]lineChannelSecretCacheEntry)
	}
	s.secretCache[credential.ID] = lineChannelSecretCacheEntry{
		ciphertext: ciphertext,
		plaintext:  plaintext,
		expiresAt:  now.Add(lineChannelSecretCacheTTL),
	}
	s.secretCacheMu.Unlock()
	return plaintext
}

func (s *lineLinkService) lookupSecretCache(id uint64, ciphertext string, now time.Time) (string, bool) {
	s.secretCacheMu.Lock()
	defer s.secretCacheMu.Unlock()
	entry, ok := s.secretCache[id]
	if !ok {
		return "", false
	}
	if entry.ciphertext == ciphertext && now.Before(entry.expiresAt) {
		return entry.plaintext, true
	}
	delete(s.secretCache, id)
	return "", false
}

// verifyLineSignature は LINE HMAC-SHA256 署名を検証する。
func verifyLineSignature(body []byte, signature, channelSecret string) bool {
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

// verifyLineIDToken は LINE API でIDトークンを検証し LINE User ID を返す。
func verifyLineIDToken(ctx context.Context, idToken string, clinicID uint64, settingRepo lstepLineSettingReader, client *http.Client) (string, error) {
	setting, err := settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to get line channel id")
	}
	if client == nil {
		client = &http.Client{Timeout: lineVerifyHTTPTimeout}
	}
	safeClient := *client
	safeClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	form := url.Values{
		"id_token":  {idToken},
		"client_id": {setting.LineChannelID},
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		line.VerifyEndpoint,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to create LINE verify request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := safeClient.Do(req)
	if err != nil {
		return "", errors.Join(
			apperrors.WrapBadGateway("LINE ID token verification request failed"),
			err,
		)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxLineVerifyResponseBytes+1))
	if err != nil {
		return "", errors.Join(
			apperrors.WrapBadGateway("failed to read LINE verification response"),
			err,
		)
	}
	if len(bodyBytes) > maxLineVerifyResponseBytes {
		return "", apperrors.WrapBadGateway("LINE verification response exceeds size limit")
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
			return "", apperrors.WrapUnauthorized("invalid LINE ID token")
		}
		return "", apperrors.WrapBadGateway("LINE ID token verification failed")
	}

	var result struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", errors.Join(
			apperrors.WrapBadGateway("invalid LINE verification response"),
			err,
		)
	}
	if result.Sub == "" || len(result.Sub) > maxLineUserIDChars {
		return "", apperrors.WrapBadGateway("invalid LINE verification response")
	}
	return result.Sub, nil
}
