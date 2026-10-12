import { useState, useCallback, useLayoutEffect, useRef } from "react";
import { useNavigate } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { updateHospitalization } from "../api/update-hospitalization";
import { swapHospitalizationCages } from "../api/swap-hospitalization-cages";
import { isInMergedGroup } from "../lib/cage-merge-model";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import { queryKeys } from "@/lib/query-keys";
import { useGetMasterItems } from "@/hooks/use-master-items";
import {
  HospitalizationFilterStatus,
  HOSPITALIZATION_FILTER_STATUS,
  HOSPITALIZATION_STATUS,
} from "../constants";
import type { Hospitalization } from "@/types";

// FE-RC-037: QueryCache は unknown を返すため、要素形状を最小限検証してから絞り込む（無検証 as を避ける）。
function isHospitalizationLike(value: unknown): value is Hospitalization {
  return (
    typeof value === "object" &&
    value !== null &&
    typeof (value as { id?: unknown }).id === "string"
  );
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
  const [statusFilter, setStatusFilter] = useState<HospitalizationFilterStatus>(
    HOSPITALIZATION_FILTER_STATUS.ACTIVE,
  );
  const [viewMode, setViewMode] = useState<"list" | "board">("board");
  const canEditRef = useRef(canEdit);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);

  const { data: cages } = useGetMasterItems("cage");

  // React Query キャッシュから現在の入院データを取得してケージ移動を処理する。
  // optimistic update は行わず、updateHospitalization 後の invalidateQueries で UI を更新する。
  const movePet = useCallback(
    async (hospitalizationId: string, targetCageId: string) => {
      // list query は HospitalizationsResult { data, total, page, limit }（BUG-009）。
      // 旧形 Hospitalization[] キャッシュが残っていても壊さないよう両対応する。
      const allEntries = queryClient.getQueriesData({ queryKey: queryKeys.hospitalizations.all() });
      const hospitalizations = allEntries.flatMap(([, data]) => extractHospitalizations(data));

      const sourceHosp = hospitalizations.find((h) => h.id === hospitalizationId);
      if (!sourceHosp) return;
      // EMR-227: 死亡ペット/退院済み入院への書込みは callback 側でも拒否する（render 側は canDrag で維持）。
      if (
        canEditRef.current !== true ||
        sourceHosp.petIsDeceased ||
        sourceHosp.status === HOSPITALIZATION_STATUS.DISCHARGED
      )
        return;

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
          // EMR-306 (Codex review P1): 入れ替えは単一トランザクションの cage-swap API で
          // 行う — 2本の独立 PATCH は片方だけ commit され得る（preflight と PATCH の間に
          // 別セッションが連結する TOCTOU 含む）。下の preflight は UX 用の早期拒否であり、
          // 最終的な拒否・原子性はバックエンドが保証する。
          if (
            isInMergedGroup(cages ?? [], sourceHosp.cageId) ||
            isInMergedGroup(cages ?? [], targetCageId)
          ) {
            toast.error("連結中のケージを含む入れ替えはできません");
            return;
          }
          if (!sourceHosp.cageId || !targetHosp.cageId) return;
          await swapHospitalizationCages(sourceHosp.id, targetHosp.id, [
            sourceHosp.cageId,
            targetHosp.cageId,
          ]);
        } else {
          await updateHospitalization(sourceHosp.id, {
            cage_id: targetCageId,
            ...(sourceHosp.cageId
              ? { expected_cage_id: sourceHosp.cageId }
              : { expected_cage_is_null: true }),
          });
        }
        queryClient.invalidateQueries({ queryKey: queryKeys.hospitalizations.all() });
      } catch (error) {
        handleApiError(error, "移動");
      }
    },
    [queryClient, cages],
  );

  const handleNavigateToForm = useCallback(
    (id?: string) => {
      navigate(
        id ? paths.hospitalization.detail.getHref(id) : paths.hospitalization.selectPet.getHref(),
      );
    },
    [navigate],
  );

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
