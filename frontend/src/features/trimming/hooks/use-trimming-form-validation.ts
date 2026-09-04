import { useCallback, useState } from "react";
import type { TrimmingFormData } from "@/types/trimming";

export type TrimmingFieldErrors = Record<string, string>;

interface TrimmingValidationSuccess {
  valid: true;
  reservationTypeId: number;
}

interface TrimmingValidationFailure {
  valid: false;
  errors: TrimmingFieldErrors;
}

type TrimmingValidationResult = TrimmingValidationSuccess | TrimmingValidationFailure;

/**
 * BUG-027: staff/course/reservationTypeId のインラインフィールドバリデーション。
 */
export function useTrimmingFormValidation() {
  const [fieldErrors, setFieldErrors] = useState<TrimmingFieldErrors>({});

  const validate = useCallback(
    (
      formData: TrimmingFormData,
      defaultTrimmingReservationTypeId: number | undefined,
    ): TrimmingValidationResult => {
      const errors: TrimmingFieldErrors = {};
      if (!formData.staffId) {
        errors.staffId = "担当者を選択してください";
      }
      if (!formData.courseId) {
        errors.courseId = "コースを選択してください";
      }
      const reservationTypeId = formData.reservationTypeId
        ? Number(formData.reservationTypeId)
        : defaultTrimmingReservationTypeId;
      if (!reservationTypeId || !Number.isFinite(reservationTypeId)) {
        errors.reservationTypeId = "トリミング予約区分が設定されていません";
      }
      if (Object.keys(errors).length > 0) {
        setFieldErrors(errors);
        return { valid: false, errors };
      }
      setFieldErrors({});
      return { valid: true, reservationTypeId: Number(reservationTypeId) };
    },
    [],
  );

  return { fieldErrors, validate };
}
