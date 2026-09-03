import { useActionState, useLayoutEffect, useRef } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, STYLE } from "@/lib/design-tokens";
import { getFormString } from "@/lib/form-data";
import { handleApiError } from "@/lib/handle-api-error";
import { todayJSTISO } from "@/lib/jst-date";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import { useRecordPetDeath } from "@/hooks/use-record-pet-death";

interface PetDeceasedDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  petId: string;
  petName: string;
  petBreed?: string;
  petGender?: string;
  petAge?: string;
  canEdit?: boolean;
  /**
   * BUG-407: 保存成功時に呼ばれる。バックエンドへの即時保存はこのダイアログが
   * 既に完結させているが、この通知が無いと外側 PetEditModal のローカル
   * formData（生死ラジオ・deceasedAt）が古いまま残り、次に外側「更新」を
   * 押すと status="生存" で上書きされ deceased_at のみ残る不整合を再現する。
   */
  onRecorded?: (result: { deceasedAt: string; deceasedReason?: string }) => void;
}

export function PetDeceasedDialog({
  open,
  onOpenChange,
  petId,
  petName,
  petBreed,
  petGender,
  petAge,
  canEdit = false,
  onRecorded,
}: PetDeceasedDialogProps) {
  const mutation = useRecordPetDeath();
  const canEditRef = useRef(canEdit);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  const today = todayJSTISO();

  const [state, formAction, isPending] = useActionState(
    async (
      _prevState: ActionState<unknown>,
      formData: FormData,
    ): Promise<ActionState<unknown>> => {
      const deceasedAt = getFormString(formData, "deceased_at").trim();
      const deceasedReason = getFormString(formData, "deceased_reason").trim();

      if (!deceasedAt) {
        return {
          success: false,
          fieldErrors: { deceased_at: "死亡日を入力してください" },
          timestamp: Date.now(),
        };
      }

      if (deceasedAt > todayJSTISO()) {
        return {
          success: false,
          fieldErrors: { deceased_at: "未来の日付は指定できません" },
          timestamp: Date.now(),
        };
      }

      try {
        const normalizedReason = deceasedReason || undefined;
        if (canEditRef.current !== true) {
          return { success: false, timestamp: Date.now() };
        }
        await mutation.mutateAsync({
          petId,
          deceasedAt,
          deceasedReason: normalizedReason,
        });
        onRecorded?.({ deceasedAt, deceasedReason: normalizedReason });
        onOpenChange(false);
        return { success: true, error: null, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "ペット死亡記録");
        return {
          success: false,
          error: "死亡の記録に失敗しました",
          timestamp: Date.now(),
        };
      }
    },
    INITIAL_ACTION_STATE,
  );
  const deceasedAtError = state.fieldErrors?.deceased_at;

  const genderLabel =
    petGender === "male" ? "オス" : petGender === "female" ? "メス" : petGender;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle className={`text-base font-semibold ${C.text}`}>
            死亡を記録する
          </DialogTitle>
          <DialogDescription>
            死亡日と理由を記録し、このペットへの自動LINE配信を停止します。
          </DialogDescription>
        </DialogHeader>

        {/* Pet summary */}
        <div
          className={`rounded-xs border ${C.borderMedium} bg-white px-3 py-2.5 text-sm space-y-0.5`}
        >
          <p className={`font-medium ${C.text}`}>{petName}</p>
          <p className={C.text50}>
            {[petBreed, genderLabel, petAge].filter(Boolean).join(" / ")}
          </p>
        </div>

        <form
          id="pet-deceased-form"
          action={formAction}
          noValidate
          className="space-y-4"
        >
          {/* 死亡日 */}
          <div className="space-y-1.5">
            <label
              htmlFor="deceased_at"
              className={`block text-sm ${C.text70}`}
            >
              死亡日
              <span className={`ml-1 ${C.textRequired}`}>*</span>
            </label>
            <input
              id="deceased_at"
              name="deceased_at"
              type="date"
              defaultValue={today}
              max={today}
              required
              aria-invalid={deceasedAtError ? true : undefined}
              aria-describedby={
                deceasedAtError ? "pet-deceased-date-error" : undefined
              }
              className={`${STYLE.formInput} w-full rounded-xs border px-3 text-sm`}
              disabled={isPending}
            />
            {deceasedAtError ? (
              <p
                id="pet-deceased-date-error"
                className={`text-xs ${C.danger}`}
                role="alert"
              >
                {deceasedAtError}
              </p>
            ) : null}
          </div>

          {/* 死亡理由 */}
          <div className="space-y-1.5">
            <label
              htmlFor="deceased_reason"
              className={`block text-sm ${C.text70}`}
            >
              死亡理由
              <span className={`ml-1.5 text-xs ${C.text40}`}>（任意）</span>
            </label>
            <textarea
              id="deceased_reason"
              name="deceased_reason"
              rows={3}
              placeholder="例: 老衰、腫瘍など"
              className={`${STYLE.textarea} text-sm focus-visible:ring-2 ${C.focusRingAccent40}`}
              disabled={isPending}
            />
          </div>

          {/* 警告文 */}
          <div
            className={`rounded-xs border ${C.borderNotice} ${C.bgNotice40} px-3 py-2 text-xs ${C.textNotice}`}
          >
            記録後、このペットへの自動LINE配信が停止されます。
          </div>

          {/* バリデーションエラー */}
          {state.error ? (
            <p className={`text-xs ${C.danger}`} role="alert">
              {state.error}
            </p>
          ) : null}

          {/* FE-RC-023: 破壊的操作の共通 SubmitButton（colorVariant="destructive"）を使うため
              form 内へ移動（useFormStatus は最寄りの祖先 form を見るため、form 属性での外部
              関連付けだけでは pending を検知できない） */}
          <DialogFooter className="gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isPending}
            >
              キャンセル
            </Button>
            <SubmitButton colorVariant="destructive" loadingText="記録中...">
              死亡を記録する
            </SubmitButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
