import { memo, useState, useCallback } from "react";
import { ArrowUp, ArrowDown, ChevronDown, X } from "lucide-react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Command,
  CommandInput,
  CommandItem,
  CommandList,
  CommandEmpty,
} from "@/components/ui/command";
import type { SortProperty, ActiveSort } from "./types";

interface SortPillProps {
  sort: ActiveSort;
  sortProperties: SortProperty[];
  onToggleDirection: (key: string) => void;
  onChangeProperty: (oldKey: string, newKey: string) => void;
  onRemove: (key: string) => void;
}

export const SortPill = memo(function SortPill({
  sort,
  sortProperties,
  onToggleDirection,
  onChangeProperty,
  onRemove,
}: SortPillProps) {
  const [open, setOpen] = useState(false);
  const property = sortProperties.find((p) => p.key === sort.key);

  const handleToggle = useCallback(() => {
    onToggleDirection(sort.key);
  }, [sort.key, onToggleDirection]);

  const handleRemove = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      onRemove(sort.key);
    },
    [sort.key, onRemove],
  );

  const handleSelectProperty = useCallback(
    (newKey: string) => {
      onChangeProperty(sort.key, newKey);
      setOpen(false);
    },
    [sort.key, onChangeProperty],
  );

  const DirectionIcon = sort.direction === "asc" ? ArrowUp : ArrowDown;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="inline-flex items-center gap-1.5 h-8 px-3 text-sm font-medium rounded-[3px] bg-[#FAEBDD] text-[#D9730D] hover:bg-[#F5DCC3] transition-colors whitespace-nowrap"
        >
          <DirectionIcon className="h-5 w-5 shrink-0" />
          <span className="truncate max-w-[140px]">
            {property?.label ?? sort.key}
          </span>
          <ChevronDown className="h-5 w-5 shrink-0 opacity-60" />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0" align="start">
        <div className="py-1">
          {/* Direction toggle */}
          <button
            type="button"
            onClick={handleToggle}
            className="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-[#37352F] hover:bg-[#F1F1EF] transition-colors"
          >
            {sort.direction === "asc" ? (
              <ArrowDown className="h-5 w-5" />
            ) : (
              <ArrowUp className="h-5 w-5" />
            )}
            {sort.direction === "asc" ? "降順に変更" : "昇順に変更"}
          </button>

          {/* Divider */}
          <div className="border-t border-[rgba(55,53,47,0.09)] my-1" />

          {/* Property change */}
          <p className="text-xs text-[#37352F]/40 px-3 py-1">
            プロパティを変更
          </p>
          <Command>
            <CommandInput placeholder="検索..." />
            <CommandList>
              <CommandEmpty>見つかりません</CommandEmpty>
              {sortProperties.map((prop) => (
                <CommandItem
                  key={prop.key}
                  onSelect={() => handleSelectProperty(prop.key)}
                  className={`text-sm ${prop.key === sort.key ? "bg-[#2383E2]/10 text-[#2383E2]" : ""}`}
                >
                  {prop.icon ? (
                    <prop.icon className="mr-2 h-5 w-5 text-[#37352F]/50" />
                  ) : null}
                  {prop.label}
                </CommandItem>
              ))}
            </CommandList>
          </Command>

          {/* Divider */}
          <div className="border-t border-[rgba(55,53,47,0.09)] my-1" />

          {/* Remove */}
          <button
            type="button"
            onClick={handleRemove}
            className="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-[#EB5757] hover:bg-[#EB5757]/5 transition-colors"
          >
            <X className="h-5 w-5" />
            削除
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
});
