import { useActionState, useLayoutEffect, useRef } from "react";
import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { UpdateMedicalRecordRequest } from "../api/types";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";
const DECEASED_SAVE_MESSAGE = "死亡したペットのカルテは保存できません";

function deniedState(error: string): ActionState {
  return { success: false, error, timestamp: Date.now() };
}

interface UseMedicalRecordSaveActionArgs {
  recordId?: string;
  activeTab: string;
  canEdit: boolean;
  isSelectedPetDeceased: boolean;
  isFinalized: boolean;
  isNextVisitDateValid: boolean;
  diagnosis1CategoryId: number | null;
  diagnosis1NameId: number | null;
  diagnosis2CategoryId: number | null;
  diagnosis2NameId: number | null;
  /** 身体検査所見 → clinical_plan.physical_exam */
  physicalExam: string;
  /** 治療方針 → clinical_plan.treatment_policy */
  plan: string;
  /** 診断詳細 → clinical_plan.diagnosis_details */
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
      physical_exam: string;
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
  isSelectedPetDeceased,
  isFinalized,
  isNextVisitDateValid,
  diagnosis1CategoryId,
  diagnosis1NameId,
  diagnosis2CategoryId,
  diagnosis2NameId,
  physicalExam,
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
  const canEditRef = useRef(canEdit);
  const isSelectedPetDeceasedRef = useRef(isSelectedPetDeceased);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  useLayoutEffect(() => {
    isSelectedPetDeceasedRef.current = isSelectedPetDeceased;
  }, [isSelectedPetDeceased]);

  // activeTab を保存時に正確に参照するための ref（commit 直後 callback 用）
  const activeTabRef = useRef(activeTab);
  useLayoutEffect(() => {
    activeTabRef.current = activeTab;
  }, [activeTab]);

  const saveSnapshotRef = useRef({
    recordId,
    isFinalized,
    isNextVisitDateValid,
    diagnosis1CategoryId,
    diagnosis1NameId,
    diagnosis2CategoryId,
    diagnosis2NameId,
    physicalExam,
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
  });
  useLayoutEffect(() => {
    saveSnapshotRef.current = {
      recordId,
      isFinalized,
      isNextVisitDateValid,
      diagnosis1CategoryId,
      diagnosis1NameId,
      diagnosis2CategoryId,
      diagnosis2NameId,
      physicalExam,
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
    };
  });

  const [formState, formAction, isSaving] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      const snapshot = saveSnapshotRef.current;
      // UI の disabled は操作補助にすぎない。programmatic submit や race でも
      // 確定済み・権限なしカルテを更新しないよう action 境界で拒否する。
      if (isSelectedPetDeceasedRef.current) {
        return deniedState(DECEASED_SAVE_MESSAGE);
      }
      if (!snapshot.recordId || canEditRef.current !== true || snapshot.isFinalized) {
        return deniedState(PERMISSION_DENIED_MESSAGE);
      }

      try {
        setManualErrors({});
        const currentTab = activeTabRef.current;

        switch (currentTab) {
          case "問診":
            if (isSelectedPetDeceasedRef.current) {
              return deniedState(DECEASED_SAVE_MESSAGE);
            }
            if (canEditRef.current !== true) {
              return deniedState(PERMISSION_DENIED_MESSAGE);
            }
            await updateInquiryMutation.mutateAsync({
              chief_complaint:
                snapshot.chiefComplaint !== snapshot.chiefComplaintDefault
                  ? snapshot.chiefComplaint
                  : undefined,
              chief_complaint_type_id: snapshot.chiefComplaintTypeId,
              notes:
                snapshot.treatmentPolicy !== snapshot.treatmentPolicyDefault
                  ? snapshot.treatmentPolicy
                  : undefined,
            });
            break;

          case "診察/治療プラン": {
            if (!snapshot.isNextVisitDateValid) {
              return { success: false, timestamp: Date.now() };
            }
            // BUG-010: clinical-plan GET/hydrate 前の空文字 PATCH は既存所見を無音で消す。
            // version 未確定（undefined）は BE が楽観ロックをスキップするため fail-closed で拒否する。
            if (typeof snapshot.existingClinicalPlanVersion !== "number") {
              toast.error("診察プランの読み込みが完了してから保存してください");
              return { success: false, timestamp: Date.now() };
            }
            if (snapshot.diagnosis1CategoryId && !snapshot.diagnosis1NameId) {
              const diagError = { diagnosis1_name_id: "診断名を選択してください" };
              setManualErrors(diagError);
              return { success: false, fieldErrors: diagError, timestamp: Date.now() };
            }
            // BUG-416 ②: diagnosis1 と同じバリデーションを diagnosis2 にも適用する（FE validation parity）
            if (snapshot.diagnosis2CategoryId && !snapshot.diagnosis2NameId) {
              const diagError = { diagnosis2_name_id: "診断名を選択してください" };
              setManualErrors(diagError);
              return { success: false, fieldErrors: diagError, timestamp: Date.now() };
            }
            // BUG-010 / BUG-102: 3欄は常に送信する（undefined 欠落は「未更新」になり、
            // テンプレ既定や別 writer の last-write-wins で入力が消える）。空文字は明示クリア。
            const treatmentPlanPayload = {
              physical_exam: snapshot.physicalExam,
              treatment_policy: snapshot.plan,
              diagnosis_details: snapshot.assessment,
              diagnosis_type_id: snapshot.diagnosis1CategoryId ?? undefined,
              diagnosis_name_id: snapshot.diagnosis1NameId ?? undefined,
              diagnosis_2_type_id: snapshot.diagnosis2CategoryId,
              diagnosis_2_name_id: snapshot.diagnosis2NameId,
              // BUG-416③: clinical_plan 楽観ロック。undefined を送ると BE は
              // バージョンチェックをスキップする（後方互換）ため常に送信する。
              version: snapshot.existingClinicalPlanVersion,
            };
            if (isSelectedPetDeceasedRef.current) {
              return deniedState(DECEASED_SAVE_MESSAGE);
            }
            if (canEditRef.current !== true) {
              return deniedState(PERMISSION_DENIED_MESSAGE);
            }
            await updateTreatmentPlanMutation.mutateAsync(treatmentPlanPayload);
            // 次回来院推奨日を更新（空欄 = クリア、値あり = 設定）
            if (isSelectedPetDeceasedRef.current) {
              return deniedState(DECEASED_SAVE_MESSAGE);
            }
            if (canEditRef.current !== true) {
              return deniedState(PERMISSION_DENIED_MESSAGE);
            }
            await updateMutation.mutateAsync({
              id: snapshot.recordId as string,
              req: {
                next_visit_recommended_date: snapshot.nextVisitDate, // "" はBEでNULLクリア
                version: snapshot.existingRecordVersion,
              } as UpdateMedicalRecordRequest,
            });
            break;
          }

          case "見積書":
            // BUG-016: 見積の永続化は post-save の estimateSave が正本。
            // ここで汎用「保存しました」を出すと API 未送信でも成功に見える。
            // 成功トーストは MedicalRecordEstimate が実 API 成功時のみ出す。
            queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
            return { success: true, timestamp: Date.now() };

          case "予防接種":
            // BUG-001: 永続化は inner VaccinationForm の createVaccination が正本。
            // success:true だと useMedicalRecordPostSave が markClean し、
            // 未保存の問診/診察 dirty を偽クリーンする。トーストも出さない。
            return { success: false, timestamp: Date.now() };

          default:
            // 治療 / 定期健診 / 検査 / 画像は子フォームが正本。未送信を成功にしない。
            return { success: false, timestamp: Date.now() };
        }

        toast.success("保存しました");
        queryClient.invalidateQueries({ queryKey: queryKeys.reception.all() });
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE,
  );

  return { formState, formAction, isSaving };
}
