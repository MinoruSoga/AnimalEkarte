import { useActionState, useEffect, useRef } from "react";
import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { UpdateMedicalRecordRequest } from "../api/types";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";

interface UseMedicalRecordSaveActionArgs {
  recordId?: string;
  activeTab: string;
  canEdit: boolean;
  isNextVisitDateValid: boolean;
  diagnosis1CategoryId: number | null;
  diagnosis1NameId: number | null;
  diagnosis2CategoryId: number | null;
  diagnosis2NameId: number | null;
  plan: string;
  assessment: string;
  chiefComplaint: string;
  chiefComplaintDefault: string;
  chiefComplaintTypeId: number | null;
  treatmentPolicy: string;
  treatmentPolicyDefault: string;
  nextVisitDate: string;
  existingRecordVersion?: number;
  existingClinicalPlanVersion?: number;
  setManualErrors: (errors: Record<string, string>) => void;
  queryClient: QueryClient;
  updateInquiryMutation: {
    mutateAsync: (variables: {
      chief_complaint?: string;
      chief_complaint_type_id: number | null;
      notes?: string;
    }) => Promise<unknown>;
  };
  updateTreatmentPlanMutation: {
    mutateAsync: (variables: {
      treatment_policy: string;
      diagnosis_details: string;
      diagnosis_type_id?: number;
      diagnosis_name_id?: number;
      diagnosis_2_type_id?: number | null;
      diagnosis_2_name_id?: number | null;
      version?: number;
    }) => Promise<unknown>;
  };
  updateMutation: {
    mutateAsync: (variables: { id: string; req: UpdateMedicalRecordRequest }) => Promise<unknown>;
  };
}

export function useMedicalRecordSaveAction({
  recordId,
  activeTab,
  canEdit,
  isNextVisitDateValid,
  diagnosis1CategoryId,
  diagnosis1NameId,
  diagnosis2CategoryId,
  diagnosis2NameId,
  plan,
  assessment,
  chiefComplaint,
  chiefComplaintDefault,
  chiefComplaintTypeId,
  treatmentPolicy,
  treatmentPolicyDefault,
  nextVisitDate,
  existingRecordVersion,
  existingClinicalPlanVersion,
  setManualErrors,
  queryClient,
  updateInquiryMutation,
  updateTreatmentPlanMutation,
  updateMutation,
}: UseMedicalRecordSaveActionArgs) {
  // activeTab を保存時に正確に参照するための ref
  const activeTabRef = useRef(activeTab);
  useEffect(() => {
    activeTabRef.current = activeTab;
  }, [activeTab]);

  const [formState, formAction, isSaving] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      if (!recordId) return { success: false, timestamp: Date.now() };

      try {
        setManualErrors({});
        const currentTab = activeTabRef.current;

        switch (currentTab) {
          case "問診":
            await updateInquiryMutation.mutateAsync({
              chief_complaint: chiefComplaint !== chiefComplaintDefault ? chiefComplaint : undefined,
              chief_complaint_type_id: chiefComplaintTypeId,
              notes: treatmentPolicy !== treatmentPolicyDefault ? treatmentPolicy : undefined,
            });
            break;

          case "診察/治療プラン": {
            if (!canEdit) break;
            if (!isNextVisitDateValid) {
              return { success: false, timestamp: Date.now() };
            }
            if (diagnosis1CategoryId && !diagnosis1NameId) {
              const diagError = { diagnosis1_name_id: "診断名を選択してください" };
              setManualErrors(diagError);
              return { success: false, fieldErrors: diagError, timestamp: Date.now() };
            }
            // BUG-416 ②: diagnosis1 と同じバリデーションを diagnosis2 にも適用する（FE validation parity）
            if (diagnosis2CategoryId && !diagnosis2NameId) {
              const diagError = { diagnosis2_name_id: "診断名を選択してください" };
              setManualErrors(diagError);
              return { success: false, fieldErrors: diagError, timestamp: Date.now() };
            }
            // BUG-102: DEFAULT値でも常に送信する（undefined を送ると BE が 400 を返す）
            const treatmentPlanPayload = {
              treatment_policy: plan,
              diagnosis_details: assessment,
              diagnosis_type_id: diagnosis1CategoryId ?? undefined,
              diagnosis_name_id: diagnosis1NameId ?? undefined,
              diagnosis_2_type_id: diagnosis2CategoryId,
              diagnosis_2_name_id: diagnosis2NameId,
              // BUG-416③: clinical_plan 楽観ロック。undefined を送ると BE は
              // バージョンチェックをスキップする（後方互換）ため常に送信する。
              version: existingClinicalPlanVersion,
            };
            await updateTreatmentPlanMutation.mutateAsync(treatmentPlanPayload);
            // 次回来院推奨日を更新（空欄 = クリア、値あり = 設定）
            await updateMutation.mutateAsync({
              id: recordId as string,
              req: {
                next_visit_recommended_date: nextVisitDate, // "" はBEでNULLクリア
                version: existingRecordVersion,
              } as UpdateMedicalRecordRequest,
            });
            break;
          }

          default:
            break;
        }

        toast.success("保存しました");
        queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

  return { formState, formAction, isSaving };
}
