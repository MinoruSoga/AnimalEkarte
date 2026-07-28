import { memo } from "react";
import { STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";

interface SidePeekFooterProps {
  onCancel: () => void;
  onSave?: () => void;
  saveLabel?: string;
  isPending?: boolean;
  /** BUG-158: true の場合、保存ボタンを非表示にし「閉じる」のみ表示 */
  readOnly?: boolean;
}

export const SidePeekFooter = memo(function SidePeekFooter({
  onCancel,
  onSave,
  saveLabel = "保存",
  isPending,
  readOnly = false,
}: SidePeekFooterProps) {
  // BUG-158: readOnly の場合のみ「閉じる」ボタン表示
  // onSave が undefined でも form action パターンでは SubmitButton が必要（BUG-161）
  if (readOnly) {
    return (
      <div className={STYLE.sidePeekFooter}>
        <button type="button" onClick={onCancel} className={STYLE.sidePeekCancelBtn}>
          閉じる
        </button>
      </div>
    );
  }

  return (
    <div className={STYLE.sidePeekFooter}>
      <button type="button" onClick={onCancel} className={STYLE.sidePeekCancelBtn}>
        キャンセル
      </button>
      <SubmitButton
        onClick={onSave}
        disabled={isPending}
        className="h-9 px-4"
        loadingText="保存中..."
      >
        {saveLabel}
      </SubmitButton>
    </div>
  );
});
