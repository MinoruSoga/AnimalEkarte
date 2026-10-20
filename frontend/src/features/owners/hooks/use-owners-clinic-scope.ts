import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";
import { useAuth } from "@/hooks/use-auth";
import { toggleClinicSelection } from "../lib/owners-clinic-scope-model";
import { writeOwnersClinicSelection } from "../lib/owners-clinic-selection-storage";
import type { OwnersClinicScopeView } from "../loaders";
import type { ClinicMembership } from "@/types/auth";

const EMPTY_CLINICS: ClinicMembership[] = [];
const EMPTY_IDS: string[] = [];

/**
 * EMR-463: 飼主・ペット一覧専有の医院スコープ adapter。
 *
 * 共有 `useClinicScope`（所属全件・未指定時は現在医院）の意味を変えないため別に持つ。
 * 表示医院は loader が GET /me `owners_view_clinic_ids` から確定した現在許可医院だけで、
 * 選択は loader が確定した有効選択（URL → 保存 → 全許可）。
 * 明示トグル時は URL `?clinics=`（適用済み選択の SoT）と保存先の両方へ反映し、page を戻す。
 */
export function useOwnersClinicScope(clinicScope: OwnersClinicScopeView | null) {
  const [, setSearchParams] = useSearchParams();
  const { user } = useAuth();
  const authUserId = user?.id ?? null;

  const clinics = clinicScope?.clinics ?? EMPTY_CLINICS;
  const selectedClinicIds = clinicScope?.selectedClinicIds ?? EMPTY_IDS;
  const scopeUserId = clinicScope?.userId ?? null;

  const permittedClinicIds = useMemo(() => clinics.map((c) => c.clinicId), [clinics]);
  const clinicNameById = useMemo(
    () => new Map(clinics.map((c) => [c.clinicId, c.clinicName])),
    [clinics],
  );

  const handleToggleClinic = useCallback(
    (clinicId: string) => {
      if (scopeUserId === null) return;
      const next = toggleClinicSelection(selectedClinicIds, clinicId, permittedClinicIds);
      if (next === null) return;
      // scope を取得した利用者と現在のログイン利用者が一致するときだけ保存する
      // （切替直後の旧画面から別ユーザーの key へ書かない）。
      if (authUserId === scopeUserId) {
        writeOwnersClinicSelection(scopeUserId, next);
      }
      setSearchParams((prev) => {
        const params = new URLSearchParams(prev);
        params.set("clinics", next.join(","));
        params.delete("page");
        return params;
      });
    },
    [scopeUserId, authUserId, selectedClinicIds, permittedClinicIds, setSearchParams],
  );

  return {
    clinics,
    selectedClinicIds,
    isMultiClinic: selectedClinicIds.length > 1,
    clinicNameById,
    handleToggleClinic,
  };
}
