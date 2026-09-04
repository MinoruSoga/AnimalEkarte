import { C, ICON } from "@/lib/design-tokens";
import { memo, useState, useCallback } from "react";
import { ArrowUp, ArrowDown, ChevronDown, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
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
          className={`inline-flex items-center gap-1.5 h-8 px-3 text-base font-medium rounded-xxs ${C.bgDiscountLight} ${C.textDiscount} ${C.bgDiscountHover} transition-colors whitespace-nowrap`}
        >
          <DirectionIcon className={`${ICON.page} shrink-0`} />
          <span className="truncate max-w-[140px]">{property?.label ?? sort.key}</span>
          <ChevronDown className={`${ICON.page} shrink-0 opacity-60`} />
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0" align="start">
        <div className="py-1">
          {/* Direction toggle */}
          <button
            type="button"
            onClick={handleToggle}
            className={`w-full flex items-center gap-2 px-3 py-1.5 text-base ${C.text} ${C.hoverBgLight} transition-colors`}
          >
            {sort.direction === "asc" ? (
              <ArrowDown className={ICON.page} />
            ) : (
              <ArrowUp className={ICON.page} />
            )}
            {sort.direction === "asc" ? "降順に変更" : "昇順に変更"}
          </button>

          {/* Divider */}
          <div className={`border-t ${C.borderLight} my-1`} />

          {/* Property change */}
          <p className={`text-base ${C.text40} px-3 py-1`}>プロパティを変更</p>
          <Command>
            <CommandInput placeholder="検索..." />
            <CommandList>
              <CommandEmpty>見つかりません</CommandEmpty>
              {sortProperties.map((prop) => (
                <CommandItem
                  key={prop.key}
                  onSelect={() => handleSelectProperty(prop.key)}
                  className={`text-base ${prop.key === sort.key ? `${C.bgBrand8} ${C.textBrand}` : ""}`}
                >
                  {prop.icon ? <prop.icon className={`mr-2 ${ICON.xs} ${C.text50}`} /> : null}
                  {prop.label}
                </CommandItem>
              ))}
            </CommandList>
          </Command>

          {/* Divider */}
          <div className={`border-t ${C.borderLight} my-1`} />

          {/* Remove */}
          <button
            type="button"
            onClick={handleRemove}
            className={`w-full flex items-center gap-2 px-3 py-1.5 text-base ${C.danger} ${C.hoverBgDanger5} transition-colors`}
          >
            <X className={ICON.page} />
            削除
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
});
