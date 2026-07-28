// React/Framework
import { ICON, C, STYLE } from "@/lib/design-tokens";

// External
import { Pencil } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";

// Relative
import { TypeIcon, StatusBadge, TimingBadges } from "./CarePlanBadges";

// Types
import type { CarePlanItem } from "../../api/care-plan-items";

interface ItemRowProps {
    item: CarePlanItem;
    onEdit?: (id: string) => void;
    onDelete?: (id: string) => void;
    isDeleting: boolean;
}

export function ItemRow({ item, onEdit, onDelete, isDeleting }: ItemRowProps) {
    return (
        <div className={`flex flex-wrap sm:flex-nowrap items-center gap-3 py-2 px-3 rounded-lg ${C.hoverBgPage} transition-colors`}>
            <TypeIcon type={item.type} />
            <span className={`min-w-[8rem] flex-1 text-sm font-medium ${C.text} truncate`}>{item.name}</span>
            <span className={`shrink-0 text-xs ${C.text60}`}>
                単価 ￥{item.unit_price.toLocaleString()}
            </span>
            <TimingBadges timing={item.timing} />
            <StatusBadge status={item.status} />
            <div className="flex gap-1 shrink-0">
                {onEdit !== undefined ? (
                    <Button
                        variant="ghost"
                        size="icon"
                        className="min-h-11 min-w-11"
                        onClick={() => onEdit(item.id)}
                        aria-label="編集"
                    >
                        <Pencil className={ICON.action} />
                    </Button>
                ) : null}
                {onDelete !== undefined ? (
                    <DeleteIconButton
                        onClick={() => onDelete(item.id)}
                        disabled={isDeleting}
                        className={`${STYLE.iconBtn28} min-h-11 min-w-11`}
                    />
                ) : null}
            </div>
        </div>
    );
}
