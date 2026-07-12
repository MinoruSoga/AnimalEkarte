package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// toMeResponse はスタッフデータと補助情報からMeResponseを構築する。
// effectivePerms は事前に計算された実効権限マップ。nil の場合はデフォルト（全権限なし）。
func toMeResponse(staff *model.Staff, account *model.Account, mainClinicID string, clinicNameMap map[string]string, allClinics []model.Clinic, effectivePerms EffectivePermissions) *MeResponse {
	meClinicList := make([]MeClinicMembership, 0)
	isSystemAdmin := account != nil && account.IsSystemAdmin
	if isSystemAdmin {
		// system_admin は全クリニックを切替候補として露出
		for i := range allClinics {
			cl := &allClinics[i]
			clIDStr := strconv.FormatUint(cl.ID, 10)
			meClinicList = append(meClinicList, MeClinicMembership{
				ClinicID:   clIDStr,
				ClinicName: cl.Name,
				IsMain:     clIDStr == mainClinicID,
			})
		}
	} else if staff != nil && len(staff.ClinicAssignments) > 0 {
		// 通常スタッフは assignments ベース（既存ロジック）
		for i := range staff.ClinicAssignments {
			asg := &staff.ClinicAssignments[i]
			clIDStr := strconv.FormatUint(asg.ClinicID, 10)
			meClinicList = append(meClinicList, MeClinicMembership{
				ClinicID:   clIDStr,
				ClinicName: clinicNameMap[clIDStr],
				IsMain:     clIDStr == mainClinicID,
			})
		}
	}

	permMap := effectivePerms
	if permMap == nil {
		permMap = make(EffectivePermissions)
	}

	var meClinic *MeClinicInfo
	for i := range allClinics {
		if strconv.FormatUint(allClinics[i].ID, 10) != mainClinicID {
			continue
		}
		cl := &allClinics[i]
		var logoURL *string
		if cl.LogoURL != "" {
			logoURL = &cl.LogoURL
		}
		meClinic = &MeClinicInfo{
			ID:                         strconv.FormatUint(cl.ID, 10),
			Name:                       cl.Name,
			PostalCode:                 cl.PostalCode,
			Address:                    cl.Address,
			PhoneNumber:                cl.PhoneNumber,
			FaxNumber:                  cl.FaxNumber,
			RegistrationNumber:         cl.RegistrationNumber,
			DirectorName:               cl.DirectorName,
			Email:                      cl.Email,
			Website:                    cl.Website,
			LogoURL:                    logoURL,
			StandardTaxRate:            cl.StandardTaxRate,
			ReducedTaxRate:             cl.ReducedTaxRate,
			AccountingDocumentShowLogo: cl.AccountingDocumentShowLogo,
			AccountingDocumentShowRegistrationWarning: cl.AccountingDocumentShowRegistrationWarning,
			AccountingDocumentShowItemCategory:        cl.AccountingDocumentShowItemCategory,
			AccountingDocumentFooterNote:              cl.AccountingDocumentFooterNote,
			AccountingDocumentShowClinicHeader:        cl.AccountingDocumentShowClinicHeader,
			AccountingDocumentShowOwnerPetInfo:        cl.AccountingDocumentShowOwnerPetInfo,
			AccountingDocumentShowItemsTable:          cl.AccountingDocumentShowItemsTable,
			AccountingDocumentShowPaymentSummary:      cl.AccountingDocumentShowPaymentSummary,
			AccountingDocumentSectionOrder:            sectionOrderToSlice(cl.AccountingDocumentSectionOrder),
		}
		break
	}

	var occupation *string
	if staff != nil && staff.Occupation != nil {
		occ := staff.Occupation.Name
		occupation = &occ
	}

	staffID := uint64(0)
	if staff != nil {
		staffID = staff.ID
	}

	email := ""
	if account != nil {
		email = account.Email
	}

	return &MeResponse{
		ID:            strconv.FormatUint(staffID, 10),
		Email:         email,
		DisplayName:   staff.Name,
		IsSystemAdmin: isSystemAdmin,
		Occupation:    occupation,
		MainClinicID:  mainClinicID,
		Clinic:        meClinic,
		Clinics:       meClinicList,
		Permissions:   permMap,
	}
}

// buildAllPermissions は全リソースに対して全CRUD true のマップを返す。
// is_system_admin=true 用。
func buildAllPermissions() EffectivePermissions {
	m := make(EffectivePermissions, len(model.AllResources))
	for _, res := range model.AllResources {
		m[string(res)] = ResourcePermission{View: true, Create: true, Edit: true, Delete: true}
	}
	return m
}

// calculateEffectivePermissions はユーザー種別に応じた実効権限を計算する。
// isSystemAdmin=true: 全リソース全権限バイパス
// isSystemAdmin=false: DB の staff_permission_groups → permission_group_rules から UNION 計算
func (h *Handler) calculateEffectivePermissions(ginCtx *gin.Context, isSystemAdmin bool, staffID uint64) EffectivePermissions {
	// system_admin は全権限バイパス
	if isSystemAdmin {
		return buildAllPermissions()
	}

	clinicID, ok := extractClinicID(ginCtx)
	if !ok {
		slog.ErrorContext(ginCtx.Request.Context(), "failed to extract clinic_id for effective permissions", "staff_id", staffID)
		return make(EffectivePermissions)
	}

	// staff: service 経由で実効権限を取得（handler → repository 直接呼び出し禁止、clinic_id スコープ付き）
	rules, err := h.svc.EffectivePermission.GetEffectivePermissions(ginCtx.Request.Context(), staffID, clinicID)
	if err != nil {
		// エラー時は空の権限（最小権限の原則）だが、ログ記録は必須（オペレーター障害認知のため）
		slog.ErrorContext(ginCtx.Request.Context(), "failed to get effective permissions", "staff_id", staffID, "clinic_id", clinicID, "error", err)
		return make(EffectivePermissions)
	}

	permMap := make(EffectivePermissions, len(rules))
	for i := range rules {
		rule := &rules[i]
		permMap[rule.Resource] = ResourcePermission{
			View:   rule.CanView,
			Create: rule.CanCreate,
			Edit:   rule.CanEdit,
			Delete: rule.CanDelete,
		}
	}
	return permMap
}

// ---- パスワードリセット ----
