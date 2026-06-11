package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/model"
)

// authenticateUser はメール/パスワードを検証してアカウントとスタッフを返す。
// 認証失敗・アカウント無効・スタッフ無効の場合は apperrors.ErrUnauthorized ラップエラーを返す。
func (h *Handler) authenticateUser(ctx context.Context, email, password, clientIP, userAgent string) (*model.Account, *model.Staff, error) {
	account, err := h.svc.Account.FindByEmail(ctx, email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			// 監査ログ: ログイン失敗（アカウント不存在）
			if logErr := h.svc.Audit.LogAuthLogin(ctx, nil, nil, model.AuditActionAuthLoginFailure, clientIP, userAgent); logErr != nil {
				slog.ErrorContext(ctx, "audit log failed for login failure (account not found)", "error", logErr)
			}
			return nil, nil, apperrors.WrapUnauthorized("メールアドレスまたはパスワードが正しくありません")
		}
		return nil, nil, err
	}
	if !account.IsActive {
		return nil, nil, apperrors.WrapUnauthorized("アカウントが無効です")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		// 監査ログ: ログイン失敗（パスワード不正）
		if logErr := h.svc.Audit.LogAuthLogin(ctx, nil, &account.ID, model.AuditActionAuthLoginFailure, clientIP, userAgent); logErr != nil {
			slog.ErrorContext(ctx, "audit log failed for login failure (invalid password)", "account_id", account.ID, "error", logErr)
		}
		return nil, nil, apperrors.WrapUnauthorized("メールアドレスまたはパスワードが正しくありません")
	}
	staff, err := h.svc.Staff.FindByAccountID(ctx, account.ID)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "スタッフ情報の取得に失敗しました")
	}
	// BUG-134: スタッフの有効性チェック（退職者・一時停止アカウント防止）
	if !staff.IsActive {
		return nil, nil, apperrors.WrapUnauthorized("このアカウントは無効です")
	}
	return account, staff, nil
}

// resolveClinicInfo はスタッフのクリニック割り当てからメインクリニックIDと所属ID一覧を返す。
// メインクリニックが未設定の場合は最初の割り当てを使用する。
func resolveClinicInfo(assignments []model.StaffClinicAssignment) (mainClinicID string, clinicIDs []uint64) {
	clinicIDs = make([]uint64, 0, len(assignments))
	for i := range assignments {
		asg := &assignments[i]
		if asg.IsMain && mainClinicID == "" {
			mainClinicID = strconv.FormatUint(asg.ClinicID, 10)
		}
		clinicIDs = append(clinicIDs, asg.ClinicID)
	}
	if mainClinicID == "" && len(assignments) > 0 {
		mainClinicID = strconv.FormatUint(assignments[0].ClinicID, 10)
	}
	return mainClinicID, clinicIDs
}

// resolveSystemAdminMainClinicID は system_admin で assignments なしの場合に
// allClinics[0] を main にフォールバックする。それ以外は元の mainClinicID を返す。
func resolveSystemAdminMainClinicID(mainClinicID string, isSystemAdmin bool, allClinics []model.Clinic) string {
	if mainClinicID != "" {
		return mainClinicID
	}
	if !isSystemAdmin || len(allClinics) == 0 {
		return mainClinicID
	}
	return strconv.FormatUint(allClinics[0].ID, 10)
}

// issueAuthCookies は JWT アクセストークン（15分）とリフレッシュトークン（7日）を生成して Cookie にセットする。
// クロスオリジン対応のため SameSite=None + Secure=true を使用する。
func (h *Handler) issueAuthCookies(c *gin.Context, staffID uint64, mainClinicID string, isSystemAdmin bool, clinicIDs []uint64) error {
	// BUG-TEST-002: 開発環境（HTTP）での Cookie 受理を許可するため、モードに応じて Secure/SameSite を切り替える
	isProduction := h.cfg.GinMode == "release"
	sameSite := http.SameSiteLaxMode
	secure := false

	if isProduction {
		sameSite = http.SameSiteNoneMode
		secure = true
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	claims := &middleware.JWTClaims{
		UserID:        strconv.FormatUint(staffID, 10),
		ClinicID:      mainClinicID,
		IsSystemAdmin: isSystemAdmin,
		ClinicIDs:     clinicIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenStr, err := accessToken.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return apperrors.Wrap(err, "failed to sign jwt")
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    accessTokenStr,
		Path:     "/",
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	// Refresh Token 発行（7日間有効、token rotation で毎回更新）
	// jti はサーバーサイド失効（ブラックリスト照合）に使用する。
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := &middleware.JWTClaims{
		UserID:        strconv.FormatUint(staffID, 10),
		ClinicID:      mainClinicID,
		IsSystemAdmin: isSystemAdmin,
		ClinicIDs:     clinicIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "refresh",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return apperrors.Wrap(err, "failed to sign refresh token")
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshTokenStr,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   int(time.Until(refreshExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
	return nil
}
