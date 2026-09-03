import { useState, useCallback, useLayoutEffect, useRef } from "react";
import { useNavigate } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { updateHospitalization } from "../api/update-hospitalization";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import { queryKeys } from "@/lib/query-keys";
import { useGetMasterItems } from "@/hooks/use-master-items";
import { HospitalizationFilterStatus, HOSPITALIZATION_FILTER_STATUS, HOSPITALIZATION_STATUS } from "../constants";
import type { Hospitalization } from "@/types";

// FE-RC-037: QueryCache は unknown を返すため、要素形状を最小限検証してから絞り込む（無検証 as を避ける）。
function isHospitalizationLike(value: unknown): value is Hospitalization {
  return typeof value === "object" && value !== null && typeof (value as { id?: unknown }).id === "string";
}

function extractHospitalizations(data: unknown): Hospitalization[] {
  if (data == null) return [];
  if (Array.isArray(data)) return data.filter(isHospitalizationLike);
  if (typeof data === "object" && Array.isArray((data as { data?: unknown }).data)) {
    return (data as { data: unknown[] }).data.filter(isHospitalizationLike);
  }
  return [];
}

export const useHospitalizationList = (canEdit = false) => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<HospitalizationFilterStatus>(HOSPITALIZATION_FILTER_STATUS.ACTIVE);
  const [viewMode, setViewMode] = useState<"list" | "board">("board");
  const canEditRef = useRef(canEdit);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);

  const { data: cages } = useGetMasterItems("cage");

  // React Query キャッシュから現在の入院データを取得してケージ移動を処理する。
  // optimistic update は行わず、updateHospitalization 後の invalidateQueries で UI を更新する。
  const movePet = useCallback(async (hospitalizationId: string, targetCageId: string) => {
    // list query は HospitalizationsResult { data, total, page, limit }（BUG-009）。
    // 旧形 Hospitalization[] キャッシュが残っていても壊さないよう両対応する。
    const allEntries = queryClient.getQueriesData({ queryKey: queryKeys.hospitalizations.all() });
    const hospitalizations = allEntries.flatMap(([, data]) => extractHospitalizations(data));

    const sourceHosp = hospitalizations.find((h) => h.id === hospitalizationId);
    if (!sourceHosp) return;
    if (canEditRef.current !== true || sourceHosp.petIsDeceased) return;

    // 移動先にアクティブな入院がある場合はスワップ
    const targetHosp = hospitalizations.find(
      (h) =>
        h.cageId === targetCageId &&
        h.status === HOSPITALIZATION_STATUS.ACTIVE &&
        h.id !== hospitalizationId,
    );
    if (targetHosp?.petIsDeceased) return;

    try {
      if (targetHosp) {
        // 元のケージがない場合は cage_id フィールドを送らない（空文字列は uint64 として不正）
        const swapPayload = sourceHosp.cageId
          ? { cage_id: sourceHosp.cageId }
          : {};
        const results = await Promise.allSettled([
          updateHospitalization(sourceHosp.id, { cage_id: targetCageId }),
          updateHospitalization(targetHosp.id, swapPayload),
        ]);
        // どちらかが失敗した場合はエラー表示（ロールバックは BE で処理）
        const failed = results.filter((r) => r.status === "rejected");
        if (failed.length > 0) {
          handleApiError(failed[0].reason, "スワップ");
        }
      } else {
        await updateHospitalization(sourceHosp.id, { cage_id: targetCageId });
      }
      queryClient.invalidateQueries({ queryKey: queryKeys.hospitalizations.all() });
    } catch (error) {
      handleApiError(error, "移動");
    }
  }, [queryClient]);

  const handleNavigateToForm = useCallback((id?: string) => {
    navigate(id ? paths.hospitalization.detail.getHref(id) : paths.hospitalization.selectPet.getHref());
  }, [navigate]);

  return {
    searchTerm,
    setSearchTerm,
    statusFilter,
    setStatusFilter,
    viewMode,
    setViewMode,
    cages,
    movePet,
    handleNavigateToForm,
  };
};
