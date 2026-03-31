import { useState, useCallback } from "react";
import { useNavigate } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { updateHospitalization } from "../api/update-hospitalization";
import { handleApiError } from "@/lib/handle-api-error";
import { useMasterItems } from "@/hooks/use-master-items";
import { HospitalizationFilterStatus, HOSPITALIZATION_FILTER_STATUS, HOSPITALIZATION_STATUS } from "../constants";
import type { Hospitalization } from "@/types";

export const useHospitalizationList = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<HospitalizationFilterStatus>(HOSPITALIZATION_FILTER_STATUS.ACTIVE);
  const [viewMode, setViewMode] = useState<"list" | "board">("board");

  const { data: cages } = useMasterItems("cage");

  // React Query キャッシュから現在の入院データを取得してケージ移動を処理する。
  // optimistic update は行わず、updateHospitalization 後の invalidateQueries で UI を更新する。
  const movePet = useCallback(async (hospitalizationId: string, targetCageId: string) => {
    // 全キャッシュエントリから入院リストを取得（フィルタに関わらず）
    const allEntries = queryClient.getQueriesData<Hospitalization[]>({ queryKey: ["hospitalizations"] });
    const hospitalizations = allEntries.flatMap(([, data]) => data ?? []);

    const sourceHosp = hospitalizations.find((h) => h.id === hospitalizationId);
    if (!sourceHosp) return;

    // 移動先にアクティブな入院がある場合はスワップ
    const targetHosp = hospitalizations.find(
      (h) =>
        h.cageId === targetCageId &&
        h.status === HOSPITALIZATION_STATUS.ACTIVE &&
        h.id !== hospitalizationId,
    );

    try {
      if (targetHosp) {
        await updateHospitalization(sourceHosp.id, { cage_id: targetCageId });
        // 元のケージがない場合は cage_id フィールドを送らない（空文字列は uint64 として不正）
        const swapPayload = sourceHosp.cageId
          ? { cage_id: sourceHosp.cageId }
          : {};
        await updateHospitalization(targetHosp.id, swapPayload);
      } else {
        await updateHospitalization(sourceHosp.id, { cage_id: targetCageId });
      }
      queryClient.invalidateQueries({ queryKey: ["hospitalizations"] });
    } catch (error) {
      handleApiError(error, "移動");
    }
  }, [queryClient]);

  const handleNavigateToForm = (id?: string) => {
    navigate(id ? `/hospitalization/${id}` : "/hospitalization/select-pet");
  };

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
