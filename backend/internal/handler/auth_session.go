package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// clearCookie は指定した Cookie を即時失効させる（MaxAge=-1, Value=""）ヘルパー
// （E-2: Logout の cookie クリアブロックの共通化）。
func clearCookie(c *gin.Context, name, path string, secure bool, sameSite http.SameSite) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func newJti() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(bytes[:])
}

func auditClinicIDFromAssignments(assignments []model.StaffClinicAssignment) (uint64, bool) {
	if len(assignments) == 0 {
		return 0, false
	}
	fallback := assignments[0].ClinicID
	for i := range assignments {
		if assignments[i].IsMain {
			return assignments[i].ClinicID, true
		}
	}
	return fallback, true
}

func (h *Handler) auditKnownAccountLoginFailure(ctx context.Context, accountID uint64, clientIP, userAgent string) {
	staff, err := h.svc.Staff.FindByAccountID(ctx, accountID)
	if err != nil {
		slog.WarnContext(ctx, "skip audit log for login failure: staff not resolved", "account_id", accountID, "error", err)
		return
	}
	assignments, err := h.svc.StaffClinicAssignment.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		slog.WarnContext(ctx, "skip audit log for login failure: clinic assignments not resolved", "staff_id", staff.ID, "error", err)
		return
	}
	clinicID, ok := auditClinicIDFromAssignments(assignments)
	if !ok {
		slog.WarnContext(ctx, "skip audit log for login failure: staff has no clinic assignments", "staff_id", staff.ID)
		return
	}
	if logErr := h.svc.Audit.LogAuthLogin(ctx, &clinicID, &staff.ID, model.AuditActionAuthLoginFailure, clientIP, userAgent); logErr != nil {
		slog.ErrorContext(ctx, "audit log failed for login failure", "staff_id", staff.ID, "clinic_id", clinicID, "error", logErr)
	}
}

// authenticateUser はメール/パスワードを検証してアカウントとスタッフを返す。
// 認証失敗・アカウント無効・スタッフ無効の場合は apperrors.ErrUnauthorized ラップエラーを返す。
func (h *Handler) authenticateUser(ctx context.Context, email, password, clientIP, userAgent string) (*model.Account, *model.Staff, error) {
	account, staff, err := h.authSvc().AuthenticateUser(ctx, email, password)
	if err != nil {
		if accountID, ok := service.IsAuthenticateWrongPassword(err); ok {
			// 監査ログ: ログイン失敗（パスワード不正）
			h.auditKnownAccountLoginFailure(ctx, accountID, clientIP, userAgent)
		}
		return nil, nil, err
	}
	return account, staff, nil
}

// resolveClinicInfo はスタッフのクリニック割り当てからメインクリニックIDと所属ID一覧を返す。
// テスト互換ヘルパ。本番 Handler 経路は h.authSvc() 経由で呼ぶ。
func resolveClinicInfo(assignments []model.StaffClinicAssignment) (mainClinicID string, clinicIDs []uint64) {
	return authServiceNoDeps().ResolveClinicInfo(assignments)
}

// resolveSystemAdminMainClinicID は system_admin で assignments なしの場合に
// allClinics[0] を main にフォールバックする。それ以外は元の mainClinicID を返す。
// テスト互換ヘルパ。本番 Handler 経路は h.authSvc() 経由で呼ぶ。
func resolveSystemAdminMainClinicID(mainClinicID string, isSystemAdmin bool, allClinics []model.Clinic) string {
	return authServiceNoDeps().ResolveSystemAdminMainClinicID(mainClinicID, isSystemAdmin, allClinics)
}

// authServiceNoDeps は依存なし純関数メソッド（ResolveClinicInfo 等）とテスト向けの単一フォールバック。
// 本番 Handler 経路は authSvc()（注入済み Auth 優先）を使うこと。
func authServiceNoDeps() service.AuthService {
	return service.NewAuthService(nil, nil, nil)
}

// authSvc は認証・権限計算用 AuthService を返す。
// 単体テストが部分 DI の Services を構築する場合に Auth 未設定でも動作するようフォールバックする。
func (h *Handler) authSvc() service.AuthService {
	if h.svc != nil && h.svc.Auth != nil {
		return h.svc.Auth
	}
	var account service.AccountService
	var staff service.StaffService
	var effectivePerm service.EffectivePermissionService
	if h.svc != nil {
		account = h.svc.Account
		staff = h.svc.Staff
		effectivePerm = h.svc.EffectivePermission
	}
	return service.NewAuthService(account, staff, effectivePerm)
}

// tokenSvc は JWT 発行・検証用 TokenService を返す。
// 単体テストが部分 DI の Services を構築する場合に Token 未設定でも動作するようフォールバックする。
func (h *Handler) tokenSvc() service.TokenService {
	if h.svc != nil && h.svc.Token != nil {
		return h.svc.Token
	}
	var blacklist service.TokenBlacklistService
	if h.svc != nil {
		blacklist = h.svc.TokenBlacklist
	}
	return service.NewTokenService(h.cfg.JWTSecret, blacklist)
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

	accessIssued, err := h.tokenSvc().IssueAccessToken(staffID, mainClinicID, isSystemAdmin, clinicIDs)
	if err != nil {
		return err
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    accessIssued.Token,
		Path:     "/",
		MaxAge:   int(time.Until(accessIssued.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	// Refresh Token 発行（7日間有効、token rotation で毎回更新）
	// jti はサーバーサイド失効（ブラックリスト照合）に使用する。
	refreshIssued, err := h.tokenSvc().IssueRefreshToken(staffID, mainClinicID, isSystemAdmin, clinicIDs)
	if err != nil {
		return err
	}
	// 旧pathの同名cookieが残るとrefresh endpointへ2値送信され得るため、移行時に明示削除する。
	clearCookie(c, refreshTokenCookieName, legacyRefreshTokenCookiePath, secure, sameSite)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshIssued.Token,
		Path:     refreshTokenCookiePath,
		MaxAge:   int(time.Until(refreshIssued.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
	return nil
}
