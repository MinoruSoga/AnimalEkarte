package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetMe returns the authenticated staff profile and effective permissions.
func (h *HTTPHandler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()
	userIDValue, exists := c.Get("user_id")
	if !exists {
		httpapi.RespondError(c, apperrors.WrapUnauthorized("missing user context"))
		return
	}
	userID, ok := userIDValue.(string)
	if !ok {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("invalid user context"))
		return
	}
	mainClinicIDValue, _ := c.Get("clinic_id")
	mainClinicID, _ := mainClinicIDValue.(string)

	parsedUserID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("invalid user id"))
		return
	}
	staff, err := h.deps.Staff.GetByID(ctx, parsedUserID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	var account *model.Account
	if staff.AccountID != nil {
		account, err = h.deps.Accounts.GetByID(ctx, *staff.AccountID)
		if err != nil {
			httpapi.RespondError(c, err)
			return
		}
	}

	assignments, err := h.deps.StaffAssignments.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to find clinic assignments",
			"error", err,
			"staff_id", staff.ID,
		)
	} else {
		staff = withClinicAssignments(staff, assignments)
	}

	allClinics, err := h.deps.Clinics.ListClinics(ctx)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	isSystemAdmin := account != nil && account.IsSystemAdmin
	c.JSON(
		http.StatusOK,
		h.BuildMeResponse(c, staff, account, mainClinicID, isSystemAdmin, allClinics),
	)
}

// AuthenticateUser validates credentials and records attributable password failures.
func (h *HTTPHandler) AuthenticateUser(
	ctx context.Context,
	email, password, clientIP, userAgent string,
) (*model.Account, *model.Staff, error) {
	timing := h.loginFailureTiming.withDefaults()
	startedAt := timing.now()
	account, staff, err := h.authService().AuthenticateUser(ctx, email, password)
	if err != nil {
		if errors.Is(err, apperrors.ErrUnauthorized) {
			accountID, knownAccount := IsAuthenticateWrongPassword(err)
			h.enqueueLoginFailureAudit(
				ctx,
				accountID,
				knownAccount,
				clientIP,
				userAgent,
			)
			timing.wait(ctx, startedAt)
		}
		return nil, nil, err
	}
	return account, staff, nil
}

func withClinicAssignments(
	staff *model.Staff,
	assignments []model.StaffClinicAssignment,
) *model.Staff {
	if staff == nil {
		return nil
	}
	result := *staff
	result.ClinicAssignments = append(
		[]model.StaffClinicAssignment(nil),
		assignments...,
	)
	return &result
}

// AuditKnownAccountLoginFailure writes an attributable wrong-password audit event.
func (h *HTTPHandler) AuditKnownAccountLoginFailure(
	ctx context.Context,
	accountID uint64,
	clientIP, userAgent string,
) {
	if h.deps.Staff == nil || h.deps.StaffAssignments == nil || h.deps.Audit == nil {
		return
	}
	staff, err := h.deps.Staff.FindByAccountID(ctx, accountID)
	if err != nil {
		slog.WarnContext(
			ctx,
			"skip audit log for login failure: staff not resolved",
			"account_id", accountID,
			"error_type", fmt.Sprintf("%T", err),
		)
		return
	}
	assignments, err := h.deps.StaffAssignments.FindAllByStaffID(ctx, staff.ID)
	if err != nil {
		slog.WarnContext(
			ctx,
			"skip audit log for login failure: clinic assignments not resolved",
			"staff_id", staff.ID,
			"error_type", fmt.Sprintf("%T", err),
		)
		return
	}
	clinicID, ok := auditClinicIDFromStaffAssignments(staff.ID, assignments)
	if !ok {
		clinicID, ok = h.systemAdminAuditClinicID(ctx, accountID)
		if !ok {
			slog.WarnContext(
				ctx,
				"skip audit log for login failure: no attributable clinic",
				"staff_id", staff.ID,
				"account_id", accountID,
			)
			return
		}
	}
	if err := h.deps.Audit.LogAuthLogin(
		ctx,
		&clinicID,
		&staff.ID,
		model.AuditActionAuthLoginFailure,
		clientIP,
		userAgent,
	); err != nil {
		slog.ErrorContext(
			ctx,
			"audit log failed for login failure",
			"staff_id", staff.ID,
			"clinic_id", clinicID,
			"error_type", fmt.Sprintf("%T", err),
		)
	}
}

func (h *HTTPHandler) systemAdminAuditClinicID(
	ctx context.Context,
	accountID uint64,
) (uint64, bool) {
	if h.deps.Accounts == nil || h.deps.Clinics == nil {
		return 0, false
	}
	account, err := h.deps.Accounts.GetByID(ctx, accountID)
	if err != nil {
		slog.WarnContext(
			ctx,
			"skip system administrator audit clinic resolution: account unavailable",
			"account_id", accountID,
			"error_type", fmt.Sprintf("%T", err),
		)
		return 0, false
	}
	if account == nil ||
		account.ID != accountID ||
		!account.IsActive ||
		account.DeletedAt.Valid ||
		!account.IsSystemAdmin {
		return 0, false
	}
	clinics, err := h.deps.Clinics.ListClinics(ctx)
	if err != nil {
		slog.WarnContext(
			ctx,
			"skip system administrator audit clinic resolution: clinics unavailable",
			"account_id", accountID,
			"error_type", fmt.Sprintf("%T", err),
		)
		return 0, false
	}
	for i := range clinics {
		if clinics[i].ID != 0 && clinics[i].IsActive {
			return clinics[i].ID, true
		}
	}
	return 0, false
}
