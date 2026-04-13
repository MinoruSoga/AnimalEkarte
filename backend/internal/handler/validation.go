package handler

import (
	"context"
	"fmt"
	"slices"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
	assignments, err := h.svc.StaffClinicAssignment.FindByStaffID(ctx, doctorID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify staff assignment")
	}
	for _, a := range assignments {
		if a.ClinicID == clinicID {
			return nil
		}
	}
	return apperrors.WrapInvalidInput("指定されたスタッフはこのクリニックに所属していません")
}
