import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";
import { useAuth } from "./use-auth";

interface UseClinicScopeOptions {
  /** 拠点トグル時に URL から削除する付随パラメータ（例: page） */
  resetParamsOnToggle?: readonly string[];
}

const EMPTY_RESET_PARAMS: readonly string[] = [];

/**
 * #86 拠点横断表示で共通して使う URL-driven の拠点スコープ管理。
 *
 * - selectedClinicIds : URL ?clinics= から解決した選択中の医院ID配列
 * - assignedClinics   : ユーザーが所属する全医院のメンバーシップ情報
 * - isMultiClinic     : 2拠点以上選択中かどうか
 * - clinicNameById    : clinicId → 医院名 のMapルックアップ
 * - handleToggleClinic: 拠点トグル（URL を更新する）
 */
export function useClinicScope({
  resetParamsOnToggle = EMPTY_RESET_PARAMS,
}: UseClinicScopeOptions = {}) {
  const [searchParams, setSearchParams] = useSearchParams();
  const { user, currentClinicId } = useAuth();

  const assignedClinics = useMemo(() => user?.clinics ?? [], [user?.clinics]);

  const clinicsParam = searchParams.get("clinics");
  const selectedClinicIds = useMemo(
    () =>
      clinicsParam
        ? clinicsParam.split(",").filter(Boolean)
        : currentClinicId
          ? [currentClinicId]
          : [],
    [clinicsParam, currentClinicId],
  );

  const isMultiClinic = selectedClinicIds.length > 1;

  const clinicNameById = useMemo(
    () => new Map(assignedClinics.map((c) => [c.clinicId, c.clinicName])),
    [assignedClinics],
  );

  const handleToggleClinic = useCallback(
    (clinicId: string) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        const current =
          next.get("clinics")?.split(",").filter(Boolean) ??
          (currentClinicId ? [currentClinicId] : []);
        const updated = current.includes(clinicId)
          ? current.filter((id) => id !== clinicId)
          : [...current, clinicId];
        if (updated.length === 0) return prev;
        if (updated.length === 1 && updated[0] === currentClinicId) {
          next.delete("clinics");
        } else {
          next.set("clinics", updated.join(","));
        }
        for (const param of resetParamsOnToggle) {
          next.delete(param);
        }
        return next;
      });
    },
    [setSearchParams, currentClinicId, resetParamsOnToggle],
  );

  return {
    assignedClinics,
    selectedClinicIds,
    isMultiClinic,
    clinicNameById,
    handleToggleClinic,
    currentClinicId,
  };
}
