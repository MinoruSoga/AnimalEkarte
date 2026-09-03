import { memo } from "react";
import { ICON, C } from "@/lib/design-tokens";
import { Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface DeleteIconButtonProps {
  onClick: () => void;
  disabled?: boolean;
  className?: string;
  title?: string;
}

/**
 * テーブル行・カード等でのアイコン削除ボタン共通コンポーネント。
 * shadcn Button variant="ghost" size="icon" + Trash2 アイコンを統一スタイルで提供する。
 */
export const DeleteIconButton = memo(function DeleteIconButton({
  onClick,
  disabled,
  className,
  title = "削除",
}: DeleteIconButtonProps) {
  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={onClick}
      disabled={disabled}
      title={title}
      className={cn(
        `size-11 min-h-11 min-w-11 ${C.text40} ${C.hoverTextDanger} ${C.hoverBgDanger5}`,
        className,
      )}
    >
      <Trash2 className={ICON.action} />
    </Button>
  );
});
