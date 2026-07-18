package handler

import (
	"context"
	"fmt"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// validateEnum はstring値vが許可されたenum値のいずれかであることを検証する。
// 有効な場合は型付きenum値とnilエラーを返す。無効な場合はゼロ値とエラーを返す。
func validateEnum[T ~string](v string, allowed ...T) (T, error) {
	if slices.Contains(allowed, T(v)) {
		return T(v), nil
	}
	var zero T
	return zero, fmt.Errorf("invalid value %q", v)
}

// checkDoctorClinicAssignment は医師が指定クリニックに所属しているかを確認する共通ヘルパー。
// doctorID が 0 の場合はスキップ（医師未指定）。
// 電カル・LINE 両方の予約登録で使用する。
func (h *Handler) checkDoctorClinicAssignment(ctx context.Context, clinicID, doctorID uint64) error {
	if doctorID == 0 {
		return nil
	}
	assignments, err := h.svc.StaffClinicAssignment.FindAllByStaffID(ctx, doctorID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify staff assignment")
	}
	for i := range assignments {
		if assignments[i].ClinicID == clinicID {
			return nil
		}
	}
	return apperrors.WrapInvalidInput("指定されたスタッフはこのクリニックに所属していません")
}

// verifyStaffClinicMembership はスタッフが指定クリニックに所属しているかを確認する。
// 所属していない場合は 404 レスポンスを書いて false を返す。
// 呼び出し元は false 時に即 return すること。
func (h *Handler) verifyStaffClinicMembership(c *gin.Context, clinicID, staffID uint64) bool {
	if err := h.svc.Staff.VerifyClinicMembership(c.Request.Context(), staffID, clinicID); err != nil {
		RespondError(c, err)
		return false
	}
	return true
}

// resolveStaffWithClinic は staff 関連エンドポイント共通の前処理を集約する:
// clinic_id 抽出 → :id パース → クリニック所属検証 を順に行う。いずれかが失敗した
// 時点で各ヘルパー（extractClinicID=401 / parseIDParam=400 / verifyStaffClinicMembership=404）
// が適切なエラー応答を書き、ok=false を返す。呼び出し元は ok==false で即 return すること。
// 挙動は従来のインライン前処理と同一（短絡順序・各エラー応答を保存）。
func (h *Handler) resolveStaffWithClinic(c *gin.Context) (clinicID, staffID uint64, ok bool) {
	clinicID, ok = extractClinicID(c)
	if !ok {
		return 0, 0, false
	}
	staffID, ok = parseIDParam(c, "id")
	if !ok {
		return 0, 0, false
	}
	if !h.verifyStaffClinicMembership(c, clinicID, staffID) {
		return 0, 0, false
	}
	return clinicID, staffID, true
}
