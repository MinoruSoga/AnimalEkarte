import { toast } from "sonner";
import type { EntityReadResult } from "@/lib/entity-read-result";
import type { Pet } from "@/types";
import type { ExaminationRecord } from "../api/transforms";
import type { CreateExaminationRequest, UpdateExaminationRequest } from "../api/types";
import type { ActionState } from "@/types/form";
import type { ExamItemRow } from "../components/ExamItemsTable";
import {
  buildCreateExaminationRequest,
  buildUpdateExaminationRequest,
  decideExaminationPatientChange,
  rowsToRequest,
  validateExaminationSave,
  type ExaminationMutationPermissions,
} from "./use-examination-form-model";

interface ExaminationSaveDeps {
  id: string | undefined;
  isEdit: boolean;
  medicalRecordId: string;
  formDataWithPetRef: { current: Partial<ExaminationRecord> };
  formItemsRef: { current: readonly ExamItemRow[] };
  itemsReadyForIDRef: { current: string | undefined };
  activeExaminationIDRef: { current: string | undefined };
  isPersistedConfirmedRef: { current: boolean };
  isPersistedResultsLockedRef: { current: boolean };
  isPatientChangeLockedRef: { current: boolean };
  existingPetIdRef: { current: string | undefined };
  entityReadRef: { current: EntityReadResult<ExaminationRecord> };
  activePetRef: { current: Pet | undefined };
  isPetExplicitlyDeceased: () => boolean;
  isMutationAllowed: (action: keyof ExaminationMutationPermissions) => boolean;
  updateMutation: {
    mutateAsync: (vars: { id: string; req: UpdateExaminationRequest }) => Promise<unknown>;
  };
  createMutation: {
    mutateAsync: (req: CreateExaminationRequest) => Promise<unknown>;
  };
}

export async function runExaminationSave(
  deps: ExaminationSaveDeps,
): Promise<ActionState> {
  const current = deps.formDataWithPetRef.current;
  const isPersistedConfirmed = deps.isPersistedConfirmedRef.current;
  const resultsLocked = deps.isPersistedResultsLockedRef.current;
  const errors = validateExaminationSave({
    current,
    isEdit: deps.isEdit,
    isCurrentEditTarget: deps.activeExaminationIDRef.current === deps.id,
    resultsLocked,
    areCurrentItemsReady: deps.itemsReadyForIDRef.current === deps.id,
    formItems: deps.formItemsRef.current,
  });

  if (Object.keys(errors).length > 0) {
    return { success: false, fieldErrors: errors, timestamp: Date.now() };
  }

  try {
    // サーバ保存済みの結果ロック中は items を省略（BE は confirmed / 初回 completed への item 更新を 409）。
    // ドラフトでステータス「確定」を選んだだけの遷移保存では items を送る（A-S02-01）。
    const items = rowsToRequest(deps.formItemsRef.current);

    if (deps.isPetExplicitlyDeceased()) {
      return { success: false, timestamp: Date.now() };
    }

    if (deps.isEdit && deps.id) {
      // BUG-016: edit route without a found entity must not create/update
      if (deps.entityReadRef.current.status !== "found") {
        return { success: false, timestamp: Date.now() };
      }
      const patientChange = decideExaminationPatientChange({
        currentPetId: current.petId,
        existingPetId: deps.existingPetIdRef.current,
        isPatientChangeLocked: deps.isPatientChangeLockedRef.current,
        changedPatient: deps.activePetRef.current,
      });
      if (patientChange.kind === "blocked") {
        toast.error(
          "患者変更の条件が変わりました。検査記録を再読み込みしてください",
        );
        return { success: false, timestamp: Date.now() };
      }
      const req = buildUpdateExaminationRequest({
        current,
        items,
        resultsLocked,
        patientChange,
      });
      // 確定済みは親フィールド更新も拒否。完了シールは status 遷移保存を許可。
      if (isPersistedConfirmed) {
        return { success: false, timestamp: Date.now() };
      }
      if (!deps.isMutationAllowed("canEdit")) {
        return { success: false, timestamp: Date.now() };
      }
      await deps.updateMutation.mutateAsync({ id: deps.id, req });
    } else {
      const pet = deps.activePetRef.current;
      if (!pet) return { success: false, timestamp: Date.now() };
      const req = buildCreateExaminationRequest({
        current,
        medicalRecordId: deps.medicalRecordId,
        petId: pet.id,
        items,
      });
      if (
        !deps.isMutationAllowed("canCreate") ||
        !deps.isMutationAllowed("canEdit")
      ) {
        return { success: false, timestamp: Date.now() };
      }
      await deps.createMutation.mutateAsync(req);
    }
    return { success: true, timestamp: Date.now() };
  } catch {
    // FE-RC-005: useCreateExamination/useUpdateExamination の onError が
    // handleApiError 済み（api/create-examination.ts, api/update-examination.ts）。
    // ここでは二重 toast を避け、失敗を呼び出し元へ伝えるだけにする。
    return { success: false, timestamp: Date.now() };
  }
}
