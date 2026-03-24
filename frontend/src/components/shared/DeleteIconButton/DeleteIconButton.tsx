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
export function DeleteIconButton({
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
        "size-8 text-[#37352F]/40 hover:text-red-600 hover:bg-red-50",
        className
      )}
    >
      <Trash2 className="size-4" />
    </Button>
  );
}
