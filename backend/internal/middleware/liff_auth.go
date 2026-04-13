package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	liffCustomerIDKey = "liff_customer_id"
	liffClinicIDKey   = "liff_clinic_id"
	liffLineUserIDKey = "liff_line_user_id"
	liffDisplayName   = "liff_display_name"
)

// lineVerifyResponse は LINE ID Token 検証 API のレスポンス。
type lineVerifyResponse struct {
	Sub       string `json:"sub"`  // line_user_id
	Name      string `json:"name"` // 表示名
	Picture   string `json:"picture"`
	ClientID  string `json:"client_id"`
	ExpiresIn int    `json:"expires_in"`
	Error     string `json:"error"`
	ErrorDesc string `json:"error_description"`
}

// LineCustomerLookup はLINE顧客を特定・作成するためのインターフェース。
type LineCustomerLookup interface {
	FindOrCreateByLineUserID(ctx context.Context, clinicID uint64, lineUserID, displayName string) (*model.LineCustomer, error)
}

// LineReservationSettingLookup はLIFF認証でクリニックの設定を取得するインターフェース。
type LineReservationSettingLookup interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
}

// LiffAuth はLIFF ID Tokenを検証してcontext に顧客情報をセットするミドルウェア。
// settingLookup でクリニックの LiffID を取得し、LINE API の client_id 照合に使用する。
func LiffAuth(lookup LineCustomerLookup, settingLookup LineReservationSettingLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ローカル開発用: LIFF_MOCK=true で認証バイパス
		if os.Getenv("LIFF_MOCK") == "true" {
			clinicIDStr := c.Param("clinicId")
			var clinicID uint64
			if _, err := fmt.Sscanf(clinicIDStr, "%d", &clinicID); err == nil && clinicID > 0 {
				customer, err := lookup.FindOrCreateByLineUserID(c.Request.Context(), clinicID, "mock-line-user-id", "テストユーザー")
				if err == nil {
					c.Set(liffCustomerIDKey, customer.ID)
					c.Set(liffClinicIDKey, clinicID)
					c.Set(liffLineUserIDKey, "mock-line-user-id")
					c.Set(liffDisplayName, "テストユーザー")
					c.Next()
					return
				}
			}
		}

		// Authorization: Bearer {ID Token}
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}
		idToken := parts[1]

		// clinicID を URL パスから取得（/api/liff/:clinicId/...）
		clinicIDStr := c.Param("clinicId")
		var clinicID uint64
		if _, err := fmt.Sscanf(clinicIDStr, "%d", &clinicID); err != nil || clinicID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clinic id"})
			c.Abort()
			return
		}

		// DB からクリニックの LiffID を取得して LINE チャンネル ID を特定する
		setting, err := settingLookup.FindByClinicID(c.Request.Context(), clinicID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load clinic setting"})
			c.Abort()
			return
		}
		if setting.LiffID == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "LINE reservation not configured for this clinic"})
			c.Abort()
			return
		}

		// LIFF ID から LINEログインチャンネル ID を抽出する
		// LIFF ID 形式: "{channelID}-{appID}"（例: "2009755586-nvKfG3Cp"）
		liffChannelID := strings.SplitN(setting.LiffID, "-", 2)[0]

		// LINE API でトークン検証（client_id にクリニックのチャンネル ID を使用）
		lineUser, err := verifyLiffIDToken(c.Request.Context(), idToken, liffChannelID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid ID token: " + err.Error()})
			c.Abort()
			return
		}

		// line_customers から顧客を特定（なければ新規作成）
		customer, err := lookup.FindOrCreateByLineUserID(c.Request.Context(), clinicID, lineUser.Sub, lineUser.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve customer"})
			c.Abort()
			return
		}

		// context にセット
		c.Set(liffCustomerIDKey, customer.ID)
		c.Set(liffClinicIDKey, clinicID)
		c.Set(liffLineUserIDKey, lineUser.Sub)
		c.Set(liffDisplayName, lineUser.Name)
		c.Next()
	}
}

// lineVerifyURL は LINE ID Token 検証 API のエンドポイント。
// テスト時のみ差し替えること（本番では変更しない）。
var lineVerifyURL = "https://api.line.me/oauth2/v2.1/verify"

// verifyLiffIDToken は LINE の ID Token 検証 API を呼び出してユーザー情報を返す。
// clientID にはクリニックの LINEログインチャンネル ID を渡すこと。
func verifyLiffIDToken(ctx context.Context, idToken, clientID string) (*lineVerifyResponse, error) {
	params := url.Values{}
	params.Set("id_token", idToken)
	params.Set("client_id", clientID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, lineVerifyURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create http request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("line verify request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // レスポンスボディのクローズ失敗は復旧不可のため無視

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read line verify response: %w", err)
	}

	var result lineVerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse line verify response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("line verify error: %s", result.ErrorDesc)
	}
	if result.Sub == "" {
		return nil, apperrors.WrapInvalidInput("empty line user id from token")
	}

	return &result, nil
}

// LiffCustomerIDKey はコンテキストキー名を返す（ハンドラーテストでのセットアップ用）。
func LiffCustomerIDKey() string { return liffCustomerIDKey }

// ExtractLiffCustomerID はコンテキストから liff_customer_id を取得する。
func ExtractLiffCustomerID(c *gin.Context) (uint64, bool) {
	v, exists := c.Get(liffCustomerIDKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}

// ExtractLiffClinicID はコンテキストから liff_clinic_id を取得する。
func ExtractLiffClinicID(c *gin.Context) (uint64, bool) {
	v, exists := c.Get(liffClinicIDKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}

// ExtractLiffDisplayName はコンテキストから LINE 表示名を取得する。
func ExtractLiffDisplayName(c *gin.Context) string {
	v, _ := c.Get(liffDisplayName)
	s, _ := v.(string)
	return s
}
