package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// RefreshToken verifies, rotates, revokes, and reissues the refresh token.
func (h *HTTPHandler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	token, err := c.Cookie(RefreshTokenCookieName)
	if err != nil || token == "" {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("refresh token not found"))
		return
	}

	claims, err := h.deps.Tokens.VerifyRefreshToken(ctx, token)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	staffID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("invalid token"))
		return
	}
	staff, err := h.deps.Staff.GetByID(ctx, staffID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			httpapi.RespondError(c, apperrors.Wrap(err, "failed to get refresh staff"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapUnauthorized("user not found"))
		return
	}
	if staff == nil ||
		!staff.IsActive ||
		staff.DeletedAt.Valid ||
		staff.AccountID == nil {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("invalid refresh identity"))
		return
	}

	account, err := h.deps.Accounts.GetByID(ctx, *staff.AccountID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			httpapi.RespondError(c, apperrors.Wrap(err, "failed to get refresh account"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapUnauthorized("invalid refresh identity"))
		return
	}
	if account == nil ||
		account.ID != *staff.AccountID ||
		!account.IsActive ||
		account.DeletedAt.Valid {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("invalid refresh identity"))
		return
	}
	accountEpoch := account.UpdatedAt.UnixNano()
	if !TokenMatchesAccountEpoch(claims, accountEpoch) {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("refresh session has expired"))
		return
	}

	assignments, err := h.deps.StaffAssignments.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to get clinic assignments"))
		return
	}
	mainClinicID, clinicIDs := h.authService().ResolveClinicInfo(assignments)
	if account.IsSystemAdmin {
		allClinics, listErr := h.deps.Clinics.ListClinics(ctx)
		if listErr != nil {
			httpapi.RespondError(c, apperrors.Wrap(listErr, "failed to get clinics"))
			return
		}
		mainClinicID = h.authService().ResolveSystemAdminMainClinicID(
			mainClinicID,
			account.IsSystemAdmin,
			allClinics,
		)
		clinicIDs = activeSystemAdminClinicIDs(allClinics)
	}
	if mainClinicID == "" {
		httpapi.RespondError(c, apperrors.WrapForbidden("no clinic access is available"))
		return
	}

	familyID, validFamily := refreshTokenFamilyID(claims)
	if !validFamily {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("invalid refresh token claims"))
		return
	}
	if h.deps.TokenBlacklist == nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("token validation unavailable"))
		return
	}
	if err := h.deps.TokenBlacklist.RevokeToken(
		ctx,
		claims.ID,
		claims.ExpiresAt.Time,
	); err != nil {
		h.handleRefreshRevokeFailure(c, ctx, familyID, err)
		return
	}

	if err := h.issueAuthCookies(
		c,
		staff.ID,
		mainClinicID,
		account.IsSystemAdmin,
		clinicIDs,
		accountEpoch,
		familyID,
		claims.ExpiresAt.Time,
	); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "token refreshed"})
}

func (h *HTTPHandler) handleRefreshRevokeFailure(c *gin.Context, ctx context.Context, familyID string, err error) {
	if !apperrors.IsAlreadyExists(err) {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to revoke old refresh token"))
		return
	}
	if familyErr := revokeRefreshTokenFamily(
		ctx,
		h.deps.TokenBlacklist,
		familyID,
	); familyErr != nil {
		slog.ErrorContext(ctx, "concurrent refresh family revocation failed")
		httpapi.RespondError(c, familyErr)
		return
	}
	httpapi.RespondError(c, apperrors.WrapUnauthorized("refresh token reuse detected"))
}
