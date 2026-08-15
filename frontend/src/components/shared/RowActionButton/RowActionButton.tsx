import { memo } from "react";
import { Button } from "@/components/ui/button";
import { Edit } from "lucide-react";
import { cn } from "@/components/ui/utils";
import { TABLE_STYLES, ICON } from "@/lib/design-tokens";

interface RowActionButtonProps {
  icon?: React.ElementType;
  onClick: (e: React.MouseEvent) => void;
  className?: string;
  "aria-label": string;
}

export const RowActionButton = memo(function RowActionButton({
  icon: Icon = Edit,
  onClick,
  className,
  "aria-label": ariaLabel,
}: RowActionButtonProps) {
  return (
    <div className="flex items-center justify-end gap-1">
      <Button
        variant="ghost"
        size="icon"
        aria-label={ariaLabel}
        className={cn(TABLE_STYLES.actionButton, className)}
        onClick={(e) => {
          e.stopPropagation();
          onClick(e);
        }}
      >
        <Icon className={ICON.page} />
      </Button>
    </div>
  );
});
