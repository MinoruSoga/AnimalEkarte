import { useCallback, useTransition } from "react";
import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { UpdateMedicalRecordRequest } from "../api/types";
import { toVisitTypeValue } from "./use-medical-record-form-model";

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
}: UseMedicalRecordQuickPatchActionsArgs) {
  // useTransition: 即時PATCH系ハンドラの pending 管理 (rerender-transitions)
  const [isSavingTransition, startSaveTransition] = useTransition();

  // 担当医変更ハンドラ
  const handleChangeDoctor = (newDoctorId: string, newDoctorName: string) => {
    if (!recordId) return;
    startSaveTransition(async () => {
      try {
        await updateMutation.mutateAsync({
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
  // eslint-disable-next-line react-hooks/preserve-manual-memoization
  const handleVisitTypeChange = useCallback((newVisitType: string) => {
    const prevVisitType = visitType;
    setVisitType(newVisitType);
    if (!recordId) return; // 新規作成時はローカルstateのみ
    startSaveTransition(async () => {
      try {
        await updateMutation.mutateAsync({
          id: recordId,
          req: {
            visit_type: toVisitTypeValue(newVisitType),
            version: existingRecordVersion,
          } as UpdateMedicalRecordRequest,
        });
        toast.success(`来院種別を ${newVisitType} に変更しました`);
      } catch (error) {
        setVisitType(prevVisitType); // H-1: rollback on PATCH failure
        handleApiError(error, "来院種別変更");
      }
    });
  }, [visitType, setVisitType, recordId, existingRecordVersion, updateMutation, startSaveTransition]);

  // 次回予定変更ハンドラ（ヘッダー NextVisitButton 用・即時PATCH）
  // existingRecordVersion のみ参照するため object 全体を dep に含めない (OCC versioning)
  // eslint-disable-next-line react-hooks/preserve-manual-memoization
  const handleNextVisitDatePatch = useCallback((newDate: string) => {
    const prev = nextVisitDate;
    setNextVisitDate(newDate);
    if (!recordId) return;
    startSaveTransition(async () => {
      try {
        await updateMutation.mutateAsync({
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
  }, [nextVisitDate, setNextVisitDate, recordId, existingRecordVersion, updateMutation, queryClient, startSaveTransition]);

  // 診察日変更ハンドラ
  // existingRecordVersion のみ参照するため object 全体を dep に含めない (OCC versioning)
  // eslint-disable-next-line react-hooks/preserve-manual-memoization
  const handleChangeDate = useCallback((newDate: string) => {
    if (!recordId) return;
    startSaveTransition(async () => {
      try {
        await updateMutation.mutateAsync({
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
  }, [recordId, existingRecordVersion, updateMutation, queryClient, startSaveTransition]);

  return {
    isSavingTransition,
    startSaveTransition,
    handleChangeDoctor,
    handleVisitTypeChange,
    handleNextVisitDatePatch,
    handleChangeDate,
  };
}
