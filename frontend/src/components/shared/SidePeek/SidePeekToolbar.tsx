import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import X from "lucide-react/dist/esm/icons/x";
import { C, STYLE, ICON } from "@/lib/design-tokens";

interface SidePeekToolbarProps {
  isNew: boolean;
  onClose: () => void;
  onDelete?: () => void;
}

export function SidePeekToolbar({ isNew, onClose, onDelete }: SidePeekToolbarProps) {
  return (
    <div className={STYLE.sidePeekToolbar}>
      <span className={`text-xs ${C.text35} pl-1 select-none`}>
        {isNew ? "新規作成" : "編集"}
      </span>
      <div className="flex items-center gap-1">
        {onDelete ? (
          <button
            type="button"
            onClick={onDelete}
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
}
