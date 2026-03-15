import { STYLE } from "@/lib/design-tokens";

interface SidePeekFooterProps {
  onCancel: () => void;
  onSave: () => void;
  saveLabel?: string;
}

export function SidePeekFooter({ onCancel, onSave, saveLabel = "保存" }: SidePeekFooterProps) {
  return (
    <div className={STYLE.sidePeekFooter}>
      <button type="button" onClick={onCancel} className={STYLE.sidePeekCancelBtn}>
        キャンセル
      </button>
      <button type="button" onClick={onSave} className={STYLE.sidePeekSaveBtn}>
        {saveLabel}
      </button>
    </div>
  );
}
