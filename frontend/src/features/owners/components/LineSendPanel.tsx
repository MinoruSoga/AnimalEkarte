import { useActionState, useState } from "react";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, STYLE } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { useGetOwnerLineTags } from "../api/get-owner-line-tags";
import { useSendLineMessage } from "../api/send-line-message";
import type { LineSendType } from "../api/send-line-message";
import { uploadLineFile } from "../api/upload-line-file";
import { LineSendTypeSelector } from "./LineSendTypeSelector";
import { LineSendHistory } from "./LineSendHistory";

interface LineSendPanelProps {
  ownerId: string;
  ownerName: string;
  open: boolean;
  onOpenChange: (o: boolean) => void;
}

interface SendFormState {
  error: string | null;
  success: boolean;
}

const INITIAL_STATE: SendFormState = { error: null, success: false };

export function LineSendPanel({
  ownerId,
  ownerName,
  open,
  onOpenChange,
}: LineSendPanelProps) {
  const { data: lineData } = useGetOwnerLineTags(ownerId);
  const { mutateAsync: sendMessage } = useSendLineMessage(ownerId);

  const [sendType, setSendType] = useState<LineSendType>("text");

  const isLinked = lineData?.is_linked ?? false;
  const isOptOut = lineData?.lstep_opt_out ?? false;
  const canSend = isLinked && !isOptOut;

  const [formState, formAction] = useActionState(
    async (
      _prev: SendFormState,
      formData: FormData
    ): Promise<SendFormState> => {
      if (!isLinked) {
        return { error: "LINE未連携のため送信できません", success: false };
      }
      if (isOptOut) {
        return { error: "配信停止中のため送信できません", success: false };
      }

      const type = formData.get("send_type") as LineSendType;
      const purpose = (formData.get("purpose") as string | null)?.trim() || undefined;

      if (type === "text") {
        const text = (formData.get("text") as string | null)?.trim();
        if (!text) {
          return { error: "メッセージを入力してください", success: false };
        }
        try {
          await sendMessage({ message_type: "text", text, purpose });
          return { error: null, success: true };
        } catch (err: unknown) {
          handleApiError(err, "LINE送信");
          return { error: "送信に失敗しました。しばらく待ってから再試行してください", success: false };
        }
      }

      // pdf_url / image_url: upload first, then send with file_id
      const file = formData.get("file") as File | null;
      if (!file || file.size === 0) {
        return { error: "ファイルを選択してください", success: false };
      }
      let uploaded: { id: number; file_name: string; file_type: string };
      try {
        uploaded = await uploadLineFile(ownerId, file);
      } catch (err: unknown) {
        handleApiError(err, "ファイルアップロード");
        return { error: "ファイルのアップロードに失敗しました", success: false };
      }
      try {
        await sendMessage({
          message_type: type,
          file_id: uploaded.id,
          file_name: uploaded.file_name,
          purpose,
        });
        return { error: null, success: true };
      } catch (err: unknown) {
        handleApiError(err, "LINE送信");
        return { error: "送信に失敗しました。しばらく待ってから再試行してください", success: false };
      }
    },
    INITIAL_STATE
  );

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className={`${C.bgPage} w-full max-w-full sm:max-w-[480px] flex flex-col`}>
        <SheetHeader className={`border-b pr-16 ${C.borderLight}`}>
          <SheetTitle className={`text-base font-semibold ${C.text}`}>
            LINE送信 — {ownerName}
          </SheetTitle>
          <SheetDescription className="sr-only">
            {ownerName}さんへLINEメッセージを送信します
          </SheetDescription>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto flex flex-col gap-6 p-4">
          {!isLinked ? (
            <div
              className={`rounded-md border ${C.borderDanger} ${C.bgDanger8} px-4 py-3 text-sm ${C.danger}`}
            >
              LINE未連携のため送信できません
            </div>
          ) : null}

          {isLinked && isOptOut ? (
            <div
              className={`rounded-md border ${C.borderDanger} ${C.bgDanger8} px-4 py-3 text-sm ${C.danger}`}
            >
              配信停止中（opt-out）のため送信できません
            </div>
          ) : null}

          {canSend ? (
            <form action={formAction} className="flex flex-col gap-4">
              {/* 送信タイプ選択 */}
              <div className="flex flex-col gap-1.5">
                <span className={STYLE.formLabel}>送信タイプ</span>
                <LineSendTypeSelector
                  value={sendType}
                  onChange={(v) => setSendType(v)}
                  disabled={false}
                />
                {/* hidden field to pass send_type into FormData */}
                <input type="hidden" name="send_type" value={sendType} />
              </div>

              {/* テキスト入力 */}
              {sendType === "text" ? (
                <div className="flex flex-col gap-1.5">
                  <label htmlFor="line_text" className={STYLE.formLabel}>
                    メッセージ
                  </label>
                  <textarea
                    id="line_text"
                    name="text"
                    rows={5}
                    placeholder="送信するメッセージを入力してください"
                    className={`w-full rounded-xxs border ${C.borderMedium} bg-white p-3 text-sm ${C.text} ${C.textPlaceholder} outline-none ${C.focusBorderAccent} transition-colors resize-none leading-relaxed`}
                  />
                </div>
              ) : null}

              {/* ファイル入力 */}
              {sendType === "pdf_url" ? (
                <div className="flex flex-col gap-1.5">
                  <label htmlFor="line_file_pdf" className={STYLE.formLabel}>
                    PDFファイル
                  </label>
                  <input
                    id="line_file_pdf"
                    name="file"
                    type="file"
                    accept=".pdf"
                    className={`text-sm ${C.text} cursor-pointer`}
                  />
                </div>
              ) : null}

              {sendType === "image_url" ? (
                <div className="flex flex-col gap-1.5">
                  <label htmlFor="line_file_image" className={STYLE.formLabel}>
                    画像ファイル
                  </label>
                  <input
                    id="line_file_image"
                    name="file"
                    type="file"
                    accept=".jpg,.jpeg,.png"
                    className={`text-sm ${C.text} cursor-pointer`}
                  />
                </div>
              ) : null}

              {/* FE-003: 月間配信数消費の注意 */}
              <div className={`rounded-md border ${C.borderNotice} ${C.bgNotice} px-3 py-2 text-xs ${C.textNotice}`}>
                この送信はLINE Messaging APIを使用します。月間配信数を1件消費します。
              </div>

              {/* 送信目的 */}
              <div className="flex flex-col gap-1.5">
                <label htmlFor="line_purpose" className={STYLE.formLabel}>
                  送信目的
                  <span className={`ml-1 text-xs ${C.text40}`}>（任意）</span>
                </label>
                <input
                  id="line_purpose"
                  name="purpose"
                  type="text"
                  placeholder="例: 術後フォロー"
                  className={`${STYLE.formInput} rounded-xxs px-3`}
                />
              </div>

              {formState.error !== null ? (
                <p className={`text-sm ${C.danger}`}>{formState.error}</p>
              ) : null}

              {formState.success ? (
                <p className={`text-sm ${C.textStatusGreen}`}>
                  送信しました
                </p>
              ) : null}

              <SubmitButton loadingText="送信中..." colorVariant="primary">送信する</SubmitButton>
            </form>
          ) : null}

          {/* 送信履歴 */}
          <div className={`border-t ${C.borderLight} pt-4 flex flex-col gap-2`}>
            <span className={`text-xs ${C.text55} uppercase`}>
              送信履歴（最新5件）
            </span>
            <LineSendHistory ownerId={ownerId} />
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
