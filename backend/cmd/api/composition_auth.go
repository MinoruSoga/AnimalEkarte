package main

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/auth"
)

// authRepositories are auth-owned persistence capabilities. Keeping this
// bundle domain-local lets staff consume narrow ports without recreating the
// former cross-domain Repositories container.
type authRepositories struct {
	Accounts            auth.AccountRepository
	PermissionGroups    auth.PermissionGroupRepository
	TokenBlacklist      auth.TokenBlacklistRepository
	PasswordResetTokens auth.PasswordResetTokenRepository
	CurrentAccessStaff  auth.CurrentAccessStaffReader
}

func newAuthRepositories(db *gorm.DB) authRepositories {
	return authRepositories{
		Accounts:            auth.NewAccountRepository(db),
		PermissionGroups:    auth.NewPermissionGroupRepository(db),
		TokenBlacklist:      auth.NewTokenBlacklistRepository(db),
		PasswordResetTokens: auth.NewPasswordResetTokenRepository(db),
		CurrentAccessStaff:  auth.NewCurrentAccessStaffReader(db),
	}
}

type authStaffServices interface {
	auth.HTTPStaffReader
	auth.StaffAccountFinder
}

type authCompositionDependencies struct {
	Transactor          auth.Transactor
	JWTSecret           string
	PasswordResetConfig auth.PasswordResetConfig
	CookieConfig        auth.CookieConfig
	Staff               authStaffServices
	StaffAssignments    auth.StaffClinicAssignmentReader
	Clinics             auth.ClinicLister
	Audit               audit.Kernel
}

// authComposition exposes only auth capabilities consumed by the composition
// root and downstream HTTP registration.
type authComposition struct {
	Handler            *auth.HTTPHandler
	Accounts           auth.AccountService
	PermissionGroups   auth.PermissionGroupApplication
	Tokens             auth.TokenService
	TokenBlacklist     auth.TokenBlacklistService
	PasswordReset      auth.PasswordResetService
	CurrentAccess      auth.CurrentAccessResolver
	DrainPasswordReset func()
}

type authServices struct {
	accounts         auth.AccountService
	permissionGroups auth.PermissionGroupApplication
	tokens           auth.TokenService
	tokenBlacklist   auth.TokenBlacklistService
	passwordReset    auth.PasswordResetService
	currentAccess    auth.CurrentAccessResolver
	login            auth.AuthService
}

func newAuthComposition(
	repositories authRepositories,
	dependencies authCompositionDependencies,
) authComposition {
	services := newAuthServices(repositories, dependencies)
	handler := newAuthHTTPHandler(services, dependencies)
	return authComposition{
		Handler:          handler,
		Accounts:         services.accounts,
		PermissionGroups: services.permissionGroups,
		Tokens:           services.tokens,
		TokenBlacklist:   services.tokenBlacklist,
		PasswordReset:    services.passwordReset,
		CurrentAccess:    services.currentAccess,
		DrainPasswordReset: func() {
			handler.Wait()
			services.passwordReset.Wait()
		},
	}
}

func newAuthServices(
	repositories authRepositories,
	dependencies authCompositionDependencies,
) authServices {
	credentialAudit := authCredentialAuditTxAdapter{
		logger: dependencies.Audit,
	}
	accounts := auth.NewAccountServiceWithCredentialAudit(
		repositories.Accounts,
		repositories.PasswordResetTokens,
		dependencies.Transactor,
		credentialAudit,
	)
	permissionGroups := auth.NewPermissionGroupService(
		repositories.PermissionGroups,
		dependencies.Transactor,
		authPermissionAuditTxAdapter{logger: dependencies.Audit},
	)
	tokenBlacklist := auth.NewTokenBlacklistService(
		repositories.TokenBlacklist,
	)
	tokens := auth.NewTokenService(
		dependencies.JWTSecret,
		tokenBlacklist,
	)
	passwordResetConfig := dependencies.PasswordResetConfig
	passwordReset := auth.NewPasswordResetServiceWithCredentialAudit(
		&passwordResetConfig,
		repositories.Accounts,
		repositories.PasswordResetTokens,
		sendAuthPasswordResetMail,
		dependencies.Transactor,
		authCredentialAuditSubjectResolver{
			accounts:    accounts,
			staff:       dependencies.Staff,
			assignments: dependencies.StaffAssignments,
			clinics:     dependencies.Clinics,
		},
		credentialAudit,
	)
	authService := auth.NewAuthService(
		accounts,
		dependencies.Staff,
		permissionGroups,
	)
	return authServices{
		accounts:         accounts,
		permissionGroups: permissionGroups,
		tokens:           tokens,
		tokenBlacklist:   tokenBlacklist,
		passwordReset:    passwordReset,
		currentAccess: auth.NewCurrentAccessResolverWithClinics(
			repositories.CurrentAccessStaff,
			accounts,
			dependencies.StaffAssignments,
			dependencies.Clinics,
		),
		login: authService,
	}
}

func newAuthHTTPHandler(
	services authServices,
	dependencies authCompositionDependencies,
) *auth.HTTPHandler {
	var auditLogger auth.AuthAuditLogger
	if dependencies.Audit != nil {
		auditLogger = authHTTPAuditAdapter{logger: dependencies.Audit}
	}
	return auth.NewHTTPHandler(auth.HTTPDependencies{
		Auth:                 services.login,
		Tokens:               services.tokens,
		TokenBlacklist:       services.tokenBlacklist,
		PasswordReset:        services.passwordReset,
		Accounts:             services.accounts,
		Staff:                dependencies.Staff,
		StaffAssignments:     dependencies.StaffAssignments,
		Clinics:              dependencies.Clinics,
		PermissionGroups:     services.permissionGroups,
		EffectivePermissions: services.permissionGroups,
		Audit:                auditLogger,
	}, dependencies.CookieConfig)
}

type authHTTPAuditAdapter struct {
	logger audit.Service
}

func (a authHTTPAuditAdapter) LogAuthLogin(
	ctx context.Context,
	clinicID, staffID *uint64,
	action, ipAddress, userAgent string,
) error {
	if a.logger == nil {
		return fmt.Errorf("auth HTTP audit logger is required")
	}
	return a.logger.LogAuthLogin(
		ctx,
		clinicID,
		staffID,
		action,
		ipAddress,
		userAgent,
	)
}

func (a authHTTPAuditAdapter) LogEntry(
	ctx context.Context,
	entry auth.AuthAuditEntry,
) error {
	if a.logger == nil {
		return fmt.Errorf("auth HTTP audit logger is required")
	}
	return a.logger.LogEntry(ctx, authAuditEntry(entry))
}

type authPermissionAuditTxAdapter struct {
	logger audit.TxLogger
}

func (a authPermissionAuditTxAdapter) LogEntryTx(
	ctx context.Context,
	entry auth.AuthAuditEntry,
) error {
	if a.logger == nil {
		return fmt.Errorf("auth permission audit logger is required")
	}
	return a.logger.LogEntryTx(ctx, authAuditEntry(entry))
}

func authAuditEntry(entry auth.AuthAuditEntry) *audit.Entry {
	return &audit.Entry{
		ClinicID:   entry.ClinicID,
		ActorID:    entry.ActorID,
		ActorType:  entry.ActorType,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
	}
}

func nilSafeDrain(drain func()) func() {
	return func() {
		if drain != nil {
			drain()
		}
	}
}
