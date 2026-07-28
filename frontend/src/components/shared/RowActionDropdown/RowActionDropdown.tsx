import { memo } from "react";
import { C, ICON } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { MoreHorizontal } from "lucide-react";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";

interface RowActionDropdownAction {
  label: string;
  icon?: React.ElementType;
  onClick: () => void;
  variant?: "default" | "destructive";
}

export interface RowActionDropdownProps {
  actions: RowActionDropdownAction[];
  ariaLabel: string;
}

export const RowActionDropdown = memo(function RowActionDropdown({
  actions,
  ariaLabel,
}: RowActionDropdownProps) {
  return (
    <div
      className="flex items-center justify-end gap-1"
      onClick={(e) => e.stopPropagation()}
    >
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            aria-label={ariaLabel}
            className={`h-11 w-11 ${C.text60} ${C.hoverText}`}
            onClick={(e) => e.stopPropagation()}
          >
            <MoreHorizontal className={ICON.page} />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {actions.map((action) => {
            const Icon = action.icon;
            return (
              <DropdownMenuItem
                key={action.label}
                onClick={(e) => {
                  e.stopPropagation();
                  action.onClick();
                }}
                className={
                  action.variant === "destructive"
                    ? `${C.danger} ${C.hoverBgDanger5}`
                    : ""
                }
              >
                {Icon ? <Icon className={`mr-2 ${ICON.action}`} /> : null}
                {action.label}
              </DropdownMenuItem>
            );
          })}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
});
