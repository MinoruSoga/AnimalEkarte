package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/crypto"
	"github.com/animal-ekarte/backend/internal/infra/line"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
	Force       bool   `json:"force"`
}

// WebhookEvent は LINE Webhook イベントを表す最小構造体。
type WebhookEvent struct {
	Type   string `json:"type"`
	Source struct {
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
	ownerRepo         repository.OwnerRepository
	lineLinkTokenRepo repository.LineLinkTokenRepository
	lineSettingRepo   repository.LineReservationSettingRepository
	auditSvc          AuditService
	// cipher は Webhook 署名検証時に line_channel_secret を復号するために使う（H-4）。
	// nil の場合は復号なしで動作する（開発環境で INTEGRATION_ENCRYPTION_KEY 未設定時）。
	cipher *crypto.AESGCMCipher
	// httpClient は LINE ID Token 検証 API 呼び出しに使う。テスト容易性のためのシームで、
	// 本番では http.DefaultClient と等価に振る舞う（挙動変更なし）。
	httpClient *http.Client
}

// NewLineLinkService は LineLinkService を初期化して返す。
// cipher が nil の場合は復号なしで動作する（lstep 連携と同一の cipher を再利用する）。
func NewLineLinkService(
	ownerRepo repository.OwnerRepository,
	lineLinkTokenRepo repository.LineLinkTokenRepository,
	lineSettingRepo repository.LineReservationSettingRepository,
	auditSvc AuditService,
	cipher *crypto.AESGCMCipher,
) LineLinkService {
	return &lineLinkService{
		ownerRepo:         ownerRepo,
		lineLinkTokenRepo: lineLinkTokenRepo,
		lineSettingRepo:   lineSettingRepo,
		auditSvc:          auditSvc,
		cipher:            cipher,
		httpClient:        http.DefaultClient,
	}
}

// GenerateLinkToken は飼い主向けの一時トークンを生成し、LIFF URL を返す（24時間有効）。
func (s *lineLinkService) GenerateLinkToken(ctx context.Context, clinicID, ownerID uint64) (*LinkTokenResult, error) {
	// 所有者の存在確認
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner for link token", "error", err)
		return nil, apperrors.Wrap(err, "failed to find owner")
	}

	// 64バイトのランダムトークン生成
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, apperrors.Wrap(fmt.Errorf("failed to generate token: %w", err), "internal error")
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(24 * time.Hour)

	t := &model.LineLinkToken{
		ClinicID:  clinicID,
		OwnerID:   owner.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := s.lineLinkTokenRepo.Create(ctx, t); err != nil {
		slog.ErrorContext(ctx, "failed to create link token", "error", err)
		return nil, apperrors.Wrap(err, "failed to create link token")
	}

	// LIFF URL 構築（clinic の LIFF ID を使用）
	setting, err := s.lineSettingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find line setting for liff url", "error", err)
		return nil, apperrors.Wrap(err, "failed to find line reservation setting")
	}

	liffURL := ""
	if setting.LiffID != "" {
		liffURL = fmt.Sprintf("https://liff.line.me/%s?token=%s", setting.LiffID, token)
	}

	return &LinkTokenResult{
		Token:     token,
		ExpiresAt: expiresAt,
		LiffURL:   liffURL,
	}, nil
}

// LinkAccount は LINE ID Token を検証してトークン対応の飼い主に LINE User ID を紐付ける。
func (s *lineLinkService) LinkAccount(ctx context.Context, clinicID uint64, input LinkAccountInput) (*model.Owner, error) {
	// 1. LINE ID Token 検証 → LINE User ID 取得
	lineUserID, err := verifyLineIDToken(ctx, input.LineIDToken, clinicID, s.lineSettingRepo, s.httpClient)
	if err != nil {
		return nil, apperrors.WrapUnauthorized(fmt.Sprintf("invalid line id token: %v", err))
	}

	// 2. link_token でオーナーを特定
	lt, err := s.lineLinkTokenRepo.FindByToken(ctx, input.LinkToken)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid or expired link token")
	}
	if lt.ClinicID != clinicID {
		return nil, apperrors.WrapInvalidInput("link token clinic mismatch")
	}

	// 3. 既存 line_user_id チェック
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, lt.OwnerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner for link account", "error", err)
		return nil, apperrors.Wrap(err, "failed to find owner")
	}

	if owner.LineUserID != nil && *owner.LineUserID != "" && !input.Force {
		// 既に別の LINE User ID が設定済み
		return nil, apperrors.WrapConflict("line user id already set")
	}

	// 4. LINE User ID を更新
	if err := s.ownerRepo.UpdateLineUserID(ctx, clinicID, lt.OwnerID, &lineUserID); err != nil {
		slog.ErrorContext(ctx, "failed to update line user id", "error", err)
		return nil, apperrors.Wrap(err, "failed to update line user id")
	}

	// 5. トークン使用済みマーク
	if err := s.lineLinkTokenRepo.MarkUsed(ctx, lt.ID, time.Now()); err != nil {
		slog.ErrorContext(ctx, "failed to mark link token used — aborting to prevent duplicate use", "error", err)
		// トークンのマーク失敗は即座に返す（二重使用リスク排除）
		return nil, apperrors.Wrap(err, "failed to mark link token used")
	}

	// 6. 監査ログ
	ownerID := lt.OwnerID
	if err := s.auditSvc.LogLstepOperation(ctx, clinicID, nil, "link_line_user_id", "owner", &ownerID); err != nil {
		slog.WarnContext(ctx, "audit log failed for line user id link", "error", err, "owner_id", ownerID)
	}

	owner.LineUserID = &lineUserID
	return owner, nil
}

// HandleWebhook は LINE Webhook を処理する。署名検証はハンドラ層で行う。
func (s *lineLinkService) HandleWebhook(ctx context.Context, body []byte, signature string) error {
	// 署名検証（全クリニックのいずれかで検証成功すれば可）
	if !s.verifySignatureAnyClinic(ctx, body, signature) {
		return apperrors.WrapInvalidInput("invalid line signature")
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return apperrors.WrapInvalidInput("invalid webhook body")
	}

	now := time.Now()
	for _, event := range payload.Events {
		lineUserID := event.Source.UserID
		if lineUserID == "" {
			continue
		}
		switch event.Type {
		case "follow":
			if err := s.handleFollowEvent(ctx, lineUserID, now); err != nil {
				slog.ErrorContext(ctx, "failed to handle follow event", "error", err, "line_user_id", lineUserID)
			}
		case "unfollow":
			if err := s.handleUnfollowEvent(ctx, lineUserID, now); err != nil {
				slog.ErrorContext(ctx, "failed to handle unfollow event", "error", err, "line_user_id", lineUserID)
			}
		}
	}
	return nil
}

func (s *lineLinkService) handleFollowEvent(ctx context.Context, lineUserID string, now time.Time) error {
	owners, err := s.ownerRepo.FindAllByLineUserID(ctx, lineUserID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find owners by line user id")
	}
	for i := range owners {
		o := &owners[i]
		if err := s.ownerRepo.UpdateLineFollowedAt(ctx, o.ClinicID, o.ID, now); err != nil {
			slog.ErrorContext(ctx, "failed to update line_followed_at", "error", err, "owner_id", o.ID)
		}
	}
	return nil
}

func (s *lineLinkService) handleUnfollowEvent(ctx context.Context, lineUserID string, now time.Time) error {
	owners, err := s.ownerRepo.FindAllByLineUserID(ctx, lineUserID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find owners by line user id")
	}
	for i := range owners {
		o := &owners[i]
		if err := s.ownerRepo.UpdateLineBlockedAt(ctx, o.ClinicID, o.ID, now); err != nil {
			slog.ErrorContext(ctx, "failed to update line_blocked_at", "error", err, "owner_id", o.ID)
		}
	}
	return nil
}

// verifySignatureAnyClinic は全クリニックの LINE Channel Secret で署名を検証する。
func (s *lineLinkService) verifySignatureAnyClinic(ctx context.Context, body []byte, signature string) bool {
	settings, err := s.lineSettingRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load line settings for signature verification", "error", err)
		return false
	}
	for i := range settings {
		setting := &settings[i]
		// DB 上の line_channel_secret は暗号文（H-4）。レガシー平文行はそのまま返る。
		secret := decryptLineCredential(ctx, s.cipher, setting.LineChannelSecret)
		if secret == "" {
			continue
		}
		if verifyLineSignature(body, signature, secret) {
			return true
		}
	}
	return false
}

// verifyLineSignature は LINE HMAC-SHA256 署名を検証する。
func verifyLineSignature(body []byte, signature, channelSecret string) bool {
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

// verifyLineIDToken は LINE API でIDトークンを検証し LINE User ID を返す。
// client は呼び出しに使う *http.Client（テスト容易性のためのシーム）。nil の場合は http.DefaultClient を使う。
func verifyLineIDToken(ctx context.Context, idToken string, clinicID uint64, settingRepo repository.LineReservationSettingRepository, client *http.Client) (string, error) {
	setting, err := settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to get line channel id")
	}
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.PostForm(line.VerifyEndpoint, url.Values{
		"id_token":  {idToken},
		"client_id": {setting.LineChannelID},
	})
	if err != nil {
		return "", apperrors.Wrap(err, "line id token verify request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to read verify response")
	}

	if resp.StatusCode != http.StatusOK {
		return "", apperrors.WrapInvalidInput(fmt.Sprintf("line id token verify failed: status=%d body=%s", resp.StatusCode, string(bodyBytes)))
	}

	var result struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", apperrors.Wrap(err, "failed to parse verify response")
	}
	if result.Sub == "" {
		return "", apperrors.WrapInvalidInput("empty sub in verify response")
	}
	return result.Sub, nil
}
