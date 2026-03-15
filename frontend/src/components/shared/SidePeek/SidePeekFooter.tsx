import { STYLE } from "@/lib/design-tokens";

interface SidePeekFooterProps {
  onCancel: () => void;
  onSave: () => void;
  saveLabel?: string;
  isPending?: boolean;
}

export function SidePeekFooter({
  onCancel,
  onSave,
  saveLabel = "保存",
  isPending = false,
}: SidePeekFooterProps) {
  return (
    <div className={STYLE.sidePeekFooter}>
      <button type="button" onClick={onCancel} className={STYLE.sidePeekCancelBtn}>
        キャンセル
      </button>
      <button
        type="button"
        onClick={onSave}
        disabled={isPending}
        className={`${STYLE.sidePeekSaveBtn} disabled:opacity-50 disabled:cursor-not-allowed`}
      >
        {isPending ? "保存中..." : saveLabel}
      </button>
    </div>
  );
}
