import { useCallback, useLayoutEffect, useRef, useTransition } from "react";
import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { usePermission } from "@/hooks/use-permission";
import type { UpdateMedicalRecordRequest } from "../api/types";
import { isSupportedVisitTypeLabel, toVisitTypeValue } from "./use-medical-record-form-model";

interface UseMedicalRecordQuickPatchActionsArgs {
  recordId?: string;
  existingRecordVersion?: number;
  visitType: string;
  setVisitType: (visitType: string) => void;
  nextVisitDate: string;
  setNextVisitDate: (date: string) => void;
  queryClient: QueryClient;
  updateMutation: {
    mutateAsync: (variables: { id: string; req: UpdateMedicalRecordRequest }) => Promise<unknown>;
  };
  canEdit?: boolean;
  isSelectedPetDeceased: boolean;
}

export function useMedicalRecordQuickPatchActions({
  recordId,
  existingRecordVersion,
  visitType,
  setVisitType,
  nextVisitDate,
  setNextVisitDate,
  queryClient,
  updateMutation,
  canEdit: canEditOverride,
  isSelectedPetDeceased,
}: UseMedicalRecordQuickPatchActionsArgs) {
  const { mutateAsync } = updateMutation;
  const { canEdit: permissionCanEdit } = usePermission("medical-records");
  const canEdit = canEditOverride ?? permissionCanEdit;
  const canEditRef = useRef(canEdit);
  const isSelectedPetDeceasedRef = useRef(isSelectedPetDeceased);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  useLayoutEffect(() => {
    isSelectedPetDeceasedRef.current = isSelectedPetDeceased;
  }, [isSelectedPetDeceased]);
  const isMutationAllowed = useCallback(
    () => canEditRef.current === true && !isSelectedPetDeceasedRef.current,
    [],
  );

  // useTransition: 即時PATCH系ハンドラの pending 管理 (rerender-transitions)
  const [isSavingTransition, startSaveTransition] = useTransition();

  // 担当医変更ハンドラ
  const handleChangeDoctor = (newDoctorId: string, newDoctorName: string) => {
    if (!recordId || !isMutationAllowed()) return;
    startSaveTransition(async () => {
      if (!isMutationAllowed()) return;
      try {
        await mutateAsync({
          id: recordId,
          req: {
            doctor_id: Number(newDoctorId),
            version: existingRecordVersion,
          } as UpdateMedicalRecordRequest,
        });
        toast.success(`担当医を ${newDoctorName} に変更しました`);
      } catch (error) {
        handleApiError(error, "担当医変更");
      }
    });
  };

  // 来院種別変更ハンドラ（即時PATCH）
  // existingRecordVersion のみ参照するため object 全体を dep に含めない (OCC versioning)
  const handleVisitTypeChange = useCallback((newVisitType: string) => {
    if (!isMutationAllowed()) return;
    if (!isSupportedVisitTypeLabel(newVisitType)) {
      toast.error("来院種別は初診または再診のみ保存できます");
      return;
    }
    const mappedVisitType = toVisitTypeValue(newVisitType);
    if (!mappedVisitType) return;
    const prevVisitType = visitType;
    setVisitType(newVisitType);
    if (!recordId) return; // 新規作成時はローカルstateのみ
    startSaveTransition(async () => {
      if (!isMutationAllowed()) return;
      try {
        await mutateAsync({
          id: recordId,
          req: {
            visit_type: mappedVisitType,
            version: existingRecordVersion,
          } as UpdateMedicalRecordRequest,
        });
        await queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.detail(recordId) });
        toast.success(`来院種別を ${newVisitType} に変更しました`);
      } catch (error) {
        setVisitType(prevVisitType); // H-1: rollback on PATCH failure
        handleApiError(error, "来院種別変更");
      }
    });
  }, [visitType, setVisitType, recordId, existingRecordVersion, mutateAsync, queryClient, startSaveTransition, isMutationAllowed]);

  // 次回予定変更ハンドラ（ヘッダー NextVisitButton 用・即時PATCH）
  // existingRecordVersion のみ参照するため object 全体を dep に含めない (OCC versioning)
  const handleNextVisitDatePatch = useCallback((newDate: string) => {
    if (!isMutationAllowed()) return;
    const prev = nextVisitDate;
    setNextVisitDate(newDate);
    if (!recordId) return; // 新規作成時はローカルstateのみ
    startSaveTransition(async () => {
      if (!isMutationAllowed()) return;
      try {
        await mutateAsync({
          id: recordId,
          req: {
            next_visit_recommended_date: newDate, // "" = クリア
            version: existingRecordVersion,
          } as UpdateMedicalRecordRequest,
        });
        await queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.detail(recordId) });
        toast.success(newDate ? `次回予定を ${newDate} に設定しました` : "次回予定をクリアしました");
      } catch (error) {
        setNextVisitDate(prev); // rollback
        handleApiError(error, "次回予定変更");
      }
    });
  }, [nextVisitDate, setNextVisitDate, recordId, existingRecordVersion, mutateAsync, queryClient, startSaveTransition, isMutationAllowed]);

  // 診察日変更ハンドラ
  // existingRecordVersion のみ参照するため object 全体を dep に含めない (OCC versioning)
  const handleChangeDate = useCallback((newDate: string) => {
    if (!recordId || !isMutationAllowed()) return;
    startSaveTransition(async () => {
      if (!isMutationAllowed()) return;
      try {
        await mutateAsync({
          id: recordId,
          req: {
            date: `${newDate}T00:00:00+09:00`,
            version: existingRecordVersion,
          } as UpdateMedicalRecordRequest,
        });
        await queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.detail(recordId) });
        toast.success(`診察日を ${newDate} に変更しました`);
      } catch (error) {
        handleApiError(error, "診察日変更");
      }
    });
  }, [recordId, existingRecordVersion, mutateAsync, queryClient, startSaveTransition, isMutationAllowed]);

  // カルテ確定（SPEC-GAP）: draft→finalized の一方向遷移。BE は確定済みカルテへの
  // 更新を 409 で拒否し（medical_record_crud.go）、確定の取り消し API は存在しない
  // （訂正は addendum のみ）。既存の quick-patch と同じ OCC versioning パターンに従う。
  // existingRecordVersion のみ参照するため object 全体を dep に含めない (OCC versioning)
  const handleFinalize = useCallback(() => {
    if (!recordId || !isMutationAllowed()) return;
    startSaveTransition(async () => {
      if (!isMutationAllowed()) return;
      try {
        await mutateAsync({
          id: recordId,
          req: {
            status: "finalized",
            version: existingRecordVersion,
          } as UpdateMedicalRecordRequest,
        });
        await queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.detail(recordId) });
        toast.success("カルテを確定しました");
      } catch (error) {
        handleApiError(error, "カルテ確定");
      }
    });
  }, [recordId, existingRecordVersion, mutateAsync, queryClient, startSaveTransition, isMutationAllowed]);

  return {
    isSavingTransition,
    startSaveTransition,
    handleChangeDoctor,
    handleVisitTypeChange,
    handleNextVisitDatePatch,
    handleChangeDate,
    handleFinalize,
  };
}
