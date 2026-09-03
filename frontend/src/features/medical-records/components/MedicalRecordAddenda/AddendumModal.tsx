import { useActionState, useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { getFormString } from "@/lib/form-data";
import { useCreateMedicalRecordAddendum } from "../../hooks/use-medical-record-addenda";

const MAX_REASON_LENGTH = 500;

interface AddendumModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  medicalRecordId: string;
}

interface FormState {
  error: string | null;
}

export function AddendumModal({ open, onOpenChange, medicalRecordId }: AddendumModalProps) {
  const { mutateAsync } = useCreateMedicalRecordAddendum(medicalRecordId);
  const [afterText, setAfterText] = useState("");
  const [reason, setReason] = useState("");
  const reasonLength = reason.length;

  const [state, formAction] = useActionState(
    async (_prev: FormState, formData: FormData): Promise<FormState> => {
      const afterText = getFormString(formData, "after_text").trim();
      const reason = getFormString(formData, "reason").trim();

      if (!afterText) return { error: "修正内容は必須です" };
      if (!reason) return { error: "修正理由は必須です" };
      if (reason.length > MAX_REASON_LENGTH) {
        return { error: `修正理由は${MAX_REASON_LENGTH}文字以内で入力してください` };
      }

      try {
        await mutateAsync({ after_text: afterText, reason });
        onOpenChange(false);
        return { error: null };
      } catch (err) {
        const isAxiosError = (e: unknown): e is { response?: { status: number } } =>
          e !== null && typeof e === 'object' && 'response' in e;
        if (isAxiosError(err) && err.response?.status === 409) {
          return { error: "確定済みカルテでないため追記できません" };
        }
        handleApiError(err, "追記の保存");
        return { error: "追記の保存に失敗しました" };
      }
    },
    { error: null },
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>追記</DialogTitle>
          <DialogDescription>
            確定済みカルテに修正内容と理由を追記します。
          </DialogDescription>
        </DialogHeader>
        <form action={formAction} className="space-y-4">
          {state.error ? (
            <div
              role="alert"
              className={`text-sm ${C.textNotionRed} ${C.bgDanger8} border ${C.borderDanger20} rounded-xs px-3 py-2`}
            >
              {state.error}
            </div>
          ) : null}

          <div>
            <label
              htmlFor="addendum-after-text"
              className={`block text-sm font-medium ${C.text} mb-1`}
            >
              修正内容
              <span className={`ml-1 text-sm ${C.textNotionRed}`} aria-hidden="true">*</span>
            </label>
            <textarea
              id="addendum-after-text"
              name="after_text"
              aria-required="true"
              rows={5}
              value={afterText}
              onChange={(e) => setAfterText(e.target.value)}
              className={`w-full border ${C.borderMedium} rounded-xs px-3 py-2 text-base ${C.text} resize-y focus:outline-none focus:ring-1 ${C.focusRingBrand}`}
            />
          </div>

          <div>
            <label
              htmlFor="addendum-reason"
              className={`block text-sm font-medium ${C.text} mb-1`}
            >
              修正理由
              <span className={`ml-1 text-sm ${C.textNotionRed}`} aria-hidden="true">*</span>
            </label>
            <textarea
              id="addendum-reason"
              name="reason"
              aria-required="true"
              rows={3}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              className={`w-full border ${C.borderMedium} rounded-xs px-3 py-2 text-base ${C.text} resize-y focus:outline-none focus:ring-1 ${C.focusRingBrand}`}
            />
            <p
              className={`text-sm mt-1 text-right ${reasonLength > MAX_REASON_LENGTH ? C.textNotionRed : C.text60}`}
              aria-live="polite"
            >
              {reasonLength} / {MAX_REASON_LENGTH}
            </p>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              キャンセル
            </Button>
            <SubmitButton colorVariant="primary" loadingText="保存中...">追記を保存</SubmitButton>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
