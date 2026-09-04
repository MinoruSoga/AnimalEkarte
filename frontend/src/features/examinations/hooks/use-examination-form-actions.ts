import { toast } from "sonner";
import type { ExaminationMutationPermissions } from "./use-examination-form-model";
import { UNCONFIRM_REASON_MAX_LENGTH } from "../lib/constants";

export function createExaminationUnconfirmHandler(input: {
  isEdit: boolean;
  id: string | undefined;
  isPersistedConfirmed: () => boolean;
  isMutationAllowed: (action: keyof ExaminationMutationPermissions) => boolean;
  unconfirm: (vars: { id: string; reason: string }) => Promise<unknown>;
}): (rawReason: string) => Promise<boolean> {
  return async (rawReason: string): Promise<boolean> => {
    const reason = rawReason.trim();
    if (!input.isEdit || !input.id) return false;
    if (!input.isPersistedConfirmed()) return false;
    if (!input.isMutationAllowed("canUnconfirm")) return false;
    if (!reason || reason.length > UNCONFIRM_REASON_MAX_LENGTH) return false;

    try {
      await input.unconfirm({ id: input.id, reason });
      toast.success("検査記録の確定を解除しました");
      return true;
    } catch {
      // onError は useUnconfirmExamination 側で handleApiError 済み。ここでは失敗を呼び出し元へ伝えるだけ。
      return false;
    }
  };
}

export function createExaminationDeleteHandler(input: {
  isEdit: boolean;
  id: string | undefined;
  isMutationAllowed: (action: keyof ExaminationMutationPermissions) => boolean;
  isResultsLocked: () => boolean;
  isPetExplicitlyDeceased: () => boolean;
  startDeleteTransition: (fn: () => void) => void;
  deleteExamination: (id: string, opts: { onSuccess: () => void }) => void;
}): (onSuccess?: () => void) => void {
  return (onSuccess?: () => void) => {
    if (!input.isEdit || !input.id) return;
    if (!input.isMutationAllowed("canDelete")) return;
    if (input.isResultsLocked()) return;
    if (input.isPetExplicitlyDeceased()) return;
    const examinationId = input.id;
    input.startDeleteTransition(() => {
      input.deleteExamination(examinationId, {
        onSuccess: () => {
          toast.success("検査記録を削除しました");
          onSuccess?.();
        },
      });
    });
  };
}
