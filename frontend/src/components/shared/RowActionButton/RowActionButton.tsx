import { Button } from "@/components/ui/button";
import { Edit } from "lucide-react";
import { cn } from "@/components/ui/utils";
import { TABLE_STYLES } from "@/config/design-tokens";

interface RowActionButtonProps {
  icon?: React.ElementType;
  onClick: (e: React.MouseEvent) => void;
  className?: string;
}

export function RowActionButton({ icon: Icon = Edit, onClick, className }: RowActionButtonProps) {
  return (
    <div className="flex items-center justify-end gap-1" onClick={(e) => e.stopPropagation()}>
        <Button
        variant="ghost"
        size="icon"
        className={cn(TABLE_STYLES.actionButton, className)}
        onClick={(e) => {
            e.stopPropagation();
            onClick(e);
        }}
        >
        <Icon className="size-5" />
        </Button>
    </div>
  );
}
