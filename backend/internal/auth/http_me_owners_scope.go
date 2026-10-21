package auth

import (
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetMeResponse は GET /me 専用の応答 DTO（EMR-463）。
// 共通 MeResponse を同じ JSON 階層に埋め込み、現在許可された owners:view 医院
// 集合だけを追加する。login/refresh の応答、MeResponse 本体、既存 builder の
// signature は変更しない。DTO を http_response.go に置かない
// （tygo include_files=http_response.go の再生成対象を増やさないため）。
type GetMeResponse struct {
	MeResponse
	// OwnersViewClinicIDs はこの要求で検証された active 医院のうち
	// owners:view が許可された医院 ID（正の正規形10進文字列・重複なし）。
	// 許可ゼロは非 nil の空配列。順序に認可上の意味はない。
	OwnersViewClinicIDs []string `json:"owners_view_clinic_ids"`
}

// ownersViewClinicIDs は GetMe の許可医院集合を算出する。
// ExtractClinicIDs(c) が返す現在検証済みの所属医院集合と、組み立て済み
// profile の医院集合（active のみ、profile 側で選別済み）を交差し、各医院へ
// checkPermissionInClinic(owners, view) を順次評価する。system_admin も
// この交差内に限定される。false は除外、評価 error は要求全体を 5xx で
// 失敗させる（部分集合や空成功へ縮退しない）。
// ok=false のとき応答は既に書き込み済み（ExtractClinicIDs の既存 401/400、
// または評価 error の 5xx）なので、呼出側は即 return すること。
func (h *HTTPHandler) ownersViewClinicIDs(
	c *gin.Context,
	profileClinics []MeClinicMembership,
) ([]string, bool) {
	verifiedIDs, ok := httpapi.ExtractClinicIDs(c)
	if !ok {
		return nil, false
	}
	verified := make(map[uint64]struct{}, len(verifiedIDs))
	for _, id := range verifiedIDs {
		verified[id] = struct{}{}
	}

	// 交差候補。profile の名称や main_clinic_id から所属を追加しない。
	candidates := make([]uint64, 0, len(profileClinics))
	seen := make(map[uint64]struct{}, len(profileClinics))
	for _, membership := range profileClinics {
		clinicID, err := strconv.ParseUint(membership.ClinicID, 10, 64)
		if err != nil || clinicID == 0 {
			continue
		}
		if _, ok := verified[clinicID]; !ok {
			continue
		}
		if _, dup := seen[clinicID]; dup {
			continue
		}
		seen[clinicID] = struct{}{}
		candidates = append(candidates, clinicID)
	}
	slices.Sort(candidates)

	permitted := make([]string, 0, len(candidates))
	for _, clinicID := range candidates {
		allowed, err := h.checkPermissionInClinic(
			c,
			clinicID,
			string(model.ResourceOwners),
			"view",
		)
		if err != nil {
			// 評価不能は分類に関わらず 5xx で要求全体を失敗させる（契約: 部分集合・空成功へ縮退しない）。
			_ = c.Error(err)
			httpapi.RespondError(c, apperrors.WrapInternalServerError("failed to evaluate owners view scope"))
			return nil, false
		}
		if allowed {
			permitted = append(permitted, strconv.FormatUint(clinicID, 10))
		}
	}
	return permitted, true
}
