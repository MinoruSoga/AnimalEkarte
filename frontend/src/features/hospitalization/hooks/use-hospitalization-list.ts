import { isAxiosError } from "axios";
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

  const { data: cages, isLoading: cagesLoading, isError: cagesError } = useGetMasterItems("cage");

  // H1: source/target は HospitalizationBoard が「表示中の行」のスナップショットを渡す。
  // any-cache scan で拾った stale/別クリニックの行ではなく、ユーザーが見ていた行を
  // expected_cage_id / swap 判定の根拠にする。optimistic update は行わず、
  // mutation 後の invalidateQueries で UI を更新する。
  const movePet = useCallback(
    async (source: Hospitalization, targetCageId: string, target?: Hospitalization) => {
      // EMR-227: 死亡ペット/退院済み入院への書込みは callback 側でも拒否する（render 側は canDrag で維持）。
      // retained callback 対策として権限・死亡・退院は mutation 直前に必ず再検査する。
      if (
        canEditRef.current !== true ||
        source.petIsDeceased ||
        source.status === HOSPITALIZATION_STATUS.DISCHARGED ||
        target?.petIsDeceased ||
        target?.status === HOSPITALIZATION_STATUS.DISCHARGED ||
        target?.id === source.id
      )
        return;

      try {
        if (target) {
          // EMR-306 (Codex review P1): 入れ替えは単一トランザクションの cage-swap API で
          // 行う — 2本の独立 PATCH は片方だけ commit され得る（preflight と PATCH の間に
          // 別セッションが連結する TOCTOU 含む）。下の preflight は UX 用の早期拒否であり、
          // 最終的な拒否・原子性はバックエンドが保証する。
          if (
            isInMergedGroup(cages ?? [], source.cageId) ||
            isInMergedGroup(cages ?? [], targetCageId)
          ) {
            toast.error("連結中のケージを含む入れ替えはできません");
            return;
          }
          if (!source.cageId || !target.cageId) return;
          await swapHospitalizationCages(source.id, target.id, [source.cageId, target.cageId]);
        } else {
          await updateHospitalization(source.id, {
            cage_id: targetCageId,
            ...(source.cageId
              ? { expected_cage_id: source.cageId }
              : { expected_cage_is_null: true }),
          });
        }
        queryClient.invalidateQueries({ queryKey: queryKeys.hospitalizations.all() });
      } catch (error) {
        handleApiError(error, "移動");
        // 409 は表示スナップショットとサーバー状態の不整合。一覧を refetch して
        // 次の gesture が新しい表示行を使えるようにする（auto replay はしない）。
        // queryKeys.hospitalizations.all() = ["hospitalizations"] — キーに clinic は
        // 含まれないが、cached entry は常に現在 clinic のデータ（clinic 切替で refetch）。
        if (isAxiosError(error) && error.response?.status === 409) {
          queryClient.invalidateQueries({ queryKey: queryKeys.hospitalizations.all() });
        }
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
    cagesLoading,
    cagesError,
    movePet,
    handleNavigateToForm,
  };
};
