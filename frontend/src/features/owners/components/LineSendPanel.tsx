import { useActionState, useState } from "react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, STYLE } from "@/lib/design-tokens";
import { useGetOwnerLineTags } from "../api/get-owner-line-tags";
import { useSendLineMessage } from "../api/send-line-message";
import type { LineSendType } from "../api/send-line-message";
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

  const [formState, formAction] = useActionState(
    async (
      _prev: SendFormState,
      formData: FormData
    ): Promise<SendFormState> => {
      if (!isLinked) {
        return { error: "LINE未連携のため送信できません", success: false };
      }

      const type = formData.get("send_type") as LineSendType;
      const purposeTag = (formData.get("purpose_tag") as string | null)?.trim() || undefined;

      if (type === "text") {
        const text = (formData.get("text") as string | null)?.trim();
        if (!text) {
          return { error: "メッセージを入力してください", success: false };
        }
        try {
          await sendMessage({ message_type: "text", text, purpose_tag: purposeTag });
          return { error: null, success: true };
        } catch {
          return { error: null, success: false };
        }
      }

      // pdf / image: file upload — use placeholder URL
      const file = formData.get("file") as File | null;
      if (!file || file.size === 0) {
        return { error: "ファイルを選択してください", success: false };
      }
      // Placeholder: in production this URL comes from a pre-signed S3 upload
      const fileUrl = `pending-upload:${file.name}`;
      try {
        await sendMessage({
          message_type: type,
          file_url: fileUrl,
          purpose_tag: purposeTag,
        });
        return { error: null, success: true };
      } catch {
        return { error: null, success: false };
      }
    },
    INITIAL_STATE
  );

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className={`${C.bgPage} w-[480px] sm:max-w-[480px] flex flex-col`}>
        <SheetHeader className={`border-b ${C.borderLight}`}>
          <SheetTitle className={`text-base font-semibold ${C.text}`}>
            LINE送信 — {ownerName}
          </SheetTitle>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto flex flex-col gap-5 p-4">
          {!isLinked ? (
            <div
              className={`rounded-md border ${C.borderNotice} ${C.bgNotice} px-4 py-3 text-sm ${C.textNotice}`}
            >
              LINE未連携のため送信できません
            </div>
          ) : null}

          {isLinked ? (
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
                    className={`w-full rounded-[3px] border ${C.borderMedium} bg-white p-3 text-sm ${C.text} ${C.textPlaceholder} outline-none ${C.focusBorderAccent} transition-colors resize-none leading-relaxed`}
                  />
                </div>
              ) : null}

              {/* ファイル入力 */}
              {sendType === "pdf" ? (
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

              {sendType === "image" ? (
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

              {/* 送信目的タグ */}
              <div className="flex flex-col gap-1.5">
                <label htmlFor="line_purpose_tag" className={STYLE.formLabel}>
                  送信目的タグ
                  <span className={`ml-1 text-xs ${C.text40}`}>（任意）</span>
                </label>
                <input
                  id="line_purpose_tag"
                  name="purpose_tag"
                  type="text"
                  placeholder="例: 術後フォロー"
                  className={`${STYLE.formInput} rounded-[3px] px-3`}
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

              <SubmitButton loadingText="送信中...">送信する</SubmitButton>
            </form>
          ) : null}

          {/* 送信履歴 */}
          <div className={`border-t ${C.borderLight} pt-4 flex flex-col gap-2`}>
            <span className={`text-xs ${C.text55} uppercase tracking-wide`}>
              送信履歴（最新5件）
            </span>
            <LineSendHistory ownerId={ownerId} />
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
