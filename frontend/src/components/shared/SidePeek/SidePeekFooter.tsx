import { STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";

interface SidePeekFooterProps {
  onCancel: () => void;
  onSave?: () => void;
  saveLabel?: string;
  isPending?: boolean;
}

export function SidePeekFooter({
  onCancel,
  onSave,
  saveLabel = "保存",
  isPending,
}: SidePeekFooterProps) {
  return (
    <div className={STYLE.sidePeekFooter}>
      <button type="button" onClick={onCancel} className={STYLE.sidePeekCancelBtn}>
        キャンセル
      </button>
      <SubmitButton
        onClick={onSave}
        disabled={isPending}
        className={STYLE.sidePeekSaveBtn}
        loadingText="保存中..."
      >
        {saveLabel}
      </SubmitButton>
    </div>
  );
}
