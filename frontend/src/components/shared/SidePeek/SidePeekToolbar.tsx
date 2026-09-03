import { memo } from "react";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import X from "lucide-react/dist/esm/icons/x";
import { C, STYLE, ICON } from "@/lib/design-tokens";

interface SidePeekToolbarProps {
  isNew: boolean;
  onClose: () => void;
  onDelete?: () => void;
  /** BUG-158: true の場合、ラベルを「詳細」に変更 */
  readOnly?: boolean;
}

export const SidePeekToolbar = memo(function SidePeekToolbar({
  isNew,
  onClose,
  onDelete,
  readOnly,
}: SidePeekToolbarProps) {
  return (
    <div className={STYLE.sidePeekToolbar}>
      <span className={`text-xs ${C.text35} pl-1 select-none`}>
        {readOnly ? "詳細" : isNew ? "新規作成" : "編集"}
      </span>
      <div className="flex items-center gap-1">
        {onDelete ? (
          <button
            type="button"
            onClick={onDelete}
            aria-label="削除"
            className={`${STYLE.sidePeekToolbarBtn} cursor-pointer ${C.danger} ${C.hoverBgDanger5}`}
          >
            <Trash2 className={ICON.toolbar} />
          </button>
        ) : null}
        <button
          type="button"
          onClick={onClose}
          className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
          aria-label="閉じる"
        >
          <X className={ICON.toolbar} />
        </button>
      </div>
    </div>
  );
});
