package service

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AuthEffectivePermissions はユーザーの実効権限マップ（handler 応答変換用）。
type AuthEffectivePermissions map[string]AuthResourcePermission

// AuthResourcePermission はリソース単位の権限。
type AuthResourcePermission struct {
	View   bool
	Create bool
	Edit   bool
	Delete bool
}

// AuthService はログイン認証・クリニック解決・実効権限計算を定義する。
type AuthService interface {
	AuthenticateUser(ctx context.Context, email, password string) (*model.Account, *model.Staff, error)
	ResolveClinicInfo(assignments []model.StaffClinicAssignment) (mainClinicID string, clinicIDs []uint64)
	ResolveSystemAdminMainClinicID(mainClinicID string, isSystemAdmin bool, allClinics []model.Clinic) string
	CalculateEffectivePermissions(ctx context.Context, isSystemAdmin bool, staffID, clinicID uint64) AuthEffectivePermissions
}

type authService struct {
	account             AccountService
	staff               StaffService
	effectivePermission EffectivePermissionService
}

// wrongPasswordError はパスワード不一致時の sentinel（監査は handler 層の責務）。
type wrongPasswordError struct {
	accountID uint64
	err       error
}

func (e *wrongPasswordError) Error() string {
	return e.err.Error()
}

func (e *wrongPasswordError) Unwrap() error {
	return e.err
}

// IsAuthenticateWrongPassword はパスワード不一致エラーかどうかを判定し、該当アカウント ID を返す。
func IsAuthenticateWrongPassword(err error) (uint64, bool) {
	var wpe *wrongPasswordError
	if errors.As(err, &wpe) {
		return wpe.accountID, true
	}
	return 0, false
}

// NewAuthService は AuthService の実装を返す。
func NewAuthService(account AccountService, staff StaffService, effectivePermission EffectivePermissionService) AuthService {
	return &authService{
		account:             account,
		staff:               staff,
		effectivePermission: effectivePermission,
	}
}

// AuthenticateUser はメール/パスワードを検証してアカウントとスタッフを返す。
// 認証失敗・アカウント無効・スタッフ無効の場合は apperrors.ErrUnauthorized ラップエラーを返す。
func (s *authService) AuthenticateUser(ctx context.Context, email, password string) (*model.Account, *model.Staff, error) {
	account, err := s.account.FindByEmail(ctx, email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			slog.InfoContext(ctx, "skip audit log for login failure: account not found")
			return nil, nil, apperrors.WrapUnauthorized("メールアドレスまたはパスワードが正しくありません")
		}
		return nil, nil, err
	}
	if !account.IsActive {
		return nil, nil, apperrors.WrapUnauthorized("アカウントが無効です")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		// 監査ログ: ログイン失敗（パスワード不正）は handler 層で記録する
		return nil, nil, &wrongPasswordError{
			accountID: account.ID,
			err:       apperrors.WrapUnauthorized("メールアドレスまたはパスワードが正しくありません"),
		}
	}
	staff, err := s.staff.FindByAccountID(ctx, account.ID)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "スタッフ情報の取得に失敗しました")
	}
	// BUG-134: スタッフの有効性チェック（退職者・一時停止アカウント防止）
	if !staff.IsActive {
		return nil, nil, apperrors.WrapUnauthorized("このアカウントは無効です")
	}
	return account, staff, nil
}

// ResolveClinicInfo はスタッフのクリニック割り当てからメインクリニックIDと所属ID一覧を返す。
// メインクリニックが未設定の場合は最初の割り当てを使用する。
func (s *authService) ResolveClinicInfo(assignments []model.StaffClinicAssignment) (mainClinicID string, clinicIDs []uint64) {
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

// ResolveSystemAdminMainClinicID は system_admin で assignments なしの場合に
// allClinics[0] を main にフォールバックする。それ以外は元の mainClinicID を返す。
func (s *authService) ResolveSystemAdminMainClinicID(mainClinicID string, isSystemAdmin bool, allClinics []model.Clinic) string {
	if mainClinicID != "" {
		return mainClinicID
	}
	if !isSystemAdmin || len(allClinics) == 0 {
		return mainClinicID
	}
	return strconv.FormatUint(allClinics[0].ID, 10)
}

func buildAuthAllPermissions() AuthEffectivePermissions {
	m := make(AuthEffectivePermissions, len(model.AllResources))
	for _, res := range model.AllResources {
		m[string(res)] = AuthResourcePermission{View: true, Create: true, Edit: true, Delete: true}
	}
	return m
}

// CalculateEffectivePermissions はユーザー種別に応じた実効権限を計算する。
// isSystemAdmin=true: 全リソース全権限バイパス
// isSystemAdmin=false: DB の staff_permission_groups → permission_group_rules から UNION 計算
func (s *authService) CalculateEffectivePermissions(ctx context.Context, isSystemAdmin bool, staffID, clinicID uint64) AuthEffectivePermissions {
	// system_admin は全権限バイパス
	if isSystemAdmin {
		return buildAuthAllPermissions()
	}

	// staff: service 経由で実効権限を取得（handler → repository 直接呼び出し禁止、clinic_id スコープ付き）
	rules, err := s.effectivePermission.GetEffectivePermissions(ctx, staffID, clinicID)
	if err != nil {
		// エラー時は空の権限（最小権限の原則）だが、ログ記録は必須（オペレーター障害認知のため）
		slog.ErrorContext(ctx, "failed to get effective permissions", "staff_id", staffID, "clinic_id", clinicID, "error", err)
		return make(AuthEffectivePermissions)
	}

	permMap := make(AuthEffectivePermissions, len(rules))
	for i := range rules {
		rule := &rules[i]
		permMap[rule.Resource] = AuthResourcePermission{
			View:   rule.CanView,
			Create: rule.CanCreate,
			Edit:   rule.CanEdit,
			Delete: rule.CanDelete,
		}
	}
	return permMap
}
